<template>
  <div>
    <div class="text-gray-500 text-sm mb-4">{{ $t('admin.roleDefaultLlm.intro') }}</div>
    <el-form label-width="120px" class="max-w-md">
      <el-form-item v-for="opt in modelTypeOptions" :key="opt.value" :label="opt.label + $t('admin.roleDefaultLlm.defaultSuffix')">
        <el-select
          v-model="defaults[opt.value]"
          :placeholder="$t('admin.roleDefaultLlm.phSelectModel')"
          clearable
          style="width: 100%"
          @change="(v) => saveDefault(opt.value, v)"
        >
          <el-option
            v-for="m in providersForType(opt.value)"
            :key="(m.source || 'user') + '-' + m.id"
            :label="`${m.name} / ${m.modelName} (${m.source === 'admin' ? $t('admin.roleDefaultLlm.sourceAdmin') : $t('admin.roleDefaultLlm.sourceCustom')})`"
            :value="(m.source || 'user') + ':' + m.id"
          />
        </el-select>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
  import { listLLMProviders, getAuthorityDefaultLLMs, setAuthorityDefaultLLM, clearAuthorityDefaultLLM } from '@/api/rag'
  import { ref, watch, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'DefaultLLMs' })

  const { t } = useI18n()

  const props = defineProps({
    row: { type: Object, default: () => ({}) }
  })

  const emit = defineEmits(['changeRow'])

  const modelTypeOptions = computed(() => [
    { value: 'chat', label: t('admin.roleDefaultLlm.typeChat') },
    { value: 'embedding', label: t('admin.roleDefaultLlm.typeEmbedding') },
    { value: 'rerank', label: t('admin.roleDefaultLlm.typeRerank') }
  ])

  const defaults = ref({})
  const providersCache = ref({})

  const providersForType = (modelType) => providersCache.value[modelType] || []

  const loadProviders = async () => {
    const all = {}
    for (const opt of modelTypeOptions.value) {
      const res = await listLLMProviders({ scenarioType: opt.value })
      all[opt.value] = res.code === 0 ? (res.data || []) : []
    }
    providersCache.value = all
  }

  const loadDefaults = async () => {
    if (!props.row?.authorityId) return
    const res = await getAuthorityDefaultLLMs({ authorityId: props.row.authorityId })
    if (res.code === 0 && res.data) {
      const map = {}
      res.data.forEach((d) => {
        map[d.modelType] = (d.llmSource || 'admin') + ':' + d.llmProviderId
      })
      defaults.value = map
    }
  }

  const saveDefault = async (modelType, value) => {
    if (!props.row?.authorityId) return
    if (!value) {
      const res = await clearAuthorityDefaultLLM({ authorityId: props.row.authorityId, modelType })
      if (res.code === 0) {
        defaults.value[modelType] = ''
        ElMessage.success(t('admin.roleDefaultLlm.cleared'))
      }
      return
    }
    const [source, idStr] = value.split(':')
    const id = parseInt(idStr, 10)
    if (isNaN(id)) return
    const res = await setAuthorityDefaultLLM({
      authorityId: props.row.authorityId,
      modelType,
      llmProviderId: id,
      llmSource: source || 'admin'
    })
    if (res.code === 0) ElMessage.success(t('admin.roleDefaultLlm.saved'))
  }

  watch(
    () => props.row?.authorityId,
    async () => {
      await loadProviders()
      await loadDefaults()
    },
    { immediate: true }
  )
</script>
