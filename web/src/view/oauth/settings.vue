<template>
  <div>
    <warning-bar :title="$t('admin.oauth.warningBar')" />
    <div class="text-sm text-gray-600 dark:text-gray-400 mb-3 max-w-3xl">
      <div class="flex flex-wrap items-center gap-2">
        <span>{{ $t('admin.oauth.callbackUrlHint') }}</span>
        <code class="text-xs bg-gray-100 dark:bg-slate-800 px-2 py-1 rounded break-all max-w-full">{{
          callbackUrlTemplate
        }}</code>
        <el-button size="small" @click="copyCallbackTpl">{{ $t('admin.oauth.copyCallbackTpl') }}</el-button>
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-500 mt-2 break-all">
        {{ $t('admin.oauth.callbackPathBackend') }}
        <code class="bg-gray-100 dark:bg-slate-800 px-1.5 py-0.5 rounded">{{ callbackPathPatternDisplay }}</code>
      </div>
    </div>
    <div class="lrag-table-box mb-4">
      <div class="text-sm font-medium mb-3">{{ $t('admin.oauth.globalSection') }}</div>
      <el-form :model="globalForm" label-width="168px" class="max-w-3xl">
        <el-form-item :label="$t('admin.oauth.globalFrontend')">
          <el-input
            v-model="globalForm.frontendRedirect"
            :placeholder="$t('admin.oauth.globalFrontendPh')"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.globalSecret')">
          <el-input
            v-model="globalForm.secretKey"
            type="password"
            show-password
            clearable
            :placeholder="$t('admin.oauth.globalSecretPh')"
          />
          <div class="text-xs text-gray-500 mt-1">{{ $t('admin.oauth.globalSecretHint') }}</div>
        </el-form-item>
        <el-form-item>
          <el-tag v-if="globalSecretConfigured" size="small" type="success" class="mr-2">{{
            $t('admin.oauth.secretSet')
          }}</el-tag>
          <el-button type="primary" :loading="globalSaving" @click="saveGlobal">{{
            $t('admin.oauth.saveGlobal')
          }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openCreate">{{ $t('admin.oauth.btnNew') }}</el-button>
      </div>
      <el-table :data="tableData" style="width: 100%" tooltip-effect="dark" row-key="ID">
        <el-table-column align="left" :label="$t('common.colDate')" width="168">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('admin.oauth.colKind')" prop="kind" width="120" />
        <el-table-column align="left" :label="$t('admin.oauth.colDisplayName')" prop="displayName" min-width="60" show-overflow-tooltip />
        <el-table-column align="center" :label="$t('admin.oauth.colButtonIcon')" width="144">
          <template #default="scope">
            <img
              v-if="tableRowOAuthIcon(scope.row)"
              :src="tableRowOAuthIcon(scope.row)"
              class="h-7 w-7 object-contain rounded mx-auto block"
              alt=""
            />
            <span v-else class="text-gray-400 text-xs">—</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('admin.oauth.colEnabled')" width="88">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.enabled ? 'success' : 'info'">
              {{ scope.row.enabled ? $t('admin.oauth.on') : $t('admin.oauth.off') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('admin.oauth.colClientId')" prop="clientId" min-width="160" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('admin.oauth.colSecret')" width="100">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.clientSecretSet ? 'success' : 'warning'">
              {{ scope.row.clientSecretSet ? $t('admin.oauth.secretSet') : $t('admin.oauth.secretUnset') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('admin.oauth.colScopes')" prop="scopes" min-width="140" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('admin.oauth.colDefaultRole')" prop="defaultAuthorityId" width="120" />
        <el-table-column align="left" :label="$t('admin.oauth.colActions')" min-width="220" fixed="right">
          <template #default="scope">
            <el-button type="primary" link icon="edit" @click="openEdit(scope.row)">{{ $t('common.change') }}</el-button>
            <el-button type="danger" link icon="delete" @click="removeRow(scope.row)">{{ $t('common.delete') }}</el-button>
            <el-button type="primary" link @click="copyRowCallbackUrl(scope.row)">{{ $t('admin.oauth.copyRowCallback') }}</el-button>
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
          @current-change="(v) => { page = v; loadList() }"
          @size-change="(v) => { pageSize = v; loadList() }"
        />
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editId ? $t('admin.oauth.dialogEdit') : $t('admin.oauth.dialogNew')"
      width="620px"
      destroy-on-close
    >
      <el-form :model="form" label-width="140px">
        <el-form-item :label="$t('admin.oauth.colKind')" required>
          <el-select v-model="form.kind" :disabled="!!editId" style="width: 100%" filterable :placeholder="$t('admin.oauth.phKind')">
            <el-option v-for="k in kindOptions" :key="k.kind" :label="`${k.displayName} (${k.kind})`" :value="k.kind" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colEnabled')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colDisplayName')">
          <el-input v-model="form.displayName" :placeholder="$t('admin.oauth.phDisplayName')" clearable />
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colButtonIcon')">
          <el-input
            v-model="form.buttonIcon"
            type="textarea"
            :rows="3"
            :placeholder="$t('admin.oauth.phButtonIcon')"
            clearable
          />
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ $t('admin.oauth.hintButtonIcon') }}</div>
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colClientId')" required>
          <el-input v-model="form.clientId" :placeholder="$t('admin.oauth.phClientId')" clearable />
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colSecret')" :required="!editId">
          <el-input v-model="form.clientSecret" type="password" show-password clearable :placeholder="$t('admin.oauth.phClientSecret')" />
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colScopes')">
          <el-input v-model="form.scopes" :placeholder="$t('admin.oauth.phScopes')" clearable />
        </el-form-item>
        <el-form-item :label="$t('admin.oauth.colDefaultRole')">
          <el-input-number v-model="form.defaultAuthorityId" :min="1" :step="1" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="$t('common.colRemark')">
          <el-input v-model="form.extra" type="textarea" :rows="4" :placeholder="$t('admin.oauth.phExtraJson')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">{{ $t('common.ok') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    oauthGlobalGet,
    oauthGlobalUpdate,
    oauthRegisteredKinds,
    oauthProviderList,
    oauthProviderFind,
    oauthProviderCreate,
    oauthProviderUpdate,
    oauthProviderDelete
  } from '@/api/sysOAuthProvider'
  import { ref, computed, watch, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate } from '@/utils/format'
  import WarningBar from '@/components/warningBar/warningBar.vue'

  defineOptions({ name: 'QuickOAuthSettings' })

  const { t } = useI18n()
  const tableData = ref([])
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const dialogVisible = ref(false)
  const editId = ref(null)
  const submitting = ref(false)
  const kindOptions = ref([])
  const globalForm = ref({
    frontendRedirect: '',
    secretKey: ''
  })
  const globalSecretConfigured = ref(false)
  const globalSaving = ref(false)
  const callbackPathFromServer = ref('')

  /** OAuth 完成回跳：未配置库表时，用当前浏览器访问的前端 origin + 控制台默认路由 */
  const defaultOAuthFrontendRedirect = () => {
    if (typeof window === 'undefined') return ''
    return `${window.location.origin}/#/layout/dashboard`
  }

  /** 与登录页 oauth 跳转、axios baseURL 一致（dev 下含 /api） */
  const oauthApiBase = () => {
    const raw = import.meta.env.VITE_BASE_API || ''
    if (raw.startsWith('http://') || raw.startsWith('https://')) {
      return raw.replace(/\/$/, '')
    }
    const path = raw.startsWith('/') ? raw : `/${raw}`
    if (typeof window !== 'undefined') {
      return `${window.location.origin}${path}`.replace(/\/$/, '')
    }
    const origin = String(import.meta.env.VITE_BASE_PATH || '').replace(/\/$/, '')
    return `${origin}${path}`.replace(/\/$/, '')
  }

  const callbackPathPatternDisplay = computed(
    () => callbackPathFromServer.value || '/base/oauth/callback/{kind}'
  )

  const callbackUrlTemplate = computed(() => {
    const base = oauthApiBase().replace(/\/$/, '')
    const path = callbackPathPatternDisplay.value.trim()
    const p = path.startsWith('/') ? path : `/${path}`
    return `${base}${p}`
  })

  const copyCallbackTpl = async () => {
    try {
      await navigator.clipboard.writeText(callbackUrlTemplate.value)
      ElMessage.success(t('admin.oauth.callbackCopied'))
    } catch {
      ElMessage.error(t('admin.oauth.callbackCopyFailed'))
    }
  }

  /** 当前浏览器 origin（协议+主机+端口）+ 后端路径模板中的 {kind} 换成本行平台 */
  const buildBrowserOriginCallbackUrl = (row) => {
    const kind = String(row?.kind || '').trim()
    if (!kind) return ''
    let pat = callbackPathPatternDisplay.value.trim() || '/base/oauth/callback/{kind}'
    pat = pat.replace(/\{kind\}/gi, kind)
    const p = pat.startsWith('/') ? pat : `/${pat}`
    if (typeof window === 'undefined') return p
    return `${window.location.origin.replace(/\/$/, '')}${p}`
  }

  const copyRowCallbackUrl = async (row) => {
    const url = buildBrowserOriginCallbackUrl(row)
    if (!url) {
      ElMessage.warning(t('admin.oauth.copyRowCallbackEmptyKind'))
      return
    }
    try {
      await navigator.clipboard.writeText(url)
      ElMessage.success(t('admin.oauth.callbackCopied'))
    } catch {
      ElMessage.error(t('admin.oauth.callbackCopyFailed'))
    }
  }

  const loadGlobal = async () => {
    try {
      const res = await oauthGlobalGet()
      if (res.code === 0 && res.data) {
        const saved = String(res.data.frontendRedirect || '').trim()
        globalForm.value.frontendRedirect = saved || defaultOAuthFrontendRedirect()
        globalForm.value.secretKey = ''
        globalSecretConfigured.value = !!res.data.secretKeyConfigured
        callbackPathFromServer.value = (res.data.callbackPathPattern || '').trim()
        return
      }
    } catch {
      /* 网络错误时仍给出可编辑的浏览器默认回跳 */
    }
    if (typeof window !== 'undefined' && !String(globalForm.value.frontendRedirect || '').trim()) {
      globalForm.value.frontendRedirect = defaultOAuthFrontendRedirect()
    }
  }

  const saveGlobal = async () => {
    globalSaving.value = true
    try {
      const res = await oauthGlobalUpdate({
        frontendRedirect: globalForm.value.frontendRedirect,
        secretKey: globalForm.value.secretKey
      })
      if (res.code === 0) {
        ElMessage.success(res.msg || t('common.update_success'))
        globalForm.value.secretKey = ''
        await loadGlobal()
      }
    } finally {
      globalSaving.value = false
    }
  }

  const form = ref({
    kind: '',
    enabled: true,
    displayName: '',
    buttonIcon: '',
    clientId: '',
    clientSecret: '',
    scopes: '',
    defaultAuthorityId: 888,
    extra: ''
  })

  const defaultIconForKind = (kind) => {
    const k = String(kind || '').trim().toLowerCase()
    const ko = kindOptions.value.find((x) => String(x.kind || '').toLowerCase() === k)
    return ko?.defaultButtonIcon || ''
  }

  /** 列表展示：优先接口 buttonIconPreview（服务端已合并默认），旧后端无该字段时回退 */
  const tableRowOAuthIcon = (row) => {
    const preview = String(row?.buttonIconPreview || '').trim()
    if (preview) return preview
    const raw = String(row?.buttonIcon || '').trim()
    if (raw) return raw
    return defaultIconForKind(row?.kind)
  }

  /** 与内置默认一致时存空库，登录页仍由后端回落默认图标 */
  const normalizeButtonIconPayload = () => {
    const v = (form.value.buttonIcon || '').trim()
    const def = defaultIconForKind(form.value.kind)
    if (def && v === def) return ''
    return v
  }

  watch(
    () => form.value.kind,
    (k) => {
      if (editId.value) return
      form.value.buttonIcon = defaultIconForKind(k)
    }
  )

  const parseExtra = () => {
    const s = (form.value.extra || '').trim()
    if (!s) return undefined
    try {
      return JSON.parse(s)
    } catch {
      ElMessage.error(t('admin.oauth.extraInvalidJson'))
      return null
    }
  }

  const loadKinds = async () => {
    const res = await oauthRegisteredKinds()
    if (res.code === 0 && Array.isArray(res.data)) {
      kindOptions.value = res.data
    }
  }

  const loadList = async () => {
    const res = await oauthProviderList({ page: page.value, pageSize: pageSize.value })
    if (res.code === 0 && res.data) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const resetForm = () => {
    const k = kindOptions.value[0]?.kind || ''
    form.value = {
      kind: k,
      enabled: true,
      displayName: '',
      buttonIcon: defaultIconForKind(k),
      clientId: '',
      clientSecret: '',
      scopes: '',
      defaultAuthorityId: 888,
      extra: ''
    }
  }

  const openCreate = () => {
    editId.value = null
    resetForm()
    dialogVisible.value = true
  }

  const openEdit = async (row) => {
    editId.value = row.ID
    const res = await oauthProviderFind(row.ID)
    if (res.code !== 0 || !res.data) return
    const d = res.data
    const storedIcon = String(d.buttonIcon || '').trim()
    form.value = {
      kind: d.kind,
      enabled: !!d.enabled,
      displayName: d.displayName || '',
      buttonIcon: storedIcon || defaultIconForKind(d.kind),
      clientId: d.clientId || '',
      clientSecret: '',
      scopes: d.scopes || '',
      defaultAuthorityId: d.defaultAuthorityId || 888,
      extra: d.extra && Object.keys(d.extra).length ? JSON.stringify(d.extra, null, 2) : ''
    }
    dialogVisible.value = true
  }

  const submit = async () => {
    const extraObj = parseExtra()
    if (extraObj === null) return
    if (!form.value.kind || !form.value.clientId) {
      ElMessage.warning(t('login.fillLoginInfo'))
      return
    }
    if (!editId.value && !form.value.clientSecret) {
      ElMessage.warning(t('admin.oauth.phClientSecret'))
      return
    }
    submitting.value = true
    try {
      if (!editId.value) {
        const payload = {
          kind: form.value.kind,
          enabled: form.value.enabled,
          displayName: form.value.displayName,
          buttonIcon: normalizeButtonIconPayload(),
          clientId: form.value.clientId,
          clientSecret: form.value.clientSecret,
          scopes: form.value.scopes,
          defaultAuthorityId: form.value.defaultAuthorityId || 888
        }
        if (extraObj !== undefined) {
          payload.extra = extraObj
        }
        const res = await oauthProviderCreate(payload)
        if (res.code === 0) {
          ElMessage.success(res.msg || t('common.create_success'))
          dialogVisible.value = false
          loadList()
        }
      } else {
        const body = {
          ID: editId.value,
          enabled: form.value.enabled,
          displayName: form.value.displayName,
          buttonIcon: normalizeButtonIconPayload(),
          clientId: form.value.clientId,
          scopes: form.value.scopes,
          defaultAuthorityId: form.value.defaultAuthorityId || 888
        }
        if (form.value.clientSecret) {
          body.clientSecret = form.value.clientSecret
        }
        if (extraObj !== undefined) {
          body.extra = extraObj
        }
        const res = await oauthProviderUpdate(body)
        if (res.code === 0) {
          ElMessage.success(res.msg || t('common.update_success'))
          dialogVisible.value = false
          loadList()
        }
      }
    } finally {
      submitting.value = false
    }
  }

  const removeRow = (row) => {
    ElMessageBox.confirm(t('common.confirmDeleteGeneric'), t('common.tipTitle'), {
      type: 'warning'
    })
      .then(async () => {
        const res = await oauthProviderDelete(row.ID)
        if (res.code === 0) {
          ElMessage.success(res.msg || t('common.delete_success'))
          loadList()
        }
      })
      .catch(() => {})
  }

  onMounted(() => {
    loadGlobal()
    loadKinds()
    loadList()
  })
</script>
