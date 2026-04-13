/**
 * Translate API description & group name based on method+path.
 * Key format: `{method}_{pathSegments}` (lowercase method, slashes → underscores, leading slash removed).
 * Example: POST /api/createApi → `post_api_createApi`
 */

function buildApiSummaryKey(method, path) {
  if (!method || !path) return ''
  const cleanPath = path.replace(/^\//, '').replace(/\//g, '_')
  return `${method.toLowerCase()}_${cleanPath}`
}

/**
 * @param {{ method: string, path: string, description?: string }} api
 * @param {(key: string) => string} t - vue-i18n t()
 * @returns {string} translated description, fallback to original
 */
export function translateApiDescription(api, t) {
  if (!api) return ''
  const key = buildApiSummaryKey(api.method, api.path)
  if (!key) return api.description || ''
  const fullKey = `apiSummary.${key}`
  const translated = t(fullKey)
  if (translated !== fullKey) return translated
  return api.description || ''
}

/**
 * @param {string} groupName - raw API group name from backend
 * @param {(key: string) => string} t - vue-i18n t()
 * @returns {string} translated group name, fallback to original
 */
export function translateApiGroup(groupName, t) {
  if (!groupName) return ''
  const fullKey = `apiGroup.${groupName}`
  const translated = t(fullKey)
  if (translated !== fullKey) return translated
  return groupName
}
