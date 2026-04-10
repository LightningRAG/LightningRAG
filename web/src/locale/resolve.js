import { DEFAULT_LOCALE, SUPPORTED_LOCALES, STORAGE_KEY } from './constants.js'

/**
 * Map BCP-47 / navigator value to one of SUPPORTED_LOCALES, or null if no match.
 */
export function normalizeLocale(raw) {
  if (!raw || typeof raw !== 'string') return null
  const tag = raw.trim().toLowerCase().replace(/_/g, '-')

  if (tag === 'zh-cn' || tag === 'zh-hans' || tag === 'zh' || tag.startsWith('zh-hans')) {
    return 'zh-CN'
  }
  if (
    tag === 'zh-tw' ||
    tag === 'zh-hk' ||
    tag === 'zh-mo' ||
    tag.startsWith('zh-hant')
  ) {
    return 'zh-TW'
  }
  if (tag.startsWith('ja')) return 'ja'
  if (tag.startsWith('ko')) return 'ko'
  if (tag.startsWith('en')) return 'en'
  if (tag.startsWith('fr')) return 'fr'
  if (tag.startsWith('de')) return 'de'
  if (tag.startsWith('es')) return 'es'
  if (tag.startsWith('it')) return 'it'
  if (tag === 'pt-br' || tag.startsWith('pt-br')) return 'pt-BR'
  if (tag.startsWith('ru')) return 'ru'
  if (tag.startsWith('vi')) return 'vi'
  if (tag.startsWith('th')) return 'th'
  if (tag.startsWith('id')) return 'id'

  return null
}

export function resolveBrowserLocale() {
  const candidates = []
  if (typeof navigator !== 'undefined') {
    if (navigator.languages?.length) {
      candidates.push(...navigator.languages)
    }
    if (navigator.language) {
      candidates.push(navigator.language)
    }
  }
  for (const raw of candidates) {
    const n = normalizeLocale(raw)
    if (n && SUPPORTED_LOCALES.includes(n)) return n
  }
  return DEFAULT_LOCALE
}

export function getInitialLocale() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved && SUPPORTED_LOCALES.includes(saved)) {
      return saved
    }
  } catch {
    /* ignore */
  }
  // No saved choice: follow browser / OS language; unmatched tags fall back via resolveBrowserLocale.
  return resolveBrowserLocale()
}
