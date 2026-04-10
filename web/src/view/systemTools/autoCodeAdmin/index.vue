<template>
  <div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="goAutoCode(null)">
          {{ $t('tools.autoCodeAdmin.btnAdd') }}
        </el-button>
      </div>
      <el-table :data="tableData">
        <el-table-column type="selection" width="55" />
        <el-table-column align="left" label="id" width="60" prop="ID" />
        <el-table-column align="left" :label="$t('common.colDate')" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('tools.autoCodeAdmin.colStructName')"
          min-width="150"
          prop="structName"
        />
        <el-table-column
          align="left"
          :label="$t('tools.autoCodeAdmin.colStructDesc')"
          min-width="150"
          prop="description"
        />
        <el-table-column
          align="left"
          :label="$t('tools.autoCodeAdmin.colTableName')"
          min-width="150"
          prop="tableName"
        />
        <el-table-column
          align="left"
          :label="$t('tools.autoCodeAdmin.colRollback')"
          min-width="150"
          prop="flag"
        >
          <template #default="scope">
            <el-tag v-if="scope.row.flag" type="danger" effect="dark">
              {{ $t('tools.autoCodeAdmin.tagRolledBack') }}
            </el-tag>
            <el-tag v-else type="success" effect="dark">{{ $t('tools.autoCodeAdmin.tagNotRolledBack') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colActions')" min-width="240">
          <template #default="scope">
            <div>
              <el-button
                type="primary"
                link
                :disabled="scope.row.flag === 1"
                @click="addFuncBtn(scope.row)"
              >
                {{ $t('tools.autoCodeAdmin.btnAddMethod') }}
              </el-button>
              <el-button type="primary" link @click="goAutoCode(scope.row, 1)">
                {{ $t('tools.autoCodeAdmin.btnAddField') }}
              </el-button>
              <el-button
                type="primary"
                link
                :disabled="scope.row.flag === 1"
                @click="openDialog(scope.row)"
              >
                {{ $t('tools.autoCodeAdmin.btnRollback') }}
              </el-button>
              <el-button type="primary" link @click="goAutoCode(scope.row)">
                {{ $t('tools.autoCodeAdmin.btnReuse') }}
              </el-button>
              <el-button type="primary" link @click="deleteRow(scope.row)">
                {{ $t('common.delete') }}
              </el-button>
            </div>
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
    <el-dialog
      v-model="dialogFormVisible"
      :title="dialogFormTitle"
      :before-close="closeDialog"
      width="600px"
    >
      <el-form :inline="true" :model="formData" label-width="80px">
        <el-form-item :label="$t('tools.autoCodeAdmin.labelOptions')">
          <el-checkbox v-model="formData.deleteApi" :label="$t('tools.autoCodeAdmin.chkDeleteApi')" />
          <el-checkbox v-model="formData.deleteMenu" :label="$t('tools.autoCodeAdmin.chkDeleteMenu')" />
          <el-checkbox
            v-model="formData.deleteTable"
            :label="$t('tools.autoCodeAdmin.chkDeleteTable')"
            @change="deleteTableCheck"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeDialog">{{ $t('tools.autoCodeAdmin.btnCancel') }}</el-button>
          <el-popconfirm
            :title="$t('tools.autoCodeAdmin.confirmRollback')"
            @confirm="enterDialog"
          >
            <template #reference>
              <el-button type="primary">{{ $t('tools.autoCodeAdmin.btnOk') }}</el-button>
            </template>
          </el-popconfirm>
        </div>
      </template>
    </el-dialog>

    <el-drawer
      v-model="funcFlag"
      size="60%"
      :show-close="false"
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('tools.autoCodeAdmin.drawerToolbar') }}</span>
          <div>
            <el-button type="primary" @click="runFunc" :loading="aiLoading">
              {{ $t('tools.autoCodeAdmin.btnGenerate') }}
            </el-button>
            <el-button type="primary" @click="closeFunc" :loading="aiLoading">
              {{ $t('common.cancel') }}
            </el-button>
          </div>
        </div>
      </template>
      <div class="">
        <el-form
          v-loading="aiLoading"
          label-position="top"
          :element-loading-text="$t('tools.autoCodeAdmin.loadingAi')"
          :model="autoFunc"
          label-width="80px"
        >
          <el-row :gutter="12">
            <el-col :span="8">
              <el-form-item :label="$t('tools.autoCodeAdmin.labelPackage')">
                <el-input
                    v-model="autoFunc.package"
                    :placeholder="$t('tools.autoCodeAdmin.phPackage')"
                    disabled
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('tools.autoCodeAdmin.labelStruct')">
                <el-input
                    v-model="autoFunc.structName"
                    :placeholder="$t('tools.autoCodeAdmin.phStruct')"
                    disabled
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('tools.autoCodeAdmin.labelWebFile')">
                <el-input
                    v-model="autoFunc.packageName"
                    :placeholder="$t('tools.autoCodeAdmin.phFileName')"
                    disabled
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="12">
            <el-col :span="8">
              <el-form-item :label="$t('tools.autoCodeAdmin.labelServerFile')">
                <el-input
                    v-model="autoFunc.humpPackageName"
                    :placeholder="$t('tools.autoCodeAdmin.phFileName')"
                    disabled
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('tools.autoCodeAdmin.labelDesc')">
                <el-input
                    v-model="autoFunc.description"
                    :placeholder="$t('tools.autoCodeAdmin.phDesc')"
                    disabled
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('tools.autoCodeAdmin.labelAbbr')">
                <el-input
                    v-model="autoFunc.abbreviation"
                    :placeholder="$t('tools.autoCodeAdmin.phAbbr')"
                    disabled
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item :label="$t('tools.autoCodeAdmin.labelAiFill')">
            <el-switch v-model="autoFunc.isAi" />
            <span class="text-sm text-red-600 p-2"
              >{{ $t('tools.autoCodeAdmin.aiUnstableHint') }}</span
            >
          </el-form-item>
          <template v-if="autoFunc.isAi">
            <el-form-item :label="$t('tools.autoCodeAdmin.labelAiAssist')">
              <div class="relative w-full">
                <el-input
                  type="textarea"
                  :placeholder="$t('tools.autoCodeAdmin.aiAssistPh')"
                  v-model="autoFunc.prompt"
                  :rows="5"
                  @input="autoFunc.router = autoFunc.router.replace(/\//g, '')"
                />
                <el-button
                  @click="aiAddFunc"
                  type="primary"
                  class="absolute right-2 bottom-2"
                  ><ai-lrag />{{ $t('tools.autoCodeAdmin.btnAiWrite') }}</el-button
                >
              </div>
            </el-form-item>
            <el-form-item :label="$t('tools.autoCodeAdmin.labelApiFunc')">
              <v-ace-editor
                v-model:value="autoFunc.apiFunc"
                lang="golang"
                theme="github_dark"
                class="h-80 w-full"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.autoCodeAdmin.labelServerFunc')">
              <v-ace-editor
                v-model:value="autoFunc.serverFunc"
                lang="golang"
                theme="github_dark"
                class="h-80 w-full"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.autoCodeAdmin.labelJsApi')">
              <v-ace-editor
                v-model:value="autoFunc.jsFunc"
                lang="javascript"
                theme="github_dark"
                class="h-80 w-full"
              />
            </el-form-item>
          </template>

          <el-form-item :label="$t('tools.autoCodeAdmin.labelFuncDesc')">
            <div class="flex w-full gap-2">
              <el-input
                class="flex-1"
                v-model="autoFunc.funcDesc"
                :placeholder="$t('tools.autoCodeAdmin.phFuncDesc')"
              />
              <el-button type="primary" @click="autoComplete"
                ><ai-lrag />{{ $t('tools.autoCodeAdmin.btnComplete') }}</el-button
              >
            </div>
          </el-form-item>
          <el-form-item :label="$t('tools.autoCodeAdmin.labelFuncName')">
            <el-input
              @blur="autoFunc.funcName = toUpperCase(autoFunc.funcName)"
              v-model="autoFunc.funcName"
              :placeholder="$t('tools.autoCodeAdmin.phFuncName')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.autoCodeAdmin.labelMethod')">
            <el-select v-model="autoFunc.method" :placeholder="$t('tools.autoCodeAdmin.phPickMethod')">
              <el-option
                v-for="item in ['GET', 'POST', 'PUT', 'DELETE']"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('tools.autoCodeAdmin.labelNeedAuth')">
            <el-switch
              v-model="autoFunc.isAuth"
              :active-text="$t('tools.autoCodeAdmin.authYes')"
              :inactive-text="$t('tools.autoCodeAdmin.authNo')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.autoCodeAdmin.labelRouterPath')">
            <el-input
              v-model="autoFunc.router"
              :placeholder="$t('tools.autoCodeAdmin.phRouterPath')"
              @input="autoFunc.router = autoFunc.router.replace(/\//g, '')"
            />
            <div>
              {{
                $t('tools.autoCodeAdmin.apiPathPreview', {
                  method: autoFunc.method,
                  abbr: autoFunc.abbreviation,
                  router: autoFunc.router
                })
              }}
            </div>
          </el-form-item>
        </el-form>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getSysHistory,
    rollback,
    delSysHistory,
    addFunc,
    llmAuto
  } from '@/api/autoCode.js'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { formatDate } from '@/utils/format'
  import { toUpperCase } from '@/utils/stringFun'

  import { VAceEditor } from 'vue3-ace-editor'
  import 'ace-builds/src-noconflict/mode-javascript'
  import 'ace-builds/src-noconflict/mode-golang'
  import 'ace-builds/src-noconflict/theme-github_dark'

  defineOptions({
    name: 'AutoCodeAdmin'
  })

  const { t } = useI18n()
  const aiLoading = ref(false)

  const formData = ref({
    id: undefined,
    deleteApi: true,
    deleteMenu: true,
    deleteTable: false
  })

  const router = useRouter()
  const dialogFormVisible = ref(false)
  const dialogFormTitle = ref('')

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])

  const activeInfo = ref('')

  const autoFunc = ref({
    package: '',
    funcName: '',
    structName: '',
    packageName: '',
    description: '',
    abbreviation: '',
    humpPackageName: '',
    businessDB: '',
    method: '',
    funcDesc: '',
    isAuth: false,
    isAi: false,
    apiFunc: '',
    serverFunc: '',
    jsFunc: ''
  })

  const addFuncBtn = (row) => {
    const req = JSON.parse(row.request)
    activeInfo.value = row.request
    autoFunc.value.package = req.package
    autoFunc.value.structName = req.structName
    autoFunc.value.packageName = req.packageName
    autoFunc.value.description = req.description
    autoFunc.value.abbreviation = req.abbreviation
    autoFunc.value.humpPackageName = req.humpPackageName
    autoFunc.value.businessDB = req.businessDB
    autoFunc.value.method = ''
    autoFunc.value.funcName = ''
    autoFunc.value.router = ''
    autoFunc.value.funcDesc = ''
    autoFunc.value.isAuth = false
    autoFunc.value.isAi = false
    autoFunc.value.apiFunc = ''
    autoFunc.value.serverFunc = ''
    autoFunc.value.jsFunc = ''
    funcFlag.value = true
  }

  const funcFlag = ref(false)

  const closeFunc = () => {
    funcFlag.value = false
  }

  const runFunc = async () => {
    // 首字母自动转换为大写
    autoFunc.value.funcName = toUpperCase(autoFunc.value.funcName)

    if (!autoFunc.value.funcName) {
      ElMessage.error(t('tools.autoCodeAdmin.errFuncName'))
      return
    }
    if (!autoFunc.value.method) {
      ElMessage.error(t('tools.autoCodeAdmin.errMethod'))
      return
    }
    if (!autoFunc.value.router) {
      ElMessage.error(t('tools.autoCodeAdmin.errRouter'))
      return
    }
    if (!autoFunc.value.funcDesc) {
      ElMessage.error(t('tools.autoCodeAdmin.errFuncDesc'))
      return
    }

    if (autoFunc.value.isAi) {
      if (
        !autoFunc.value.apiFunc ||
        !autoFunc.value.serverFunc ||
        !autoFunc.value.jsFunc
      ) {
        ElMessage.error(t('tools.autoCodeAdmin.errAiFirst'))
        return
      }
    }

    const res = await addFunc(autoFunc.value)
    if (res.code === 0) {
      ElMessage.success(t('tools.autoCodeAdmin.addFuncOk'))
      closeFunc()
    }
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

  // 查询
  const getTableData = async () => {
    const table = await getSysHistory({
      page: page.value,
      pageSize: pageSize.value
    })
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
    }
  }

  getTableData()

  const deleteRow = async (row) => {
    ElMessageBox.confirm(t('tools.autoCodeAdmin.confirmDeleteHistory'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await delSysHistory({ id: Number(row.ID) })
      if (res.code === 0) {
        ElMessage.success(t('common.deleteOk'))
        getTableData()
      }
    })
  }

  // 打开弹窗
  const openDialog = (row) => {
    dialogFormTitle.value = t('tools.autoCodeAdmin.rollbackTitle', { name: row.structName })
    formData.value.id = row.ID
    dialogFormVisible.value = true
  }

  // 关闭弹窗
  const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
      id: undefined,
      deleteApi: true,
      deleteMenu: true,
      deleteTable: false
    }
  }

  // 确认删除表
  const deleteTableCheck = (flag) => {
    if (flag) {
      ElMessageBox.confirm(
        t('tools.autoCodeAdmin.rollbackDropTableBody1'),
        t('common.tipTitle'),
        {
          closeOnClickModal: false,
          distinguishCancelAndClose: true,
          confirmButtonText: t('common.ok'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
        .then(() => {
          ElMessageBox.confirm(
            t('tools.autoCodeAdmin.rollbackDropTableBody2'),
            t('tools.autoCodeAdmin.rollbackDropTableTitle2'),
            {
              closeOnClickModal: false,
              distinguishCancelAndClose: true,
              confirmButtonText: t('common.ok'),
              cancelButtonText: t('common.cancel'),
              type: 'warning'
            }
          ).catch(() => {
            formData.value.deleteTable = false
          })
        })
        .catch(() => {
          formData.value.deleteTable = false
        })
    }
  }

  const enterDialog = async () => {
    const res = await rollback(formData.value)
    if (res.code === 0) {
      ElMessage.success(t('tools.autoCodeAdmin.rollbackOk'))
      getTableData()
    }
  }

  const goAutoCode = (row, isAdd) => {
    if (row) {
      router.push({
        name: 'autoCodeEdit',
        params: {
          id: row.ID
        },
        query: {
          isAdd: isAdd
        }
      })
    } else {
      router.push({ name: 'autoCode' })
    }
  }

  const aiAddFunc = async () => {
    aiLoading.value = true
    autoFunc.value.apiFunc = ''
    autoFunc.value.serverFunc = ''
    autoFunc.value.jsFunc = ''

    if (!autoFunc.value.prompt) {
      ElMessage.error(t('tools.autoCodeAdmin.errPromptRequired'))
      return
    }

    const res = await addFunc({ ...autoFunc.value, isPreview: true })
    if (res.code !== 0) {
      aiLoading.value = false
      ElMessage.error(res.msg)
      return
    }

    const aiRes = await llmAuto({
      structInfo: activeInfo.value,
      template: JSON.stringify(res.data),
      prompt: autoFunc.value.prompt,
      mode: 'addFunc'
    })
    aiLoading.value = false
    if (aiRes.code === 0) {
      try {
        const aiData = JSON.parse(aiRes.data)
        autoFunc.value.apiFunc = aiData.api
        autoFunc.value.serverFunc = aiData.server
        autoFunc.value.jsFunc = aiData.js
        autoFunc.value.method = aiData.method
        autoFunc.value.funcName = aiData.funcName
        const routerArr = aiData.router.split('/')
        autoFunc.value.router = routerArr[routerArr.length - 1]
        autoFunc.value.funcDesc = autoFunc.value.prompt
      } catch (_) {
        ElMessage.error(t('tools.autoCodeAdmin.errAiBusy'))
      }
    }
  }

  const autoComplete = async () => {
    aiLoading.value = true
    const aiRes = await llmAuto({
      prompt: autoFunc.value.funcDesc,
      mode: 'autoCompleteFunc'
    })
    aiLoading.value = false
    if (aiRes.code === 0) {
      try {
        const aiData = JSON.parse(aiRes.data)
        autoFunc.value.method = aiData.method
        autoFunc.value.funcName = aiData.funcName
        autoFunc.value.router = aiData.router
        autoFunc.value.prompt = autoFunc.value.funcDesc
      } catch (_) {
        ElMessage.error(t('tools.autoCodeAdmin.errAiGlitch'))
      }
    }
  }
</script>
