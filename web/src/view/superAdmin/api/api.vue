<template>
  <div>
    <div class="lrag-search-box">
      <el-form ref="searchForm" :inline="true" :model="searchInfo">
        <el-form-item :label="$t('common.colPath')">
          <el-input v-model="searchInfo.path" :placeholder="$t('common.colPath')" />
        </el-form-item>
        <el-form-item :label="$t('common.colDescription')">
          <el-input v-model="searchInfo.description" :placeholder="$t('common.colDescription')" />
        </el-form-item>
        <el-form-item :label="$t('common.colApiGroup')">
          <el-select
            v-model="searchInfo.apiGroup"
            clearable
            :placeholder="$t('common.phSelect')"
          >
            <el-option
              v-for="item in apiGroupOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('common.colHttpMethod')">
          <el-select v-model="searchInfo.method" clearable :placeholder="$t('common.phSelect')">
            <el-option
              v-for="item in methodOptions"
              :key="item.value"
              :label="`${item.label}(${item.value})`"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">
            {{ $t('common.query') }}
          </el-button>
          <el-button icon="refresh" @click="onReset">{{ $t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openDialog('addApi')">
          {{ $t('common.add') }}
        </el-button>
        <el-button icon="delete" :disabled="!apis.length" @click="onDelete">
          {{ $t('common.delete') }}
        </el-button>
        <el-button icon="Refresh" @click="onFresh">{{ $t('adminApi.refreshCache') }}</el-button>
        <el-button icon="Compass" @click="onSync">{{ $t('adminApi.syncApi') }}</el-button>
        <ExportTemplate template-id="api" />
        <ExportExcel template-id="api" :limit="9999" />
        <ImportExcel template-id="api" @on-success="getTableData" />
      </div>
      <el-table
        :data="tableData"
        @sort-change="sortChange"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column
          align="left"
          :label="$t('common.colId')"
          min-width="60"
          prop="ID"
          sortable="custom"
        />
        <el-table-column
          align="left"
          :label="$t('common.colApiPath')"
          min-width="150"
          prop="path"
          sortable="custom"
        />
        <el-table-column
          align="left"
          :label="$t('common.colApiGroup')"
          min-width="150"
          prop="apiGroup"
          sortable="custom"
        >
          <template #default="scope">
            {{ translateApiGroup(scope.row.apiGroup, t) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colApiSummary')"
          min-width="150"
          prop="description"
          sortable="custom"
        >
          <template #default="scope">
            {{ translateApiDescription(scope.row, t) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colHttpMethod')"
          min-width="150"
          prop="method"
          sortable="custom"
        >
          <template #default="scope">
            <div>
              {{ scope.row.method }} / {{ methodFilter(scope.row.method) }}
            </div>
          </template>
        </el-table-column>

        <el-table-column align="left" fixed="right" :label="$t('common.colActions')" :min-width="appStore.operateMinWith">
          <template #default="scope">
            <el-button
              icon="edit"
              type="primary"
              link
              @click="editApiFunc(scope.row)"
            >
              {{ $t('common.edit') }}
            </el-button>
            <el-button
              icon="user"
              type="primary"
              link
              @click="openAssignRoleDrawer(scope.row)"
            >
              {{ $t('admin.menu.assignRole') }}
            </el-button>
            <el-button
              icon="delete"
              type="primary"
              link
              @click="deleteApiFunc(scope.row)"
            >
              {{ $t('common.delete') }}
            </el-button>
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

    <el-drawer
      v-model="syncApiFlag"
      :size="appStore.drawerSize"
      :before-close="closeSyncDialog"
      :show-close="false"
    >
      <warning-bar
        :title="$t('adminApi.syncDrawerWarningBar')"
      />
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('adminApi.syncRoutesTitle') }}</span>
          <div>
            <el-button :loading="apiCompletionLoading" @click="closeSyncDialog">
              {{ $t('common.cancel') }}
            </el-button>
            <el-button
              type="primary"
              :loading="syncing || apiCompletionLoading"
              @click="enterSyncDialog"
            >
              {{ $t('common.ok') }}
            </el-button>
          </div>
        </div>
      </template>

      <h4>
        {{ $t('adminApi.sectionNewRoutes') }}
        <span class="text-xs text-gray-500 mx-2 font-normal"
          >{{ $t('adminApi.hintNewRoutes') }}</span
        >
        <el-button type="primary" size="small" @click="apiCompletion">
          <el-icon size="18">
            <ai-lrag />
          </el-icon>
          {{ $t('adminApi.btnAiAutofill') }}
        </el-button>
      </h4>
      <el-table
        v-loading="syncing || apiCompletionLoading"
        :element-loading-text="$t('adminApi.aiLoadingText')"
        :data="syncApiData.newApis"
      >
        <el-table-column
          align="left"
          :label="$t('common.colApiPath')"
          min-width="150"
          prop="path"
        />
        <el-table-column
          align="left"
          :label="$t('common.colApiGroup')"
          min-width="150"
          prop="apiGroup"
        >
          <template #default="{ row }">
            <el-select
              v-model="row.apiGroup"
              :placeholder="$t('common.phSelectOrCreate')"
              allow-create
              filterable
              default-first-option
            >
              <el-option
                v-for="item in apiGroupOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colApiSummary')"
          min-width="150"
          prop="description"
        >
          <template #default="{ row }">
            <el-input v-model="row.description" autocomplete="off" />
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colHttpMethod')"
          min-width="150"
          prop="method"
        >
          <template #default="scope">
            <div>
              {{ scope.row.method }} / {{ methodFilter(scope.row.method) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.colActions')" min-width="150" fixed="right">
          <template #default="{ row }">
            <el-button icon="plus" type="primary" link @click="addApiFunc(row)">
              {{ $t('adminApi.addSingleRow') }}
            </el-button>
            <el-button
              icon="sunrise"
              type="primary"
              link
              @click="ignoreApiFunc(row, true)"
            >
              {{ $t('adminApi.ignoreRoute') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <h4>
        {{ $t('adminApi.sectionDeletedRoutes') }}
        <span class="text-xs text-gray-500 ml-2 font-normal"
          >{{ $t('adminApi.hintDeletedRoutes') }}</span
        >
      </h4>
      <el-table :data="syncApiData.deleteApis">
        <el-table-column
          align="left"
          :label="$t('common.colApiPath')"
          min-width="150"
          prop="path"
        />
        <el-table-column
          align="left"
          :label="$t('common.colApiGroup')"
          min-width="150"
          prop="apiGroup"
        >
          <template #default="scope">
            {{ translateApiGroup(scope.row.apiGroup, t) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colApiSummary')"
          min-width="150"
          prop="description"
        >
          <template #default="scope">
            {{ translateApiDescription(scope.row, t) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colHttpMethod')"
          min-width="150"
          prop="method"
        >
          <template #default="scope">
            <div>
              {{ scope.row.method }} / {{ methodFilter(scope.row.method) }}
            </div>
          </template>
        </el-table-column>
      </el-table>

      <h4>
        {{ $t('adminApi.sectionIgnoredRoutes') }}
        <span class="text-xs text-gray-500 ml-2 font-normal"
          >{{ $t('adminApi.hintIgnoredRoutes') }}</span
        >
      </h4>
      <el-table :data="syncApiData.ignoreApis">
        <el-table-column
          align="left"
          :label="$t('common.colApiPath')"
          min-width="150"
          prop="path"
        />
        <el-table-column
          align="left"
          :label="$t('common.colApiGroup')"
          min-width="150"
          prop="apiGroup"
        >
          <template #default="scope">
            {{ translateApiGroup(scope.row.apiGroup, t) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colApiSummary')"
          min-width="150"
          prop="description"
        >
          <template #default="scope">
            {{ translateApiDescription(scope.row, t) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colHttpMethod')"
          min-width="150"
          prop="method"
        >
          <template #default="scope">
            <div>
              {{ scope.row.method }} / {{ methodFilter(scope.row.method) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.colActions')" min-width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              icon="sunny"
              type="primary"
              link
              @click="ignoreApiFunc(row, false)"
            >
              {{ $t('adminApi.unignoreRoute') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>

    <el-drawer
      v-model="dialogFormVisible"
      :size="appStore.drawerSize"
      :before-close="closeDialog"
      :show-close="false"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ dialogTitle }}</span>
          <div>
            <el-button @click="closeDialog"> {{ $t('common.cancel') }} </el-button>
            <el-button type="primary" @click="enterDialog"> {{ $t('common.ok') }} </el-button>
          </div>
        </div>
      </template>

      <warning-bar :title="$t('adminApi.warningNewApiNeedRole')" />
      <el-form ref="apiForm" :model="form" :rules="rules" label-width="80px">
        <el-form-item :label="$t('common.colPath')" prop="path">
          <el-input v-model="form.path" autocomplete="off" />
        </el-form-item>
        <el-form-item :label="$t('common.colHttpMethod')" prop="method">
          <el-select
            v-model="form.method"
            :placeholder="$t('common.phSelect')"
            style="width: 100%"
          >
            <el-option
              v-for="item in methodOptions"
              :key="item.value"
              :label="`${item.label}(${item.value})`"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('common.colApiGroup')" prop="apiGroup">
          <el-select
            v-model="form.apiGroup"
            :placeholder="$t('common.phSelectOrCreate')"
            allow-create
            filterable
            default-first-option
          >
            <el-option
              v-for="item in apiGroupOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('common.colApiSummary')" prop="description">
          <el-input v-model="form.description" autocomplete="off" />
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 分配给角色抽屉 -->
    <el-drawer
      v-model="assignRoleDrawerVisible"
      :size="appStore.drawerSize"
      :show-close="false"
      destroy-on-close
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{
            $t('adminApi.assignRoleTitle', { desc: assignApiRow.description || '' })
          }}</span>
          <div>
            <el-button @click="assignRoleDrawerVisible = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" :loading="assignRoleSubmitting" @click="confirmAssignRole">{{ $t('common.ok') }}</el-button>
          </div>
        </div>
      </template>
      <warning-bar :title="$t('adminApi.assignRoleWarningBar')" />
      <el-tree
        ref="roleTreeRef"
        v-loading="assignRoleLoading"
        :data="authorityTreeData"
        :props="{ label: 'authorityName', children: 'children' }"
        node-key="authorityId"
        show-checkbox
        check-strictly
        default-expand-all
      >
        <template #default="{ data }">
          <span>{{ authorityDisplayName(data, t) }}</span>
        </template>
      </el-tree>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getApiById,
    getApiList,
    createApi,
    updateApi,
    deleteApi,
    deleteApisByIds,
    freshCasbin,
    syncApi,
    getApiGroups,
    ignoreApi,
    enterSyncApi,
    getApiRoles,
    setApiRoles
  } from '@/api/api'
  import { getAuthorityList } from '@/api/authority'
  import { toSQLLine } from '@/utils/stringFun'
  import { authorityDisplayName } from '@/utils/authorityI18n'
  import { translateApiDescription, translateApiGroup } from '@/utils/apiI18n'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, nextTick, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import ExportExcel from '@/components/exportExcel/exportExcel.vue'
  import ExportTemplate from '@/components/exportExcel/exportTemplate.vue'
  import ImportExcel from '@/components/exportExcel/importExcel.vue'
  import { llmAuto } from '@/api/autoCode'
  import { useAppStore } from "@/pinia";

  defineOptions({
    name: 'Api'
  })

  const { t } = useI18n()
  const appStore = useAppStore()

  const methodFilter = (value) => {
    const target = methodOptions.value.find((item) => item.value === value)
    return target ? target.label : ''
  }

  const apis = ref([])
  const form = ref({
    path: '',
    apiGroup: '',
    method: '',
    description: ''
  })
  const methodOptions = computed(() => [
    {
      value: 'POST',
      label: t('adminApi.methodCreate'),
      type: 'success'
    },
    {
      value: 'GET',
      label: t('adminApi.methodRead'),
      type: ''
    },
    {
      value: 'PUT',
      label: t('adminApi.methodUpdate'),
      type: 'warning'
    },
    {
      value: 'DELETE',
      label: t('adminApi.methodDelete'),
      type: 'danger'
    }
  ])

  const type = ref('')
  const rules = computed(() => ({
    path: [{ required: true, message: t('adminApi.ruleApiPath'), trigger: 'blur' }],
    apiGroup: [{ required: true, message: t('adminApi.ruleApiGroupName'), trigger: 'blur' }],
    method: [{ required: true, message: t('adminApi.ruleHttpMethod'), trigger: 'blur' }],
    description: [{ required: true, message: t('adminApi.ruleApiDesc'), trigger: 'blur' }]
  }))

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const searchInfo = ref({})
  const apiGroupOptions = ref([])
  const apiGroupMap = ref({})

  const getGroup = async () => {
    const res = await getApiGroups()
    if (res.code === 0) {
      const groups = res.data.groups
      apiGroupOptions.value = groups.map((item) => ({
        label: translateApiGroup(item, t),
        value: item
      }))
      apiGroupMap.value = res.data.apiGroupMap
    }
  }

  const ignoreApiFunc = async (row, flag) => {
    const res = await ignoreApi({ path: row.path, method: row.method, flag })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: res.msg
      })
      if (flag) {
        syncApiData.value.newApis = syncApiData.value.newApis.filter(
          (item) => !(item.path === row.path && item.method === row.method)
        )
        syncApiData.value.ignoreApis.push(row)
        return
      }
      syncApiData.value.ignoreApis = syncApiData.value.ignoreApis.filter(
        (item) => !(item.path === row.path && item.method === row.method)
      )
      syncApiData.value.newApis.push(row)
    }
  }

  const addApiFunc = async (row) => {
    if (!row.apiGroup) {
      ElMessage({
        type: 'error',
        message: t('adminApi.errPickApiGroup')
      })
      return
    }
    if (!row.description) {
      ElMessage({
        type: 'error',
        message: t('adminApi.errFillApiDesc')
      })
      return
    }
    const res = await createApi(row)
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: t('adminApi.addOkNeedRoleAssign'),
        showClose: true
      })
      syncApiData.value.newApis = syncApiData.value.newApis.filter(
        (item) => !(item.path === row.path && item.method === row.method)
      )
    }
    getTableData()
    getGroup()
  }

  const closeSyncDialog = () => {
    syncApiFlag.value = false
  }

  const syncing = ref(false)

  const enterSyncDialog = async () => {
    if (
      syncApiData.value.newApis.some(
        (item) => !item.apiGroup || !item.description
      )
    ) {
      ElMessage({
        type: 'error',
        message: t('adminApi.errSyncIncomplete')
      })
      return
    }

    syncing.value = true
    const res = await enterSyncApi(syncApiData.value)
    syncing.value = false
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: res.msg
      })
      syncApiFlag.value = false
      getTableData()
    }
  }

  const onReset = () => {
    searchInfo.value = {}
    getTableData()
  }
  // 搜索

  const onSubmit = () => {
    page.value = 1
    getTableData()
  }

  // 分页
  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  // 排序
  const sortChange = ({ prop, order }) => {
    if (prop) {
      if (prop === 'ID') {
        prop = 'id'
      }
      searchInfo.value.orderKey = toSQLLine(prop)
      searchInfo.value.desc = order === 'descending'
    }
    getTableData()
  }

  // 查询
  const getTableData = async () => {
    const table = await getApiList({
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
  getGroup()
  // 批量操作
  const handleSelectionChange = (val) => {
    apis.value = val
  }

  const onDelete = async () => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const ids = apis.value.map((item) => item.ID)
      const res = await deleteApisByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: res.msg
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
    })
  }
  const onFresh = async () => {
    ElMessageBox.confirm(t('adminApi.confirmRefreshCache'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await freshCasbin()
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: res.msg
        })
      }
    })
  }

  const syncApiData = ref({
    newApis: [],
    deleteApis: [],
    ignoreApis: []
  })

  const syncApiFlag = ref(false)

  const onSync = async () => {
    const res = await syncApi()
    if (res.code === 0) {
      res.data.newApis.forEach((item) => {
        item.apiGroup = apiGroupMap.value[item.path.split('/')[1]]
      })

      syncApiData.value = res.data
      syncApiFlag.value = true
    }
  }

  // 弹窗相关
  const apiForm = ref(null)
  const initForm = () => {
    apiForm.value.resetFields()
    form.value = {
      path: '',
      apiGroup: '',
      method: '',
      description: ''
    }
  }

  const dialogTitle = ref('')
  const dialogFormVisible = ref(false)
  const openDialog = (key) => {
    switch (key) {
      case 'addApi':
        dialogTitle.value = t('adminApi.drawerAddApi')
        break
      case 'edit':
        dialogTitle.value = t('adminApi.drawerEditApi')
        break
      default:
        break
    }
    type.value = key
    dialogFormVisible.value = true
  }
  const closeDialog = () => {
    initForm()
    dialogFormVisible.value = false
  }

  const editApiFunc = async (row) => {
    const res = await getApiById({ id: row.ID })
    form.value = res.data.api
    openDialog('edit')
  }

  const enterDialog = async () => {
    apiForm.value.validate(async (valid) => {
      if (valid) {
        switch (type.value) {
          case 'addApi':
            {
              const res = await createApi(form.value)
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: t('adminApi.addOkShort'),
                  showClose: true
                })
              }
              getTableData()
              getGroup()
              closeDialog()
            }

            break
          case 'edit':
            {
              const res = await updateApi(form.value)
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: t('adminApi.editOkShort'),
                  showClose: true
                })
              }
              getTableData()
              closeDialog()
            }
            break
          default:
            {
              ElMessage({
                type: 'error',
                message: t('adminApi.unknownAction'),
                showClose: true
              })
            }
            break
        }
      }
    })
  }

  const deleteApiFunc = async (row) => {
    ElMessageBox.confirm(t('adminApi.confirmDeleteRoleApi'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteApi(row)
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('common.deleteOk')
        })
        if (tableData.value.length === 1 && page.value > 1) {
          page.value--
        }
        getTableData()
        getGroup()
      }
    })
  }
  const apiCompletionLoading = ref(false)
  const apiCompletion = async () => {
    apiCompletionLoading.value = true
    const routerPaths = syncApiData.value.newApis
      .filter((item) => !item.apiGroup || !item.description)
      .map((item) => item.path)
    const res = await llmAuto({ data: String(routerPaths), mode: 'apiCompletion' })
    apiCompletionLoading.value = false
    if (res.code === 0) {
      try {
        const data = JSON.parse(res.data)
        syncApiData.value.newApis.forEach((item) => {
          const target = data.find((d) => d.path === item.path)
          if (target) {
            if (!item.apiGroup) {
              item.apiGroup = target.apiGroup
            }
            if (!item.description) {
              item.description = target.description
            }
          }
        })
      } catch (_) {
        ElMessage({
          type: 'error',
          message: t('adminApi.aiAutofillFail')
        })
      }
    }
  }

  // 分配给角色
  const assignRoleDrawerVisible = ref(false)
  const assignApiRow = ref({})
  const authorityTreeData = ref([])
  const assignRoleLoading = ref(false)
  const assignRoleSubmitting = ref(false)
  const roleTreeRef = ref(null)

  const openAssignRoleDrawer = async (row) => {
    assignApiRow.value = row
    assignRoleDrawerVisible.value = true
    assignRoleLoading.value = true
    const [authRes, rolesRes] = await Promise.all([
      getAuthorityList(),
      getApiRoles(row.path, row.method)
    ])
    if (authRes.code === 0) {
      authorityTreeData.value = authRes.data
    }
    if (rolesRes.code === 0 && rolesRes.data) {
      nextTick(() => {
        roleTreeRef.value?.setCheckedKeys(rolesRes.data)
      })
    }
    assignRoleLoading.value = false
  }

  const confirmAssignRole = async () => {
    assignRoleSubmitting.value = true
    try {
      const checkedKeys = roleTreeRef.value?.getCheckedKeys(false) || []
      const res = await setApiRoles({
        path: assignApiRow.value.path,
        method: assignApiRow.value.method,
        authorityIds: checkedKeys
      })
      if (res.code === 0) {
        ElMessage({ type: 'success', message: t('adminApi.assignOk') })
        assignRoleDrawerVisible.value = false
      }
    } catch {
      ElMessage({ type: 'error', message: t('adminApi.assignFailRetry') })
    }
    assignRoleSubmitting.value = false
  }
</script>

<style scoped lang="scss">
  .warning {
    color: #dc143c;
  }
</style>
