<template>
  <div>
    <div class="lrag-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item :label="$t('common.colUsername')">
          <el-input v-model="searchInfo.username" :placeholder="$t('tools.loginLog.phSearchUsername')" />
        </el-form-item>
        <el-form-item :label="$t('common.colStatus')">
             <el-select v-model="searchInfo.status" :placeholder="$t('common.phSelect')" clearable>
                 <el-option :label="$t('common.loginStatusSuccess')" :value="true" />
                 <el-option :label="$t('common.loginStatusFail')" :value="false" />
             </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">{{ $t('common.query') }}</el-button>
          <el-button icon="refresh" @click="onReset">{{ $t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button
          icon="delete"
          style="margin-left: 10px;"
          :disabled="!multipleSelection.length"
          @click="onDelete"
        >{{ $t('common.delete') }}</el-button>
      </div>
      <el-table
        ref="multipleTable"
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column align="left" :label="$t('common.colId')" prop="ID" width="80" />
        <el-table-column align="left" :label="$t('common.colUsername')" prop="username" width="150" />
        <el-table-column align="left" :label="$t('common.colLoginIp')" prop="ip" width="150" />
        <el-table-column align="left" :label="$t('common.colStatus')" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status ? 'success' : 'danger'">
              {{ scope.row.status ? $t('common.loginStatusSuccess') : $t('common.loginStatusFail') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.detail')" show-overflow-tooltip>
             <template #default="scope">
                 {{ formatLoginLogDetailCell(scope.row.status, scope.row.errorMessage) }}
             </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colBrowserAgent')" prop="agent" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('common.colLoginTime')" width="180">
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colActions')" width="120">
          <template #default="scope">
            <el-popover v-model:visible="scope.row.visible" placement="top" width="160">
              <p>{{ $t('tools.loginLog.confirmDelete') }}</p>
              <div style="text-align: right; margin: 0">
                <el-button size="small" type="primary" link @click="scope.row.visible = false">{{ $t('settings.general.cancel') }}</el-button>
                <el-button size="small" type="primary" @click="deleteRow(scope.row)">{{ $t('settings.general.confirm') }}</el-button>
              </div>
              <template #reference>
                <el-button icon="delete" type="primary" link @click="scope.row.visible = true">{{ $t('common.delete') }}</el-button>
              </template>
            </el-popover>
          </template>
        </el-table-column>
      </el-table>
      <div class="lrag-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  getLoginLogList,
  deleteLoginLog,
  deleteLoginLogByIds
} from '@/api/sysLoginLog'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDate } from '@/utils/format'
import { formatLoginLogDetailCell as loginLogDetailCell } from '@/utils/loginLogDetail'

const { t } = useI18n()

const formatLoginLogDetailCell = (statusOk, errorMessage) =>
  loginLogDetailCell(statusOk, errorMessage, t)
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})
const multipleSelection = ref([])

const handleSelectionChange = (val) => {
  multipleSelection.value = val
}

const getTableData = async () => {
  const table = await getLoginLogList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

const deleteRow = async (row) => {
  row.visible = false
  const res = await deleteLoginLog(row)
  if (res.code === 0) {
    ElMessage({
      type: 'success',
      message: t('common.deleteOk')
    })
    if (tableData.value.length === 1 && page.value > 1) {
      page.value--
    }
    getTableData()
  }
}

const onDelete = async() => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
    }).then(async() => {
        const ids = multipleSelection.value.map(item => item.ID)
        const res = await deleteLoginLogByIds({ ids })
        if (res.code === 0) {
            ElMessage({
                type: 'success',
                message: t('common.deleteOk')
            })
            if (tableData.value.length === ids.length && page.value > 1) {
                page.value--
            }
            getTableData()
        }
    })
}

const onSubmit = () => {
  page.value = 1
  pageSize.value = 10
  getTableData()
}

const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

// 首次加载
getTableData()
</script>

<style scoped>
</style>
