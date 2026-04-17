<template>
  <div>
    <warning-bar :title="$t('rag.kb.warningBar')" />
    <div class="lrag-table-box">
          <div class="lrag-btn-list">
            <el-button type="primary" icon="plus" @click="openDrawer">{{ $t('rag.kb.newKb') }}</el-button>
          </div>
          <el-table :data="tableData" style="width: 100%" tooltip-effect="dark" row-key="ID">
        <el-table-column align="left" :label="$t('rag.kb.colName')" prop="name" width="180" />
        <el-table-column align="left" :label="$t('rag.kb.colDescription')" prop="description" min-width="200" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('rag.kb.colChunkMethod')" width="140">
          <template #default="scope">
            <el-tag size="small" :type="chunkMethodTagType(scope.row.chunkMethod)">{{ chunkMethodLabel(scope.row.chunkMethod) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.kb.colRerank')" width="80">
          <template #default="scope">
            <el-tag v-if="scope.row.useRerank" size="small" type="success">{{ $t('rag.kb.on') }}</el-tag>
            <el-tag v-else size="small" type="info">{{ $t('rag.kb.off') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.kb.colCreatedAt')" width="180">
          <template #default="scope">
            <span>{{ formatDate(scope.row.CreatedAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.kb.colActions')" min-width="560" fixed="right">
          <template #default="scope">
            <div class="lrag-kb-actions">
              <el-button type="primary" link icon="document" @click="goDocuments(scope.row)">{{ $t('rag.kb.documents') }}</el-button>
              <el-button type="primary" link icon="connection" @click="goKnowledgeGraph(scope.row)">{{ $t('rag.kb.knowledgeGraph') }}</el-button>
              <el-button type="primary" link icon="edit" @click="updateKb(scope.row)">{{ $t('rag.kb.edit') }}</el-button>
              <el-button type="primary" link icon="share" @click="openShare(scope.row)">{{ $t('rag.kb.share') }}</el-button>
              <el-button type="danger" link icon="delete" @click="deleteKb(scope.row)">{{ $t('rag.kb.delete') }}</el-button>
            </div>
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
          @current-change="(v) => { page = v; getTableData() }"
          @size-change="(v) => { pageSize = v; getTableData() }"
        />
      </div>
    </div>

    <el-drawer v-model="drawerVisible" :before-close="closeDrawer" :title="$t('rag.kb.drawerTitle')" size="480px" destroy-on-close>
      <div style="padding: 0 16px">
        <el-form :model="form" label-width="120px">
        <el-form-item :label="$t('rag.kb.colName')" required>
          <el-input v-model="form.name" :placeholder="$t('rag.kb.namePh')" />
        </el-form-item>
        <el-form-item :label="$t('rag.kb.colDescription')">
          <el-input v-model="form.description" type="textarea" :placeholder="$t('rag.kb.descPh')" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('rag.kb.embedding')" required>
          <el-select v-model="form.embeddingId" :placeholder="$t('rag.kb.embeddingPh')" style="width: 100%" filterable @change="onEmbeddingModelChange">
            <el-option-group v-if="adminEmbeddingModels.length" :label="$t('rag.kb.adminModels')">
              <el-option
                v-for="m in adminEmbeddingModels"
                :key="'admin-' + m.id"
                :label="m.modelName + ' (' + m.name + ')'"
                :value="m.id"
              />
            </el-option-group>
            <el-option-group v-if="userEmbeddingModels.length" :label="$t('rag.kb.myModels')">
              <el-option
                v-for="m in userEmbeddingModels"
                :key="'user-' + m.id"
                :label="m.modelName + ' (' + (m.provider || m.name) + ')'"
                :value="m.id"
              />
            </el-option-group>
          </el-select>
          <div v-if="form.id" class="text-xs text-amber-600 mt-1">{{ $t('rag.kb.embedHintEdit') }}</div>
          <div v-else-if="!adminEmbeddingModels.length && !userEmbeddingModels.length" class="text-xs text-amber-600 mt-1">{{ $t('rag.kb.embedHintEmpty') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.kb.vectorStore')" required>
          <el-select v-model="form.vectorStoreId" :placeholder="$t('rag.kb.vectorStorePh')" style="width: 100%" filterable :disabled="!!form.id">
            <el-option
              v-for="v in vectorStoreOptions"
              :key="v.id"
              :label="v.label"
              :value="v.id"
            />
          </el-select>
          <div v-if="form.id" class="text-xs text-amber-600 mt-1">{{ $t('rag.kb.vectorHintEdit') }}</div>
          <div v-else-if="!vectorStoreOptions.length" class="text-xs text-amber-600 mt-1">{{ $t('rag.kb.vectorHintEmpty') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.kb.fileStorage')" required>
          <el-select v-model="form.fileStorageId" :placeholder="$t('rag.kb.fileStoragePh')" style="width: 100%" filterable :disabled="!!form.id">
            <el-option
              v-for="f in fileStorageOptionsForForm"
              :key="f.id"
              :label="f.label"
              :value="f.id"
            />
          </el-select>
          <div v-if="form.id" class="text-xs text-amber-600 mt-1">{{ $t('rag.kb.fileHintEdit') }}</div>
          <div v-else-if="!fileStorageOptions.length" class="text-xs text-amber-600 mt-1">{{ $t('rag.kb.fileHintEmpty') }}</div>
        </el-form-item>
        <el-divider content-position="left">{{ $t('rag.kb.sectionRetrieval') }}</el-divider>
        <el-form-item :label="$t('rag.kb.enableKnowledgeGraph')">
          <el-switch v-model="form.enableKnowledgeGraph" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.enableKnowledgeGraphHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.kb.retrieverType')">
          <el-select v-model="form.retrieverType" :placeholder="$t('rag.kb.retrieverPh')" style="width: 100%">
            <el-option :label="$t('rag.kb.retrieverVector')" value="vector" />
            <el-option :label="$t('rag.kb.retrieverKeyword')" value="keyword" />
            <el-option :label="$t('rag.kb.retrieverLocal')" value="local" />
            <el-option :label="$t('rag.kb.retrieverGlobal')" value="global" />
            <el-option :label="$t('rag.kb.retrieverHybrid')" value="hybrid" />
            <el-option :label="$t('rag.kb.retrieverMix')" value="mix" />
            <el-option :label="$t('rag.kb.retrieverBypass')" value="bypass" />
            <el-option :label="$t('rag.kb.retrieverPageindex')" value="pageindex" />
          </el-select>
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.retrieverLightragHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.kb.chunkMethod')">
          <el-select v-model="form.chunkMethod" :placeholder="$t('rag.kb.chunkMethodPh')" style="width: 100%">
            <el-option :label="$t('rag.kb.chunk_general')" value="general" />
            <el-option :label="$t('rag.kb.chunk_qa')" value="qa" />
            <el-option :label="$t('rag.kb.chunk_book')" value="book" />
            <el-option :label="$t('rag.kb.chunk_paper')" value="paper" />
            <el-option :label="$t('rag.kb.chunk_laws')" value="laws" />
            <el-option :label="$t('rag.kb.chunk_presentation')" value="presentation" />
            <el-option :label="$t('rag.kb.chunk_table')" value="table" />
            <el-option :label="$t('rag.kb.chunk_one')" value="one" />
          </el-select>
          <div class="text-xs text-gray-400 mt-1">{{ chunkMethodHints[form.chunkMethod] || '' }}</div>
        </el-form-item>
        <el-form-item v-if="form.chunkMethod !== 'one'" :label="$t('rag.kb.chunkSize')">
          <el-input-number v-model="form.chunkSize" :min="100" :max="4000" :step="100" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.chunkSizeHint') }}</div>
        </el-form-item>
        <el-form-item v-if="showOverlapField" :label="$t('rag.kb.chunkOverlap')">
          <el-input-number v-model="form.chunkOverlap" :min="0" :max="500" :step="10" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.chunkOverlapHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.kb.concurrentTasks')">
          <el-input-number v-model="form.concurrentSliceJobs" :min="1" :max="32" :step="1" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.concurrentTasksHint') }}</div>
        </el-form-item>
        <el-form-item v-if="showDelimiterField" :label="$t('rag.kb.delimiter')">
          <el-input v-model="form.delimiter" :placeholder="$t('rag.kb.delimiterPlaceholder')">
            <template #append>
              <el-tooltip :content="$t('rag.kb.delimiterTooltip')" placement="top">
                <el-icon><QuestionFilled /></el-icon>
              </el-tooltip>
            </template>
          </el-input>
          <div class="delimiter-preview mt-1">
            <el-tag v-for="(d, i) in delimiterPreview" :key="i" size="small" type="info" class="mr-1 mb-1">{{ d }}</el-tag>
          </div>
        </el-form-item>
        <el-divider content-position="left">{{ $t('rag.kb.sectionRerank') }}</el-divider>
        <el-form-item :label="$t('rag.kb.useRerank')">
          <el-switch v-model="form.useRerank" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.useRerankHint') }}</div>
        </el-form-item>
        <template v-if="form.useRerank">
          <el-form-item :label="$t('rag.kb.rerankModel')" required>
            <el-select v-model="form.rerankId" :placeholder="$t('rag.kb.rerankModelPh')" style="width: 100%" filterable @change="onRerankModelChange">
              <el-option-group v-if="adminRerankModels.length" :label="$t('rag.kb.adminModels')">
                <el-option
                  v-for="m in adminRerankModels"
                  :key="'admin-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
              <el-option-group v-if="userRerankModels.length" :label="$t('rag.kb.myModels')">
                <el-option
                  v-for="m in userRerankModels"
                  :key="'user-' + m.id"
                  :label="m.modelName + ' (' + m.provider + ')'"
                  :value="m.id"
                />
              </el-option-group>
            </el-select>
            <div v-if="!adminRerankModels.length && !userRerankModels.length" class="text-xs text-amber-600 mt-1">
              {{ $t('rag.kb.rerankEmpty') }}
            </div>
          </el-form-item>
          <el-form-item :label="$t('rag.kb.rerankTopK')">
            <el-input-number v-model="form.rerankTopK" :min="0" :max="200" :step="10" />
            <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.rerankTopKHint') }}</div>
          </el-form-item>
        </template>

        <template v-if="form.retrieverType === 'pageindex'">
          <el-divider content-position="left">{{ $t('rag.kb.sectionPageIndex') }}</el-divider>
          <el-form-item :label="$t('rag.kb.pageIndexLlm')">
            <el-select v-model="form.pageIndexLlmId" :placeholder="$t('rag.kb.pageIndexLlmPh')" style="width: 100%" filterable clearable @change="onPageIndexLlmChange">
              <el-option :value="0" :label="$t('rag.kb.pageIndexAuto')" />
              <el-option-group v-if="adminChatModels.length" :label="$t('rag.kb.adminModels')">
                <el-option
                  v-for="m in adminChatModels"
                  :key="'admin-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
              <el-option-group v-if="userChatModels.length" :label="$t('rag.kb.myModels')">
                <el-option
                  v-for="m in userChatModels"
                  :key="'user-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
            </el-select>
            <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.pageIndexHint') }}</div>
          </el-form-item>
        </template>

        <el-divider content-position="left">{{ $t('rag.kb.sectionParse') }}</el-divider>
        <el-form-item :label="$t('rag.kb.enableOcr')">
          <el-switch v-model="form.useOcr" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.enableOcrHint') }}</div>
        </el-form-item>
        <template v-if="form.useOcr">
          <el-form-item :label="$t('rag.kb.ocrModel')">
            <el-select v-model="form.ocrId" :placeholder="$t('rag.kb.pageIndexLlmPh')" style="width: 100%" filterable clearable @change="onOcrModelChange">
              <el-option :value="0" :label="$t('rag.kb.ocrAuto')" />
              <el-option-group v-if="adminOcrModels.length" :label="$t('rag.kb.adminModels')">
                <el-option
                  v-for="m in adminOcrModels"
                  :key="'admin-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
              <el-option-group v-if="userOcrModels.length" :label="$t('rag.kb.myModels')">
                <el-option
                  v-for="m in userOcrModels"
                  :key="'user-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
            </el-select>
          </el-form-item>
        </template>

        <el-form-item :label="$t('rag.kb.imageCaption')">
          <el-switch v-model="form.useImageDescription" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.imageCaptionHint') }}</div>
        </el-form-item>
        <template v-if="form.useImageDescription">
          <el-form-item :label="$t('rag.kb.cvModel')">
            <el-select v-model="form.imageDescriptionId" :placeholder="$t('rag.kb.pageIndexLlmPh')" style="width: 100%" filterable clearable @change="onCvModelChange">
              <el-option :value="0" :label="$t('rag.kb.cvAuto')" />
              <el-option-group v-if="adminCvModels.length" :label="$t('rag.kb.adminModels')">
                <el-option
                  v-for="m in adminCvModels"
                  :key="'admin-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
              <el-option-group v-if="userCvModels.length" :label="$t('rag.kb.myModels')">
                <el-option
                  v-for="m in userCvModels"
                  :key="'user-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
            </el-select>
          </el-form-item>
        </template>

        <el-form-item :label="$t('rag.kb.speech2text')">
          <el-switch v-model="form.useSpeech2Text" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('rag.kb.speech2textHint') }}</div>
        </el-form-item>
        <template v-if="form.useSpeech2Text">
          <el-form-item :label="$t('rag.kb.sttModel')">
            <el-select v-model="form.speech2TextId" :placeholder="$t('rag.kb.pageIndexLlmPh')" style="width: 100%" filterable clearable @change="onSpeech2TextModelChange">
              <el-option :value="0" :label="$t('rag.kb.sttAuto')" />
              <el-option-group v-if="adminSpeech2TextModels.length" :label="$t('rag.kb.adminModels')">
                <el-option
                  v-for="m in adminSpeech2TextModels"
                  :key="'admin-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
              <el-option-group v-if="userSpeech2TextModels.length" :label="$t('rag.kb.myModels')">
                <el-option
                  v-for="m in userSpeech2TextModels"
                  :key="'user-' + m.id"
                  :label="m.modelName + ' (' + m.name + ')'"
                  :value="m.id"
                />
              </el-option-group>
            </el-select>
          </el-form-item>
        </template>
      </el-form>
      </div>
      <template #footer>
        <el-button @click="closeDrawer">{{ $t('rag.kb.cancel') }}</el-button>
        <el-button type="primary" @click="enterDrawer">{{ $t('rag.kb.confirm') }}</el-button>
      </template>
    </el-drawer>

    <el-dialog v-model="shareVisible" :title="$t('rag.kb.shareTitle')" width="400px">
      <el-form :model="shareForm" label-width="100px">
        <el-form-item :label="$t('rag.kb.targetType')">
          <el-select v-model="shareForm.targetType" :placeholder="$t('rag.kb.selectPh')" @change="onShareTargetTypeChange">
            <el-option :label="$t('rag.kb.targetUser')" value="user" />
            <el-option :label="$t('rag.kb.targetRole')" value="role" />
          </el-select>
        </el-form-item>
        <el-form-item :label="shareForm.targetType === 'user' ? $t('rag.kb.targetUserLabel') : $t('rag.kb.targetRoleLabel')" required>
          <el-select
            v-model="shareForm.targetId"
            :placeholder="shareForm.targetType === 'user' ? $t('rag.kb.selectUserPh') : $t('rag.kb.selectRolePh')"
            style="width: 100%"
            filterable
            clearable
          >
            <el-option
              v-for="item in shareTargetOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.kb.permission')">
          <el-select v-model="shareForm.permission">
            <el-option :label="$t('rag.kb.permRead')" value="read" />
            <el-option :label="$t('rag.kb.permWrite')" value="write" />
            <el-option :label="$t('rag.kb.permAdmin')" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shareVisible = false">{{ $t('rag.kb.cancel') }}</el-button>
        <el-button type="primary" @click="doShare">{{ $t('rag.kb.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    createKnowledgeBase,
    updateKnowledgeBase,
    deleteKnowledgeBase,
    getKnowledgeBaseList,
    shareKnowledgeBase,
    listVectorStoreConfigs,
    listFileStorageConfigs,
    listLLMProviders
  } from '@/api/rag'
  import { getUserList } from '@/api/user'
  import { getAuthorityList } from '@/api/authority'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, computed, onMounted, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import { formatDate } from '@/utils/format'
  import { authorityDisplayName } from '@/utils/authorityI18n'
  import { useRouter } from 'vue-router'

  defineOptions({ name: 'KnowledgeBase' })

  const { t } = useI18n()

  const chunkMethodHints = computed(() => ({
    general: t('rag.kb.chunkDesc_general'),
    qa: t('rag.kb.chunkDesc_qa'),
    book: t('rag.kb.chunkDesc_book'),
    paper: t('rag.kb.chunkDesc_paper'),
    laws: t('rag.kb.chunkDesc_laws'),
    presentation: t('rag.kb.chunkDesc_presentation'),
    table: t('rag.kb.chunkDesc_table'),
    one: t('rag.kb.chunkDesc_one')
  }))

  const showOverlapField = computed(() => {
    return ['general', 'book', 'paper', 'laws'].includes(form.value.chunkMethod)
  })

  const showDelimiterField = computed(() => {
    return ['general'].includes(form.value.chunkMethod)
  })

  const delimiterPreview = computed(() => {
    const d = form.value.delimiter || ''
    const parsed = d.replace(/\\n/g, '\n').replace(/\\t/g, '\t').replace(/\\r/g, '\r')
    const chars = [...new Set([...parsed])]
    return chars.map(c => {
      if (c === '\n') return t('rag.kb.newlineDisplay')
      if (c === '\t') return t('rag.kb.tabDisplay')
      if (c === '\r') return '\\r'
      if (c === ' ') return t('rag.kb.spaceDisplay')
      return c
    }).filter(Boolean)
  })

  const chunkMethodKnown = ['general', 'qa', 'book', 'paper', 'laws', 'presentation', 'table', 'one']
  const chunkMethodLabel = (method) => {
    if (method && chunkMethodKnown.includes(method)) {
      return t(`rag.kb.chunkShort_${method}`)
    }
    return method || t('rag.kb.chunkShort_general')
  }

  const chunkMethodTagType = (method) => {
    const map = {
      general: '', qa: 'success', book: 'warning', paper: 'info',
      laws: 'danger', presentation: 'warning', table: 'info', one: 'success'
    }
    return map[method] || ''
  }

  // ---- 模型列表 ----
  const adminEmbeddingModels = ref([])
  const userEmbeddingModels = ref([])
  const adminRerankModels = ref([])
  const userRerankModels = ref([])
  const adminChatModels = ref([])
  const userChatModels = ref([])
  const adminOcrModels = ref([])
  const userOcrModels = ref([])
  const adminCvModels = ref([])
  const userCvModels = ref([])
  const adminSpeech2TextModels = ref([])
  const userSpeech2TextModels = ref([])

  const loadModelsByScenario = async (scenarioType) => {
    try {
      const res = await listLLMProviders({ scenarioType })
      const list = res.code === 0 ? (Array.isArray(res.data) ? res.data : res.data?.list || []) : []
      if (list.length) {
        return {
          admin: list.filter(m => m.source === 'admin'),
          user: list.filter(m => m.source === 'user')
        }
      }
    } catch (e) {
      console.warn(`Failed to load ${scenarioType} model list`, e)
    }
    return { admin: [], user: [] }
  }

  const loadAllModels = async () => {
    const [embRes, rerankRes, chatRes, ocrRes, cvRes, s2tRes] = await Promise.all([
      loadModelsByScenario('embedding'),
      loadModelsByScenario('rerank'),
      loadModelsByScenario('chat'),
      loadModelsByScenario('ocr'),
      loadModelsByScenario('cv'),
      loadModelsByScenario('speech2text')
    ])
    adminEmbeddingModels.value = embRes.admin
    userEmbeddingModels.value = embRes.user
    adminRerankModels.value = rerankRes.admin
    userRerankModels.value = rerankRes.user
    adminChatModels.value = chatRes.admin
    userChatModels.value = chatRes.user
    adminOcrModels.value = ocrRes.admin
    userOcrModels.value = ocrRes.user
    adminCvModels.value = cvRes.admin
    userCvModels.value = cvRes.user
    adminSpeech2TextModels.value = s2tRes.admin
    userSpeech2TextModels.value = s2tRes.user
  }

  const resolveModelSource = (val, adminList, userList) => {
    if (adminList.find(m => m.id === val)) return 'admin'
    if (userList.find(m => m.id === val)) return 'user'
    return 'admin'
  }

  const onEmbeddingModelChange = (val) => {
    form.value.embeddingSource = resolveModelSource(val, adminEmbeddingModels.value, userEmbeddingModels.value)
  }

  const onRerankModelChange = (val) => {
    form.value.rerankSource = resolveModelSource(val, adminRerankModels.value, userRerankModels.value)
  }

  const onPageIndexLlmChange = (val) => {
    form.value.pageIndexLlmSource = resolveModelSource(val, adminChatModels.value, userChatModels.value)
  }

  const onOcrModelChange = (val) => {
    form.value.ocrSource = resolveModelSource(val, adminOcrModels.value, userOcrModels.value)
  }

  const onCvModelChange = (val) => {
    form.value.imageDescriptionSource = resolveModelSource(val, adminCvModels.value, userCvModels.value)
  }

  const onSpeech2TextModelChange = (val) => {
    form.value.speech2TextSource = resolveModelSource(val, adminSpeech2TextModels.value, userSpeech2TextModels.value)
  }

  const router = useRouter()
  const vectorStoreOptions = ref([])
  const fileStorageOptions = ref([])
  const shareTargetOptions = ref([])
  // 文件存储选项：编辑时若 fileStorageId 为 0（旧数据）则补充“默认”占位
  const fileStorageOptionsForForm = computed(() => {
    const opts = fileStorageOptions.value
    if (form.value.id && form.value.fileStorageId === 0 && opts.length) {
      return [{ id: 0, label: t('rag.kb.fileStorageDefault') }, ...opts]
    }
    return opts
  })
  const defaultFormValues = () => ({
    name: '',
    description: '',
    embeddingId: undefined,
    embeddingSource: 'admin',
    vectorStoreId: undefined,
    fileStorageId: undefined,
    retrieverType: 'vector',
    chunkMethod: 'general',
    chunkSize: 500,
    chunkOverlap: 50,
    concurrentSliceJobs: 1,
    delimiter: '\\n!?。；！？',
    autoKeywords: 0,
    autoQuestions: 0,
    useRerank: false,
    rerankId: undefined,
    rerankSource: 'admin',
    rerankTopK: 0,
    pageIndexLlmId: 0,
    pageIndexLlmSource: 'admin',
    useOcr: false,
    ocrId: 0,
    ocrSource: 'admin',
    useImageDescription: false,
    imageDescriptionId: 0,
    imageDescriptionSource: 'admin',
    useSpeech2Text: false,
    speech2TextId: 0,
    speech2TextSource: 'admin',
    enableKnowledgeGraph: true
  })
  const form = ref(defaultFormValues())
  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const drawerVisible = ref(false)
  const shareVisible = ref(false)
  const type = ref('create')
  const currentKbId = ref(0)
  const shareForm = ref({
    targetType: 'user',
    targetId: undefined,
    permission: 'read'
  })

  const getTableData = async () => {
    const res = await getKnowledgeBaseList({
      page: page.value,
      pageSize: pageSize.value
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const loadVectorStoreOptions = async () => {
    const res = await listVectorStoreConfigs()
    if (res.code === 0) vectorStoreOptions.value = res.data || []
  }

  const loadFileStorageOptions = async () => {
    const res = await listFileStorageConfigs()
    if (res.code === 0) fileStorageOptions.value = res.data || []
  }

  const loadShareTargetOptions = async () => {
    shareTargetOptions.value = []
    shareForm.value.targetId = undefined
    if (shareForm.value.targetType === 'user') {
      const res = await getUserList({ page: 1, pageSize: 500 })
      if (res.code === 0 && res.data?.list) {
        shareTargetOptions.value = res.data.list.map(u => ({
          value: u.ID,
          label: `${u.nickName || u.userName || ''} (${u.userName || u.ID})`
        }))
      }
    } else if (shareForm.value.targetType === 'role') {
      const res = await getAuthorityList()
      if (res.code === 0 && res.data) {
        const flatten = (list) => {
          let out = []
          for (const a of list) {
            out.push({
              value: a.authorityId,
              label: `${authorityDisplayName(a, t)} (${a.authorityId})`
            })
            if (a.children?.length) out = out.concat(flatten(a.children))
          }
          return out
        }
        shareTargetOptions.value = flatten(Array.isArray(res.data) ? res.data : [res.data])
      }
    }
  }

  const onShareTargetTypeChange = () => loadShareTargetOptions()

  onMounted(() => {
    getTableData()
    loadVectorStoreOptions()
    loadFileStorageOptions()
  })

  watch(drawerVisible, (v) => {
    if (v) {
      loadVectorStoreOptions()
      loadFileStorageOptions()
    }
  })

  const openDrawer = async () => {
    type.value = 'create'
    await Promise.all([loadVectorStoreOptions(), loadFileStorageOptions(), loadAllModels()])
    const firstEmb = adminEmbeddingModels.value[0] || userEmbeddingModels.value[0]
    form.value = {
      ...defaultFormValues(),
      embeddingId: firstEmb?.id,
      embeddingSource: firstEmb ? (adminEmbeddingModels.value[0] ? 'admin' : 'user') : 'admin',
      vectorStoreId: vectorStoreOptions.value[0]?.id,
      fileStorageId: fileStorageOptions.value[0]?.id,
    }
    drawerVisible.value = true
  }

  const updateKb = async (row) => {
    type.value = 'update'
    await Promise.all([loadVectorStoreOptions(), loadFileStorageOptions(), loadAllModels()])
    form.value = {
      id: row.ID,
      name: row.name,
      description: row.description,
      embeddingId: row.embeddingId,
      embeddingSource: row.embeddingSource || 'legacy',
      vectorStoreId: row.vectorStoreId,
      fileStorageId: row.fileStorageId,
      retrieverType: row.retrieverType || 'vector',
      chunkMethod: row.chunkMethod || 'general',
      chunkSize: row.chunkSize || 500,
      chunkOverlap: row.chunkOverlap || 50,
      concurrentSliceJobs: row.concurrentSliceJobs >= 1 ? row.concurrentSliceJobs : 1,
      delimiter: row.delimiter || '\\n!?。；！？',
      autoKeywords: row.autoKeywords || 0,
      autoQuestions: row.autoQuestions || 0,
      useRerank: row.useRerank || false,
      rerankId: row.rerankId || undefined,
      rerankSource: row.rerankSource || 'admin',
      rerankTopK: row.rerankTopK || 0,
      pageIndexLlmId: row.pageIndexLlmId || 0,
      pageIndexLlmSource: row.pageIndexLlmSource || 'admin',
      useOcr: row.useOcr !== false,
      ocrId: row.ocrId || 0,
      ocrSource: row.ocrSource || 'admin',
      useImageDescription: row.useImageDescription || false,
      imageDescriptionId: row.imageDescriptionId || 0,
      imageDescriptionSource: row.imageDescriptionSource || 'admin',
      useSpeech2Text: row.useSpeech2Text || false,
      speech2TextId: row.speech2TextId || 0,
      speech2TextSource: row.speech2TextSource || 'admin',
      enableKnowledgeGraph: row.enableKnowledgeGraph === true
    }
    drawerVisible.value = true
  }

  const closeDrawer = () => {
    drawerVisible.value = false
  }

  const enterDrawer = async () => {
    if (!form.value.name) {
      ElMessage.warning(t('rag.kb.validateName'))
      return
    }
    if (!form.value.embeddingId) {
      ElMessage.warning(t('rag.kb.validateEmbedding'))
      return
    }
    if (!form.value.vectorStoreId) {
      ElMessage.warning(t('rag.kb.validateVector'))
      return
    }
    if (!form.value.id && !form.value.fileStorageId) {
      ElMessage.warning(t('rag.kb.validateFile'))
      return
    }
    let res
    if (type.value === 'create') {
      res = await createKnowledgeBase(form.value)
    } else {
      res = await updateKnowledgeBase(form.value)
    }
    if (res.code === 0) {
      ElMessage.success(t('rag.kb.opSuccess'))
      closeDrawer()
      getTableData()
    }
  }

  const deleteKb = (row) => {
    ElMessageBox.confirm(t('rag.kb.deleteConfirm'), t('rag.kb.deleteTitle'), {
      confirmButtonText: t('rag.kb.confirm'),
      cancelButtonText: t('rag.kb.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteKnowledgeBase({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('rag.kb.deleteOk'))
        getTableData()
      }
    })
  }

  const goDocuments = (row) => {
    router.push({ name: 'ragDocuments', query: { id: row.ID, name: row.name } })
  }

  const goKnowledgeGraph = (row) => {
    router.push({ name: 'ragKnowledgeGraph', query: { id: row.ID, name: row.name } })
  }

  const openShare = async (row) => {
    currentKbId.value = row.ID
    shareForm.value = { targetType: 'user', targetId: undefined, permission: 'read' }
    shareVisible.value = true
    await loadShareTargetOptions()
  }

  const doShare = async () => {
    if (!shareForm.value.targetId) {
      ElMessage.warning(t('rag.kb.shareNeedTarget'))
      return
    }
    const res = await shareKnowledgeBase({
      id: currentKbId.value,
      targetType: shareForm.value.targetType,
      targetId: shareForm.value.targetId,
      permission: shareForm.value.permission
    })
    if (res.code === 0) {
      ElMessage.success(t('rag.kb.shareOk'))
      shareVisible.value = false
    }
  }
</script>

<style scoped>
  .lrag-kb-actions {
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    column-gap: 2px;
  }
</style>
