<template>
  <el-sub-menu
    ref="subMenu"
    :index="routerInfo.name"
    class="lrag-sub-menu dark:text-slate-300 relative"
  >
    <template #title>
      <div
        v-if="!isCollapse"
        class="flex items-center"
        :style="{
          height: sideHeight
        }"
      >
        <el-icon v-if="routerInfo.meta.icon">
          <component :is="routerInfo.meta.icon" />
        </el-icon>
        <span>{{ displayTitle }}</span>
      </div>
      <template v-else>
        <el-icon v-if="routerInfo.meta.icon">
          <component :is="routerInfo.meta.icon" />
        </el-icon>
        <span>{{ displayTitle }}</span>
      </template>
    </template>
    <slot />
  </el-sub-menu>
</template>

<script setup>
  import { inject, computed } from 'vue'
  import { useRoute } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { useAppStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  import { resolveMenuTitle } from '@/utils/menuTitle'

  const appStore = useAppStore()
  const { config } = storeToRefs(appStore)
  const route = useRoute()
  const { locale } = useI18n()

  defineOptions({
    name: 'AsyncSubmenu'
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

<style lang="scss">
  .lrag-sub-menu {
    .el-sub-menu__title {
      height: v-bind('sideHeight') !important;
    }
  }
</style>
