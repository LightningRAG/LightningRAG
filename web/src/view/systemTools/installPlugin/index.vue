<template>
  <div class="lrag-form-box">
    <el-upload
      drag
      :action="`${getBaseUrl()}/autoCode/installPlugin`"
      :show-file-list="false"
      :on-success="handleSuccess"
      :on-error="handleSuccess"
      :headers="{'x-token': token}"
      name="plug"
    >
      <el-icon class="el-icon--upload"><upload-filled /></el-icon>
      <div class="el-upload__text">{{ $t('tools.installPlugin.dragText') }}<em>{{ $t('tools.installPlugin.clickUploadEm') }}</em></div>
      <template #tip>
        <div class="el-upload__tip">{{ $t('tools.installPlugin.zipTip') }}</div>
      </template>
    </el-upload>

    <!-- Plugin List Table -->
    <div style="margin-top: 20px;">
      <el-table :data="pluginList" style="width: 100%">
        <el-table-column type="expand">
            <template #default="props">
                <div style="padding: 20px;">
                    <h3>{{ $t('tools.installPlugin.apiListTitle') }}</h3>
                    <el-table :data="props.row.apis" border>
                        <el-table-column prop="path" :label="$t('common.colPath')" />
                        <el-table-column prop="method" :label="$t('common.colMethod')" />
                        <el-table-column :label="$t('common.colDescription')">
                          <template #default="{ row }">
                            {{ translateApiDescription(row, t) || row.description }}
                          </template>
                        </el-table-column>
                        <el-table-column :label="$t('common.colApiGroup')">
                          <template #default="{ row }">
                            {{ translateApiGroup(row.apiGroup, t) }}
                          </template>
                        </el-table-column>
                    </el-table>
                    <h3>{{ $t('tools.installPlugin.menuListTitle') }}</h3>
                    <el-table :data="props.row.menus" row-key="name" :tree-props="{children: 'children', hasChildren: 'hasChildren'}" border>
                        <el-table-column :label="$t('common.colTitle')">
                          <template #default="cell">
                            {{ resolveMenuTitle(cell.row.meta, route) }}
                          </template>
                        </el-table-column>
                        <el-table-column prop="name" :label="$t('tools.installPlugin.colNameEn')" />
                        <el-table-column prop="path" :label="$t('tools.installPlugin.colPathEn')" />
                    </el-table>
                     <h3>{{ $t('tools.installPlugin.dictListTitle') }}</h3>
                     <el-table :data="props.row.dictionaries" border>
                         <el-table-column prop="name" :label="$t('common.colName')" />
                         <el-table-column prop="type" :label="$t('common.colDictType')" />
                         <el-table-column prop="desc" :label="$t('common.colDescription')" />
                     </el-table>
                </div>
            </template>
        </el-table-column>
        <el-table-column prop="pluginName" :label="$t('common.colPluginName')" />
        <el-table-column prop="pluginType" :label="$t('common.colPluginType')">
          <template #default="scope">
              {{ typeLabel(scope.row.pluginType) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.colActions')">
          <template #default="scope">
            <el-button type="primary" link icon="delete" @click="deletePlugin(scope.row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { resolveMenuTitle } from '@/utils/menuTitle'
  import { ElMessage } from 'element-plus'
  import { getBaseUrl } from '@/utils/format'
  import { useUserStore } from "@/pinia";
  import { getPluginList, removePlugin } from '@/api/autoCode'
  import { ElMessageBox } from 'element-plus'
  import { translateApiDescription, translateApiGroup } from '@/utils/apiI18n'

  const route = useRoute()
  const { t } = useI18n()
  const userStore = useUserStore()
  const token = userStore.token
  const pluginList = ref([])

  const getTableData = async () => {
    const res = await getPluginList()
    if (res.code === 0) {
      pluginList.value = res.data
    }
  }

  const typeLabel = (pluginType) => {
    const map = {
      server: t('tools.installPlugin.typeServer'),
      web: t('tools.installPlugin.typeWeb'),
      full: t('tools.installPlugin.typeFull')
    }
    return map[pluginType] || t('tools.installPlugin.unknownType')
  }

  const deletePlugin = (row) => {
    ElMessageBox.confirm(
    t('tools.installPlugin.confirmDeletePlugin'),
    t('common.tipTitle'),
    {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  )
    .then(async () => {
      const res = await removePlugin({ pluginName: row.pluginName, pluginType: row.pluginType })
      if (res.code === 0) {
        ElMessage.success(t('common.deleteOk'))
        getTableData()
      }
    })
    .catch(() => {
    })
  }

  onMounted(() => {
    getTableData()
  })

  const handleSuccess = (res) => {
    if (res.code === 0) {
      let msg = ``
      res.data &&
        res.data.forEach((item, index) => {
          msg += `${index + 1}.${item.msg}\n`
        })
      ElMessage.success(msg.trim() || t('tools.installPlugin.installOk'))
      getTableData() // Refresh list on success
    } else {
      ElMessage.error(res.msg)
    }
  }
</script>
