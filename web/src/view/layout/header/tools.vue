<template>
  <div class="flex items-center mx-4 gap-4">
    <el-tooltip class="" effect="dark" :content="$t('layout.tools.search')" placement="bottom">
        <span class="w-8 h-8 p-2 rounded-full flex items-center justify-center shadow border border-gray-200 dark:border-gray-600 cursor-pointer border-solid">
        <el-icon
            @click="handleCommand"
        >
        <Search />
      </el-icon>
        </span>

    </el-tooltip>

    <el-tooltip class="" effect="dark" :content="$t('layout.tools.systemSettings')" placement="bottom">
        <span class="w-8 h-8 p-2 rounded-full flex items-center justify-center shadow border border-gray-200 dark:border-gray-600 cursor-pointer border-solid">
         <el-icon
             @click="toggleSetting"
         >
        <Setting />
      </el-icon>
        </span>

    </el-tooltip>

    <el-tooltip class="" effect="dark" :content="$t('layout.tools.refresh')" placement="bottom">
      <span class="w-8 h-8 p-2 rounded-full flex items-center justify-center shadow border border-gray-200 dark:border-gray-600 cursor-pointer border-solid">
      <el-icon
          :class="showRefreshAnmite ? 'animate-spin' : ''"
          @click="toggleRefresh"
      >
        <Refresh />
      </el-icon>
      </span>

    </el-tooltip>

    <el-tooltip
      class=""
      effect="dark"
      :content="$t('settings.language.label')"
      placement="bottom"
    >
      <div class="inline-block">
        <el-dropdown trigger="click" @command="handleLocaleChange">
          <span
            class="w-8 h-8 p-2 rounded-full flex items-center justify-center shadow border border-gray-200 dark:border-gray-600 cursor-pointer border-solid"
          >
            <el-icon>
              <Icon icon="mdi:web" />
            </el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="opt in localeMenuOptions"
                :key="opt.value"
                :command="opt.value"
              >
                <span class="flex items-center justify-between gap-6 min-w-[10rem]">
                  <span>{{ $t(opt.labelKey) }}</span>
                  <el-icon v-if="locale === opt.value" class="text-[var(--el-color-primary)]">
                    <Check />
                  </el-icon>
                </span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-tooltip>

    <el-tooltip
      class=""
      effect="dark"
      :content="$t('layout.tools.toggleTheme')"
      placement="bottom"
    >
      <span class="w-8 h-8 p-2 rounded-full flex items-center justify-center shadow border border-gray-200 dark:border-gray-600 cursor-pointer border-solid">
        <el-icon
            v-if="appStore.isDark"
            @click="appStore.toggleTheme(false)"
        >
        <Sunny />
      </el-icon>
      <el-icon
          v-else
          @click="appStore.toggleTheme(true)"
      >
        <Moon />
      </el-icon>
      </span>

    </el-tooltip>

    <lrag-setting v-model:drawer="showSettingDrawer"></lrag-setting>
    <command-menu ref="command" />
  </div>
</template>

<script setup>
  import { useAppStore } from '@/pinia'
  import LragSetting from '@/view/layout/setting/index.vue'
  import { ref, onBeforeUnmount } from 'vue'
  import { emitter } from '@/utils/bus.js'
  import CommandMenu from '@/components/commandMenu/index.vue'
  import { useI18n } from 'vue-i18n'
  import { Check } from '@element-plus/icons-vue'
  import { Icon } from '@iconify/vue'
  import { setLocale } from '@/locale'

  const { locale } = useI18n()

  const localeMenuOptions = [
    { value: 'en', labelKey: 'settings.language.optionEn' },
    { value: 'zh-CN', labelKey: 'settings.language.optionZhCN' },
    { value: 'zh-TW', labelKey: 'settings.language.optionZhTW' },
    { value: 'ja', labelKey: 'settings.language.optionJa' },
    { value: 'ko', labelKey: 'settings.language.optionKo' },
    { value: 'fr', labelKey: 'settings.language.optionFr' },
    { value: 'de', labelKey: 'settings.language.optionDe' },
    { value: 'es', labelKey: 'settings.language.optionEs' },
    { value: 'it', labelKey: 'settings.language.optionIt' },
    { value: 'pt-BR', labelKey: 'settings.language.optionPtBR' },
    { value: 'ru', labelKey: 'settings.language.optionRu' },
    { value: 'vi', labelKey: 'settings.language.optionVi' },
    { value: 'th', labelKey: 'settings.language.optionTh' },
    { value: 'id', labelKey: 'settings.language.optionId' }
  ]

  const handleLocaleChange = (code) => setLocale(code)

  const appStore = useAppStore()
  const showSettingDrawer = ref(false)
  const showRefreshAnmite = ref(false)
  const toggleRefresh = () => {
    showRefreshAnmite.value = true
    emitter.emit('reload')
    setTimeout(() => {
      showRefreshAnmite.value = false
    }, 1000)
  }

  const toggleSetting = () => {
    showSettingDrawer.value = true
  }

  const first = ref('')
  const command = ref()

  const handleCommand = () => {
    command.value.open()
  }
  const initPage = () => {
    // 判断当前用户的操作系统
    if (window.localStorage.getItem('osType') === 'WIN') {
      first.value = 'Ctrl'
    } else {
      first.value = '⌘'
    }
    window.addEventListener('keydown', handleKeyDown)
  }

  const handleKeyDown = (e) => {
    if (e.ctrlKey && e.key === 'k') {
      e.preventDefault()
      handleCommand()
    }
  }

  initPage()

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyDown)
  })
</script>

<style scoped lang="scss"></style>
