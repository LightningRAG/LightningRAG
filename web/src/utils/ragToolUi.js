/**
 * RAG 对话工具列表（listTools）的展示名/描述：优先 vue-i18n rag.tools.*，否则回退接口字段。
 */
export function formatRagToolDisplayName(tool, t) {
  if (!tool) return ''
  const name = tool.name
  if (!name) return tool.displayName || ''
  const k = `rag.tools.names.${name}`
  const tr = t(k)
  if (tr !== k) return tr
  return tool.displayName || name
}

export function formatRagToolDescription(tool, t) {
  if (!tool) return ''
  const name = tool.name
  if (!name) return tool.description || ''
  const k = `rag.tools.desc.${name}`
  const tr = t(k)
  if (tr !== k) return tr
  return tool.description || ''
}

/** 流式工具调用条：仅名称，无 tool 对象时用 */
export function formatRagToolNameOnly(name, apiDisplayName, t) {
  if (!name) return ''
  const k = `rag.tools.names.${name}`
  const tr = t(k)
  if (tr !== k) return tr
  if (apiDisplayName) return apiDisplayName
  return name
}
