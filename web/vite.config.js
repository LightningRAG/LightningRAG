import legacyPlugin from '@vitejs/plugin-legacy'
import { viteLogo } from './src/core/config'
import Banner from 'vite-plugin-banner'
import { createRequire } from 'node:module'
import * as path from 'path'
import { loadEnv } from 'vite'
import vuePlugin from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import VueFilePathPlugin from './vitePlugin/componentName/index.js'
import vueRootValidator from 'vite-check-multiple-dom'

/** 本地 CJS 插件若用 ESM import，会被 Vite 打包 config 时内联并触发 Dynamic require of "fs" */
const nodeRequire = createRequire(import.meta.url)
const { svgBuilder } = nodeRequire(
  path.resolve(process.cwd(), 'vitePlugin/vite-auto-import-svg/index.min.js')
)
import { AddSecret } from './vitePlugin/secret'
import UnoCSS from '@unocss/vite'

// @see https://cn.vitejs.dev/config/
export default ({ mode }) => {
  AddSecret('')
  const env = loadEnv(mode, process.cwd())
  viteLogo(env)

  /** Docker / 低配 CI 下 terser 易导致 OOM(137)，改用 esbuild 并限流 Rollup */
  const dockerBuild = process.env.DOCKER_BUILD === '1'

  const timestamp = Date.parse(new Date())

  // @form-create/designer 走源码时会直接 import 若干 CJS 包，需预构建才有 default 导出
  const optimizeDeps = {
    include: [
      'codemirror/lib/codemirror',
      'codemirror/mode/javascript/javascript',
      'codemirror/mode/xml/xml',
      'codemirror/addon/hint/show-hint',
      'codemirror/addon/hint/javascript-hint',
      'js-beautify'
    ]
  }

  const alias = {
    '@': path.resolve(__dirname, './src'),
    vue$: 'vue/dist/vue.runtime.esm-bundler.js',
    // 指向包根（勿指到 src/index.js，否则 @form-create/designer/src/locale/... 会错解成 index.js/src/...）
    '@form-create/designer': path.resolve(
      __dirname,
      'node_modules/@form-create/designer'
    )
  }

  const esbuild =
    dockerBuild && mode === 'production'
      ? { drop: ['console', 'debugger'] }
      : {}

  const rollupOptions = {
    ...(dockerBuild ? { maxParallelFileOps: 1 } : {}),
    output: {
      entryFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
      chunkFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
      assetFileNames: 'assets/087AC4D233B64EB0[name].[hash].[ext]'
    }
  }

  const base = '/'
  const root = './'
  const outDir = 'dist'

  const config = {
    base: base, // 编译后js导入的资源路径
    root: root, // index.html文件所在位置
    publicDir: 'public', // 静态资源文件夹
    resolve: {
      alias
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler' // or "modern"
        }
      }
    },
    server: {
      // 如果使用docker-compose开发模式，设置为false
      open: true,
      port: Number(env.VITE_CLI_PORT),
      proxy: {
        // 把key的路径代理到target位置
        // detail: https://cli.vuejs.org/config/#devserver-proxy
        [env.VITE_BASE_API]: {
          // 需要代理的路径   例如 '/api'
          target: `${env.VITE_BASE_PATH}:${env.VITE_SERVER_PORT}/`, // 代理到 目标路径
          changeOrigin: true,
          rewrite: (path) =>
            path.replace(new RegExp('^' + env.VITE_BASE_API), '')
        },
        '/plugin': {
          // 需要代理的路径，插件市场 API。可通过 VITE_PLUGIN_API 环境变量覆盖
          target: env.VITE_PLUGIN_API || 'https://plugin.lightningrag.com/api/',
          changeOrigin: true,
          rewrite: (path) =>
            path.replace(new RegExp('^/plugin'), '')
        }
      }
    },
    build: {
      minify: dockerBuild ? 'esbuild' : 'terser',
      manifest: false, // 是否产出manifest.json
      sourcemap: false, // 是否产出sourcemap.json
      outDir: outDir, // 产出目录
      commonjsOptions: {
        transformMixedEsModules: true
      },
      ...(dockerBuild
        ? {}
        : {
            terserOptions: {
              compress: {
                drop_console: true,
                drop_debugger: true
              }
            }
          }),
      rollupOptions
    },
    esbuild,
    optimizeDeps,
    plugins: [
      env.VITE_POSITION === 'open' &&
      vueDevTools({ launchEditor: env.VITE_EDITOR }),
      legacyPlugin({
        targets: [
          'Android > 39',
          'Chrome >= 60',
          'Safari >= 10.1',
          'iOS >= 10.3',
          'Firefox >= 54',
          'Edge >= 15'
        ]
      }),
      vuePlugin(),
      svgBuilder(['./src/plugin/', './src/assets/icons/'], base, outDir, 'assets', mode),
      [Banner(`\n Build based on LightningRAG \n Time : ${timestamp}`)],
      VueFilePathPlugin('./src/pathInfo.json'),
      UnoCSS(),
      vueRootValidator()
    ]
  }
  return config
}
