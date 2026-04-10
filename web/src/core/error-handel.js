import { createSysError } from '@/api/system/sysError'

/** 入库使用英文标签，Error log 页按界面语言展示（见 sysErrorDisplay.js） */
function sendErrorTip(errorInfo) {
  setTimeout(() => {
    const lines = []
    lines.push(`Message: ${errorInfo.message}`)
    lines.push(`Stack: ${errorInfo.stack || 'No stack trace available.'}`)
    if (errorInfo.component) {
      lines.push(`Component: ${errorInfo.component.name || 'Unknown'}`)
    }
    if (errorInfo.vueInfo) {
      lines.push(`Vue Info: ${errorInfo.vueInfo}`)
    }
    if (errorInfo.source) {
      lines.push(
        `Source: ${errorInfo.source}:${errorInfo.lineno}:${errorInfo.colno}`
      )
    }
    const errorData = {
      form: 'frontend',
      info: lines.join('\n'),
      level: 'error',
      solution: null
    }

    createSysError(errorData).catch((apiErr) => {
      console.error('Failed to create error record:', apiErr)
    })
  }, 0)
}

window.addEventListener('unhandledrejection', (event) => {
  const reason = event.reason
  sendErrorTip({
    message: String(reason != null ? reason : ''),
    stack: reason?.stack || ''
  })
})
