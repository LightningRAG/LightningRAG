<template>
  <div>
    <warning-bar :title="$t('rag.channelConnector.warningBar')" />
    <div class="lrag-table-box">
      <div class="lrag-btn-list">
        <el-button type="primary" icon="plus" @click="openCreate">{{ $t('rag.channelConnector.btnNew') }}</el-button>
      </div>
      <el-table :data="tableData" style="width: 100%" tooltip-effect="dark" row-key="id">
        <el-table-column align="left" :label="$t('rag.channelConnector.colName')" prop="name" width="160" />
        <el-table-column align="left" :label="$t('rag.channelConnector.colChannel')" prop="channel" width="120" />
        <el-table-column align="left" :label="$t('rag.channelConnector.colAgent')" prop="agentId" width="88" />
        <el-table-column align="left" :label="$t('rag.channelConnector.colEnabled')" width="88">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.enabled ? 'success' : 'info'">
              {{ scope.row.enabled ? $t('rag.kb.on') : $t('rag.kb.off') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.colWebhook')" min-width="220" show-overflow-tooltip>
          <template #default="scope">
            <span class="text-xs font-mono">{{ publicWebhookUrl(scope.row.webhookUrl) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.colSecret')" width="100">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.webhookSecretSet ? 'success' : 'warning'">
              {{ scope.row.webhookSecretSet ? $t('rag.channelConnector.secretSet') : $t('rag.channelConnector.secretUnset') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.colCreatedAt')" width="168">
          <template #default="scope">
            <span>{{ formatDate(scope.row.createdAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.colActions')" width="200" fixed="right">
          <template #default="scope">
            <el-button type="primary" link icon="document-copy" @click="copyUrl(scope.row)">{{ $t('rag.channelConnector.copyUrl') }}</el-button>
            <el-button type="primary" link icon="edit" @click="openEdit(scope.row)">{{ $t('rag.channelConnector.edit') }}</el-button>
            <el-button type="danger" link icon="delete" @click="removeRow(scope.row)">{{ $t('rag.channelConnector.delete') }}</el-button>
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

    <el-divider content-position="left">{{ $t('rag.channelConnector.outboundSection') }}</el-divider>
    <div class="lrag-table-box mb-8">
      <div class="lrag-btn-list flex flex-wrap items-center gap-2">
        <el-select
          v-model="outboundFilterConnectorId"
          clearable
          :placeholder="$t('rag.channelConnector.outboundFilterConnector')"
          style="width: 220px"
          @change="() => { obPage = 1; loadOutbound() }"
        >
          <el-option
            v-for="c in connectorFilterOptions"
            :key="c.id"
            :label="c.name + ' (' + c.channel + ')'"
            :value="c.id"
          />
        </el-select>
        <el-button icon="refresh" @click="loadOutbound">{{ $t('rag.channelConnector.outboundRefresh') }}</el-button>
        <el-button :loading="outboundRunLoading" @click="runOutboundOnce">{{ $t('rag.channelConnector.outboundRunOnce') }}</el-button>
      </div>
      <el-table :data="outboundRows" style="width: 100%" tooltip-effect="dark" row-key="id" v-loading="outboundLoading">
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColConnector')" min-width="140" show-overflow-tooltip>
          <template #default="scope">
            {{ scope.row.connectorName || ('#' + scope.row.connectorId) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.colChannel')" prop="channel" width="100" />
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColKind')" prop="kind" width="140" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColPreview')" prop="textPreview" min-width="160" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColAttempts')" prop="attempts" width="72" />
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColNextRetry')" width="168">
          <template #default="scope">
            <span>{{ formatDate(scope.row.nextRetryAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColLeaseUntil')" width="168">
          <template #default="scope">
            <span>{{ scope.row.leaseUntil ? formatDate(scope.row.leaseUntil) : '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('rag.channelConnector.outboundColLastErr')" prop="lastErr" min-width="140" show-overflow-tooltip />
        <el-table-column align="left" :label="$t('rag.channelConnector.colActions')" width="88" fixed="right">
          <template #default="scope">
            <el-button type="danger" link icon="delete" @click="removeOutbound(scope.row)">{{ $t('rag.channelConnector.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="!outboundLoading && !outboundRows.length" class="text-sm text-gray-500 py-2">{{ $t('rag.channelConnector.outboundEmpty') }}</div>
      <div class="lrag-pagination">
        <el-pagination
          :current-page="obPage"
          :page-size="obPageSize"
          :page-sizes="[10, 20, 30]"
          :total="obTotal"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="(v) => { obPage = v; loadOutbound() }"
          @size-change="(v) => { obPageSize = v; loadOutbound() }"
        />
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editId ? $t('rag.channelConnector.dialogEdit') : $t('rag.channelConnector.dialogNew')"
      width="560px"
      destroy-on-close
    >
      <el-form :model="form" label-width="120px">
        <el-form-item :label="$t('rag.channelConnector.colName')" required>
          <el-input v-model="form.name" :placeholder="$t('rag.channelConnector.phName')" />
        </el-form-item>
        <el-form-item :label="$t('rag.channelConnector.colChannel')" required>
          <el-select v-model="form.channel" :disabled="!!editId" style="width: 100%" filterable>
            <el-option v-for="c in channelOptions" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.channelConnector.colAgent')" required>
          <el-select v-model="form.agentId" style="width: 100%" filterable :placeholder="$t('rag.channelConnector.phAgent')">
            <el-option v-for="a in agents" :key="a.ID" :label="agentLabel(a)" :value="a.ID" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.channelConnector.colEnabled')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item v-if="!editId" :label="$t('rag.channelConnector.webhookSecret')">
          <el-input v-model="form.webhookSecret" type="password" show-password clearable :placeholder="$t('rag.channelConnector.phSecret')" />
        </el-form-item>
        <el-form-item v-else :label="$t('rag.channelConnector.rotateSecret')">
          <el-input v-model="form.webhookSecret" type="password" show-password clearable :placeholder="$t('rag.channelConnector.phRotateSecret')" />
        </el-form-item>
        <el-form-item :label="$t('rag.channelConnector.extraJson')">
          <el-input v-model="form.extra" type="textarea" :rows="5" :placeholder="$t('rag.channelConnector.phExtra')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">{{ $t('settings.general.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    channelConnectorList,
    channelConnectorChannelTypes,
    channelConnectorCreate,
    channelConnectorUpdate,
    channelConnectorDelete,
    channelOutboundList,
    channelOutboundDelete,
    channelOutboundRunOnce
  } from '@/api/rag'
  import { agentList } from '@/api/rag'
  import { ref, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate } from '@/utils/format'
  import WarningBar from '@/components/warningBar/warningBar.vue'

  defineOptions({ name: 'RagChannelConnectors' })

  const { t } = useI18n()
  const tableData = ref([])
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const dialogVisible = ref(false)
  const editId = ref(null)
  const submitting = ref(false)
  const agents = ref([])
  const outboundRows = ref([])
  const outboundLoading = ref(false)
  const obPage = ref(1)
  const obPageSize = ref(10)
  const obTotal = ref(0)
  const outboundFilterConnectorId = ref(null)
  const connectorFilterOptions = ref([])
  const outboundRunLoading = ref(false)

  const CHANNEL_OPTIONS_FALLBACK = ['dingtalk', 'discord', 'feishu', 'line', 'mock', 'slack', 'teams', 'telegram', 'wechat_mp', 'wecom', 'whatsapp']
  const channelOptions = ref([...CHANNEL_OPTIONS_FALLBACK])

  const loadChannelTypes = async () => {
    try {
      const res = await channelConnectorChannelTypes()
      if (res.code === 0 && Array.isArray(res.data) && res.data.length > 0) {
        channelOptions.value = res.data
      }
    } catch {
      /* 保持 FALLBACK */
    }
  }

  const form = ref({
    name: '',
    channel: 'mock',
    agentId: null,
    enabled: true,
    webhookSecret: '',
    extra: ''
  })

  const publicWebhookUrl = (path) => {
    if (!path) return ''
    const base = import.meta.env.VITE_BASE_PATH
    if (base) {
      const b = String(base).replace(/\/$/, '')
      return b + (path.startsWith('/') ? path : `/${path}`)
    }
    if (typeof window !== 'undefined') {
      return `${window.location.origin}${path.startsWith('/') ? path : `/${path}`}`
    }
    return path
  }

  const agentLabel = (a) => (a.name || t('rag.channelConnector.unnamedAgent')) + ` (id=${a.ID})`

  const loadAgents = async () => {
    const res = await agentList({ page: 1, pageSize: 500 })
    if (res.code === 0 && res.data?.list) {
      agents.value = res.data.list
    }
  }

  const loadList = async () => {
    const res = await channelConnectorList({ page: page.value, pageSize: pageSize.value })
    if (res.code === 0 && res.data) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const loadConnectorFilterOptions = async () => {
    const res = await channelConnectorList({ page: 1, pageSize: 500 })
    if (res.code === 0 && res.data?.list) {
      connectorFilterOptions.value = res.data.list
    }
  }

  const runOutboundOnce = async () => {
    outboundRunLoading.value = true
    try {
      const res = await channelOutboundRunOnce()
      if (res.code === 0) {
        const n = res.data?.processed ?? 0
        ElMessage.success(t('rag.channelConnector.outboundRunOnceOk', { n }))
        loadOutbound()
      }
    } finally {
      outboundRunLoading.value = false
    }
  }

  const loadOutbound = async () => {
    outboundLoading.value = true
    try {
      const payload = { page: obPage.value, pageSize: obPageSize.value }
      if (outboundFilterConnectorId.value) {
        payload.connectorId = outboundFilterConnectorId.value
      }
      const res = await channelOutboundList(payload)
      if (res.code === 0 && res.data) {
        outboundRows.value = res.data.list || []
        obTotal.value = res.data.total || 0
      }
    } finally {
      outboundLoading.value = false
    }
  }

  const removeOutbound = (row) => {
    ElMessageBox.confirm(t('rag.channelConnector.outboundDeleteConfirm', { id: row.id }), t('rag.channelConnector.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    })
      .then(async () => {
        const res = await channelOutboundDelete({ id: row.id })
        if (res.code === 0) {
          ElMessage.success(t('rag.channelConnector.deleteOk'))
          loadOutbound()
        }
      })
      .catch(() => {})
  }

  const openCreate = () => {
    editId.value = null
    const firstCh = channelOptions.value[0] ?? 'mock'
    form.value = {
      name: t('rag.channelConnector.defaultName'),
      channel: firstCh,
      agentId: agents.value[0]?.ID ?? null,
      enabled: true,
      webhookSecret: '',
      extra: ''
    }
    dialogVisible.value = true
  }

  const openEdit = (row) => {
    editId.value = row.id
    form.value = {
      name: row.name || '',
      channel: row.channel || 'mock',
      agentId: row.agentId,
      enabled: !!row.enabled,
      webhookSecret: '',
      extra: row.extra || ''
    }
    dialogVisible.value = true
  }

  const isValidConnectorExtraJson = (raw) => {
    const s = String(raw ?? '').trim()
    if (!s) return true
    let v
    try {
      v = JSON.parse(s)
    } catch {
      return false
    }
    return v !== null && typeof v === 'object' && !Array.isArray(v)
  }

  const submit = async () => {
    if (!form.value.name?.trim()) {
      ElMessage.warning(t('rag.channelConnector.needName'))
      return
    }
    if (!form.value.agentId) {
      ElMessage.warning(t('rag.channelConnector.needAgent'))
      return
    }
    if (!isValidConnectorExtraJson(form.value.extra)) {
      ElMessage.warning(t('rag.channelConnector.extraInvalidJson'))
      return
    }
    submitting.value = true
    try {
      if (editId.value) {
        const payload = {
          id: editId.value,
          name: form.value.name.trim(),
          agentId: form.value.agentId,
          enabled: form.value.enabled,
          extra: form.value.extra || ''
        }
        if (form.value.webhookSecret?.trim()) {
          payload.webhookSecret = form.value.webhookSecret.trim()
        }
        const res = await channelConnectorUpdate(payload)
        if (res.code === 0) {
          ElMessage.success(t('rag.channelConnector.updateOk'))
          dialogVisible.value = false
          loadList()
          loadConnectorFilterOptions()
          loadOutbound()
        }
      } else {
        const payload = {
          name: form.value.name.trim(),
          channel: form.value.channel,
          agentId: form.value.agentId,
          enabled: form.value.enabled,
          extra: form.value.extra || ''
        }
        if (form.value.webhookSecret?.trim()) {
          payload.webhookSecret = form.value.webhookSecret.trim()
        }
        const res = await channelConnectorCreate(payload)
        if (res.code === 0 && res.data) {
          ElMessage.success(t('rag.channelConnector.createOk'))
          dialogVisible.value = false
          loadList()
          loadConnectorFilterOptions()
          loadOutbound()
          if (res.data.webhookSecret) {
            ElMessageBox.alert(
              `${t('rag.channelConnector.secretOnceHint')}\n\n${res.data.webhookSecret}`,
              t('rag.channelConnector.secretOnceTitle'),
              { confirmButtonText: t('settings.general.confirm') }
            )
          }
        }
      }
    } finally {
      submitting.value = false
    }
  }

  const copyUrl = async (row) => {
    const text = publicWebhookUrl(row.webhookUrl)
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success(t('rag.channelConnector.copyOk'))
    } catch {
      ElMessage.warning(t('rag.channelConnector.copyFail'))
    }
  }

  const removeRow = (row) => {
    ElMessageBox.confirm(t('rag.channelConnector.deleteConfirm', { name: row.name }), t('rag.channelConnector.deleteTitle'), {
      confirmButtonText: t('settings.general.confirm'),
      cancelButtonText: t('settings.general.cancel'),
      type: 'warning'
    })
      .then(async () => {
        const res = await channelConnectorDelete({ id: row.id })
        if (res.code === 0) {
          ElMessage.success(t('rag.channelConnector.deleteOk'))
          loadList()
          loadConnectorFilterOptions()
          loadOutbound()
        }
      })
      .catch(() => {})
  }

  onMounted(async () => {
    await loadChannelTypes()
    await loadAgents()
    await loadList()
    await loadConnectorFilterOptions()
    await loadOutbound()
  })
</script>
