<template>
  <el-dialog
    v-model="dialogVisible"
    width="30%"
    class="overlay"
    :show-close="false"
  >
    <template #header>
      <input
        v-model="searchInput"
        class="quick-input"
        :placeholder="t('layout.commandMenu.phSearch')"
      />
    </template>

    <div v-for="(option, index) in options" :key="index">
      <div v-if="option.children.length" class="quick-title">
        {{ option.label }}
      </div>
      <div
        v-for="(item, key) in option.children"
        :key="index + '-' + key"
        class="quick-item"
        @click="item.func"
      >
        {{ item.label }}
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="close">{{ t('layout.commandMenu.close') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
  import { reactive, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { useRouterStore } from '@/pinia/modules/router'
  import { useAppStore, useUserStore } from '@/pinia'
  import { resolveMenuTitle } from '@/utils/menuTitle'
  import { useI18n } from 'vue-i18n'

  defineOptions({
    name: 'CommandMenu'
  })

  const { t } = useI18n()
  const appStore = useAppStore()
  const userStore = useUserStore()

  const router = useRouter()
  const route = useRoute()
  const routerStore = useRouterStore()
  const dialogVisible = ref(false)
  const searchInput = ref('')
  const options = reactive([])
  const deepMenus = (menus) => {
    const arr = []
    menus?.forEach((menu) => {
      if (menu.children && menu.children.length > 0) {
        arr.push(...deepMenus(menu.children))
      } else {
        const label = resolveMenuTitle(menu.meta, route)
        const q = searchInput.value
        const hit =
          (label && label.indexOf(q) > -1) ||
          (menu.name && String(menu.name).indexOf(q) > -1) ||
          (menu.meta?.title && String(menu.meta.title).indexOf(q) > -1)
        if (hit) {
          arr.push({
            label,
            func: () => changeRouter(menu)
          })
        }
      }
    })
    return arr
  }

  const addQuickMenu = () => {
    const option = {
      label: t('layout.commandMenu.sectionJump'),
      children: []
    }
    const menus = deepMenus(routerStore.asyncRouters[0]?.children || [])
    option.children.push(...menus)
    options.push(option)
  }

  const addQuickOption = () => {
    const option = {
      label: t('layout.commandMenu.sectionActions'),
      children: []
    }
    const quickArr = [
      {
        label: t('layout.commandMenu.themeLight'),
        func: () => changeMode(false)
      },
      {
        label: t('layout.commandMenu.themeDark'),
        func: () => changeMode(true)
      },
      {
        label: t('layout.header.logout'),
        func: () => userStore.LoginOut()
      }
    ]
    option.children.push(
      ...quickArr.filter((item) => item.label.indexOf(searchInput.value) > -1)
    )
    options.push(option)
  }

  addQuickMenu()
  addQuickOption()

  const open = () => {
    dialogVisible.value = true
  }

  const changeRouter = (e) => {
    const index = e.name
    const query = {}
    const params = {}
    routerStore.routeMap[index]?.parameters &&
      routerStore.routeMap[index]?.parameters.forEach((item) => {
        if (item.type === 'query') {
          query[item.key] = item.value
        } else {
          params[item.key] = item.value
        }
      })
    if (index === route.name) return
    if (e.name.indexOf('http://') > -1 || e.name.indexOf('https://') > -1) {
      window.open(e.name)
    } else {
      router.push({ name: index, query, params })
    }
    dialogVisible.value = false
  }

  const changeMode = (darkMode) => {
    appStore.toggleTheme(darkMode)
  }

  const close = () => {
    dialogVisible.value = false
  }

  defineExpose({ open })

  watch(searchInput, () => {
    options.length = 0
    addQuickMenu()
    addQuickOption()
  })
</script>

<style lang="scss">
  .overlay {
    border-radius: 4px;
    .el-dialog__header {
      padding: 0 !important;
      margin-right: 0 !important;
    }
    .el-dialog__body {
      padding: 12px !important;
      height: 50vh;
      overflow: auto !important;
    }
    .quick-title {
      margin-top: 8px;
      font-size: 12px;
      font-weight: 600;
      color: #666;
    }
    .quick-input {
      @apply bg-gray-50 dark:bg-gray-800;
      color: #666;
      border-radius: 4px 4px 0 0;
      border: none;
      padding: 12px 16px;
      box-sizing: border-box;
      width: 100%;
      font-size: 16px;
      border-bottom: 1px solid #ddd;
    }
    .quick-item {
      font-size: 14px;
      padding: 8px;
      margin: 4px 0;
      &:hover {
        @apply bg-gray-200 dark:bg-slate-500;
        cursor: pointer;
        border-radius: 4px;
      }
    }
  }
</style>
