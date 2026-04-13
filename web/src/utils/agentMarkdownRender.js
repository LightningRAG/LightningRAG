import { Marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js'
import DOMPurify from 'dompurify'
import { i18n } from '@/locale'
import { splitModelThinking } from '@/utils/ragThinking'

const marked = new Marked(
  markedHighlight({
    langPrefix: 'hljs language-',
    highlight(code, lang) {
      const language = hljs.getLanguage(lang) ? lang : 'plaintext'
      return hljs.highlight(code, { language }).value
    }
  })
)

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

/** 与对话页折叠标题一致（rag.conv.reasoningBlock / reasoningStreaming） */
function reasoningDetailsTitle(streaming) {
  try {
    const t = i18n.global.t
    return streaming ? t('rag.conv.reasoningStreaming') : t('rag.conv.reasoningBlock')
  } catch {
    return streaming ? 'Reasoning (streaming…)' : 'Reasoning'
  }
}

/** Agent 运行调试输出：与对话页一致，将 think 推理块用 details 展示，正文再走 marked */
export function renderAgentMarkdown(text) {
  if (text == null || !String(text).trim()) return ''
  const raw = String(text)
  const { think, main, streaming } = splitModelThinking(raw)
  let prefix = ''
  if (think) {
    const title = reasoningDetailsTitle(streaming)
    prefix =
      `<details class="agent-thinking-block mb-2 rounded border border-slate-200 dark:border-slate-600 bg-slate-50/80 dark:bg-slate-900/40"` +
      (streaming ? ' open' : '') +
      `><summary class="px-2 py-1.5 cursor-pointer text-xs font-medium text-slate-600 dark:text-slate-300 select-none list-none marker:content-none">${escapeHtml(title)}</summary>` +
      `<div class="px-2 pb-2 text-xs text-slate-600 dark:text-slate-400 whitespace-pre-wrap max-h-56 overflow-y-auto border-t border-slate-200/80 dark:border-slate-600/80">${escapeHtml(think)}</div></details>`
  }
  if (!main.trim()) return prefix || escapeHtml(raw)
  try {
    return prefix + DOMPurify.sanitize(marked.parse(main))
  } catch {
    return prefix + escapeHtml(main)
  }
}
