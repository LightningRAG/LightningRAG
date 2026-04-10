<template>
  <div class="flow-editor">
    <div class="flow-left">
      <component-palette />
    </div>
    <div class="flow-center" @dragenter="dragging = true" @dragleave="onDragLeave">
      <VueFlow
        v-model:nodes="nodes"
        v-model:edges="edges"
        :default-viewport="{ x: 0, y: 0, zoom: 1 }"
        :min-zoom="0.2"
        :max-zoom="2"
        fit-view-on-init
        :elements-selectable="true"
        :delete-key-code="['Delete', 'Backspace']"
        @connect="onConnect"
        @node-click="onNodeClick"
        @pane-click="onPaneClick"
        @pane-context-menu="closeContextMenu"
        @node-context-menu="onNodeContextMenu"
      >
        <drop-zone :class="{ active: dragging }" @add-node="onAddNodeFromDrop" />
        <!-- 右键菜单 -->
        <Teleport to="body">
          <ul
            v-show="contextMenuVisible"
            :style="{ left: contextMenuLeft + 'px', top: contextMenuTop + 'px' }"
            class="flow-contextmenu"
          >
            <li @click="onContextMenuDelete">删除</li>
          </ul>
        </Teleport>
        <Background pattern-color="#e5e7eb" :gap="16" />
        <Controls />
        <MiniMap />
        <template #node-agent="nodeProps">
          <agent-node v-bind="nodeProps" />
        </template>
      </VueFlow>
    </div>
    <agent-editor-side-rail
      v-model:selected-node="selectedNode"
      :nodes="nodes"
      :edges="edges"
      :knowledge-bases="knowledgeBases"
      :llm-models="llmModels"
      :run-query="runQuery"
      :run-output="runOutput"
      :run-output-html="runOutputHtml"
      :run-loading="runLoading"
      :run-conversation-id="runConversationId"
      @update:run-query="emit('update:runQuery', $event)"
      @node-params-update="onNodeParamsUpdate"
      @run="emit('run')"
      @clear-conversation="emit('clear-conversation')"
    />
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { VueFlow } from '@vue-flow/core'
import { Background, Controls, MiniMap } from '@vue-flow/additional-components'
import ComponentPalette from './componentPalette.vue'
import DropZone from './dropZone.vue'
import AgentNode from './agentNode.vue'
import AgentEditorSideRail from '../AgentEditorSideRail.vue'
import { dslToFlow, flowToDsl } from './dslConverter'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'

defineOptions({ name: 'FlowEditor' })

const props = defineProps({
  modelValue: { type: Object, default: null },
  knowledgeBases: { type: Array, default: () => [] },
  llmModels: { type: Array, default: () => [] },
  runQuery: { type: String, default: '' },
  runOutput: { type: String, default: '' },
  runOutputHtml: { type: String, default: '' },
  runLoading: { type: Boolean, default: false },
  runConversationId: { type: [String, Number], default: null }
})

const emit = defineEmits(['update:modelValue', 'update:runQuery', 'run', 'clear-conversation'])

const nodes = ref([])
const edges = ref([])
const selectedNode = ref(null)
const dragging = ref(false)

// 右键菜单
const contextMenuVisible = ref(false)
const contextMenuLeft = ref(0)
const contextMenuTop = ref(0)
const contextMenuNode = ref(null)

const genId = (componentName) => {
  const prefix = componentName.toLowerCase()
  const existing = nodes.value.filter((n) => n.id.startsWith(prefix))
  const max = existing.reduce((acc, n) => {
    const m = n.id.match(/_(\d+)$/)
    return Math.max(acc, m ? parseInt(m[1], 10) : 0)
  }, -1)
  return `${prefix}_${max + 1}`
}

const onDragLeave = (e) => {
  if (!e.relatedTarget || !e.currentTarget.contains(e.relatedTarget)) {
    dragging.value = false
  }
}

const onAddNodeFromDrop = (payload) => {
  dragging.value = false
  onAddNode(payload)
}

const onAddNode = ({ componentName, label, color, defaultParams, position }) => {
  const id = genId(componentName.toLowerCase().replace(/([A-Z])/g, (m) => m.toLowerCase()))
  nodes.value = [
    ...nodes.value,
    {
      id,
      type: 'agent',
      position: position || { x: 100, y: 100 },
      data: {
        componentName,
        label,
        color: color || '#6366f1',
        params: { ...defaultParams }
      }
    }
  ]
}

const onConnect = (params) => {
  edges.value = [...edges.value, { id: `e-${params.source}-${params.target}`, ...params }]
}

const onNodeClick = ({ node }) => {
  selectedNode.value = node
}

const onPaneClick = () => {
  selectedNode.value = null
  contextMenuVisible.value = false
}

const onNodeContextMenu = ({ event, node }) => {
  event.preventDefault()
  contextMenuNode.value = node
  contextMenuLeft.value = event.clientX
  contextMenuTop.value = event.clientY
  contextMenuVisible.value = true
}

const onContextMenuDelete = () => {
  const node = contextMenuNode.value
  if (!node) return
  const nodeId = node.id
  nodes.value = nodes.value.filter((n) => n.id !== nodeId)
  edges.value = edges.value.filter((e) => e.source !== nodeId && e.target !== nodeId)
  if (selectedNode.value?.id === nodeId) {
    selectedNode.value = null
  }
  contextMenuVisible.value = false
  contextMenuNode.value = null
}

const closeContextMenu = () => {
  contextMenuVisible.value = false
  contextMenuNode.value = null
}

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})

const onNodeParamsUpdate = (newParams) => {
  if (!selectedNode.value) return
  const id = selectedNode.value.id
  nodes.value = nodes.value.map((n) => {
    if (n.id !== id) return n
    return { ...n, data: { ...n.data, params: newParams } }
  })
  selectedNode.value = nodes.value.find((n) => n.id === id) || selectedNode.value
}

// 从 DSL 初始化
const initFromDsl = (dsl) => {
  if (!dsl?.components) return
  const { nodes: n, edges: e } = dslToFlow(dsl)
  nodes.value = n
  edges.value = e
}

// 同步 selectedNode 到 nodes 变化；侧栏关闭时去掉画布选中态
watch(selectedNode, (node) => {
  if (node) {
    const n = nodes.value.find((n) => n.id === node.id)
    if (n) selectedNode.value = n
  } else {
    nodes.value = nodes.value.map((n) => ({ ...n, selected: false }))
  }
}, { deep: true })

watch([nodes, edges], () => {
  const dsl = flowToDsl(nodes.value, edges.value)
  emit('update:modelValue', dsl)
}, { deep: true })

watch(() => props.modelValue, (dsl) => {
  if (!dsl?.components || Object.keys(dsl.components).length === 0) return
  // 避免循环更新：若当前 nodes/edges 已对应此 dsl，则跳过
  const current = flowToDsl(nodes.value, edges.value)
  if (current?.components && JSON.stringify(current.components) === JSON.stringify(dsl.components)) return
  initFromDsl(dsl)
}, { immediate: true })

defineExpose({
  getDsl: () => flowToDsl(nodes.value, edges.value),
  initFromDsl
})
</script>

<style scoped>
/* 根节点不设 overflow:hidden，避免右侧收起的展开按钮被裁切；裁剪交给画布区 */
.flow-editor {
  display: flex;
  flex: 1;
  min-height: 0;
  border-radius: 8px;
  border: 1px solid var(--el-border-color);
  overflow: visible;
}

.flow-left {
  flex-shrink: 0;
  overflow: hidden;
}

.flow-center {
  flex: 1;
  min-width: 0;
  min-height: 400px;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.flow-contextmenu {
  position: fixed;
  z-index: 9999;
  margin: 0;
  padding: 4px 0;
  min-width: 100px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  box-shadow: var(--el-box-shadow-light);
  list-style: none;
}

.flow-contextmenu li {
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  color: var(--el-text-color-regular);
}

.flow-contextmenu li:hover {
  background: var(--el-fill-color-light);
  color: var(--el-color-danger);
}
</style>
