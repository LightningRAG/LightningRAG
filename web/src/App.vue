<template>
  <div
    id="app"
    class="bg-gray-50 text-slate-700 !dark:text-slate-500 dark:bg-slate-800"
  >
    <el-config-provider :locale="elementLocale" :size="appStore.config.global_size">
      <router-view />
      <Application />
    </el-config-provider>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { elementLocaleMap } from '@/locale/element-locale'
  import Application from '@/components/application/index.vue'
  import { useAppStore } from '@/pinia'

  const appStore = useAppStore()
  const { locale } = useI18n()
  const elementLocale = computed(
    () => elementLocaleMap[locale.value] || elementLocaleMap.en
  )
  defineOptions({
    name: 'App'
  })
</script>
<style lang="scss">
  // 引入初始化样式
  #app {
    height: 100vh;
    overflow: hidden;
    font-weight: 400 !important;
  }

  .el-button {
    font-weight: 400 !important;
  }

  .lrag-body-h {
    min-height: calc(100% - 3rem);
  }

  .lrag-container {
    height: calc(100% - 2.5rem);
  }

  .lrag-container2 {
    height: calc(100% - 4.5rem);
  }
</style>
