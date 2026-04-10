<template>
  <div>
    <warning-bar :title="$t('tools.autoCode.page.devOnlyWarn')" />
    <div class="lrag-search-box" v-if="!isAdd">
      <div class="text-lg mb-2 text-gray-600">
        {{ $t('tools.autoCode.page.aiCreateTitle') }}
      </div>
      <div class="relative">
        <el-input
          v-model="prompt"
          type="textarea"
          :rows="5"
          :maxlength="2000"
          :placeholder="$t('tools.autoCode.page.promptPh')"
          resize="none"
          @focus="handleFocus"
          @blur="handleBlur"
        />

        <div class="flex absolute right-28 bottom-2">
          <el-tooltip effect="light">
            <template #content>
              <div>
                {{ $t('tools.autoCode.page.aiFreeTooltip') }}
              </div>
            </template>
            <el-button
                :disabled="form.onlyTemplate"
                type="primary"
                @click="eyeFunc()"
            >
              <el-icon size="18">
                <ai-lrag />
              </el-icon>
              {{ $t('tools.autoCode.page.btnVision') }}
            </el-button>
          </el-tooltip>
        </div>

        <div class="flex absolute right-2 bottom-2">
          <el-tooltip effect="light">
            <template #content>
              <div>
                {{ $t('tools.autoCode.page.aiFreeTooltip') }}
              </div>
            </template>
            <el-button
              :disabled="form.onlyTemplate"
              type="primary"
              @click="llmAutoFunc()"
            >
              <el-icon size="18">
                <ai-lrag />
              </el-icon>
              {{ $t('tools.autoCode.page.btnGenerate') }}
            </el-button>
          </el-tooltip>
        </div>
      </div>
    </div>
    <!-- 从数据库直接获取字段 -->
    <div class="lrag-search-box" v-if="!isAdd">
      <div class="text-lg mb-2 text-gray-600">{{ $t('tools.autoCode.page.fromDbTitle') }}</div>
      <el-form
        ref="getTableForm"
        :inline="true"
        :model="dbform"
        label-width="120px"
      >
        <el-row class="w-full">
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.bizDbLabel')" prop="selectDBtype" class="w-full">
              <template #label>
                <el-tooltip
                  :content="$t('tools.autoCode.page.ttBizDbPicker')"
                  placement="bottom"
                  effect="light"
                >
                  <div>
                    {{ $t('tools.autoCode.page.bizDbLabel') }} <el-icon><QuestionFilled /></el-icon>
                  </div>
                </el-tooltip>
              </template>
              <el-select
                v-model="dbform.businessDB"
                clearable
                :placeholder="$t('tools.autoCode.page.phBizDb')"
                @change="getDbFunc"
                class="w-full"
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
                    <span
                      style="float: right; color: #8492a6; font-size: 13px"
                      >{{ item.dbName }}</span
                    >
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.dbNameLabel')" prop="structName" class="w-full">
              <el-select
                v-model="dbform.dbName"
                clearable
                filterable
                :placeholder="$t('tools.autoCode.page.phDbName')"
                class="w-full"
                @change="getTableFunc"
              >
                <el-option
                  v-for="item in dbOptions"
                  :key="item.database"
                  :label="item.database"
                  :value="item.database"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.tableLabel')" prop="structName" class="w-full">
              <el-select
                v-model="dbform.tableName"
                :disabled="!dbform.dbName"
                class="w-full"
                filterable
                :placeholder="$t('tools.autoCode.page.phTable')"
              >
                <el-option
                  v-for="item in tableOptions"
                  :key="item.tableName"
                  :label="item.tableName"
                  :value="item.tableName"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item class="w-full">
              <div class="flex justify-end w-full">
                <el-button type="primary" @click="getColumnFunc">
                  {{ $t('tools.autoCode.page.btnUseTable') }}
                </el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>
    <div class="lrag-search-box">
      <!-- 初始版本自动化代码工具 -->
      <div class="text-lg mb-2 text-gray-600">{{ $t('tools.autoCode.page.autoStructTitle') }}</div>
      <el-form
        :disabled="isAdd"
        ref="autoCodeForm"
        :rules="rules"
        :model="form"
        label-width="120px"
        :inline="true"
      >
        <el-row class="w-full">
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.structNameLabel')" prop="structName" class="w-full">
              <div class="flex gap-2 w-full">
                <el-input
                  v-model="form.structName"
                  :placeholder="$t('tools.autoCode.page.phStructUpper')"
                />
                <el-button
                  :disabled="form.onlyTemplate"
                  type="primary"
                  @click="llmAutoFunc(true)"
                >
                  <el-icon size="18">
                    <ai-lrag />
                  </el-icon>
                  {{ $t('tools.autoCode.page.btnGenerate') }}
                </el-button>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.structAbbr')" prop="abbreviation" class="w-full">
              <template #label>
                <el-tooltip
                  :content="$t('tools.autoCode.page.ttAbbr')"
                  placement="bottom"
                  effect="light"
                >
                  <div>
                    {{ $t('tools.autoCode.page.structAbbr') }} <el-icon><QuestionFilled /></el-icon>
                  </div>
                </el-tooltip>
              </template>
              <el-input
                v-model="form.abbreviation"
                :placeholder="$t('tools.autoCode.page.phStructAbbr')"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.displayNameLabel')" prop="description" class="w-full">
              <el-input
                v-model="form.description"
                :placeholder="$t('tools.autoCode.page.phApiDesc')"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.tableNameLabel')" prop="tableName" class="w-full">
              <el-input
                v-model="form.tableName"
                :placeholder="$t('tools.autoCode.page.phTableOptional')"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row class="w-full">
          <el-col :span="6">
            <el-form-item prop="packageName" class="w-full">
              <template #label>
                <el-tooltip
                  :content="$t('tools.autoCode.page.ttFileName')"
                  placement="bottom"
                  effect="light"
                >
                  <div>
                    {{ $t('tools.autoCode.page.fileNameLabel') }} <el-icon><QuestionFilled /></el-icon>
                  </div>
                </el-tooltip>
              </template>
              <el-input
                v-model="form.packageName"
                :placeholder="$t('tools.autoCode.page.phFileName')"
                @blur="toLowerCaseFunc(form, 'packageName')"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item
              :label="$t('tools.autoCode.page.pickTemplateLabel')"
              prop="package"
              class="w-full relative"
            >
              <el-select v-model="form.package" class="w-full pr-12" filterable>
                <el-option
                  v-for="item in pkgs"
                  :key="item.ID"
                  :value="item.packageName"
                  :label="item.packageName"
                />
              </el-select>
              <span class="absolute right-0">
                <el-icon
                  class="cursor-pointer ml-2 text-gray-600"
                  @click="getPkgs"
                >
                  <refresh />
                </el-icon>
                <el-icon
                  class="cursor-pointer ml-2 text-gray-600"
                  @click="goPkgs"
                >
                  <document-add />
                </el-icon>
              </span>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="$t('tools.autoCode.page.bizDbLabel')" prop="businessDB" class="w-full">
              <template #label>
                <el-tooltip
                  :content="$t('tools.autoCode.page.ttBizDbCodegen')"
                  placement="bottom"
                  effect="light"
                >
                  <div>
                    {{ $t('tools.autoCode.page.bizDbLabel') }} <el-icon><QuestionFilled /></el-icon>
                  </div>
                </el-tooltip>
              </template>
              <el-select
                v-model="form.businessDB"
                clearable
                :placeholder="$t('tools.autoCode.page.phBizDb')"
                class="w-full"
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
                    <span
                      style="float: right; color: #8492a6; font-size: 13px"
                      >{{ item.dbName }}</span
                    >
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>
    <div class="lrag-search-box">
      <el-collapse class="no-border-collapse">
        <el-collapse-item>
          <template #title>
            <div class="text-lg text-gray-600 font-normal">
              {{ $t('tools.autoCode.page.expertMode') }}
            </div>
          </template>
          <template #icon="{ isActive }">
          <span class="text-lg ml-auto mr-4 font-normal">
            {{ isActive ? $t('tools.autoCode.page.collapse') : $t('tools.autoCode.page.expand') }}
          </span>
          </template>
          <div class="p-4">
            <!-- 基础设置组 -->
            <div class="border-b border-gray-200 last:border-0">
              <h3 class="text-lg font-medium mb-4 text-gray-700">{{ $t('tools.autoCode.page.basicSettings') }}</h3>
              <el-row :gutter="20">
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttLragModel')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.useLragModel')">
                      <el-checkbox v-model="form.lragModel" @change="useLrag" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttBtnAuth')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.createBtnAuth')">
                      <el-checkbox :disabled="!form.generateWeb" v-model="form.autoCreateBtnAuth" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
                <el-col :span="3">
                  <el-form-item :label="$t('tools.autoCode.page.genWeb')">
                    <el-checkbox v-model="form.generateWeb" />
                  </el-form-item>
                </el-col>
                <el-col :span="3">
                  <el-form-item :label="$t('tools.autoCode.page.genServer')">
                    <el-checkbox disabled v-model="form.generateServer" />
                  </el-form-item>
                </el-col>
              </el-row>
            </div>

            <!-- 自动化设置组 -->
            <div class="border-b border-gray-200 last:border-0">
              <h3 class="text-lg font-medium mb-4 text-gray-700">{{ $t('tools.autoCode.page.autoSettings') }}</h3>
              <el-row :gutter="20">
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttAutoApi')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.autoCreateApi')">
                      <el-checkbox  :disabled="!form.generateServer" v-model="form.autoCreateApiToSql" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttAutoMenu')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.autoCreateMenu')">
                      <el-checkbox :disabled="!form.generateWeb" v-model="form.autoCreateMenuToSql" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttAutoMigrate')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.autoMigrate')">
                      <el-checkbox  :disabled="!form.generateServer" v-model="form.autoMigrate" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
              </el-row>
            </div>

            <!-- 高级设置组 -->
            <div class="border-b border-gray-200 last:border-0">
              <h3 class="text-lg font-medium mb-4 text-gray-700">{{ $t('tools.autoCode.page.advancedSettings') }}</h3>
              <el-row :gutter="20">
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttResource')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.createResource')">
                      <el-checkbox v-model="form.autoCreateResource" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
                <el-col :span="3">
                  <el-tooltip
                      :content="$t('tools.autoCode.page.ttOnlyTpl')"
                      placement="top"
                      effect="light"
                  >
                    <el-form-item :label="$t('tools.autoCode.page.onlyTemplate')">
                      <el-checkbox v-model="form.onlyTemplate" />
                    </el-form-item>
                  </el-tooltip>
                </el-col>
              </el-row>
            </div>

            <!-- 树形结构设置 -->
            <div class="last:pb-0">
              <h3 class="text-lg font-medium mb-4 text-gray-700">{{ $t('tools.autoCode.page.treeSettings') }}</h3>
              <el-row :gutter="20" align="middle">
                <el-col :span="24">
                    <el-form-item :label="$t('tools.autoCode.page.treeStruct')">
                      <div class="flex items-center gap-4">
                        <el-tooltip
                            :content="$t('tools.autoCode.page.ttTree')"
                            placement="top"
                            effect="light"
                        >
                          <el-checkbox v-model="form.isTree" />
                        </el-tooltip>
                        <el-input
                            v-model="form.treeJson"
                            :disabled="!form.isTree"
                            :placeholder="$t('tools.autoCode.page.phTreeJson')"
                            class="flex-1"
                        />
                      </div>
                    </el-form-item>
                </el-col>
              </el-row>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>
    <!-- 组件列表 -->
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
          <el-button
          type="primary"
          @click="editAndAddField()"
          :disabled="form.onlyTemplate"
        >
          {{ $t('tools.autoCode.fields.btnAddField') }}
        </el-button>
      </div>
      <div class="draggable">
        <el-table :data="form.fields" row-key="fieldName">
          <el-table-column
            v-if="!isAdd"
            fixed="left"
            align="left"
            type="index"
            width="60"
          >
            <template #default>
              <el-icon class="cursor-grab drag-column">
                <MoreFilled />
              </el-icon>
            </template>
          </el-table-column>
          <el-table-column
            fixed="left"
            align="left"
            type="index"
            :label="$t('tools.autoCode.fields.colSeq')"
            width="60"
          />
          <el-table-column
            fixed="left"
            align="left"
            type="index"
            :label="$t('tools.autoCode.fields.colPk')"
            width="60"
          >
            <template #default="{ row }">
              <el-checkbox :disabled="row.disabled" v-model="row.primaryKey" />
            </template>
          </el-table-column>
          <el-table-column
            fixed="left"
            align="left"
            prop="fieldName"
            :label="$t('tools.autoCode.fields.colFieldName')"
            width="160"
          >
            <template #default="{ row }">
              <el-input disabled v-model="row.fieldName" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="fieldDesc"
            :label="$t('tools.autoCode.fields.colDisplayName')"
            width="160"
          >
            <template #default="{ row }">
              <el-input :disabled="row.disabled" v-model="row.fieldDesc" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="defaultValue"
            :label="$t('tools.autoCode.fields.colDefaultValue')"
            width="160"
          >
            <template #default="{ row }">
              <el-input :disabled="row.disabled" v-model="row.defaultValue" />
            </template>
          </el-table-column>
          <el-table-column align="left" prop="require" :label="$t('tools.autoCode.fields.colRequired')">
            <template #default="{ row }">
              <el-checkbox :disabled="row.disabled" v-model="row.require" />
            </template>
          </el-table-column>
          <el-table-column align="left" prop="sort" :label="$t('tools.autoCode.fields.colSort')">
            <template #default="{ row }">
              <el-checkbox :disabled="row.disabled" v-model="row.sort" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="form"
            width="100"
            :label="$t('tools.autoCode.fields.colFormEdit')"
          >
            <template #default="{ row }">
              <el-checkbox :disabled="row.disabled" v-model="row.form" />
            </template>
          </el-table-column>
          <el-table-column align="left" prop="table" :label="$t('tools.autoCode.fields.colTable')">
            <template #default="{ row }">
              <el-checkbox :disabled="row.disabled" v-model="row.table" />
            </template>
          </el-table-column>
          <el-table-column align="left" prop="desc" :label="$t('tools.autoCode.fields.colDetail')">
            <template #default="{ row }">
              <el-checkbox :disabled="row.disabled" v-model="row.desc" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="excel"
            width="100"
            :label="$t('tools.autoCode.fields.colImportExport')"
            v-if="!isAdd"
          >
            <template #default="{ row }">
              <el-checkbox v-model="row.excel" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="fieldJson"
            width="160px"
            :label="$t('tools.autoCode.fields.colFieldJson')"
          >
            <template #default="{ row }">
              <el-input :disabled="row.disabled" v-model="row.fieldJson" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="fieldType"
            :label="$t('tools.autoCode.fields.colFieldType')"
            width="160"
          >
            <template #default="{ row }">
              <el-select
                v-model="row.fieldType"
                style="width: 100%"
                :placeholder="$t('tools.autoCode.fields.phFieldType')"
                :disabled="row.disabled"
                clearable
              >
                <el-option
                  v-for="item in typeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="fieldIndexType"
            :label="$t('tools.autoCode.fields.colIndexType')"
            width="160"
          >
            <template #default="{ row }">
              <el-select
                v-model="row.fieldIndexType"
                style="width: 100%"
                :placeholder="$t('tools.autoCode.fields.phIndexType')"
                :disabled="row.disabled"
                clearable
              >
                <el-option
                  v-for="item in typeIndexOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="dataTypeLong"
            :label="$t('tools.autoCode.fields.colFieldLenEnum')"
            width="160"
          >
            <template #default="{ row }">
              <el-input :disabled="row.disabled" v-model="row.dataTypeLong" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="columnName"
            :label="$t('tools.autoCode.fields.colDbColumn')"
            width="160"
          >
            <template #default="{ row }">
              <el-input :disabled="row.disabled" v-model="row.columnName" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="comment"
            :label="$t('tools.autoCode.fields.colDbComment')"
            width="160"
          >
            <template #default="{ row }">
              <el-input :disabled="row.disabled" v-model="row.comment" />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            prop="fieldSearchType"
            :label="$t('tools.autoCode.fields.colSearchCond')"
            width="130"
          >
            <template #default="{ row }">
              <el-select
                v-model="row.fieldSearchType"
                style="width: 100%"
                :placeholder="$t('tools.autoCode.fields.phSearchOp')"
                clearable
                :disabled="row.fieldType === 'json' || row.disabled"
              >
                <el-option
                  v-for="item in typeSearchOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                  :disabled="canSelect(row.fieldType,item.value)"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column align="left" :label="$t('common.colActions')" width="300" fixed="right">
            <template #default="scope">
              <el-button
                v-if="!scope.row.disabled"
                type="primary"
                link
                icon="edit"
                @click="editAndAddField(scope.row)"
              >
                {{ $t('tools.autoCode.fields.btnAdvancedEdit') }}
              </el-button>
              <el-button
                v-if="!scope.row.disabled"
                type="primary"
                link
                icon="delete"
                @click="deleteField(scope.$index)"
              >
                {{ $t('tools.autoCode.fields.btnDelete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <!-- 组件列表 -->
      <div class="lrag-btn-list justify-end mt-4">
        <el-button type="primary" :disabled="isAdd" @click="exportJson()">
          {{ $t('tools.autoCode.fields.btnExportJson') }}
        </el-button>
        <el-upload
          class="flex items-center"
          :before-upload="importJson"
          :show-file-list="false"
          :headers="{'x-token': token}"
          accept=".json"
        >
          <el-button type="primary" class="mx-2" :disabled="isAdd"
            >{{ $t('tools.autoCode.fields.btnImportJson') }}</el-button
          >
        </el-upload>
        <el-button type="primary" :disabled="isAdd" @click="clearCatch()">
          {{ $t('tools.autoCode.page.clearDraft') }}
        </el-button>
        <el-button type="primary" :disabled="isAdd" @click="catchData()">
          {{ $t('tools.autoCode.page.saveDraft') }}
        </el-button>
        <el-button type="primary" :disabled="isAdd" @click="enterForm(false)">
          {{ $t('tools.autoCode.page.genCode') }}
        </el-button>
        <el-button type="primary" @click="enterForm(true)">
          {{ isAdd ? $t('tools.autoCode.page.viewCode') : $t('tools.autoCode.page.previewCode') }}
        </el-button>
      </div>
    </div>
    <!-- 组件弹窗 -->
    <el-drawer v-model="dialogFlag" size="70%" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('tools.autoCode.page.drawerComponent') }}</span>
          <div>
            <el-button @click="closeDialog">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="enterDialog">{{ $t('common.ok') }}</el-button>
          </div>
        </div>
      </template>

      <FieldDialog
        v-if="dialogFlag"
        ref="fieldDialogNode"
        :dialog-middle="dialogMiddle"
        :type-options="typeOptions"
        :type-search-options="typeSearchOptions"
        :type-index-options="typeIndexOptions"
      />
    </el-drawer>

    <el-drawer v-model="previewFlag" size="80%" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('tools.autoCode.page.drawerOps') }}</span>
          <div>
            <el-button type="primary" @click="selectText">{{ $t('tools.autoCode.page.selectAll') }}</el-button>
            <el-button type="primary" @click="copy">{{ $t('tools.autoCode.page.copy') }}</el-button>
          </div>
        </div>
      </template>
      <PreviewCodeDialog
        v-if="previewFlag"
        :is-add="isAdd"
        ref="previewNode"
        :preview-code="preViewCode"
      />
    </el-drawer>
  </div>
</template>

<script setup>
  import FieldDialog from '@/view/systemTools/autoCode/component/fieldDialog.vue'
  import PreviewCodeDialog from '@/view/systemTools/autoCode/component/previewCodeDialog.vue'
  import {
    toUpperCase,
    toHump,
    toSQLLine,
    toLowerCase
  } from '@/utils/stringFun'
  import {
    createTemp,
    getDB,
    getTable,
    getColumn,
    preview,
    getMeta,
    getPackageApi,
    llmAuto
  } from '@/api/autoCode'
  import { getDict } from '@/utils/dictionary'
  import { ref, watch, toRaw, onMounted, nextTick, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import Sortable from 'sortablejs'
  import { useUserStore } from "@/pinia";

  const userStore = useUserStore()
  const { t } = useI18n()

  const token = userStore.token

  const handleFocus = () => {
    document.addEventListener('keydown', handleKeydown);
    document.addEventListener('paste', handlePaste);
  }

  const handleBlur = () => {
    document.removeEventListener('keydown', handleKeydown);
    document.removeEventListener('paste', handlePaste);
  }

  const handleKeydown = (event) => {
    if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
      llmAutoFunc()
    }
  }

  const handlePaste = (event) => {
    const items = event.clipboardData.items;
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf('image') !== -1) {
        const file = items[i].getAsFile();
        const reader = new FileReader();
        reader.onload =async (e) => {
          const base64String = e.target.result;
          const res = await llmAuto({ _file_path: base64String,mode:"eye" })
          if (res.code === 0) {
            prompt.value = res.data.text
            llmAutoFunc()
          }
        };
        reader.readAsDataURL(file);
      }
    }
  };

  const getOnlyNumber = () => {
    let randomNumber = ''
    while (randomNumber.length < 16) {
      randomNumber += Math.random().toString(16).substring(2)
    }
    return randomNumber.substring(0, 16)
  }

  const prompt = ref('')

  const eyeFunc = async () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';

    input.onchange = (event) => {
      const file = event.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = async (e) => {
          const base64String = e.target.result;

          const res = await llmAuto({ _file_path: base64String,mode:'eye' })
          if (res.code === 0) {
            prompt.value = res.data.text
            llmAutoFunc()
          }
        };
        reader.readAsDataURL(file);
      }
    };

    input.click();
  }


  const llmAutoFunc = async (flag) => {
    if (flag && !form.value.structName) {
      ElMessage.error(t('tools.autoCode.page.errStructName'))
      return
    }
    if (!flag && !prompt.value) {
      ElMessage.error(t('tools.autoCode.page.errPrompt'))
      return
    }

    if (form.value.fields.length > 0) {
      const res = await ElMessageBox.confirm(
        t('tools.autoCode.page.confirmAiClear'),
        t('common.tipTitle'),
        {
          confirmButtonText: t('common.ok'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
      if (res !== 'confirm') {
        return
      }
    }

    const res = await llmAuto({
      prompt: flag
        ? t('tools.autoCode.page.aiPromptStructPrefix', { name: form.value.structName })
        : prompt.value,
      mode: "ai"
    })
    if (res.code === 0) {
      form.value.fields = []
      const json = JSON.parse(res.data.text)
      json.fields?.forEach((item) => {
        item.fieldName = toUpperCase(item.fieldName)
      })

      for (let key in json) {
        form.value[key] = json[key]
      }

      form.value.generateServer = true
      form.value.generateWeb = true

    }
  }

  const isAdd = ref(false)

  // 行拖拽
  const rowDrop = () => {
    // 要拖拽元素的父容器
    const tbody = document.querySelector(
      '.draggable .el-table__body-wrapper tbody'
    )
    Sortable.create(tbody, {
      //  可被拖拽的子元素
      draggable: '.draggable .el-table__row',
      handle: '.drag-column',
      onEnd: async ({ newIndex, oldIndex }) => {
        await nextTick()
        const currRow = form.value.fields.splice(oldIndex, 1)[0]
        form.value.fields.splice(newIndex, 0, currRow)
      }
    })
  }

  onMounted(() => {
    rowDrop()
  })

  defineOptions({
    name: 'AutoCode'
  })
  const gormModelList = ['id', 'created_at', 'updated_at', 'deleted_at']

  const dataModelList = ['created_by', 'updated_by', 'deleted_by']

  const typeOptions = computed(() => {
    const ft = (key) => ({
      label: t(`tools.autoCode.fieldTypes.${key}`),
      value: key
    })
    return [
      ft('string'),
      ft('richtext'),
      ft('int'),
      ft('bool'),
      ft('float64'),
      ft('time.Time'),
      ft('enum'),
      ft('picture'),
      ft('pictures'),
      ft('video'),
      ft('file'),
      ft('json'),
      ft('array')
    ]
  })

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
    }
  ])

  const typeIndexOptions = ref([
    {
      label: 'index',
      value: 'index'
    },
    {
      label: 'uniqueIndex',
      value: 'uniqueIndex'
    }
  ])

  const fieldTemplate = {
    fieldName: '',
    fieldDesc: '',
    fieldType: '',
    dataType: '',
    fieldJson: '',
    columnName: '',
    dataTypeLong: '',
    comment: '',
    defaultValue: '',
    require: false,
    sort: false,
    form: true,
    desc: true,
    table: true,
    excel: false,
    errorText: '',
    primaryKey: false,
    clearable: true,
    fieldSearchType: '',
    fieldIndexType: '',
    dictType: '',
    dataSource: {
      dbName: '',
      association: 1,
      table: '',
      label: '',
      value: '',
      hasDeletedAt: false
    }
  }
  const route = useRoute()
  const router = useRouter()
  const preViewCode = ref({})
  const dbform = ref({
    businessDB: '',
    dbName: '',
    tableName: ''
  })
  const tableOptions = ref([])
  const addFlag = ref('')
  const fdMap = ref({})
  const form = ref({
    structName: '',
    tableName: '',
    packageName: '',
    package: '',
    abbreviation: '',
    description: '',
    businessDB: '',
    autoCreateApiToSql: true,
    autoCreateMenuToSql: true,
    autoCreateBtnAuth: false,
    autoMigrate: true,
    lragModel: true,
    autoCreateResource: false,
    onlyTemplate: false,
    isTree: false,
    generateWeb:true,
    generateServer:true,
    treeJson: "",
    fields: []
  })
  const rules = computed(() => ({
    structName: [
      { required: true, message: t('tools.autoCode.page.ruleStructName'), trigger: 'blur' }
    ],
    abbreviation: [
      { required: true, message: t('tools.autoCode.page.ruleAbbr'), trigger: 'blur' }
    ],
    description: [
      { required: true, message: t('tools.autoCode.page.ruleDesc'), trigger: 'blur' }
    ],
    packageName: [
      {
        required: true,
        message: t('tools.autoCode.page.rulePkgName'),
        trigger: 'blur'
      }
    ],
    package: [{ required: true, message: t('tools.autoCode.page.rulePackage'), trigger: 'blur' }]
  }))
  const dialogMiddle = ref({})
  const bk = ref({})
  const dialogFlag = ref(false)
  const previewFlag = ref(false)

  const useLrag = (e) => {
    if (e && form.value.fields.length) {
      ElMessageBox.confirm(
        t('tools.autoCode.page.confirmLragModel'),
        t('tools.autoCode.page.warnTitle'),
        {
          confirmButtonText: t('tools.autoCode.fieldDialog.btnContinue'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
        .then(() => {
          form.value.fields = form.value.fields.filter(
            (item) =>
              !gormModelList.some((gormfd) => gormfd === item.columnName)
          )
        })
        .catch(() => {
          form.value.lragModel = false
        })
    }
  }

  const toLowerCaseFunc = (form, key) => {
    form[key] = toLowerCase(form[key])
  }
  const previewNode = ref(null)
  const selectText = () => {
    previewNode.value.selectText()
  }
  const copy = () => {
    previewNode.value.copy()
  }
  const editAndAddField = (item) => {
    dialogFlag.value = true
    if (item) {
      addFlag.value = 'edit'
      if (!item.dataSource) {
        item.dataSource = {
          dbName: '',
          association: 1,
          table: '',
          label: '',
          value: '',
          hasDeletedAt: false
        }
      }
      bk.value = JSON.parse(JSON.stringify(item))
      dialogMiddle.value = item
    } else {
      addFlag.value = 'add'
      fieldTemplate.onlyNumber = getOnlyNumber()
      dialogMiddle.value = JSON.parse(JSON.stringify(fieldTemplate))
    }
  }

  const fieldDialogNode = ref(null)
  const enterDialog = () => {
    fieldDialogNode.value.fieldDialogForm.validate((valid) => {
      if (valid) {
        dialogMiddle.value.fieldName = toUpperCase(dialogMiddle.value.fieldName)
        if (addFlag.value === 'add') {
          form.value.fields.push(dialogMiddle.value)
        }
        dialogFlag.value = false
      } else {
        return false
      }
    })
  }
  const closeDialog = () => {
    if (addFlag.value === 'edit') {
      dialogMiddle.value = bk.value
    }
    dialogFlag.value = false
  }
  const deleteField = (index) => {
    ElMessageBox.confirm(t('tools.autoCode.page.confirmDeleteField'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      form.value.fields.splice(index, 1)
    })
  }
  const autoCodeForm = ref(null)
  const enterForm = async (isPreview) => {
    if (form.value.isTree && !form.value.treeJson){
      ElMessage({
        type: 'error',
        message: t('tools.autoCode.page.errTreeJson')
      })
      return false
    }
    if(!form.value.generateWeb && !form.value.generateServer){
      ElMessage({
        type: 'error',
        message: t('tools.autoCode.page.errPickGen')
      })
      return false
    }
    if (!form.value.onlyTemplate) {
      if (form.value.fields.length <= 0) {
        ElMessage({
          type: 'error',
          message: t('tools.autoCode.page.errOneField')
        })
        return false
      }

      if (
        !form.value.lragModel &&
        form.value.fields.every((item) => !item.primaryKey)
      ) {
        ElMessage({
          type: 'error',
          message: t('tools.autoCode.page.errPkRequired')
        })
        return false
      }

      if (
        form.value.fields.some(
          (item) => item.fieldName === form.value.structName
        )
      ) {
        ElMessage({
          type: 'error',
          message: t('tools.autoCode.page.errFieldNameConflict')
        })
        return false
      }

      if (
        form.value.fields.some((item) => item.fieldJson === form.value.package)
      ) {
        ElMessage({
          type: 'error',
          message: t('tools.autoCode.page.errJsonConflict')
        })
        return false
      }

      if (form.value.fields.some((item) => !item.fieldType)) {
        ElMessage({
          type: 'error',
          message: t('tools.autoCode.page.errAllFieldTypes')
        })
        return false
      }

      if (form.value.package === form.value.abbreviation) {
        ElMessage({
          type: 'error',
          message: t('tools.autoCode.page.errPkgAbbrSame')
        })
        return false
      }
    }

    autoCodeForm.value.validate(async (valid) => {
      if (valid) {
        for (const key in form.value) {
          if (typeof form.value[key] === 'string') {
            form.value[key] = form.value[key].trim()
          }
        }
        form.value.structName = toUpperCase(form.value.structName)
        form.value.tableName = form.value.tableName.replace(' ', '')
        if (!form.value.tableName) {
          form.value.tableName = toSQLLine(toLowerCase(form.value.structName))
        }
        if (form.value.structName === form.value.abbreviation) {
          ElMessage({
            type: 'error',
            message: t('tools.autoCode.page.errStructAbbrSame')
          })
          return false
        }
        form.value.humpPackageName = toSQLLine(form.value.packageName)

        form.value.fields?.forEach((item) => {
          item.fieldName = toUpperCase(item.fieldName)
          if (item.fieldType === 'enum') {
            // 判断一下 item.dataTypeLong 按照,切割后的每个元素是否都使用 '' 包裹，如果没包 则修改为包裹起来的 然后再转为字符串赋值给 item.dataTypeLong
            item.dataTypeLong = item.dataTypeLong.replace(/[\[\]{}()]/g, '')
            const arr = item.dataTypeLong.split(',')
            arr.forEach((ele, index) => {
              if (ele.indexOf("'") === -1) {
                arr[index] = `'${ele}'`
              }
            })
            item.dataTypeLong = arr.join(',')
          }
        })

        delete form.value.primaryField
        if (isPreview) {
          const res = await preview({
            ...form.value,
            isAdd: !!isAdd.value,
            fields: form.value.fields.filter((item) => !item.disabled)
          })
          if(res.code !== 0){
            return
          }
          preViewCode.value = res.data.autoCode
          previewFlag.value = true
        } else {
          const res = await createTemp(form.value)
          if (res.code !== 0) {
            return
          }
          ElMessage({
            type: 'success',
            message: t('tools.autoCode.page.successCreated')
          })
          clearCatch()
        }
      }
    })
  }

  const dbList = ref([])
  const dbOptions = ref([])

  const getDbFunc = async () => {
    dbform.value.dbName = ''
    dbform.value.tableName = ''
    const res = await getDB({ businessDB: dbform.value.businessDB })
    if (res.code === 0) {
      dbOptions.value = res.data.dbs
      dbList.value = res.data.dbList
    }
  }
  const getTableFunc = async () => {
    const res = await getTable({
      businessDB: dbform.value.businessDB,
      dbName: dbform.value.dbName
    })
    if (res.code === 0) {
      tableOptions.value = res.data.tables
    }
    dbform.value.tableName = ''
  }

  const getColumnFunc = async () => {
    const res = await getColumn(dbform.value)
    if (res.code === 0) {
      let dbtype = ''
      if (dbform.value.businessDB !== '') {
        const dbtmp = dbList.value.find(
          (item) => item.aliasName === dbform.value.businessDB
        )
        const dbraw = toRaw(dbtmp)
        dbtype = dbraw.dbtype
      }
      form.value.lragModel = false
      const tbHump = toHump(dbform.value.tableName)
      form.value.structName = toUpperCase(tbHump)
      form.value.tableName = dbform.value.tableName
      form.value.packageName = toLowerCase(tbHump)
      form.value.abbreviation = toLowerCase(tbHump)
      form.value.description = tbHump + t('tools.autoCode.page.tableSuffix')
      form.value.autoCreateApiToSql = true
      form.value.generateServer = true
      form.value.generateWeb = true
      form.value.fields = []
      res.data.columns &&
        res.data.columns.forEach((item) => {
          if (needAppend(item)) {
            const fbHump = toHump(item.columnName)
            form.value.fields.push({
              onlyNumber: getOnlyNumber(),
              fieldName: toUpperCase(fbHump),
              fieldDesc: item.columnComment || fbHump + t('tools.autoCode.page.fieldSuffix'),
              fieldType: fdMap.value[item.dataType],
              dataType: item.dataType,
              fieldJson: fbHump,
              primaryKey: item.primaryKey,
              dataTypeLong:
                item.dataTypeLong && item.dataTypeLong.split(',')[0],
              columnName:
                dbtype === 'oracle'
                  ? item.columnName.toUpperCase()
                  : item.columnName,
              comment: item.columnComment,
              require: false,
              errorText: '',
              clearable: true,
              fieldSearchType: '',
              fieldIndexType: '',
              dictType: '',
              form: true,
              table: true,
              excel: false,
              desc: true,
              dataSource: {
                dbName: '',
                association: 1,
                table: '',
                label: '',
                value: '',
                hasDeletedAt: false
              }
            })
          }
        })
    }
  }

  const needAppend = (item) => {
    let isAppend = true
    if (
      form.value.lragModel &&
      gormModelList.some((gormfd) => gormfd === item.columnName)
    ) {
      isAppend = false
    }
    if (
      form.value.autoCreateResource &&
      dataModelList.some((datafd) => datafd === item.columnName)
    ) {
      isAppend = false
    }
    return isAppend
  }

  const setFdMap = async () => {
    const fdTypes = ['string', 'int', 'bool', 'float64', 'time.Time']
    fdTypes.forEach(async (fdtype) => {
      const res = await getDict(fdtype)
      res &&
        res.forEach((item) => {
          fdMap.value[item.label] = fdtype
        })
    })
  }
  const getAutoCodeJson = async (id) => {
    const res = await getMeta({ id: Number(id) })
    if (res.code === 0) {
      const add = route.query.isAdd
      isAdd.value = add
      form.value = JSON.parse(res.data.meta)
      if (isAdd.value) {
        form.value.fields.forEach((item) => {
          item.disabled = true
        })
      }
    }
  }

  const pkgs = ref([])
  const getPkgs = async () => {
    const res = await getPackageApi()
    if (res.code === 0) {
      pkgs.value = res.data.pkgs
    }
  }

  const goPkgs = () => {
    router.push({ name: 'autoPkg' })
  }

  const init = () => {
    getDbFunc()
    setFdMap()
    getPkgs()
    const id = route.params.id
    if (id) {
      getAutoCodeJson(id)
    }
  }
  init()

  watch(()=>form.value.generateServer,()=>{
    if(!form.value.generateServer){
      form.value.autoCreateApiToSql = false
      form.value.autoMigrate = false
    }
  })

  watch(()=>form.value.generateWeb,()=>{
    if(!form.value.generateWeb){
      form.value.autoCreateMenuToSql = false
      form.value.autoCreateBtnAuth = false
    }
  })

  const catchData = () => {
    window.sessionStorage.setItem('autoCode', JSON.stringify(form.value))
    ElMessage.success(t('tools.autoCode.page.draftSaved'))
  }

  const getCatch = () => {
    const data = window.sessionStorage.getItem('autoCode')
    if (data) {
      form.value = JSON.parse(data)
    }
  }

  const clearCatch = async () => {
    form.value = {
      structName: '',
      tableName: '',
      packageName: '',
      package: '',
      abbreviation: '',
      description: '',
      businessDB: '',
      autoCreateApiToSql: true,
      autoCreateMenuToSql: true,
      autoCreateBtnAuth: false,
      autoMigrate: true,
      lragModel: true,
      autoCreateResource: false,
      onlyTemplate: false,
      isTree: false,
      treeJson: "",
      fields: []
    }
    await nextTick()
    window.sessionStorage.removeItem('autoCode')
  }

  getCatch()

  const exportJson = () => {
    const dataStr = JSON.stringify(form.value, null, 2)
    const blob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'form_data.json'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const importJson = (file) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        form.value = JSON.parse(e.target.result)
        form.value.generateServer = true
        form.value.generateWeb = true
        ElMessage.success(t('tools.autoCode.page.jsonImportOk'))
      } catch (_) {
        ElMessage.error(t('tools.autoCode.page.jsonImportBad'))
      }
    }
    reader.readAsText(file)
    return false
  }

  watch(
    () => form.value.onlyTemplate,
    (val) => {
      if (val) {
        ElMessageBox.confirm(
          t('tools.autoCode.page.confirmOnlyTemplate'),
          t('tools.autoCode.page.warnTitle'),
          {
            confirmButtonText: t('tools.autoCode.fieldDialog.btnContinue'),
            cancelButtonText: t('common.cancel'),
            type: 'warning'
          }
        )
          .then(() => {
            form.value.fields = []
          })
          .catch(() => {
            form.value.onlyTemplate = false
          })
      }
    }
  )

  const canSelect = (fieldType,item) => {
    if (fieldType === 'richtext') {
      return item !== 'LIKE';
    }

    if (fieldType !== 'string' && item === 'LIKE') {
      return true;
    }

    const nonNumericTypes = ['int', 'time.Time', 'float64'];
    if (!nonNumericTypes.includes(fieldType) && ['BETWEEN', 'NOT BETWEEN'].includes(item)) {
      return true;
    }

    return false;
  }
</script>

<style>
.no-border-collapse{
  @apply border-none;
  .el-collapse-item__header{
    @apply border-none;
  }
  .el-collapse-item__wrap{
    @apply border-none;
  }
  .el-collapse-item__content{
    @apply pb-0;
  }
}
</style>
