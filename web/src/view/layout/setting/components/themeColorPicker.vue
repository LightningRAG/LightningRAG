<template>
  <div class="lrag-theme-font">
    <div class="lrag-theme-card-bg p-4">
      <div class="mb-4">
        <p class="text-base font-semibold text-gray-700 dark:text-gray-300 mb-4">{{ $t('themeColors.curated') }}</p>
        <div class="grid grid-cols-3 gap-4">
          <div
            v-for="colorItem in presetColors"
            :key="colorItem.color"
            class="flex items-center gap-4 p-2 bg-white dark:bg-gray-700 border-2 border-gray-200 dark:border-gray-600 rounded-xl cursor-pointer transition-all duration-150 ease-in-out hover:transform hover:-translate-y-1 hover:shadow-lg"
            :class="{
              'ring-2 ring-offset-2 ring-offset-gray-50 dark:ring-offset-gray-800 transform -translate-y-1 shadow-lg': modelValue === colorItem.color
            }"
            :style="modelValue === colorItem.color ? {
              borderColor: colorItem.color,
              ringColor: colorItem.color + '40'
            } : {}"
            @click="handleColorChange(colorItem.color)"
          >
            <div
              class="relative w-10 h-10 rounded-lg border border-gray-300 dark:border-gray-500 flex-shrink-0 shadow-sm"
              :style="{ backgroundColor: colorItem.color }"
            >
              <div
                v-if="modelValue === colorItem.color"
                class="absolute inset-0 flex items-center justify-center text-white text-base"
                style="text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);"
              >
                <el-icon>
                  <Check />
                </el-icon>
              </div>
            </div>
            <div class="min-w-0 flex-1">
              <span class="block text-sm font-semibold lrag-theme-text-main">{{ colorItem.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-between p-4 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl mb-6 shadow-sm">
        <div class="flex-1">
          <h4 class="text-base font-semibold lrag-theme-text-main">{{ $t('themeColors.customTitle') }}</h4>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ $t('themeColors.customHint') }}</p>
        </div>
        <el-color-picker
          v-model="customColor"
          size="large"
          :predefine="presetColors.map(item => item.color)"
          @change="handleCustomColorChange"
          class="custom-color-picker"
        />
      </div>

      <div class="bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl p-4 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-base font-semibold text-gray-700 dark:text-gray-300">{{ $t('themeColors.current') }}</span>
          <div class="flex items-center gap-3">
            <div
              class="w-6 h-6 rounded-lg border border-gray-300 dark:border-gray-500 shadow-sm"
              :style="{ backgroundColor: modelValue }"
            ></div>
            <code class="text-sm font-mono bg-gray-100 dark:bg-gray-600 text-gray-700 dark:text-gray-300 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-500">
              {{ modelValue }}
            </code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { Check } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

defineOptions({
  name: 'ThemeColorPicker'
})

const props = defineProps({
  modelValue: {
    type: String,
    default: '#3b82f6'
  }
})

const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()

const customColor = ref(props.modelValue)

const presetColorDefs = [
  { color: '#4E80EE', key: 'presetDefault' },
  { color: '#8bb5d1', key: 'presetMorningMist' },
  { color: '#a8c8a8', key: 'presetMintGreen' },
  { color: '#d4a5a5', key: 'presetRosePink' },
  { color: '#c8a8d8', key: 'presetLavender' },
  { color: '#f0c674', key: 'presetWarmYellow' },
  { color: '#b8b8b8', key: 'presetMoonSilver' },
  { color: '#d8a8a8', key: 'presetCoralOrange' },
  { color: '#a8d8d8', key: 'presetSeaMist' },
  { color: '#c8c8a8', key: 'presetOliveGreen' },
  { color: '#d8c8a8', key: 'presetMilkTea' },
  { color: '#a8a8d8', key: 'presetDreamPurple' },
  { color: '#c8d8a8', key: 'presetMatchaGreen' }
]

const presetColors = computed(() =>
  presetColorDefs.map((p) => ({
    color: p.color,
    name: t(`themeColors.${p.key}`)
  }))
)

const handleColorChange = (color) => {
  customColor.value = color
  emit('update:modelValue', color)
}

const handleCustomColorChange = (color) => {
  if (color) {
    emit('update:modelValue', color)
  }
}

watch(() => props.modelValue, (newValue) => {
  customColor.value = newValue
})
</script>

<style scoped>


.custom-color-picker {
  ::v-deep(.el-color-picker__trigger) {
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    transition: all 150ms ease-in-out;

    &:hover {
      border-color: #9ca3af;
      transform: translateY(-1px);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
  }
}

.dark .custom-color-picker {
  ::v-deep(.el-color-picker__trigger) {
    border-color: #4b5563;

    &:hover {
      border-color: #6b7280;
    }
  }
}
</style>
