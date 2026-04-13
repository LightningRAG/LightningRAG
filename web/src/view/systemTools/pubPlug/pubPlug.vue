<template>
  <div class="lrag-form-box">
    <div class="p-4 bg-white dark:bg-slate-900">
      <WarningBar :title="$t('tools.pubPlug.warnStandardOnly')" />
      <div class="flex items-center gap-3">
        <el-input
          v-model="plugName"
          :placeholder="$t('tools.pubPlug.phPlugName')"
        />
      </div>
      <el-card class="mt-2 text-center">
        <WarningBar :title="$t('tools.pubPlug.warnTransferChildren')" />
        <el-input
          v-model="parentMenu"
          :placeholder="$t('tools.pubPlug.phParentMenu')"
          class="mb-2"
        />
        <el-transfer
          v-model="menus"
          :props="{
            key: 'ID'
          }"
          class="plugin-transfer"
          :data="menusData"
          filterable
          :filter-method="filterMenuMethod"
          :filter-placeholder="$t('tools.pubPlug.transferFilterPhMenu')"
          :titles="[ $t('tools.pubPlug.transferAvailMenus'), $t('tools.pubPlug.transferSelMenus') ]"
          :button-texts="[ $t('tools.pubPlug.transferBtnRemove'), $t('tools.pubPlug.transferBtnAdd') ]"
        >
          <template #default="{ option }">
            {{ resolveMenuTitle(option.meta, route) }} {{ option.component }}
          </template>
        </el-transfer>
        <div class="flex justify-end mt-2">
          <el-button type="primary" @click="fmtInitMenu">
            {{ $t('tools.pubPlug.btnDefineMenu') }}
          </el-button>
        </div>
      </el-card>
      <el-card class="mt-2 text-center">
        <el-transfer
          v-model="apis"
          :props="{
            key: 'ID'
          }"
          class="plugin-transfer"
          :data="apisData"
          filterable
          :filter-method="filterApiMethod"
          :filter-placeholder="$t('tools.pubPlug.transferFilterPhApi')"
          :titles="[ $t('tools.pubPlug.transferAvailApis'), $t('tools.pubPlug.transferSelApis') ]"
          :button-texts="[ $t('tools.pubPlug.transferBtnRemove'), $t('tools.pubPlug.transferBtnAdd') ]"
        >
          <template #default="{ option }">
            {{ option._displayDesc || option.description }} {{ option.path }}
          </template>
        </el-transfer>
        <div class="flex justify-end mt-2">
          <el-button type="primary" @click="fmtInitAPI">
            {{ $t('tools.pubPlug.btnDefineApi') }}
          </el-button>
        </div>
      </el-card>
      <el-card class="mt-2 text-center">
        <el-transfer
          v-model="dictionaries"
          :props="{
            key: 'ID'
          }"
          class="plugin-transfer"
          :data="dictionariesData"
          filterable
          :filter-method="filterDictionaryMethod"
          :filter-placeholder="$t('tools.pubPlug.transferFilterPhDict')"
          :titles="[ $t('tools.pubPlug.transferAvailDicts'), $t('tools.pubPlug.transferSelDicts') ]"
          :button-texts="[ $t('tools.pubPlug.transferBtnRemove'), $t('tools.pubPlug.transferBtnAdd') ]"
        >
          <template #default="{ option }">
            {{ option.name }} {{ option.type }}
          </template>
        </el-transfer>
        <div class="flex justify-end mt-2">
          <el-button type="primary" @click="fmtInitDictionary">
            {{ $t('tools.pubPlug.btnDefineDictionary') }}
          </el-button>
        </div>
      </el-card>
    </div>
    <div class="flex justify-end">
      <el-button type="primary" @click="pubPlugin">
        {{ $t('tools.pubPlug.btnPackagePlugin') }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { useRoute } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { resolveMenuTitle } from '@/utils/menuTitle'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { pubPlug, initMenu, initAPI, initDictionary } from '@/api/autoCode.js'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getAllApis } from '@/api/api'
  import { translateApiDescription } from '@/utils/apiI18n'
  import { getMenuList } from '@/api/menu'
  import { getSysDictionaryList } from '@/api/sysDictionary'

  const route = useRoute()
  const { t } = useI18n()

  const plugName = ref('')

  const menus = ref([])
  const menusData = ref([])
  const apis = ref([])
  const apisData = ref([])
  const dictionaries = ref([])
  const dictionariesData = ref([])
  const parentMenu = ref('')

  const fmtMenu = (menus) => {
    const res = []
    menus.forEach((item) => {
      if (item.children) {
        res.push(...fmtMenu(item.children))
      } else {
        res.push(item)
      }
    })
    return res
  }

  const initData = async () => {
    const menuRes = await getMenuList()
    if (menuRes.code === 0) {
      menusData.value = fmtMenu(menuRes.data)
    }
    const apiRes = await getAllApis()
    if (apiRes.code === 0) {
      apisData.value = apiRes.data.apis.map(api => ({
        ...api,
        _displayDesc: translateApiDescription(api, t) || api.description
      }))
    }
    const dictionaryRes = await getSysDictionaryList({
      page: 1,
      pageSize: 9999
    })
    if (dictionaryRes.code === 0) {
      dictionariesData.value = dictionaryRes.data
    }
  }

  const filterMenuMethod = (query, item) => {
    const label = resolveMenuTitle(item.meta, route)
    const raw = String(item.meta?.title ?? '')
    return (
      label.indexOf(query) > -1 ||
      raw.indexOf(query) > -1 ||
      item.component.indexOf(query) > -1
    )
  }

  const filterApiMethod = (query, item) => {
    return (item._displayDesc || item.description).indexOf(query) > -1 || item.path.indexOf(query) > -1
  }

  const filterDictionaryMethod = (query, item) => {
    return item.name.indexOf(query) > -1 || item.type.indexOf(query) > -1
  }

  initData()

  const pubPlugin = async () => {
    ElMessageBox.confirm(
      t('tools.pubPlug.confirmPackageBody', { plugName: plugName.value || '…' }),
      t('tools.pubPlug.confirmPackageTitle'),
      {
        confirmButtonText: t('tools.pubPlug.confirmPackageOk'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
      .then(async () => {
        const res = await pubPlug({ plugName: plugName.value })
        if (res.code === 0) {
          ElMessage.success(res.msg)
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: t('tools.pubPlug.msgPackageCancelled')
        })
      })
  }

  const fmtInitMenu = () => {
    if (!parentMenu.value) {
      ElMessage.error(t('tools.pubPlug.errParentMenuRequired'))
      return
    }
    if (menus.value.length === 0) {
      ElMessage.error(t('tools.pubPlug.errPickMenus'))
      return
    }
    if (plugName.value === '') {
      ElMessage.error(t('tools.pubPlug.errPlugNameRequired'))
      return
    }
    ElMessageBox.confirm(
      t('tools.pubPlug.confirmMenuBody', { plugName: plugName.value }),
      t('tools.pubPlug.confirmMenuTitle'),
      {
        confirmButtonText: t('tools.pubPlug.confirmGenerateOk'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
      .then(async () => {
        const req = {
          plugName: plugName.value,
          parentMenu: parentMenu.value,
          menus: menus.value
        }
        const res = await initMenu(req)
        if (res.code === 0) {
          ElMessage.success(t('tools.pubPlug.successMenuInjected'))
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: t('tools.pubPlug.cancelMenuGen')
        })
      })
  }
  const fmtInitAPI = () => {
    if (apis.value.length === 0) {
      ElMessage.error(t('tools.pubPlug.errPickApis'))
      return
    }
    if (plugName.value === '') {
      ElMessage.error(t('tools.pubPlug.errPlugNameRequired'))
      return
    }
    ElMessageBox.confirm(
      t('tools.pubPlug.confirmApiBody', { plugName: plugName.value }),
      t('tools.pubPlug.confirmApiTitle'),
      {
        confirmButtonText: t('tools.pubPlug.confirmGenerateOk'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
      .then(async () => {
        const req = {
          plugName: plugName.value,
          apis: apis.value
        }
        const res = await initAPI(req)
        if (res.code === 0) {
          ElMessage.success(t('tools.pubPlug.successApiInjected'))
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: t('tools.pubPlug.cancelApiGen')
        })
      })
  }

  const fmtInitDictionary = () => {
    if (dictionaries.value.length === 0) {
      ElMessage.error(t('tools.pubPlug.errPickDicts'))
      return
    }
    if (plugName.value === '') {
      ElMessage.error(t('tools.pubPlug.errPlugNameRequired'))
      return
    }
    ElMessageBox.confirm(
      t('tools.pubPlug.confirmDictBody', { plugName: plugName.value }),
      t('tools.pubPlug.confirmDictTitle'),
      {
        confirmButtonText: t('tools.pubPlug.confirmGenerateOk'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
      .then(async () => {
        const req = {
          plugName: plugName.value,
          dictionaries: dictionaries.value
        }
        const res = await initDictionary(req)
        if (res.code === 0) {
          ElMessage.success(t('tools.pubPlug.successDictInjected'))
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: t('tools.pubPlug.cancelDictGen')
        })
      })
  }
</script>

<style lang="scss">
  .plugin-transfer {
    .el-transfer-panel {
      width: 400px !important;
    }
  }
</style>
