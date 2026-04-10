/**
 * Ensures every locale message module has the same flattened key set as en.js
 * (after defineLocale merge, partial locales must still match en).
 */
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const messagesDir = path.join(__dirname, '../src/locale/messages')

const en = (await import(path.join(messagesDir, 'en.js'))).default
const zhCN = (await import(path.join(messagesDir, 'zh-CN.js'))).default
const zhTW = (await import(path.join(messagesDir, 'zh-TW.js'))).default
const ja = (await import(path.join(messagesDir, 'ja.js'))).default
const ko = (await import(path.join(messagesDir, 'ko.js'))).default
const fr = (await import(path.join(messagesDir, 'fr.js'))).default
const de = (await import(path.join(messagesDir, 'de.js'))).default
const es = (await import(path.join(messagesDir, 'es.js'))).default
const it = (await import(path.join(messagesDir, 'it.js'))).default
const ptBR = (await import(path.join(messagesDir, 'pt-BR.js'))).default
const ru = (await import(path.join(messagesDir, 'ru.js'))).default
const vi = (await import(path.join(messagesDir, 'vi.js'))).default
const th = (await import(path.join(messagesDir, 'th.js'))).default
const id = (await import(path.join(messagesDir, 'id.js'))).default

function flattenKeys(obj, prefix = '') {
  const keys = []
  if (obj === null || obj === undefined) return keys
  if (typeof obj !== 'object' || Array.isArray(obj)) {
    keys.push(prefix)
    return keys
  }
  for (const k of Object.keys(obj)) {
    const p = prefix ? `${prefix}.${k}` : k
    keys.push(...flattenKeys(obj[k], p))
  }
  return keys
}

const modules = {
  en,
  'zh-CN': zhCN,
  'zh-TW': zhTW,
  ja,
  ko,
  fr,
  de,
  es,
  it,
  'pt-BR': ptBR,
  ru,
  vi,
  th,
  id
}

const base = new Set(flattenKeys(en))
let failed = false

for (const [code, mod] of Object.entries(modules)) {
  if (code === 'en') continue
  const keys = new Set(flattenKeys(mod))
  const missing = [...base].filter((k) => !keys.has(k))
  const extra = [...keys].filter((k) => !base.has(k))
  if (missing.length || extra.length) {
    failed = true
    console.error(`[locale:check] ${code}: missing ${missing.length}, extra ${extra.length}`)
    if (missing.length) console.error('  missing (first 20):', missing.slice(0, 20).join(', '))
    if (extra.length) console.error('  extra (first 20):', extra.slice(0, 20).join(', '))
  }
}

if (failed) {
  console.error('[locale:check] failed')
  process.exit(1)
}

console.log(`[locale:check] OK — ${base.size} keys, all locales match en.`)
