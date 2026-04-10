<template>
  <div class="h-full lrag-container2 overflow-auto bg-slate-50/60 dark:bg-slate-900">
    <div class="space-y-4 p-4 lg:p-6">
      <section
        class="relative overflow-hidden rounded-xl border border-slate-200/80 bg-white px-5 py-6 shadow-sm dark:border-slate-700 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900"
      >
        
        <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-xs tracking-[0.2em] text-slate-500 dark:text-slate-400">{{ $t('dashboard.label') }}</p>
            <h1 class="mt-2 text-xl font-semibold text-slate-900 dark:text-slate-100 lg:text-2xl">
              {{ $t('dashboard.welcome') }}
            </h1>
            <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">
              {{ sublineText }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <el-button type="primary" @click="goLicense">{{ $t('dashboard.buyLicense') }}</el-button>
          </div>
        </div>
      </section>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <lrag-card>
          <lrag-chart :type="1" :title="$t('dashboard.chartVisitors')" />
        </lrag-card>
        <lrag-card>
          <lrag-chart :type="2" :title="$t('dashboard.chartCustomers')" />
        </lrag-card>
        <lrag-card>
          <lrag-chart :type="3" :title="$t('dashboard.chartResolved')" />
        </lrag-card>
      </div>

      <div class="grid grid-cols-1 items-stretch gap-4 xl:grid-cols-12">
        <div class="grid grid-cols-1 gap-4 content-start xl:col-span-8 xl:h-full">
          <lrag-card :title="$t('dashboard.cardContent')">
            <lrag-chart :type="4" />
          </lrag-card>

          <lrag-card :title="$t('dashboard.cardUpdates')">
            <lrag-table />
          </lrag-card>
        </div>

        <div class="flex flex-col gap-4 xl:col-span-4 xl:h-full">
          <lrag-card :title="$t('dashboard.cardShortcuts')" show-action custom-class="min-h-[300px]">
            <lrag-quick-link />
          </lrag-card>
          <lrag-card :title="$t('dashboard.cardNotice')" show-action custom-class="min-h-[300px]">
            <lrag-notice />
          </lrag-card>
          <lrag-card :title="$t('dashboard.cardDocs')" show-action custom-class="min-h-[120px]">
            <lrag-wiki />
          </lrag-card>
          <div
            class="relative min-h-[200px] flex-1 overflow-hidden rounded-lg border border-slate-200 bg-slate-900 p-5 text-white shadow-sm dark:border-slate-700"
          >
            
            <div class="relative">
              <div class="inline-flex rounded-full bg-white/10 px-3 py-1 text-xs">{{ $t('dashboard.licenseBadge') }}</div>
              <h3 class="mt-3 text-lg font-semibold">{{ $t('dashboard.licenseTitle') }}</h3>
              <p class="mt-2 text-sm text-slate-200/90">
                {{ $t('dashboard.licenseDesc') }}
              </p>
              <div class="mt-4 flex flex-wrap gap-2 text-xs">
                <span class="rounded-full bg-white/10 px-2.5 py-1">{{ $t('dashboard.licenseTag1') }}</span>
                <span class="rounded-full bg-white/10 px-2.5 py-1">{{ $t('dashboard.licenseTag2') }}</span>
                <span class="rounded-full bg-white/10 px-2.5 py-1">{{ $t('dashboard.licenseTag3') }}</span>
              </div>
              <div class="mt-5 flex items-center gap-3">
                <el-button type="primary" @click="goLicense">{{ $t('dashboard.buyNow') }}</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    LragTable,
    LragChart,
    LragWiki,
    LragNotice,
    LragQuickLink,
    LragCard
  } from './components'

  const { t, locale } = useI18n()

  const dateLocaleTag = computed(() => {
    const m = {
      en: 'en-US',
      'zh-CN': 'zh-CN',
      'zh-TW': 'zh-TW',
      ja: 'ja-JP',
      ko: 'ko-KR',
      fr: 'fr-FR',
      de: 'de-DE',
      es: 'es-ES',
      it: 'it-IT',
      'pt-BR': 'pt-BR',
      ru: 'ru-RU',
      vi: 'vi-VN',
      th: 'th-TH',
      id: 'id-ID'
    }
    return m[locale.value] || 'en-US'
  })

  const today = computed(() => {
    try {
      const d = new Date()
      return d.toLocaleDateString(dateLocaleTag.value, {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    } catch (e) {
      return new Date().toISOString().slice(0, 10)
    }
  })

  const sublineText = computed(() =>
    t('dashboard.subline', { date: today.value })
  )

  const goLicense = () => {
    window.open('https://plugin.LightningRAG.com/license', '_blank', 'noopener,noreferrer')
  }

  defineOptions({
    name: 'Dashboard'
  })
</script>

<style lang="scss" scoped></style>

