import { i18n } from '@/locale'
import { fmtTitle } from '@/utils/fmtRouterTitle'

/**
 * 侧栏 / Tab / 面包屑等展示用：若 meta.titleKey 存在则用 vue-i18n，否则用 meta.title；再套用 ${param} 替换。
 * 与后端一致：英文缺省标题见 server/model/system/menu_default_title_en.go（DefaultMenuTitleEnglish）与 en.js → menu.names。
 * @param {Record<string, unknown>|undefined|null} meta
 * @param {{ params?: object, query?: object }} routeLike
 */
export function resolveMenuTitle(meta, routeLike) {
  if (!meta) return ''
  const key = meta.titleKey
  const base =
    typeof key === 'string' && key.length > 0
      ? i18n.global.t(key)
      : String(meta.title ?? '')
  return fmtTitle(base, routeLike || {})
}
