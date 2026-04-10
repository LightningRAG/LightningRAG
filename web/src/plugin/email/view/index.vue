<template>
  <div>
    <warning-bar :title="t('example.emailPlugin.warningBar')" />
    <div class="lrag-form-box">
      <el-form
        ref="emailForm"
        label-position="right"
        label-width="80px"
        :model="form"
      >
        <el-form-item :label="t('example.emailPlugin.labelTo')">
          <el-input v-model="form.to" />
        </el-form-item>
        <el-form-item :label="t('example.emailPlugin.labelSubject')">
          <el-input v-model="form.subject" />
        </el-form-item>
        <el-form-item :label="t('example.emailPlugin.labelBody')">
          <el-input v-model="form.body" type="textarea" />
        </el-form-item>
        <el-form-item>
          <el-button @click="sendTestEmail">{{ t('example.emailPlugin.sendTest') }}</el-button>
          <el-button @click="sendEmail">{{ t('example.emailPlugin.send') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { emailTest } from '@/plugin/email/api/email.js'
  import { ElMessage } from 'element-plus'
  import { reactive, ref } from 'vue'
  import { useI18n } from 'vue-i18n'

  defineOptions({
    name: 'Email'
  })

  const { t } = useI18n()
  const emailForm = ref(null)
  const form = reactive({
    to: '',
    subject: '',
    body: ''
  })
  const sendTestEmail = async () => {
    const res = await emailTest()
    if (res.code === 0) {
      ElMessage.success(t('example.emailPlugin.sendOk'))
    }
  }

  const sendEmail = async () => {
    const res = await emailTest()
    if (res.code === 0) {
      ElMessage.success(t('example.emailPlugin.sendOkCheck'))
    }
  }
</script>
