<template>
  <div>
    <warning-bar :title="$t('admin.user.warningBar')" />
    <div class="lrag-search-box">
      <el-form ref="searchForm" :inline="true" :model="searchInfo">
        <el-form-item :label="$t('common.colUsername')">
          <el-input v-model="searchInfo.username" :placeholder="$t('common.colUsername')" />
        </el-form-item>
        <el-form-item :label="$t('common.colNickname')">
          <el-input v-model="searchInfo.nickname" :placeholder="$t('common.colNickname')" />
        </el-form-item>
        <el-form-item :label="$t('common.colPhone')">
          <el-input v-model="searchInfo.phone" :placeholder="$t('common.colPhone')" />
        </el-form-item>
        <el-form-item :label="$t('common.colEmail')">
          <el-input v-model="searchInfo.email" :placeholder="$t('common.colEmail')" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">
            {{ $t('common.query') }}
          </el-button>
          <el-button icon="refresh" @click="onReset">{{ $t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="addUser"
          >{{ $t('common.addUser') }}</el-button
        >
      </div>
      <el-table :data="tableData" row-key="ID" :default-sort="{ prop: 'ID', order: 'descending' }" @sort-change="sortChange">
        <el-table-column align="left" :label="$t('common.colAvatar')" min-width="75">
          <template #default="scope">
            <CustomPic style="margin-top: 8px" :pic-src="scope.row.headerImg" />
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colId')" min-width="50" prop="ID" sortable="custom" />
        <el-table-column
          align="left"
          :label="$t('common.colUsername')"
          min-width="150"
          prop="userName"
        />
        <el-table-column
          align="left"
          :label="$t('common.colNickname')"
          min-width="150"
          prop="nickName"
        />
        <el-table-column
          align="left"
          :label="$t('common.colPhone')"
          min-width="180"
          prop="phone"
        />
        <el-table-column
          align="left"
          :label="$t('common.colEmail')"
          min-width="180"
          prop="email"
        />
        <el-table-column align="left" :label="$t('common.colUserRoles')" min-width="200">
          <template #default="scope">
            <el-cascader
              v-model="scope.row.authorityIds"
              :options="authOptions"
              :show-all-levels="false"
              collapse-tags
              :props="{
                multiple: true,
                checkStrictly: true,
                label: 'authorityName',
                value: 'authorityId',
                disabled: 'disabled',
                emitPath: false
              }"
              :clearable="false"
              @visible-change="
                (flag) => {
                  changeAuthority(scope.row, flag, 0)
                }
              "
              @remove-tag="
                (removeAuth) => {
                  changeAuthority(scope.row, false, removeAuth)
                }
              "
            />
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colEnabled')" min-width="150">
          <template #default="scope">
            <el-switch
              v-model="scope.row.enable"
              inline-prompt
              :active-value="1"
              :inactive-value="2"
              @change="
                () => {
                  switchEnable(scope.row)
                }
              "
            />
          </template>
        </el-table-column>

        <el-table-column :label="$t('common.colActions')" :min-width="appStore.operateMinWith" fixed="right">
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="delete"
              @click="deleteUserFunc(scope.row)"
              >{{ $t('common.delete') }}</el-button
            >
            <el-button
              type="primary"
              link
              icon="edit"
              @click="openEdit(scope.row)"
              >{{ $t('common.edit') }}</el-button
            >
            <el-button
              type="primary"
              link
              icon="magic-stick"
              @click="resetPasswordFunc(scope.row)"
              >{{ $t('admin.user.resetPassword') }}</el-button
            >
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
    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="resetPwdDialog"
      :title="$t('admin.user.resetPwdTitle')"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <el-form :model="resetPwdInfo" ref="resetPwdForm" label-width="100px">
        <el-form-item :label="$t('admin.user.userAccount')">
          <el-input v-model="resetPwdInfo.userName" disabled />
        </el-form-item>
        <el-form-item :label="$t('admin.user.userNickname')">
          <el-input v-model="resetPwdInfo.nickName" disabled />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <div class="flex w-full">
            <el-input class="flex-1" v-model="resetPwdInfo.password" :placeholder="$t('admin.user.phNewPassword')" show-password />
            <el-button type="primary" @click="generateRandomPassword" style="margin-left: 10px">
              {{ $t('admin.user.generateRandomPwd') }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeResetPwdDialog">{{ $t('settings.general.cancel') }}</el-button>
          <el-button type="primary" @click="confirmResetPassword">{{ $t('settings.general.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <el-drawer
      v-model="addUserDialog"
      :size="appStore.drawerSize"
      :show-close="false"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ $t('admin.user.drawerTitle') }}</span>
          <div>
            <el-button @click="closeAddUserDialog">{{ $t('settings.general.cancel') }}</el-button>
            <el-button type="primary" @click="enterAddUserDialog"
              >{{ $t('settings.general.confirm') }}</el-button
            >
          </div>
        </div>
      </template>

      <el-form
        ref="userForm"
        :rules="rules"
        :model="userInfo"
        label-width="80px"
      >
        <el-form-item
          v-if="dialogFlag === 'add'"
          :label="$t('common.colUsername')"
          prop="userName"
        >
          <el-input v-model="userInfo.userName" />
        </el-form-item>
        <el-form-item v-if="dialogFlag === 'add'" :label="$t('admin.user.password')" prop="password">
          <el-input v-model="userInfo.password" />
        </el-form-item>
        <el-form-item :label="$t('common.colNickname')" prop="nickName">
          <el-input v-model="userInfo.nickName" />
        </el-form-item>
        <el-form-item :label="$t('common.colPhone')" prop="phone">
          <el-input v-model="userInfo.phone" />
        </el-form-item>
        <el-form-item :label="$t('common.colEmail')" prop="email">
          <el-input v-model="userInfo.email" />
        </el-form-item>
        <el-form-item :label="$t('common.colUserRoles')" prop="authorityId">
          <el-cascader
            v-model="userInfo.authorityIds"
            style="width: 100%"
            :options="authOptions"
            :show-all-levels="false"
            :props="{
              multiple: true,
              checkStrictly: true,
              label: 'authorityName',
              value: 'authorityId',
              disabled: 'disabled',
              emitPath: false
            }"
            :clearable="false"
          />
        </el-form-item>
        <el-form-item :label="$t('common.colEnabled')" prop="disabled">
          <el-switch
            v-model="userInfo.enable"
            inline-prompt
            :active-value="1"
            :inactive-value="2"
          />
        </el-form-item>
        <el-form-item :label="$t('common.colAvatar')" label-width="80px">
          <SelectImage v-model="userInfo.headerImg" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getUserList,
    setUserAuthorities,
    register,
    deleteUser
  } from '@/api/user'

  import { getAuthorityList } from '@/api/authority'
  import CustomPic from '@/components/customPic/index.vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { setUserInfo, resetPassword } from '@/api/user.js'

  import { computed, nextTick, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import { useAppStore } from "@/pinia";
  import { toSQLLine } from '@/utils/stringFun'
  import { authorityDisplayName } from '@/utils/authorityI18n'

  defineOptions({
    name: 'User'
  })

  const appStore = useAppStore()
  const { t } = useI18n()

  const searchInfo = ref({
    username: '',
    nickname: '',
    phone: '',
    email: ''
  })

  const onSubmit = () => {
    page.value = 1
    getTableData()
  }

  const onReset = () => {
    searchInfo.value = {
      username: '',
      nickname: '',
      phone: '',
      email: ''
    }
    orderKey.value = 'id'
    desc.value = true
    getTableData()
  }
  // 初始化相关
  const setAuthorityOptions = (AuthorityData, optionsData) => {
    AuthorityData &&
      AuthorityData.forEach((item) => {
        if (item.children && item.children.length) {
          const option = {
            authorityId: item.authorityId,
            authorityName: authorityDisplayName(item, t),
            children: []
          }
          setAuthorityOptions(item.children, option.children)
          optionsData.push(option)
        } else {
          const option = {
            authorityId: item.authorityId,
            authorityName: authorityDisplayName(item, t)
          }
          optionsData.push(option)
        }
      })
  }

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const orderKey = ref('id')
  const desc = ref(true)

  const sortChange = ({ prop, order }) => {
    if (prop) {
      orderKey.value = prop === 'ID' ? 'id' : toSQLLine(prop)
      desc.value = order === 'descending'
    }
    getTableData()
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
    const table = await getUserList({
      page: page.value,
      pageSize: pageSize.value,
      orderKey: orderKey.value,
      desc: desc.value,
      ...searchInfo.value
    })
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
    }
  }

  watch(
    () => tableData.value,
    () => {
      setAuthorityIds()
    }
  )

  const authOptions = ref([])
  const setOptions = (authData) => {
    authOptions.value = []
    setAuthorityOptions(authData, authOptions.value)
  }

  const initPage = async () => {
    getTableData()
    const res = await getAuthorityList()
    setOptions(res.data)
  }

  initPage()

  // 重置密码对话框相关
  const resetPwdDialog = ref(false)
  const resetPwdForm = ref(null)
  const resetPwdInfo = ref({
    ID: '',
    userName: '',
    nickName: '',
    password: ''
  })

  // 生成随机密码
  const generateRandomPassword = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*'
    let password = ''
    for (let i = 0; i < 12; i++) {
      password += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    resetPwdInfo.value.password = password
    // 复制到剪贴板
    navigator.clipboard.writeText(password).then(() => {
      ElMessage({
        type: 'success',
        message: t('admin.user.pwdCopied')
      })
    }).catch(() => {
      ElMessage({
        type: 'error',
        message: t('admin.user.copyFailed')
      })
    })
  }

  // 打开重置密码对话框
  const resetPasswordFunc = (row) => {
    resetPwdInfo.value.ID = row.ID
    resetPwdInfo.value.userName = row.userName
    resetPwdInfo.value.nickName = row.nickName
    resetPwdInfo.value.password = ''
    resetPwdDialog.value = true
  }

  // 确认重置密码
  const confirmResetPassword = async () => {
    if (!resetPwdInfo.value.password) {
      ElMessage({
        type: 'warning',
        message: t('admin.user.enterOrGenPwd')
      })
      return
    }

    const res = await resetPassword({
      ID: resetPwdInfo.value.ID,
      password: resetPwdInfo.value.password
    })

    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: res.msg || t('admin.user.resetPwdOk')
      })
      resetPwdDialog.value = false
    } else {
      ElMessage({
        type: 'error',
        message: res.msg || t('admin.user.resetPwdFail')
      })
    }
  }

  // 关闭重置密码对话框
  const closeResetPwdDialog = () => {
    resetPwdInfo.value.password = ''
    resetPwdDialog.value = false
  }
  const setAuthorityIds = () => {
    tableData.value &&
      tableData.value.forEach((user) => {
        user.authorityIds =
          user.authorities &&
          user.authorities.map((i) => {
            return i.authorityId
          })
      })
  }

  const deleteUserFunc = async (row) => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    }).then(async () => {
      const res = await deleteUser({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success(t('common.deleteOk'))
        await getTableData()
      }
    })
  }

  // 弹窗相关
  const userInfo = ref({
    userName: '',
    password: '',
    nickName: '',
    headerImg: '',
    authorityId: '',
    authorityIds: [],
    enable: 1
  })

  const rules = computed(() => ({
    userName: [
      { required: true, message: t('admin.user.ruleUsername'), trigger: 'blur' },
      { min: 5, message: t('admin.user.ruleUsernameMin'), trigger: 'blur' }
    ],
    password: [
      { required: true, message: t('admin.user.rulePassword'), trigger: 'blur' },
      { min: 6, message: t('admin.user.rulePasswordMin'), trigger: 'blur' }
    ],
    nickName: [{ required: true, message: t('admin.user.ruleNickName'), trigger: 'blur' }],
    phone: [
      {
        pattern: /^1([38][0-9]|4[014-9]|[59][0-35-9]|6[2567]|7[0-8])\d{8}$/,
        message: t('admin.user.rulePhone'),
        trigger: 'blur'
      }
    ],
    email: [
      {
        pattern: /^([0-9A-Za-z\-_.]+)@([0-9a-z]+\.[a-z]{2,3}(\.[a-z]{2})?)$/g,
        message: t('admin.user.ruleEmail'),
        trigger: 'blur'
      }
    ],
    authorityId: [
      { required: true, message: t('admin.user.ruleAuthority'), trigger: 'blur' }
    ]
  }))
  const userForm = ref(null)
  const enterAddUserDialog = async () => {
    userInfo.value.authorityId = userInfo.value.authorityIds[0]
    userForm.value.validate(async (valid) => {
      if (valid) {
        const req = {
          ...userInfo.value
        }
        if (dialogFlag.value === 'add') {
          const res = await register(req)
          if (res.code === 0) {
            ElMessage({ type: 'success', message: t('admin.user.createOk') })
            await getTableData()
            closeAddUserDialog()
          }
        }
        if (dialogFlag.value === 'edit') {
          const res = await setUserInfo(req)
          if (res.code === 0) {
            ElMessage({ type: 'success', message: t('admin.user.editOk') })
            await getTableData()
            closeAddUserDialog()
          }
        }
      }
    })
  }

  const addUserDialog = ref(false)
  const closeAddUserDialog = () => {
    userForm.value.resetFields()
    userInfo.value.headerImg = ''
    userInfo.value.authorityIds = []
    addUserDialog.value = false
  }

  const dialogFlag = ref('add')

  const addUser = () => {
    dialogFlag.value = 'add'
    addUserDialog.value = true
  }

  const tempAuth = {}
  const changeAuthority = async (row, flag, removeAuth) => {
    if (flag) {
      if (!removeAuth) {
        tempAuth[row.ID] = [...row.authorityIds]
      }
      return
    }
    await nextTick()
    const res = await setUserAuthorities({
      ID: row.ID,
      authorityIds: row.authorityIds
    })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: t('admin.user.roleSetOk') })
    } else {
      if (!removeAuth) {
        row.authorityIds = [...tempAuth[row.ID]]
        delete tempAuth[row.ID]
      } else {
        row.authorityIds = [removeAuth, ...row.authorityIds]
      }
    }
  }

  const openEdit = (row) => {
    dialogFlag.value = 'edit'
    userInfo.value = JSON.parse(JSON.stringify(row))
    addUserDialog.value = true
  }

  const switchEnable = async (row) => {
    userInfo.value = JSON.parse(JSON.stringify(row))
    await nextTick()
    const req = {
      ...userInfo.value
    }
    const res = await setUserInfo(req)
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: req.enable === 2 ? t('admin.user.disabledOk') : t('admin.user.enabledOk')
      })
      await getTableData()
      userInfo.value.headerImg = ''
      userInfo.value.authorityIds = []
    }
  }
</script>

<style lang="scss">
  .header-img-box {
    @apply w-52 h-52 border border-solid border-gray-300 rounded-xl flex justify-center items-center cursor-pointer;
  }
</style>
