/** 与后端各 LLM（OpenAI/Azure/Ollama/Anthropic/Bedrock/Cohere 等）注入的 think 标签对一致（避免字面量被工具误替换，使用拼接） */
const MODEL_THINK_OPEN = '<' + 'think' + '>'
const MODEL_THINK_CLOSE = '<' + '/' + 'think' + '>'

export function stripAllThinkingBlocks (text) {
  if (!text) return ''
  let out = text
  for (;;) {
    const i0 = out.indexOf(MODEL_THINK_OPEN)
    if (i0 === -1) break
    const i1 = out.indexOf(MODEL_THINK_CLOSE, i0 + MODEL_THINK_OPEN.length)
    if (i1 === -1) break
    out = out.slice(0, i0) + out.slice(i1 + MODEL_THINK_CLOSE.length)
  }
  return out
}

/** @returns {{ think: string, main: string, streaming: boolean }} */
export function splitModelThinking (text) {
  if (!text) return { think: '', main: '', streaming: false }
  const i0 = text.indexOf(MODEL_THINK_OPEN)
  if (i0 === -1) return { think: '', main: text, streaming: false }
  const afterOpen = i0 + MODEL_THINK_OPEN.length
  const i1 = text.indexOf(MODEL_THINK_CLOSE, afterOpen)
  if (i1 === -1) {
    return {
      think: text.slice(afterOpen),
      main: text.slice(0, i0),
      streaming: true
    }
  }
  const think = text.slice(afterOpen, i1)
  let main = text.slice(i1 + MODEL_THINK_CLOSE.length).replace(/^\s+/, '')
  main = stripAllThinkingBlocks(main)
  return { think, main, streaming: false }
}
