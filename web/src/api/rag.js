import service from '@/utils/request'
import { useUserStore } from '@/pinia/modules/user'
import { i18n, getLocaleHeaders } from '@/locale'

// ========== 知识库 ==========
export const createKnowledgeBase = (data) => {
  return service({
    url: '/rag/knowledgeBase/create',
    method: 'post',
    data
  })
}

export const getKnowledgeBaseList = (data) => {
  return service({
    url: '/rag/knowledgeBase/list',
    method: 'post',
    data
  })
}

export const listEmbeddingProviders = () => {
  return service({
    url: '/rag/knowledgeBase/listEmbeddingProviders',
    method: 'post',
    data: {}
  })
}

export const listVectorStoreConfigs = () => {
  return service({
    url: '/rag/knowledgeBase/listVectorStoreConfigs',
    method: 'post',
    data: {}
  })
}

export const listFileStorageConfigs = () => {
  return service({
    url: '/rag/knowledgeBase/listFileStorageConfigs',
    method: 'post',
    data: {}
  })
}

// ========== 设置（向量存储、文件存储 CRUD） ==========
export const listVectorStoreConfigsFull = (data) => {
  return service({
    url: '/rag/settings/vectorStore/list',
    method: 'post',
    data: data || { page: 1, pageSize: 10 }
  })
}

export const createVectorStoreConfig = (data) => {
  return service({
    url: '/rag/settings/vectorStore/create',
    method: 'post',
    data
  })
}

export const updateVectorStoreConfig = (data) => {
  return service({
    url: '/rag/settings/vectorStore/update',
    method: 'post',
    data
  })
}

export const deleteVectorStoreConfig = (data) => {
  return service({
    url: '/rag/settings/vectorStore/delete',
    method: 'post',
    data
  })
}

export const listFileStorageConfigsFull = (data) => {
  return service({
    url: '/rag/settings/fileStorage/list',
    method: 'post',
    data: data || { page: 1, pageSize: 10 }
  })
}

export const createFileStorageConfig = (data) => {
  return service({
    url: '/rag/settings/fileStorage/create',
    method: 'post',
    data
  })
}

export const updateFileStorageConfig = (data) => {
  return service({
    url: '/rag/settings/fileStorage/update',
    method: 'post',
    data
  })
}

export const deleteFileStorageConfig = (data) => {
  return service({
    url: '/rag/settings/fileStorage/delete',
    method: 'post',
    data
  })
}

export const getKnowledgeBase = (data) => {
  return service({
    url: '/rag/knowledgeBase/get',
    method: 'post',
    data
  })
}

export const updateKnowledgeBase = (data) => {
  return service({
    url: '/rag/knowledgeBase/update',
    method: 'post',
    data
  })
}

export const deleteKnowledgeBase = (data) => {
  return service({
    url: '/rag/knowledgeBase/delete',
    method: 'post',
    data
  })
}

export const getDocumentList = (data) => {
  return service({
    url: '/rag/knowledgeBase/listDocuments',
    method: 'post',
    data
  })
}

export const uploadDocument = (formData) => {
  return service({
    url: '/rag/knowledgeBase/uploadDocument',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': undefined // 让浏览器自动设置 multipart/form-data 及 boundary
    }
  })
}

export const getDocument = (data) => {
  return service({
    url: '/rag/knowledgeBase/getDocument',
    method: 'post',
    data
  })
}

export const deleteDocument = (data) => {
  return service({
    url: '/rag/knowledgeBase/deleteDocument',
    method: 'post',
    data
  })
}

export const retryDocument = (data) => {
  return service({
    url: '/rag/knowledgeBase/retryDocument',
    method: 'post',
    data
  })
}

export const batchDeleteDocuments = (data) => {
  return service({
    url: '/rag/knowledgeBase/batchDeleteDocuments',
    method: 'post',
    data
  })
}

export const batchReindexDocuments = (data) => {
  return service({
    url: '/rag/knowledgeBase/batchReindexDocuments',
    method: 'post',
    data
  })
}

export const batchCancelDocumentIndexing = (data) => {
  return service({
    url: '/rag/knowledgeBase/batchCancelDocumentIndexing',
    method: 'post',
    data
  })
}

export const batchSetDocumentRetrieval = (data) => {
  return service({
    url: '/rag/knowledgeBase/batchSetDocumentRetrieval',
    method: 'post',
    data
  })
}

export const batchSetDocumentPriority = (data) => {
  return service({
    url: '/rag/knowledgeBase/batchSetDocumentPriority',
    method: 'post',
    data
  })
}

/** 下载/预览文档（带 token，返回 blob 用于保存或预览）
 * @param {number} docId - 文档 ID
 * @param {Object} options - { preview: true } 时 URL 加 preview=1（后端返回 inline，用于新窗口打开）
 */
export const downloadDocument = async (docId, options = {}) => {
  const userStore = useUserStore()
  const baseURL = import.meta.env.VITE_BASE_API || ''
  let url = `${baseURL}/rag/knowledgeBase/downloadDocument?id=${docId}`
  if (options.preview) url += '&preview=1'
  const res = await fetch(url, {
    headers: {
      ...getLocaleHeaders(),
      'x-token': userStore.token || '',
      'x-user-id': String(userStore.userInfo?.ID || '')
    }
  })
  if (!res.ok) throw new Error(res.statusText || i18n.global.t('common.request.failed'))
  return res.blob()
}

// ========== 文档切片 ==========
export const getDocumentChunks = (data) => {
  return service({
    url: '/rag/knowledgeBase/listChunks',
    method: 'post',
    data
  })
}

/** 在选定知识库中按查询检索切片（测试检索，与对话引用结构一致） */
export const retrieveKnowledgeChunks = (data) => {
  return service({
    url: '/rag/knowledgeBase/retrieve',
    method: 'post',
    data
  })
}

/** 知识库图谱可视化子集（实体 / 关系，大库截断） */
export const getKnowledgeGraph = (data) => {
  return service({
    url: '/rag/knowledgeBase/knowledgeGraph',
    method: 'post',
    data
  })
}

export const updateDocumentChunk = (data) => {
  return service({
    url: '/rag/knowledgeBase/updateChunk',
    method: 'post',
    data
  })
}

export const shareKnowledgeBase = (data) => {
  return service({
    url: '/rag/knowledgeBase/share',
    method: 'post',
    data
  })
}

export const transferKnowledgeBase = (data) => {
  return service({
    url: '/rag/knowledgeBase/transfer',
    method: 'post',
    data
  })
}

// ========== 对话 ==========
export const createConversation = (data) => {
  return service({
    url: '/rag/conversation/create',
    method: 'post',
    data
  })
}

export const chatConversation = (data) => {
  return service({
    url: '/rag/conversation/chat',
    method: 'post',
    data
  })
}

/** 对话上下文纯检索（LightningRAG /query/data 风格，无 LLM） */
export const queryConversationData = (data) => {
  return service({
    url: '/rag/conversation/queryData',
    method: 'post',
    data
  })
}

/**
 * 流式对话，SSE 输出
 * @param {Object} data - { conversationId, content }
 * @param {Object} callbacks
 * @param {Function} callbacks.onChunk - (chunk: string) => void 每收到一块内容时调用
 * @param {Function} callbacks.onToolCall - (name: string, status: string, result?: string, toolCall?: {name, displayName, status, result}) => void 工具调用时调用，status: "start"|"done"
 * @param {Function} callbacks.onReferences - (references: Array) => void 检索完成后提前到达的引用数据
 * @param {Function} callbacks.onDone - (references?: Array, meta?: object) => void 完成时调用；meta 含 done、retrievalMode、retrievalQuery、searchQuery、onlyNeedContext 等（与 SSE 末帧一致）
 * @param {Function} callbacks.onError - (err: string) => void 错误时调用
 * @returns {Promise<void>}
 */
export const chatConversationStream = async (data, { onChunk, onToolCall, onReferences, onDone, onError }) => {
  const userStore = useUserStore()
  const baseURL = import.meta.env.VITE_BASE_API || ''
  const url = `${baseURL}/rag/conversation/chatStream`
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getLocaleHeaders(),
      'x-token': userStore.token || '',
      'x-user-id': String(userStore.userInfo?.ID || '')
    },
    body: JSON.stringify(data)
  })
  if (!res.ok) {
    onError?.(res.statusText || i18n.global.t('common.request.failed'))
    return
  }
  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  let doneHandled = false
  try {
    let readResult = await reader.read()
    while (!readResult.done) {
      buf += decoder.decode(readResult.value, { stream: true })
      const lines = buf.split('\n')
      buf = lines.pop() || ''
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const json = JSON.parse(line.slice(6))
            if (json.error) {
              onError?.(json.error)
              return
            }
            if (json.content) onChunk?.(json.content)
            if (json.toolCall) onToolCall?.(json.toolCall.name, json.toolCall.status, json.toolCall.result, json.toolCall)
            if (json.references && !json.done) onReferences?.(json.references)
            if (json.done) {
              doneHandled = true
              onDone?.(json.references || [], json)
            }
          } catch {
            // 忽略非 JSON 的 SSE 行（如不完整 chunk）
          }
        }
      }
      readResult = await reader.read()
    }
    if (buf.startsWith('data: ')) {
      try {
        const json = JSON.parse(buf.slice(6))
        if (json.content) onChunk?.(json.content)
        if (json.toolCall) onToolCall?.(json.toolCall.name, json.toolCall.status, json.toolCall.result, json.toolCall)
        if (json.references && !json.done) onReferences?.(json.references)
        if (json.done) {
          doneHandled = true
          onDone?.(json.references || [], json)
        }
      } catch {
        // 忽略尾部缓冲区中非合法 JSON
      }
    }
    if (!doneHandled) onDone?.([])
  } catch (e) {
    onError?.(e.message || i18n.global.t('common.request.failed'))
  }
}

export const getConversationList = (data) => {
  return service({
    url: '/rag/conversation/list',
    method: 'post',
    data
  })
}

export const getConversation = (data) => {
  return service({
    url: '/rag/conversation/get',
    method: 'post',
    data
  })
}

/**
 * 更新对话（如修改启用的工具）
 * @param {Object} data - { id, enabledToolNames }
 */
export const updateConversation = (data) => {
  return service({
    url: '/rag/conversation/update',
    method: 'post',
    data
  })
}

export const deleteConversation = (data) => {
  return service({
    url: '/rag/conversation/delete',
    method: 'post',
    data
  })
}

/**
 * 获取对话可用工具列表（供展示及扩展）
 * @returns {Promise<{code, data: Array<{name, displayName, description}>}>}
 */
export const listConversationTools = () => {
  return service({
    url: '/rag/conversation/listTools',
    method: 'post',
    data: {}
  })
}

/**
 * 获取对话消息列表（历史记录）
 * @param {Object} data - { conversationId, page?, pageSize? }
 */
export const getConversationMessages = (data) => {
  return service({
    url: '/rag/conversation/listMessages',
    method: 'post',
    data: { page: 1, pageSize: 100, ...data }
  })
}

// ========== 模型 ==========
// scenarioType: 场景类型 chat|embedding|rerank|speech2text|tts|ocr|cv，空则返回全部
export const listLLMProviders = (data = {}) => {
  return service({
    url: '/rag/llm/listProviders',
    method: 'post',
    data
  })
}

/**
 * 获取指定场景类型可用的提供商列表（用于添加/编辑模型时的下拉选项）
 * @param {Object} data - { scenarioType?: string, scenarioTypes?: string[] }
 */
export const listAvailableProviders = (data = {}) => {
  return service({
    url: '/rag/llm/listAvailableProviders',
    method: 'post',
    data
  })
}

export const listUserModels = () => {
  return service({
    url: '/rag/llm/listUserModels',
    method: 'post',
    data: {}
  })
}

export const addUserModel = (data) => {
  return service({
    url: '/rag/llm/addUserModel',
    method: 'post',
    data
  })
}

export const updateUserModel = (data) => {
  return service({
    url: '/rag/llm/updateUserModel',
    method: 'post',
    data
  })
}

export const deleteUserModel = (data) => {
  return service({
    url: '/rag/llm/deleteUserModel',
    method: 'post',
    data
  })
}

// 角色默认模型（管理员）
export const setAuthorityDefaultLLM = (data) => {
  return service({
    url: '/rag/llm/setAuthorityDefaultLLM',
    method: 'post',
    data
  })
}

export const getAuthorityDefaultLLMs = (data) => {
  return service({
    url: '/rag/llm/getAuthorityDefaultLLMs',
    method: 'post',
    data
  })
}

export const clearAuthorityDefaultLLM = (data) => {
  return service({
    url: '/rag/llm/clearAuthorityDefaultLLM',
    method: 'post',
    data
  })
}

// 用户默认模型
export const setUserDefaultLLM = (data) => {
  return service({
    url: '/rag/llm/setUserDefaultLLM',
    method: 'post',
    data
  })
}

export const getUserDefaultLLMs = () => {
  return service({
    url: '/rag/llm/getUserDefaultLLMs',
    method: 'post',
    data: {}
  })
}

export const clearUserDefaultLLM = (data) => {
  return service({
    url: '/rag/llm/clearUserDefaultLLM',
    method: 'post',
    data
  })
}

// 互联网搜索配置
export const listWebSearchProviders = () => {
  return service({
    url: '/rag/llm/listWebSearchProviders',
    method: 'post',
    data: {}
  })
}

export const getWebSearchConfig = () => {
  return service({
    url: '/rag/llm/getWebSearchConfig',
    method: 'post',
    data: {}
  })
}

export const setWebSearchConfig = (data) => {
  return service({
    url: '/rag/llm/setWebSearchConfig',
    method: 'post',
    data
  })
}

// ========== 系统全局模型配置 ==========
export const listAdminModels = (data = {}) => {
  return service({
    url: '/rag/systemModel/listAdminModels',
    method: 'post',
    data: { page: 1, pageSize: 50, ...data }
  })
}

export const createAdminModel = (data) => {
  return service({
    url: '/rag/systemModel/createAdminModel',
    method: 'post',
    data
  })
}

export const updateAdminModel = (data) => {
  return service({
    url: '/rag/systemModel/updateAdminModel',
    method: 'post',
    data
  })
}

export const deleteAdminModel = (data) => {
  return service({
    url: '/rag/systemModel/deleteAdminModel',
    method: 'post',
    data
  })
}

export const getSystemDefaults = () => {
  return service({
    url: '/rag/systemModel/getSystemDefaults',
    method: 'post',
    data: {}
  })
}

export const setSystemDefault = (data) => {
  return service({
    url: '/rag/systemModel/setSystemDefault',
    method: 'post',
    data
  })
}

export const clearSystemDefault = (data) => {
  return service({
    url: '/rag/systemModel/clearSystemDefault',
    method: 'post',
    data
  })
}

// ========== 系统互联网搜索配置 ==========
export const listSystemWebSearchProviders = () => {
  return service({
    url: '/rag/systemModel/listSystemWebSearchProviders',
    method: 'post',
    data: {}
  })
}

export const getSystemWebSearchConfig = () => {
  return service({
    url: '/rag/systemModel/getSystemWebSearchConfig',
    method: 'post',
    data: {}
  })
}

export const setSystemWebSearchConfig = (data) => {
  return service({
    url: '/rag/systemModel/setSystemWebSearchConfig',
    method: 'post',
    data
  })
}

export const clearSystemWebSearchConfig = () => {
  return service({
    url: '/rag/systemModel/clearSystemWebSearchConfig',
    method: 'post',
    data: {}
  })
}

// ========== 全局共享知识库 ==========
export const listGlobalKnowledgeBases = () => {
  return service({
    url: '/rag/systemModel/listGlobalKnowledgeBases',
    method: 'post',
    data: {}
  })
}

export const setGlobalKnowledgeBase = (data) => {
  return service({
    url: '/rag/systemModel/setGlobalKnowledgeBase',
    method: 'post',
    data
  })
}

export const removeGlobalKnowledgeBase = (data) => {
  return service({
    url: '/rag/systemModel/removeGlobalKnowledgeBase',
    method: 'post',
    data
  })
}

export const listAllKnowledgeBases = () => {
  return service({
    url: '/rag/systemModel/listAllKnowledgeBases',
    method: 'post',
    data: {}
  })
}

// ========== Agent 流程编排 ==========
export const agentRun = (data) => {
  return service({
    url: '/rag/agent/run',
    method: 'post',
    data
  })
}

/**
 * 流式运行 Agent，SSE 输出，支持多轮对话
 * @param {Object} data - { agentId?, dsl?, query, conversationId?, workflowGlobals? }
 * @param {Object} callbacks
 * @param {Function} callbacks.onChunk - (chunk: string) => void
 * @param {Function} callbacks.onDone - (conversationId?: number, meta?: { workflowPausedAtEntry?: boolean }) => void
 * @param {Function} callbacks.onError - (err: string) => void
 */
export const agentRunStream = async (data, { onChunk, onDone, onError }) => {
  const userStore = useUserStore()
  const baseURL = import.meta.env.VITE_BASE_API || ''
  const url = `${baseURL}/rag/agent/runStream`
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getLocaleHeaders(),
      'x-token': userStore.token || '',
      'x-user-id': String(userStore.userInfo?.ID || '')
    },
    body: JSON.stringify(data)
  })
  if (!res.ok) {
    onError?.(res.statusText || i18n.global.t('common.request.failed'))
    return
  }
  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  try {
    let readResult = await reader.read()
    while (!readResult.done) {
      buf += decoder.decode(readResult.value, { stream: true })
      const lines = buf.split('\n')
      buf = lines.pop() || ''
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const json = JSON.parse(line.slice(6))
            if (json.error) {
              onError?.(json.error)
              return
            }
            if (json.content) onChunk?.(json.content)
            if (json.done) {
              onDone?.(json.conversationId, {
                workflowPausedAtEntry: !!json.workflowPausedAtEntry
              })
            }
          } catch {
            // 忽略非 JSON 的 SSE 行（如不完整 chunk）
          }
        }
      }
      readResult = await reader.read()
    }
    if (buf.startsWith('data: ')) {
      try {
        const json = JSON.parse(buf.slice(6))
        if (json.content) onChunk?.(json.content)
        if (json.done) {
          onDone?.(json.conversationId, {
            workflowPausedAtEntry: !!json.workflowPausedAtEntry
          })
        }
      } catch {
        // 忽略尾部缓冲区中非合法 JSON
      }
    }
  } catch (e) {
    onError?.(e.message || i18n.global.t('common.request.failed'))
  }
}

export const agentListTemplates = () => {
  return service({
    url: '/rag/agent/templates',
    method: 'post',
    data: {}
  })
}

export const agentLoadTemplate = (data) => {
  return service({
    url: '/rag/agent/loadTemplate',
    method: 'post',
    data
  })
}

export const agentCreate = (data) => {
  return service({
    url: '/rag/agent/create',
    method: 'post',
    data
  })
}

export const agentList = (data) => {
  return service({
    url: '/rag/agent/list',
    method: 'post',
    data
  })
}

export const agentGet = (data) => {
  return service({
    url: '/rag/agent/get',
    method: 'post',
    data
  })
}

export const agentUpdate = (data) => {
  return service({
    url: '/rag/agent/update',
    method: 'post',
    data
  })
}

export const agentDelete = (data) => {
  return service({
    url: '/rag/agent/delete',
    method: 'post',
    data
  })
}

export const agentCreateFromTemplate = (data) => {
  return service({
    url: '/rag/agent/createFromTemplate',
    method: 'post',
    data
  })
}

// ========== 第三方渠道连接器（飞书 / 钉钉 / Discord 等 Webhook）==========
export const channelConnectorList = (data) => {
  return service({
    url: '/rag/channelConnector/list',
    method: 'post',
    data: data || { page: 1, pageSize: 10 }
  })
}

/** 后端 Register 的渠道列表（字典序），用于新建连接器下拉 */
export const channelConnectorChannelTypes = () => {
  return service({
    url: '/rag/channelConnector/channelTypes',
    method: 'post',
    data: {}
  })
}

export const channelConnectorCreate = (data) => {
  return service({
    url: '/rag/channelConnector/create',
    method: 'post',
    data
  })
}

export const channelConnectorUpdate = (data) => {
  return service({
    url: '/rag/channelConnector/update',
    method: 'post',
    data
  })
}

export const channelConnectorDelete = (data) => {
  return service({
    url: '/rag/channelConnector/delete',
    method: 'post',
    data
  })
}

export const channelOutboundList = (data) => {
  return service({
    url: '/rag/channelConnector/outbound/list',
    method: 'post',
    data: data || { page: 1, pageSize: 10 }
  })
}

export const channelOutboundDelete = (data) => {
  return service({
    url: '/rag/channelConnector/outbound/delete',
    method: 'post',
    data
  })
}

export const channelOutboundRunOnce = () => {
  return service({
    url: '/rag/channelConnector/outbound/runOnce',
    method: 'post',
    data: {}
  })
}
