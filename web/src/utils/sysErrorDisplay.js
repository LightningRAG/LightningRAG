/**
 * 错误日志 form / info：库中可能为中文、英文或稳定代码（frontend / backend）。
 * 列表与详情按当前 vue-i18n 展示，英文文案以 en 为准。
 */

const FORM_TO_I18N = {
  frontend: 'sysError.sourceFrontend',
  client: 'sysError.sourceFrontend',
  前端: 'sysError.sourceFrontend',
  backend: 'sysError.sourceBackend',
  后端: 'sysError.sourceBackend',
  後端: 'sysError.sourceBackend',
  Backend: 'sysError.sourceBackend'
}

/**
 * @param {string|null|undefined} form
 * @param {(k: string) => string} t
 */
export function formatSysErrorForm(form, t) {
  if (form == null || form === '') return ''
  const s = String(form).trim()
  const path = FORM_TO_I18N[s]
  if (path) {
    const out = t(path)
    return out === path ? s : out
  }
  return s
}

function localizeInfoLine(line, t) {
  const rules = [
    { prefix: '错误信息:', key: 'sysError.logPrefixMessage' },
    { prefix: '调用栈：', key: 'sysError.logPrefixStack' },
    { prefix: '调用栈:', key: 'sysError.logPrefixStack' },
    { prefix: 'Message:', key: 'sysError.logPrefixMessage' },
    { prefix: 'Stack:', key: 'sysError.logPrefixStack' },
    { prefix: 'Component:', key: 'sysError.logPrefixComponent' },
    { prefix: 'Vue Info:', key: 'sysError.logPrefixVueInfo' },
    { prefix: 'Source:', key: 'sysError.logPrefixSource' },
    { prefix: ' 源文件:', key: 'sysError.logPrefixSourceFile' },
    { prefix: '源文件:', key: 'sysError.logPrefixSourceFile' },
    { prefix: 'Source file:', key: 'sysError.logPrefixSourceFile' },
    { prefix: '最终调用方法:', key: 'sysError.logPrefixFinalCaller' },
    { prefix: 'Final caller:', key: 'sysError.logPrefixFinalCaller' },
    { prefix: 'Panic:', key: 'sysError.logPrefixPanic' },
    { prefix: 'Request:', key: 'sysError.logPrefixRequest' }
  ]

  const trimmedStart = line.trimStart()
  for (const r of rules) {
    if (trimmedStart.startsWith(r.prefix)) {
      const indent = line.slice(0, line.length - trimmedStart.length)
      const rest = trimmedStart.slice(r.prefix.length)
      return `${indent}${t(r.key)}${rest}`
    }
  }

  let out = line
  if (out.includes('| 错误：')) {
    out = out.split('| 错误：').join(t('sysError.logInlineError'))
  }
  if (out.includes('| 错误:')) {
    out = out.split('| 错误:').join(t('sysError.logInlineError'))
  }
  if (out.includes('| Error:')) {
    out = out.split('| Error:').join(t('sysError.logInlineError'))
  }

  const cnTitle = '----- 产生日志的方法代码如下 -----'
  const enTitle = '----- Source excerpt -----'
  if (out.includes(cnTitle)) {
    out = out.split(cnTitle).join(t('sysError.logCodeBlockTitle'))
  }
  if (out.includes(enTitle)) {
    out = out.split(enTitle).join(t('sysError.logCodeBlockTitle'))
  }

  if (out.trim() === '没有调用栈信息') {
    return t('sysError.noStackTrace')
  }

  return out
}

/**
 * @param {string|null|undefined} info
 * @param {(k: string) => string} t
 */
export function formatSysErrorInfo(info, t) {
  if (info == null) return ''
  return String(info)
    .split('\n')
    .map((line) => localizeInfoLine(line, t))
    .join('\n')
}
