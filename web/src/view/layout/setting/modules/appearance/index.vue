<template>
  <div class="lrag-theme-font">
    <!-- Theme Mode Section -->
    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.appearance.themeModeTitle') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <ThemeModeSelector v-model="config.darkMode" @update:modelValue="appStore.toggleDarkMode" />
      </div>
    </div>

    <!-- Theme Color Section -->
    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.appearance.themeColorTitle') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <ThemeColorPicker v-model="config.primaryColor" @update:modelValue="appStore.togglePrimaryColor" />
      </div>
    </div>

    <!-- Global Size Section -->
    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.appearance.globalSizeSection') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <SettingItem :label="$t('settings.appearance.globalSizeLabel')">
            <template #suffix>
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">{{ $t('settings.appearance.globalSizeHint') }}</span>
            </template>
            <div class="w-39">
              <el-select v-model="config.global_size" :placeholder="$t('settings.appearance.selectPlaceholder')" @change="appStore.toggleGlobalSize">
                <el-option :label="$t('settings.appearance.sizeDefault')" value="default" />
                <el-option :label="$t('settings.appearance.sizeLarge')" value="large" />
                <el-option :label="$t('settings.appearance.sizeSmall')" value="small" />
              </el-select>
            </div>
          </SettingItem>
        </div>
      </div>
    </div>

    <!-- Visual Accessibility Section -->
    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.appearance.visualTitle') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <SettingItem :label="$t('settings.appearance.greyMode')">
            <template #suffix>
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">{{ $t('settings.appearance.greyHint') }}</span>
            </template>
            <el-switch v-model="config.grey" @change="appStore.toggleGrey" />
          </SettingItem>

          <SettingItem :label="$t('settings.appearance.weaknessMode')">
            <template #suffix>
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">{{ $t('settings.appearance.weaknessHint') }}</span>
            </template>
            <el-switch v-model="config.weakness" @change="appStore.toggleWeakness" />
          </SettingItem>

          <SettingItem :label="$t('settings.appearance.watermark')">
            <template #suffix>
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">{{ $t('settings.appearance.watermarkHint') }}</span>
            </template>
            <el-switch v-model="config.show_watermark" @change="appStore.toggleConfigWatermark" />
          </SettingItem>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/pinia'
import ThemeModeSelector from '../../components/themeModeSelector.vue'
import ThemeColorPicker from '../../components/themeColorPicker.vue'
import SettingItem from '../../components/settingItem.vue'

defineOptions({
  name: 'AppearanceSettings'
})

const appStore = useAppStore()
const { config } = storeToRefs(appStore)
</script>


