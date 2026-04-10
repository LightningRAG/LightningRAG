import { resolveMenuTitle } from '@/utils/menuTitle'
import config from '@/core/config'

export default function getPageTitle(route) {
  const title = resolveMenuTitle(route?.meta, route)
  if (title) {
    return `${title} - ${config.appName}`
  }
  return `${config.appName}`
}
