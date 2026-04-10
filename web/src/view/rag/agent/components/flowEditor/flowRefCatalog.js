/**
 * 与 server/agent/component 各组件 output 字段对齐，供编排侧栏「插入变量」列举可选引用。
 * VariableAggregator 另有各 group_name 动态键，此处仅列出固定输出。
 */
export const COMPONENT_OUTPUT_KEYS = {
  Begin: [],
  Retrieval: ['formalized_content'],
  LLM: ['content'],
  Agent: ['content'],
  Message: ['content'],
  Switch: ['next_id', 'selected_case'],
  Categorize: ['selected_category', 'next_id'],
  HTTPRequest: ['result', 'status_code'],
  Iteration: ['formalized_content', 'next_id'],
  TextProcessing: ['content', 'result'],
  ExecuteSQL: ['formalized_content', 'json', 'json_string'],
  DocsGenerator: ['file_path', 'filename', 'success', 'download', 'pdf_base64'],
  MCP: ['result', 'raw_content'],
  SetVariable: ['formalized_content', 'summary'],
  Transformer: ['content', 'formalized_content'],
  AwaitResponse: ['user_input', 'formalized_content', 'content'],
  DuckDuckGo: ['formalized_content', 'json'],
  Wikipedia: ['formalized_content', 'json'],
  ArXiv: ['formalized_content', 'json'],
  TavilySearch: ['formalized_content', 'json', 'answer'],
  VariableAssigner: ['formalized_content', 'success'],
  VariableAggregator: ['formalized_content'],
  ListOperations: ['result', 'first', 'last', 'formalized_content'],
  StringTransform: ['result', 'formalized_content'],
  Invoke: ['content', 'formalized_content', 'status_code']
}

/** i18n 键后缀：rag.flowEditor.panel.<值> */
export const SYS_VAR_DESC_KEYS = {
  'sys.query': 'refDescSysQuery',
  'sys.user_id': 'refDescSysUserId',
  'sys.conversation_turns': 'refDescSysConversationTurns',
  'sys.files': 'refDescSysFiles',
  'sys.await_reply': 'refDescSysAwaitReply',
  'iteration.current': 'refDescIterationCurrent'
}

/** 输出字段名 → 说明文案 i18n 键（与组件类型无关，同一字段语义一致） */
export const OUTPUT_FIELD_DESC_KEYS = {
  formalized_content: 'refDescOutFormalizedContent',
  content: 'refDescOutContent',
  json: 'refDescOutJson',
  json_string: 'refDescOutJsonString',
  next_id: 'refDescOutNextId',
  selected_case: 'refDescOutSelectedCase',
  selected_category: 'refDescOutSelectedCategory',
  result: 'refDescOutResult',
  status_code: 'refDescOutStatusCode',
  user_input: 'refDescOutUserInput',
  raw_content: 'refDescOutRawContent',
  file_path: 'refDescOutFilePath',
  filename: 'refDescOutFilename',
  success: 'refDescOutSuccess',
  download: 'refDescOutDownload',
  pdf_base64: 'refDescOutPdfBase64',
  summary: 'refDescOutSummary',
  answer: 'refDescOutAnswer',
  first: 'refDescOutFirst',
  last: 'refDescOutLast'
}

/** 画布上始终可用的系统变量（与默认 DSL globals 一致） */
export const SYSTEM_REF_VARS = [
  { key: 'sys.query' },
  { key: 'sys.user_id' },
  { key: 'sys.conversation_turns' },
  { key: 'sys.files' },
  { key: 'sys.await_reply' },
  { key: 'iteration.current' }
]

/**
 * 根据占位符片段解析说明（用于预览区悬停）。tPanel 接收 panel 内的 i18n 键（不含前缀）。
 */
export function describeRefToken(token, tPanel) {
  if (!token || typeof token !== 'string' || !tPanel) return ''
  let norm = token.trim()
  if (norm.startsWith('{') && norm.endsWith('}')) {
    norm = norm.slice(1, -1).trim()
  }
  if (SYS_VAR_DESC_KEYS[norm]) {
    return tPanel(SYS_VAR_DESC_KEYS[norm])
  }
  const at = norm.match(/^[\w]+@([\w]+)$/)
  if (at && OUTPUT_FIELD_DESC_KEYS[at[1]]) {
    return tPanel(OUTPUT_FIELD_DESC_KEYS[at[1]])
  }
  return ''
}

/**
 * 收集所有能通过边「逆流」到达当前节点的上游节点 id（含多跳）。
 */
export function collectUpstreamNodeIds(currentId, edges) {
  if (!currentId || !edges?.length) return []
  const seen = new Set()
  const walk = (targetId) => {
    for (const e of edges) {
      if (e.target === targetId && e.source && !seen.has(e.source)) {
        seen.add(e.source)
        walk(e.source)
      }
    }
  }
  walk(currentId)
  return [...seen]
}

export function buildRefPickerModel({ nodes, edges, currentNodeId, tComp }) {
  const upstreamIds = collectUpstreamNodeIds(currentNodeId, edges)
  const sysOptions = SYSTEM_REF_VARS.map(({ key }) => ({
    ref: key,
    title: key,
    subtitle: '',
    descKey: SYS_VAR_DESC_KEYS[key] || null
  }))

  const nodeGroups = upstreamIds
    .map((nid) => {
      const n = nodes.find((x) => x.id === nid)
      if (!n) return null
      const comp = n.data?.componentName
      const keys = COMPONENT_OUTPUT_KEYS[comp]
      if (!keys?.length) return null
      const compLabel = tComp(comp) || comp
      const nodeTitle = n.data?.label || nid
      return {
        nodeId: nid,
        collapseTitle: `${nodeTitle} (${compLabel})`,
        options: keys.map((k) => ({
          ref: `${nid}@${k}`,
          title: k,
          subtitle: `${nodeTitle} · ${k}`,
          descKey: OUTPUT_FIELD_DESC_KEYS[k] || null
        }))
      }
    })
    .filter(Boolean)

  return { sysOptions, nodeGroups }
}

/**
 * 从文本中解析占位符，用于预览区标签。
 * - 统一展示为单层花括号 `{…}`，与插入变量一致。
 * - 同一语义只出现一次：已写在 `{node@key}` 内的不再单独出现裸 `node@key`。
 */
export function extractRefTokens(text) {
  if (!text || typeof text !== 'string') return []
  const seenInner = new Set()
  const ordered = []

  const pushCanonical = (inner) => {
    let key = String(inner || '').trim()
    if (!key) return
    let dedupeKey = key
    if (key.includes('@') || key.startsWith('sys.') || key.startsWith('iteration.')) {
      dedupeKey = key.toLowerCase()
      key = dedupeKey
    }
    if (seenInner.has(dedupeKey)) return
    seenInner.add(dedupeKey)
    ordered.push(`{${key}}`)
  }

  let m
  const reBrace = /\{[^{}]+\}/g
  while ((m = reBrace.exec(text))) {
    pushCanonical(m[0].slice(1, -1))
  }

  // 裸 node@output：画布节点多为 begin、xxx_0，避免匹配 llm 配置里的 ollama@llama3.2
  const reAt = /\b(begin|[a-z][a-z0-9_]*_\d+)@[\w]+\b/gi
  while ((m = reAt.exec(text))) {
    if (!seenInner.has(m[0])) pushCanonical(m[0])
  }

  const reSys = /\bsys\.\w+\b/g
  while ((m = reSys.exec(text))) {
    if (!seenInner.has(m[0])) pushCanonical(m[0])
  }

  const reIter = /\biteration\.current\b/g
  while ((m = reIter.exec(text))) {
    if (!seenInner.has(m[0])) pushCanonical(m[0])
  }

  return ordered
}

/** 插入占位符时统一为单层花括号 `{…}`，避免重复包裹 */
export function formatRefForInsert(rawRef) {
  let inner = String(rawRef || '').trim()
  if (inner.startsWith('{') && inner.endsWith('}')) {
    inner = inner.slice(1, -1).trim()
  }
  if (!inner) return ''
  return `{${inner}}`
}
