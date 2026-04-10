/**
 * 网站配置文件
 */
import packageInfo from '../../package.json'

const greenText = (text) => `\x1b[32m${text}\x1b[0m`

export const config = {
  appName: 'LightningRAG',
  showViteLogo: true,
  keepAliveTabs: false,
  logs: []
}

export const viteLogo = (env) => {
  if (config.showViteLogo) {
    console.log(
      greenText(
        `> 欢迎使用 LightningRAG，开源地址：https://github.com/LightningRAG`
      )
    )
    console.log(greenText(`> 当前版本:v${packageInfo.version}`))
    console.log(
      greenText(`> 官网：https://lightningrag.com`)
    )
    console.log(
      greenText(
        `> 默认自动化文档地址:http://127.0.0.1:${env.VITE_SERVER_PORT}/swagger/index.html`
      )
    )
    console.log(
      greenText(`> 默认前端文件运行地址:http://127.0.0.1:${env.VITE_CLI_PORT}`)
    )
    console.log('\n')
  }
}

export default config
