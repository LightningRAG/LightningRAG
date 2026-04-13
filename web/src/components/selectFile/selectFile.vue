<template>
  <div>
    <el-upload
      v-model:file-list="fileList"
      multiple
      :action="`${getBaseUrl()}/fileUploadAndDownload/upload?noSave=1`"
      :on-error="uploadError"
      :on-success="uploadSuccess"
      :on-remove="uploadRemove"
      :show-file-list="true"
      :limit="limit"
      :accept="accept"
      class="upload-btn"
      :headers="{'x-token': token}"
    >
      <el-button type="primary"> {{ $t('common.uploadFile') }} </el-button>
    </el-upload>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { getBaseUrl } from '@/utils/format'
  import { useUserStore } from "@/pinia";

  defineOptions({
    name: 'UploadCommon'
  })

  const { t } = useI18n()

  defineProps({
    limit: {
      type: Number,
      default: 3
    },
    accept: {
      type: String,
      default: ''
    }
  })

  const userStore = useUserStore()

  const token = userStore.token

  const fullscreenLoading = ref(false)

  const model = defineModel({ type: Array })

  const fileList = ref(model.value || [])

  const emits = defineEmits(['on-success', 'on-error'])

  const uploadSuccess = (res) => {
    const { data, code } = res
    if (code !== 0) {
      ElMessage({
        type: 'error',
        message: t('common.uploadFailed') + res.msg
      })
      fileList.value.pop()
      return
    }
    model.value.push({
      name: data.file.name,
      url: data.file.url
    })
    emits('on-success', res)
  }

  const uploadRemove = (file) => {
    const index = model.value.indexOf(file)
    if (index > -1) {
      model.value.splice(index, 1)
      fileList.value = model.value
    }
  }

  const uploadError = (err) => {
    ElMessage({
      type: 'error',
      message: t('common.uploadFailed')
    })
    fullscreenLoading.value = false
    emits('on-error', err)
  }
</script>
