<template>
  <div class="agent-node" :class="{ selected: selected }" :style="{ borderColor: nodeColor }">
    <Handle type="target" :position="Position.Top" class="node-handle" />
    <div class="node-header">
      <el-icon class="node-icon" :style="{ color: nodeColor }">
        <component :is="iconComponent" />
      </el-icon>
      <span class="node-label">{{ label }}</span>
      <span v-if="nodeId" class="node-id">{{ nodeId }}</span>
    </div>
    <div v-if="nodeData?.params?.prologue" class="node-preview">
      {{ truncate(nodeData.params.prologue, 30) }}
    </div>
    <div v-else-if="nodeData?.params?.llm_id" class="node-preview">
      {{ nodeData.params.llm_id }}
    </div>
    <div v-else-if="nodeData?.params?.top_n && ['Retrieval','DuckDuckGo','Wikipedia','ArXiv'].includes(nodeData?.componentName)" class="node-preview">
      top_n: {{ nodeData.params.top_n }}
    </div>
    <div v-else-if="nodeData?.params?.tool_name" class="node-preview">
      {{ nodeData.params.tool_name }}
    </div>
    <div v-else-if="nodeData?.params?.input && nodeData?.componentName === 'Transformer'" class="node-preview">
      {{ truncate(nodeData.params.input, 28) }}
    </div>
    <div v-else-if="nodeData?.componentName === 'AwaitResponse'" class="node-preview">
      {{ truncate(nodeData.params.variable_key || 'sys.await_reply', 24) }}
    </div>
    <Handle type="source" :position="Position.Bottom" class="node-handle" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Handle, Position } from '@vue-flow/core'
import { VideoPlay, Search, ChatDotRound, Message, Box, Switch, CollectionTag, Link, Refresh, Document, DocumentAdd, Avatar, DataAnalysis, Share, Setting, MagicStick, ChatLineRound, Compass, Reading, Notebook, TrendCharts, EditPen, Connection, Sort, CopyDocument, Promotion } from '@element-plus/icons-vue'

const { t, te } = useI18n()

const props = defineProps({
  id: String,
  data: Object,
  selected: Boolean
})

const iconMap = {
  Begin: VideoPlay,
  Retrieval: Search,
  LLM: ChatDotRound,
  Agent: Avatar,
  Message: Message,
  Switch,
  Categorize: CollectionTag,
  HTTPRequest: Link,
  Iteration: Refresh,
  TextProcessing: Document,
  ExecuteSQL: DataAnalysis,
  DocsGenerator: DocumentAdd,
  MCP: Share,
  SetVariable: Setting,
  Transformer: MagicStick,
  AwaitResponse: ChatLineRound,
  DuckDuckGo: Compass,
  Wikipedia: Reading,
  ArXiv: Notebook,
  TavilySearch: TrendCharts,
  VariableAssigner: EditPen,
  VariableAggregator: Connection,
  ListOperations: Sort,
  StringTransform: CopyDocument,
  Invoke: Promotion
}

const nodeId = computed(() => props.id)
const nodeData = computed(() => props.data || {})
const componentName = computed(() => nodeData.value.componentName || '')
const label = computed(() => {
  const name = componentName.value
  const key = name ? `rag.flowEditor.comp.${name}` : ''
  if (key && te(key)) {
    return t(key)
  }
  return nodeData.value.label || name || t('rag.flowEditor.nodeFallback')
})
const nodeColor = computed(() => nodeData.value.color || '#6366f1')
const iconComponent = computed(() => iconMap[componentName.value] || 'Box')

const truncate = (s, len) => {
  if (!s || typeof s !== 'string') return ''
  return s.length > len ? s.slice(0, len) + '...' : s
}
</script>

<style scoped>
.agent-node {
  min-width: 160px;
  padding: 0;
  border-radius: 8px;
  border: 2px solid;
  background: var(--el-bg-color);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.agent-node.selected {
  box-shadow: 0 0 0 2px var(--el-color-primary);
}

.node-handle {
  width: 10px;
  height: 10px;
  background: var(--el-border-color);
  border: 2px solid var(--el-bg-color);
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
}

.node-icon {
  font-size: 18px;
}

.node-label {
  font-weight: 600;
  font-size: 13px;
  flex: 1;
}

.node-id {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.node-preview {
  padding: 0 12px 10px;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
