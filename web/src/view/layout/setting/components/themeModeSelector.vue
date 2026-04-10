<template>
  <div class="flex justify-center">
    <div class="lrag-theme-mode-selector">
      <div
        v-for="mode in themeModes"
        :key="mode.value"
        class="lrag-theme-mode-item"
        :class="[
          modelValue === mode.value
            ? 'text-white shadow-sm transform -translate-y-0.5'
            : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
        ]"
        :style="modelValue === mode.value ? { backgroundColor: primaryColor } : {}"
        @click="handleModeChange(mode.value)"
      >
        <el-icon class="text-lg mb-1">
          <component :is="mode.icon" />
        </el-icon>
        <span class="text-xs font-medium">{{ mode.label }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { Sunny, Moon, Monitor } from '@element-plus/icons-vue'
import { useAppStore } from '@/pinia'
import { useI18n } from 'vue-i18n'

defineOptions({
  name: 'ThemeModeSelector'
})

defineProps({
  modelValue: {
    type: String,
    default: 'auto'
  }
})

const emit = defineEmits(['update:modelValue'])

const appStore = useAppStore()
const { config } = storeToRefs(appStore)
const { t } = useI18n()

const primaryColor = computed(() => config.value.primaryColor)

const themeModes = computed(() => [
  {
    value: 'light',
    label: t('settings.appearance.themeLight'),
    icon: Sunny
  },
  {
    value: 'dark',
    label: t('settings.appearance.themeDark'),
    icon: Moon
  },
  {
    value: 'auto',
    label: t('settings.appearance.themeAuto'),
    icon: Monitor
  }
])

const handleModeChange = (mode) => {
  emit('update:modelValue', mode)
}
</script>
