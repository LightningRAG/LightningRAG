<template>
  <div>
    <warning-bar :title="$t('rag.userModel.warningBar')" />
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openAdd">{{ $t('rag.userModel.addMyModel') }}</el-button>
      </div>
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="$t('rag.userModel.tabProviders')" name="providers">
          <el-table :data="providers" style="width: 100%">
            <el-table-column align="left" :label="$t('rag.userModel.colProvider')" prop="name" width="120" />
            <el-table-column align="left" :label="$t('rag.userModel.colModelName')" prop="modelName" />
            <el-table-column align="left" :label="$t('rag.userModel.colScenarios')" min-width="200">
              <template #default="scope">
                <el-tag v-for="mt in (scope.row.modelTypes || ['chat'])" :key="mt" size="small" class="mr-1">
                  {{ modelTypeOptions.find(o => o.value === mt)?.label || mt }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.userModel.colSource')" prop="source" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.source === 'admin' ? 'success' : 'info'">
                  {{ scope.row.source === 'admin' ? $t('rag.conv.sourceAdmin') : $t('rag.conv.sourceCustom') }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="$t('rag.userModel.tabDefaults')" name="defaults">
          <div class="text-gray-500 text-sm mb-4">{{ $t('rag.userModel.defaultsHint') }}</div>
          <el-form label-width="120px" class="max-w-md">
            <el-form-item v-for="opt in modelTypeOptions" :key="opt.value" :label="opt.label + $t('rag.userModel.defaultSuffix')">
              <el-select
                v-model="userDefaults[opt.value]"
                :placeholder="$t('rag.userModel.pickDefaultModel')"
                clearable
                style="width: 100%"
                @change="(v) => saveUserDefault(opt.value, v)"
              >
                <el-option
                  v-for="m in providersForType(opt.value)"
                  :key="(m.source || 'user') + '-' + m.id"
                  :label="formatProviderRowLabel(m)"
                  :value="(m.source || 'user') + ':' + m.id"
                />
              </el-select>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane :label="$t('rag.userModel.tabMine')" name="user">
          <el-table :data="userModels" style="width: 100%">
            <el-table-column align="left" :label="$t('rag.userModel.colProvider')" prop="provider" width="120" />
            <el-table-column align="left" :label="$t('rag.userModel.colModelName')" prop="modelName" />
            <el-table-column align="left" :label="$t('rag.userModel.colScenarios')" min-width="200">
              <template #default="scope">
                <el-tag v-for="mt in (scope.row.modelTypes || ['chat'])" :key="mt" size="small" class="mr-1">
                  {{ modelTypeOptions.find(o => o.value === mt)?.label || mt }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.userModel.colBaseUrl')" prop="baseUrl" show-overflow-tooltip />
            <el-table-column align="left" :label="$t('rag.docs.colActions')" width="150">
              <template #default="scope">
                <el-button type="primary" link icon="edit" @click="openEdit(scope.row)">{{ $t('rag.docs.edit') }}</el-button>
                <el-button type="danger" link icon="delete" @click="deleteModel(scope.row)">{{ $t('rag.docs.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="$t('rag.userModel.tabWebSearch')" name="webSearch">
          <div class="text-gray-500 text-sm mb-4">{{ $t('rag.userModel.webSearchHint') }}</div>
          <el-form label-width="140px" class="max-w-md">
            <el-form-item :label="$t('rag.userModel.useSystemDefault')">
              <el-switch v-model="webSearchForm.useSystemDefault" />
              <div class="text-gray-400 text-xs mt-1">{{ $t('rag.userModel.useSystemDefaultHint') }}</div>
            </el-form-item>
            <template v-if="!webSearchForm.useSystemDefault">
              <el-form-item :label="$t('rag.userModel.defaultSearchEngine')">
                <el-select v-model="webSearchForm.provider" :placeholder="$t('rag.userModel.pickSearchEngine')" style="width: 100%">
                  <el-option
                    v-for="p in webSearchProviders"
                    :key="p.id"
                    :label="p.displayName"
                    :value="p.id"
                  />
                </el-select>
              </el-form-item>
              <template v-for="f in currentWebSearchSchema" :key="f.key">
                <el-form-item :label="f.label" :required="f.required">
                  <el-input
                    v-model="webSearchForm.config[f.key]"
                    :type="f.secret ? 'password' : 'text'"
                    :placeholder="f.placeholder || ''"
                    show-password
                    clearable
                    style="width: 100%"
                  />
                </el-form-item>
              </template>
            </template>
            <el-form-item>
              <el-button type="primary" @click="saveWebSearchConfig">{{ $t('rag.docs.save') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="addVisible" :title="$t('rag.userModel.dialogAddTitle')" width="450px">
      <el-form :model="addForm" label-width="100px">
        <el-form-item :label="$t('rag.userModel.colProvider')" required>
          <el-select
            v-model="addForm.provider"
            :placeholder="$t('rag.userModel.phSelectSearch')"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="p in providerOptions"
              :key="p.value"
              :label="p.label"
              :value="p.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colModelName')" required>
          <el-input v-model="addForm.modelName" :placeholder="$t('rag.userModel.phModelExample')" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colScenarios')" required>
          <el-checkbox-group v-model="addForm.modelTypes" @change="(v) => loadProviderOptions(v?.length ? v : ['chat'])">
            <el-checkbox
              v-for="opt in modelTypeOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.userModel.scenariosHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colBaseUrl')">
          <el-input v-model="addForm.baseUrl" :placeholder="$t('rag.userModel.phOllamaDefault')" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelApiKey')">
          <el-input v-model="addForm.apiKey" type="password" :placeholder="$t('rag.userModel.phApiKeyOpenai')" show-password />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelMaxCtx')">
          <el-input-number v-model="addForm.maxContextTokens" :min="0" :max="1000000" :placeholder="$t('rag.userModel.phMaxCtx0')" style="width: 100%" />
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.userModel.maxCtxHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelCapabilities')">
          <el-checkbox v-model="addForm.supportsDeepThinking">{{ $t('rag.userModel.capDeepThink') }}</el-checkbox>
          <el-checkbox v-model="addForm.supportsToolCall">{{ $t('rag.userModel.capToolCall') }}</el-checkbox>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.userModel.capUiHint') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="doAdd">{{ $t('rag.sysModel.add') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editVisible" :title="$t('rag.userModel.dialogEditTitle')" width="450px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item :label="$t('rag.userModel.colProvider')" required>
          <el-select
            v-model="editForm.provider"
            :placeholder="$t('rag.userModel.phSelectSearch')"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="p in providerOptions"
              :key="p.value"
              :label="p.label"
              :value="p.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colModelName')" required>
          <el-input v-model="editForm.modelName" :placeholder="$t('rag.userModel.phModelExampleR1')" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colScenarios')" required>
          <el-checkbox-group v-model="editForm.modelTypes" @change="(v) => loadProviderOptions(v?.length ? v : ['chat'])">
            <el-checkbox
              v-for="opt in modelTypeOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colBaseUrl')">
          <el-input v-model="editForm.baseUrl" :placeholder="$t('rag.userModel.phOllamaDefault')" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelApiKey')">
          <el-input v-model="editForm.apiKey" type="password" :placeholder="$t('rag.userModel.phApiKeyUnchanged')" show-password />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelMaxCtx')">
          <el-input-number v-model="editForm.maxContextTokens" :min="0" :max="1000000" :placeholder="$t('rag.userModel.phMaxCtx0')" style="width: 100%" />
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.userModel.maxCtxHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelCapabilities')">
          <el-checkbox v-model="editForm.supportsDeepThinking">{{ $t('rag.userModel.capDeepThink') }}</el-checkbox>
          <el-checkbox v-model="editForm.supportsToolCall">{{ $t('rag.userModel.capToolCall') }}</el-checkbox>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.userModel.capUiHint') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="doEdit">{{ $t('rag.docs.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { listLLMProviders, listUserModels, listAvailableProviders, addUserModel, updateUserModel, deleteUserModel, getUserDefaultLLMs, setUserDefaultLLM, clearUserDefaultLLM, listWebSearchProviders, getWebSearchConfig, setWebSearchConfig } from '@/api/rag'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, computed, onMounted, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'RagModel' })

  const { t } = useI18n()

  const modelTypeOptions = computed(() => [
    { value: 'chat', label: t('rag.userModel.typeChat') },
    { value: 'embedding', label: t('rag.userModel.typeEmbedding') },
    { value: 'rerank', label: t('rag.userModel.typeRerank') },
    { value: 'speech2text', label: t('rag.userModel.typeSpeech2text') },
    { value: 'tts', label: t('rag.userModel.typeTts') },
    { value: 'ocr', label: t('rag.userModel.typeOcr') },
    { value: 'cv', label: t('rag.userModel.typeCv') }
  ])

  const formatProviderRowLabel = (m) => {
    if (!m) return ''
    const src = m.source === 'admin' ? t('rag.conv.sourceAdmin') : t('rag.conv.sourceCustom')
    return `${m.name} / ${m.modelName} (${src})`
  }

  const activeTab = ref('providers')
  const providers = ref([])
  const userModels = ref([])
  const addVisible = ref(false)
  const editVisible = ref(false)
  const providerOptions = ref([])
  const addForm = ref({
    provider: 'openai',
    modelName: '',
    modelTypes: ['chat'],
    baseUrl: '',
    apiKey: '',
    maxContextTokens: 10000,
    supportsDeepThinking: false,
    supportsToolCall: true
  })
  const editForm = ref({
    id: 0,
    provider: 'ollama',
    modelName: '',
    modelTypes: ['chat'],
    baseUrl: '',
    apiKey: '',
    maxContextTokens: 0,
    supportsDeepThinking: false,
    supportsToolCall: true
  })
  const userDefaults = ref({}) // { chat: 'admin:1', ... }

  const webSearchProviders = ref([])
  const webSearchForm = ref({ useSystemDefault: true, provider: 'duckduckgo', config: {} })
  const currentWebSearchSchema = ref([])

  const loadProviders = async () => {
    const res = await listLLMProviders()
    if (res.code === 0) providers.value = res.data || []
  }

  const loadUserModels = async () => {
    const res = await listUserModels()
    if (res.code === 0) userModels.value = res.data || []
  }

  onMounted(() => {
    loadProviders()
    loadUserModels()
    loadWebSearchProviders()
  })

  watch(activeTab, async () => {
    if (activeTab.value === 'providers') loadProviders()
    else if (activeTab.value === 'user') loadUserModels()
    else if (activeTab.value === 'defaults') {
      await loadProvidersForDefaults()
      await loadUserDefaults()
    } else if (activeTab.value === 'webSearch') {
      await loadWebSearchConfig()
    }
  })

  const loadWebSearchProviders = async () => {
    const res = await listWebSearchProviders()
    if (res.code === 0 && res.data?.length) {
      webSearchProviders.value = res.data
    }
  }

  const loadWebSearchConfig = async () => {
    await loadWebSearchProviders()
    const res = await getWebSearchConfig()
    if (res.code === 0 && res.data) {
      webSearchForm.value = {
        useSystemDefault: res.data.useSystemDefault !== false,
        provider: res.data.provider || 'duckduckgo',
        config: { ...(res.data.config || {}) }
      }
    }
    const p = webSearchProviders.value.find((x) => x.id === webSearchForm.value.provider)
    currentWebSearchSchema.value = p?.configSchema || []
    ensureConfigKeys()
  }

  const ensureConfigKeys = () => {
    const cfg = webSearchForm.value.config || {}
    for (const f of currentWebSearchSchema.value) {
      if (cfg[f.key] === undefined) cfg[f.key] = ''
    }
    webSearchForm.value.config = cfg
  }

  watch(() => webSearchForm.value.provider, (v) => {
    const p = webSearchProviders.value.find((x) => x.id === v)
    currentWebSearchSchema.value = p?.configSchema || []
    ensureConfigKeys()
  })

  const saveWebSearchConfig = async () => {
    const useSystemDefault = webSearchForm.value.useSystemDefault
    if (useSystemDefault) {
      const res = await setWebSearchConfig({ useSystemDefault: true })
      if (res.code === 0) {
        ElMessage.success(t('rag.userModel.msgSaveOk'))
      }
      return
    }
    const provider = webSearchForm.value.provider
    const schema = webSearchProviders.value.find((x) => x.id === provider)?.configSchema || []
    const config = {}
    for (const f of schema) {
      const v = webSearchForm.value.config?.[f.key]
      if (f.required && !v) {
        ElMessage.warning(t('rag.userModel.fillField', { label: f.label }))
        return
      }
      if (v) config[f.key] = v
    }
    const res = await setWebSearchConfig({ useSystemDefault: false, provider, config })
    if (res.code === 0) {
      ElMessage.success(t('rag.userModel.msgSaveOk'))
    }
  }

  const loadUserDefaults = async () => {
    const res = await getUserDefaultLLMs()
    if (res.code === 0 && res.data) {
      const map = {}
      res.data.forEach((d) => {
        map[d.modelType] = (d.llmSource || 'user') + ':' + d.llmProviderId
      })
      userDefaults.value = map
    }
  }

  const providersCache = ref({})
  const providersForType = (modelType) => {
    return providersCache.value[modelType] || []
  }

  const loadProvidersForDefaults = async () => {
    const types = modelTypeOptions.value.map((x) => x.value)
    const all = {}
    for (const t of types) {
      const res = await listLLMProviders({ scenarioType: t })
      all[t] = res.code === 0 ? (res.data || []) : []
    }
    providersCache.value = all
  }

  const saveUserDefault = async (modelType, value) => {
    if (!value) {
      const res = await clearUserDefaultLLM({ modelType })
      if (res.code === 0) {
        userDefaults.value[modelType] = ''
        ElMessage.success(t('rag.userModel.msgCleared'))
      }
      return
    }
    const [source, idStr] = value.split(':')
    const id = parseInt(idStr, 10)
    if (isNaN(id)) return
    const res = await setUserDefaultLLM({ modelType, llmProviderId: id, llmSource: source || 'user' })
    if (res.code === 0) ElMessage.success(t('rag.userModel.msgSaved'))
  }

  const loadProviderOptions = async (scenarioTypes = ['chat']) => {
    const res = await listAvailableProviders({ scenarioTypes })
    if (res.code === 0 && res.data?.length) {
      providerOptions.value = res.data
      // 若当前选中的 provider 不在新列表中，则选第一个
      const vals = res.data.map((p) => p.value)
      const form = addVisible.value ? addForm.value : editForm.value
      if (form?.provider && !vals.includes(form.provider)) {
        form.provider = vals[0]
      }
    } else {
      providerOptions.value = [{ value: 'openai', label: 'OpenAI' }, { value: 'ollama', label: 'Ollama' }]
    }
  }

  const openAdd = async () => {
    addForm.value = { provider: 'openai', modelName: '', modelTypes: ['chat'], baseUrl: '', apiKey: '', maxContextTokens: 10000, supportsDeepThinking: false, supportsToolCall: true }
    await loadProviderOptions(['chat'])
    addVisible.value = true
  }

  const doAdd = async () => {
    if (!addForm.value.modelName) {
      ElMessage.warning(t('rag.userModel.needModelName'))
      return
    }
    if (!addForm.value.modelTypes?.length) {
      ElMessage.warning(t('rag.userModel.needScenario'))
      return
    }
    const res = await addUserModel(addForm.value)
    if (res.code === 0) {
      ElMessage.success(t('rag.userModel.addOk'))
      addVisible.value = false
      loadProviders()
      loadUserModels()
    }
  }

  const openEdit = async (row) => {
    const types = row.modelTypes && row.modelTypes.length ? row.modelTypes : ['chat']
    editForm.value = {
      id: row.ID,
      provider: (row.provider || '').toLowerCase() || 'ollama',
      modelName: row.modelName || '',
      modelTypes: [...types],
      baseUrl: row.baseUrl || '',
      apiKey: '',
      maxContextTokens: row.maxContextTokens || 0,
      supportsDeepThinking: !!row.supportsDeepThinking,
      supportsToolCall: row.supportsToolCall !== false
    }
    await loadProviderOptions(types)
    editVisible.value = true
  }

  const doEdit = async () => {
    if (!editForm.value.modelName) {
      ElMessage.warning(t('rag.userModel.needModelName'))
      return
    }
    if (!editForm.value.modelTypes?.length) {
      ElMessage.warning(t('rag.userModel.needScenario'))
      return
    }
    const res = await updateUserModel(editForm.value)
    if (res.code === 0) {
      ElMessage.success(t('rag.userModel.updateOk'))
      editVisible.value = false
      loadProviders()
      loadUserModels()
    }
  }

  const deleteModel = (row) => {
    ElMessageBox.confirm(t('rag.userModel.deleteConfirm'), t('rag.userModel.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteUserModel({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('rag.userModel.deleteOk'))
        loadUserModels()
        loadProviders()
      }
    })
  }
</script>
