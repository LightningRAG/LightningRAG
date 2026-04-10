<template>
  <div>
    <div class="lrag-form-box">
      <el-form
        :model="formData"
        ref="elFormRef"
        label-position="right"
        :rules="rule"
        label-width="80px"
      >
        <el-form-item :label="t('example.announcement.labelTitleColon')" prop="title">
          <el-input
            v-model="formData.title"
            :clearable="true"
            :placeholder="t('example.announcement.phTitle')"
          />
        </el-form-item>
        <el-form-item :label="t('example.announcement.labelContentColon')" prop="content">
          <RichEdit v-model="formData.content" />
        </el-form-item>
        <el-form-item :label="t('example.announcement.labelAuthorColon')" prop="userID">
          <el-select
            v-model="formData.userID"
            :placeholder="t('example.announcement.phAuthor')"
            style="width: 100%"
            :clearable="true"
          >
            <el-option
              v-for="(item, key) in dataSource.userID"
              :key="key"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('example.announcement.labelAttachmentsColon')" prop="attachments">
          <SelectFile v-model="formData.attachments" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">{{ t('example.announcement.btnSave') }}</el-button>
          <el-button type="primary" @click="back">{{ t('example.announcement.btnBack') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
  import {
    getInfoDataSource,
    createInfo,
    updateInfo,
    findInfo
  } from '@/plugin/announcement/api/info'

  defineOptions({
    name: 'InfoForm'
  })

  // 自动获取字典
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { ref, reactive } from 'vue'
  import SelectFile from '@/components/selectFile/selectFile.vue'
  // 富文本组件
  import RichEdit from '@/components/richtext/rich-edit.vue'
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  const type = ref('')
  const formData = ref({
    title: '',
    content: '',
    userID: undefined,
    attachments: []
  })
  // 验证规则
  const rule = reactive({})

  const elFormRef = ref()
  const dataSource = ref([])
  const getDataSourceFunc = async () => {
    const res = await getInfoDataSource()
    if (res.code === 0) {
      dataSource.value = res.data
    }
  }
  getDataSourceFunc()

  // 初始化方法
  const init = async () => {
    // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findInfo({ ID: route.query.id })
      if (res.code === 0) {
        formData.value = res.data
        type.value = 'update'
      }
    } else {
      type.value = 'create'
    }
  }

  init()
  // 保存按钮
  const save = async () => {
    elFormRef.value?.validate(async (valid) => {
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createInfo(formData.value)
          break
        case 'update':
          res = await updateInfo(formData.value)
          break
        default:
          res = await createInfo(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: t('example.announcement.saveOk')
        })
      }
    })
  }

  // 返回按钮
  const back = () => {
    router.go(-1)
  }
</script>

<style></style>
