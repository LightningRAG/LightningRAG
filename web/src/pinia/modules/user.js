import { login, getUserInfo } from '@/api/user'
import { jsonInBlacklist } from '@/api/jwt'
import router from '@/router/index'
import { ElLoading, ElMessage } from 'element-plus'
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouterStore } from './router'
import { useCookies } from '@vueuse/integrations/useCookies'
import { useStorage } from '@vueuse/core'

import { useAppStore } from '@/pinia'
import { i18n } from '@/locale'

export const useUserStore = defineStore('user', () => {
  const appStore = useAppStore()
  const loadingInstance = ref(null)

  const userInfo = ref({
    uuid: '',
    nickName: '',
    headerImg: '',
    authority: {}
  })
  const token = useStorage('token', '')
  const xToken = useCookies()
  const currentToken = computed(() => token.value || xToken.get('x-token') || '')

  const setUserInfo = (val) => {
    userInfo.value = val
    if (val.originSetting) {
      Object.keys(appStore.config).forEach((key) => {
        if (val.originSetting[key] !== undefined) {
          appStore.config[key] = val.originSetting[key]
        }
      })
    }
  }

  const setToken = (val) => {
    token.value = val
    xToken.value = val
  }

  const NeedInit = async () => {
    await ClearStorage()
    await router.push({ name: 'Init', replace: true })
  }

  const ResetUserInfo = (value = {}) => {
    userInfo.value = {
      ...userInfo.value,
      ...value
    }
  }
  /* 获取用户信息*/
  const GetUserInfo = async () => {
    const res = await getUserInfo()
    if (res.code === 0) {
      setUserInfo(res.data.userInfo)
    }
    return res
  }
  const finishLoginNavigation = async (oauthRedirectPath) => {
    const routerStore = useRouterStore()
    await routerStore.SetAsyncRouter()
    const asyncRouters = routerStore.asyncRouters
    asyncRouters.forEach((asyncRouter) => {
      router.addRoute(asyncRouter)
    })

    if (
      oauthRedirectPath &&
      typeof oauthRedirectPath === 'string' &&
      oauthRedirectPath.startsWith('/')
    ) {
      await router.replace(oauthRedirectPath)
    } else if (router.currentRoute.value.query.redirect) {
      await router.replace(router.currentRoute.value.query.redirect)
    } else if (!router.hasRoute(userInfo.value.authority.defaultRouter)) {
      ElMessage.error(i18n.global.t('common.user.noHomeConfigured'))
    } else {
      await router.replace({ name: userInfo.value.authority.defaultRouter })
    }

    const isWindows = /windows/i.test(navigator.userAgent)
    window.localStorage.setItem('osType', isWindows ? 'WIN' : 'MAC')
    return true
  }

  const LoginIn = async (loginInfo) => {
    try {
      loadingInstance.value = ElLoading.service({
        fullscreen: true,
        text: i18n.global.t('common.user.loggingIn')
      })

      const res = await login(loginInfo)

      if (res.code !== 0) {
        return false
      }
      setUserInfo(res.data.user)
      setToken(res.data.token)
      await finishLoginNavigation()
      return true
    } catch (error) {
      console.error('LoginIn error:', error)
      return false
    } finally {
      loadingInstance.value?.close()
    }
  }

  const LoginInWithOAuthPayload = async (payload) => {
    try {
      loadingInstance.value = ElLoading.service({
        fullscreen: true,
        text: i18n.global.t('common.user.loggingIn')
      })
      if (!payload?.user || !payload?.token) {
        return false
      }
      setUserInfo(payload.user)
      setToken(payload.token)
      await finishLoginNavigation(payload.redirect)
      return true
    } catch (error) {
      console.error('LoginInWithOAuthPayload error:', error)
      return false
    } finally {
      loadingInstance.value?.close()
    }
  }
  /* 登出*/
  const LoginOut = async () => {
    const res = await jsonInBlacklist()

    // 登出失败
    if (res.code !== 0) {
      return
    }

    await ClearStorage()

    // 把路由定向到登录页，无需等待直接reload
    router.push({ name: 'Login', replace: true })
    window.location.reload()
  }
  /* 清理数据 */
  const ClearStorage = async () => {
    token.value = ''
    // 使用remove方法正确删除cookie
    xToken.remove()
    sessionStorage.clear()
    // 清理所有相关的localStorage项
    localStorage.removeItem('originSetting')
    localStorage.removeItem('token')
  }

  return {
    userInfo,
    token: currentToken,
    NeedInit,
    ResetUserInfo,
    GetUserInfo,
    LoginIn,
    LoginInWithOAuthPayload,
    LoginOut,
    setToken,
    loadingInstance,
    ClearStorage
  }
})
