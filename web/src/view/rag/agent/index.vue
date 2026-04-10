<template>
  <div>
    <warning-bar :title="$t('rag.agent.warningBar')" />
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openCreate">{{ $t('rag.agent.btnNew') }}</el-button>
        <el-button icon="document-copy" @click="openTemplateDialog">{{ $t('rag.agent.btnFromTemplate') }}</el-button>
      </div>
      <el-table :data="tableData" style="width: 100%" tooltip-effect="dark" row-key="ID">
        <el-table-column align="left" :label="$t('rag.agent.colName')" prop="name" width="200" />
        <el-table-column align="left" :label="$t('rag.agent.colDesc')" prop="desc" min-width="200" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('rag.agent.colCreatedAt')" width="180">
          <template #default="scope">
            <span>{{ formatDate(scope.row.CreatedAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.agent.colActions')" width="280" fixed="right">
          <template #default="scope">
            <el-button type="primary" link icon="edit" @click="goEdit(scope.row)">{{ $t('rag.agent.actionEdit') }}</el-button>
            <el-button type="primary" link icon="video-play" @click="goRun(scope.row)">{{ $t('rag.agent.actionRun') }}</el-button>
            <el-button type="danger" link icon="delete" @click="deleteAgent(scope.row)">{{ $t('rag.agent.actionDelete') }}</el-button>
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

    <!-- 新建/编辑弹窗 -->
    <el-dialog
      v-model="drawerVisible"
      :title="editId ? $t('rag.agent.dialogEditTitle') : $t('rag.agent.dialogNewTitle')"
      width="500px"
      destroy-on-close
    >
      <el-form :model="form" label-width="80px">
        <el-form-item :label="$t('rag.agent.labelName')" required>
          <el-input v-model="form.name" :placeholder="$t('rag.agent.placeholderName')" />
        </el-form-item>
        <el-form-item :label="$t('rag.agent.labelDesc')">
          <el-input v-model="form.desc" type="textarea" :placeholder="$t('rag.agent.placeholderDesc')" :rows="3" />
        </el-form-item>
        <el-form-item v-if="!editId" :label="$t('rag.agent.labelDsl')" required>
          <el-button size="small" @click="loadDefaultDSL">{{ $t('rag.agent.btnUseDefaultTemplate') }}</el-button>
          <div class="text-xs text-gray-500 mt-1">{{ $t('rag.agent.hintDslOrTemplate') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="drawerVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="submitForm">{{ $t('settings.general.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 从模板创建弹窗（卡片 + 分类 + 搜索，对齐多实例管理体验） -->
    <el-dialog v-model="templateVisible" :title="$t('rag.agent.dialogTemplateTitle')" width="720px" destroy-on-close class="agent-template-dialog">
      <div class="template-toolbar">
        <el-input
          v-model="templateSearch"
          clearable
          :placeholder="$t('rag.agent.templateSearchPh')"
          style="width: 100%; max-width: 320px"
        />
        <el-radio-group v-model="templateCategory" size="small">
          <el-radio-button label="all">{{ $t('rag.agent.categoryAll') }}</el-radio-button>
          <el-radio-button v-for="c in templateCategoryList" :key="c" :label="c">{{ c }}</el-radio-button>
        </el-radio-group>
      </div>
      <div class="template-grid">
        <div
          v-for="t in filteredTemplates"
          :key="templateKey(t)"
          class="template-card"
          :class="{ 'template-card--active': templateForm.templateName === templateKey(t) }"
          @click="selectTemplateCard(t)"
        >
          <div class="template-card-title">{{ t.title || t.templateName || $t('rag.agent.unnamed') }}</div>
          <div class="template-card-desc">{{ t.description || $t('rag.agent.noDesc') }}</div>
          <div class="template-card-tags">
            <el-tag v-if="t.category" size="small" type="info">{{ t.category }}</el-tag>
            <el-tag v-for="tag in (t.tags || []).slice(0, 4)" :key="tag" size="small" class="ml-1">{{ tag }}</el-tag>
          </div>
        </div>
      </div>
      <div v-if="!filteredTemplates.length" class="text-center text-gray-500 py-8">{{ $t('rag.agent.templateNoMatch') }}</div>
      <el-divider />
      <el-form :model="templateForm" label-width="100px">
        <el-form-item :label="$t('rag.agent.labelSelectedTemplate')">
          <el-input :model-value="selectedTemplateTitle" readonly />
        </el-form-item>
        <el-form-item :label="$t('rag.agent.labelAgentName')" required>
          <el-input v-model="templateForm.name" :placeholder="$t('rag.agent.placeholderAgentName')" />
        </el-form-item>
        <el-form-item :label="$t('rag.agent.labelDesc')">
          <el-input v-model="templateForm.desc" type="textarea" :placeholder="$t('rag.agent.templateDescPh')" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="templateVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" :loading="templateSubmitting" @click="submitFromTemplate">{{ $t('rag.agent.btnCreateAndEdit') }}</el-button>
      </template>
    </el-dialog>

    <!-- 运行弹窗 -->
    <el-dialog v-model="runVisible" :title="$t('rag.agent.dialogRunTitle')" width="600px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item :label="$t('rag.agent.labelQuestion')">
          <el-input v-model="runQuery" type="textarea" :rows="3" :placeholder="$t('rag.agent.placeholderQuestion')" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="runLoading" @click="doRun">{{ $t('rag.agent.btnRun') }}</el-button>
          <el-button v-if="runConversationId" size="small" @click="runConversationId = null">{{ $t('rag.agent.btnNewConv') }}</el-button>
          <span v-if="runConversationId" class="text-xs text-gray-500 ml-2">{{ $t('rag.agent.contextHint') }}</span>
        </el-form-item>
        <el-form-item v-if="runResult !== null" :label="$t('rag.agent.labelOutput')">
          <div class="p-3 rounded bg-slate-100 dark:bg-slate-800 max-h-60 overflow-y-auto text-sm">
            <div v-if="!String(runResult).trim()" class="text-slate-400 dark:text-slate-500">{{ $t('rag.agent.emptyOutput') }}</div>
            <div v-else class="agent-run-markdown" v-html="runResultHtml" />
          </div>
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script setup>
  import { agentList, agentCreate, agentUpdate, agentDelete, agentListTemplates, agentCreateFromTemplate, agentRunStream } from '@/api/rag'
  import { ref, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { renderAgentMarkdown } from '@/utils/agentMarkdownRender'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate } from '@/utils/format'
  import WarningBar from '@/components/warningBar/warningBar.vue'

  defineOptions({ name: 'RagAgent' })

  const { t, locale } = useI18n()
  const router = useRouter()
  const tableData = ref([])
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const drawerVisible = ref(false)
  const templateVisible = ref(false)
  const runVisible = ref(false)
  const editId = ref(null)
  const form = ref({ name: '', desc: '', dsl: null })
  const templates = ref([])
  const templateForm = ref({ templateName: '', name: '', desc: '' })
  const templateSubmitting = ref(false)
  const templateSearch = ref('')
  const templateCategory = ref('all')

  const templateKey = (t) => t.templateName || (t.title && String(t.title).replace(/\s+/g, '_')) || 'template'

  const templateCategoryList = computed(() => {
    const s = new Set()
    templates.value.forEach((t) => {
      if (t.category) s.add(t.category)
    })
    return Array.from(s).sort((a, b) =>
      a.localeCompare(b, locale.value?.replace('_', '-') || undefined)
    )
  })

  const filteredTemplates = computed(() => {
    let list = [...templates.value]
    list.sort((a, b) => {
      const ca = a.category || '\uffff'
      const cb = b.category || '\uffff'
      if (ca !== cb) return ca.localeCompare(cb, locale.value?.replace('_', '-') || undefined)
      return (a.title || '').localeCompare(
        b.title || '',
        locale.value?.replace('_', '-') || undefined
      )
    })
    const q = templateSearch.value.trim().toLowerCase()
    const cat = templateCategory.value
    return list.filter((t) => {
      if (cat !== 'all' && t.category !== cat) return false
      if (!q) return true
      const tags = Array.isArray(t.tags) ? t.tags.join(' ') : ''
      const hay = `${t.title || ''} ${t.description || ''} ${tags}`.toLowerCase()
      return hay.includes(q)
    })
  })

  const selectedTemplateTitle = computed(() => {
    const t = templates.value.find((x) => templateKey(x) === templateForm.value.templateName)
    if (!t) return templateForm.value.templateName || '—'
    return (t.title || t.templateName || '') + (t.description ? ` — ${t.description}` : '')
  })

  const selectTemplateCard = (t) => {
    templateForm.value.templateName = templateKey(t)
  }
  const runQuery = ref('')
  const runLoading = ref(false)
  const runResult = ref(null)
  const runResultHtml = computed(() => renderAgentMarkdown(runResult.value ?? ''))
  const runAgentId = ref(0)
  const runConversationId = ref(null)

  const defaultDSL = () => ({
    components: {
      begin: {
        obj: { component_name: 'Begin', params: { prologue: t('rag.agent.dslPrologue') } },
        downstream: ['retrieval_0'],
        upstream: []
      },
      retrieval_0: {
        obj: {
          component_name: 'Retrieval',
          params: {
            query: '{sys.query}',
            top_n: 20,
            empty_response: t('rag.agent.dslEmptyResponse'),
            kb_ids: []
          }
        },
        downstream: ['generate_0'],
        upstream: ['begin']
      },
      generate_0: {
        obj: {
          component_name: 'LLM',
          params: {
            llm_id: 'ollama@llama3.2',
            sys_prompt: t('rag.agent.dslSysPrompt'),
            prompts: [{ role: 'user', content: t('rag.agent.dslUserQuestion') }]
          }
        },
        downstream: ['message_0'],
        upstream: ['retrieval_0']
      },
      message_0: {
        obj: { component_name: 'Message', params: { content: ['{generate_0@content}'] } },
        downstream: [],
        upstream: ['generate_0']
      }
    },
    path: ['begin', 'retrieval_0', 'generate_0', 'message_0'],
    globals: { 'sys.query': '', 'sys.user_id': '', 'sys.conversation_turns': 0, 'sys.files': [] },
    history: [],
    retrieval: []
  })

  const getTableData = async () => {
    const res = await agentList({ page: page.value, pageSize: pageSize.value })
    if (res.code === 0 && res.data) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const openCreate = () => {
    editId.value = null
    form.value = { name: t('rag.agent.defaultNewName'), desc: '', dsl: defaultDSL() }
    drawerVisible.value = true
  }

  const loadDefaultDSL = () => {
    form.value.dsl = defaultDSL()
  }

  const submitForm = async () => {
    if (!form.value.name?.trim()) {
      ElMessage.warning(t('rag.agent.needName'))
      return
    }
    if (editId.value) {
      const res = await agentUpdate({ id: editId.value, name: form.value.name, desc: form.value.desc })
      if (res.code === 0) {
        ElMessage.success(t('rag.agent.updateOk'))
        drawerVisible.value = false
        getTableData()
      }
    } else {
      const res = await agentCreate({ name: form.value.name, desc: form.value.desc, dsl: form.value.dsl || defaultDSL() })
      if (res.code === 0) {
        ElMessage.success(t('rag.agent.createOk'))
        drawerVisible.value = false
        router.push({ name: 'ragAgentEditor', query: { id: res.data.ID } })
      }
    }
  }

  const goEdit = (row) => {
    router.push({ name: 'ragAgentEditor', query: { id: row.ID } })
  }

  const openTemplateDialog = async () => {
    const res = await agentListTemplates()
    if (res.code === 0 && res.data) {
      templates.value = res.data
      templateSearch.value = ''
      templateCategory.value = 'all'
      const first = templates.value[0]
      const key = first ? templateKey(first) : ''
      templateForm.value = { templateName: key, name: '', desc: '' }
      templateVisible.value = true
    }
  }

  const submitFromTemplate = async () => {
    if (!templateForm.value.templateName || !templateForm.value.name?.trim()) {
      ElMessage.warning(t('rag.agent.pickTemplateAndName'))
      return
    }
    templateSubmitting.value = true
    try {
      const res = await agentCreateFromTemplate({
        templateName: templateForm.value.templateName,
        name: templateForm.value.name.trim(),
        desc: templateForm.value.desc || ''
      })
      if (res.code === 0) {
        ElMessage.success(t('rag.agent.createOk'))
        templateVisible.value = false
        router.push({ name: 'ragAgentEditor', query: { id: res.data.ID } })
      }
    } finally {
      templateSubmitting.value = false
    }
  }

  const goRun = (row) => {
    runAgentId.value = row.ID
    runQuery.value = ''
    runResult.value = null
    runConversationId.value = null
    runVisible.value = true
  }

  const doRun = async () => {
    if (!runQuery.value?.trim()) {
      ElMessage.warning(t('rag.agent.needQuestion'))
      return
    }
    runLoading.value = true
    runResult.value = ''
    try {
      const payload = { agentId: runAgentId.value, query: runQuery.value.trim() }
      if (runConversationId.value) payload.conversationId = runConversationId.value
      await agentRunStream(payload, {
        onChunk: (chunk) => { runResult.value += chunk },
        onDone: (convId, meta) => {
          if (convId) runConversationId.value = convId
          if (meta?.workflowPausedAtEntry) {
            ElMessage.info(t('rag.agent.workflowPausedHint'))
          }
        },
        onError: (err) => {
          runResult.value = t('rag.agent.runFailedPrefix') + err
        }
      })
    } catch (e) {
      runResult.value =
        t('rag.agent.runFailedPrefix') + (e.message || t('rag.agent.networkError'))
    } finally {
      runLoading.value = false
    }
  }

  const deleteAgent = (row) => {
    ElMessageBox.confirm(t('rag.agent.deleteConfirm', { name: row.name }), t('rag.agent.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await agentDelete({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('rag.agent.deleteOk'))
        getTableData()
      }
    }).catch(() => {})
  }

  onMounted(() => {
    getTableData()
  })
</script>

<style scoped>
@import 'highlight.js/styles/github.css';

.template-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
  max-height: 360px;
  overflow-y: auto;
  padding: 2px;
}
.template-card {
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  background: var(--el-bg-color);
  transition: border-color 0.15s, box-shadow 0.15s;
}
.template-card:hover {
  border-color: var(--el-color-primary-light-5);
}
.template-card--active {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px var(--el-color-primary-light-7);
}
.template-card-title {
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 6px;
  line-height: 1.35;
}
.template-card-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.45;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.template-card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}
.template-card-tags .ml-1 {
  margin-left: 0;
}

.agent-run-markdown :deep(p) { margin: 0.5em 0; }
.agent-run-markdown :deep(p:first-child) { margin-top: 0; }
.agent-run-markdown :deep(p:last-child) { margin-bottom: 0; }
.agent-run-markdown :deep(ul), .agent-run-markdown :deep(ol) { margin: 0.5em 0; padding-left: 1.5em; }
.agent-run-markdown :deep(li) { margin: 0.25em 0; }
.agent-run-markdown :deep(h1), .agent-run-markdown :deep(h2), .agent-run-markdown :deep(h3) { margin: 0.75em 0 0.5em; font-weight: 600; }
.agent-run-markdown :deep(blockquote) { margin: 0.5em 0; padding-left: 1em; border-left: 4px solid #94a3b8; opacity: 0.9; }
.agent-run-markdown :deep(code) { padding: 0.2em 0.4em; border-radius: 4px; font-size: 0.9em; background: rgba(0,0,0,0.08); }
.agent-run-markdown :deep(pre) { margin: 0.5em 0; padding: 0.75em; border-radius: 6px; overflow-x: auto; background: rgba(0,0,0,0.06); }
.agent-run-markdown :deep(pre code) { padding: 0; background: none; }
.agent-run-markdown :deep(a) { color: #3b82f6; text-decoration: underline; }
.agent-run-markdown :deep(table) { border-collapse: collapse; margin: 0.5em 0; }
.agent-run-markdown :deep(th), .agent-run-markdown :deep(td) { border: 1px solid rgba(0,0,0,0.1); padding: 0.25em 0.5em; }
.dark .agent-run-markdown :deep(code) { background: rgba(255,255,255,0.1); }
.dark .agent-run-markdown :deep(pre) { background: rgba(255,255,255,0.06); }
.dark .agent-run-markdown :deep(blockquote) { border-left-color: #64748b; }
</style>
