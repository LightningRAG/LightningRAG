<template>
  <aside class="agent-side-rail" :class="{ 'agent-side-rail--collapsed': railCollapsed }">
    <button
      type="button"
      class="agent-side-rail__edge-tab"
      :aria-expanded="!railCollapsed"
      :aria-label="railCollapsed ? $t('rag.agent.sideRailExpand') : $t('rag.agent.sideRailCollapse')"
      @click="railCollapsed = !railCollapsed"
    >
      <el-icon :size="14">
        <DArrowLeft v-if="!railCollapsed" />
        <DArrowRight v-else />
      </el-icon>
    </button>
    <div v-show="!railCollapsed" class="agent-side-rail__inner">
      <!-- 节点配置 -->
      <section v-if="showConfig" class="rail-block rail-block--config">
        <header class="rail-block__head" @click="openConfig = !openConfig">
          <span class="rail-block__title">{{ $t('rag.agent.panelNodeConfig') }}</span>
          <el-icon class="rail-block__chev">
            <ArrowUp v-if="openConfig" />
            <ArrowDown v-else />
          </el-icon>
        </header>
        <div v-show="openConfig" class="rail-block__body rail-block__body--config">
          <node-config-panel
            v-if="selectedNode"
            class="node-config-panel--embedded"
            :node="selectedNode"
            :nodes="nodes"
            :edges="edges"
            :knowledge-bases="knowledgeBases"
            :llm-models="llmModels"
            @update="onNodeParamsUpdate"
            @close="onCloseConfig"
          />
          <div v-else class="rail-placeholder">{{ $t('rag.flowEditor.selectNodeHint') }}</div>
        </div>
      </section>

      <!-- 运行调试 -->
      <section class="rail-block rail-block--debug">
        <header class="rail-block__head" @click="openDebug = !openDebug">
          <span class="rail-block__title">{{ $t('rag.agent.panelDebug') }}</span>
          <el-icon class="rail-block__chev">
            <ArrowUp v-if="openDebug" />
            <ArrowDown v-else />
          </el-icon>
        </header>
        <div v-show="openDebug" class="rail-block__body rail-block__body--debug">
          <el-input
            :model-value="runQuery"
            type="textarea"
            :rows="3"
            :placeholder="$t('rag.agent.placeholderQuery')"
            @update:model-value="emit('update:runQuery', $event)"
          />
          <div class="rail-debug-actions">
            <el-button type="primary" :loading="runLoading" @click="emit('run')">{{ $t('rag.agent.btnExecute') }}</el-button>
            <el-button v-if="runConversationId" size="small" @click="emit('clear-conversation')">{{ $t('rag.agent.btnNewConv') }}</el-button>
          </div>
        </div>
      </section>

      <!-- 输出 -->
      <section class="rail-block rail-block--output">
        <header class="rail-block__head" @click="openOutput = !openOutput">
          <span class="rail-block__title">{{ $t('rag.agent.panelOutput') }}</span>
          <el-icon class="rail-block__chev">
            <ArrowUp v-if="openOutput" />
            <ArrowDown v-else />
          </el-icon>
        </header>
        <div v-show="openOutput" class="rail-block__body rail-block__body--output">
          <div class="rail-output-scroll">
            <div v-if="!runOutput.trim()" class="rail-output-empty">{{ $t('rag.agent.outputPlaceholder') }}</div>
            <div v-else class="agent-run-markdown" v-html="runOutputHtml" />
          </div>
        </div>
      </section>
    </div>
  </aside>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ArrowDown, ArrowUp, DArrowLeft, DArrowRight } from '@element-plus/icons-vue'
import NodeConfigPanel from './flowEditor/nodeConfigPanel.vue'

defineOptions({ name: 'AgentEditorSideRail' })

const props = defineProps({
  showConfig: { type: Boolean, default: true },
  nodes: { type: Array, default: () => [] },
  edges: { type: Array, default: () => [] },
  knowledgeBases: { type: Array, default: () => [] },
  llmModels: { type: Array, default: () => [] },
  runQuery: { type: String, default: '' },
  runOutput: { type: String, default: '' },
  runOutputHtml: { type: String, default: '' },
  runLoading: { type: Boolean, default: false },
  runConversationId: { type: [String, Number], default: null }
})

const selectedNode = defineModel('selectedNode', { type: Object, default: null })

const emit = defineEmits(['update:runQuery', 'node-params-update', 'run', 'clear-conversation'])

const railCollapsed = ref(false)
const openConfig = ref(true)
const openDebug = ref(true)
const openOutput = ref(true)

const onNodeParamsUpdate = (newParams) => {
  emit('node-params-update', newParams)
}

const onCloseConfig = () => {
  selectedNode.value = null
}

// 选中节点时自动展开侧栏并打开「节点配置」区块（画布模式）
watch(
  () => selectedNode.value,
  (node) => {
    if (node && props.showConfig) {
      railCollapsed.value = false
      openConfig.value = true
    }
  }
)
</script>

<style scoped>
.agent-side-rail {
  --agent-rail-width: 440px;
  /* 节点配置折叠块最大高度（表单较长时多露出一些） */
  --agent-rail-config-max-h: min(58vh, 560px);
  position: relative;
  flex-shrink: 0;
  width: var(--agent-rail-width);
  min-width: var(--agent-rail-width);
  height: 100%;
  max-height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border-left: 1px solid var(--el-border-color);
  transition: width 0.2s ease, min-width 0.2s ease;
  /* 展开条用 translate 伸出侧栏外，不能被裁剪 */
  overflow: visible;
}

.agent-side-rail--collapsed {
  width: 0;
  min-width: 0;
  border-left: none;
  overflow: visible;
}

.agent-side-rail__inner {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  min-width: var(--agent-rail-width);
  overflow: hidden;
}

.agent-side-rail__edge-tab {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translate(-100%, -50%);
  z-index: 50;
  width: 22px;
  height: 48px;
  padding: 0;
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--el-border-color);
  border-right: none;
  border-radius: 6px 0 0 6px;
  background: var(--el-bg-color);
  color: var(--el-text-color-regular);
  cursor: pointer;
  box-shadow: var(--el-box-shadow-light);
}

.agent-side-rail__edge-tab:hover {
  color: var(--el-color-primary);
  background: var(--el-fill-color-light);
}

.agent-side-rail--collapsed .agent-side-rail__edge-tab {
  left: 0;
  transform: translate(0, -50%);
  border-radius: 6px 0 0 6px;
  border-right: none;
}

.rail-block {
  display: flex;
  flex-direction: column;
  min-height: 0;
  border-bottom: 1px solid var(--el-border-color);
}

.rail-block:last-child {
  border-bottom: none;
  flex: 1;
  min-height: 120px;
}

.rail-block__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 10px 12px;
  cursor: pointer;
  user-select: none;
  background: var(--el-fill-color-lighter);
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.rail-block__head:hover {
  background: var(--el-fill-color-light);
}

.rail-block__title {
  flex: 1;
  min-width: 0;
}

.rail-block__chev {
  flex-shrink: 0;
  color: var(--el-text-color-secondary);
}

.rail-block__body {
  padding: 12px;
  min-height: 0;
}

.rail-block__body--config {
  max-height: var(--agent-rail-config-max-h);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.rail-block__body--debug {
  flex-shrink: 0;
}

.rail-block__body--output {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 100px;
  padding: 12px;
  padding-top: 0;
}

.rail-block--output .rail-block__head {
  margin-bottom: 0;
}

.rail-block--output .rail-block__body--output {
  padding-top: 12px;
}

.rail-placeholder {
  padding: 24px 14px;
  text-align: center;
  font-size: 13px;
  color: var(--el-text-color-placeholder);
  line-height: 1.5;
}

.rail-debug-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}

.rail-output-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 10px 12px;
  border-radius: 6px;
  background: var(--el-fill-color-lighter);
  font-size: 13px;
}

.rail-output-empty {
  color: var(--el-text-color-placeholder);
}

:deep(.node-config-panel--embedded.node-config-panel) {
  width: 100%;
  border-left: none;
  height: 100%;
  max-height: var(--agent-rail-config-max-h);
}

:deep(.node-config-panel--embedded .panel-header) {
  padding: 10px 12px;
}

:deep(.node-config-panel--embedded .panel-body) {
  max-height: calc(var(--agent-rail-config-max-h) - 49px);
}

/* 与 editor.vue 中运行输出 Markdown 样式对齐 */
.rail-output-scroll .agent-run-markdown :deep(p) {
  margin: 0.5em 0;
}
.rail-output-scroll .agent-run-markdown :deep(p:first-child) {
  margin-top: 0;
}
.rail-output-scroll .agent-run-markdown :deep(p:last-child) {
  margin-bottom: 0;
}
.rail-output-scroll .agent-run-markdown :deep(ul),
.rail-output-scroll .agent-run-markdown :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.5em;
}
.rail-output-scroll .agent-run-markdown :deep(li) {
  margin: 0.25em 0;
}
.rail-output-scroll .agent-run-markdown :deep(h1),
.rail-output-scroll .agent-run-markdown :deep(h2),
.rail-output-scroll .agent-run-markdown :deep(h3) {
  margin: 0.75em 0 0.5em;
  font-weight: 600;
}
.rail-output-scroll .agent-run-markdown :deep(blockquote) {
  margin: 0.5em 0;
  padding-left: 1em;
  border-left: 4px solid #94a3b8;
  opacity: 0.9;
}
.rail-output-scroll .agent-run-markdown :deep(code) {
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-size: 0.9em;
  background: rgba(0, 0, 0, 0.08);
}
.rail-output-scroll .agent-run-markdown :deep(pre) {
  margin: 0.5em 0;
  padding: 0.75em;
  border-radius: 6px;
  overflow-x: auto;
  background: rgba(0, 0, 0, 0.06);
}
.rail-output-scroll .agent-run-markdown :deep(pre code) {
  padding: 0;
  background: none;
}
.rail-output-scroll .agent-run-markdown :deep(a) {
  color: #3b82f6;
  text-decoration: underline;
}
.rail-output-scroll .agent-run-markdown :deep(table) {
  border-collapse: collapse;
  margin: 0.5em 0;
}
.rail-output-scroll .agent-run-markdown :deep(th),
.rail-output-scroll .agent-run-markdown :deep(td) {
  border: 1px solid rgba(0, 0, 0, 0.1);
  padding: 0.25em 0.5em;
}
.dark .rail-output-scroll .agent-run-markdown :deep(code) {
  background: rgba(255, 255, 255, 0.1);
}
.dark .rail-output-scroll .agent-run-markdown :deep(pre) {
  background: rgba(255, 255, 255, 0.06);
}
.dark .rail-output-scroll .agent-run-markdown :deep(blockquote) {
  border-left-color: #64748b;
}
</style>

