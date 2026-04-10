<template>
  <div
    class="drop-zone"
    @drop="onDrop"
    @dragover="onDragOver"
  />
</template>

<script setup>
import { useVueFlow } from '@vue-flow/core'

const emit = defineEmits(['add-node'])

const { screenToFlowCoordinate } = useVueFlow()

const onDragOver = (e) => {
  e.preventDefault()
  e.dataTransfer.dropEffect = 'move'
}

const onDrop = (e) => {
  e.preventDefault()
  const data = e.dataTransfer.getData('application/json')
  if (!data) return
  try {
    const payload = JSON.parse(data)
    if (payload.type !== 'component') return
    const { x, y } = screenToFlowCoordinate({ x: e.clientX, y: e.clientY })
    emit('add-node', { ...payload, position: { x: x - 90, y: y - 40 } })
  } catch (_) {}
}
</script>

<style scoped>
.drop-zone {
  position: absolute;
  inset: 0;
  z-index: 5;
  pointer-events: none;
}

.drop-zone.active {
  pointer-events: auto;
}
</style>
