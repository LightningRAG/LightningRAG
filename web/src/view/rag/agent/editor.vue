<template>
  <div class="flex flex-col h-[calc(100vh-120px)]">
    <div class="flex items-center justify-between py-2 border-b dark:border-slate-700">
      <div class="flex items-center gap-2">
        <el-button link icon="arrow-left" @click="goBack">{{ $t('rag.agent.back') }}</el-button>
        <span class="font-medium">{{ agent?.name || $t('rag.agent.editorTitle') }}</span>
        <el-tag v-if="agentId" type="info" size="small">{{ $t('rag.agent.tagSaved') }}</el-tag>
      </div>
      <div class="flex gap-2">
        <el-radio-group v-model="editMode" size="small">
          <el-radio-button label="flow">{{ $t('rag.agent.modeFlow') }}</el-radio-button>
          <el-radio-button label="json">{{ $t('rag.agent.modeJson') }}</el-radio-button>
        </el-radio-group>
        <el-button :loading="runLoading" type="primary" icon="video-play" @click="runDSL">{{ $t('rag.agent.btnRunHeader') }}</el-button>
        <el-button :loading="saveLoading" type="success" icon="check" @click="saveDSL">{{ $t('rag.agent.btnSave') }}</el-button>
      </div>
    </div>

    <div class="flex-1 flex min-h-0 gap-0 py-4 agent-editor-body">
      <!-- 画布编排模式：用 v-show 避免切换时销毁 FlowEditor 导致 patch 错误 -->
      <div v-show="editMode === 'flow'" class="flex-1 flex flex-col min-w-0 min-h-0 rounded border dark:border-slate-600 overflow-visible" style="min-height: 400px">
        <flow-editor
          ref="flowEditorRef"
          v-model="dslFromFlow"
          :knowledge-bases="knowledgeBases"
          :llm-models="llmModels"
          :run-query="runQuery"
          :run-output="runOutput"
          :run-output-html="runOutputHtml"
          :run-loading="runLoading"
          :run-conversation-id="runConversationId"
          @update:run-query="runQuery = $event"
          @run="runDSL"
          @clear-conversation="runConversationId = null"
        />
      </div>

      <!-- JSON 编辑模式：v-if 确保切换时以最新 dslText 挂载，避免空内容 -->
      <div v-if="editMode === 'json'" class="flex-1 flex flex-col min-w-0 min-h-0">
        <div class="text-sm text-gray-500 mb-2">{{ $t('rag.agent.dslJsonLabel') }}</div>
        <div class="flex-1 min-h-0 rounded border dark:border-slate-600 overflow-hidden">
          <v-ace-editor
            v-model:value="dslText"
            :options="aceOptions"
            lang="json"
            theme="github_dark"
            style="height: 100%; min-height: 400px"
          />
        </div>
        <div v-if="dslError" class="text-red-500 text-sm mt-2">{{ dslError }}</div>
      </div>
      <agent-editor-side-rail
        v-show="editMode === 'json'"
        v-model:selected-node="jsonSideSelectedNode"
        :show-config="false"
        :nodes="[]"
        :edges="[]"
        :knowledge-bases="knowledgeBases"
        :llm-models="llmModels"
        :run-query="runQuery"
        :run-output="runOutput"
        :run-output-html="runOutputHtml"
        :run-loading="runLoading"
        :run-conversation-id="runConversationId"
        @update:run-query="runQuery = $event"
        @run="runDSL"
        @clear-conversation="runConversationId = null"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VAceEditor } from 'vue3-ace-editor'
import 'ace-builds/src-noconflict/mode-json'
import 'ace-builds/src-noconflict/theme-github_dark'
import { agentGet, agentUpdate, agentRunStream, getKnowledgeBaseList, listLLMProviders } from '@/api/rag'
import { renderAgentMarkdown } from '@/utils/agentMarkdownRender'
import FlowEditor from './components/flowEditor/index.vue'
import AgentEditorSideRail from './components/AgentEditorSideRail.vue'

defineOptions({ name: 'RagAgentEditor' })

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const agentId = computed(() => {
  const id = route.query.id || route.params.id
  return id ? parseInt(id, 10) : null
})
const agent = ref(null)
const editMode = ref('flow')
const dslText = ref('')
const dslFromFlow = ref(null)
const dslError = ref('')
const runQuery = ref('')
const runOutput = ref('')
const runOutputHtml = computed(() => renderAgentMarkdown(runOutput.value))
const runLoading = ref(false)
const saveLoading = ref(false)
const flowEditorRef = ref(null)
const knowledgeBases = ref([])
const llmModels = ref([])
/** JSON 模式下侧栏无节点配置，仅占位 v-model，避免 defineModel 告警 */
const jsonSideSelectedNode = ref(null)

const aceOptions = {
  fontSize: 13,
  showPrintMargin: false,
  wrap: true,
  useWorker: false  // 避免 Vite 下 worker-json.js 加载失败
}

const defaultDSL = () => ({
  components: {
    begin: {
      obj: { component_name: 'Begin', params: { prologue: t('rag.agent.dslPrologue') } },
      downstream: ['retrieval_0'],
      upstream: []
    },
    retrieval_0: {
      obj: {
        component_name: 'Retrieval',
        params: {
          query: '{sys.query}',
          top_n: 20,
          empty_response: t('rag.agent.dslEmptyResponse'),
          kb_ids: []
        }
      },
      downstream: ['generate_0'],
      upstream: ['begin']
    },
    generate_0: {
      obj: {
        component_name: 'LLM',
        params: {
          llm_id: 'ollama@llama3.2',
          sys_prompt: t('rag.agent.dslSysPromptShort'),
          prompts: [{ role: 'user', content: '{sys.query}' }]
        }
      },
      downstream: ['message_0'],
      upstream: ['retrieval_0']
    },
    message_0: { obj: { component_name: 'Message', params: { content: ['{generate_0@content}'] } }, downstream: [], upstream: ['generate_0'] }
  },
  path: ['begin', 'retrieval_0', 'generate_0', 'message_0'],
  globals: { 'sys.query': '', 'sys.user_id': '', 'sys.conversation_turns': 0, 'sys.files': [] },
  history: [],
  retrieval: []
})

const parseDSL = () => {
  try {
    const v = JSON.parse(dslText.value || '{}')
    dslError.value = ''
    return v
  } catch (e) {
    dslError.value = t('rag.agent.jsonErrorPrefix') + e.message
    return null
  }
}

const getCurrentDSL = () => {
  if (editMode.value === 'flow') {
    return flowEditorRef.value?.getDsl?.() ?? dslFromFlow.value ?? defaultDSL()
  }
  return parseDSL() ?? defaultDSL()
}

const loadAgent = async () => {
  if (!agentId.value) {
    const dsl = defaultDSL()
    dslText.value = JSON.stringify(dsl, null, 2)
    dslFromFlow.value = dsl
    return
  }
  const res = await agentGet({ id: agentId.value })
  if (res.code === 0 && res.data) {
    agent.value = res.data
    try {
      const dsl = typeof res.data.dsl === 'string' ? JSON.parse(res.data.dsl) : res.data.dsl
      dslText.value = JSON.stringify(dsl, null, 2)
      dslFromFlow.value = dsl
    } catch {
      dslText.value = res.data.dsl || '{}'
      dslFromFlow.value = null
    }
  } else {
    ElMessage.error(t('rag.agent.loadFail'))
  }
}

const loadKnowledgeBases = async () => {
  const res = await getKnowledgeBaseList({ page: 1, pageSize: 100 })
  if (res.code === 0 && res.data?.list) {
    knowledgeBases.value = res.data.list
  }
}

const loadLLMModels = async () => {
  const res = await listLLMProviders({ scenarioType: 'chat' })
  if (res.code === 0 && res.data) {
    const list = res.data || []
    llmModels.value = list.map((m) => {
      const provider = m.name || m.provider || 'ollama'
      const modelName = m.modelName || 'llama3.2'
      return { id: `${provider}@${modelName}`, name: provider, modelName }
    })
  }
}

const saveDSL = async () => {
  const dsl = getCurrentDSL()
  if (!dsl) return
  if (!agentId.value) {
    ElMessage.warning(t('rag.agent.saveFirstWarning'))
    return
  }
  saveLoading.value = true
  try {
    const res = await agentUpdate({ id: agentId.value, dsl })
    if (res.code === 0) {
      ElMessage.success(t('rag.agent.saveOk'))
    } else {
      ElMessage.error(res.msg || t('rag.agent.saveFail'))
    }
  } finally {
    saveLoading.value = false
  }
}

const runConversationId = ref(null)
const runDSL = async () => {
  const dsl = getCurrentDSL()
  if (!dsl) return
  if (!runQuery.value?.trim()) {
    ElMessage.warning(t('rag.agent.needQuestion'))
    return
  }
  runLoading.value = true
  runOutput.value = ''
  try {
    const payload = { query: runQuery.value.trim() }
    if (agentId.value) {
      payload.agentId = agentId.value
    } else {
      payload.dsl = dsl
    }
    if (runConversationId.value) payload.conversationId = runConversationId.value
    await agentRunStream(payload, {
      onChunk: (chunk) => { runOutput.value += chunk },
      onDone: (convId, meta) => {
        if (convId) runConversationId.value = convId
        if (meta?.workflowPausedAtEntry) {
          ElMessage.info(t('rag.agent.workflowPausedHint'))
        }
      },
      onError: (err) => {
        runOutput.value = t('rag.agent.runFailedPrefix') + err
      }
    })
  } catch (e) {
    runOutput.value =
      t('rag.agent.runFailedPrefix') + (e.message || t('rag.agent.networkError'))
  } finally {
    runLoading.value = false
  }
}

const goBack = () => {
  router.push({ name: 'ragAgent' })
}

// 画布模式时持续同步 DSL 到 dslText，确保切换时已有最新内容
watch(
  () => [dslFromFlow.value, editMode.value],
  () => {
    if (editMode.value === 'flow' && dslFromFlow.value) {
      dslText.value = JSON.stringify(dslFromFlow.value, null, 2)
    }
  },
  { deep: true }
)

// 切换模式时同步 DSL
watch(editMode, async (mode) => {
  if (mode === 'flow') {
    const dsl = parseDSL()
    if (dsl) {
      dslFromFlow.value = dsl
      await nextTick()
      flowEditorRef.value?.initFromDsl?.(dsl)
    }
  } else {
    // 切换到 JSON 时，同步从画布获取最新 DSL（同步执行，在 DOM 切换前读取 flowEditorRef）
    const dsl = flowEditorRef.value?.getDsl?.() ?? dslFromFlow.value ?? defaultDSL()
    dslText.value = JSON.stringify(dsl, null, 2)
  }
})

onMounted(() => {
  loadAgent()
  loadKnowledgeBases()
  loadLLMModels()
})

watch(agentId, () => {
  runConversationId.value = null
  loadAgent()
})
</script>

<style scoped>
@import 'highlight.js/styles/github.css';
.agent-editor-body {
  position: relative;
}
.agent-run-markdown :deep(p) { margin: 0.5em 0; }
.agent-run-markdown :deep(p:first-child) { margin-top: 0; }
.agent-run-markdown :deep(p:last-child) { margin-bottom: 0; }
.agent-run-markdown :deep(ul), .agent-run-markdown :deep(ol) { margin: 0.5em 0; padding-left: 1.5em; }
.agent-run-markdown :deep(li) { margin: 0.25em 0; }
.agent-run-markdown :deep(h1), .agent-run-markdown :deep(h2), .agent-run-markdown :deep(h3) { margin: 0.75em 0 0.5em; font-weight: 600; }
.agent-run-markdown :deep(blockquote) { margin: 0.5em 0; padding-left: 1em; border-left: 4px solid #94a3b8; opacity: 0.9; }
.agent-run-markdown :deep(code) { padding: 0.2em 0.4em; border-radius: 4px; font-size: 0.9em; background: rgba(0,0,0,0.08); }
.agent-run-markdown :deep(pre) { margin: 0.5em 0; padding: 0.75em; border-radius: 6px; overflow-x: auto; background: rgba(0,0,0,0.06); }
.agent-run-markdown :deep(pre code) { padding: 0; background: none; }
.agent-run-markdown :deep(a) { color: #3b82f6; text-decoration: underline; }
.agent-run-markdown :deep(table) { border-collapse: collapse; margin: 0.5em 0; }
.agent-run-markdown :deep(th), .agent-run-markdown :deep(td) { border: 1px solid rgba(0,0,0,0.1); padding: 0.25em 0.5em; }
.dark .agent-run-markdown :deep(code) { background: rgba(255,255,255,0.1); }
.dark .agent-run-markdown :deep(pre) { background: rgba(255,255,255,0.06); }
.dark .agent-run-markdown :deep(blockquote) { border-left-color: #64748b; }
</style>
