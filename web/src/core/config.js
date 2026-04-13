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
        `> Welcome to LightningRAG — https://github.com/LightningRAG`
      )
    )
    console.log(greenText(`> Version: v${packageInfo.version}`))
    console.log(
      greenText(`> Website: https://lightningrag.com`)
    )
    console.log(
      greenText(
        `> Swagger docs: http://127.0.0.1:${env.VITE_SERVER_PORT}/swagger/index.html`
      )
    )
    console.log(
      greenText(`> Frontend: http://127.0.0.1:${env.VITE_CLI_PORT}`)
    )
    console.log('\n')
  }
}

export default config
