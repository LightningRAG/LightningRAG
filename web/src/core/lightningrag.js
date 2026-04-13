/*
 * LightningRAG web 框架组
 *
 * */
// 加载网站配置文件夹
import { register } from './global'
import packageInfo from '../../package.json'

export default {
  install: (app) => {
    register(app)
    console.log(`
       Welcome to LightningRAG
       Version: v${packageInfo.version}
       Repository: https://github.com/LightningRAG/LightningRAG
       Swagger docs: http://127.0.0.1:${import.meta.env.VITE_SERVER_PORT}/swagger/index.html
       Frontend: http://127.0.0.1:${import.meta.env.VITE_CLI_PORT}
    `)
  }
}
