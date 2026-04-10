<template>
  <div class="p-2">
    <warning-bar :title="$t('rag.retrieval.warningBar')" />
    <div class="mt-4 bg-white dark:bg-slate-900 rounded-lg p-4 shadow-sm border border-slate-100 dark:border-slate-800">
      <el-form label-width="100px" class="max-w-4xl">
        <el-form-item :label="$t('rag.retrieval.labelKb')" required>
          <el-select
            v-model="selectedKbIds"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            :placeholder="$t('rag.retrieval.kbPh')"
            class="w-full"
            style="width: 100%"
          >
            <el-option
              v-for="kb in kbList"
              :key="kb.ID"
              :label="kb.name"
              :value="kb.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.retrieval.labelQuery')" required>
          <el-input
            v-model="query"
            type="textarea"
            :rows="4"
            :placeholder="$t('rag.retrieval.queryPh')"
            maxlength="2000"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('rag.retrieval.labelTopN')">
          <el-input-number v-model="topN" :min="1" :max="50" :step="1" />
          <span class="text-xs text-gray-500 ml-2">{{ $t('rag.retrieval.topNHint') }}</span>
        </el-form-item>
        <el-form-item :label="$t('rag.retrieval.labelPoolTopK')">
          <el-input-number v-model="poolTopK" :min="0" :max="50" :step="1" />
          <span class="text-xs text-gray-500 ml-2">{{ $t('rag.retrieval.poolTopKHint') }}</span>
        </el-form-item>
        <el-form-item :label="$t('rag.retrieval.labelMode')">
          <el-select v-model="retrieveMode" clearable :placeholder="$t('rag.retrieval.modeKbDefault')" class="w-full" style="width: 100%">
            <el-option :label="$t('rag.retrieval.modeKbDefault')" value="" />
            <el-option :label="$t('rag.kb.retrieverVector')" value="vector" />
            <el-option :label="$t('rag.kb.retrieverKeyword')" value="keyword" />
            <el-option :label="$t('rag.kb.retrieverLocal')" value="local" />
            <el-option :label="$t('rag.kb.retrieverGlobal')" value="global" />
            <el-option :label="$t('rag.kb.retrieverHybrid')" value="hybrid" />
            <el-option :label="$t('rag.kb.retrieverMix')" value="mix" />
            <el-option :label="$t('rag.kb.retrieverBypass')" value="bypass" />
            <el-option :label="$t('rag.kb.retrieverPageindex')" value="pageindex" />
          </el-select>
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.retrieval.modeHint') }}</div>
        </el-form-item>
        <el-collapse class="mt-1 border-0">
          <el-collapse-item :title="$t('rag.retrieval.advTitle')" name="adv">
            <el-form-item :label="$t('rag.retrieval.labelRerank')">
              <el-select v-model="rerankMode" clearable class="w-full" style="width: 100%">
                <el-option :label="$t('rag.retrieval.rerankKb')" value="" />
                <el-option :label="$t('rag.retrieval.rerankOn')" value="on" />
                <el-option :label="$t('rag.retrieval.rerankOff')" value="off" />
              </el-select>
              <div class="text-xs text-gray-400 mt-1">{{ $t('rag.retrieval.rerankHint') }}</div>
            </el-form-item>
            <el-form-item :label="$t('rag.retrieval.labelTocEnhance')">
              <el-select v-model="tocEnhanceChoice" clearable class="w-full" style="width: 100%">
                <el-option :label="$t('rag.retrieval.tocEnhanceDefault')" value="" />
                <el-option :label="$t('rag.retrieval.tocEnhanceOn')" value="on" />
                <el-option :label="$t('rag.retrieval.tocEnhanceOff')" value="off" />
              </el-select>
              <div class="text-xs text-gray-400 mt-1">{{ $t('rag.retrieval.tocEnhanceHint') }}</div>
            </el-form-item>
            <el-form-item :label="$t('rag.retrieval.labelHlKw')">
              <el-select
                v-model="hlKeywords"
                multiple
                filterable
                allow-create
                default-first-option
                collapse-tags
                collapse-tags-tooltip
                :placeholder="$t('rag.retrieval.kwPh')"
                class="w-full"
                style="width: 100%"
              />
            </el-form-item>
            <el-form-item :label="$t('rag.retrieval.labelLlKw')">
              <el-select
                v-model="llKeywords"
                multiple
                filterable
                allow-create
                default-first-option
                collapse-tags
                collapse-tags-tooltip
                :placeholder="$t('rag.retrieval.kwPh')"
                class="w-full"
                style="width: 100%"
              />
            </el-form-item>
            <div class="text-xs text-gray-400 mb-2">{{ $t('rag.retrieval.kwHint') }}</div>
          </el-collapse-item>
        </el-collapse>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="runRetrieve">{{ $t('rag.retrieval.btnRetrieve') }}</el-button>
          <el-button @click="clearResults" :disabled="!results.length">{{ $t('rag.retrieval.btnClear') }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div v-if="results.length" class="mt-4 lrag-table-box">
      <div class="text-sm text-gray-600 dark:text-gray-400 mb-2">
        {{ $t('rag.retrieval.resultSummary', { n: results.length }) }}
      </div>
      <el-table :data="results" stripe style="width: 100%">
        <el-table-column :label="$t('rag.retrieval.colId')" prop="index" width="56" align="center" />
        <el-table-column :label="$t('rag.retrieval.colRelevance')" width="100" align="center">
          <template #default="{ row }">
            <span v-if="row.score != null && row.score !== ''">{{ formatScore(row.score) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>
        </el-table-column>
        <el-table-column :label="$t('rag.retrieval.colDoc')" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button
              v-if="documentIdOf(row)"
              type="primary"
              link
              class="!p-0 !h-auto font-normal"
              @click="handlePreview(row)"
            >
              {{ row.docName || row.title || $t('rag.retrieval.previewDocLink') }}
            </el-button>
            <span v-else>{{ row.docName || row.title || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="$t('rag.retrieval.colChunkNo')" width="88" align="center">
          <template #default="{ row }">
            {{ row.chunkIndex != null ? row.chunkIndex : '—' }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('rag.retrieval.colSourceSummary')" min-width="200" show-overflow-tooltip prop="sourceLabel" />
        <el-table-column :label="$t('rag.retrieval.colChunkBody')" min-width="320">
          <template #default="{ row }">
            <div class="text-sm whitespace-pre-wrap max-h-40 overflow-y-auto text-slate-700 dark:text-slate-200">
              {{ row.content }}
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('rag.docs.colActions')" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <template v-if="documentIdOf(row)">
              <el-button type="primary" link size="small" @click="handlePreview(row)">{{ $t('rag.docs.preview') }}</el-button>
              <el-button type="primary" link size="small" @click="handleDownload(row)">{{ $t('rag.docs.download') }}</el-button>
            </template>
            <span v-else class="text-xs text-gray-400">—</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-empty v-else-if="searched && !loading" :description="$t('rag.retrieval.emptyHint')" class="mt-8" />

    <!-- 与知识库「文档管理」相同的预览弹窗 -->
    <el-dialog
      v-model="previewVisible"
      :title="previewTitle"
      width="80%"
      top="5vh"
      destroy-on-close
      class="document-preview-dialog"
      @closed="onPreviewClosed"
    >
      <div v-loading="previewLoading" class="preview-content">
        <template v-if="previewType === 'pdf'">
          <iframe v-if="previewUrl" :src="previewUrl" class="preview-iframe" />
        </template>
        <template v-else-if="previewType === 'image'">
          <img v-if="previewUrl" :src="previewUrl" class="preview-img" />
        </template>
        <template v-else-if="previewType === 'text'">
          <pre class="preview-text">{{ previewText }}</pre>
        </template>
        <template v-else-if="previewType === 'unsupported'">
          <el-empty :description="$t('rag.docs.previewUnsupported')" />
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
  import { getKnowledgeBaseList, retrieveKnowledgeChunks, downloadDocument } from '@/api/rag'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { formatRagRelevanceScore } from '@/utils/format'
  import { ref, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'RagDocumentRetrieval' })

  const { t } = useI18n()

  const kbList = ref([])
  const selectedKbIds = ref([])
  const query = ref('')
  const topN = ref(8)
  /** LightningRAG top_k：候选池大于返回条数时先多取再截断；0 表示不传 */
  const poolTopK = ref(0)
  /** 覆盖本次请求的 LightningRAG 风格检索模式；空则使用各知识库配置 */
  const retrieveMode = ref('')
  /** '' | 'on' | 'off' → enableRerank；空=跟随知识库 */
  const rerankMode = ref('')
  /** PageIndex tocEnhance：''=不传（服务端默认混合）；on/off 显式 true/false */
  const tocEnhanceChoice = ref('')
  const hlKeywords = ref([])
  const llKeywords = ref([])
  const loading = ref(false)
  const results = ref([])
  const searched = ref(false)

  const loadKbList = async () => {
    try {
      const res = await getKnowledgeBaseList({ page: 1, pageSize: 500 })
      if (res.code === 0) {
        kbList.value = res.data?.list || []
      }
    } catch (e) {
      console.warn(e)
    }
  }

  const formatScore = formatRagRelevanceScore

  const runRetrieve = async () => {
    const q = (query.value || '').trim()
    if (!q) {
      ElMessage.warning(t('rag.retrieval.warnNeedQuery'))
      return
    }
    if (!selectedKbIds.value.length) {
      ElMessage.warning(t('rag.retrieval.warnNeedKb'))
      return
    }
    loading.value = true
    searched.value = true
    results.value = []
    try {
      const payload = {
        knowledgeBaseIds: selectedKbIds.value,
        query: q,
        topN: topN.value || 8
      }
      const pk = Number(poolTopK.value)
      if (Number.isFinite(pk) && pk > 0 && pk > (topN.value || 8)) {
        payload.topK = pk
      }
      if ((retrieveMode.value || '').trim()) {
        payload.mode = retrieveMode.value.trim()
      }
      if (rerankMode.value === 'on') {
        payload.enableRerank = true
      } else if (rerankMode.value === 'off') {
        payload.enableRerank = false
      }
      if (tocEnhanceChoice.value === 'on') {
        payload.tocEnhance = true
      } else if (tocEnhanceChoice.value === 'off') {
        payload.tocEnhance = false
      }
      const hk = (hlKeywords.value || []).map((s) => String(s).trim()).filter(Boolean)
      if (hk.length) {
        payload.hlKeywords = hk
      }
      const lk = (llKeywords.value || []).map((s) => String(s).trim()).filter(Boolean)
      if (lk.length) {
        payload.llKeywords = lk
      }
      const res = await retrieveKnowledgeChunks(payload)
      if (res.code === 0) {
        results.value = Array.isArray(res.data) ? res.data : []
        if (!results.value.length) {
          ElMessage.info(t('rag.retrieval.infoNoResults'))
        }
      } else {
        ElMessage.error(res.msg || t('rag.retrieval.errRetrieve'))
      }
    } catch (e) {
      ElMessage.error(e?.message || t('rag.retrieval.errRequest'))
    } finally {
      loading.value = false
    }
  }

  const clearResults = () => {
    results.value = []
    searched.value = false
  }

  /** 检索结果里的 documentId（后端 metadata，PageIndex 等场景可能无） */
  const documentIdOf = (row) => {
    const v = row?.documentId
    if (v == null || v === '') return 0
    const n = typeof v === 'number' ? v : parseInt(String(v), 10)
    return Number.isFinite(n) && n > 0 ? n : 0
  }

  const PREVIEWABLE_IMAGE = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp']
  const PREVIEWABLE_TEXT = ['txt', 'md', 'json', 'xml', 'csv', 'log']

  const getFileExt = (row) => {
    const name = row?.docName || row?.title || ''
    if (name) {
      const i = name.lastIndexOf('.')
      if (i >= 0) return name.slice(i + 1).toLowerCase()
    }
    return ''
  }

  const previewVisible = ref(false)
  const previewLoading = ref(false)
  const previewTitle = ref('')
  const previewUrl = ref('')
  const previewText = ref('')
  const previewType = ref('')

  const handlePreview = async (row) => {
    const id = documentIdOf(row)
    if (!id) {
      ElMessage.warning(t('rag.retrieval.warnNoPreviewDoc'))
      return
    }
    previewVisible.value = true
    previewTitle.value = row.docName || row.title || t('rag.docs.previewTitleFallback')
    previewLoading.value = true
    previewUrl.value = ''
    previewText.value = ''
    previewType.value = 'unsupported'
    const ext = getFileExt(row)
    try {
      const blob = await downloadDocument(id, { preview: true })
      if (!blob || !(blob instanceof Blob)) {
        ElMessage.error(t('rag.docs.previewFail'))
        return
      }
      const mime = (blob.type || '').toLowerCase()
      const isPdf = ext === 'pdf' || mime.includes('pdf')
      const isImage = PREVIEWABLE_IMAGE.includes(ext) || mime.startsWith('image/')
      const isText =
        PREVIEWABLE_TEXT.includes(ext) ||
        mime.startsWith('text/') ||
        mime === 'application/json' ||
        mime === 'application/xml'
      if (isPdf) {
        previewType.value = 'pdf'
        previewUrl.value = URL.createObjectURL(blob)
      } else if (isImage) {
        previewType.value = 'image'
        previewUrl.value = URL.createObjectURL(blob)
      } else if (isText) {
        previewType.value = 'text'
        previewText.value = await blob.text()
      }
    } catch (e) {
      ElMessage.error(e?.message || t('rag.docs.previewFail'))
    } finally {
      previewLoading.value = false
    }
  }

  const onPreviewClosed = () => {
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value)
      previewUrl.value = ''
    }
  }

  const handleDownload = async (row) => {
    const id = documentIdOf(row)
    if (!id) {
      ElMessage.warning(t('rag.retrieval.warnNoDownloadDoc'))
      return
    }
    try {
      const blob = await downloadDocument(id)
      if (!blob || !(blob instanceof Blob)) {
        ElMessage.error(t('rag.docs.downloadFail'))
        return
      }
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = row.docName || row.title || 'document'
      a.click()
      URL.revokeObjectURL(url)
    } catch (e) {
      ElMessage.error(e?.message || t('rag.docs.downloadFail'))
    }
  }

  onMounted(() => {
    loadKbList()
  })
</script>

<style scoped>
.preview-content {
  min-height: 60vh;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}
.preview-iframe {
  width: 100%;
  min-height: 70vh;
  border: none;
}
.preview-img {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
}
.preview-text {
  width: 100%;
  max-height: 70vh;
  overflow: auto;
  padding: 12px;
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}
</style>
