<template>
  <div class="lrag-form-box">
    <el-form :model="form" ref="formRef" label-width="100px" :rules="rules">
      <el-form-item :label="$t('tools.autoCodeMcp.labelToolName')" prop="name">
        <el-input v-model="form.name" :placeholder="$t('tools.autoCodeMcp.phToolName')" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCodeMcp.labelToolDesc')" prop="description">
        <el-input type="textarea" v-model="form.description" :placeholder="$t('tools.autoCodeMcp.phToolDesc')" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCodeMcp.labelParamList')">
        <el-table :data="form.params"  style="width: 100%">
          <el-table-column prop="name" :label="$t('tools.autoCodeMcp.colParamName')" width="120">
            <template #default="scope">
              <el-input v-model="scope.row.name" :placeholder="$t('tools.autoCodeMcp.phParamName')" />
            </template>
          </el-table-column>
          <el-table-column prop="description" :label="$t('common.colDescription')" min-width="180">
            <template #default="scope">
              <el-input v-model="scope.row.description" :placeholder="$t('tools.autoCodeMcp.phDesc')" />
            </template>
          </el-table-column>
          <el-table-column prop="type" :label="$t('tools.autoCodeMcp.colType')" width="120">
            <template #default="scope">
              <el-select v-model="scope.row.type" :placeholder="$t('tools.autoCodeMcp.phType')">
                <el-option label="string" value="string" />
                <el-option label="number" value="number" />
                <el-option label="boolean" value="boolean" />
                <el-option label="object" value="object" />
                <el-option label="array" value="array" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column :label="$t('tools.autoCodeMcp.colDefault')" width="300">
            <template #default="scope">
              <el-input :disabled="scope.row.type === 'object'" v-model="scope.row.default" />
            </template>
          </el-table-column>
          <el-table-column prop="required" :label="$t('tools.autoCodeMcp.colRequired')" width="80">
            <template #default="scope">
              <el-checkbox v-model="scope.row.required" />
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.colActions')" width="80">
            <template #default="scope">
              <el-button type="text" @click="removeParam(scope.$index)">{{ $t('common.delete') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-form-item>
      <div class="flex justify-end">
        <el-button type="primary" icon="plus" @click="addParam" style="margin-top: 10px;">{{ $t('tools.autoCodeMcp.btnAddParam') }}</el-button>
      </div>
      <el-form-item :label="$t('tools.autoCodeMcp.labelResponse')">
        <el-table :data="form.response" style="width: 100%">
          <el-table-column prop="type" :label="$t('tools.autoCodeMcp.colType')" min-width="120">
            <template #default="scope">
              <el-select v-model="scope.row.type" :placeholder="$t('tools.autoCodeMcp.phType')">
                <el-option label="text" value="text" />
                <el-option label="image" value="image" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.colActions')" width="80">
            <template #default="scope">
              <el-button type="text" @click="removeResponse(scope.$index)">{{ $t('common.delete') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-form-item>
      <div class="flex justify-end">
        <el-button type="primary" icon="plus" @click="addResponse" style="margin-top: 10px;">{{ $t('tools.autoCodeMcp.btnAddResponse') }}</el-button>
      </div>

      <div class="flex justify-end mt-8">
        <el-button type="primary" @click="submit">{{ $t('tools.autoCodeMcp.btnGenerate') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { mcp } from '@/api/autoCode'

defineOptions({
  name: 'MCP'
})

const { t } = useI18n()

const formRef = ref(null)
const form = reactive({
  name: '',
  description: '',
  type: '',
  params: [],
  response: []
})

const rules = computed(() => ({
  name: [{ required: true, message: t('tools.autoCodeMcp.ruleName'), trigger: 'blur' }],
  description: [{ required: true, message: t('tools.autoCodeMcp.ruleDesc'), trigger: 'blur' }],
  type: [{ required: true, message: t('tools.autoCodeMcp.ruleType'), trigger: 'change' }]
}))

function addParam() {
  form.params.push({
    name: '',
    description: '',
    type: '',
    required: false
  })
}

function removeParam(index) {
  form.params.splice(index, 1)
}

function addResponse() {
  form.response.push({
    type: ''
  })
}

function removeResponse(index) {
  form.response.splice(index, 1)
}

function submit() {
  formRef.value.validate(async (valid) => {
    if (!valid) return
    // 简单校验参数
    for (const p of form.params) {
      if (!p.name || !p.description || !p.type) {
        ElMessage.error(t('tools.autoCodeMcp.errCompleteParams'))
        return
      }
    }
    for (const r of form.response) {
      if (!r.type) {
        ElMessage.error(t('tools.autoCodeMcp.errCompleteResponse'))
        return
      }
    }
      const res = await mcp(form)
      if (res.code === 0) {
        ElMessage.success(res.msg)
      }
  })
}
</script>
