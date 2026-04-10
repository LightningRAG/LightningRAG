<template>
  <div class="component-palette">
    <div class="palette-title">{{ $t('rag.flowEditor.paletteTitle') }}</div>
    <div class="palette-list">
      <div class="palette-section">
        <div class="palette-section-label">{{ $t('rag.flowEditor.paletteCore') }}</div>
        <div
          v-for="c in coreList"
          :key="c.key"
          class="palette-item"
          :style="{ borderColor: c.color }"
          draggable="true"
          @dragstart="onDragStart($event, c)"
        >
          <el-icon class="palette-icon" :style="{ color: c.color }">
            <component :is="iconMap[c.icon] || 'Box'" />
          </el-icon>
          <span class="palette-label">{{ $t('rag.flowEditor.comp.' + c.componentName) }}</span>
        </div>
      </div>

      <div class="palette-divider" role="separator" :aria-label="$t('rag.flowEditor.paletteToolsAria')">
        <span class="palette-divider-line" />
        <span class="palette-divider-text">{{ $t('rag.flowEditor.paletteTools') }}</span>
        <span class="palette-divider-line" />
      </div>

      <div class="palette-section palette-section--tools">
        <div
          v-for="c in toolList"
          :key="c.key"
          class="palette-item"
          :style="{ borderColor: c.color }"
          draggable="true"
          @dragstart="onDragStart($event, c)"
        >
          <el-icon class="palette-icon" :style="{ color: c.color }">
            <component :is="iconMap[c.icon] || 'Box'" />
          </el-icon>
          <span class="palette-label">{{ $t('rag.flowEditor.comp.' + c.componentName) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { VideoPlay, Search, ChatDotRound, Message, Box, Switch, CollectionTag, Link, Refresh, Document, DocumentAdd, Avatar, DataAnalysis, Share, Setting, MagicStick, ChatLineRound, Compass, Reading, Notebook, TrendCharts, EditPen, Connection, Sort, CopyDocument, Promotion } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { PALETTE_CORE_LIST, PALETTE_TOOL_LIST } from './componentTypes'

defineOptions({ name: 'ComponentPalette' })

const { t } = useI18n()

const coreList = PALETTE_CORE_LIST
const toolList = PALETTE_TOOL_LIST

const iconMap = {
  'play-circle': VideoPlay,
  'search': Search,
  'chat-dot-round': ChatDotRound,
  'avatar': Avatar,
  'message': Message,
  'switch': Switch,
  'collection-tag': CollectionTag,
  'link': Link,
  'refresh': Refresh,
  'document': Document,
  'document-add': DocumentAdd,
  'data-analysis': DataAnalysis,
  'share': Share,
  'setting': Setting,
  'magic-stick': MagicStick,
  'chat-line-round': ChatLineRound,
  'compass': Compass,
  'reading': Reading,
  'notebook': Notebook,
  'trend-charts': TrendCharts,
  'edit-pen': EditPen,
  'connection': Connection,
  'sort': Sort,
  'copy-document': CopyDocument,
  'promotion': Promotion
}

const onDragStart = (e, comp) => {
  e.dataTransfer.setData('application/json', JSON.stringify({
    type: 'component',
    componentName: comp.componentName,
    label: t('rag.flowEditor.comp.' + comp.componentName),
    defaultParams: comp.defaultParams,
    color: comp.color
  }))
  e.dataTransfer.effectAllowed = 'move'
}
</script>

<style scoped>
.component-palette {
  width: 148px;
  height: 100%;
  background: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color);
  padding: 12px 8px;
  overflow-y: auto;
}

.palette-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 10px;
  padding: 0 4px;
}

.palette-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.palette-section-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  text-transform: none;
  letter-spacing: 0.02em;
  margin-bottom: 8px;
  padding: 0 4px;
}

.palette-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.palette-section--tools {
  margin-top: 0;
}

.palette-divider {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 14px 0 12px;
  padding: 0 2px;
}

.palette-divider-line {
  flex: 1;
  height: 1px;
  background: var(--el-border-color);
  min-width: 8px;
}

.palette-divider-text {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-placeholder);
  white-space: nowrap;
}

.palette-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 8px;
  border: 2px solid;
  background: var(--el-fill-color-light);
  cursor: grab;
  transition: all 0.2s;
}

.palette-item:hover {
  background: var(--el-fill-color);
  transform: translateY(-1px);
}

.palette-item:active {
  cursor: grabbing;
}

.palette-icon {
  font-size: 18px;
}

.palette-label {
  font-size: 13px;
  font-weight: 500;
}
</style>
