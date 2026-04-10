import './style/element_visiable.scss'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'uno.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import { setupVueRootValidator } from 'vite-check-multiple-dom/client'

import 'element-plus/dist/index.css'
// 引入 LightningRAG 前端初始化相关内容
import router from '@/router/index'
import '@/permission'
import run from '@/core/lightningrag.js'
import auth from '@/directive/auth'
import clickOutSide from '@/directive/clickOutSide'
import { store } from '@/pinia'
import App from './App.vue'
import '@/core/error-handel'
import { i18n, validatorHtmlLangFromLocale, getInitialLocale } from '@/locale'

const app = createApp(App)

app.config.productionTip = false

const initialLocale = getInitialLocale()
setupVueRootValidator(app, {
  lang: validatorHtmlLangFromLocale(initialLocale)
})
if (typeof document !== 'undefined') {
  document.documentElement.lang = validatorHtmlLangFromLocale(initialLocale)
}

app
  .use(i18n)
  .use(run)
  .use(ElementPlus)
  .use(store)
  .use(auth)
  .use(clickOutSide)
  .use(router)
  .mount('#app')
export default app
