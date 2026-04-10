<template>
  <div
    class="rounded-lg flex items-center justify-evenly w-full h-full relative md:w-screen md:h-screen md:bg-[#194bfb] overflow-hidden"
  >
    <div
      class="rounded-md w-full h-full flex items-center justify-center overflow-hidden"
    >
      <div
        class="oblique h-[130%] w-3/5 bg-white dark:bg-slate-900 transform -rotate-12 absolute -ml-80"
      />
      <div
        v-if="!page.showForm"
        :class="[page.showReadme ? 'slide-out-right' : 'slide-in-fwd-top']"
      >
        <div class="text-lg">
          <div
            class="font-sans text-4xl font-bold text-center mb-4 dark:text-white"
          >
            LightningRAG
          </div>
          <p class="text-gray-600 dark:text-gray-300 mb-2">{{ $t('init.noticeTitle') }}</p>
          <p class="text-gray-600 dark:text-gray-300 mb-2">
            {{ $t('init.line1') }}
          </p>
          <p class="text-gray-600 dark:text-gray-300 mb-2">
            {{ $t('init.line2a') }}<a
              class="text-blue-600 font-bold"
              href="https://lightningrag.com"
              target="_blank"
              >{{ $t('init.officialDocsLink') }}</a
            >{{ $t('init.line2b') }}
          </p>
          <p class="text-gray-600 dark:text-gray-300 mb-2">
            {{ $t('init.line3') }}
          </p>
          <p class="text-gray-600 dark:text-gray-300 mb-2">
            {{ $t('init.line4a') }}<span
              class="text-red-600 font-bold text-3xl ml-2"
              >{{ $t('init.engineInnoDB') }}</span
            >{{ $t('init.line4b') }}
          </p>
          <p class="text-gray-600 dark:text-gray-300 mb-2">
            {{ $t('init.note') }}
          </p>
          <p class="flex items-center justify-between mt-8">
            <el-button type="primary" size="large" @click="goDoc">
              {{ $t('init.readDoc') }}
            </el-button>
            <el-button type="primary" size="large" @click="showNext">
              {{ $t('init.confirmed') }}
            </el-button>
          </p>
        </div>
      </div>
      <div
        v-if="page.showForm"
        :class="[page.showForm ? 'slide-in-left' : 'slide-out-right']"
        class="w-96"
      >
        <el-form ref="formRef" :model="form" label-width="100px" size="large">
          <el-form-item :label="$t('init.adminPassword')">
            <el-input
              v-model="form.adminPassword"
              :placeholder="$t('init.adminPasswordPh')"
            ></el-input>
          </el-form-item>
          <el-form-item :label="$t('init.dbType')">
            <el-select
              v-model="form.dbType"
              :placeholder="$t('init.selectPlaceholder')"
              class="w-full"
              @change="changeDB"
            >
              <el-option key="mysql" label="mysql" value="mysql" />
              <el-option key="pgsql" label="pgsql" value="pgsql" />
              <el-option key="oracle" label="oracle" value="oracle" />
              <el-option key="mssql" label="mssql" value="mssql" />
              <el-option key="sqlite" label="sqlite" value="sqlite" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="form.dbType !== 'sqlite'" :label="$t('init.host')">
            <el-input v-model="form.host" :placeholder="$t('init.hostPh')" />
          </el-form-item>
          <el-form-item v-if="form.dbType !== 'sqlite'" :label="$t('init.port')">
            <el-input v-model="form.port" :placeholder="$t('init.portPh')" />
          </el-form-item>
          <el-form-item v-if="form.dbType !== 'sqlite'" :label="$t('init.userName')">
            <el-input
              v-model="form.userName"
              :placeholder="$t('init.userNamePh')"
            />
          </el-form-item>
          <el-form-item v-if="form.dbType !== 'sqlite'" :label="$t('init.password')">
            <el-input
              v-model="form.password"
              :placeholder="$t('init.passwordPh')"
            />
          </el-form-item>
          <el-form-item :label="$t('init.dbName')">
            <el-input v-model="form.dbName" :placeholder="$t('init.dbNamePh')" />
          </el-form-item>
          <el-form-item v-if="form.dbType === 'sqlite'" :label="$t('init.dbPath')">
            <el-input
              v-model="form.dbPath"
              :placeholder="$t('init.dbPathPh')"
            />
          </el-form-item>
          <el-form-item v-if="form.dbType === 'pgsql'" :label="$t('init.template')">
            <el-input
              v-model="form.template"
              :placeholder="$t('init.templatePh')"
            />
          </el-form-item>
          <el-form-item>
            <div style="text-align: right">
              <el-button type="primary" @click="onSubmit">{{ $t('init.submit') }}</el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <div class="hidden md:block w-1/2 h-full float-right bg-[#194bfb]">
      <img class="h-full" src="@/assets/login_right_banner.jpg" alt="banner" />
    </div>
  </div>
</template>

<script setup>
  // @ts-ignore
  import { initDB } from '@/api/initdb'
  import { reactive, ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElLoading, ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'

  defineOptions({
    name: 'Init'
  })

  const router = useRouter()
  const { t } = useI18n()

  const page = reactive({
    showReadme: false,
    showForm: false
  })

  const showNext = () => {
    page.showReadme = false
    setTimeout(() => {
      page.showForm = true
    }, 20)
  }

  const goDoc = () => {
    window.open('https://lightningrag.com/guide/start-quickly/env.html')
  }

  const out = ref(false)

  const form = reactive({
    adminPassword: '123456',
    dbType: 'mysql',
    host: '127.0.0.1',
    port: '3306',
    userName: 'root',
    password: '',
    dbName: 'lrag',
    dbPath: ''
  })

  const changeDB = (val) => {
    switch (val) {
      case 'mysql':
        Object.assign(form, {
          adminPassword: '123456',
          reAdminPassword: '',
          dbType: 'mysql',
          host: '127.0.0.1',
          port: '3306',
          userName: 'root',
          password: '',
          dbName: 'lrag',
          dbPath: ''
        })
        break
      case 'pgsql':
        Object.assign(form, {
          adminPassword: '123456',
          dbType: 'pgsql',
          host: '127.0.0.1',
          port: '5432',
          userName: 'postgres',
          password: '',
          dbName: 'lrag',
          dbPath: '',
          template: 'template0'
        })
        break
      case 'oracle':
        Object.assign(form, {
          adminPassword: '123456',
          dbType: 'oracle',
          host: '127.0.0.1',
          port: '1521',
          userName: 'oracle',
          password: '',
          dbName: 'lrag',
          dbPath: ''
        })
        break
      case 'mssql':
        Object.assign(form, {
          adminPassword: '123456',
          dbType: 'mssql',
          host: '127.0.0.1',
          port: '1433',
          userName: 'mssql',
          password: '',
          dbName: 'lrag',
          dbPath: ''
        })
        break
      case 'sqlite':
        Object.assign(form, {
          adminPassword: '123456',
          dbType: 'sqlite',
          host: '',
          port: '',
          userName: '',
          password: '',
          dbName: 'lrag',
          dbPath: ''
        })
        break
      default:
        Object.assign(form, {
          adminPassword: '123456',
          dbType: 'mysql',
          host: '127.0.0.1',
          port: '3306',
          userName: 'root',
          password: '',
          dbName: 'lrag',
          dbPath: ''
        })
    }
  }
  const onSubmit = async () => {
    if (form.adminPassword.length < 6) {
      ElMessage({
        type: 'error',
        message: t('init.passwordTooShort')
      })
      return
    }

    const loading = ElLoading.service({
      lock: true,
      text: t('init.loading'),
      spinner: 'loading',
      background: 'rgba(0, 0, 0, 0.7)'
    })
    try {
      const res = await initDB(form)
      if (res.code === 0) {
        out.value = true
        ElMessage({
          type: 'success',
          message: res.msg
        })
        
        // 显示AI助手配置提示弹窗
        ElMessageBox.confirm(
          t('init.doneBody'),
          t('init.doneTitle'),
          {
            confirmButtonText: t('init.doneConfirm'),
            cancelButtonText: t('init.doneCancel'),
            type: 'success',
            center: true
          }
        ).then(() => {
          // 点击确认按钮，打开AI配置文档
          window.open('https://lightningrag.com/guide/server/mcp.html', '_blank')
          router.push({ name: 'Login' })
        }).catch(() => {
          // 点击取消按钮或关闭弹窗，直接跳转到登录页
          router.push({ name: 'Login' })
        })
      }
      loading.close()
    } catch (_) {
      loading.close()
    }
  }
</script>

<style lang="scss" scoped>
  .slide-in-fwd-top {
    -webkit-animation: slide-in-fwd-top 0.4s
      cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
    animation: slide-in-fwd-top 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
  }
  .slide-out-right {
    -webkit-animation: slide-out-right 0.5s
      cubic-bezier(0.55, 0.085, 0.68, 0.53) both;
    animation: slide-out-right 0.5s cubic-bezier(0.55, 0.085, 0.68, 0.53) both;
  }
  .slide-in-left {
    -webkit-animation: slide-in-left 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94)
      both;
    animation: slide-in-left 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
  }
  @-webkit-keyframes slide-in-fwd-top {
    0% {
      transform: translateZ(-1400px) translateY(-800px);
      opacity: 0;
    }
    100% {
      transform: translateZ(0) translateY(0);
      opacity: 1;
    }
  }
  @keyframes slide-in-fwd-top {
    0% {
      transform: translateZ(-1400px) translateY(-800px);
      opacity: 0;
    }
    100% {
      transform: translateZ(0) translateY(0);
      opacity: 1;
    }
  }
  @-webkit-keyframes slide-out-right {
    0% {
      transform: translateX(0);
      opacity: 1;
    }
    100% {
      transform: translateX(1000px);
      opacity: 0;
    }
  }
  @keyframes slide-out-right {
    0% {
      transform: translateX(0);
      opacity: 1;
    }
    100% {
      transform: translateX(1000px);
      opacity: 0;
    }
  }
  @-webkit-keyframes slide-in-left {
    0% {
      transform: translateX(-1000px);
      opacity: 0;
    }
    100% {
      transform: translateX(0);
      opacity: 1;
    }
  }
  @keyframes slide-in-left {
    0% {
      transform: translateX(-1000px);
      opacity: 0;
    }
    100% {
      transform: translateX(0);
      opacity: 1;
    }
  }
  @media (max-width: 750px) {
    .form {
      width: 94vw !important;
      padding: 0;
    }
  }
</style>
