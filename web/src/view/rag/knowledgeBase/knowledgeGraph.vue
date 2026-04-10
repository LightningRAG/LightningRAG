<template>
  <div class="kg-viz-page">
    <div class="kg-viz-toolbar">
      <el-button icon="arrow-left" @click="goBack">{{ $t('rag.kgGraph.back') }}</el-button>
      <span class="kg-viz-title">{{ $t('rag.kgGraph.title') }}：{{ kbName }}</span>
      <el-tag v-if="payload && !payload.enableKnowledgeGraph" type="warning" size="small" class="ml-2">
        {{ $t('rag.kgGraph.kgDisabledHint') }}
      </el-tag>
      <el-tag
        v-if="payload?.scopeDocumentId"
        type="success"
        size="small"
        closable
        class="ml-1"
        @close="clearDocumentScope"
      >
        {{ $t('rag.kgGraph.scopeDoc', { id: payload.scopeDocumentId }) }}
      </el-tag>
      <div class="flex-1" />
      <span v-if="payload && graphEntities.length" class="kg-stats text-sm text-gray-500">
        {{ statsLine }}
      </span>
      <el-popover placement="bottom-end" :width="340" trigger="click">
        <template #reference>
          <el-button size="small">{{ $t('rag.kgGraph.dataScope') }}</el-button>
        </template>
        <div class="kg-scope-popover">
          <div class="kg-scope-row">
            <span class="kg-scope-label">{{ $t('rag.kgGraph.maxEntitiesLabel') }}</span>
            <el-input-number v-model="loadMaxEntities" :min="50" :max="800" :step="50" size="small" controls-position="right" />
          </div>
          <div class="kg-scope-row">
            <span class="kg-scope-label">{{ $t('rag.kgGraph.maxRelsLabel') }}</span>
            <el-input-number v-model="loadMaxRels" :min="50" :max="2000" :step="100" size="small" controls-position="right" />
          </div>
          <el-checkbox v-model="hideEdgeLabels" class="kg-scope-check">{{ $t('rag.kgGraph.hideEdgeLabels') }}</el-checkbox>
          <p class="kg-scope-hint">{{ $t('rag.kgGraph.exportJsonHint') }}</p>
          <el-button size="small" class="kg-scope-export" :disabled="!graphEntities.length" @click="exportGraphJson">
            {{ $t('rag.kgGraph.exportJson') }}
          </el-button>
          <p class="kg-scope-hint">{{ $t('rag.kgGraph.exportVisibleJsonHint') }}</p>
          <el-button size="small" class="kg-scope-export" :disabled="!displayEntities.length" @click="exportVisibleGraphJson">
            {{ $t('rag.kgGraph.exportVisibleJson') }}
          </el-button>
          <el-button type="primary" size="small" class="kg-scope-reload" :loading="loading" @click="loadGraph">
            {{ $t('rag.kgGraph.reloadWithScope') }}
          </el-button>
        </div>
      </el-popover>
      <el-tooltip :content="isFs ? $t('rag.kgGraph.fsExit') : $t('rag.kgGraph.fsEnter')" placement="bottom">
        <el-button size="small" icon="full-screen" :disabled="!screenfullOk" @click="toggleCanvasFullscreen" />
      </el-tooltip>
      <el-button :loading="loading" icon="refresh" @click="loadGraph">{{ $t('rag.kgGraph.refresh') }}</el-button>
    </div>

    <div v-if="graphEntities.length" class="kg-viz-subbar">
      <el-input
        v-model="searchQuery"
        clearable
        class="kg-search"
        :placeholder="$t('rag.kgGraph.searchPh')"
        prefix-icon="search"
      />
      <span class="text-sm text-gray-500 mr-1 shrink-0">{{ $t('rag.kgGraph.layoutLabel') }}</span>
      <el-tooltip :disabled="forceAllowed" :content="$t('rag.kgGraph.layoutForceLimited')" placement="bottom">
        <el-radio-group v-model="layoutMode" size="small">
          <el-radio-button value="circle">{{ $t('rag.kgGraph.layoutCircle') }}</el-radio-button>
          <el-radio-button value="force" :disabled="!forceAllowed">{{ $t('rag.kgGraph.layoutForce') }}</el-radio-button>
        </el-radio-group>
      </el-tooltip>
      <el-checkbox v-model="egoFilter" size="small">{{ $t('rag.kgGraph.egoNeighborhood') }}</el-checkbox>
      <el-select
        v-model="typeFilterSelected"
        multiple
        filterable
        clearable
        collapse-tags
        collapse-tags-tooltip
        class="kg-type-filter"
        :placeholder="$t('rag.kgGraph.typeFilterPh')"
      >
        <el-option v-for="typ in allEntityTypes" :key="typ" :label="typ" :value="typ" />
      </el-select>
      <el-input
        v-model="relKeywordInput"
        clearable
        class="kg-rel-filter"
        :placeholder="$t('rag.kgGraph.relKeywordPh')"
      />
      <el-button size="small" :disabled="!graphFilterActive" @click="resetGraphFilters">
        {{ $t('rag.kgGraph.resetFilters') }}
      </el-button>
    </div>

    <div v-if="typeLegend.length" class="kg-legend">
      <span class="kg-legend-title">{{ $t('rag.kgGraph.legendTitle') }}：</span>
      <span v-for="item in typeLegend" :key="item.key" class="kg-legend-item">
        <i class="kg-legend-dot" :style="{ background: item.border }" />
        {{ item.label }}
      </span>
    </div>

    <el-alert
      v-if="payload?.truncated"
      type="info"
      :closable="false"
      class="kg-viz-alert"
      :title="$t('rag.kgGraph.truncatedTitle')"
      :description="truncateDetail"
    />

    <div class="kg-viz-body">
      <div ref="canvasHostRef" class="kg-viz-canvas" v-loading="loading">
        <VueFlow
          v-if="nodes.length"
          :key="flowMountKey"
          v-model:nodes="nodes"
          v-model:edges="edges"
          :min-zoom="0.12"
          :max-zoom="2.2"
          :nodes-connectable="false"
          :edges-selectable="true"
          :default-edge-options="{ type: 'smoothstep', animated: false }"
          fit-view-on-init
          class="kg-vue-flow"
          @node-click="onNodeClick"
          @node-double-click="onNodeDoubleClick"
          @edge-click="onEdgeClick"
          @edge-mouse-enter="onEdgeMouseEnter"
          @edge-mouse-move="onEdgeMouseMove"
          @edge-mouse-leave="onEdgeMouseLeave"
          @pane-click="onPaneClick"
        >
          <Background pattern-color="#cbd5e1" :gap="20" />
          <Controls />
          <MiniMap
            pannable
            zoomable
            :node-color="minimapNodeColor"
            mask-color="rgb(15 23 42 / 14%)"
            :width="168"
            :height="112"
          />
          <KgGraphFlowPanel
            :focus-node-id="selectedId"
            :auto-fit-pulse="nodeAutoFitPulse"
            :auto-fit-node-id="nodeAutoFitNodeId"
          />
          <template #node-kg="nodeProps">
            <div
              class="kg-node"
              :class="{
                'is-selected': selectedId === nodeProps.id,
                'is-dimmed': isNodeDimmed(nodeProps)
              }"
              :style="nodeChromeStyle(nodeProps)"
              :title="nodeProps.data?.descTitle || nodeProps.data?.label"
            >
              <div class="kg-node-label">{{ nodeProps.data?.label }}</div>
              <div v-if="nodeProps.data?.sub" class="kg-node-type">{{ nodeProps.data.sub }}</div>
            </div>
          </template>
        </VueFlow>
        <div v-else-if="!loading" class="kg-viz-empty">
          <p class="kg-empty-text">{{ emptyHint }}</p>
          <el-button v-if="isFilteredViewEmpty" type="primary" size="small" @click="resetGraphFilters">
            {{ $t('rag.kgGraph.resetFilters') }}
          </el-button>
        </div>
      </div>
      <aside class="kg-viz-side">
        <template v-if="selectedEntity">
          <div class="kg-side-head-row">
            <span class="kg-side-head">{{ $t('rag.kgGraph.entityDetail') }}</span>
            <el-button text type="primary" size="small" @click="clearSelection">{{ $t('rag.kgGraph.clearSelection') }}</el-button>
          </div>
          <div class="kg-side-name-row">
            <span class="kg-side-name">{{ selectedEntity.name }}</span>
            <el-button text type="primary" size="small" @click="copyText(selectedEntity.name)">{{ $t('rag.kgGraph.copyName') }}</el-button>
          </div>
          <div v-if="selectedEntity.entityType" class="kg-side-row">
            <span class="kg-side-k">{{ $t('rag.kgGraph.entityType') }}</span>
            <span class="kg-type-pill" :style="{ borderColor: typeColor(selectedEntity.entityType).border }">
              {{ selectedEntity.entityType }}
            </span>
          </div>
          <div v-if="selectedEntity.description" class="kg-side-desc">{{ selectedEntity.description }}</div>
        </template>
        <template v-else-if="selectedRel">
          <div class="kg-side-head-row">
            <span class="kg-side-head">{{ $t('rag.kgGraph.relDetail') }}</span>
            <el-button text type="primary" size="small" @click="clearSelection">{{ $t('rag.kgGraph.clearSelection') }}</el-button>
          </div>
          <div class="kg-side-row">
            <span class="kg-side-k">{{ $t('rag.kgGraph.srcEntity') }}</span>
            {{ entityName(selectedRel.sourceEntityId) }}
          </div>
          <div class="kg-side-row">
            <span class="kg-side-k">{{ $t('rag.kgGraph.tgtEntity') }}</span>
            {{ entityName(selectedRel.targetEntityId) }}
          </div>
          <div v-if="selectedRel.keywords" class="kg-side-block">
            <div class="kg-side-k-row">
              <span class="kg-side-k">{{ $t('rag.kgGraph.relKeywords') }}</span>
              <el-button text type="primary" size="small" @click="copyText(selectedRel.keywords)">{{ $t('rag.kgGraph.copyField') }}</el-button>
            </div>
            <div class="kg-side-desc">{{ selectedRel.keywords }}</div>
          </div>
          <div v-if="selectedRel.description" class="kg-side-block">
            <div class="kg-side-k-row">
              <span class="kg-side-k">{{ $t('rag.kgGraph.relDesc') }}</span>
              <el-button text type="primary" size="small" @click="copyText(selectedRel.description)">{{ $t('rag.kgGraph.copyField') }}</el-button>
            </div>
            <div class="kg-side-desc">{{ selectedRel.description }}</div>
          </div>
        </template>
        <template v-else>
          <div class="kg-side-placeholder">{{ sidePlaceholder }}</div>
        </template>
      </aside>
    </div>

    <Teleport to="body">
      <div
        v-show="edgeTip.show && edgeTip.rel"
        class="kg-edge-tip"
        :style="{ left: edgeTip.x + 'px', top: edgeTip.y + 'px' }"
      >
        <template v-if="edgeTip.rel">
          <div class="kg-edge-tip-line">
            <span class="kg-edge-tip-k">{{ entityName(edgeTip.rel.sourceEntityId) }}</span>
            <span class="kg-edge-tip-arrow">→</span>
            <span class="kg-edge-tip-k">{{ entityName(edgeTip.rel.targetEntityId) }}</span>
          </div>
          <div v-if="edgeTip.rel.keywords" class="kg-edge-tip-block">
            <div class="kg-edge-tip-h">{{ $t('rag.kgGraph.edgeTipKw') }}</div>
            <div class="kg-edge-tip-t">{{ edgeTip.rel.keywords }}</div>
          </div>
          <div v-if="edgeTip.rel.description" class="kg-edge-tip-block">
            <div class="kg-edge-tip-h">{{ $t('rag.kgGraph.edgeTipDesc') }}</div>
            <div class="kg-edge-tip-t">{{ edgeTip.rel.description }}</div>
          </div>
        </template>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useEventListener, watchDebounced } from '@vueuse/core'
import { VueFlow } from '@vue-flow/core'
import { Background, Controls, MiniMap } from '@vue-flow/additional-components'
import { MarkerType } from '@vue-flow/core'
import screenfull from 'screenfull'
import { getKnowledgeGraph } from '@/api/rag'
import { ElMessage } from 'element-plus'
import KgGraphFlowPanel from './kgGraphFlowPanel.vue'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'

defineOptions({ name: 'RagKnowledgeGraph' })

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const kbId = computed(() => Number(route.query.id) || 0)
const kbName = computed(() => route.query.name || t('rag.kgGraph.unnamedKb'))
const fromDocs = computed(() => route.query.from === 'docs')
const documentScopeId = computed(() => {
  const d = Number(route.query.documentId)
  return Number.isFinite(d) && d > 0 ? d : 0
})

const loading = ref(false)
const payload = ref(null)
const graphEntities = ref([])
const graphRels = ref([])
const nodes = ref([])
const edges = ref([])
const selectedId = ref('')
const selectedEdgeId = ref('')
const flowMountKey = ref(0)
const layoutMode = ref('circle')
const searchQuery = ref('')
const egoFilter = ref(false)
const typeFilterSelected = ref([])
/** 防抖后的关系关键词，驱动拓扑筛选 */
const relKeywordQuery = ref('')
/** 输入框即时值（用于重置按钮状态与防抖源） */
const relKeywordInput = ref('')
const loadMaxEntities = ref(400)
const loadMaxRels = ref(800)
const hideEdgeLabels = ref(false)
const canvasHostRef = ref(null)
const isFs = ref(false)
const edgeTip = reactive({
  show: false,
  x: 0,
  y: 0,
  rel: null
})

const nodeAutoFitPulse = ref(0)
const nodeAutoFitNodeId = ref('')

watchDebounced(
  relKeywordInput,
  (v) => {
    relKeywordQuery.value = String(v ?? '')
  },
  { debounce: 280 }
)

const screenfullOk = computed(() => screenfull.isEnabled)

const allEntityTypes = computed(() => {
  const s = new Set()
  const unk = t('rag.kgGraph.typeUnknown')
  for (const e of graphEntities.value) {
    s.add((e.entityType || '').trim() || unk)
  }
  return Array.from(s).sort((a, b) => a.localeCompare(b))
})

const typeFilteredEntities = computed(() => {
  const base = graphEntities.value
  if (!typeFilterSelected.value.length) {
    return base
  }
  const allowed = new Set(typeFilterSelected.value)
  const unk = t('rag.kgGraph.typeUnknown')
  return base.filter((e) => {
    const typ = (e.entityType || '').trim() || unk
    return allowed.has(typ)
  })
})

const typeFilteredRels = computed(() => {
  const ids = new Set(typeFilteredEntities.value.map((e) => e.id))
  return graphRels.value.filter((r) => ids.has(r.sourceEntityId) && ids.has(r.targetEntityId))
})

const stageEntities = computed(() => {
  const q = relKeywordQuery.value.trim().toLowerCase()
  if (!q) {
    return typeFilteredEntities.value
  }
  const rels = typeFilteredRels.value.filter(
    (r) =>
      (r.keywords || '').toLowerCase().includes(q) || (r.description || '').toLowerCase().includes(q)
  )
  if (!rels.length) {
    return []
  }
  const entIds = new Set()
  for (const r of rels) {
    entIds.add(r.sourceEntityId)
    entIds.add(r.targetEntityId)
  }
  return typeFilteredEntities.value.filter((e) => entIds.has(e.id))
})

const stageRels = computed(() => {
  const q = relKeywordQuery.value.trim().toLowerCase()
  if (!q) {
    return typeFilteredRels.value
  }
  return typeFilteredRels.value.filter(
    (r) =>
      (r.keywords || '').toLowerCase().includes(q) || (r.description || '').toLowerCase().includes(q)
  )
})

const egoCenterEntityId = computed(() => {
  const s = selectedId.value
  if (!s || !String(s).startsWith('e-')) return 0
  const n = Number(String(s).slice(2))
  return Number.isFinite(n) ? n : 0
})

const displayEntities = computed(() => {
  const base = stageEntities.value
  if (!egoFilter.value || !egoCenterEntityId.value) {
    return base
  }
  const nid = egoCenterEntityId.value
  if (!base.some((e) => e.id === nid)) {
    return base
  }
  const nbr = new Set([nid])
  for (const r of stageRels.value) {
    if (r.sourceEntityId === nid) {
      nbr.add(r.targetEntityId)
    }
    if (r.targetEntityId === nid) {
      nbr.add(r.sourceEntityId)
    }
  }
  return base.filter((e) => nbr.has(e.id))
})

const displayRels = computed(() => {
  const ids = new Set(displayEntities.value.map((e) => e.id))
  return stageRels.value.filter((r) => ids.has(r.sourceEntityId) && ids.has(r.targetEntityId))
})

const forceAllowed = computed(() => displayEntities.value.length > 0 && displayEntities.value.length <= 200)

const isFilteredViewEmpty = computed(
  () => !loading.value && graphEntities.value.length > 0 && displayEntities.value.length === 0
)

const graphFilterActive = computed(
  () =>
    typeFilterSelected.value.length > 0 ||
    !!relKeywordInput.value.trim() ||
    egoFilter.value ||
    !!searchQuery.value.trim()
)

const statsLine = computed(() =>
  t('rag.kgGraph.statsLine', {
    ne: displayEntities.value.length,
    nr: displayRels.value.length
  })
)

const nameById = computed(() => {
  const m = new Map()
  for (const e of graphEntities.value) {
    m.set(e.id, e.name || `ID:${e.id}`)
  }
  return m
})

function entityName(id) {
  return nameById.value.get(id) || `#${id}`
}

const selectedEntity = computed(() => {
  if (!selectedId.value || !payload.value?.entities) return null
  const raw = selectedId.value.replace(/^e-/, '')
  const id = Number(raw)
  return payload.value.entities.find((e) => e.id === id) || null
})

const selectedRel = computed(() => {
  if (!selectedEdgeId.value || !payload.value?.relationships) return null
  const rid = Number(selectedEdgeId.value.replace(/^r-/, ''))
  return payload.value.relationships.find((r) => r.id === rid) || null
})

const sidePlaceholder = computed(() => {
  if (!graphEntities.value.length) return ''
  return t('rag.kgGraph.sideHint')
})

const typeLegend = computed(() => {
  const seen = new Map()
  for (const e of displayEntities.value) {
    const label = (e.entityType || '').trim() || t('rag.kgGraph.typeUnknown')
    if (!seen.has(label)) {
      const c = typeColor(e.entityType)
      seen.set(label, { key: label, label, border: c.border })
    }
  }
  return Array.from(seen.values()).sort((a, b) => a.label.localeCompare(b.label))
})

const emptyHint = computed(() => {
  if (!kbId.value) return t('rag.kgGraph.noKbId')
  if (!graphEntities.value.length) return t('rag.kgGraph.emptyGraph')
  if (isFilteredViewEmpty.value) return t('rag.kgGraph.filterNoMatch')
  return t('rag.kgGraph.emptyGraph')
})

const truncateDetail = computed(() => {
  if (!payload.value) return ''
  return t('rag.kgGraph.truncatedBody', {
    ec: payload.value.entityCount,
    rc: payload.value.relationshipCount,
    ne: payload.value.entities?.length ?? 0,
    nr: payload.value.relationships?.length ?? 0,
    me: payload.value.maxEntitiesRequested,
    mr: payload.value.maxRelationshipsRequested
  })
})

function minimapNodeColor(node) {
  const b = node.data?.borderColor
  return typeof b === 'string' && b ? b : '#64748b'
}

/** 按类型生成稳定配色（HSL） */
function typeColor(entityType) {
  const s = entityType || ''
  let h = 216
  for (let i = 0; i < s.length; i++) {
    h = (h * 31 + s.charCodeAt(i)) >>> 0
  }
  h = h % 360
  return {
    border: `hsl(${h} 58% 40%)`,
    bg: `hsl(${h} 35% 96%)`
  }
}

function nodeChromeStyle(nodeProps) {
  const b = nodeProps.data?.borderColor
  const bg = nodeProps.data?.bg
  if (!b) return {}
  return {
    borderColor: b,
    background: bg
  }
}

function isNodeDimmed(nodeProps) {
  const o = nodeProps.style?.opacity
  return typeof o === 'number' && o < 0.5
}

function layoutCirclePositions(entities, w, h) {
  const n = entities.length
  const cx = w / 2
  const cy = h / 2
  const R = Math.min(w, h) * 0.36
  return entities.map((_, i) => {
    const a = (2 * Math.PI * i) / Math.max(n, 1) - Math.PI / 2
    return { x: cx + R * Math.cos(a), y: cy + R * Math.sin(a) }
  })
}

function layoutForcePositions(entities, relationships, width, height) {
  const n = entities.length
  const idxById = new Map(entities.map((e, i) => [e.id, i]))
  const pos = layoutCirclePositions(entities, width, height)
  const links = []
  for (const r of relationships) {
    const s = idxById.get(r.sourceEntityId)
    const t = idxById.get(r.targetEntityId)
    if (s !== undefined && t !== undefined && s !== t) {
      links.push({ s, t })
    }
  }
  const area = width * height
  const k = Math.sqrt(area / Math.max(n, 1))
  let temp = 0.12 * Math.sqrt(area)
  const inward = 0.07
  const iterCount = Math.min(130, 45 + Math.floor(n * 1.8))
  for (let iter = 0; iter < iterCount; iter++) {
    const disp = Array.from({ length: n }, () => ({ x: 0, y: 0 }))
    for (let i = 0; i < n; i++) {
      for (let j = i + 1; j < n; j++) {
        let dx = pos[j].x - pos[i].x
        let dy = pos[j].y - pos[i].y
        const dist = Math.hypot(dx, dy) || 0.01
        const f = (k * k) / dist
        dx /= dist
        dy /= dist
        disp[i].x -= dx * f
        disp[i].y -= dy * f
        disp[j].x += dx * f
        disp[j].y += dy * f
      }
    }
    for (const { s, t } of links) {
      let dx = pos[t].x - pos[s].x
      let dy = pos[t].y - pos[s].y
      const dist = Math.hypot(dx, dy) || 0.01
      const f = ((dist * dist) / k) * 0.035
      dx /= dist
      dy /= dist
      disp[s].x += dx * f
      disp[s].y += dy * f
      disp[t].x -= dx * f
      disp[t].y -= dy * f
    }
    for (let i = 0; i < n; i++) {
      disp[i].x += inward * (width / 2 - pos[i].x)
      disp[i].y += inward * (height / 2 - pos[i].y)
    }
    for (let i = 0; i < n; i++) {
      const d = Math.hypot(disp[i].x, disp[i].y) || 1
      const step = Math.min(temp, d)
      pos[i].x += (disp[i].x / d) * step
      pos[i].y += (disp[i].y / d) * step
      pos[i].x = Math.max(56, Math.min(width - 56, pos[i].x))
      pos[i].y = Math.max(56, Math.min(height - 56, pos[i].y))
    }
    temp *= 0.965
  }
  return pos
}

function computeSearchHighlight(entities, relationships, q) {
  const qq = q.trim().toLowerCase()
  if (!qq) return null
  const matchEnt = new Set()
  for (const e of entities) {
    const blob = `${e.name || ''} ${e.entityType || ''} ${e.description || ''}`.toLowerCase()
    if (blob.includes(qq)) {
      matchEnt.add(e.id)
    }
  }
  const nodeIds = new Set()
  matchEnt.forEach((id) => nodeIds.add(`e-${id}`))
  for (const r of relationships) {
    if (matchEnt.has(r.sourceEntityId) || matchEnt.has(r.targetEntityId)) {
      nodeIds.add(`e-${r.sourceEntityId}`)
      nodeIds.add(`e-${r.targetEntityId}`)
    }
  }
  const edgeIds = new Set()
  for (const r of relationships) {
    if (nodeIds.has(`e-${r.sourceEntityId}`) && nodeIds.has(`e-${r.targetEntityId}`)) {
      edgeIds.add(`r-${r.id}`)
    }
  }
  return { nodes: nodeIds, edges: edgeIds }
}

function relFromEdgeId(edgeId) {
  const rid = Number(String(edgeId).replace(/^r-/, ''))
  if (!Number.isFinite(rid)) return null
  return graphRels.value.find((r) => r.id === rid) || null
}

function rebuildGraph(remountFlow = true) {
  const ents = displayEntities.value
  const rels = displayRels.value
  const w = 960
  const h = 640
  if (!ents.length) {
    nodes.value = []
    edges.value = []
    nextTick(() => {
      if (remountFlow) flowMountKey.value += 1
    })
    return
  }
  let mode = layoutMode.value
  if (mode === 'force' && ents.length > 200) {
    mode = 'circle'
  }
  const pos =
    mode === 'force' ? layoutForcePositions(ents, rels, w, h) : layoutCirclePositions(ents, w, h)
  const hi = computeSearchHighlight(ents, rels, searchQuery.value)
  const tw = 90
  const th = 42
  nodes.value = ents.map((e, i) => {
    const id = `e-${e.id}`
    const dim = hi && !hi.nodes.has(id)
    const c = typeColor(e.entityType)
    const parts = [e.name, e.entityType, e.description].filter((x) => x && String(x).trim())
    const descTitle = parts.join(' — ').slice(0, 480)
    return {
      id,
      type: 'kg',
      position: { x: pos[i].x - tw, y: pos[i].y - th },
      data: {
        label: e.name || `ID:${e.id}`,
        sub: e.entityType || '',
        borderColor: c.border,
        bg: c.bg,
        descTitle: descTitle || undefined
      },
      style: dim ? { opacity: 0.2 } : { opacity: 1 }
    }
  })
  edges.value = rels.map((r) => {
    const id = `r-${r.id}`
    const kw = (r.keywords || '').trim()
    const short = kw.length > 28 ? `${kw.slice(0, 28)}…` : kw
    const isSelEdge = id === selectedEdgeId.value
    const dim = hi && !hi.edges.has(id) && !isSelEdge
    const lbl = hideEdgeLabels.value ? '' : short || t('rag.kgGraph.relationFallback')
    return {
      id,
      source: `e-${r.sourceEntityId}`,
      target: `e-${r.targetEntityId}`,
      label: lbl,
      markerEnd: { type: MarkerType.ArrowClosed, width: 14, height: 14 },
      style: {
        stroke: isSelEdge ? 'var(--el-color-primary)' : dim ? '#94a3b8' : '#475569',
        strokeWidth: isSelEdge ? 3.2 : dim ? 1 : 1.35,
        opacity: dim ? 0.15 : 1
      }
    }
  })
  nextTick(() => {
    if (remountFlow) {
      flowMountKey.value += 1
    }
  })
}

watch(
  () => [
    layoutMode.value,
    graphEntities.value,
    graphRels.value,
    egoFilter.value,
    typeFilterSelected.value,
    relKeywordQuery.value,
    displayEntities.value,
    displayRels.value
  ],
  () => rebuildGraph(true),
  { deep: true }
)

watch(searchQuery, () => rebuildGraph(false))

watch(selectedEdgeId, () => rebuildGraph(false))

watch(hideEdgeLabels, () => rebuildGraph(false))

watch(forceAllowed, (ok) => {
  if (!ok && layoutMode.value === 'force') {
    layoutMode.value = 'circle'
  }
})

async function loadGraph() {
  if (!kbId.value) {
    ElMessage.warning(t('rag.kgGraph.noKbId'))
    return
  }
  loading.value = true
  clearSelection()
  try {
    const req = {
      id: kbId.value,
      maxEntities: loadMaxEntities.value,
      maxRelationships: loadMaxRels.value
    }
    if (documentScopeId.value > 0) {
      req.documentId = documentScopeId.value
    }
    const res = await getKnowledgeGraph(req)
    if (res.code !== 0) {
      ElMessage.error(res.msg || t('rag.kgGraph.loadFail'))
      payload.value = null
      graphEntities.value = []
      graphRels.value = []
      nodes.value = []
      edges.value = []
      return
    }
    payload.value = res.data
    graphEntities.value = res.data?.entities || []
    graphRels.value = res.data?.relationships || []
    await nextTick()
    const validTypes = new Set(allEntityTypes.value)
    typeFilterSelected.value = typeFilterSelected.value.filter((x) => validTypes.has(x))
  } catch {
    ElMessage.error(t('rag.kgGraph.loadFail'))
    payload.value = null
    graphEntities.value = []
    graphRels.value = []
    nodes.value = []
    edges.value = []
  } finally {
    loading.value = false
  }
}

async function copyText(text) {
  const s = String(text || '').trim()
  if (!s) return
  try {
    await navigator.clipboard.writeText(s)
    ElMessage.success(t('rag.kgGraph.copied'))
  } catch {
    ElMessage.error(t('rag.kgGraph.copyFail'))
  }
}

function clearSelection() {
  selectedId.value = ''
  selectedEdgeId.value = ''
}

function onNodeClick({ node }) {
  selectedId.value = node?.id || ''
  selectedEdgeId.value = ''
}

function onNodeDoubleClick({ node }) {
  const id = node?.id || ''
  if (!id) return
  selectedId.value = id
  selectedEdgeId.value = ''
  nodeAutoFitNodeId.value = id
  nodeAutoFitPulse.value += 1
}

function onEdgeClick({ edge }) {
  selectedEdgeId.value = edge?.id || ''
  selectedId.value = ''
}

function onPaneClick() {
  clearSelection()
}

function onEdgeMouseEnter({ event, edge }) {
  const rel = relFromEdgeId(edge.id)
  if (!rel) return
  const ev = event
  edgeTip.rel = rel
  edgeTip.show = true
  edgeTip.x = (ev.clientX || 0) + 12
  edgeTip.y = (ev.clientY || 0) + 12
}

function onEdgeMouseMove({ event }) {
  if (!edgeTip.show) return
  const ev = event
  edgeTip.x = (ev.clientX || 0) + 12
  edgeTip.y = (ev.clientY || 0) + 12
}

function onEdgeMouseLeave() {
  edgeTip.show = false
  edgeTip.rel = null
}

function toggleCanvasFullscreen() {
  if (!screenfull.isEnabled) {
    ElMessage.warning(t('rag.kgGraph.fsUnavailable'))
    return
  }
  const el = canvasHostRef.value
  if (!el) return
  screenfull.toggle(el)
}

function onScreenfullChange() {
  isFs.value = screenfull.isFullscreen
}

useEventListener(window, 'keydown', (e) => {
  if (e.key === 'Escape') {
    clearSelection()
    edgeTip.show = false
    edgeTip.rel = null
  }
})

onMounted(() => {
  if (screenfull.isEnabled) {
    screenfull.on('change', onScreenfullChange)
    isFs.value = screenfull.isFullscreen
  }
})

onUnmounted(() => {
  if (screenfull.isEnabled) {
    screenfull.off('change', onScreenfullChange)
  }
})

function goBack() {
  if (fromDocs.value) {
    router.push({ name: 'ragDocuments', query: { id: kbId.value, name: route.query.name } })
  } else {
    router.push({ name: 'knowledgeBase' })
  }
}

function clearDocumentScope() {
  router.replace({
    name: 'ragKnowledgeGraph',
    query: {
      id: String(kbId.value),
      name: route.query.name || '',
      ...(fromDocs.value ? { from: 'docs' } : {})
    }
  })
}

function exportGraphJson() {
  if (!graphEntities.value.length) return
  const docId = payload.value?.scopeDocumentId || 0
  const body = {
    knowledgeBaseId: kbId.value,
    scopeDocumentId: docId,
    exportedAt: new Date().toISOString(),
    entities: graphEntities.value,
    relationships: graphRels.value
  }
  const blob = new Blob([JSON.stringify(body, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `kg-kb${kbId.value}${docId ? `-doc${docId}` : ''}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function resetGraphFilters() {
  typeFilterSelected.value = []
  relKeywordInput.value = ''
  relKeywordQuery.value = ''
  egoFilter.value = false
  searchQuery.value = ''
}

function exportVisibleGraphJson() {
  const ents = displayEntities.value
  const rels = displayRels.value
  if (!ents.length) {
    ElMessage.warning(t('rag.kgGraph.exportVisibleEmpty'))
    return
  }
  const docId = payload.value?.scopeDocumentId || 0
  const body = {
    knowledgeBaseId: kbId.value,
    scopeDocumentId: docId,
    exportedAt: new Date().toISOString(),
    exportKind: 'visibleSubgraph',
    entityCountLoaded: graphEntities.value.length,
    relationshipCountLoaded: graphRels.value.length,
    entities: ents,
    relationships: rels
  }
  const blob = new Blob([JSON.stringify(body, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `kg-kb${kbId.value}${docId ? `-doc${docId}` : ''}-visible.json`
  a.click()
  URL.revokeObjectURL(url)
}

watch(
  [kbId, documentScopeId],
  ([bid]) => {
    if (bid) {
      loadGraph()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.kg-viz-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  min-height: 480px;
}
.kg-viz-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}
.kg-viz-title {
  font-size: 16px;
  font-weight: 600;
}
.kg-stats {
  margin-right: 8px;
}
.kg-viz-subbar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}
.kg-search {
  width: min(320px, 100%);
  max-width: 100%;
}
.kg-type-filter {
  width: min(200px, 100%);
  max-width: 100%;
}
.kg-rel-filter {
  width: min(220px, 100%);
  max-width: 100%;
}
.kg-side-name-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}
.kg-side-name-row .kg-side-name {
  margin-bottom: 0;
  flex: 1;
  min-width: 0;
}
.kg-side-k-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}
.kg-legend {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 14px;
  margin-bottom: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.kg-legend-title {
  font-weight: 500;
  color: var(--el-text-color-regular);
}
.kg-legend-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.kg-legend-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.kg-viz-alert {
  margin-bottom: 10px;
}
.kg-viz-body {
  flex: 1;
  display: flex;
  gap: 12px;
  min-height: 0;
}
.kg-viz-canvas {
  flex: 1;
  min-width: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  background: var(--el-fill-color-blank);
  position: relative;
}
.kg-vue-flow {
  width: 100%;
  height: 100%;
  min-height: 400px;
}
.kg-viz-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  height: 100%;
  min-height: 400px;
  color: var(--el-text-color-secondary);
  font-size: 14px;
  text-align: center;
  padding: 16px;
  box-sizing: border-box;
}
.kg-empty-text {
  margin: 0;
  max-width: 420px;
  line-height: 1.55;
}
.kg-node {
  padding: 8px 12px;
  border-radius: 8px;
  border: 2px solid #64748b;
  min-width: 100px;
  max-width: 180px;
  box-shadow: 0 1px 3px rgb(15 23 42 / 10%);
  cursor: pointer;
  transition:
    opacity 0.2s,
    border-color 0.15s,
    box-shadow 0.15s;
}
.kg-node.is-dimmed {
  filter: grayscale(0.35);
}
.kg-node.is-selected {
  box-shadow: 0 0 0 3px var(--el-color-primary-light-5);
  z-index: 2;
}
.kg-node-label {
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
  word-break: break-word;
  line-height: 1.35;
}
.kg-node-type {
  margin-top: 4px;
  font-size: 11px;
  color: #475569;
  word-break: break-all;
}
.kg-viz-side {
  width: 300px;
  flex-shrink: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  padding: 12px;
  background: var(--el-fill-color-blank);
  overflow: auto;
}
.kg-side-head-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}
.kg-side-head {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.kg-side-name {
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 8px;
  word-break: break-word;
}
.kg-side-row {
  font-size: 13px;
  margin-bottom: 8px;
  line-height: 1.45;
  word-break: break-word;
}
.kg-side-block {
  margin-top: 10px;
}
.kg-side-k {
  color: var(--el-text-color-secondary);
  margin-right: 6px;
}
.kg-type-pill {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid;
  font-size: 12px;
  background: var(--el-fill-color-light);
}
.kg-side-desc {
  font-size: 13px;
  line-height: 1.5;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
}
.kg-side-placeholder {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.6;
  padding-top: 8px;
}
.kg-scope-popover {
  padding: 4px 0;
}
.kg-scope-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}
.kg-scope-label {
  flex: 1;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.kg-scope-check {
  display: block;
  margin: 4px 0 12px;
}
.kg-scope-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.45;
  margin: 0 0 8px;
}
.kg-scope-export {
  width: 100%;
  margin-bottom: 8px;
}
.kg-scope-reload {
  width: 100%;
}
</style>

<style>
.kg-edge-tip {
  position: fixed;
  z-index: 8000;
  max-width: min(360px, calc(100vw - 24px));
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--el-bg-color-overlay, #ffffff);
  border: 1px solid var(--el-border-color-lighter, #e2e8f0);
  box-shadow: 0 8px 24px rgb(15 23 42 / 18%);
  pointer-events: none;
  font-size: 12px;
  line-height: 1.45;
  color: var(--el-text-color-primary, #0f172a);
}
.kg-edge-tip-line {
  font-weight: 600;
  margin-bottom: 8px;
  word-break: break-word;
}
.kg-edge-tip-arrow {
  margin: 0 6px;
  opacity: 0.55;
  font-weight: 400;
}
.kg-edge-tip-k {
  word-break: break-word;
}
.kg-edge-tip-block {
  margin-top: 8px;
}
.kg-edge-tip-h {
  color: var(--el-text-color-secondary, #64748b);
  font-size: 11px;
  margin-bottom: 2px;
}
.kg-edge-tip-t {
  white-space: pre-wrap;
  word-break: break-word;
}
.kg-viz-canvas:fullscreen {
  min-height: 100vh !important;
  border-radius: 0;
  display: flex;
  flex-direction: column;
}
.kg-viz-canvas:fullscreen .kg-vue-flow {
  flex: 1;
  min-height: calc(100vh - 8px) !important;
}
</style>
