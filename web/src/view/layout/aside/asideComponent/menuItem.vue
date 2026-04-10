<template>
  <el-menu-item
    :index="routerInfo.name"
    :style="{
          height: sideHeight
        }"
  >
    <el-icon v-if="routerInfo.meta.icon">
      <component :is="routerInfo.meta.icon" />
    </el-icon>
    <template v-else>
      {{ isCollapse ? displayTitle[0] : "" }}
    </template>
    <template #title>
      {{ displayTitle }}
    </template>
  </el-menu-item>
</template>

<script setup>
import { computed, inject } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/pinia'
import { storeToRefs } from 'pinia'
import { resolveMenuTitle } from '@/utils/menuTitle'
import { useI18n } from 'vue-i18n'

  const { locale } = useI18n()
  const appStore = useAppStore()
  const { config } = storeToRefs(appStore)
  const route = useRoute()

  defineOptions({
    name: 'MenuItem'
  })

  const props = defineProps({
    routerInfo: {
      default: function () {
        return null
      },
      type: Object
    }
  })

  const isCollapse = inject('isCollapse', {
    default: false
  })

  const sideHeight = computed(() => {
    return config.value.layout_side_item_height + 'px'
  })

  const displayTitle = computed(() => {
    void locale.value
    return resolveMenuTitle(props.routerInfo?.meta, route)
  })
</script>

<style lang="scss"></style>
