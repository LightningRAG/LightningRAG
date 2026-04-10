<template>
  <div>
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <el-button icon="arrow-left" @click="goBack">{{ $t('rag.docs.back') }}</el-button>
      <span class="text-lg font-medium">{{ $t('rag.docs.kbPrefix') }}{{ kbName }}</span>
      <el-button
        v-if="kbId"
        type="primary"
        plain
        icon="connection"
        @click="goKnowledgeGraph"
      >
        {{ $t('rag.kb.knowledgeGraph') }}
      </el-button>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list flex flex-wrap items-center gap-2">
        <el-upload
          :auto-upload="true"
          :http-request="handleUpload"
          :show-file-list="false"
        >
          <el-button type="primary" icon="upload">{{ $t('rag.docs.upload') }}</el-button>
        </el-upload>
        <el-button
          type="danger"
          :disabled="!selectedDocumentIdList.length"
          @click="handleBatchDelete"
        >
          {{ $t('rag.docs.batchDelete') }}
        </el-button>
        <el-button
          type="warning"
          :disabled="!selectedDocumentIdList.length"
          @click="handleBatchReindex"
        >
          {{ $t('rag.docs.batchReslice') }}
        </el-button>
        <el-button
          :disabled="!selectedDocumentIdList.length"
          @click="handleBatchCancelIndexing"
        >
          {{ $t('rag.docs.batchCancel') }}
        </el-button>
        <el-button
          type="success"
          :disabled="!selectedDocumentIdList.length"
          @click="handleBatchSetRetrieval(true)"
        >
          {{ $t('rag.docs.batchEnableSearch') }}
        </el-button>
        <el-button
          :disabled="!selectedDocumentIdList.length"
          @click="handleBatchSetRetrieval(false)"
        >
          {{ $t('rag.docs.batchDisableSearch') }}
        </el-button>
      </div>
      <el-table
        ref="tableRef"
        :data="tableData"
        style="width: 100%"
        row-key="ID"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="42" :reserve-selection="true" />
        <el-table-column align="center" label="" width="60">
          <template #default="scope">
            <el-image
              v-if="scope.row.thumbnail"
              :src="scope.row.thumbnail"
              fit="cover"
              class="doc-thumbnail"
              :preview-src-list="scope.row.thumbnail ? [scope.row.thumbnail] : []"
            />
            <div v-else class="doc-type-icon" :class="'icon-' + (scope.row.fileType || 'file')">
              {{ (scope.row.fileType || '?').toUpperCase().slice(0, 3) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.docs.colFileName')" prop="name" min-width="200" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('rag.docs.colType')" width="80">
          <template #default="scope">
            <el-tag size="small">{{ scope.row.fileType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.docs.colSize')" width="100">
          <template #default="scope">
            {{ formatSize(scope.row.fileSize) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.docs.colStatus')" prop="status" width="100">
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'completed' ? 'success' : scope.row.status === 'failed' ? 'danger' : scope.row.status === 'cancelled' ? 'info' : 'warning'"
            >
              {{ statusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="center" :label="$t('rag.docs.colInSearch')" width="100">
          <template #default="scope">
            <el-switch
              :model-value="isRetrievalEnabled(scope.row)"
              :disabled="scope.row.status === 'processing'"
              @change="(v) => onRetrievalToggle(scope.row, v)"
            />
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.docs.colChunksTokens')" width="120">
          <template #default="scope">
            <span>{{ scope.row.chunkCount || 0 }}</span>
            <span v-if="scope.row.tokenCount" class="text-gray-400"> / {{ scope.row.tokenCount }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.docs.colUploadedAt')" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.docs.colActions')" width="420" fixed="right">
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 'completed' && scope.row.chunkCount > 0"
              type="primary"
              link
              size="small"
              @click="handleViewChunks(scope.row)"
            >
              {{ $t('rag.docs.actionSlice') }}
            </el-button>
            <el-button
              v-if="scope.row.status === 'completed' && scope.row.chunkCount > 0"
              type="primary"
              link
              size="small"
              @click="goDocumentKnowledgeGraph(scope.row)"
            >
              {{ $t('rag.docs.docKgGraph') }}
            </el-button>
            <el-button
              v-if="scope.row.storagePath"
              type="primary"
              link
              size="small"
              @click="handlePreview(scope.row)"
            >
              {{ $t('rag.docs.preview') }}
            </el-button>
            <el-button
              v-if="scope.row.storagePath"
              type="primary"
              link
              size="small"
              @click="handleDownload(scope.row)"
            >
              {{ $t('rag.docs.download') }}
            </el-button>
            <el-button
              v-if="scope.row.status === 'failed' || scope.row.status === 'completed' || scope.row.status === 'cancelled'"
              type="warning"
              link
              size="small"
              :loading="retryingId === scope.row.ID"
              @click="handleRetry(scope.row)"
            >
              {{ scope.row.status === 'failed' || scope.row.status === 'cancelled' ? $t('rag.docs.retry') : $t('rag.docs.rebuildSlice') }}
            </el-button>
            <el-tooltip
              v-if="scope.row.status === 'failed' && scope.row.errorMsg"
              :content="scope.row.errorMsg"
              placement="top"
              :show-after="300"
            >
              <el-button type="info" link size="small">{{ $t('rag.docs.errDetail') }}</el-button>
            </el-tooltip>
            <el-button
              type="danger"
              link
              size="small"
              @click="handleDelete(scope.row)"
            >
              {{ $t('rag.docs.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="lrag-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="(v) => { page = v; loadDocuments() }"
          @size-change="(v) => { pageSize = v; page = 1; loadDocuments() }"
        />
      </div>
    </div>

    <!-- 文档预览弹窗 -->
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

    <!-- 切片列表弹窗 -->
    <el-dialog
      v-model="chunksVisible"
      :title="chunksTitle"
      width="75%"
      top="5vh"
      destroy-on-close
    >
      <div v-loading="chunksLoading" class="chunks-container">
        <el-table :data="chunksData" style="width: 100%" row-key="ID" border>
          <el-table-column align="center" :label="$t('common.colIndexNo')" width="70">
            <template #default="scope">
              {{ scope.row.chunkIndex }}
            </template>
          </el-table-column>
          <el-table-column align="left" :label="$t('rag.docs.chunkColContent')" min-width="500">
            <template #default="scope">
              <div class="chunk-content-cell">{{ scope.row.content }}</div>
            </template>
          </el-table-column>
          <el-table-column align="center" :label="$t('rag.docs.colActions')" width="100">
            <template #default="scope">
              <el-button type="primary" link size="small" @click="handleEditChunk(scope.row)">
                {{ $t('rag.docs.edit') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="lrag-pagination" style="margin-top: 12px;">
          <el-pagination
            :current-page="chunksPage"
            :page-size="chunksPageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="chunksTotal"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="(v) => { chunksPage = v; loadChunks() }"
            @size-change="(v) => { chunksPageSize = v; chunksPage = 1; loadChunks() }"
          />
        </div>
      </div>
    </el-dialog>

    <!-- 编辑切片弹窗 -->
    <el-dialog
      v-model="editChunkVisible"
      :title="$t('rag.docs.editChunkTitle')"
      width="60%"
      top="10vh"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="$t('rag.docs.chunkIndex')">
          <el-input :model-value="'#' + editChunkData.chunkIndex" disabled />
        </el-form-item>
        <el-form-item :label="$t('rag.docs.chunkContent')">
          <el-input
            v-model="editChunkData.content"
            type="textarea"
            :autosize="{ minRows: 6, maxRows: 20 }"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editChunkVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" :loading="editChunkSaving" @click="handleSaveChunk">{{ $t('rag.docs.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import {
    uploadDocument,
    getDocumentList,
    deleteDocument,
    retryDocument,
    downloadDocument,
    getDocumentChunks,
    updateDocumentChunk,
    batchDeleteDocuments,
    batchReindexDocuments,
    batchCancelDocumentIndexing,
    batchSetDocumentRetrieval
  } from '@/api/rag'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate } from '@/utils/format'

  defineOptions({ name: 'RagDocuments' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  const kbId = computed(() => Number(route.query.id) || Number(route.params.id) || 0)
  const kbName = computed(
    () => route.query.name || route.params.name || t('rag.docs.listTitleFallback')
  )

  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const retryingId = ref(null)
  const tableRef = ref(null)
  /** 当前页表格勾选行（随数据刷新会失效，仅作展示用） */
  const selectedRows = ref([])
  /** 跨页、跨定时刷新的勾选文档 ID（批量操作以此为准） */
  const selectedDocumentIdList = ref([])
  let selectionSyncSuspended = false

  const statusLabel = (s) => {
    if (s === 'completed') return t('rag.docs.statusCompleted')
    if (s === 'failed') return t('rag.docs.statusFailed')
    if (s === 'cancelled') return t('rag.docs.statusCancelled')
    return t('rag.docs.statusProcessing')
  }

  const isRetrievalEnabled = (row) => row.retrievalEnabled !== false

  const onSelectionChange = (rows) => {
    selectedRows.value = rows || []
    if (selectionSyncSuspended) return
    const pageIds = tableData.value.map((r) => r.ID).filter(Boolean)
    const onPage = new Set((rows || []).map((r) => r.ID).filter(Boolean))
    const next = new Set(selectedDocumentIdList.value)
    pageIds.forEach((id) => {
      if (!onPage.has(id)) next.delete(id)
    })
    onPage.forEach((id) => next.add(id))
    selectedDocumentIdList.value = Array.from(next)
  }

  const selectedDocumentIds = () => [...selectedDocumentIdList.value]

  /** 列表数据替换后（含定时静默刷新）按 ID 恢复勾选，避免勾选丢失 */
  const restoreTableSelectionFromIds = async () => {
    const ids = new Set(selectedDocumentIdList.value)
    selectionSyncSuspended = true
    await nextTick()
    const table = tableRef.value
    if (table && tableData.value.length) {
      tableData.value.forEach((row) => {
        table.toggleRowSelection(row, ids.has(row.ID))
      })
    }
    selectedRows.value = tableData.value.filter((r) => ids.has(r.ID))
    await nextTick()
    selectionSyncSuspended = false
  }

  const clearDocumentSelection = () => {
    selectedDocumentIdList.value = []
    selectedRows.value = []
    tableRef.value?.clearSelection()
  }

  const formatSize = (bytes) => {
    if (!bytes || bytes === 0) return '-'
    const units = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
  }

  // 预览相关
  const previewVisible = ref(false)
  const previewLoading = ref(false)
  const previewTitle = ref('')
  const previewUrl = ref('')
  const previewText = ref('')
  const previewType = ref('') // pdf | image | text | unsupported

  const goBack = () => router.push({ name: 'knowledgeBase' })

  const goKnowledgeGraph = () => {
    if (!kbId.value) return
    router.push({
      name: 'ragKnowledgeGraph',
      query: { id: kbId.value, name: String(kbName.value || ''), from: 'docs' }
    })
  }

  const goDocumentKnowledgeGraph = (row) => {
    if (!kbId.value || !row?.ID) return
    router.push({
      name: 'ragKnowledgeGraph',
      query: {
        id: kbId.value,
        name: String(kbName.value || ''),
        from: 'docs',
        documentId: row.ID
      }
    })
  }

  /** 文档列表轮询间隔（切片状态等后台任务） */
  const DOC_LIST_POLL_MS = 5000
  let documentListPollTimer = null

  const stopDocumentListPolling = () => {
    if (documentListPollTimer != null) {
      clearInterval(documentListPollTimer)
      documentListPollTimer = null
    }
  }

  const startDocumentListPolling = () => {
    stopDocumentListPolling()
    documentListPollTimer = setInterval(() => {
      if (kbId.value) {
        loadDocuments({ silent: true })
      }
    }, DOC_LIST_POLL_MS)
  }

  const loadDocuments = async (options = {}) => {
    const silent = options.silent === true
    if (!kbId.value) {
      if (!silent) ElMessage.warning(t('rag.docs.warnNoKbId'))
      return
    }
    const res = await getDocumentList({
      knowledgeBaseId: kbId.value,
      page: page.value,
      pageSize: pageSize.value
    })
    if (res?.code === 0) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
      await restoreTableSelectionFromIds()
    }
  }

  watch(kbId, (val) => {
    if (val) {
      page.value = 1
      clearDocumentSelection()
      loadDocuments()
    }
  })

  const handleUpload = async ({ file }) => {
    const formData = new FormData()
    formData.append('knowledgeBaseId', String(kbId.value))
    formData.append('file', file)
    const res = await uploadDocument(formData)
    if (res?.code === 0) {
      ElMessage.success(t('rag.docs.uploadOk'))
      loadDocuments()
    } else {
      ElMessage.error(res?.msg || t('rag.docs.uploadFail'))
    }
  }

  const PREVIEWABLE_IMAGE = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp']
  const PREVIEWABLE_TEXT = ['txt', 'md', 'json', 'xml', 'csv', 'log']

  const getFileExt = (row) => {
    if (row?.name) {
      const i = row.name.lastIndexOf('.')
      if (i >= 0) return row.name.slice(i + 1).toLowerCase()
    }
    return (row?.fileType || '').toLowerCase()
  }

  const handlePreview = async (row) => {
    previewVisible.value = true
    previewTitle.value = row.name || t('rag.docs.previewTitleFallback')
    previewLoading.value = true
    previewUrl.value = ''
    previewText.value = ''
    previewType.value = 'unsupported'
    const ext = getFileExt(row)
    try {
      const blob = await downloadDocument(row.ID, { preview: true })
      if (!blob || !(blob instanceof Blob)) {
        ElMessage.error(t('rag.docs.previewFail'))
        return
      }
      if (ext === 'pdf') {
        previewType.value = 'pdf'
        previewUrl.value = URL.createObjectURL(blob)
      } else if (PREVIEWABLE_IMAGE.includes(ext)) {
        previewType.value = 'image'
        previewUrl.value = URL.createObjectURL(blob)
      } else if (PREVIEWABLE_TEXT.includes(ext)) {
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
    try {
      const blob = await downloadDocument(row.ID)
      if (!blob || !(blob instanceof Blob)) {
        ElMessage.error(t('rag.docs.downloadFail'))
        return
      }
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = row.name || 'document'
      a.click()
      URL.revokeObjectURL(url)
    } catch (e) {
      ElMessage.error(e?.message || t('rag.docs.downloadFail'))
    }
  }

  const handleRetry = async (row) => {
    retryingId.value = row.ID
    try {
      const res = await retryDocument({ id: row.ID })
      if (res?.code === 0) {
        ElMessage.success(t('rag.docs.sliceSubmitted'))
        loadDocuments()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.retryFail'))
      }
    } catch (e) {
      ElMessage.error(e?.message || t('rag.docs.retryFail'))
    } finally {
      retryingId.value = null
    }
  }

  const onRetrievalToggle = async (row, enabled) => {
    const res = await batchSetDocumentRetrieval({
      knowledgeBaseId: kbId.value,
      documentIds: [row.ID],
      enabled
    })
    if (res?.code === 0) {
      ElMessage.success(t('rag.docs.updateOk'))
      loadDocuments()
    } else {
      ElMessage.error(res?.msg || t('rag.docs.updateFail'))
    }
  }

  const handleBatchDelete = () => {
    const ids = selectedDocumentIds()
    if (!ids.length) return
    ElMessageBox.confirm(
      t('rag.docs.batchDeleteConfirm', { n: ids.length }),
      t('rag.docs.batchDeleteTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      const res = await batchDeleteDocuments({ knowledgeBaseId: kbId.value, documentIds: ids })
      if (res?.code === 0) {
        ElMessage.success(t('rag.docs.deleteOk'))
        clearDocumentSelection()
        loadDocuments()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.deleteFail'))
      }
    }).catch(() => {})
  }

  const handleBatchReindex = () => {
    const ids = selectedDocumentIds()
    if (!ids.length) return
    ElMessageBox.confirm(t('rag.docs.batchResliceBody'), t('rag.docs.batchResliceTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'info'
    }).then(async () => {
      const res = await batchReindexDocuments({ knowledgeBaseId: kbId.value, documentIds: ids })
      if (res?.code === 0) {
        ElMessage.success(res.msg || t('rag.docs.submitOk'))
        clearDocumentSelection()
        loadDocuments()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.submitFail'))
      }
    }).catch(() => {})
  }

  const handleBatchCancelIndexing = () => {
    const ids = selectedDocumentIds()
    if (!ids.length) return
    ElMessageBox.confirm(t('rag.docs.batchCancelBody'), t('rag.docs.batchCancelTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await batchCancelDocumentIndexing({ knowledgeBaseId: kbId.value, documentIds: ids })
      if (res?.code === 0) {
        ElMessage.success(t('rag.docs.cancelOk'))
        clearDocumentSelection()
        loadDocuments()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.cancelFail'))
      }
    }).catch(() => {})
  }

  const handleBatchSetRetrieval = (enabled) => {
    const ids = selectedDocumentIds()
    if (!ids.length) return
    const body = enabled ? t('rag.docs.batchEnableSearchBody') : t('rag.docs.batchDisableSearchBody')
    const title = enabled ? t('rag.docs.batchEnableTitle') : t('rag.docs.batchDisableTitle')
    ElMessageBox.confirm(body, title, {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'info'
    }).then(async () => {
      const res = await batchSetDocumentRetrieval({ knowledgeBaseId: kbId.value, documentIds: ids, enabled })
      if (res?.code === 0) {
        ElMessage.success(t('rag.docs.updateOk'))
        clearDocumentSelection()
        loadDocuments()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.updateFail'))
      }
    }).catch(() => {})
  }

  const handleDelete = (row) => {
    ElMessageBox.confirm(
      t('rag.docs.singleDeleteConfirm', { name: row.name }),
      t('rag.docs.singleDeleteTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      const res = await deleteDocument({ id: row.ID })
      if (res?.code === 0) {
        ElMessage.success(t('rag.docs.deleteOk'))
        selectedDocumentIdList.value = selectedDocumentIdList.value.filter((id) => id !== row.ID)
        loadDocuments()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.deleteFail'))
      }
    }).catch(() => {})
  }

  // ========== 切片查看与编辑 ==========
  const chunksVisible = ref(false)
  const chunksLoading = ref(false)
  const chunksTitle = ref('')
  const chunksData = ref([])
  const chunksTotal = ref(0)
  const chunksPage = ref(1)
  const chunksPageSize = ref(10)
  const currentDocId = ref(null)

  const handleViewChunks = (row) => {
    currentDocId.value = row.ID
    chunksTitle.value = t('rag.docs.chunksTitle', { name: row.name, n: row.chunkCount })
    chunksPage.value = 1
    chunksVisible.value = true
    loadChunks()
  }

  const loadChunks = async () => {
    if (!currentDocId.value) return
    chunksLoading.value = true
    try {
      const res = await getDocumentChunks({
        documentId: currentDocId.value,
        page: chunksPage.value,
        pageSize: chunksPageSize.value
      })
      if (res?.code === 0) {
        chunksData.value = res.data?.list || []
        chunksTotal.value = res.data?.total || 0
      }
    } finally {
      chunksLoading.value = false
    }
  }

  const editChunkVisible = ref(false)
  const editChunkSaving = ref(false)
  const editChunkData = reactive({ id: 0, chunkIndex: 0, content: '' })

  const handleEditChunk = (row) => {
    editChunkData.id = row.ID
    editChunkData.chunkIndex = row.chunkIndex
    editChunkData.content = row.content
    editChunkVisible.value = true
  }

  const handleSaveChunk = async () => {
    if (!editChunkData.content.trim()) {
      ElMessage.warning(t('rag.docs.chunkEmptyWarn'))
      return
    }
    editChunkSaving.value = true
    try {
      const res = await updateDocumentChunk({
        id: editChunkData.id,
        content: editChunkData.content
      })
      if (res?.code === 0) {
        ElMessage.success(t('rag.docs.saveChunkOk'))
        editChunkVisible.value = false
        loadChunks()
      } else {
        ElMessage.error(res?.msg || t('rag.docs.saveChunkFail'))
      }
    } catch (e) {
      ElMessage.error(e?.message || t('rag.docs.saveChunkFail'))
    } finally {
      editChunkSaving.value = false
    }
  }

  onMounted(() => {
    loadDocuments()
    startDocumentListPolling()
  })

  onUnmounted(() => {
    stopDocumentListPolling()
  })
</script>

<style scoped>
.doc-thumbnail {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  border: 1px solid var(--el-border-color-lighter);
}

.doc-type-icon {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: bold;
  color: #fff;
  background: var(--el-color-info);
  margin: 0 auto;
}

.doc-type-icon.icon-pdf { background: #dc3545; }
.doc-type-icon.icon-doc, .doc-type-icon.icon-docx { background: #2b579a; }
.doc-type-icon.icon-xls, .doc-type-icon.icon-xlsx { background: #217346; }
.doc-type-icon.icon-ppt, .doc-type-icon.icon-pptx { background: #d24726; }
.doc-type-icon.icon-txt, .doc-type-icon.icon-md { background: #6c757d; }
.doc-type-icon.icon-csv { background: #28a745; }
.doc-type-icon.icon-html, .doc-type-icon.icon-htm { background: #e44d26; }
.doc-type-icon.icon-jpg, .doc-type-icon.icon-jpeg,
.doc-type-icon.icon-png, .doc-type-icon.icon-gif,
.doc-type-icon.icon-webp { background: #9b59b6; }

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

.chunks-container {
  min-height: 300px;
}

.chunk-content-cell {
  max-height: 120px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.5;
  padding: 4px 0;
}
</style>
