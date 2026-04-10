<template>
  <Panel position="top-right" class="kg-flow-panel">
    <el-button-group class="kg-flow-btns">
      <el-button size="small" @click="zoomOut()">-</el-button>
      <el-button size="small" @click="onFit">{{ $t('rag.kgGraph.fitView') }}</el-button>
      <el-button size="small" @click="zoomIn()">+</el-button>
    </el-button-group>
    <el-button
      class="kg-flow-focus"
      size="small"
      :disabled="!focusNodeId"
      @click="onFocusSelection"
    >
      {{ $t('rag.kgGraph.focusSelection') }}
    </el-button>
  </Panel>
</template>

<script setup>
import { watch, nextTick } from 'vue'
import { Panel, useVueFlow } from '@vue-flow/core'

defineOptions({ name: 'KgGraphFlowPanel' })

const props = defineProps({
  focusNodeId: { type: String, default: '' },
  /** 递增时在 `autoFitNodeId` 上执行一次居中（如双击节点） */
  autoFitPulse: { type: Number, default: 0 },
  autoFitNodeId: { type: String, default: '' }
})

const { fitView, zoomIn, zoomOut } = useVueFlow()

function onFit() {
  fitView({ padding: 0.2, duration: 220 })
}

function fitNodeById(id) {
  const nid = String(id || '').trim()
  if (!nid) return
  fitView({ nodes: [nid], padding: 0.38, duration: 300, maxZoom: 2 })
}

function onFocusSelection() {
  fitNodeById(props.focusNodeId)
}

watch(
  () => props.autoFitPulse,
  () => {
    if (!props.autoFitPulse) return
    nextTick(() => fitNodeById(props.autoFitNodeId))
  }
)
</script>

<style scoped>
.kg-flow-panel {
  margin: 8px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}
.kg-flow-btns {
  box-shadow: 0 1px 4px rgb(15 23 42 / 12%);
  border-radius: 6px;
  overflow: hidden;
}
.kg-flow-focus {
  box-shadow: 0 1px 4px rgb(15 23 42 / 12%);
}
</style>
