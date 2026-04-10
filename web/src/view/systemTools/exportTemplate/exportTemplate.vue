<template>
  <div>
    <WarningBar
      :title="$t('tools.exportTemplate.warningBar')"
      href="https://lightningrag.feishu.cn/docx/KwjxdnvatozgwIxGV0rcpkZSn4d"
    />
    <div class="lrag-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        class="demo-form-inline"
        :rules="searchRule"
        @keyup.enter="onSubmit"
      >
        <el-form-item :label="$t('tools.exportTemplate.labelCreatedDate')" prop="createdAt">
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
            v-model="searchInfo.startCreatedAt"
            type="datetime"
            :placeholder="$t('tools.exportTemplate.startDate')"
            :disabled-date="
              (time) =>
                searchInfo.endCreatedAt
                  ? time.getTime() > searchInfo.endCreatedAt.getTime()
                  : false
            "
          />
          —
          <el-date-picker
            v-model="searchInfo.endCreatedAt"
            type="datetime"
            :placeholder="$t('tools.exportTemplate.endDate')"
            :disabled-date="
              (time) =>
                searchInfo.startCreatedAt
                  ? time.getTime() < searchInfo.startCreatedAt.getTime()
                  : false
            "
          />
        </el-form-item>
        <el-form-item :label="$t('common.colTemplateName')" prop="name">
          <el-input v-model="searchInfo.name" :placeholder="$t('common.phSearchCondition')" />
        </el-form-item>
        <el-form-item :label="$t('common.colTableName')" prop="tableName">
          <el-input v-model="searchInfo.tableName" :placeholder="$t('common.phSearchCondition')" />
        </el-form-item>
        <el-form-item :label="$t('common.colTemplateId')" prop="templateID">
          <el-input v-model="searchInfo.templateID" :placeholder="$t('common.phSearchCondition')" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit"
            >{{ $t('common.query') }}</el-button
          >
          <el-button icon="refresh" @click="onReset">{{ $t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openDialog"
          >{{ $t('common.add') }}</el-button
        >

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
        <el-table-column align="left" :label="$t('common.colDate')" width="180">
          <template #default="scope">{{
            formatDate(scope.row.CreatedAt)
          }}</template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colDatabase')" width="120">
          <template #default="scope">
            <span>{{ scope.row.dbName || $t('tools.exportTemplate.mainDb') }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('common.colTemplateId')"
          prop="templateID"
          width="120"
        />
        <el-table-column
          align="left"
          :label="$t('common.colTemplateName')"
          prop="name"
          width="120"
        />
        <el-table-column
          align="left"
          :label="$t('common.colTableName')"
          prop="tableName"
          width="120"
        />
        <el-table-column
          align="left"
          :label="$t('common.colTemplateInfo')"
          prop="templateInfo"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column align="left" :label="$t('common.colActions')" min-width="280">
          <template #default="scope">
            <el-button
                type="primary"
                link
                icon="documentCopy"
                class="table-button"
                @click="copyFunc(scope.row)"
            >{{ $t('common.btnCopy') }}</el-button
            >
            <el-button
              type="primary"
              link
              icon="edit-pen"
              class="table-button"
              @click="showCode(scope.row)"
              >{{ $t('tools.exportTemplate.btnCodeSqlPreview') }}</el-button
            >
            <el-button
              type="primary"
              link
              icon="edit"
              class="table-button"
              @click="updateSysExportTemplateFunc(scope.row)"
              >{{ $t('tools.exportTemplate.btnChange') }}</el-button
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
      v-model="dialogFormVisible"
      size="60%"
      :before-close="closeDialog"
      :title="type === 'create' ? $t('tools.exportTemplate.drawerCreate') : $t('tools.exportTemplate.drawerUpdate')"
      :show-close="false"
      destroy-on-close
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ type === 'create' ? $t('tools.exportTemplate.drawerCreate') : $t('tools.exportTemplate.drawerUpdate') }}</span>
          <div>
            <el-button @click="closeDialog">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="enterDialog">{{ $t('common.ok') }}</el-button>
          </div>
        </div>
      </template>

      <el-form
        ref="elFormRef"
        :model="formData"
        label-position="right"
        :rules="rule"
        label-width="100px"
        v-loading="aiLoading"
        :element-loading-text="$t('tools.exportTemplate.aiLoadingText')"
      >
        <el-form-item :label="$t('tools.exportTemplate.dbLabel')" prop="dbName">
          <template #label>
            <el-tooltip
              :content="$t('tools.exportTemplate.dbTooltip')"
              placement="bottom"
              effect="light"
            >
              <div>
                {{ $t('tools.exportTemplate.dbLabel') }} <el-icon><QuestionFilled /></el-icon>
              </div>
            </el-tooltip>
          </template>
          <el-select
            v-model="formData.dbName"
            clearable
            @change="dbNameChange"
            :placeholder="$t('tools.exportTemplate.pickDb')"
          >
            <el-option
              v-for="item in dbList"
              :key="item.aliasName"
              :value="item.aliasName"
              :label="item.aliasName"
              :disabled="item.disable"
            >
              <div>
                <span>{{ item.aliasName }}</span>
                <span style="float: right; color: #8492a6; font-size: 13px">{{
                  item.dbName
                }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('tools.exportTemplate.tablesLabel')" prop="tables">
          <el-select
            multiple
            v-model="tables"
            clearable
            :placeholder="$t('tools.exportTemplate.tablesPhAi')"
          >
            <el-option
              v-for="item in tableOptions"
              :key="item.tableName"
              :label="item.tableName"
              :value="item.tableName"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('tools.exportTemplate.aiAssistLabel')" prop="ai">
          <div class="relative w-full">
            <el-input
              type="textarea"
              v-model="prompt"
              :clearable="true"
              :rows="5"
              :placeholder="$t('tools.exportTemplate.aiPromptPh')"
            />
            <el-button
              class="absolute bottom-2 right-2"
              type="primary"
              @click="autoExport"
              ><el-icon><ai-lrag /></el-icon>{{ $t('tools.exportTemplate.btnAiWrite') }}</el-button
            >
          </div>
        </el-form-item>

        <el-form-item :label="$t('tools.exportTemplate.tableNameLabel')" clearable prop="tableName">
          <div class="w-full flex gap-4">
            <el-select
              v-model="formData.tableName"
              class="flex-1"
              filterable
              :placeholder="$t('tools.exportTemplate.pickTable')"
            >
              <el-option
                v-for="item in tableOptions"
                :key="item.tableName"
                :label="item.tableName"
                :value="item.tableName"
              />
            </el-select>
            <el-button
              :disabled="!formData.tableName"
              type="primary"
              @click="getColumnFunc(true)"
              ><el-icon><ai-lrag /></el-icon>{{ $t('tools.exportTemplate.btnAiComplete') }}</el-button
            >
            <el-button
              :disabled="!formData.tableName"
              type="primary"
              @click="getColumnFunc(false)"
              >{{ $t('tools.exportTemplate.btnGenTemplate') }}</el-button
            >
          </div>
        </el-form-item>

        <el-form-item :label="$t('tools.exportTemplate.nameLabel')" prop="name">
          <el-input
            v-model="formData.name"
            :clearable="true"
            :placeholder="$t('tools.exportTemplate.phTemplateName')"
          />
        </el-form-item>

        <el-form-item :label="$t('tools.exportTemplate.templateIdLabel')" prop="templateID">
          <el-input
            v-model="formData.templateID"
            :clearable="true"
            :placeholder="$t('tools.exportTemplate.phTemplateId')"
          />
        </el-form-item>

        <el-tabs v-model="activeName">
          <el-tab-pane :label="$t('tools.exportTemplate.tabAuto')" name="auto" class="pt-2">
            <el-form-item :label="$t('tools.exportTemplate.joinLabel')">
              <div
                v-for="(join, key) in formData.joinTemplate"
                :key="key"
                class="flex gap-4 w-full mb-2"
              >
                <el-select v-model="join.joins" :placeholder="$t('tools.exportTemplate.pickJoinType')">
                  <el-option label="LEFT JOIN" value="LEFT JOIN" />
                  <el-option label="INNER JOIN" value="INNER JOIN" />
                  <el-option label="RIGHT JOIN" value="RIGHT JOIN" />
                </el-select>
                <el-input v-model="join.table" :placeholder="$t('tools.exportTemplate.phJoinTable')" />
                <el-input
                  v-model="join.on"
                  :placeholder="$t('tools.exportTemplate.phJoinOn')"
                />
                <el-button
                  type="danger"
                  icon="delete"
                  @click="() => formData.joinTemplate.splice(key, 1)"
                  >{{ $t('tools.exportTemplate.btnDel') }}</el-button
                >
              </div>
              <div class="flex justify-end w-full">
                <el-button type="primary" icon="plus" @click="addJoin"
                  >{{ $t('tools.exportTemplate.btnAddJoin') }}</el-button
                >
              </div>
            </el-form-item>

            <el-form-item :label="$t('tools.exportTemplate.defaultLimit')">
              <el-input-number
                v-model="formData.limit"
                :step="1"
                :step-strictly="true"
                :precision="0"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.exportTemplate.defaultOrder')">
              <el-input v-model="formData.order" :placeholder="$t('tools.exportTemplate.phOrderExample')" />
            </el-form-item>
            <el-form-item :label="$t('tools.exportTemplate.exportConditions')">
              <div
                v-for="(condition, key) in formData.conditions"
                :key="key"
                class="flex gap-4 w-full mb-2"
              >
                <el-input
                  v-model="condition.from"
                  :placeholder="$t('tools.exportTemplate.phCondJsonKey')"
                />
                <el-input v-model="condition.column" :placeholder="$t('tools.exportTemplate.phCondColumn')" />
                <el-select
                  v-model="condition.operator"
                  :placeholder="$t('tools.exportTemplate.pickSearchOp')"
                >
                  <el-option
                    v-for="item in typeSearchOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
                <el-button
                  type="danger"
                  icon="delete"
                  @click="() => formData.conditions.splice(key, 1)"
                  >{{ $t('tools.exportTemplate.btnDel') }}</el-button
                >
              </div>
              <div class="flex justify-end w-full">
                <el-button type="primary" icon="plus" @click="addCondition"
                  >{{ $t('tools.exportTemplate.btnAddCond') }}</el-button
                >
              </div>
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane :label="$t('tools.exportTemplate.tabSql')" name="sql"  class="pt-2">
            <el-form-item :label="$t('tools.exportTemplate.exportSqlLabel')" prop="sql">
              <el-input
                v-model="formData.sql"
                type="textarea"
                :rows="10"
                :placeholder="$t('tools.exportTemplate.phExportSql')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.exportTemplate.importSqlLabel')" prop="importSql">
              <el-input
                v-model="formData.importSql"
                type="textarea"
                :rows="10"
                :placeholder="$t('tools.exportTemplate.phImportSql')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.exportTemplate.exportConditions')">
              <span class="text-sm text-gray-600 dark:text-gray-300">{{ $t('tools.exportTemplate.sqlCondHint') }}</span>
            </el-form-item>
          </el-tab-pane>
        </el-tabs>

        <el-form-item :label="$t('tools.exportTemplate.templateInfoLabel')" prop="templateInfo">
          <el-input
            v-model="formData.templateInfo"
            type="textarea"
            :rows="12"
            :clearable="true"
            :placeholder="templatePlaceholder"
          />
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 合并：代码模板 + SQL预览 抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      size="70%"
      :title="$t('tools.exportTemplate.drawerPreviewTitle')"
      :show-close="true"
      destroy-on-close
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('tools.exportTemplate.drawerPreviewTitle') }}</span>
          <div>
            <el-button @click="drawerVisible = false">{{ $t('common.btnClose') }}</el-button>
            <el-button v-if="activeTab === 'sql'" type="primary" @click="runPreview">{{ $t('tools.exportTemplate.btnRunPreview') }}</el-button>
          </div>
        </div>
      </template>
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane :label="$t('tools.exportTemplate.tabCodeTpl')" name="code">
          <v-ace-editor
            v-model:value="webCode"
            lang="vue"
            theme="github_dark"
            class="w-full h-96"
            :options="{ showPrintMargin: false, fontSize: 14 }"
          />
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.exportTemplate.tabSqlPreview')" name="sql">
          <div class="flex flex-col gap-4">
            <div class="w-full">
              <el-form :model="previewForm" label-width="120px">
                <el-form-item :label="$t('tools.exportTemplate.filterDeleted')">
                  <el-switch v-model="previewForm.filterDeleted" />
                </el-form-item>
                <el-form-item :label="$t('tools.exportTemplate.previewOrder')">
                  <el-input v-model="previewForm.order" :placeholder="$t('tools.exportTemplate.phPreviewOrder')" />
                </el-form-item>
                <el-form-item :label="$t('tools.exportTemplate.limitRows')">
                  <el-input-number v-model="previewForm.limit" :min="0" />
                </el-form-item>
                <el-form-item :label="$t('tools.exportTemplate.offsetRows')">
                  <el-input-number v-model="previewForm.offset" :min="0" />
                </el-form-item>

                <el-divider content-position="left">{{ $t('tools.exportTemplate.dividerQuery') }}</el-divider>
                <div v-if="previewConditions.length === 0" class="text-gray">{{ $t('tools.exportTemplate.noPreviewCond') }}</div>
                <template v-for="(cond, idx) in previewConditions" :key="idx">
                  <el-form-item :label="cond.column + ' ' + cond.operator">
                    <template v-if="cond.operator === 'BETWEEN'">
                      <div class="flex gap-2 w-full">
                        <el-input v-model="previewForm['start' + cond.from]" :placeholder="$t('tools.exportTemplate.phStartVal', { from: cond.from })" />
                        <el-input v-model="previewForm['end' + cond.from]" :placeholder="$t('tools.exportTemplate.phEndVal', { from: cond.from })" />
                      </div>
                    </template>
                    <template v-else>
                      <el-input v-model="previewForm[cond.from]" :placeholder="$t('tools.exportTemplate.phVarNamed', { name: cond.from })" />
                    </template>
                  </el-form-item>
                </template>
              </el-form>
            </div>
            <div class="w-full">
              <v-ace-editor
                v-model:value="previewSQLCode"
                lang="sql"
                theme="github_dark"
                class="w-full h-96"
                :options="aceOptions"
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createSysExportTemplate,
    deleteSysExportTemplate,
    deleteSysExportTemplateByIds,
    updateSysExportTemplate,
    findSysExportTemplate,
    getSysExportTemplateList
  } from '@/api/exportTemplate.js'
  import { previewSQL } from '@/api/exportTemplate.js'

  // 全量引入格式化工具 请按需保留
  import { formatDate } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, reactive, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { getDB, getTable, getColumn, llmAuto } from '@/api/autoCode'
  import { getCode } from './code'
  import { VAceEditor } from 'vue3-ace-editor'
  import { QuestionFilled } from '@element-plus/icons-vue'

  import 'ace-builds/src-noconflict/mode-vue'
  import 'ace-builds/src-noconflict/theme-github_dark'
  import 'ace-builds/src-noconflict/mode-sql'

  defineOptions({
    name: 'ExportTemplate'
  })

  const { t, locale } = useI18n()
  const templatePlaceholder = computed(() => t('tools.exportTemplate.templateInfoHelp'))

  // 自动化生成的字典（可能为空）以及字段
  const formData = ref({
    name: '',
    tableName: '',
    dbName: '',
    templateID: '',
    templateInfo: '',
    limit: 0,
    order: '',
    conditions: [],
    joinTemplate: [],
    sql: '',
    importSql: ''
  })

  const activeName = ref('auto')

  const prompt = ref('')
  const tables = ref([])

  const typeSearchOptions = ref([
    {
      label: '=',
      value: '='
    },
    {
      label: '<>',
      value: '<>'
    },
    {
      label: '>',
      value: '>'
    },
    {
      label: '<',
      value: '<'
    },
    {
      label: 'LIKE',
      value: 'LIKE'
    },
    {
      label: 'BETWEEN',
      value: 'BETWEEN'
    },
    {
      label: 'NOT BETWEEN',
      value: 'NOT BETWEEN'
    },
    {
      label: 'IN',
      value: 'IN'
    },
    {
      label: 'NOT IN',
      value: 'NOT IN'
    },
  ])

  const addCondition = () => {
    formData.value.conditions.push({
      from: '',
      column: '',
      operator: ''
    })
  }

  const addJoin = () => {
    formData.value.joinTemplate.push({
      joins: 'LEFT JOIN',
      table: '',
      on: ''
    })
  }

  const rule = computed(() => ({
    name: [
      {
        required: true,
        message: t('tools.exportTemplate.ruleNameRequired'),
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: t('common.ruleWhitespaceOnly'),
        trigger: ['input', 'blur']
      }
    ],
    tableName: [
      {
        required: true,
        message: t('tools.exportTemplate.ruleTableNameRequired'),
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: t('common.ruleWhitespaceOnly'),
        trigger: ['input', 'blur']
      }
    ],
    templateID: [
      {
        required: true,
        message: t('tools.exportTemplate.ruleTemplateIDRequired'),
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: t('common.ruleWhitespaceOnly'),
        trigger: ['input', 'blur']
      }
    ],
    templateInfo: [
      {
        required: true,
        message: t('tools.exportTemplate.ruleTemplateInfoRequired'),
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: t('common.ruleWhitespaceOnly'),
        trigger: ['input', 'blur']
      }
    ]
  }))

  const searchRule = reactive({
    createdAt: [
      {
        validator: (rule, value, callback) => {
          if (
            searchInfo.value.startCreatedAt &&
            !searchInfo.value.endCreatedAt
          ) {
            callback(new Error(t('common.dateNeedEnd')))
          } else if (
            !searchInfo.value.startCreatedAt &&
            searchInfo.value.endCreatedAt
          ) {
            callback(new Error(t('common.dateNeedStart')))
          } else if (
            searchInfo.value.startCreatedAt &&
            searchInfo.value.endCreatedAt &&
            (searchInfo.value.startCreatedAt.getTime() ===
              searchInfo.value.endCreatedAt.getTime() ||
              searchInfo.value.startCreatedAt.getTime() >
                searchInfo.value.endCreatedAt.getTime())
          ) {
            callback(new Error(t('common.dateStartBeforeEnd')))
          } else {
            callback()
          }
        },
        trigger: 'change'
      }
    ]
  })

  const elFormRef = ref()
  const elSearchFormRef = ref()

  // =========== 表格控制部分 ===========
  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const searchInfo = ref({})

  const dbList = ref([])
  const tableOptions = ref([])
  const aiLoading = ref(false)

  const getTablesCloumn = async () => {
    const tablesMap = {}
    const promises = tables.value.map(async (item) => {
      const res = await getColumn({
        businessDB: formData.value.dbName,
        tableName: item
      })
      if (res.code === 0) {
        tablesMap[item] = res.data.columns
      }
    })
    await Promise.all(promises)
    return tablesMap
  }

  const autoExport = async () => {
    if (tables.value.length === 0) {
      ElMessage({
        type: 'error',
        message: t('tools.exportTemplate.errPickTablesForAi')
      })
      return
    }
    aiLoading.value = true
    const tableMap = await getTablesCloumn()
    const aiRes = await llmAuto({
      prompt: prompt.value,
      tableMap: JSON.stringify(tableMap),
      mode: 'autoExportTemplate'
    })
    aiLoading.value = false
    if (aiRes.code === 0) {
      const aiData = JSON.parse(aiRes.data)
      formData.value.name = aiData.name
      formData.value.tableName = aiData.tableName
      formData.value.templateID = aiData.templateID
      formData.value.templateInfo = JSON.stringify(aiData.templateInfo, null, 2)
      formData.value.joinTemplate = aiData.joinTemplate
    }
  }

  const getDbFunc = async () => {
    const res = await getDB()
    if (res.code === 0) {
      dbList.value = res.data.dbList
    }
  }

  getDbFunc()

  const dbNameChange = () => {
    formData.value.tableName = ''
    formData.value.templateInfo = ''
    tables.value = []
    getTableFunc()
  }

  const getTableFunc = async () => {
    const res = await getTable({ businessDB: formData.value.dbName })
    if (res.code === 0) {
      tableOptions.value = res.data.tables
    }
    formData.value.tableName = ''
  }
  getTableFunc()
  const getColumnFunc = async (aiFLag) => {
    if (!formData.value.tableName) {
      ElMessage({
        type: 'error',
        message: t('tools.exportTemplate.errPickDbTable')
      })
      return
    }
    formData.value.templateInfo = ''
    aiLoading.value = true
    const res = await getColumn({
      businessDB: formData.value.dbName,
      tableName: formData.value.tableName
    })
    if (res.code === 0) {
      if (aiFLag) {
        const aiRes = await llmAuto({
          data: JSON.stringify(res.data.columns),
          mode: 'exportCompletion'
        })
        if (aiRes.code === 0) {
          const aiData = JSON.parse(aiRes.data)
          aiLoading.value = false
          formData.value.templateInfo = JSON.stringify(
            aiData.templateInfo,
            null,
            2
          )
          formData.value.name = aiData.name
          formData.value.templateID = aiData.templateID
          return
        }
        ElMessage.warning(t('tools.exportTemplate.warnAiFallback'))
      }

      // 中文界面用库表注释作 Excel 列名；其它语言用字段名，避免英文界面出现大量中文
      const templateInfo = {}
      const useZhComment = String(locale.value || '')
        .toLowerCase()
        .startsWith('zh')
      res.data.columns.forEach((item) => {
        templateInfo[item.columnName] = useZhComment
          ? item.columnComment || item.columnName
          : item.columnName
      })
      formData.value.templateInfo = JSON.stringify(templateInfo, null, 2)
    }
    aiLoading.value = false
  }

  // 重置
  const onReset = () => {
    searchInfo.value = {}
    getTableData()
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
    const table = await getSysExportTemplateList({
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
      deleteSysExportTemplateFunc(row)
    })
  }

  // 多选删除
  const onDelete = async () => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: t('common.pickRowsToDelete')
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map((item) => {
          ids.push(item.ID)
        })
      const res = await deleteSysExportTemplateByIds({ ids })
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

  // 行为控制标记（弹窗内部需要增还是改）
  const type = ref('')

  // 复制
  const copyFunc = async (row) => {
    let copyData
    const res = await findSysExportTemplate({ ID: row.ID })
    if (res.code === 0) {
      copyData = JSON.parse(JSON.stringify(res.data.resysExportTemplate))
      if (!copyData.conditions) {
        copyData.conditions = []
      }
      if (!copyData.joinTemplate) {
        copyData.joinTemplate = []
      }
      if (!copyData.sql) {
        copyData.sql = ''
      }
      if (!copyData.importSql) {
        copyData.importSql = ''
      }
      delete copyData.ID
      delete copyData.CreatedAt
      delete copyData.UpdatedAt
      copyData.templateID = copyData.templateID + '_copy'
      copyData.name = copyData.name + '_copy'
      formData.value = copyData
      dialogFormVisible.value = true
    }
  }

  // 更新行
  const updateSysExportTemplateFunc = async (row) => {
    const res = await findSysExportTemplate({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data.resysExportTemplate
      if (!formData.value.conditions) {
        formData.value.conditions = []
      }
      if (!formData.value.joinTemplate) {
        formData.value.joinTemplate = []
      }
      if (!formData.value.sql) {
        formData.value.sql = ''
      }
      if (!formData.value.importSql) {
        formData.value.importSql = ''
      }
      if (formData.value.sql || formData.value.importSql) {
        activeName.value = 'sql'
      } else {
        activeName.value = 'auto'
      }
      dialogFormVisible.value = true
    }
  }

  // 删除行
  const deleteSysExportTemplateFunc = async (row) => {
    const res = await deleteSysExportTemplate({ ID: row.ID })
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
  const drawerVisible = ref(false)
  const activeTab = ref('code')
  // 弹窗控制标记
  const dialogFormVisible = ref(false)

  const webCode = ref('')

  const showCode = (row) => {
    webCode.value = getCode(row.templateID)
    activeTab.value = 'code'
    drawerVisible.value = true
  }

  // 预览 SQL
  const previewForm = ref({ filterDeleted: true, order: '', limit: 0, offset: 0 })
  const previewSQLCode = ref('')
  const previewTemplate = ref(null)
  const previewConditions = ref([])
  const aceOptions = { wrap: true, showPrintMargin: false, fontSize: 14 }

  const openPreview = async (row) => {
    // 获取模板完整信息以展示条件输入项
    const res = await findSysExportTemplate({ ID: row.ID })
    if (res.code === 0) {
      previewTemplate.value = res.data.resysExportTemplate
      previewConditions.value = (previewTemplate.value.conditions || []).map((c) => ({
        from: c.from,
        column: c.column,
        operator: c.operator
      }))
      // 预填默认的排序与限制
      previewForm.value.order = previewTemplate.value.order || ''
      previewForm.value.limit = previewTemplate.value.limit || 0
      previewForm.value.offset = 0
      previewSQLCode.value = ''
      activeTab.value = 'sql'
      drawerVisible.value = true
    }
  }

  const runPreview = async () => {
    if (!previewTemplate.value) return
    // 组装 params，与导出组件保持一致
    const paramsCopy = JSON.parse(JSON.stringify(previewForm.value))
    // 将布尔与数值等按照导出组件规则编码
    if (paramsCopy.filterDeleted) paramsCopy.filterDeleted = 'true'
    const entries = Object.entries(paramsCopy).filter(([key, v]) => {
      if (v === '' || v === null || v === undefined) return false
      if ((key === 'limit' || key === 'offset') && Number(v) === 0) return false
      return true
    })
    const params = entries
      .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
      .join('&')

    const res = await previewSQL({ templateID: previewTemplate.value.templateID, params })
    if (res.code === 0) {
      previewSQLCode.value = res.data.sql || ''
    }
  }

  // 打开弹窗
  const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
  }

  // 关闭弹窗
  const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
      name: '',
      tableName: '',
      templateID: '',
      templateInfo: '',
      limit: 0,
      order: '',
      conditions: [],
      joinTemplate: [],
      sql: '',
      importSql: ''
    }
    activeName.value = 'auto'
  }
  // 弹窗确定
  const enterDialog = async () => {
    // 判断 formData.templateInfo 是否为标准json格式 如果不是标准json 则辅助调整
    try {
      JSON.parse(formData.value.templateInfo)
    } catch (_) {
      ElMessage({
        type: 'error',
        message: t('tools.exportTemplate.errTemplateInfoJson')
      })
      return
    }

    const reqData = JSON.parse(JSON.stringify(formData.value))
    if (activeName.value === 'sql') {
      reqData.conditions = []
      reqData.joinTemplate = []
      reqData.limit = 0
      reqData.order = ''
    } else {
      reqData.sql = ''
      reqData.importSql = ''
    }

    for (let i = 0; i < reqData.conditions.length; i++) {
      if (
        !reqData.conditions[i].from ||
        !reqData.conditions[i].column ||
        !reqData.conditions[i].operator
      ) {
        ElMessage({
          type: 'error',
          message: t('tools.exportTemplate.errExportCondIncomplete')
        })
        return
      }
      reqData.conditions[i].templateID = reqData.templateID
    }

    for (let i = 0; i < reqData.joinTemplate.length; i++) {
      if (!reqData.joinTemplate[i].joins || !reqData.joinTemplate[i].on) {
        ElMessage({
          type: 'error',
          message: t('tools.exportTemplate.errJoinIncomplete')
        })
        return
      }
      reqData.joinTemplate[i].templateID = reqData.templateID
    }

    elFormRef.value?.validate(async (valid) => {
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createSysExportTemplate(reqData)
          break
        case 'update':
          res = await updateSysExportTemplate(reqData)
          break
        default:
          res = await createSysExportTemplate(reqData)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('tools.exportTemplate.saveOk')
        })
        closeDialog()
        getTableData()
      }
    })
  }
</script>

<style></style>
