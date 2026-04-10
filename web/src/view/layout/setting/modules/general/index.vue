<template>
  <div class="lrag-theme-font">
    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.language.label') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>
      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
            {{ $t('settings.language.hint') }}
          </p>
          <el-select
            :model-value="locale"
            class="w-full max-w-xs"
            @change="handleLocaleChange"
          >
            <el-option :label="$t('settings.language.optionEn')" value="en" />
            <el-option :label="$t('settings.language.optionZhCN')" value="zh-CN" />
            <el-option :label="$t('settings.language.optionZhTW')" value="zh-TW" />
            <el-option :label="$t('settings.language.optionJa')" value="ja" />
            <el-option :label="$t('settings.language.optionKo')" value="ko" />
            <el-option :label="$t('settings.language.optionFr')" value="fr" />
            <el-option :label="$t('settings.language.optionDe')" value="de" />
            <el-option :label="$t('settings.language.optionEs')" value="es" />
            <el-option :label="$t('settings.language.optionIt')" value="it" />
            <el-option :label="$t('settings.language.optionPtBR')" value="pt-BR" />
            <el-option :label="$t('settings.language.optionRu')" value="ru" />
            <el-option :label="$t('settings.language.optionVi')" value="vi" />
            <el-option :label="$t('settings.language.optionTh')" value="th" />
            <el-option :label="$t('settings.language.optionId')" value="id" />
          </el-select>
        </div>
      </div>
    </div>

    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.general.systemInfo') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div class="flex justify-between items-center py-3 border-b border-gray-200 dark:border-gray-600">
              <span class="lrag-theme-text-sub font-medium">{{ $t('settings.general.version') }}</span>
              <span class="font-mono lrag-theme-text-main font-semibold">v2.7.4</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-gray-200 dark:border-gray-600">
              <span class="lrag-theme-text-sub font-medium">{{ $t('settings.general.frontendFramework') }}</span>
              <span class="font-mono lrag-theme-text-main font-semibold">Vue 3</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-gray-200 dark:border-gray-600">
              <span class="lrag-theme-text-sub font-medium">{{ $t('settings.general.uiLibrary') }}</span>
              <span class="font-mono lrag-theme-text-main font-semibold">Element Plus</span>
            </div>
            <div class="flex justify-between items-center py-3 border-b border-gray-200 dark:border-gray-600">
              <span class="lrag-theme-text-sub font-medium">{{ $t('settings.general.buildTool') }}</span>
              <span class="font-mono lrag-theme-text-main font-semibold">Vite</span>
            </div>
            <div class="flex justify-between items-center py-3">
              <span class="lrag-theme-text-sub font-medium">{{ $t('settings.general.browser') }}</span>
              <span class="font-mono lrag-theme-text-main font-semibold">{{ browserInfo }}</span>
            </div>
            <div class="flex justify-between items-center py-3">
              <span class="lrag-theme-text-sub font-medium">{{ $t('settings.general.screenResolution') }}</span>
              <span class="font-mono lrag-theme-text-main font-semibold">{{ screenResolution }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.general.configManagement') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <div class="space-y-5">
            <div
              class="lrag-theme-card-white flex items-center justify-between">
              <div class="flex items-center gap-4">
                <div
                  class="w-12 h-12 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl flex items-center justify-center text-red-600 dark:text-red-400 text-xl">
                  🔄
                </div>
                <div>
                  <h4 class="text-sm font-semibold lrag-theme-text-main">{{ $t('settings.general.resetConfig') }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('settings.general.resetConfigDesc') }}</p>
                </div>
              </div>
              <el-button type="danger" size="small"
                class="rounded-lg font-medium transition-all duration-150 ease-in-out hover:-translate-y-0.5"
                @click="handleResetConfig">
                {{ $t('settings.general.resetConfigBtn') }}
              </el-button>
            </div>

            <div
              class="lrag-theme-card-white flex items-center justify-between">
              <div class="flex items-center gap-4">
                <div
                  class="w-12 h-12 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl flex items-center justify-center text-blue-600 dark:text-blue-400 text-xl">
                  📤
                </div>
                <div>
                  <h4 class="text-sm font-semibold lrag-theme-text-main">{{ $t('settings.general.exportConfig') }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('settings.general.exportConfigDesc') }}</p>
                </div>
              </div>
              <el-button type="primary" size="small"
                class="rounded-lg font-medium transition-all duration-150 ease-in-out hover:-translate-y-0.5"
                :style="{ backgroundColor: config.primaryColor, borderColor: config.primaryColor }"
                @click="handleExportConfig">
                {{ $t('settings.general.exportConfigBtn') }}
              </el-button>
            </div>

            <div
              class="lrag-theme-card-white flex items-center justify-between">
              <div class="flex items-center gap-4">
                <div
                  class="w-12 h-12 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-xl flex items-center justify-center text-green-600 dark:text-green-400 text-xl">
                  📥
                </div>
                <div>
                  <h4 class="text-sm font-semibold lrag-theme-text-main">{{ $t('settings.general.importConfig') }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('settings.general.importConfigDesc') }}</p>
                </div>
              </div>
              <el-upload ref="uploadRef" :auto-upload="false" :show-file-list="false" accept=".json"
                @change="handleImportConfig">
                <el-button type="success" size="small"
                  class="rounded-lg font-medium transition-all duration-150 ease-in-out hover:-translate-y-0.5">
                  {{ $t('settings.general.importConfigBtn') }}
                </el-button>
              </el-upload>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.general.aboutProject') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <div class="flex items-start gap-5">
            <div
              class="w-16 h-16 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-sm">
              <Logo />
            </div>
            <div class="flex-1">
              <h4 class="text-xl font-semibold lrag-theme-text-main mb-3">{{ $t('settings.general.projectName') }}</h4>
              <p class="text-sm lrag-theme-text-sub mb-5 leading-relaxed">
                {{ $t('settings.general.projectDesc') }}
              </p>
              <div class="flex items-center gap-3 text-sm">
                <a href="https://github.com/LightningRAG/LightningRAG" target="_blank"
                  class="font-medium transition-colors duration-150 hover:underline"
                  :style="{ color: config.primaryColor }">
                  {{ $t('settings.general.githubRepo') }}
                </a>
                <span class="text-gray-400 dark:text-gray-500">·</span>
                <a href="https://lightningrag.com/" target="_blank"
                  class="font-medium transition-colors duration-150 hover:underline"
                  :style="{ color: config.primaryColor }">
                  {{ $t('settings.general.officialDocs') }}
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/pinia'
import Logo from '@/components/logo/index.vue'
import { setLocale } from '@/locale'

defineOptions({
  name: 'GeneralSettings'
})

const { t, locale } = useI18n()
const appStore = useAppStore()
const { config } = storeToRefs(appStore)
const uploadRef = ref()

const handleLocaleChange = (code) => setLocale(code)

const browserInfo = ref('')
const screenResolution = ref('')

onMounted(() => {
  const userAgent = navigator.userAgent
  if (userAgent.includes('Chrome')) {
    browserInfo.value = 'Chrome'
  } else if (userAgent.includes('Firefox')) {
    browserInfo.value = 'Firefox'
  } else if (userAgent.includes('Safari')) {
    browserInfo.value = 'Safari'
  } else if (userAgent.includes('Edge')) {
    browserInfo.value = 'Edge'
  } else {
    browserInfo.value = t('settings.general.browserUnknown')
  }

  screenResolution.value = `${screen.width}×${screen.height}`
})

const handleResetConfig = async () => {
  try {
    await ElMessageBox.confirm(
      t('settings.general.confirmResetBody'),
      t('settings.general.confirmResetTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    )

    appStore.resetConfig()
    ElMessage.success(t('settings.general.configResetOk'))
  } catch {
    // User cancelled
  }
}

const handleExportConfig = () => {
  const configData = JSON.stringify(config.value, null, 2)
  const blob = new Blob([configData], { type: 'application/json' })
  const url = URL.createObjectURL(blob)

  const link = document.createElement('a')
  link.href = url
  link.download = `lightningrag-config-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)

  ElMessage.success(t('settings.general.configExported'))
}

const handleImportConfig = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const importedConfig = JSON.parse(e.target.result)

      Object.keys(importedConfig).forEach(key => {
        if (key in config.value) {
          config.value[key] = importedConfig[key]
        }
      })

      ElMessage.success(t('settings.general.configImported'))
    } catch (error) {
      ElMessage.error(t('settings.general.invalidConfigFile'))
    }
  }
  reader.readAsText(file.raw)
}
</script>


