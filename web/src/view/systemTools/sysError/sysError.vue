<template>
  <div>
    <div class="lrag-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        class="demo-form-inline"
        @keyup.enter="onSubmit"
      >
        <el-form-item :label="$t('tools.exportTemplate.labelCreatedDate')" prop="createdAtRange">
          <template #label>
            <span>
              {{ $t('tools.exportTemplate.labelCreatedDate') }}
              <el-tooltip
                :content="$t('tools.exportTemplate.dateRangeTip')"
              >
                <el-icon><QuestionFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>

          <el-date-picker
            v-model="searchInfo.createdAtRange"
            class="!w-380px"
            type="datetimerange"
            :range-separator="$t('sysError.rangeTo')"
            :start-placeholder="$t('sysError.startTime')"
            :end-placeholder="$t('sysError.endTime')"
          />
        </el-form-item>

        <el-form-item :label="$t('common.colErrorSource')" prop="form">
          <el-input v-model="searchInfo.form" :placeholder="$t('common.phSearchCondition')" />
        </el-form-item>

        <el-form-item :label="$t('common.colErrorContent')" prop="info">
          <el-input v-model="searchInfo.info" :placeholder="$t('common.phSearchCondition')" />
        </el-form-item>

        <template v-if="showAllQuery">
          <!-- 将需要控制显示状态的查询条件添加到此范围内 -->
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit"
            >{{ $t('common.query') }}</el-button
          >
          <el-button icon="refresh" @click="onReset">{{ $t('common.reset') }}</el-button>
          <el-button
            link
            type="primary"
            icon="arrow-down"
            @click="showAllQuery = true"
            v-if="!showAllQuery"
            >{{ $t('sysError.expand') }}</el-button
          >
          <el-button
            link
            type="primary"
            icon="arrow-up"
            @click="showAllQuery = false"
            v-else
            >{{ $t('sysError.collapse') }}</el-button
          >
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button
          icon="delete"
          style="margin-left: 10px"
          :disabled="!multipleSelection.length"
          @click="onDelete"
          >{{ $t('common.delete') }}</el-button
        >
      </div>
      <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />

        <el-table-column
          sortable
          align="left"
          :label="$t('common.colDate')"
          prop="CreatedAt"
          width="180"
        >
          <template #default="scope">{{
            formatDate(scope.row.CreatedAt)
          }}</template>
        </el-table-column>

        <el-table-column
          align="left"
          :label="$t('common.colErrorSource')"
          prop="form"
          width="120"
        >
          <template #default="scope">
            {{ formatSysErrForm(scope.row.form) }}
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          :label="$t('common.colErrorLevel')"
          prop="level"
          width="120"
        >
          <template #default="scope">
            <el-tag
              effect="dark"
              :type="levelTagMap[scope.row.level] || 'info'"
            >
              {{ levelLabelMap[scope.row.level] || defaultLevelLabel }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          :label="$t('common.colHandleStatus')"
          prop="status"
          width="140"
        >
          <template #default="scope">
            <el-tag
              effect="light"
              :type="statusTagMap[scope.row.status] || 'info'"
            >
              {{ statusLabelMap[scope.row.status] || defaultStatusLabel }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          :label="$t('common.colErrorContent')"
          prop="info"
          show-overflow-tooltip
          width="240"
        >
          <template #default="scope">
            {{ formatSysErrInfo(scope.row.info) }}
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          :label="$t('common.colSolution')"
          show-overflow-tooltip
          prop="solution"
          width="120"
        />

        <el-table-column
          align="left"
          :label="$t('common.colActions')"
          fixed="right"
          :min-width="appStore.operateMinWith"
        >
          <template #default="scope">
            <el-button
              v-if="scope.row.status !== ERR_STATUS.PROCESSING"
              type="primary"
              link
              class="table-button"
              @click="getSolution(scope.row.ID)"
            >
              <el-icon><ai-lrag /></el-icon>{{ $t('sysError.btnAiSolution') }}
            </el-button>
            <el-button
              type="primary"
              link
              class="table-button"
              @click="getDetails(scope.row)"
              ><el-icon style="margin-right: 5px"><InfoFilled /></el-icon
              >{{ $t('common.view') }}</el-button
            >
            <el-button
              type="primary"
              link
              icon="delete"
              @click="deleteRow(scope.row)"
              >{{ $t('common.delete') }}</el-button
            >
          </template>
        </el-table-column>
      </el-table>
      <div class="lrag-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-drawer
      destroy-on-close
      :size="appStore.drawerSize"
      v-model="detailShow"
      :show-close="true"
      :before-close="closeDetailShow"
      :title="$t('sysError.detailDrawerTitle')"
    >
      <el-descriptions :column="2" border direction="vertical">
        <el-descriptions-item :label="$t('common.colErrorSource')">
          {{ formatSysErrForm(detailForm.form) }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.colErrorLevel')">
          <el-tag
            effect="dark"
            :type="levelTagMap[detailForm.level] || 'info'"
          >
            {{ levelLabelMap[detailForm.level] || defaultLevelLabel }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.colHandleStatus')">
          <el-tag
            effect="light"
            :type="statusTagMap[detailForm.status] || 'info'"
          >
            {{ statusLabelMap[detailForm.status] || defaultStatusLabel }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.colErrorContent')" :span="2">
          <pre class="whitespace-pre-wrap break-words">{{ formatSysErrInfo(detailForm.info) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.colSolution')" :span="2">
          <pre class="whitespace-pre-wrap break-words">{{ detailForm.solution }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    deleteSysError,
    deleteSysErrorByIds,
    findSysError,
    getSysErrorList,
    getSysErrorSolution
  } from '@/api/system/sysError'

  import { formatDate } from '@/utils/format'
  import {
    formatSysErrorForm as formatSysErrorFormUtil,
    formatSysErrorInfo as formatSysErrorInfoUtil
  } from '@/utils/sysErrorDisplay'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useAppStore } from '@/pinia'

  defineOptions({
    name: 'SysError'
  })

  const { t } = useI18n()
  const appStore = useAppStore()

  const formatSysErrForm = (form) => formatSysErrorFormUtil(form, t)
  const formatSysErrInfo = (info) => formatSysErrorInfoUtil(info, t)

  /** Backend stores handle status as these Chinese literals */
  const ERR_STATUS = Object.freeze({
    PENDING: '未处理',
    PROCESSING: '处理中',
    DONE: '处理完成',
    FAILED: '处理失败'
  })

  // 控制更多查询条件显示/隐藏状态
  const showAllQuery = ref(false)

  const elSearchFormRef = ref()

  // =========== 表格控制部分 ===========
  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const searchInfo = ref({})
  // 重置
  const onReset = () => {
    searchInfo.value = {}
    getTableData()
  }

  const getSolution = async (id) => {
    const confirmed = await ElMessageBox.confirm(
      t('sysError.aiAnalysisConfirmBody'),
      t('sysError.aiAnalysisConfirmTitle'),
      {
        confirmButtonText: t('common.ok'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    ).catch(() => false)
    if (!confirmed) return
    const res = await getSysErrorSolution({ id })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: res.msg || t('sysError.aiSubmittedFallback')
      })
      getTableData()
    }
  }
  // 搜索
  const onSubmit = () => {
    elSearchFormRef.value?.validate(async (valid) => {
      if (!valid) return
      page.value = 1
      getTableData()
    })
  }

  // 分页
  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  // 修改页面容量
  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  // 查询
  const getTableData = async () => {
    const table = await getSysErrorList({
      page: page.value,
      pageSize: pageSize.value,
      ...searchInfo.value
    })
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
    }
  }

  getTableData()

  // ============== 表格控制部分结束 ===============

  // 获取需要的字典 可能为空 按需保留
  const setOptions = async () => {}

  // 获取需要的字典 可能为空 按需保留
  setOptions()

  // 多选数据
  const multipleSelection = ref([])
  // 多选
  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }

  // 删除行
  const deleteRow = (row) => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(() => {
      deleteSysErrorFunc(row)
    })
  }

  // 多选删除
  const onDelete = async () => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const IDs = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: t('common.pickRowsToDelete')
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map((item) => {
          IDs.push(item.ID)
        })
      const res = await deleteSysErrorByIds({ IDs })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('common.deleteOk')
        })
        if (tableData.value.length === IDs.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
    })
  }

  // 删除行
  const deleteSysErrorFunc = async (row) => {
    const res = await deleteSysError({ ID: row.ID })
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

  const detailForm = ref({})

  // 查看详情控制标记
  const detailShow = ref(false)

  // 打开详情弹窗
  const openDetailShow = () => {
    detailShow.value = true
  }

  // 打开详情
  const getDetails = async (row) => {
    // 打开弹窗
    const res = await findSysError({ ID: row.ID })
    if (res.code === 0) {
      detailForm.value = res.data
      openDetailShow()
    }
  }

  // 关闭详情弹窗
  const closeDetailShow = () => {
    detailShow.value = false
    detailForm.value = {}
  }

  const statusLabelMap = computed(() => ({
    [ERR_STATUS.PENDING]: t('sysError.statusPending'),
    [ERR_STATUS.PROCESSING]: t('sysError.statusProcessing'),
    [ERR_STATUS.DONE]: t('sysError.statusDone'),
    [ERR_STATUS.FAILED]: t('sysError.statusFailed')
  }))
  const statusTagMap = {
    [ERR_STATUS.PENDING]: 'info',
    [ERR_STATUS.PROCESSING]: 'warning',
    [ERR_STATUS.DONE]: 'success',
    [ERR_STATUS.FAILED]: 'danger'
  }
  const defaultStatusLabel = computed(() => t('sysError.statusPending'))

  const levelLabelMap = computed(() => ({
    fatal: t('sysError.levelFatal'),
    error: t('sysError.levelError')
  }))
  const levelTagMap = {
    fatal: 'danger',
    error: 'warning'
  }
  const defaultLevelLabel = computed(() => t('sysError.levelError'))
</script>
