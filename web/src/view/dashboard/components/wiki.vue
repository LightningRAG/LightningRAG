<template>
  <div class="grid grid-cols-2 gap-2">
    <a
      v-for="item in wikis"
      :key="item.url"
      :href="item.url"
      class="text-sm text-black/70 dark:text-white/70 no-underline hover:text-[var(--el-color-primary)] dark:hover:text-white"
      target="_blank"
    >
      {{ item.title }}
    </a>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const wikis = computed(() => [
  {
    title: t('dashboard.wikiVue3'),
    url: 'https://v3.cn.vuejs.org/guide/introduction.html'
  },
  {
    title: t('dashboard.wikiGin'),
    url: 'https://gin-gonic.com/'
  },
  {
    title: t('dashboard.wikiLrag'),
    url: 'https://lightningrag.com/'
  },
  {
    title: t('dashboard.wikiGithub'),
    url: 'https://github.com/LightningRAG/LightningRAG'
  }
])
</script>

<style scoped lang="scss"></style>
