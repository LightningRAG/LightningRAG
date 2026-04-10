<template>
  <div>
    <warning-bar :title="$t('tools.autoCode.fieldDialog.warningBar')" />
    <el-form
      ref="fieldDialogForm"
      :model="middleDate"
      label-width="120px"
      label-position="right"
      :rules="rules"
      class="grid grid-cols-2"
    >
      <el-form-item :label="$t('tools.autoCode.fields.colFieldName')" prop="fieldName">
        <el-input
          v-model="middleDate.fieldName"
          autocomplete="off"
          style="width: 80%"
        />
        <el-button style="width: 18%; margin-left: 2%" @click="autoFill">
          <span style="font-size: 12px">{{ $t('tools.autoCode.fieldDialog.btnAutoFill') }}</span>
        </el-button>
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colDisplayName')" prop="fieldDesc">
        <el-input v-model="middleDate.fieldDesc" autocomplete="off" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colFieldJson')" prop="fieldJson">
        <el-input v-model="middleDate.fieldJson" autocomplete="off" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colDbColumn')" prop="columnName">
        <el-input v-model="middleDate.columnName" autocomplete="off" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colDbComment')" prop="comment">
        <el-input v-model="middleDate.comment" autocomplete="off" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colFieldType')" prop="fieldType">
        <el-select
          v-model="middleDate.fieldType"
          style="width: 100%"
          :placeholder="$t('tools.autoCode.fields.phFieldType')"
          clearable
          @change="clearOther"
        >
          <el-option
            v-for="item in typeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
            :disabled="item.disabled"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        :label="middleDate.fieldType === 'enum' ? $t('tools.autoCode.fieldDialog.labelEnumValues') : $t('tools.autoCode.fieldDialog.labelTypeLength')"
        prop="dataTypeLong"
      >
        <el-input
          v-model="middleDate.dataTypeLong"
          :placeholder="
            middleDate.fieldType === 'enum'
              ? $t('tools.autoCode.fieldDialog.phEnumExample')
              : $t('tools.autoCode.fieldDialog.phDbTypeLength')
          "
        />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colSearchCond')" prop="fieldSearchType">
        <el-select
          v-model="middleDate.fieldSearchType"
          :disabled="middleDate.fieldType === 'json'"
          style="width: 100%"
          :placeholder="$t('tools.autoCode.fields.phSearchOp')"
          clearable
        >
          <el-option
            v-for="item in typeSearchOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
            :disabled="canSelect(item.value)"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.labelAssocDict')" prop="dictType">
        <el-select
          v-model="middleDate.dictType"
          style="width: 100%"
          :disabled="middleDate.fieldType !== 'string' && middleDate.fieldType !== 'array'"
          :placeholder="$t('tools.autoCode.fieldDialog.phSelectDict')"
          clearable
        >
          <el-option
            v-for="item in dictOptions"
            :key="item.type"
            :label="`${item.type}(${item.name})`"
            :value="item.type"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colDefaultValue')">
        <el-input
          v-model="middleDate.defaultValue"
          :placeholder="$t('tools.autoCode.fieldDialog.phDefaultValue')"
        />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colPk')">
        <el-checkbox v-model="middleDate.primaryKey" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fields.colIndexType')" prop="fieldIndexType">
        <el-select
          v-model="middleDate.fieldIndexType"
          :disabled="middleDate.fieldType === 'json'"
          style="width: 100%"
          :placeholder="$t('tools.autoCode.fields.phIndexType')"
          clearable
        >
          <el-option
            v-for="item in typeIndexOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
            :disabled="canSelect(item.value)"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swFormEdit')">
        <el-switch v-model="middleDate.form" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swTable')">
        <el-switch v-model="middleDate.table" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swDetail')">
        <el-switch v-model="middleDate.desc" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swExcel')">
        <el-switch v-model="middleDate.excel" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swSort')">
        <el-switch v-model="middleDate.sort" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swRequired')">
        <el-switch v-model="middleDate.require" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swClearable')">
        <el-switch v-model="middleDate.clearable" />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.swHideSearch')">
        <el-switch
          v-model="middleDate.fieldSearchHide"
          :disabled="!middleDate.fieldSearchType"
        />
      </el-form-item>
      <el-form-item :label="$t('tools.autoCode.fieldDialog.labelErrorText')">
        <el-input v-model="middleDate.errorText" />
      </el-form-item>
    </el-form>
    <el-collapse v-model="activeNames">
      <el-collapse-item
        :title="$t('tools.autoCode.fieldDialog.dataSourceCollapse')"
        name="1"
      >
        <el-row :gutter="8">
          <el-col :span="4">
            <el-select
              v-model="middleDate.dataSource.dbName"
              :placeholder="$t('tools.autoCode.fieldDialog.phDbMain')"
              @change="dbNameChange"
              clearable
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
          </el-col>
          <el-col :span="4">
            <el-select
              v-model="middleDate.dataSource.association"
              :placeholder="$t('tools.autoCode.fieldDialog.associationMode')"
              @change="associationChange"
            >
              <el-option :label="$t('tools.autoCode.fieldDialog.oneToOne')" :value="1" />
              <el-option :label="$t('tools.autoCode.fieldDialog.oneToMany')" :value="2" />
            </el-select>
          </el-col>
          <el-col :span="5">
            <el-select
              v-model="middleDate.dataSource.table"
              :placeholder="$t('tools.autoCode.fieldDialog.phPickDataTable')"
              filterable
              allow-create
              clearable
              @focus="getDBTableList"
              @change="selectDB"
              @clear="clearAccress"
            >
              <el-option
                v-for="item in dbTableList"
                :key="item.tableName"
                :label="item.tableName"
                :value="item.tableName"
              />
            </el-select>
          </el-col>
          <el-col :span="5">
            <el-select
              v-model="middleDate.dataSource.value"
              :placeholder="$t('tools.autoCode.fieldDialog.phPickStoreValue')"
            >
              <template #label="{ value }">
                <span>{{ $t('tools.autoCode.fieldDialog.storePrefix') }}</span>
                <span style="font-weight: bold">{{ value }}</span>
              </template>
              <el-option
                v-for="item in dbColumnList"
                :key="item.columnName"
                :value="item.columnName"
              >
                <span style="float: left">
                  <el-tag :type="item.isPrimary ? 'primary' : 'info'">
                    {{ item.isPrimary ? $t('tools.autoCode.fieldDialog.tagPk') : $t('tools.autoCode.fieldDialog.tagNonPk') }}
                  </el-tag>
                  {{ item.columnName }}</span
                >
                <span
                  style="
                    float: right;
                    margin-left: 5px;
                    color: var(--el-text-color-secondary);
                    font-size: 13px;
                  "
                >
                  {{ $t('tools.autoCode.fieldDialog.metaType') }}{{ item.type }}
                  <span v-if="item.comment != ''"
                    >{{ $t('tools.autoCode.fieldDialog.commaBeforeComment') }}{{ $t('tools.autoCode.fieldDialog.metaComment') }}{{ item.comment }}</span
                  >
                </span>
              </el-option>
            </el-select>
          </el-col>
          <el-col :span="5">
            <el-select
              v-model="middleDate.dataSource.label"
              :placeholder="$t('tools.autoCode.fieldDialog.phPickDisplayValue')"
            >
              <template #label="{ value }">
                <span>{{ $t('tools.autoCode.fieldDialog.displayPrefix') }}</span>
                <span style="font-weight: bold">{{ value }}</span>
              </template>
              <el-option
                v-for="item in dbColumnList"
                :key="item.columnName"
                :value="item.columnName"
              >
                <span style="float: left">
                  <el-tag :type="item.isPrimary ? 'primary' : 'info'">
                    {{ item.isPrimary ? $t('tools.autoCode.fieldDialog.tagPk') : $t('tools.autoCode.fieldDialog.tagNonPk') }}
                  </el-tag>
                  {{ item.columnName }}</span
                >
                <span
                  style="
                    float: right;
                    margin-left: 5px;
                    color: var(--el-text-color-secondary);
                    font-size: 13px;
                  "
                >
                  {{ $t('tools.autoCode.fieldDialog.metaType') }}{{ item.type }}
                  <span v-if="item.comment != ''"
                    >{{ $t('tools.autoCode.fieldDialog.commaBeforeComment') }}{{ $t('tools.autoCode.fieldDialog.metaComment') }}{{ item.comment }}</span
                  >
                </span>
              </el-option>
            </el-select>
            <!-- <el-input v-model="middleDate.dataSource.label" placeholder="展示用字段" /> -->
          </el-col>
        </el-row>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
  import { toLowerCase, toSQLLine } from '@/utils/stringFun'
  import { getSysDictionaryList } from '@/api/sysDictionary'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, onMounted, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessageBox } from 'element-plus'
  import { getColumn, getDB, getTable } from '@/api/autoCode'

  defineOptions({
    name: 'FieldDialog'
  })

  const props = defineProps({
    dialogMiddle: {
      type: Object,
      default: function () {
        return {}
      }
    },
    typeOptions: {
      type: Array,
      default: function () {
        return []
      }
    },
    typeSearchOptions: {
      type: Array,
      default: function () {
        return []
      }
    },
    typeIndexOptions: {
      type: Array,
      default: function () {
        return []
      }
    }
  })

  const { t } = useI18n()

  const activeNames = ref([])

  const middleDate = ref({})
  const dictOptions = ref([])

  const dbList = ref([])

  const getDbFunc = async () => {
    const res = await getDB()
    if (res.code === 0) {
      dbList.value = res.data.dbList
    }
  }

  const validateDataTypeLong = (rule, value, callback) => {
    const regex = /^('([^']*)'(?:,'([^']+)'*)*)$/
    if (middleDate.value.fieldType == 'enum' && !regex.test(value)) {
      callback(new Error(t('tools.autoCode.fieldDialog.errEnumFormat')))
    } else {
      callback()
    }
  }

  const rules = computed(() => ({
    fieldName: [
      { required: true, message: t('tools.autoCode.fieldDialog.ruleFieldName'), trigger: 'blur' }
    ],
    fieldDesc: [
      { required: true, message: t('tools.autoCode.fieldDialog.ruleFieldDesc'), trigger: 'blur' }
    ],
    fieldJson: [
      { required: true, message: t('tools.autoCode.fieldDialog.ruleFieldJson'), trigger: 'blur' }
    ],
    columnName: [
      { required: true, message: t('tools.autoCode.fieldDialog.ruleColumnName'), trigger: 'blur' }
    ],
    fieldType: [{ required: true, message: t('tools.autoCode.fieldDialog.ruleFieldType'), trigger: 'blur' }],
    dataTypeLong: [{ validator: validateDataTypeLong, trigger: 'blur' }]
  }))

  const init = async () => {
    middleDate.value = props.dialogMiddle
    const dictRes = await getSysDictionaryList({
      page: 1,
      pageSize: 999999
    })

    dictOptions.value = dictRes.data
  }
  init()

  const autoFill = () => {
    middleDate.value.fieldJson = toLowerCase(middleDate.value.fieldName)
    middleDate.value.columnName = toSQLLine(middleDate.value.fieldJson)
  }

  const canSelect = (item) => {
    const fieldType = middleDate.value.fieldType;

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

  const clearOther = () => {
    middleDate.value.fieldSearchType = ''
    middleDate.value.dictType = ''
  }

  const associationChange = (val) => {
    if (val === 2) {
      ElMessageBox.confirm(
        t('tools.autoCode.fieldDialog.confirmOneToMany'),
        t('common.tipTitle'),
        {
          confirmButtonText: t('tools.autoCode.fieldDialog.btnContinue'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
        .then(() => {
          middleDate.value.fieldType = 'array'
        })
        .catch(() => {
          middleDate.value.dataSource.association = 1
        })
    }
  }

  const clearAccress = () => {
    middleDate.value.dataSource.value = ''
    middleDate.value.dataSource.label = ''
  }

  const clearDataSourceTable = () => {
    middleDate.value.dataSource.table = ''
  }

  const dbNameChange = () => {
    getDBTableList()
    clearDataSourceTable()
    clearAccress()
  }

  const dbTableList = ref([])

  const getDBTableList = async () => {
    const res = await getTable({
      businessDB: middleDate.value.dataSource.dbName
    })
    if (res.code === 0) {
      let list = res.data.tables // 确保这里正确获取到 tables 数组
      dbTableList.value = list.map((item) => ({
        tableName: item.tableName,
        value: item.tableName // 这里假设 value 也是 tableName，如果不同请调整
      }))
    }
    clearAccress()
  }

  const dbColumnList = ref([])
  const selectDB = async (val, isInit) => {
    middleDate.value.dataSource.hasDeletedAt = false
    middleDate.value.dataSource.table = val
    const res = await getColumn({
      businessDB: middleDate.value.dataSource.dbName,
      tableName: val
    })

    if (res.code === 0) {
      let list = res.data.columns // 确保这里正确获取到 tables 数组
      dbColumnList.value = list.map((item) => {
        if (item.columnName === 'deleted_at') {
          middleDate.value.dataSource.hasDeletedAt = true
        }
        return {
          columnName: item.columnName,
          value: item.columnName,
          type: item.dataType,
          isPrimary: item.primaryKey,
          comment: item.columnComment
        }
      })
      if (dbColumnList.value.length > 0 && !isInit) {
        middleDate.value.dataSource.label = dbColumnList.value[0].columnName
        middleDate.value.dataSource.value = dbColumnList.value[0].columnName
      }
    }
  }

  const fieldDialogForm = ref(null)
  defineExpose({ fieldDialogForm })

  onMounted(() => {
    getDbFunc()
    if (middleDate.value.dataSource.table) {
      selectDB(middleDate.value.dataSource.table, true)
    }
  })
</script>
