import { createI18n } from 'vue-i18n'
import {
  DEFAULT_LOCALE,
  SUPPORTED_LOCALES,
  STORAGE_KEY,
  LOCALE_ACCEPT_LANGUAGE
} from './constants.js'
import { getInitialLocale } from './resolve.js'
import en from './messages/en.js'
import zhCN from './messages/zh-CN.js'
import zhTW from './messages/zh-TW.js'
import ja from './messages/ja.js'
import ko from './messages/ko.js'
import fr from './messages/fr.js'
import de from './messages/de.js'
import es from './messages/es.js'
import it from './messages/it.js'
import ptBR from './messages/pt-BR.js'
import ru from './messages/ru.js'
import vi from './messages/vi.js'
import th from './messages/th.js'
import id from './messages/id.js'

const initialLocale = getInitialLocale()

export const i18n = createI18n({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: DEFAULT_LOCALE,
  fallbackWarn: false,
  missingWarn: false,
  globalInjection: true,
  messages: {
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
})

/**
 * @param {string} code
 * @param {boolean} persistUserChoice - false when only syncing document (e.g. tests)
 */
export function setLocale(code, persistUserChoice = true) {
  if (!SUPPORTED_LOCALES.includes(code)) return
  i18n.global.locale.value = code
  if (typeof document !== 'undefined') {
    document.documentElement.lang = validatorHtmlLangFromLocale(code)
  }
  if (persistUserChoice) {
    try {
      localStorage.setItem(STORAGE_KEY, code)
    } catch {
      /* ignore */
    }
  }
}

export function validatorHtmlLangFromLocale(code) {
  if (code === 'zh-CN') return 'zh-CN'
  if (code === 'zh-TW') return 'zh-TW'
  if (code === 'ja') return 'ja'
  if (code === 'ko') return 'ko'
  if (code === 'pt-BR') return 'pt-BR'
  if (
    code === 'en' ||
    code === 'fr' ||
    code === 'de' ||
    code === 'es' ||
    code === 'it' ||
    code === 'ru' ||
    code === 'vi' ||
    code === 'th' ||
    code === 'id'
  ) {
    return code
  }
  return 'en'
}

export { getInitialLocale, resolveBrowserLocale, normalizeLocale } from './resolve.js'
export {
  SUPPORTED_LOCALES,
  DEFAULT_LOCALE,
  STORAGE_KEY,
  LOCALE_ACCEPT_LANGUAGE
} from './constants.js'

/** Value for the Accept-Language HTTP header from the active UI locale */
export function getAcceptLanguageHeader() {
  try {
    const code = i18n.global.locale.value
    return LOCALE_ACCEPT_LANGUAGE[code] || LOCALE_ACCEPT_LANGUAGE.en
  } catch {
    return LOCALE_ACCEPT_LANGUAGE.en
  }
}

/** Same locale headers as axios interceptors — use on raw fetch() so API messages match UI language */
export function getLocaleHeaders() {
  try {
    return {
      'X-Locale': i18n.global.locale.value,
      'Accept-Language': getAcceptLanguageHeader()
    }
  } catch {
    return {
      'X-Locale': DEFAULT_LOCALE,
      'Accept-Language': LOCALE_ACCEPT_LANGUAGE.en
    }
  }
}
