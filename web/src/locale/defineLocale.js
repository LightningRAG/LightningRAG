import en from './messages/en.js'

function clone(obj) {
  return JSON.parse(JSON.stringify(obj))
}

/**
 * Deep-merge locale overrides onto a clone of English messages.
 * Omitted keys keep English from the base (same behavior as vue-i18n fallback, but per-file).
 */
function deepMerge(target, source) {
  if (!source || typeof source !== 'object') return target
  for (const key of Object.keys(source)) {
    const sv = source[key]
    const tv = target[key]
    if (
      sv &&
      typeof sv === 'object' &&
      !Array.isArray(sv) &&
      tv &&
      typeof tv === 'object' &&
      !Array.isArray(tv)
    ) {
      deepMerge(tv, sv)
    } else {
      target[key] = sv
    }
  }
  return target
}

export function defineLocale(overrides) {
  const base = clone(en)
  return deepMerge(base, overrides || {})
}

/** Merge overrides onto a clone of any base locale (e.g. zh-CN → zh-TW tweaks). */
export function defineLocaleFrom(baseModule, overrides) {
  const base = clone(baseModule?.default ?? baseModule)
  return deepMerge(base, overrides || {})
}
