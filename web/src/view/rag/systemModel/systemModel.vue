<template>
  <div>
    <warning-bar :title="$t('rag.sysModel.warningBar')" />
    <div class="lrag-table-box">
      <el-tabs v-model="activeTab">
        <!-- Tab 1: 管理员模型管理 -->
        <el-tab-pane :label="$t('rag.sysModel.tabModels')" name="models">
          <div class="lrag-btn-list mb-4">
            <el-button type="primary" icon="plus" @click="openAdd">{{ $t('rag.sysModel.btnAddModel') }}</el-button>
          </div>
          <el-table :data="adminModels" style="width: 100%">
            <el-table-column align="left" :label="$t('rag.sysModel.colId')" prop="ID" width="60" />
            <el-table-column align="left" :label="$t('rag.sysModel.colProvider')" prop="name" width="120" />
            <el-table-column align="left" :label="$t('rag.sysModel.colModelName')" prop="modelName" min-width="160" />
            <el-table-column align="left" :label="$t('rag.sysModel.colScenarios')" min-width="240">
              <template #default="scope">
                <el-tag v-for="mt in (scope.row.modelTypes || ['chat'])" :key="mt" size="small" class="mr-1">
                  {{ modelTypeLabel(mt) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.userModel.colBaseUrl')" prop="baseUrl" min-width="180" show-overflow-tooltip />
            <el-table-column align="left" :label="$t('rag.sysModel.colStatus')" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.enabled ? 'success' : 'danger'" size="small">
                  {{ scope.row.enabled ? $t('rag.sysModel.enabled') : $t('rag.sysModel.disabled') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.sysModel.colShareScope')" width="100">
              <template #default="scope">
                {{ shareScopeLabel(scope.row.shareScope) }}
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.docs.colActions')" width="150" fixed="right">
              <template #default="scope">
                <el-button type="primary" link icon="edit" @click="openEdit(scope.row)">{{ $t('rag.sysModel.edit') }}</el-button>
                <el-button type="danger" link icon="delete" @click="deleteModel(scope.row)">{{ $t('rag.sysModel.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 2: 系统默认模型 -->
        <el-tab-pane :label="$t('rag.sysModel.tabDefaults')" name="defaults">
          <div class="text-gray-500 text-sm mb-4">
            {{ $t('rag.sysModel.defaultsHint') }}
          </div>
          <el-form label-width="140px" class="max-w-lg">
            <el-form-item v-for="opt in modelTypeOptions" :key="opt.value" :label="opt.label">
              <div class="flex items-center w-full gap-2">
                <el-select
                  v-model="systemDefaults[opt.value]"
                  :placeholder="$t('rag.sysModel.pickDefaultModel')"
                  clearable
                  style="flex: 1"
                  @change="(v) => saveSystemDefault(opt.value, v)"
                >
                  <el-option
                    v-for="m in modelsForType(t.value)"
                    :key="m.ID"
                    :label="`${m.name} / ${m.modelName}`"
                    :value="m.ID"
                  />
                </el-select>
              </div>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 3: 全局共享知识库 -->
        <el-tab-pane :label="$t('rag.sysModel.tabGlobalKb')" name="globalKb">
          <div class="text-gray-500 text-sm mb-4">
            {{ $t('rag.sysModel.globalKbHint') }}
          </div>
          <div class="lrag-btn-list mb-4">
            <el-select
              v-model="globalKbAddId"
              :placeholder="$t('rag.sysModel.pickKb')"
              filterable
              style="width: 300px"
              class="mr-2"
            >
              <el-option
                v-for="kb in availableKbsForGlobal"
                :key="kb.id"
                :label="kb.name"
                :value="kb.id"
              />
            </el-select>
            <el-button type="primary" icon="plus" :disabled="!globalKbAddId" @click="addGlobalKb">
              {{ $t('rag.sysModel.addGlobalKb') }}
            </el-button>
          </div>
          <el-table :data="globalKbList" style="width: 100%" :empty-text="$t('rag.sysModel.emptyGlobalKb')">
            <el-table-column align="left" :label="$t('rag.sysModel.colId')" prop="knowledgeBaseId" width="80" />
            <el-table-column align="left" :label="$t('rag.sysModel.colKbName')" prop="knowledgeBaseName" min-width="200" />
            <el-table-column align="left" :label="$t('rag.sysModel.colRemark')" prop="description" min-width="200" show-overflow-tooltip />
            <el-table-column align="left" :label="$t('rag.docs.colActions')" width="120" fixed="right">
              <template #default="scope">
                <el-button type="danger" link icon="delete" @click="removeGlobalKb(scope.row)">
                  {{ $t('rag.sysModel.remove') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 4: 系统默认互联网搜索配置 -->
        <el-tab-pane :label="$t('rag.sysModel.tabWebSearch')" name="webSearch">
          <div class="text-gray-500 text-sm mb-4">
            {{ $t('rag.sysModel.webSearchHint') }}
          </div>
          <el-form label-width="140px" class="max-w-lg">
            <el-form-item :label="$t('rag.sysModel.defaultEngine')">
              <el-select v-model="sysWebSearchForm.provider" :placeholder="$t('rag.sysModel.pickEngine')" style="width: 100%">
                <el-option
                  v-for="p in sysWebSearchProviders"
                  :key="p.id"
                  :label="p.displayName"
                  :value="p.id"
                />
              </el-select>
            </el-form-item>
            <template v-for="f in sysCurrentWebSearchSchema" :key="f.key">
              <el-form-item :label="f.label" :required="f.required">
                <el-input
                  v-model="sysWebSearchForm.config[f.key]"
                  :type="f.secret ? 'password' : 'text'"
                  :placeholder="f.placeholder || ''"
                  show-password
                  clearable
                  style="width: 100%"
                />
              </el-form-item>
            </template>
            <el-form-item>
              <el-button type="primary" @click="saveSysWebSearchConfig">{{ $t('rag.docs.save') }}</el-button>
              <el-button @click="clearSysWebSearchConfig">{{ $t('rag.sysModel.clearConfig') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 向量存储（知识库向量索引） -->
        <el-tab-pane :label="$t('rag.sysModel.tabVector')" name="vectorStorage">
          <div class="text-gray-500 text-sm mb-4">
            {{ $t('rag.sysModel.vectorHint') }}
          </div>
          <div class="lrag-btn-list mb-4">
            <el-button type="primary" icon="plus" @click="openVectorDrawer()">{{ $t('rag.sysModel.btnAddVector') }}</el-button>
          </div>
          <el-table :data="vectorList" style="width: 100%">
            <el-table-column align="left" :label="$t('rag.sysModel.colName')" prop="name" width="160" />
            <el-table-column align="left" :label="$t('rag.sysModel.colType')" prop="provider" width="120" />
            <el-table-column align="left" :label="$t('rag.sysModel.colScope')" min-width="200">
              <template #default="scope">
                <span v-if="scope.row.allowAll !== false">{{ $t('rag.sysModel.allRoles') }}</span>
                <span v-else>{{ formatRagSettingRoleLabels(scope.row.allowedAuthorityIds) }}</span>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.sysModel.colStatus')" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
                  {{ scope.row.enabled ? $t('rag.sysModel.enabled') : $t('rag.sysModel.disabled') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.docs.colActions')" width="150" fixed="right">
              <template #default="scope">
                <el-button type="primary" link icon="edit" @click="openVectorDrawer(scope.row)">{{ $t('rag.sysModel.edit') }}</el-button>
                <el-button type="danger" link icon="delete" @click="deleteVector(scope.row)">{{ $t('rag.sysModel.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="lrag-pagination mt-4">
            <el-pagination
              :current-page="vectorPage"
              :page-size="vectorPageSize"
              :page-sizes="[10, 20, 50]"
              :total="vectorTotal"
              layout="total, sizes, prev, pager, next"
              @current-change="(v) => { vectorPage = v; loadVectorList() }"
              @size-change="(v) => { vectorPageSize = v; loadVectorList() }"
            />
          </div>
        </el-tab-pane>

        <!-- 文件存储（知识库文档文件落盘 / 对象存储） -->
        <el-tab-pane :label="$t('rag.sysModel.tabFile')" name="fileStorage">
          <div class="text-gray-500 text-sm mb-4">
            {{ $t('rag.sysModel.fileHint') }}
          </div>
          <div class="lrag-btn-list mb-4">
            <el-button type="primary" icon="plus" @click="openFileDrawer()">{{ $t('rag.sysModel.btnAddFile') }}</el-button>
          </div>
          <el-table :data="fileList" style="width: 100%">
            <el-table-column align="left" :label="$t('rag.sysModel.colName')" prop="name" width="160" />
            <el-table-column align="left" :label="$t('rag.sysModel.colType')" prop="provider" width="120" />
            <el-table-column align="left" :label="$t('rag.sysModel.colScope')" min-width="200">
              <template #default="scope">
                <span v-if="scope.row.allowAll !== false">{{ $t('rag.sysModel.allRoles') }}</span>
                <span v-else>{{ formatRagSettingRoleLabels(scope.row.allowedAuthorityIds) }}</span>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.sysModel.colStatus')" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
                  {{ scope.row.enabled ? $t('rag.sysModel.enabled') : $t('rag.sysModel.disabled') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column align="left" :label="$t('rag.docs.colActions')" width="150" fixed="right">
              <template #default="scope">
                <el-button type="primary" link icon="edit" @click="openFileDrawer(scope.row)">{{ $t('rag.sysModel.edit') }}</el-button>
                <el-button type="danger" link icon="delete" @click="deleteFile(scope.row)">{{ $t('rag.sysModel.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="lrag-pagination mt-4">
            <el-pagination
              :current-page="filePage"
              :page-size="filePageSize"
              :page-sizes="[10, 20, 50]"
              :total="fileTotal"
              layout="total, sizes, prev, pager, next"
              @current-change="(v) => { filePage = v; loadFileList() }"
              @size-change="(v) => { filePageSize = v; loadFileList() }"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 文件存储编辑抽屉 -->
    <el-drawer v-model="fileDrawerVisible" :title="fileEditId ? $t('rag.sysModel.fileDrawerEdit') : $t('rag.sysModel.fileDrawerAdd')" size="480px" destroy-on-close>
      <div style="padding: 0 16px">
        <el-form :model="fileForm" label-width="120px">
          <el-form-item :label="$t('rag.sysModel.labelName')" required>
            <el-input v-model="fileForm.name" :placeholder="$t('rag.sysModel.phConfigName')" />
          </el-form-item>
          <el-form-item :label="$t('rag.sysModel.labelType')" required>
            <el-select v-model="fileForm.provider" :placeholder="$t('rag.sysModel.phSelect')" style="width: 100%" :disabled="!!fileEditId">
              <el-option :label="$t('rag.sysModel.provLocal')" value="local" />
              <el-option :label="$t('rag.sysModel.provQiniu')" value="qiniu" />
              <el-option :label="$t('rag.sysModel.provTencentCos')" value="tencent-cos" />
              <el-option :label="$t('rag.sysModel.provAliyunOss')" value="aliyun-oss" />
              <el-option :label="$t('rag.sysModel.provHuaweiObs')" value="huawei-obs" />
              <el-option :label="$t('rag.sysModel.provAwsS3')" value="aws-s3" />
              <el-option :label="$t('rag.sysModel.provCfR2')" value="cloudflare-r2" />
              <el-option :label="$t('rag.sysModel.provMinio')" value="minio" />
            </el-select>
          </el-form-item>
          <template v-if="fileForm.provider === 'local'">
            <el-form-item :label="$t('rag.sysModel.labelStorePath')">
              <el-input v-model="fileForm.config.storePath" :placeholder="$t('rag.sysModel.phStorePath')" />
            </el-form-item>
          </template>
          <template v-else-if="fileForm.provider === 'minio'">
            <el-form-item :label="$t('rag.sysModel.labelEndpoint')">
              <el-input v-model="fileForm.config.endpoint" :placeholder="$t('rag.sysModel.phEndpoint')" />
            </el-form-item>
            <el-form-item :label="$t('rag.sysModel.labelAccessKey')">
              <el-input v-model="fileForm.config.accessKeyId" :placeholder="$t('rag.sysModel.phAccessKey')" />
            </el-form-item>
            <el-form-item :label="$t('rag.sysModel.labelSecretKey')">
              <el-input v-model="fileForm.config.accessKeySecret" type="password" :placeholder="$t('rag.sysModel.phSecretKey')" show-password />
            </el-form-item>
            <el-form-item :label="$t('rag.sysModel.labelBucket')">
              <el-input v-model="fileForm.config.bucketName" :placeholder="$t('rag.sysModel.phBucket')" />
            </el-form-item>
          </template>
          <el-form-item :label="$t('rag.sysModel.allowAllRoles')">
            <el-switch v-model="fileForm.allowAll" />
            <div class="text-gray-400 text-xs mt-1">{{ $t('rag.sysModel.allowAllHintFile') }}</div>
          </el-form-item>
          <el-form-item v-if="!fileForm.allowAll" :label="$t('rag.sysModel.allowedRoles')" required>
            <el-select
              v-model="fileForm.allowedAuthorityIds"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              :placeholder="$t('rag.sysModel.phRolesFile')"
              style="width: 100%"
            >
              <el-option
                v-for="opt in authorityFlatOptions"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.sysModel.enableSwitch')">
            <el-switch v-model="fileForm.enabled" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="fileDrawerVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="saveFile">{{ $t('settings.general.confirm') }}</el-button>
      </template>
    </el-drawer>

    <!-- 向量存储编辑抽屉 -->
    <el-drawer v-model="vectorDrawerVisible" :title="vectorEditId ? $t('rag.sysModel.vectorDrawerEdit') : $t('rag.sysModel.vectorDrawerAdd')" size="480px" destroy-on-close>
      <div style="padding: 0 16px">
        <el-form :model="vectorForm" label-width="120px">
          <el-form-item :label="$t('rag.sysModel.labelName')" required>
            <el-input v-model="vectorForm.name" :placeholder="$t('rag.sysModel.phConfigName')" />
          </el-form-item>
          <el-form-item :label="$t('rag.sysModel.labelType')" required>
            <el-select v-model="vectorForm.provider" :placeholder="$t('rag.sysModel.phSelect')" style="width: 100%" :disabled="!!vectorEditId">
              <el-option :label="$t('rag.sysModel.provPg')" value="postgresql" />
              <el-option :label="$t('rag.sysModel.provEs')" value="elasticsearch" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="vectorForm.provider === 'elasticsearch'" :label="$t('rag.sysModel.labelAddress')">
            <el-input v-model="vectorForm.config.address" :placeholder="$t('rag.sysModel.phEsAddr')" />
          </el-form-item>
          <el-form-item v-if="vectorForm.provider === 'elasticsearch'" :label="$t('rag.sysModel.labelUsername')">
            <el-input v-model="vectorForm.config.username" :placeholder="$t('rag.sysModel.phOptional')" />
          </el-form-item>
          <el-form-item v-if="vectorForm.provider === 'elasticsearch'" :label="$t('rag.sysModel.labelPassword')">
            <el-input v-model="vectorForm.config.password" type="password" :placeholder="$t('rag.sysModel.phOptional')" show-password />
          </el-form-item>
          <el-form-item v-if="vectorForm.provider === 'postgresql'" :label="$t('rag.sysModel.noteLabel')">
            <span class="text-gray-500 text-sm">{{ $t('rag.sysModel.pgVectorNote') }}</span>
          </el-form-item>
          <el-form-item :label="$t('rag.sysModel.allowAllRoles')">
            <el-switch v-model="vectorForm.allowAll" />
            <div class="text-gray-400 text-xs mt-1">{{ $t('rag.sysModel.allowAllHintVector') }}</div>
          </el-form-item>
          <el-form-item v-if="!vectorForm.allowAll" :label="$t('rag.sysModel.allowedRoles')" required>
            <el-select
              v-model="vectorForm.allowedAuthorityIds"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              :placeholder="$t('rag.sysModel.phRolesVector')"
              style="width: 100%"
            >
              <el-option
                v-for="opt in authorityFlatOptions"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.sysModel.enableSwitch')">
            <el-switch v-model="vectorForm.enabled" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="vectorDrawerVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="saveVector">{{ $t('settings.general.confirm') }}</el-button>
      </template>
    </el-drawer>

    <!-- 添加模型对话框 -->
    <el-dialog v-model="addVisible" :title="$t('rag.sysModel.dialogAddAdmin')" width="500px">
      <el-form :model="addForm" label-width="110px">
        <el-form-item :label="$t('rag.sysModel.colProvider')" required>
          <el-select
            v-model="addForm.name"
            :placeholder="$t('rag.userModel.phSelectSearch')"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="p in providerOptions"
              :key="p.value"
              :label="p.label"
              :value="p.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.colModelName')" required>
          <el-input v-model="addForm.modelName" :placeholder="$t('rag.userModel.phModelExample')" />
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.colScenarios')" required>
          <el-checkbox-group v-model="addForm.modelTypes" @change="(v) => loadProviderOpts(v?.length ? v : ['chat'])">
            <el-checkbox v-for="opt in modelTypeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colBaseUrl')">
          <el-input v-model="addForm.baseUrl" :placeholder="$t('rag.sysModel.phBaseUrlHttps')" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelApiKey')">
          <el-input v-model="addForm.apiKey" type="password" :placeholder="$t('rag.sysModel.phApiKey')" show-password />
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.labelMaxContext')">
          <el-input-number v-model="addForm.maxContextTokens" :min="0" :max="1000000" style="width: 100%" />
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.sysModel.maxContextHint') }}</div>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelCapabilities')">
          <el-checkbox v-model="addForm.supportsDeepThinking">{{ $t('rag.sysModel.capDeepShort') }}</el-checkbox>
          <el-checkbox v-model="addForm.supportsToolCall">{{ $t('rag.sysModel.capToolShort') }}</el-checkbox>
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.labelShareScope')">
          <el-select v-model="addForm.shareScope" style="width: 100%">
            <el-option :label="$t('rag.sysModel.shareAll')" value="all" />
            <el-option :label="$t('rag.sysModel.sharePrivate')" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.enableSwitch')">
          <el-switch v-model="addForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="doAdd">{{ $t('rag.sysModel.add') }}</el-button>
      </template>
    </el-dialog>

    <!-- 编辑模型对话框 -->
    <el-dialog v-model="editVisible" :title="$t('rag.sysModel.dialogEditAdmin')" width="500px">
      <el-form :model="editForm" label-width="110px">
        <el-form-item :label="$t('rag.sysModel.colProvider')" required>
          <el-select
            v-model="editForm.name"
            :placeholder="$t('rag.userModel.phSelectSearch')"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="p in providerOptions"
              :key="p.value"
              :label="p.label"
              :value="p.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.colModelName')" required>
          <el-input v-model="editForm.modelName" :placeholder="$t('rag.userModel.phModelExample')" />
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.colScenarios')" required>
          <el-checkbox-group v-model="editForm.modelTypes" @change="(v) => loadProviderOpts(v?.length ? v : ['chat'])">
            <el-checkbox v-for="opt in modelTypeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.colBaseUrl')">
          <el-input v-model="editForm.baseUrl" :placeholder="$t('rag.sysModel.phBaseUrlHttps')" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelApiKey')">
          <el-input v-model="editForm.apiKey" type="password" :placeholder="$t('rag.userModel.phApiKeyUnchanged')" show-password />
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.labelMaxContext')">
          <el-input-number v-model="editForm.maxContextTokens" :min="0" :max="1000000" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="$t('rag.userModel.labelCapabilities')">
          <el-checkbox v-model="editForm.supportsDeepThinking">{{ $t('rag.sysModel.capDeepShort') }}</el-checkbox>
          <el-checkbox v-model="editForm.supportsToolCall">{{ $t('rag.sysModel.capToolShort') }}</el-checkbox>
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.labelShareScope')">
          <el-select v-model="editForm.shareScope" style="width: 100%">
            <el-option :label="$t('rag.sysModel.shareAll')" value="all" />
            <el-option :label="$t('rag.sysModel.sharePrivate')" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.sysModel.enableSwitch')">
          <el-switch v-model="editForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="doEdit">{{ $t('rag.docs.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    listAdminModels,
    createAdminModel,
    updateAdminModel,
    deleteAdminModel,
    getSystemDefaults,
    setSystemDefault,
    clearSystemDefault,
    listAvailableProviders,
    listSystemWebSearchProviders,
    getSystemWebSearchConfig,
    setSystemWebSearchConfig,
    clearSystemWebSearchConfig,
    listGlobalKnowledgeBases,
    setGlobalKnowledgeBase,
    removeGlobalKnowledgeBase,
    listAllKnowledgeBases,
    listVectorStoreConfigsFull,
    createVectorStoreConfig,
    updateVectorStoreConfig,
    deleteVectorStoreConfig,
    listFileStorageConfigsFull,
    createFileStorageConfig,
    updateFileStorageConfig,
    deleteFileStorageConfig
  } from '@/api/rag'
  import { getAuthorityList } from '@/api/authority'
  import { authorityDisplayName } from '@/utils/authorityI18n'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, computed, onMounted, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'RagSystemModel' })

  const { t } = useI18n()

  const modelTypeOptions = computed(() => [
    { value: 'chat', label: t('rag.sysModel.typeChat') },
    { value: 'embedding', label: t('rag.sysModel.typeEmbedding') },
    { value: 'rerank', label: t('rag.sysModel.typeRerank') },
    { value: 'speech2text', label: t('rag.sysModel.typeSpeech2text') },
    { value: 'tts', label: t('rag.sysModel.typeTts') },
    { value: 'ocr', label: t('rag.sysModel.typeOcr') },
    { value: 'cv', label: t('rag.sysModel.typeCv') }
  ])

  const modelTypeLabel = (type) =>
    modelTypeOptions.value.find((o) => o.value === type)?.label || type

  const shareScopeLabel = (s) => {
    if (s === 'all') return t('rag.sysModel.shareAll')
    if (s === 'private') return t('rag.sysModel.sharePrivate')
    if (s === 'role') return t('rag.sysModel.shareRole')
    if (s === 'org') return t('rag.sysModel.shareOrg')
    return t('rag.sysModel.shareAll')
  }

  const activeTab = ref('models')
  const adminModels = ref([])
  const providerOptions = ref([])
  const addVisible = ref(false)
  const editVisible = ref(false)
  const systemDefaults = ref({})

  const defaultAddForm = () => ({
    name: 'openai',
    modelName: '',
    modelTypes: ['chat'],
    baseUrl: '',
    apiKey: '',
    maxContextTokens: 10000,
    supportsDeepThinking: false,
    supportsToolCall: true,
    shareScope: 'all',
    enabled: true
  })

  const addForm = ref(defaultAddForm())
  const editForm = ref({
    id: 0,
    name: 'openai',
    modelName: '',
    modelTypes: ['chat'],
    baseUrl: '',
    apiKey: '',
    maxContextTokens: 0,
    supportsDeepThinking: false,
    supportsToolCall: true,
    shareScope: 'all',
    enabled: true
  })

  // ===== 管理员模型 =====
  const loadAdminModels = async () => {
    const res = await listAdminModels()
    if (res.code === 0) {
      adminModels.value = res.data?.list || []
    }
  }

  const loadProviderOpts = async (scenarioTypes = ['chat']) => {
    const res = await listAvailableProviders({ scenarioTypes })
    if (res.code === 0 && res.data?.length) {
      providerOptions.value = res.data
    } else {
      providerOptions.value = [
        { value: 'openai', label: 'OpenAI' },
        { value: 'ollama', label: 'Ollama' }
      ]
    }
  }

  const openAdd = async () => {
    addForm.value = defaultAddForm()
    await loadProviderOpts(['chat'])
    addVisible.value = true
  }

  const doAdd = async () => {
    if (!addForm.value.modelName) {
      ElMessage.warning(t('rag.sysModel.needModelName'))
      return
    }
    if (!addForm.value.modelTypes?.length) {
      ElMessage.warning(t('rag.sysModel.needScenario'))
      return
    }
    const res = await createAdminModel(addForm.value)
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.addOk'))
      addVisible.value = false
      loadAdminModels()
    }
  }

  const openEdit = async (row) => {
    const types = row.modelTypes?.length ? row.modelTypes : ['chat']
    editForm.value = {
      id: row.ID,
      name: (row.name || '').toLowerCase() || 'openai',
      modelName: row.modelName || '',
      modelTypes: [...types],
      baseUrl: row.baseUrl || '',
      apiKey: '',
      maxContextTokens: row.maxContextTokens || 0,
      supportsDeepThinking: !!row.supportsDeepThinking,
      supportsToolCall: row.supportsToolCall !== false,
      shareScope: row.shareScope || 'all',
      enabled: row.enabled !== false
    }
    await loadProviderOpts(types)
    editVisible.value = true
  }

  const doEdit = async () => {
    if (!editForm.value.modelName) {
      ElMessage.warning(t('rag.sysModel.needModelName'))
      return
    }
    const res = await updateAdminModel(editForm.value)
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.updateOk'))
      editVisible.value = false
      loadAdminModels()
    }
  }

  const deleteModel = (row) => {
    ElMessageBox.confirm(t('rag.sysModel.deleteModelBody'), t('rag.userModel.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteAdminModel({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('rag.sysModel.deleteOk'))
        loadAdminModels()
      }
    })
  }

  // ===== 系统默认模型 =====
  const modelsForType = (modelType) => {
    return adminModels.value.filter((m) =>
      m.enabled && (m.modelTypes || []).includes(modelType)
    )
  }

  const loadSystemDefaults = async () => {
    const res = await getSystemDefaults()
    if (res.code === 0 && res.data) {
      const map = {}
      res.data.forEach((d) => {
        map[d.modelType] = d.llmProviderId
      })
      systemDefaults.value = map
    }
  }

  const saveSystemDefault = async (modelType, value) => {
    if (!value) {
      const res = await clearSystemDefault({ modelType })
      if (res.code === 0) {
        systemDefaults.value[modelType] = null
        ElMessage.success(t('rag.sysModel.msgCleared'))
      }
      return
    }
    const res = await setSystemDefault({ modelType, llmProviderId: value })
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.msgSaved'))
    }
  }

  // ===== 系统互联网搜索配置 =====
  const sysWebSearchProviders = ref([])
  const sysWebSearchForm = ref({ provider: 'duckduckgo', config: {} })
  const sysCurrentWebSearchSchema = ref([])

  const loadSysWebSearchProviders = async () => {
    const res = await listSystemWebSearchProviders()
    if (res.code === 0 && res.data?.length) {
      sysWebSearchProviders.value = res.data
    }
  }

  const loadSysWebSearchConfig = async () => {
    await loadSysWebSearchProviders()
    const res = await getSystemWebSearchConfig()
    if (res.code === 0 && res.data) {
      sysWebSearchForm.value = {
        provider: res.data.provider || 'duckduckgo',
        config: { ...(res.data.config || {}) }
      }
    }
    const p = sysWebSearchProviders.value.find((x) => x.id === sysWebSearchForm.value.provider)
    sysCurrentWebSearchSchema.value = p?.configSchema || []
    ensureSysConfigKeys()
  }

  const ensureSysConfigKeys = () => {
    const cfg = sysWebSearchForm.value.config || {}
    for (const f of sysCurrentWebSearchSchema.value) {
      if (cfg[f.key] === undefined) cfg[f.key] = ''
    }
    sysWebSearchForm.value.config = cfg
  }

  watch(() => sysWebSearchForm.value.provider, () => {
    const p = sysWebSearchProviders.value.find((x) => x.id === sysWebSearchForm.value.provider)
    sysCurrentWebSearchSchema.value = p?.configSchema || []
    ensureSysConfigKeys()
  })

  const saveSysWebSearchConfig = async () => {
    const provider = sysWebSearchForm.value.provider
    if (!provider) {
      ElMessage.warning(t('rag.sysModel.pickEngineWarn'))
      return
    }
    const schema = sysWebSearchProviders.value.find((x) => x.id === provider)?.configSchema || []
    const config = {}
    for (const f of schema) {
      const v = sysWebSearchForm.value.config?.[f.key]
      if (f.required && !v) {
        ElMessage.warning(t('rag.sysModel.fillField', { label: f.label }))
        return
      }
      if (v) config[f.key] = v
    }
    const res = await setSystemWebSearchConfig({ provider, config })
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.saveOk'))
    }
  }

  const clearSysWebSearchConfig = async () => {
    const res = await clearSystemWebSearchConfig()
    if (res.code === 0) {
      sysWebSearchForm.value = { provider: 'duckduckgo', config: {} }
      sysCurrentWebSearchSchema.value = []
      ElMessage.success(t('rag.sysModel.msgCleared'))
    }
  }

  // ===== 初始化 =====
  onMounted(async () => {
    await loadAdminModels()
    await loadProviderOpts()
  })

  // ===== 全局共享知识库 =====
  const globalKbList = ref([])
  const allKbList = ref([])
  const globalKbAddId = ref(null)

  const availableKbsForGlobal = computed(() => {
    const usedIds = new Set(globalKbList.value.map(g => g.knowledgeBaseId))
    return allKbList.value.filter(kb => !usedIds.has(kb.id))
  })

  const loadGlobalKbs = async () => {
    const [gRes, aRes] = await Promise.all([
      listGlobalKnowledgeBases(),
      listAllKnowledgeBases()
    ])
    if (gRes.code === 0) globalKbList.value = gRes.data || []
    if (aRes.code === 0) allKbList.value = aRes.data || []
  }

  const addGlobalKb = async () => {
    if (!globalKbAddId.value) return
    const res = await setGlobalKnowledgeBase({ knowledgeBaseId: globalKbAddId.value })
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.addOk'))
      globalKbAddId.value = null
      loadGlobalKbs()
    } else {
      ElMessage.error(res.msg || t('rag.sysModel.addGlobalFail'))
    }
  }

  const removeGlobalKb = (row) => {
    ElMessageBox.confirm(
      t('rag.sysModel.removeGlobalKbBody', {
        name:
          row.knowledgeBaseName ||
          t('rag.sysModel.kbFallback', { id: row.knowledgeBaseId })
      }),
      t('rag.sysModel.removeGlobalKbTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      const res = await removeGlobalKnowledgeBase({ knowledgeBaseId: row.knowledgeBaseId })
      if (res.code === 0) {
        ElMessage.success(t('rag.sysModel.removedOk'))
        loadGlobalKbs()
      }
    }).catch(() => {})
  }

  watch(activeTab, () => {
    if (activeTab.value === 'models') {
      loadAdminModels()
    } else if (activeTab.value === 'defaults') {
      loadAdminModels()
      loadSystemDefaults()
    } else if (activeTab.value === 'globalKb') {
      loadGlobalKbs()
    } else if (activeTab.value === 'webSearch') {
      loadSysWebSearchConfig()
    } else if (activeTab.value === 'vectorStorage') {
      loadAuthorityOptions()
      loadVectorList()
    } else if (activeTab.value === 'fileStorage') {
      loadAuthorityOptions()
      loadFileList()
    }
  })

  // ===== 向量存储 =====
  const vectorList = ref([])
  const vectorPage = ref(1)
  const vectorPageSize = ref(10)
  const vectorTotal = ref(0)
  const vectorDrawerVisible = ref(false)
  const vectorEditId = ref(0)
  const defaultVectorForm = () => ({
    name: '',
    provider: 'postgresql',
    config: { address: '', username: '', password: '' },
    enabled: true,
    allowAll: true,
    allowedAuthorityIds: []
  })
  const vectorForm = ref(defaultVectorForm())

  const loadVectorList = async () => {
    const res = await listVectorStoreConfigsFull({
      page: vectorPage.value,
      pageSize: vectorPageSize.value
    })
    if (res.code === 0) {
      vectorList.value = res.data.list || []
      vectorTotal.value = res.data.total || 0
    }
  }

  const openVectorDrawer = async (row) => {
    await loadAuthorityOptions()
    vectorEditId.value = row?.ID || 0
    if (row) {
      vectorForm.value = {
        name: row.name,
        provider: row.provider || 'postgresql',
        config: {
          ...(row.config || {}),
          address: row.config?.address || '',
          username: row.config?.username || '',
          password: row.config?.password || ''
        },
        enabled: row.enabled ?? true,
        allowAll: row.allowAll !== false,
        allowedAuthorityIds: Array.isArray(row.allowedAuthorityIds)
          ? row.allowedAuthorityIds.map((x) => Number(x))
          : []
      }
    } else {
      vectorForm.value = defaultVectorForm()
    }
    vectorDrawerVisible.value = true
  }

  const saveVector = async () => {
    if (!vectorForm.value.name) {
      ElMessage.warning(t('rag.sysModel.needName'))
      return
    }
    if (!vectorForm.value.allowAll && !(vectorForm.value.allowedAuthorityIds || []).length) {
      ElMessage.warning(t('rag.sysModel.needRolesOrAllowAll'))
      return
    }
    const config = {}
    if (vectorForm.value.provider === 'elasticsearch') {
      if (vectorForm.value.config?.address) config.address = vectorForm.value.config.address
      if (vectorForm.value.config?.username) config.username = vectorForm.value.config.username
      if (vectorForm.value.config?.password) config.password = vectorForm.value.config.password
    }
    const rolePayload = {
      allowAll: vectorForm.value.allowAll,
      allowedAuthorityIds: vectorForm.value.allowAll ? [] : [...(vectorForm.value.allowedAuthorityIds || [])]
    }
    let res
    if (vectorEditId.value) {
      res = await updateVectorStoreConfig({
        id: vectorEditId.value,
        name: vectorForm.value.name,
        provider: vectorForm.value.provider,
        config: Object.keys(config).length ? config : undefined,
        enabled: vectorForm.value.enabled,
        ...rolePayload
      })
    } else {
      res = await createVectorStoreConfig({
        name: vectorForm.value.name,
        provider: vectorForm.value.provider,
        config,
        enabled: vectorForm.value.enabled,
        ...rolePayload
      })
    }
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.saveOk'))
      vectorDrawerVisible.value = false
      loadVectorList()
    } else {
      ElMessage.error(res.msg || t('rag.sysModel.saveFail'))
    }
  }

  const deleteVector = (row) => {
    ElMessageBox.confirm(t('rag.sysModel.deleteVectorConfirm'), t('rag.userModel.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteVectorStoreConfig({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('rag.sysModel.deleteOk'))
        loadVectorList()
      } else {
        ElMessage.error(res.msg || t('rag.docs.deleteFail'))
      }
    }).catch(() => {})
  }

  // ===== 文件存储 =====
  const fileList = ref([])
  const filePage = ref(1)
  const filePageSize = ref(10)
  const fileTotal = ref(0)
  const fileDrawerVisible = ref(false)
  const fileEditId = ref(0)
  const authorityFlatOptions = ref([])

  const defaultFileForm = () => ({
    name: '',
    provider: 'local',
    config: { storePath: '', endpoint: '', accessKeyId: '', accessKeySecret: '', bucketName: '' },
    enabled: true,
    allowAll: true,
    allowedAuthorityIds: []
  })
  const fileForm = ref(defaultFileForm())

  const flattenAuthorities = (list) => {
    const out = []
    const walk = (arr) => {
      for (const a of arr || []) {
        const id = Number(a.authorityId)
        if (!Number.isFinite(id)) continue
        out.push({
          value: id,
          label: `${authorityDisplayName(a, t)} (${id})`
        })
        if (a.children?.length) walk(a.children)
      }
    }
    walk(Array.isArray(list) ? list : list ? [list] : [])
    return out
  }

  const loadAuthorityOptions = async () => {
    const res = await getAuthorityList()
    if (res.code === 0 && res.data) {
      authorityFlatOptions.value = flattenAuthorities(Array.isArray(res.data) ? res.data : [res.data])
    }
  }

  const formatRagSettingRoleLabels = (ids) => {
    if (!ids?.length) return t('rag.sysModel.rolesNoneConfigured')
    const opts = authorityFlatOptions.value
    const labelOf = (id) => {
      const n = Number(id)
      const hit = opts.find((o) => o.value === n)
      return hit ? hit.label : String(id)
    }
    return ids.map(labelOf).join(t('rag.sysModel.listJoiner'))
  }

  const loadFileList = async () => {
    const res = await listFileStorageConfigsFull({
      page: filePage.value,
      pageSize: filePageSize.value
    })
    if (res.code === 0) {
      fileList.value = res.data.list || []
      fileTotal.value = res.data.total || 0
    }
  }

  const openFileDrawer = async (row) => {
    await loadAuthorityOptions()
    fileEditId.value = row?.ID || 0
    if (row) {
      const c = row.config || {}
      fileForm.value = {
        name: row.name,
        provider: row.provider || 'local',
        config: {
          storePath: c.storePath || c.store_path || '',
          endpoint: c.endpoint || '',
          accessKeyId: c.accessKeyId || c.access_key_id || '',
          accessKeySecret: c.accessKeySecret || c.access_key_secret || '',
          bucketName: c.bucketName || c.bucket_name || ''
        },
        enabled: row.enabled ?? true,
        allowAll: row.allowAll !== false,
        allowedAuthorityIds: Array.isArray(row.allowedAuthorityIds)
          ? row.allowedAuthorityIds.map((x) => Number(x))
          : []
      }
    } else {
      fileForm.value = defaultFileForm()
    }
    fileDrawerVisible.value = true
  }

  const saveFile = async () => {
    if (!fileForm.value.name) {
      ElMessage.warning(t('rag.sysModel.needName'))
      return
    }
    if (!fileForm.value.allowAll && !(fileForm.value.allowedAuthorityIds || []).length) {
      ElMessage.warning(t('rag.sysModel.needRolesOrAllowAll'))
      return
    }
    const config = {}
    if (fileForm.value.provider === 'local' && fileForm.value.config?.storePath) {
      config.storePath = fileForm.value.config.storePath
      config.store_path = fileForm.value.config.storePath
    } else if (fileForm.value.provider === 'minio') {
      if (fileForm.value.config?.endpoint) config.endpoint = fileForm.value.config.endpoint
      if (fileForm.value.config?.accessKeyId) config.accessKeyId = fileForm.value.config.accessKeyId
      if (fileForm.value.config?.accessKeySecret) config.accessKeySecret = fileForm.value.config.accessKeySecret
      if (fileForm.value.config?.bucketName) config.bucketName = fileForm.value.config.bucketName
    }
    const payloadBase = {
      name: fileForm.value.name,
      provider: fileForm.value.provider,
      config: Object.keys(config).length ? config : undefined,
      enabled: fileForm.value.enabled,
      allowAll: fileForm.value.allowAll,
      allowedAuthorityIds: fileForm.value.allowAll ? [] : [...(fileForm.value.allowedAuthorityIds || [])]
    }
    let res
    if (fileEditId.value) {
      res = await updateFileStorageConfig({ id: fileEditId.value, ...payloadBase })
    } else {
      res = await createFileStorageConfig({
        name: fileForm.value.name,
        provider: fileForm.value.provider,
        config,
        enabled: fileForm.value.enabled,
        allowAll: fileForm.value.allowAll,
        allowedAuthorityIds: payloadBase.allowedAuthorityIds
      })
    }
    if (res.code === 0) {
      ElMessage.success(t('rag.sysModel.saveOk'))
      fileDrawerVisible.value = false
      loadFileList()
    } else {
      ElMessage.error(res.msg || t('rag.sysModel.saveFail'))
    }
  }

  const deleteFile = (row) => {
    ElMessageBox.confirm(t('rag.sysModel.deleteFileConfirm'), t('rag.userModel.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteFileStorageConfig({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('rag.sysModel.deleteOk'))
        loadFileList()
      } else {
        ElMessage.error(res.msg || t('rag.docs.deleteFail'))
      }
    }).catch(() => {})
  }
</script>
