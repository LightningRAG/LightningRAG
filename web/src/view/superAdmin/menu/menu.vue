<template>
  <div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="addMenu(0)">
          {{ $t('admin.menu.addRoot') }}
        </el-button>
      </div>

      <!-- 由于此处菜单跟左侧列表一一对应所以不需要分页 pageSize默认999 -->
      <el-table :data="tableData" row-key="ID">
        <el-table-column align="left" :label="$t('common.colId')" min-width="100" prop="ID" />
        <el-table-column
          align="left"
          :label="$t('admin.menu.colDisplayTitle')"
          min-width="120"
          prop="authorityName"
        >
          <template #default="scope">
            <span>{{ resolveMenuTitle(scope.row.meta, route) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('admin.menu.colIcon')"
          min-width="140"
          prop="authorityName"
        >
          <template #default="scope">
            <div v-if="scope.row.meta.icon" class="icon-column">
              <el-icon>
                <component :is="scope.row.meta.icon" />
              </el-icon>
              <span>{{ scope.row.meta.icon }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('admin.menu.colRouteName')"
          show-overflow-tooltip
          min-width="160"
          prop="name"
        />
        <el-table-column
          align="left"
          :label="$t('admin.menu.colRoutePath')"
          show-overflow-tooltip
          min-width="160"
          prop="path"
        />
        <el-table-column
          align="left"
          :label="$t('admin.menu.colVisible')"
          min-width="100"
          prop="hidden"
        >
          <template #default="scope">
            <span>{{ scope.row.hidden ? $t('admin.menu.hiddenYes') : $t('admin.menu.hiddenNo') }}</span>
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          :label="$t('admin.menu.colParent')"
          min-width="90"
          prop="parentId"
        />
        <el-table-column align="left" :label="$t('admin.menu.colSort')" min-width="70" prop="sort" />
        <el-table-column
          align="left"
          :label="$t('admin.menu.colFilePath')"
          min-width="360"
          prop="component"
        />
        <el-table-column align="left" fixed="right" :label="$t('common.colActions')" :min-width="appStore.operateMinWith">
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="plus"
              @click="addMenu(scope.row.ID)"
            >
              {{ $t('admin.menu.addChild') }}
            </el-button>
            <el-button
              type="primary"
              link
              icon="edit"
              @click="editMenu(scope.row.ID)"
            >
              {{ $t('common.edit') }}
            </el-button>
            <el-button
              type="primary"
              link
              icon="user"
              @click="openAssignRoleDrawer(scope.row)"
            >
              {{ $t('admin.menu.assignRole') }}
            </el-button>
            <el-button
              type="primary"
              link
              icon="delete"
              @click="deleteMenu(scope.row.ID)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-drawer
      v-model="dialogFormVisible"
      :size="appStore.drawerSize"
      :before-close="handleClose"
      :show-close="false"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ dialogTitle }}</span>
          <div>
            <el-button @click="closeDialog">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="enterDialog">{{ $t('common.ok') }}</el-button>
          </div>
        </div>
      </template>

      <warning-bar :title="$t('admin.menu.warningNewMenu')" />
      
      <!-- 基础信息区域 -->
      <div class="border-b border-gray-200">
        <h3 class="font-semibold text-gray-700 mb-4">{{ $t('admin.menu.sectionBasic') }}</h3>
        <el-form
          v-if="dialogFormVisible"
          ref="menuForm"
          :inline="true"
          :model="form"
          :rules="rules"
          label-position="top"
        >
          <el-row class="w-full">
            <el-col :span="24">
              <el-form-item :label="$t('admin.menu.colFilePath')" prop="component">
                <components-cascader
                  :component="form.component"
                  @change="fmtComponent"
                />
                <div class="form-tip">
                  <el-icon><InfoFilled /></el-icon>
                  <span>{{ $t('admin.menu.componentPathTip') }}</span>
                  <el-button
                    size="small"
                    type="text"
                    @click="form.component = 'view/routerHolder.vue'"
                  >
                    {{ $t('admin.menu.setRouterHolder') }}
                  </el-button>
                </div>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row class="w-full">
            <el-col :span="12">
              <el-form-item :label="$t('admin.menu.colDisplayTitle')" prop="meta.title">
                <el-input 
                  v-model="form.meta.title" 
                  autocomplete="off" 
                  :placeholder="$t('admin.menu.phDisplayTitle')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('admin.menu.colRouteName')" prop="path">
                <el-input
                  v-model="form.name"
                  autocomplete="off"
                  :placeholder="$t('admin.menu.phRouteNameUnique')"
                  @change="changeName"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </div>
       
      <!-- 路由配置区域 -->
      <div class="border-b border-gray-200">
        <h3 class="font-semibold text-gray-700 mb-4">{{ $t('admin.menu.sectionRoute') }}</h3>
        <el-form
          :inline="true"
          :model="form"
          :rules="rules"
          label-position="top"
        >
           <el-row class="w-full">
             <el-col :span="12">
               <el-form-item :label="$t('admin.menu.labelParentNodeId')">
                 <el-cascader
                   v-model="form.parentId"
                   style="width: 100%"
                   :disabled="!isEdit"
                   :options="menuOption"
                   :props="{
                     checkStrictly: true,
                     label: 'title',
                     value: 'ID',
                     disabled: 'disabled',
                     emitPath: false
                   }"
                   :show-all-levels="false"
                   filterable
                   :placeholder="$t('admin.menu.phSelectParent')"
                 />
               </el-form-item>
             </el-col>
             <el-col :span="12">
               <el-form-item prop="path">
                 <template #label>
                  <div class="inline-flex items-center h-4">
                     <span>{{ $t('admin.menu.routePathLabel') }}</span>
                     <el-checkbox
                       class="ml-2"
                       v-model="checkFlag"
                       >{{ $t('admin.menu.addQueryParams') }}</el-checkbox
                     >
                    </div>
                 </template>
                 <el-input
                   v-model="form.path"
                   :disabled="!checkFlag"
                   autocomplete="off"
                   :placeholder="$t('admin.menu.phAppendParamsHint')"
                 />
               </el-form-item>
             </el-col>
           </el-row>
        </el-form>
      </div>
       
      <!-- 显示设置区域 -->
      <div class="border-b border-gray-200">
        <h3 class="font-semibold text-gray-700 mb-4">{{ $t('admin.menu.sectionDisplay') }}</h3>
        <el-form
          :inline="true"
          :model="form"
          :rules="rules"
          label-position="top"
        >
           <el-row class="w-full">
              <el-col :span="8">
                <el-form-item :label="$t('admin.menu.colIcon')" prop="meta.icon">
                  <icon v-model="form.meta.icon" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item :label="$t('admin.menu.colSortBadge')" prop="sort">
                  <el-input 
                    v-model.number="form.sort" 
                    autocomplete="off" 
                    :placeholder="$t('admin.menu.phSortNumber')"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item :label="$t('admin.menu.labelHiddenInList')">
                  <el-select
                    v-model="form.hidden"
                    style="width: 100%"
                    :placeholder="$t('admin.menu.phHiddenInList')"
                  >
                    <el-option :value="false" :label="$t('common.no')" />
                    <el-option :value="true" :label="$t('common.yes')" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
        </el-form>
      </div>
        
      <!-- 高级配置区域 -->
      <div class="border-b border-gray-200">
        <h3 class="font-semibold text-gray-700 mb-4">{{ $t('admin.menu.sectionAdvanced') }}</h3>
        <el-form
          :inline="true"
          :model="form"
          :rules="rules"
          label-position="top"
        >
            <el-row class="w-full">
              <el-col :span="12">
                <el-form-item prop="meta.activeName">
                  <template #label>
                    <div class="label-with-tooltip">
                      <span>{{ $t('admin.menu.colActiveMenu') }}</span>
                      <el-tooltip
                        :content="$t('admin.menu.tooltipActiveMenu')"
                        placement="top"
                        effect="light"
                      >
                        <el-icon><QuestionFilled /></el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input
                    v-model="form.meta.activeName"
                    :placeholder="form.name || $t('admin.menu.phActiveMenuName')"
                    autocomplete="off"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item :label="$t('admin.menu.labelKeepAlive')" prop="meta.keepAlive">
                  <el-select
                    v-model="form.meta.keepAlive"
                    style="width: 100%"
                    :placeholder="$t('admin.menu.phKeepAlive')"
                  >
                    <el-option :value="false" :label="$t('common.no')" />
                    <el-option :value="true" :label="$t('common.yes')" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
             <el-row class="w-full">
               <el-col :span="8">
                 <el-form-item :label="$t('admin.menu.labelCloseTab')" prop="meta.closeTab">
                   <el-select
                     v-model="form.meta.closeTab"
                     style="width: 100%"
                     :placeholder="$t('admin.menu.phCloseTab')"
                   >
                     <el-option :value="false" :label="$t('common.no')" />
                     <el-option :value="true" :label="$t('common.yes')" />
                   </el-select>
                 </el-form-item>
               </el-col>
               <el-col :span="8">
                 <el-form-item>
                   <template #label>
                     <div class="label-with-tooltip">
                       <span>{{ $t('admin.menu.labelBasePage') }}</span>
                       <el-tooltip
                         :content="$t('admin.menu.tooltipBasePage')"
                         placement="top"
                         effect="light"
                       >
                         <el-icon><QuestionFilled /></el-icon>
                       </el-tooltip>
                     </div>
                   </template>
                   <el-select
                     v-model="form.meta.defaultMenu"
                     style="width: 100%"
                     :placeholder="$t('admin.menu.phBasePage')"
                   >
                     <el-option :value="false" :label="$t('common.no')" />
                     <el-option :value="true" :label="$t('common.yes')" />
                   </el-select>
                 </el-form-item>
               </el-col>
               <el-col :span="8">
                 <el-form-item>
                   <template #label>
                     <div class="label-with-tooltip">
                       <span>{{ $t('admin.menu.labelRouteTransition') }}</span>
                       <el-tooltip
                         :content="$t('admin.menu.tooltipRouteTransition')"
                         placement="top"
                         effect="light"
                       >
                         <el-icon><QuestionFilled /></el-icon>
                       </el-tooltip>
                     </div>
                   </template>
                   <el-select
                     v-model="form.meta.transitionType"
                     style="width: 100%"
                     :placeholder="$t('admin.menu.phTransitionGlobal')"
                     clearable
                   >
                     <el-option value="fade" :label="$t('settings.layoutPrefs.animFade')" />
                     <el-option value="slide" :label="$t('settings.layoutPrefs.animSlide')" />
                     <el-option value="zoom" :label="$t('settings.layoutPrefs.animZoom')" />
                     <el-option value="none" :label="$t('settings.layoutPrefs.animNone')" />
                   </el-select>
                 </el-form-item>
               </el-col>
             </el-row>
        </el-form>
      </div>
          
      <!-- 菜单参数配置区域 -->
      <div class="border-b border-gray-200">
        <div class="flex justify-between items-center mb-4">
          <h3 class="font-semibold text-gray-700">{{ $t('admin.menu.sectionParams') }}</h3>
          <el-button type="primary" size="small" @click="addParameter(form)">
            {{ $t('admin.menu.btnAddParam') }}
          </el-button>
        </div>
            <el-table 
              :data="form.parameters" 
              style="width: 100%"
              class="parameter-table"
            >
              <el-table-column
                align="center"
                prop="type"
                :label="$t('admin.menu.colParamType')"
                width="150"
              >
                <template #default="scope">
                  <el-select 
                    v-model="scope.row.type" 
                    :placeholder="$t('common.phSelect')"
                    size="small"
                  >
                    <el-option key="query" value="query" label="query" />
                    <el-option key="params" value="params" label="params" />
                  </el-select>
                </template>
              </el-table-column>
              <el-table-column align="center" prop="key" :label="$t('common.colParamKey')" width="150">
                <template #default="scope">
                  <el-input 
                    v-model="scope.row.key" 
                    size="small"
                    :placeholder="$t('admin.menu.phParamKeyInput')"
                  />
                </template>
              </el-table-column>
              <el-table-column align="center" prop="value" :label="$t('common.colParamValue')">
                <template #default="scope">
                  <el-input 
                    v-model="scope.row.value" 
                    size="small"
                    :placeholder="$t('admin.menu.phParamValueInput')"
                  />
                </template>
              </el-table-column>
              <el-table-column align="center" :label="$t('common.colActions')" width="100">
                <template #default="scope">
                  <el-button
                    type="danger"
                    size="small"
                    @click="deleteParameter(form.parameters, scope.$index)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
      </div>
           
      <!-- 可控按钮配置区域 -->
      <div class="mb-2 mt-2">
        <div class="flex justify-between items-center mb-4">
          <h3 class="font-semibold text-gray-700">{{ $t('admin.menu.sectionButtons') }}</h3>
          <div class="flex items-center gap-2">
            <el-button type="primary" size="small" @click="addBtn(form)">
              {{ $t('admin.menu.btnAddMenuBtn') }}
            </el-button>
            <el-tooltip
              :content="$t('admin.menu.tooltipBtnAuthDoc')"
              placement="top"
              effect="light"
            >
              <el-icon
                class="cursor-pointer text-blue-500 hover:text-blue-700"
                @click="toDoc('https://lightningrag.com/guide/web/button-auth.html')"
              >
                <QuestionFilled />
              </el-icon>
            </el-tooltip>
          </div>
        </div>
             <el-table 
               :data="form.menuBtn" 
               style="width: 100%"
               class="button-table"
             >
               <el-table-column
                 align="center"
                 prop="name"
                 :label="$t('admin.menu.colBtnName')"
                 width="150"
               >
                 <template #default="scope">
                   <el-input 
                     v-model="scope.row.name" 
                     size="small"
                     :placeholder="$t('admin.menu.phBtnName')"
                   />
                 </template>
               </el-table-column>
               <el-table-column align="center" prop="desc" :label="$t('common.colRemark')">
                 <template #default="scope">
                   <el-input 
                     v-model="scope.row.desc" 
                     size="small"
                     :placeholder="$t('admin.menu.phBtnRemark')"
                   />
                 </template>
               </el-table-column>
               <el-table-column align="center" :label="$t('common.colActions')" width="100">
                 <template #default="scope">
                   <el-button
                     type="danger"
                     size="small"
                     @click="deleteBtn(form.menuBtn, scope.$index)"
                   >
                     <el-icon><Delete /></el-icon>
                   </el-button>
                 </template>
               </el-table-column>
             </el-table>
       </div>
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
            $t('admin.menu.assignRoleTitle', { title: assignMenuRow.meta?.title || '' })
          }}</span>
          <div>
            <el-button @click="assignRoleDrawerVisible = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" :loading="assignRoleSubmitting" @click="confirmAssignRole">{{ $t('common.ok') }}</el-button>
          </div>
        </div>
      </template>
      <warning-bar :title="$t('admin.menu.assignRoleWarningBar')" />
      <el-tree
        ref="roleTreeRef"
        v-loading="assignRoleLoading"
        :data="authorityTreeData"
        :props="{ label: 'authorityName', children: 'children', disabled: isRoleDisabled }"
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
    updateBaseMenu,
    getMenuList,
    addBaseMenu,
    deleteBaseMenu,
    getBaseMenuById,
    getMenuRoles,
    setMenuRoles
  } from '@/api/menu'
  import { getAuthorityList } from '@/api/authority'
  import icon from '@/view/superAdmin/menu/icon.vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { canRemoveAuthorityBtnApi } from '@/api/authorityBtn'
  import { ref, nextTick, computed } from 'vue'
  import { useRoute } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { resolveMenuTitle } from '@/utils/menuTitle'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { QuestionFilled, InfoFilled, Delete } from '@element-plus/icons-vue'
  import { toDoc } from '@/utils/doc'
  import { toLowerCase } from '@/utils/stringFun'
  import { authorityDisplayName } from '@/utils/authorityI18n'
  import ComponentsCascader from '@/view/superAdmin/menu/components/components-cascader.vue'

  import pathInfo from '@/pathInfo.json'
  import { useAppStore } from "@/pinia";

  defineOptions({
    name: 'Menus'
  })

  const appStore = useAppStore()
  const route = useRoute()
  const { t } = useI18n()

  const rules = computed(() => ({
    path: [{ required: true, message: t('admin.menu.ruleMenuPath'), trigger: 'blur' }],
    component: [{ required: true, message: t('admin.menu.ruleComponentPath'), trigger: 'blur' }],
    'meta.title': [
      { required: true, message: t('admin.menu.ruleMenuTitle'), trigger: 'blur' }
    ]
  }))

  const tableData = ref([])
  // 查询
  const getTableData = async () => {
    const table = await getMenuList()
    if (table.code === 0) {
      tableData.value = table.data
    }
  }

  getTableData()

  // 新增参数
  const addParameter = (form) => {
    if (!form.parameters) {
      form.parameters = []
    }
    form.parameters.push({
      type: 'query',
      key: '',
      value: ''
    })
  }

  const fmtComponent = (component) => {
    form.value.component = component.replace(/\\/g, '/')
    form.value.name = toLowerCase(pathInfo['/src/' + component])
    form.value.path = form.value.name
  }

  // 删除参数
  const deleteParameter = (parameters, index) => {
    parameters.splice(index, 1)
  }

  // 新增可控按钮
  const addBtn = (form) => {
    if (!form.menuBtn) {
      form.menuBtn = []
    }
    form.menuBtn.push({
      name: '',
      desc: ''
    })
  }
  // 删除可控按钮
  const deleteBtn = async (btns, index) => {
    const btn = btns[index]
    if (btn.ID === 0) {
      btns.splice(index, 1)
      return
    }
    const res = await canRemoveAuthorityBtnApi({ id: btn.ID })
    if (res.code === 0) {
      btns.splice(index, 1)
    }
  }

  const form = ref({
    ID: 0,
    path: '',
    name: '',
    hidden: false,
    parentId: 0,
    component: '',
    meta: {
      activeName: '',
      title: '',
      icon: '',
      defaultMenu: false,
      closeTab: false,
      keepAlive: false
    },
    parameters: [],
    menuBtn: []
  })
  const changeName = () => {
    form.value.path = form.value.name
  }

  const handleClose = (done) => {
    initForm()
    done()
  }
  // 删除菜单
  const deleteMenu = (ID) => {
    ElMessageBox.confirm(
      t('admin.menu.confirmDeleteMenu'),
      t('common.tipTitle'),
      {
        confirmButtonText: t('common.ok'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
      .then(async () => {
        const res = await deleteBaseMenu({ ID })
        if (res.code === 0) {
          ElMessage({
            type: 'success',
            message: t('common.deleteOk')
          })

          getTableData()
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: t('admin.authority.deleteCancelled')
        })
      })
  }
  // 初始化弹窗内表格方法
  const menuForm = ref(null)
  const checkFlag = ref(false)
  const initForm = () => {
    checkFlag.value = false
    menuForm.value.resetFields()
    form.value = {
      ID: 0,
      path: '',
      name: '',
      hidden: false,
      parentId: 0,
      component: '',
      meta: {
        title: '',
        icon: '',
        defaultMenu: false,
        closeTab: false,
        keepAlive: false
      }
    }
  }
  // 关闭弹窗

  const dialogFormVisible = ref(false)
  const closeDialog = () => {
    initForm()
    dialogFormVisible.value = false
  }
  // 添加menu
  const enterDialog = async () => {
    menuForm.value.validate(async (valid) => {
      if (valid) {
        let res
        if (isEdit.value) {
          res = await updateBaseMenu(form.value)
        } else {
          res = await addBaseMenu(form.value)
        }
        if (res.code === 0) {
          ElMessage({
            type: 'success',
            message: isEdit.value
              ? t('admin.menu.editOkShort')
              : t('admin.menu.addOkNeedRoleAssign')
          })
          getTableData()
        }
        initForm()
        dialogFormVisible.value = false
      }
    })
  }

  const menuOption = ref([
    {
      ID: '0',
      title: ''
    }
  ])
  const setOptions = () => {
    menuOption.value = [
      {
        ID: 0,
        title: t('admin.menu.rootDirTitle')
      }
    ]
    setMenuOptions(tableData.value, menuOption.value, false)
  }
  const setMenuOptions = (menuData, optionsData, disabled) => {
    menuData &&
      menuData.forEach((item) => {
        if (item.children && item.children.length) {
          const option = {
            title: resolveMenuTitle(item.meta, route),
            ID: item.ID,
            disabled: disabled || item.ID === form.value.ID,
            children: []
          }
          setMenuOptions(
            item.children,
            option.children,
            disabled || item.ID === form.value.ID
          )
          optionsData.push(option)
        } else {
          const option = {
            title: resolveMenuTitle(item.meta, route),
            ID: item.ID,
            disabled: disabled || item.ID === form.value.ID
          }
          optionsData.push(option)
        }
      })
  }

  // 添加菜单方法，id为 0则为添加根菜单
  const isEdit = ref(false)
  const dialogTitle = ref('')
  const addMenu = (id) => {
    dialogTitle.value = t('admin.menu.drawerAddMenu')
    form.value.parentId = id
    isEdit.value = false
    setOptions()
    dialogFormVisible.value = true
  }
  // 修改菜单方法
  const editMenu = async (id) => {
    dialogTitle.value = t('admin.menu.drawerEditMenu')
    const res = await getBaseMenuById({ id })
    form.value = res.data.menu
    isEdit.value = true
    setOptions()
    dialogFormVisible.value = true
  }

  // 分配给角色
  const assignRoleDrawerVisible = ref(false)
  const assignMenuRow = ref({})
  const authorityTreeData = ref([])
  const assignRoleLoading = ref(false)
  const assignRoleSubmitting = ref(false)
  const roleTreeRef = ref(null)
  const defaultRouterAuthorityIds = ref(new Set())

  const isRoleDisabled = (data) => {
    return defaultRouterAuthorityIds.value.has(data.authorityId)
  }

  const openAssignRoleDrawer = async (row) => {
    assignMenuRow.value = row
    defaultRouterAuthorityIds.value = new Set()
    assignRoleDrawerVisible.value = true
    assignRoleLoading.value = true
    // 并行加载角色树和当前菜单已分配的角色
    const [authRes, rolesRes] = await Promise.all([
      getAuthorityList(),
      getMenuRoles(row.ID)
    ])
    if (authRes.code === 0) {
      authorityTreeData.value = authRes.data
    }
    if (rolesRes.code === 0 && rolesRes.data) {
      if (rolesRes.data.defaultRouterAuthorityIds) {
        defaultRouterAuthorityIds.value = new Set(rolesRes.data.defaultRouterAuthorityIds)
      }
      nextTick(() => {
        roleTreeRef.value?.setCheckedKeys(rolesRes.data.authorityIds || [])
      })
    }
    assignRoleLoading.value = false
  }

  const confirmAssignRole = async () => {
    assignRoleSubmitting.value = true
    try {
      const checkedKeys = roleTreeRef.value?.getCheckedKeys(false) || []
      const halfCheckedKeys = roleTreeRef.value?.getHalfCheckedKeys() || []
      const authorityIds = [...checkedKeys, ...halfCheckedKeys]
      const res = await setMenuRoles({
        menuId: assignMenuRow.value.ID,
        authorityIds
      })
      if (res.code === 0) {
        ElMessage({ type: 'success', message: t('admin.menu.assignOk') })
        assignRoleDrawerVisible.value = false
      }
    } catch {
      ElMessage({ type: 'error', message: t('admin.menu.assignFailRetry') })
    }
    assignRoleSubmitting.value = false
  }
</script>

<style scoped lang="scss">
  .warning {
    color: #dc143c;
  }
  .icon-column {
    display: flex;
    align-items: center;
    .el-icon {
      margin-right: 8px;
    }
  }


  
  .form-tip {
    margin-top: 8px;
    font-size: 12px;
    color: #909399;
    display: flex;
    align-items: center;
    gap: 8px;
    
    .el-icon {
      color: #409eff;
    }
  }
  
  .label-with-tooltip {
    display: flex;
    align-items: center;
    gap: 6px;
    
    .el-icon {
      color: #909399;
      cursor: help;
      
      &:hover {
        color: #409eff;
      }
    }
  }
  
  .parameter-table,
  .button-table {
    border: 1px solid #ebeef5;
    border-radius: 6px;
    
    :deep(.el-table__header) {
      background-color: #fafafa;
    }
    
    :deep(.el-table__body) {
      .el-table__row {
        &:hover {
          background-color: #f5f7fa;
        }
      }
    }
  }
</style>
