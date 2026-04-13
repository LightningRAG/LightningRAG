<template>
  <div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list justify-between flex items-center">
        <span class="text font-bold">{{ $t('admin.dictionaryDetail.detailTitle') }}</span>
        <div class="flex items-center gap-2">
          <el-input
            :placeholder="$t('admin.dictionaryDetail.searchLabelPh')"
            v-model="searchName"
            clearable
            class="!w-64"
            @clear="clearSearchInput"
            :prefix-icon="Search"
            v-click-outside="handleCloseSearchInput"
            @keydown="handleInputKeyDown"
          >
            <template #append>
              <el-button
                :type="searchName ? 'primary' : 'info'"
                @click="applySearch"
                >{{ $t('admin.dictionary.btnSearch') }}</el-button
              >
            </template>
          </el-input>
          <el-button type="primary" icon="plus" @click="openDrawer">
            {{ $t('admin.dictionaryDetail.addEntry') }}
          </el-button>
        </div>
      </div>
      <!-- 表格视图 -->
      <el-table
        :data="displayTreeData"
        style="width: 100%"
        tooltip-effect="dark"
        :tree-props="{ children: 'children'}"
        row-key="ID"
        default-expand-all
      >
        <el-table-column type="selection" width="55" />

        <el-table-column align="left" :label="$t('admin.dictionaryDetail.colLabel')" prop="label" min-width="100"/>

        <el-table-column align="left" :label="$t('admin.dictionaryDetail.colValue')" prop="value" />

        <el-table-column align="left" :label="$t('admin.dictionaryDetail.colExtend')" prop="extend" />

        <el-table-column align="left" :label="$t('admin.dictionaryDetail.colLevel')" prop="level" width="80" />

        <el-table-column
          align="left"
          :label="$t('admin.dictionaryDetail.colEnabled')"
          prop="status"
          width="100"
        >
          <template #default="scope">
            {{ formatBoolI18n(scope.row.status) }}
          </template>
        </el-table-column>

        <el-table-column
          align="left"
          :label="$t('common.colDetailSort')"
          prop="sort"
          width="100"
        />

        <el-table-column
          align="left"
          :label="$t('common.colActions')"
          :min-width="appStore.operateMinWith"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="plus"
              @click="addChildNode(scope.row)"
            >
              {{ $t('admin.dictionaryDetail.addChildEntry') }}
            </el-button>
            <el-button
              type="primary"
              link
              icon="edit"
              @click="updateSysDictionaryDetailFunc(scope.row)"
            >
              {{ $t('common.change') }}
            </el-button>
            <el-button
              type="primary"
              link
              icon="delete"
              @click="deleteSysDictionaryDetailFunc(scope.row)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-drawer
      v-model="drawerFormVisible"
      :size="appStore.drawerSize"
      :show-close="false"
      :before-close="closeDrawer"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{
            type === 'create' ? $t('admin.dictionaryDetail.drawerAddEntry') : $t('admin.dictionaryDetail.drawerEditEntry')
          }}</span>
          <div>
            <el-button @click="closeDrawer"> {{ $t('common.cancel') }} </el-button>
            <el-button type="primary" @click="enterDrawer"> {{ $t('common.ok') }} </el-button>
          </div>
        </div>
      </template>
      <el-form
        ref="drawerForm"
        :model="formData"
        :rules="rules"
        label-width="110px"
      >
        <el-form-item :label="$t('admin.dictionaryDetail.parentEntry')" prop="parentID">
          <el-cascader
            v-model="formData.parentID"
            :options="[rootOption,...treeData]"
            :props="cascadeProps"
            :placeholder="$t('admin.dictionaryDetail.phParentEntry')"
            clearable
            filterable
            :style="{ width: '100%' }"
            @change="handleParentChange"
          />
        </el-form-item>
        <el-form-item :label="$t('admin.dictionaryDetail.colLabel')" prop="label">
          <el-input
            v-model="formData.label"
            :placeholder="$t('admin.dictionaryDetail.phLabel')"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item :label="$t('admin.dictionaryDetail.colValue')" prop="value">
          <el-input
            v-model="formData.value"
            :placeholder="$t('admin.dictionaryDetail.phValue')"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item :label="$t('admin.dictionaryDetail.colExtend')" prop="extend">
          <el-input
            v-model="formData.extend"
            :placeholder="$t('admin.dictionaryDetail.phExtend')"
            clearable
            :style="{ width: '100%' }"
          />
        </el-form-item>
        <el-form-item :label="$t('admin.dictionaryDetail.colEnabled')" prop="status" required>
          <el-switch
            v-model="formData.status"
            :active-text="$t('admin.dictionary.statusOn')"
            :inactive-text="$t('admin.dictionary.statusOff')"
          />
        </el-form-item>
        <el-form-item :label="$t('common.colDetailSort')" prop="sort">
          <el-input-number
            v-model.number="formData.sort"
            :placeholder="$t('admin.dictionaryDetail.phSort')"
          />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createSysDictionaryDetail,
    deleteSysDictionaryDetail,
    updateSysDictionaryDetail,
    findSysDictionaryDetail,
    getDictionaryTreeList
  } from '@/api/sysDictionaryDetail' // 此处请自行替换地址
  import { ref, watch, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import { Search } from '@element-plus/icons-vue'

  defineOptions({
    name: 'SysDictionaryDetail'
  })

  const { t } = useI18n()
  const appStore = useAppStore()

  const formatBoolI18n = (bool) => {
    if (bool === null || bool === undefined) return ''
    return bool ? t('common.yes') : t('common.no')
  }
  const searchName = ref('')

  const props = defineProps({
    sysDictionaryID: {
      type: Number,
      default: 0
    }
  })

  const formData = ref({
    label: null,
    value: null,
    status: true,
    sort: null,
    parentID: null
  })

  const rules = computed(() => ({
    label: [
      {
        required: true,
        message: t('admin.dictionaryDetail.ruleLabel'),
        trigger: 'blur'
      }
    ],
    value: [
      {
        required: true,
        message: t('admin.dictionaryDetail.ruleValue'),
        trigger: 'blur'
      }
    ],
    sort: [
      {
        required: true,
        message: t('admin.dictionaryDetail.ruleSort'),
        trigger: 'blur'
      }
    ]
  }))

  const treeData = ref([])
  const displayTreeData = ref([])

  // 级联选择器配置
  const cascadeProps = {
    value: 'ID',
    label: 'label',
    children: 'children',
    checkStrictly: true, // 允许选择任意级别
    emitPath: false // 只返回选中节点的值
  }


  const normalizeSearch = (value) => (value ?? '').toString().toLowerCase()

  const filterTree = (nodes, keyword) => {
    const trimmed = normalizeSearch(keyword).trim()
    if (!trimmed) {
      return nodes
    }
    const walk = (list) => {
      const result = []
      for (const node of list) {
        const label = normalizeSearch(node.label)
        const children = Array.isArray(node.children) ? walk(node.children) : []
        if (label.includes(trimmed) || children.length > 0) {
          result.push({
            ...node,
            children
          })
        }
      }
      return result
    }
    return walk(nodes)
  }

  const applySearch = () => {
    displayTreeData.value = filterTree(treeData.value, searchName.value)
  }

  // 获取树形数据
  const getTreeData = async () => {
    if (!props.sysDictionaryID) return
    try {
      const res = await getDictionaryTreeList({
        sysDictionaryID: props.sysDictionaryID
      })
      if (res.code === 0) {
        treeData.value = res.data.list || []
        applySearch()
      }
    } catch (error) {
      console.error('Failed to load tree data:', error)
      ElMessage.error(t('admin.dictionaryDetail.treeLoadFail'))
    }
  }

  const rootOption = computed(() => ({
    ID: null,
    label: t('admin.dictionaryDetail.rootOption')
  }))


  // 初始加载
  getTreeData()

  const type = ref('')
  const drawerFormVisible = ref(false)

  const updateSysDictionaryDetailFunc = async (row) => {
    drawerForm.value && drawerForm.value.clearValidate()
    const res = await findSysDictionaryDetail({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data.reSysDictionaryDetail
      drawerFormVisible.value = true
    }
  }

  // 添加子节点
  const addChildNode = (parentNode) => {
    type.value = 'create'
    formData.value = {
      label: null,
      value: null,
      status: true,
      sort: null,
      parentID: parentNode.ID,
      sysDictionaryID: props.sysDictionaryID
    }
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
  }

  // 处理父级选择变化
  const handleParentChange = (value) => {
    formData.value.parentID = value
  }

  const closeDrawer = () => {
    drawerFormVisible.value = false
    formData.value = {
      label: null,
      value: null,
      status: true,
      sort: null,
      parentID: null,
      sysDictionaryID: props.sysDictionaryID
    }
  }

  const deleteSysDictionaryDetailFunc = async (row) => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteSysDictionaryDetail({ ID: row.ID })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('common.deleteOk')
        })
        await getTreeData() // 重新加载数据
      }
    })
  }

  const drawerForm = ref(null)
  const enterDrawer = async () => {
    drawerForm.value.validate(async (valid) => {
      formData.value.sysDictionaryID = props.sysDictionaryID
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createSysDictionaryDetail(formData.value)
          break
        case 'update':
          res = await updateSysDictionaryDetail(formData.value)
          break
        default:
          res = await createSysDictionaryDetail(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('common.createUpdateOk')
        })
        closeDrawer()
        await getTreeData() // 重新加载数据
      }
    })
  }

  const openDrawer = () => {
    type.value = 'create'
    formData.value.parentID = null
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
  }

  const clearSearchInput = () => {
    searchName.value = ''
    applySearch()
  }

  const handleCloseSearchInput = () => {
    // 处理搜索输入框关闭
  }

  const handleInputKeyDown = (e) => {
    if (e.key === 'Enter') {
      applySearch()
    }
  }

  watch(
    () => props.sysDictionaryID,
    () => {
      getTreeData()
    }
  )
</script>

<style scoped>

</style>
