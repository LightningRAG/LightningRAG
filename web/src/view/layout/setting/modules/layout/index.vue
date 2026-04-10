<template>
  <div class="lrag-theme-font">
    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.layoutPrefs.layoutModeTitle') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <LayoutModeCard
          v-model="config.side_mode"
          @update:modelValue="appStore.toggleSideMode"
        />
      </div>
    </div>

    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.layoutPrefs.uiConfigTitle') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <SettingItem :label="$t('settings.layoutPrefs.showTabs')">
            <template #suffix>
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">{{ $t('settings.layoutPrefs.showTabsHint') }}</span>
            </template>
            <el-switch
              v-model="config.showTabs"
              @change="appStore.toggleTabs"
            />
          </SettingItem>

          <SettingItem :label="$t('settings.layoutPrefs.pageTransition')">
            <template #suffix>
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">{{ $t('settings.layoutPrefs.pageTransitionHint') }}</span>
            </template>
            <el-select
              v-model="config.transition_type"
              @change="appStore.toggleTransition"
              class="w-32"
              size="small"
            >
              <el-option value="fade" :label="$t('settings.layoutPrefs.animFade')" />
              <el-option value="slide" :label="$t('settings.layoutPrefs.animSlide')" />
              <el-option value="zoom" :label="$t('settings.layoutPrefs.animZoom')" />
              <el-option value="none" :label="$t('settings.layoutPrefs.animNone')" />
            </el-select>
          </SettingItem>
        </div>
      </div>
    </div>

    <div class="mb-10">
      <div class="lrag-theme-section-header">
        <div class="lrag-theme-divider"></div>
        <span class="lrag-theme-section-title">{{ $t('settings.layoutPrefs.sizeConfigTitle') }}</span>
        <div class="lrag-theme-divider"></div>
      </div>

      <div class="lrag-theme-section-content">
        <div class="lrag-theme-card-bg">
          <div class="space-y-4">
            <div class="lrag-theme-card-white">
              <div class="flex items-center justify-between">
                <div>
                  <h4 class="text-sm font-medium lrag-theme-text-main">{{ $t('settings.layoutPrefs.sidebarWidthExpanded') }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('settings.layoutPrefs.sidebarWidthExpandedDesc') }}</p>
                </div>
                <div class="flex items-center gap-2">
                  <el-input-number
                    v-model="config.layout_side_width"
                    :min="150"
                    :max="400"
                    :step="10"
                    size="small"
                    class="w-24"
                  />
                  <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ $t('settings.layoutPrefs.unitPx') }}</span>
                </div>
              </div>
            </div>

            <div class="lrag-theme-card-white">
              <div class="flex items-center justify-between">
                <div>
                  <h4 class="text-sm font-medium lrag-theme-text-main">{{ $t('settings.layoutPrefs.sidebarWidthCollapsed') }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('settings.layoutPrefs.sidebarWidthCollapsedDesc') }}</p>
                </div>
                <div class="flex items-center gap-2">
                  <el-input-number
                    v-model="config.layout_side_collapsed_width"
                    :min="60"
                    :max="100"
                    size="small"
                    class="w-24"
                  />
                  <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ $t('settings.layoutPrefs.unitPx') }}</span>
                </div>
              </div>
            </div>

            <div class="lrag-theme-card-white">
              <div class="flex items-center justify-between">
                <div>
                  <h4 class="text-sm font-medium lrag-theme-text-main">{{ $t('settings.layoutPrefs.menuItemHeight') }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('settings.layoutPrefs.menuItemHeightDesc') }}</p>
                </div>
                <div class="flex items-center gap-2">
                  <el-input-number
                    v-model="config.layout_side_item_height"
                    :min="30"
                    :max="50"
                    size="small"
                    class="w-24"
                  />
                  <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ $t('settings.layoutPrefs.unitPx') }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/pinia'
import LayoutModeCard from '../../components/layoutModeCard.vue'
import SettingItem from '../../components/settingItem.vue'

defineOptions({
  name: 'LayoutSettings'
})

const appStore = useAppStore()
const { config } = storeToRefs(appStore)
</script>


