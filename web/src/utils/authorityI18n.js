/**
 * Built-in roles from server seed (server/source/system/authority.go).
 * UI shows i18n strings by authorityId so names follow the active locale.
 */
export const BUILTIN_AUTHORITY_DISPLAY_KEYS = {
  888: 'admin.authority.builtinDefaultUser',
  9528: 'admin.authority.builtinTestRole',
  8881: 'admin.authority.builtinDefaultUserSubRole'
}

/**
 * @param {{ authorityId?: number|string, authorityName?: string }}|null|undefined} authority
 * @param {(key: string) => string} t - vue-i18n t()
 */
export function authorityDisplayName(authority, t) {
  if (!authority) return ''
  const id = Number(authority.authorityId)
  const key = BUILTIN_AUTHORITY_DISPLAY_KEYS[id]
  const raw = authority.authorityName ?? ''
  if (key) return t(key)
  return raw
}
