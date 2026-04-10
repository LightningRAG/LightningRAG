<template>
  <div>
    <div class="lrag-search-box">
      <el-form :inline="true" :model="searchInfo">
          <el-form-item :label="$t('common.colUserId')">
              <el-input v-model.number="searchInfo.userId" :placeholder="$t('common.phSearchUserId')" />
          </el-form-item>
        <el-form-item :label="$t('common.colStatus')">
             <el-select v-model="searchInfo.status" :placeholder="$t('common.phSelect')" clearable>
                 <el-option :label="$t('common.statusActive')" :value="true" />
                 <el-option :label="$t('common.statusInactive')" :value="false" />
             </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">{{ $t('common.query') }}</el-button>
          <el-button icon="refresh" @click="onReset">{{ $t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openDrawer">{{ $t('common.issue') }}</el-button>
      </div>
      <el-table
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        row-key="ID"
      >
        <el-table-column align="left" :label="$t('common.colId')" prop="ID" width="80" />
        <el-table-column align="left" :label="$t('common.colUser')" min-width="150">
             <template #default="scope">
                 {{ scope.row.user.nickName }} ({{ scope.row.user.userName }})
             </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colAuthorityId')" prop="authorityId" width="100" />
        <el-table-column align="left" :label="$t('common.colStatus')" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status ? 'success' : 'danger'">
              {{ scope.row.status ? $t('common.statusActive') : $t('common.statusRevoked') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('common.colExpiresAt')" width="180">
          <template #default="scope">{{ formatDate(scope.row.expiresAt) }}</template>
        </el-table-column>
         <el-table-column align="left" :label="$t('common.colRemark')" prop="remark" min-width="150" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('common.colActions')" width="220">
          <template #default="scope">
            <el-button type="primary" link icon="link" @click="openCurl(scope.row)">{{ $t('toolsApiToken.curlExample') }}</el-button>
            <el-popover v-if="scope.row.status" v-model:visible="scope.row.visible" placement="top" width="160">
              <p>{{ $t('toolsApiToken.confirmRevoke') }}</p>
              <div style="text-align: right; margin: 0">
                <el-button size="small" type="primary" link @click="scope.row.visible = false">{{ $t('settings.general.cancel') }}</el-button>
                <el-button size="small" type="primary" @click="invalidateToken(scope.row)">{{ $t('settings.general.confirm') }}</el-button>
              </div>
              <template #reference>
                <el-button icon="delete" type="danger" link @click="scope.row.visible = true">{{ $t('toolsApiToken.revoke') }}</el-button>
              </template>
            </el-popover>
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

    <el-drawer v-model="drawerVisible" size="400px" :title="$t('toolsApiToken.drawerTitle')">
         <el-form ref="formRef" :model="form" label-width="80px">
             <el-form-item :label="$t('common.colUser')" required>
                 <el-select 
                    v-model="form.userId" 
                    :placeholder="$t('toolsApiToken.phPickUser')" 
                    filterable 
                    style="width:100%"
                    @change="handleUserChange"
                 >
                     <el-option
                        v-for="item in userOptions"
                        :key="item.ID"
                        :label="`${item.nickName} (${item.userName})`"
                        :value="item.ID"
                     />
                 </el-select>
             </el-form-item>
             <el-form-item :label="$t('common.colRole')" required>
                 <el-select v-model="form.authorityId" :placeholder="$t('toolsApiToken.phPickRole')" style="width:100%" :disabled="!form.userId">
                     <el-option
                        v-for="item in authorityOptions"
                        :key="item.authorityId"
                        :label="authorityOptionLabel(item)"
                        :value="item.authorityId"
                     />
                 </el-select>
             </el-form-item>
            <el-form-item :label="$t('toolsApiToken.validity')">
                <el-select v-model="form.days" :placeholder="$t('common.phSelect')" style="width:100%">
                    <el-option :label="$t('toolsApiToken.days1')" :value="1" />
                    <el-option :label="$t('toolsApiToken.days7')" :value="7" />
                    <el-option :label="$t('toolsApiToken.days30')" :value="30" />
                    <el-option :label="$t('toolsApiToken.days90')" :value="90" />
                    <el-option :label="$t('common.daysPermanent')" :value="-1" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('common.colRemark')">
                <el-input v-model="form.remark" type="textarea" />
            </el-form-item>
         </el-form>
         <template #footer>
             <div style="flex: auto">
                 <el-button @click="drawerVisible = false">{{ $t('settings.general.cancel') }}</el-button>
                 <el-button type="primary" @click="submitIssuer">{{ $t('toolsApiToken.issueJwt') }}</el-button>
             </div>
         </template>
    </el-drawer>

    <el-dialog v-model="tokenDialogVisible" :title="$t('toolsApiToken.issueSuccess')" width="500px">
        <div style="text-align: center; margin-bottom: 20px;">
            <el-alert :title="$t('toolsApiToken.copyTokenWarn')" type="warning" :closable="false" show-icon />
        </div>
        <el-input type="textarea" :rows="6" v-model="tokenResult" readonly />
        <template #footer>
            <el-button @click="copyText(tokenResult)">{{ $t('common.btnCopy') }}</el-button>
            <el-button type="primary" @click="tokenDialogVisible = false">{{ $t('common.btnClose') }}</el-button>
        </template>
    </el-dialog>

    <el-drawer v-model="curlDrawerVisible" size="500px" :title="$t('toolsApiToken.curlDrawerTitle')">
        <div style="padding: 10px;">
            <p style="margin-bottom: 10px;">{{ $t('toolsApiToken.curlHeaderMode') }}</p>
            <el-input type="textarea" :rows="4" v-model="curlHeader" readonly />
            <el-button style="margin-top: 5px;" size="small" @click="copyText(curlHeader)">{{ $t('common.btnCopy') }}</el-button>
            
            <el-divider />
            
            <p style="margin-bottom: 10px;">{{ $t('toolsApiToken.curlCookieMode') }}</p>
            <el-input type="textarea" :rows="4" v-model="curlCookie" readonly />
            <el-button style="margin-top: 5px;" size="small" @click="copyText(curlCookie)">{{ $t('common.btnCopy') }}</el-button>
        </div>
    </el-drawer>
  </div>
</template>

<script setup>
import {
  getApiTokenList,
  createApiToken,
  deleteApiToken
} from '@/api/sysApiToken'
import { getUserList } from '@/api/user'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { formatDate } from '@/utils/format'
import { authorityDisplayName } from '@/utils/authorityI18n'

const { t } = useI18n()
const authorityOptionLabel = (item) =>
  `${authorityDisplayName(item, t)} (${item.authorityId})`
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})

const drawerVisible = ref(false)
const tokenDialogVisible = ref(false)
const tokenResult = ref('')
const curlDrawerVisible = ref(false)
const curlHeader = ref('')
const curlCookie = ref('')

const form = ref({
    userId: '',
    authorityId: '',
    days: 30,
    remark: ''
})

const userOptions = ref([])
const authorityOptions = ref([])

const getTableData = async () => {
  const table = await getApiTokenList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

const openDrawer = async () => {
    form.value = { userId: '', authorityId: '', days: 30, remark: '' }
    authorityOptions.value = []
    drawerVisible.value = true
    if (userOptions.value.length === 0) {
        const res = await getUserList({ page: 1, pageSize: 999 })
        if (res.code === 0) {
            userOptions.value = res.data.list
        }
    }
}

const handleUserChange = (val) => {
    form.value.authorityId = ''
    const user = userOptions.value.find(u => u.ID === val)
    if (user) {
        authorityOptions.value = user.authorities || []
        // 默认选中第一个
        if (authorityOptions.value.length > 0) {
            form.value.authorityId = authorityOptions.value[0].authorityId
        }
    } else {
        authorityOptions.value = []
    }
}

const submitIssuer = async () => {
    if (!form.value.userId || !form.value.authorityId) {
        ElMessage.warning(t('toolsApiToken.pickUserAndRole'))
        return
    }
    const res = await createApiToken(form.value)
    if (res.code === 0) {
        tokenResult.value = res.data.token
        drawerVisible.value = false
        tokenDialogVisible.value = true
        getTableData()
    }
}

const invalidateToken = async (row) => {
    row.visible = false
    const res = await deleteApiToken({ ID: row.ID })
    if (res.code === 0) {
        ElMessage.success(t('toolsApiToken.revokeOk'))
        getTableData()
    }
}

const openCurl = (row) => {
    // 假设 API Host 为当前 origin
    const origin = window.location.origin
    // 构造示例 URL
    const url = `${origin}/api/menu/getMenu`
    
    curlHeader.value = `curl -X POST "${url}" \\
  -H "x-token: ${row.token}" \\
  -H "Content-Type: application/json"`

    curlCookie.value = `curl -X POST "${url}" \\
  -b "x-token=${row.token}" \\
  -H "Content-Type: application/json"`

    curlDrawerVisible.value = true
}

const copyText = (text) => {
    if (!text) return
    const input = document.createElement('textarea')
    input.value = text
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    ElMessage.success(t('common.copyOk'))
}

const onSubmit = () => {
  page.value = 1
  pageSize.value = 10
  getTableData()
}

const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

getTableData()
</script>

<style scoped>
</style>
