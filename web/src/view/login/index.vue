<template>
  <div id="userLayout" class="w-full h-full relative">
    <div
      class="rounded-lg flex items-center justify-evenly w-full h-full md:w-screen md:h-screen md:bg-[#194bfb] bg-white"
    >
      <div class="md:w-3/5 w-10/12 h-full flex items-center justify-evenly">
        <div
          class="oblique h-[130%] w-3/5 bg-white dark:bg-slate-900 transform -rotate-12 absolute -ml-52"
        />
        <!-- 分割斜块 -->
        <div
          class="z-[999] pt-12 pb-10 md:w-96 w-full rounded-lg flex flex-col justify-between box-border"
        >
          <div>
            <div class="flex items-center justify-center">
              <Logo :size="6" />
            </div>
            <div class="mb-9">
              <p class="text-center text-4xl font-bold">
                {{ $LIGHTNINGRAG.appName }}
              </p>
              <p class="text-center text-sm font-normal text-gray-500 mt-2.5">
                {{ $t('login.subtitle') }}
              </p>
            </div>
            <el-form
              ref="loginForm"
              :model="loginFormData"
              :rules="rules"
              :validate-on-rule-change="false"
              @keyup.enter="submitForm"
            >
              <el-form-item prop="username" class="mb-6">
                <el-input
                  v-model="loginFormData.username"
                  size="large"
                  :placeholder="$t('login.usernamePlaceholder')"
                  suffix-icon="user"
                />
              </el-form-item>
              <el-form-item prop="password" class="mb-6">
                <el-input
                  v-model="loginFormData.password"
                  show-password
                  size="large"
                  type="password"
                  :placeholder="$t('login.passwordPlaceholder')"
                />
              </el-form-item>
              <el-form-item
                v-if="loginFormData.openCaptcha"
                prop="captcha"
                class="mb-6"
              >
                <div class="flex w-full justify-between">
                  <el-input
                    v-model="loginFormData.captcha"
                    :placeholder="$t('login.captchaPlaceholder')"
                    size="large"
                    class="flex-1 mr-5"
                  />
                  <div class="w-1/3 h-11 bg-[#c3d4f2] rounded">
                    <img
                      v-if="picPath"
                      class="w-full h-full"
                      :src="picPath"
                      :alt="$t('login.captchaAlt')"
                      @click="loginVerify()"
                    />
                  </div>
                </div>
              </el-form-item>
              <el-form-item class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="submitForm"
                  >{{ $t('login.loginBtn') }}</el-button
                >
              </el-form-item>
              <template v-if="oauthProviders.length">
                <el-divider content-position="center">{{ $t('login.oauthDivider') }}</el-divider>
                <div class="flex flex-wrap gap-2 justify-center mb-6">
                  <el-button
                    v-for="p in oauthProviders"
                    :key="p.kind"
                    size="default"
                    @click="startOAuth(p.kind)"
                  >
                    <span class="inline-flex items-center gap-2">
                      <img
                        v-if="oauthProviderButtonIcon(p)"
                        :src="oauthProviderButtonIcon(p)"
                        alt=""
                        class="h-5 w-5 object-contain shrink-0"
                      />
                      <span>{{ oauthProviderLabel(p) }}</span>
                    </span>
                  </el-button>
                </div>
              </template>
              <el-form-item class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="checkInit"
                  >{{ $t('login.goInitBtn') }}</el-button
                >
              </el-form-item>
            </el-form>
          </div>
        </div>
      </div>
      <div class="hidden md:block w-1/2 h-full float-right bg-[#194bfb]">
        <img
          class="h-full"
          src="@/assets/login_right_banner.jpg"
          alt="banner"
        />
      </div>
    </div>

    <BottomInfo class="left-0 right-0 absolute bottom-3 mx-auto w-full z-20">
      <div class="links items-center justify-center gap-2 hidden md:flex">
        <a href="https://lightningrag.com/" target="_blank">
          <img src="@/assets/docs.png" class="w-8 h-8" :alt="$t('login.docsAlt')" />
        </a>
        <a href="https://support.qq.com/product/371961" target="_blank">
          <img src="@/assets/kefu.png" class="w-8 h-8" :alt="$t('login.supportAlt')" />
        </a>
        <a
          href="https://github.com/LightningRAG/LightningRAG"
          target="_blank"
        >
          <img src="@/assets/github.png" class="w-8 h-8" :alt="$t('login.githubAlt')" />
        </a>
      </div>
    </BottomInfo>
  </div>
</template>

<script setup>
  import { captcha, oauthPublicProviders, oauthExchange } from '@/api/user'
  import { checkDB } from '@/api/initdb'
  import BottomInfo from '@/components/bottomInfo/bottomInfo.vue'
  import { reactive, ref, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import { useRouter, useRoute } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'
  import Logo from '@/components/logo/index.vue'

  defineOptions({
    name: 'Login'
  })

  const router = useRouter()
  const route = useRoute()
  const { t, te } = useI18n()
  const captchaRequiredLength = ref(6)
  // 验证函数
  const checkUsername = (rule, value, callback) => {
    if (value.length < 5) {
      return callback(new Error(t('login.rulesUsername')))
    } else {
      callback()
    }
  }
  const checkPassword = (rule, value, callback) => {
    if (value.length < 6) {
      return callback(new Error(t('login.rulesPassword')))
    } else {
      callback()
    }
  }
  const checkCaptcha = (rule, value, callback) => {
    if (!loginFormData.openCaptcha) {
      return callback()
    }
    const sanitizedValue = (value || '').replace(/\s+/g, '')
    if (!sanitizedValue) {
      return callback(new Error(t('login.rulesCaptchaEmpty')))
    }
    if (!/^\d+$/.test(sanitizedValue)) {
      return callback(new Error(t('login.rulesCaptchaDigits')))
    }
    if (sanitizedValue.length < captchaRequiredLength.value) {
      return callback(
        new Error(
          t('login.rulesCaptchaLength', { n: captchaRequiredLength.value })
        )
      )
    }
    if (sanitizedValue !== value) {
      loginFormData.captcha = sanitizedValue
    }
    callback()
  }

  // 获取验证码
  const loginVerify = async () => {
    const ele = await captcha()
    captchaRequiredLength.value = Number(ele.data?.captchaLength) || 0
    picPath.value = ele.data?.picPath
    loginFormData.captchaId = ele.data?.captchaId
    loginFormData.openCaptcha = ele.data?.openCaptcha
  }
  loginVerify()

  // 登录相关操作
  const loginForm = ref(null)
  const picPath = ref('')
  const loginFormData = reactive({
    username: 'admin',
    password: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })
  const rules = reactive({
    username: [{ validator: checkUsername, trigger: 'blur' }],
    password: [{ validator: checkPassword, trigger: 'blur' }],
    captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
  })

  const userStore = useUserStore()
  const oauthProviders = ref([])
  const oauthCallbackPathPattern = ref('')
  const oauthCallbackPathSuffix = '/base/oauth/callback/{kind}'

  /** 接口 buttonIcon 与 defaultButtonIconsByKind 合并后展示（未配置时用各平台默认） */
  const oauthProviderButtonIcon = (p) => String(p?.buttonIcon || '').trim()

  /** 登录按钮文案：优先 i18n（login.oauthProviders.{kind}），否则回退接口 displayName */
  const oauthProviderLabel = (p) => {
    const kind = String(p?.kind || '').trim().toLowerCase()
    const key = kind ? `login.oauthProviders.${kind}` : ''
    if (key && te(key)) {
      return t(key)
    }
    const name = String(p?.displayName || '').trim()
    return name || kind || 'OAuth'
  }

  /** 与 axios baseURL 一致：相对路径时以当前页 origin 为前缀（dev 下即 /api 走 Vite 代理） */
  const oauthApiBase = () => {
    const raw = import.meta.env.VITE_BASE_API || ''
    if (raw.startsWith('http://') || raw.startsWith('https://')) {
      return raw.replace(/\/$/, '')
    }
    const path = raw.startsWith('/') ? raw : `/${raw}`
    if (typeof window !== 'undefined') {
      return `${window.location.origin}${path}`.replace(/\/$/, '')
    }
    const origin = String(import.meta.env.VITE_BASE_PATH || '').replace(/\/$/, '')
    return `${origin}${path}`.replace(/\/$/, '')
  }

  const buildOAuthAuthorizeUrl = (kind, callbackPattern) => {
    const k = encodeURIComponent(kind)
    const apiBase = oauthApiBase().replace(/\/$/, '')
    const pat = (callbackPattern || '').trim()
    // 必须用 apiBase（含 /api），不能只用 origin：否则 dev 会打开 /base/oauth/... 被 Vite 当前端路由，出现 …/authorize/github#/login
    if (pat.endsWith(oauthCallbackPathSuffix)) {
      const routerPrefix = pat.slice(0, -oauthCallbackPathSuffix.length)
      const authorizePath = `${routerPrefix}/base/oauth/authorize/${k}`.replace(
        /\/{2,}/g,
        '/'
      )
      const p = authorizePath.startsWith('/') ? authorizePath : `/${authorizePath}`
      return `${apiBase}${p}`
    }
    return `${apiBase}/base/oauth/authorize/${k}`
  }

  const startOAuth = (kind) => {
    const url = buildOAuthAuthorizeUrl(kind, oauthCallbackPathPattern.value)
    window.location.assign(url)
  }

  const oauthFailureMessage = (code) => {
    const c = String(code || '').trim()
    if (c === '1' || c === 'true') {
      return t('login.oauthErr.unknown')
    }
    const key = `login.oauthErr.${c}`
    const msg = t(key)
    return msg !== key ? msg : t('login.oauthFailed')
  }

  const tryOAuthExchange = async () => {
    const ex = route.query.oauth_ex
    const errRaw = route.query.oauth_err
    const err =
      errRaw === undefined || errRaw === null
        ? ''
        : Array.isArray(errRaw)
          ? errRaw[0]
          : errRaw
    if (String(err || '').length > 0) {
      ElMessage.error(oauthFailureMessage(err))
      await router.replace({ name: 'Login', query: {} })
      return
    }
    if (!ex) return
    const exId = Array.isArray(ex) ? ex[0] : ex
    const exTrim = String(exId || '').trim()
    if (exTrim.length < 20 || exTrim.length > 128) {
      ElMessage.error(t('login.oauthExchangeFailed'))
      await router.replace({ name: 'Login', query: {} })
      return
    }
    try {
      const res = await oauthExchange(exTrim)
      if (res.code !== 0 || !res.data) {
        ElMessage.error(res.msg || t('login.oauthExchangeFailed'))
        await router.replace({ name: 'Login', query: {} })
        return
      }
      const ok = await userStore.LoginInWithOAuthPayload(res.data)
      if (!ok) {
        ElMessage.error(t('login.oauthExchangeFailed'))
        await router.replace({ name: 'Login', query: {} })
      }
    } catch (e) {
      if (e?.response?.status === 429) {
        ElMessage.error(t('common.request.tooManyRequests'))
      } else {
        ElMessage.error(t('login.oauthExchangeFailed'))
      }
      await router.replace({ name: 'Login', query: {} })
    }
  }

  const normalizeOAuthPublicPayload = (d) => {
    let list = []
    let callbackPath = ''
    let defaults = {}
    if (Array.isArray(d)) {
      list = d
    } else if (d && typeof d === 'object') {
      list = Array.isArray(d.providers) ? d.providers : []
      callbackPath = String(d.callbackPathPattern || '').trim()
      const dm = d.defaultButtonIconsByKind
      if (dm && typeof dm === 'object' && !Array.isArray(dm)) {
        defaults = dm
      }
    }
    const normKind = (k) => String(k || '').trim().toLowerCase()
    const iconFallback = (kind) => {
      const k = normKind(kind)
      const v = defaults[k] ?? defaults[String(kind || '').trim()] ?? defaults._default
      return String(v || '').trim()
    }
    const merged = list.map((p) => {
      const bi = String(p.buttonIcon || '').trim()
      return { ...p, buttonIcon: bi || iconFallback(p.kind) }
    })
    return { list: merged, callbackPath }
  }

  onMounted(async () => {
    try {
      const pr = await oauthPublicProviders()
      if (pr.code === 0 && pr.data) {
        const { list, callbackPath } = normalizeOAuthPublicPayload(pr.data)
        oauthProviders.value = list
        oauthCallbackPathPattern.value = callbackPath
      }
    } catch {
      /* 未配置后端或网络错误时不展示第三方入口 */
    }
    await tryOAuthExchange()
  })
  const login = async () => {
    return await userStore.LoginIn(loginFormData)
  }
  const submitForm = () => {
    loginForm.value.validate(async (v) => {
      if (!v) {
        // 未通过前端静态验证
        ElMessage({
          type: 'error',
          message: t('login.fillLoginInfo'),
          showClose: true
        })
        return false
      }

      // 通过验证，请求登陆
      const flag = await login()

      // 登陆失败，刷新验证码
      if (!flag) {
        await loginVerify()
        return false
      }

      // 登陆成功
      return true
    })
  }

  // 跳转初始化
  const checkInit = async () => {
    const res = await checkDB()
    if (res.code === 0) {
      if (res.data?.needInit) {
        userStore.NeedInit()
        await router.push({ name: 'Init' })
      } else {
        ElMessage({
          type: 'info',
          message: t('login.dbAlreadyConfigured'),
          showClose: true
        })
      }
    }
  }
</script>
