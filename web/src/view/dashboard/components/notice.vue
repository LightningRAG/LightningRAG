<template>
  <div class="space-y-3">
    <el-scrollbar max-height="320px">
      <div class="space-y-2 pr-1">
        <div
          v-for="(item, index) in notices"
          :key="index"
          class="group rounded-lg border border-black/10 bg-white/70 p-3 transition-all duration-200 hover:-translate-y-0.5 hover:border-black/20 hover:shadow-sm dark:border-white/10 dark:bg-white/[0.02] dark:hover:border-white/20"
        >
          <div class="flex items-start gap-3">
            <span class="mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full" :class="item.dotClass" />
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <span class="rounded-md px-2 py-0.5 text-[11px] font-semibold leading-4" :class="item.tagClass">
                  {{ item.typeTitle }}
                </span>
                <span class="shrink-0 text-[11px] text-black/45 dark:text-white/45">{{ item.time }}</span>
              </div>
              <el-tooltip effect="light" :content="item.title" placement="top">
                <p class="mt-1.5 line-clamp-2 text-sm text-black/75 dark:text-white/75">
                  {{ item.title }}
                </p>
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const notices = computed(() => [
  {
    typeTitle: t('dashboard.noticeTypes.notice'),
    time: t('dashboard.noticeTimeToday'),
    title: t('dashboard.notice1'),
    dotClass: 'bg-cyan-500',
    tagClass: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-200'
  },
  {
    typeTitle: t('dashboard.noticeTypes.activity'),
    time: t('dashboard.noticeTime2d'),
    title: t('dashboard.notice2'),
    dotClass: 'bg-emerald-500',
    tagClass: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-200'
  },
  {
    typeTitle: t('dashboard.noticeTypes.compliance'),
    time: t('dashboard.noticeTime3d'),
    title: t('dashboard.notice3'),
    dotClass: 'bg-amber-500',
    tagClass: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-200'
  },
  {
    typeTitle: t('dashboard.noticeTypes.service'),
    time: t('dashboard.noticeTime5d'),
    title: t('dashboard.notice4'),
    dotClass: 'bg-violet-500',
    tagClass: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-200'
  }
])
</script>

<style scoped lang="scss"></style>
