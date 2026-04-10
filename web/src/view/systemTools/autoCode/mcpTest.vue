<template>
  <div class="p-2">

    <el-card class="mb-2">
      <template #header>
        <div class="flex justify-between items-center font-bold">
          <span>{{ $t('tools.mcpTest.configExampleTitle') }}</span>
          <el-tooltip :content="$t('tools.mcpTest.copyConfigTooltip')" placement="top">
            <el-button :icon="DocumentCopy" circle @click="copyMcpConfig" />
          </el-tooltip>
        </div>
      </template>
      <pre class="font-mono whitespace-pre-wrap break-words bg-gray-100 p-2.5 rounded text-gray-700">{{ mcpServerConfig }}</pre>
    </el-card>

    
    <el-row :gutter="8">
      <el-col v-for="tool in mcpTools" :key="tool.name" :xs="24" :sm="12" :md="12" :lg="8">
        <el-card class="mb-5 min-h-[150px] flex flex-col overflow-hidden">
          <template #header>
            <div class="flex justify-between items-center font-bold">
              <span>{{ tool.name }}</span>
              <el-tooltip :content="$t('tools.mcpTest.testToolTooltip')" placement="top">
                <el-button :icon="VideoPlay" circle @click="openTestDialog(tool)" />
              </el-tooltip>
            </div>
          </template>
          <div class="text-sm mb-1">{{ tool.description }}</div>
          <div v-if="tool.inputSchema && tool.inputSchema.properties && Object.keys(tool.inputSchema.properties).length > 0" class="mt-1 text-xs overflow-y-auto max-h-[100px] p-2 border-t border-gray-200 bg-gray-50 rounded-b">
            <p class="font-semibold mb-1 text-gray-700 flex items-center">
              <span class="mr-1 my-2">{{ $t('tools.mcpTest.paramList') }}</span>
              <span class="text-xs text-gray-500">({{ Object.keys(tool.inputSchema.properties).length }})</span>
            </p>
            <div class="space-y-2">
              <div v-for="(propDetails, propName) in tool.inputSchema.properties" :key="propName" class="flex flex-col p-1.5 bg-white rounded border border-gray-100 hover:border-gray-300 transition-colors">
                <div class="flex items-center justify-between">
                  <div class="flex items-center">
                    <span class="font-medium text-gray-800">{{ propName }}</span>
                    <span v-if="tool.inputSchema.required && tool.inputSchema.required.includes(propName)" class="ml-1 text-red-500 text-xs">*</span>
                  </div>
                  <span class="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded">{{ propDetails.type }}</span>
                </div>
                <div class="text-gray-500 mt-0.5 text-xs line-clamp-2" :title="propDetails.description || $t('tools.mcpTest.noDesc')">
                  {{ propDetails.description || $t('tools.mcpTest.noDesc') }}
                </div>
              </div>
            </div>
          </div>
          <div v-else class="mt-1 text-xs p-2 border-t border-gray-200 bg-gray-50 rounded-b flex items-center justify-center">
            <span class="text-gray-500 italic py-3">{{ $t('tools.mcpTest.noInputParams') }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>


    <el-dialog
      v-model="testDialogVisible"
      :title="currentTestingTool ? $t('tools.mcpTest.dialogTitleNamed', { name: currentTestingTool.name }) : $t('tools.mcpTest.paramTestTitle')"
      width="60%"
      :before-close="handleCloseDialog"
    >
      <el-form v-if="currentTestingTool" :model="testParamsForm" ref="testParamsFormRef" label-width="120px" label-position="top" class="max-h-[calc(60vh-120px)] overflow-y-auto">
        <el-form-item
          v-for="(propDetails, propName) in currentTestingTool.inputSchema.properties"
          :key="propName"
          :label="propDetails.description || propName"
          :prop="propName"
          :rules="paramRules(propName, propDetails)"
        >
          <el-input 
            v-if="propDetails.type === 'string' && !propDetails.enum" 
            v-model="testParamsForm[propName]" 
            :placeholder="propDetails.description || $t('tools.mcpTest.phEnter', { field: propName })" 
          />
          <el-input 
            v-else-if="propDetails.type === 'number'" 
            v-model.number="testParamsForm[propName]" 
            type="number" 
            :placeholder="propDetails.description || $t('tools.mcpTest.phEnterNumber', { field: propName })" 
          />
          <el-select 
            v-else-if="propDetails.type === 'boolean'" 
            v-model="testParamsForm[propName]" 
            :placeholder="propDetails.description || $t('tools.mcpTest.phSelect')"
          >
            <el-option :label="$t('tools.mcpTest.boolTrue')" :value="true" />
            <el-option :label="$t('tools.mcpTest.boolFalse')" :value="false" />
          </el-select>
          <el-select 
            v-else-if="propDetails.type === 'string' && propDetails.enum" 
            v-model="testParamsForm[propName]" 
            :placeholder="propDetails.description || $t('tools.mcpTest.phSelectEnum', { field: propName })"
          >
            <el-option v-for="enumValue in propDetails.enum" :key="enumValue" :label="enumValue" :value="enumValue" />
          </el-select>
          <el-input 
            v-else 
            type="textarea" 
            v-model="testParamsForm[propName]" 
            :placeholder="$t('tools.mcpTest.phJson', { field: propDetails.description || propName })" 
            :autosize="{ minRows: 2, maxRows: 6 }"
          />
        </el-form-item>
      </el-form>
      <div v-if="apiDialogResponse" class="mt-5 p-[15px] border border-gray-200 rounded bg-gray-50">
        <h4 class="mt-0 mb-2.5 text-base">{{ $t('tools.mcpTest.apiResponseTitle') }}</h4>
        <div v-if="typeof apiDialogResponse === 'string'">
          <pre class="bg-gray-100 p-2.5 rounded whitespace-pre-wrap break-words overflow-y-auto">{{ apiDialogResponse }}</pre>
        </div>
        <div v-else-if="apiDialogResponse.type === 'image' && apiDialogResponse.content">
           <el-image
              class="max-w-full max-h-[300px]"
              :src="apiDialogResponse.content"
              :preview-src-list="[apiDialogResponse.content]"
              fit="contain"
            />
        </div>
         <div v-else-if="apiDialogResponse.type === 'text' && apiDialogResponse.content">
          <pre class="bg-gray-100 p-2.5 rounded whitespace-pre-wrap break-words overflow-y-auto">{{ apiDialogResponse.content }}</pre>
        </div>
        <div v-else>
          <pre class="bg-gray-100 p-2.5 rounded whitespace-pre-wrap break-words overflow-y-auto">{{ JSON.stringify(apiDialogResponse, null, 2) }}</pre>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="testDialogVisible = false">{{ $t('tools.mcpTest.btnCancel') }}</el-button>
          <el-button type="primary" @click="handleDialogTestTool">{{ $t('tools.mcpTest.btnTest') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { VideoPlay, DocumentCopy } from '@element-plus/icons-vue'
import { mcpList, mcpTest } from '@/api/autoCode'

defineOptions({
  name: 'MCPTest'
})

const { t } = useI18n()

const paramRules = (propName, propDetails) => {
  const req = currentTestingTool.value?.inputSchema?.required
  if (!req || !req.includes(propName)) return []
  const field = propDetails.description || propName
  return [{ required: true, message: t('tools.mcpTest.ruleRequiredEnter', { field }), trigger: 'blur' }]
}

const mcpTools = ref([])
const testDialogVisible = ref(false)
const currentTestingTool = ref(null)
const testParamsForm = reactive({})
const testParamsFormRef = ref(null)
const apiDialogResponse = ref(null)

const mcpServerConfig = ref(JSON.stringify({
  "mcpServers": {
    "lightningrag": {
      "url": "https://127.0.0.1/sse"
    }
  }
}, null, 2))

const fetchMcpTools = async () => {
    const res = await mcpList()
    if (res.code === 0 && res.data && res.data.list.tools) {
      mcpTools.value = res.data.list.tools
      mcpServerConfig.value = JSON.stringify(res.data.mcpServerConfig, null, 2)
    } else {
      ElMessage.error(res.msg || t('tools.mcpTest.fetchListFail'))
    }
}

onMounted(() => {
  fetchMcpTools()
})

const copyMcpConfig = async () => {
  try {
    await navigator.clipboard.writeText(mcpServerConfig.value)
    ElMessage.success(t('tools.mcpTest.copyOk'))
  } catch (err) {
    ElMessage.error(t('tools.mcpTest.copyFail', { err: String(err) }))
  }
}

const openTestDialog = (tool) => {
  currentTestingTool.value = tool
  apiDialogResponse.value = null

  for (const key in testParamsForm) {
    delete testParamsForm[key]
  }
  if (tool.inputSchema && tool.inputSchema.properties) {
    Object.keys(tool.inputSchema.properties).forEach(propName => {
      const propDetails = tool.inputSchema.properties[propName]
      if (propDetails.default !== undefined) {
        testParamsForm[propName] = propDetails.default
      } else if (propDetails.type === 'boolean') {
        testParamsForm[propName] = false
      } else if (propDetails.type === 'number') {
        testParamsForm[propName] = null
      } else if (propDetails.type === 'object' || propDetails.type === 'array') {
        testParamsForm[propName] = ''
      } else {
        testParamsForm[propName] = ''
      }
    })
  }
  testDialogVisible.value = true
  if (testParamsFormRef.value) {
    testParamsFormRef.value.clearValidate()
  }
}

const handleCloseDialog = (done) => {
  apiDialogResponse.value = null
  done()
}

const handleDialogTestTool = async () => {
  if (!currentTestingTool.value) {
    ElMessage.warning(t('tools.mcpTest.noToolSelected'))
    return
  }
  if (testParamsFormRef.value) {
    testParamsFormRef.value.validate(async (valid) => {
      if (valid) {
          const toolName = currentTestingTool.value.name
          const payload = { ...testParamsForm }

          if (currentTestingTool.value.inputSchema && currentTestingTool.value.inputSchema.properties) {
            for (const propName of Object.keys(currentTestingTool.value.inputSchema.properties)) {
              const propDetails = currentTestingTool.value.inputSchema.properties[propName]
              if ((propDetails.type === 'object' || propDetails.type === 'array') && payload[propName] && typeof payload[propName] === 'string') {
                try {
                  payload[propName] = JSON.parse(payload[propName])
                } catch (e) {
                  ElMessage.error(t('tools.mcpTest.invalidJsonParam', { name: propName, msg: e.message }))
                  return
                }
              }
            }
          }

          const res = await mcpTest({
            name:toolName,
            arguments:payload
          })
          apiDialogResponse.value = res.data
          if (res.code === 0) {
            ElMessage.success(t('tools.mcpTest.apiCallOk'))
          }
      }
    })
  }
}

</script>
