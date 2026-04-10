<template>
  <div>
    <warning-bar
      :title="$t('tools.autoPkg.devOnlyWarn')"
    />
    <div class="lrag-table-box">
      <div class="lrag-btn-list gap-3 flex items-center">
        <el-button type="primary" icon="plus" @click="openDialog('addApi')">
          {{ $t('tools.autoPkg.addNew') }}
        </el-button>
      </div>
      <el-table :data="tableData">
        <el-table-column align="left" :label="$t('common.colId')" width="120" prop="ID" />
        <el-table-column
          align="left"
          :label="$t('common.colPackageName')"
          width="150"
          prop="packageName"
        />
        <el-table-column
          align="left"
          :label="$t('common.colTemplate')"
          width="150"
          prop="template"
        />
        <el-table-column align="left" :label="$t('common.colDisplayName')" width="150" prop="label" />
        <el-table-column
          align="left"
          :label="$t('common.colDescription')"
          min-width="150"
          prop="desc"
        />

        <el-table-column align="left" :label="$t('common.colActions')" width="200">
          <template #default="scope">
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
    </div>

    <el-drawer v-model="dialogFormVisible" size="40%" :show-close="false">
      <warning-bar
        :title="$t('tools.autoPkg.templateHint')"
      />
      <el-form ref="pkgForm" :model="form" :rules="rules" label-width="80px">
        <el-form-item :label="$t('common.colPackageName')" prop="packageName">
          <el-input v-model="form.packageName" autocomplete="off" />
        </el-form-item>
        <el-form-item :label="$t('common.colTemplate')" prop="template">
          <el-select v-model="form.template">
            <el-option
              v-for="template in templatesOptions"
              :label="template"
              :value="template"
              :key="template"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('common.colDisplayName')" prop="label">
          <el-input v-model="form.label" autocomplete="off" />
        </el-form-item>
        <el-form-item :label="$t('common.colDescription')" prop="desc">
          <el-input v-model="form.desc" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('tools.autoPkg.createPackage') }}</span>
          <div>
            <el-button @click="closeDialog">{{ $t('settings.general.cancel') }}</el-button>
            <el-button type="primary" @click="enterDialog">{{ $t('settings.general.confirm') }}</el-button>
          </div>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createPackageApi,
    getPackageApi,
    deletePackageApi,
    getTemplatesApi
  } from '@/api/autoCode'
  import { computed, ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({
    name: 'AutoPkg'
  })

  const { t } = useI18n()

  const form = ref({
    packageName: '',
    template: '',
    label: '',
    desc: ''
  })
  const templatesOptions = ref([])

  const getTemplates = async () => {
    const res = await getTemplatesApi()
    if (res.code === 0) {
      templatesOptions.value = res.data
    }
  }

  getTemplates()

  const validateData = (rule, value, callback) => {
    if (/[\u4E00-\u9FA5]/g.test(value)) {
      callback(new Error(t('tools.autoPkg.validateNoZh')))
    } else if (/^\d+$/.test(value[0])) {
      callback(new Error(t('tools.autoPkg.validateNoDigitStart')))
    } else if (!/^[a-zA-Z0-9_]+$/.test(value)) {
      callback(new Error(t('tools.autoPkg.validatePattern')))
    } else {
      callback()
    }
  }

  const rules = computed(() => ({
    packageName: [
      { required: true, message: t('tools.autoPkg.rulePackageName'), trigger: 'blur' },
      { validator: validateData, trigger: 'blur' }
    ],
    template: [
      { required: true, message: t('tools.autoPkg.ruleTemplate'), trigger: 'change' },
      { validator: validateData, trigger: 'blur' }
    ]
  }))

  const dialogFormVisible = ref(false)
  const openDialog = () => {
    dialogFormVisible.value = true
  }

  const closeDialog = () => {
    dialogFormVisible.value = false
    form.value = {
      packageName: '',
      template: '',
      label: '',
      desc: ''
    }
  }

  const pkgForm = ref(null)
  const enterDialog = async () => {
    pkgForm.value.validate(async (valid) => {
      if (valid) {
        const res = await createPackageApi(form.value)
        if (res.code === 0) {
          ElMessage({
            type: 'success',
            message: t('tools.autoPkg.addOk'),
            showClose: true
          })
        }
        getTableData()
        closeDialog()
      }
    })
  }

  const tableData = ref([])
  const getTableData = async () => {
    const table = await getPackageApi()
    if (table.code === 0) {
      tableData.value = table.data.pkgs
    }
  }

  const deleteApiFunc = async (row) => {
    ElMessageBox.confirm(
      t('tools.autoPkg.deleteWarn'),
      t('common.tipTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      const res = await deletePackageApi(row)
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('tools.autoPkg.deleteOk')
        })
        getTableData()
      }
    })
  }

  getTableData()
</script>
