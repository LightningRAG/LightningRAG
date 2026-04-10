/**
 * Regenerate zh-TW locale from zh-CN using OpenCC (cn → tw).
 * Run from repo root: npm run locale:zh-tw (in web/)
 */
import * as OpenCC from 'opencc-js'
import { writeFileSync } from 'fs'
import { fileURLToPath, pathToFileURL } from 'url'
import { dirname, join } from 'path'

const __dirname = dirname(fileURLToPath(import.meta.url))
const root = join(__dirname, '..')

const zhCNModUrl = pathToFileURL(join(root, 'src/locale/messages/zh-CN.js')).href
const zhCN = (await import(zhCNModUrl)).default
const converter = OpenCC.Converter({ from: 'cn', to: 'tw' })

function convertDeep(value) {
  if (value === null || value === undefined) return value
  if (typeof value === 'string') return converter(value)
  if (Array.isArray(value)) return value.map(convertDeep)
  if (typeof value === 'object') {
    const out = {}
    for (const key of Object.keys(value)) {
      out[key] = convertDeep(value[key])
    }
    return out
  }
  return value
}

const zhTW = convertDeep(zhCN)
const outPath = join(root, 'src/locale/messages/zh-TW.js')
const banner =
  '// Auto-generated from zh-CN.js — do not edit by hand; run: npm run locale:zh-tw\n\n'
writeFileSync(outPath, `${banner}export default ${JSON.stringify(zhTW, null, 2)}\n`, 'utf8')
console.log('Wrote', outPath)
