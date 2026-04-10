export const DEFAULT_LOCALE = 'en'
export const SUPPORTED_LOCALES = [
  'en',
  'zh-CN',
  'zh-TW',
  'ja',
  'ko',
  'fr',
  'de',
  'es',
  'it',
  'pt-BR',
  'ru',
  'vi',
  'th',
  'id'
]
export const STORAGE_KEY = 'lightningrag-locale'

/** Primary Accept-Language values per app locale (for API / future server i18n) */
export const LOCALE_ACCEPT_LANGUAGE = {
  en: 'en-US,en;q=0.9',
  'zh-CN': 'zh-CN,zh;q=0.9,en;q=0.8',
  'zh-TW': 'zh-TW,zh-Hant;q=0.9,en;q=0.8',
  ja: 'ja,en;q=0.8',
  ko: 'ko,en;q=0.8',
  fr: 'fr,en;q=0.8',
  de: 'de,en;q=0.8',
  es: 'es,en;q=0.8',
  it: 'it,en;q=0.8',
  'pt-BR': 'pt-BR,pt;q=0.9,en;q=0.8',
  ru: 'ru,en;q=0.8',
  vi: 'vi,en;q=0.8',
  th: 'th,en;q=0.8',
  id: 'id,en;q=0.8'
}
