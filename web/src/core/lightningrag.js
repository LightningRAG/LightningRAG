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
       欢迎使用 LightningRAG
       当前版本:v${packageInfo.version}
       项目地址：https://github.com/LightningRAG/LightningRAG
       默认自动化文档地址:http://127.0.0.1:${import.meta.env.VITE_SERVER_PORT}/swagger/index.html
       默认前端文件运行地址:http://127.0.0.1:${import.meta.env.VITE_CLI_PORT}
       --------------------------------------版权声明--------------------------------------
       ** 版权所有方：LightningRAG 开源团队 **
       ** 版权持有公司：合肥云亿连智能科技有限公司 **
       ** 剔除授权标识需购买商用授权：https://plugin.LightningRAG.com/license.html **
       ** 感谢您对 LightningRAG 的支持与关注 **
    `)
  }
}
