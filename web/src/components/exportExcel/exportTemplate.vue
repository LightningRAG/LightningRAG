<template>
  <el-button type="primary" icon="download" @click="exportTemplateFunc">{{
    $t('common.exportExcel.btnDownloadTemplate')
  }}</el-button>
</template>

<script setup>
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import {exportTemplate} from "@/api/exportTemplate";

  const { t } = useI18n()

  const props = defineProps({
    templateId: {
      type: String,
      required: true
    }
  })


  const exportTemplateFunc = async () => {
    if (props.templateId === '') {
      ElMessage.error(t('common.exportExcel.noTemplateId'))
      return
    }
    let baseUrl = import.meta.env.VITE_BASE_API
    if (baseUrl === "/"){
      baseUrl = ""
    }

    const res = await exportTemplate({
      templateID: props.templateId
    })

    if(res.code === 0){
      ElMessage.success(t('common.exportExcel.taskCreated'))
      const url = `${baseUrl}${res.data}`
      window.open(url, '_blank')
    }

  }
</script>
