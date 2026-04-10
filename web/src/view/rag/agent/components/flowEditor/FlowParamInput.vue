<template>
  <div class="flow-param-input" :class="{ 'flow-param-input--small': small }">
    <div class="flow-param-input__row" :class="inputType === 'textarea' ? 'flow-param-input__row--textarea' : 'flow-param-input__row--text'">
      <el-input
        ref="inputRef"
        :model-value="modelValue"
        :type="inputType"
        :rows="rows"
        :placeholder="placeholder"
        :size="small ? 'small' : 'default'"
        class="flow-param-input__field"
        :class="{ 'flow-param-input__field--has-refs': refTokensWithHints.length > 0 }"
        @update:model-value="(v) => emit('update:modelValue', v)"
      />
      <el-popover
        v-model:visible="popoverVisible"
        placement="left-start"
        :width="340"
        trigger="click"
      >
        <template #reference>
          <el-button
            :size="small ? 'small' : 'small'"
            type="primary"
            plain
            class="flow-param-input__pick-btn"
            :title="$t('rag.flowEditor.panel.refInsertTitle')"
          >
            <span class="flow-param-input__pick-label">{{ $t('rag.flowEditor.panel.refInsert') }}</span>
          </el-button>
        </template>
        <div class="flow-param-ref-popover">
          <el-input
            v-model="filterText"
            size="small"
            clearable
            :placeholder="$t('rag.flowEditor.panel.refFilterPh')"
            class="flow-param-ref-popover__search"
          />
          <div v-if="!currentNodeId" class="flow-param-ref-popover__empty">{{ $t('rag.flowEditor.panel.refNoNode') }}</div>
          <template v-else>
            <div v-if="filteredSys.length" class="flow-param-ref-popover__section">
              <div class="flow-param-ref-popover__section-title">{{ $t('rag.flowEditor.panel.refGroupSys') }}</div>
              <div
                v-for="opt in filteredSys"
                :key="opt.ref"
                class="flow-param-ref-popover__opt-row"
                role="button"
                tabindex="0"
                @click="insertRef(opt.ref)"
                @keydown.enter.prevent="insertRef(opt.ref)"
              >
                <code class="flow-param-ref-popover__code">{{ formatRefForInsert(opt.ref) }}</code>
                <span v-if="opt.descText" class="flow-param-ref-popover__opt-desc">{{ opt.descText }}</span>
              </div>
            </div>
            <div v-if="filteredNodeGroups.length" class="flow-param-ref-popover__section">
              <div class="flow-param-ref-popover__section-title">{{ $t('rag.flowEditor.panel.refGroupUpstream') }}</div>
              <el-collapse class="flow-param-ref-popover__collapse">
                <el-collapse-item v-for="g in filteredNodeGroups" :key="g.nodeId" :title="g.collapseTitle" :name="g.nodeId">
                  <div
                    v-for="opt in g.options"
                    :key="opt.ref"
                    class="flow-param-ref-popover__opt-row"
                    role="button"
                    tabindex="0"
                    @click="insertRef(opt.ref)"
                    @keydown.enter.prevent="insertRef(opt.ref)"
                  >
                    <code class="flow-param-ref-popover__code">{{ formatRefForInsert(opt.ref) }}</code>
                    <span v-if="opt.descText" class="flow-param-ref-popover__opt-desc">{{ opt.descText }}</span>
                  </div>
                </el-collapse-item>
              </el-collapse>
            </div>
            <div
              v-if="currentNodeId && !filteredSys.length && !filteredNodeGroups.length"
              class="flow-param-ref-popover__empty"
            >
              {{ filterText.trim() ? $t('rag.flowEditor.panel.refNoMatch') : $t('rag.flowEditor.panel.refNoUpstream') }}
            </div>
          </template>
        </div>
      </el-popover>
    </div>
    <div v-if="refTokensWithHints.length" class="flow-param-ref-preview">
      <div class="flow-param-ref-preview__head">
        <span class="flow-param-ref-preview__label">{{ $t('rag.flowEditor.panel.refParsedLabel') }}</span>
        <span class="flow-param-ref-preview__hint">{{ $t('rag.flowEditor.panel.refChipTooltipHint') }}</span>
      </div>
      <div class="flow-param-ref-preview__chips">
        <el-tooltip
          v-for="item in refTokensWithHints"
          :key="item.token"
          placement="top"
          :show-after="400"
          :disabled="!item.hint"
          :content="item.hint"
        >
          <span class="flow-param-ref-chip">{{ item.token }}</span>
        </el-tooltip>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { buildRefPickerModel, describeRefToken, extractRefTokens, formatRefForInsert } from './flowRefCatalog.js'

defineOptions({ name: 'FlowParamInput' })

const emit = defineEmits(['update:modelValue'])

const props = defineProps({
  modelValue: { type: String, default: '' },
  nodes: { type: Array, default: () => [] },
  edges: { type: Array, default: () => [] },
  currentNodeId: { type: String, default: '' },
  inputType: { type: String, default: 'text' },
  rows: { type: Number, default: 2 },
  placeholder: { type: String, default: '' },
  small: { type: Boolean, default: false }
})

const { t } = useI18n()
const inputRef = ref(null)
const popoverVisible = ref(false)
const filterText = ref('')

watch(popoverVisible, (open) => {
  if (open) filterText.value = ''
})

const tComp = (name) => t(`rag.flowEditor.comp.${name}`, name)

const pickerModel = computed(() =>
  buildRefPickerModel({
    nodes: props.nodes,
    edges: props.edges,
    currentNodeId: props.currentNodeId,
    tComp
  })
)

const fq = (s) => s.toLowerCase()

const panelDesc = (descKey) => (descKey ? t(`rag.flowEditor.panel.${descKey}`) : '')

const filteredSys = computed(() => {
  const q = filterText.value.trim().toLowerCase()
  let list = pickerModel.value.sysOptions.map((o) => ({
    ...o,
    descText: panelDesc(o.descKey)
  }))
  if (q) {
    list = list.filter((o) => fq(o.ref).includes(q) || fq(o.descText).includes(q))
  }
  return list
})

const filteredNodeGroups = computed(() => {
  const q = filterText.value.trim().toLowerCase()
  const enrich = (g) => ({
    ...g,
    options: g.options.map((o) => ({
      ...o,
      descText: panelDesc(o.descKey)
    }))
  })
  let groups = pickerModel.value.nodeGroups.map(enrich)
  if (q) {
    groups = groups
      .map((g) => ({
        ...g,
        options: g.options.filter(
          (o) =>
            fq(o.ref).includes(q) ||
            fq(o.subtitle).includes(q) ||
            fq(o.title).includes(q) ||
            fq(o.descText).includes(q)
        )
      }))
      .filter((g) => g.options.length > 0)
  }
  return groups
})

const refTokensWithHints = computed(() =>
  extractRefTokens(props.modelValue || '').map((token) => ({
    token,
    hint: describeRefToken(token, (k) => t(`rag.flowEditor.panel.${k}`))
  }))
)

const insertRef = (rawRef) => {
  const piece = formatRefForInsert(rawRef)
  const cur = props.modelValue ?? ''
  const el = inputRef.value?.textarea ?? inputRef.value?.input
  let newVal
  let caret = (cur?.length ?? 0)
  if (el && typeof el.selectionStart === 'number') {
    const start = el.selectionStart
    const end = el.selectionEnd ?? start
    newVal = cur.slice(0, start) + piece + cur.slice(end)
    caret = start + piece.length
  } else {
    newVal = cur + piece
  }
  emit('update:modelValue', newVal)
  popoverVisible.value = false
  nextTick(() => {
    const el2 = inputRef.value?.textarea ?? inputRef.value?.input
    if (el2 && typeof el2.setSelectionRange === 'function') {
      el2.setSelectionRange(caret, caret)
      el2.focus()
    }
  })
}

</script>

<style scoped>
.flow-param-input__row {
  display: flex;
  gap: 6px;
}

.flow-param-input__row--textarea {
  align-items: flex-start;
}

.flow-param-input__row--text {
  align-items: center;
}

.flow-param-input__field {
  flex: 1;
  min-width: 0;
}

.flow-param-input__field--has-refs :deep(.el-textarea__inner),
.flow-param-input__field--has-refs :deep(.el-input__wrapper) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.flow-param-input__pick-btn {
  flex-shrink: 0;
  margin-top: 0;
  font-weight: 600;
  border-style: dashed;
}

.flow-param-input--small .flow-param-input__pick-btn {
  padding: 5px 8px;
}

.flow-param-input__pick-label {
  font-size: 11px;
  white-space: nowrap;
}

.flow-param-ref-popover__search {
  margin-bottom: 8px;
}

.flow-param-ref-popover__section-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

.flow-param-ref-popover__opt-row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  width: 100%;
  padding: 8px 6px;
  margin: 0 0 4px;
  border-radius: 6px;
  cursor: pointer;
  outline: none;
}

.flow-param-ref-popover__opt-row:hover,
.flow-param-ref-popover__opt-row:focus-visible {
  background: var(--el-fill-color-light);
}

.flow-param-ref-popover__code {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px dashed var(--el-color-primary-light-5);
}

.flow-param-ref-popover__opt-desc {
  font-size: 11px;
  line-height: 1.4;
  color: var(--el-text-color-secondary);
  padding-left: 2px;
}

.flow-param-ref-popover__collapse {
  border: none;
  --el-collapse-header-height: 32px;
}

.flow-param-ref-popover__collapse :deep(.el-collapse-item__header) {
  font-size: 12px;
  padding-left: 0;
}

.flow-param-ref-popover__collapse :deep(.el-collapse-item__wrap) {
  border-bottom: none;
}

.flow-param-ref-popover__empty {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  padding: 8px 0;
}

.flow-param-ref-preview {
  margin-top: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--el-color-warning-light-9);
  border: 1px dashed var(--el-color-warning-light-3);
}

.flow-param-ref-preview__head {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 6px;
}

.flow-param-ref-preview__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-color-warning-dark-2);
}

.flow-param-ref-preview__hint {
  font-size: 10px;
  color: var(--el-text-color-placeholder);
  line-height: 1.35;
}

.flow-param-ref-preview__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.flow-param-ref-chip {
  display: inline-block;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
  color: var(--el-color-primary);
  background: var(--el-bg-color);
  border: 2px dashed var(--el-color-primary);
  line-height: 1.3;
}
</style>
