<template>
  <div class="form-designer-container">
    <fc-designer
      ref="designer"
      :config="config"
      :locale="fcDesignerLocale"
      height="calc(100vh - 160px)"
    >
      <template #handle>
        <el-button type="primary" size="small" plain @click="exportVueTemplate">
          {{ t('example.formCreate.exportVueNative') }}
        </el-button>
      </template>
    </fc-designer>

    <el-dialog v-model="dialogVisible" :title="t('example.formCreate.dialogCodeTitle')" width="70%" top="5vh">
      <el-input 
        type="textarea" 
        :rows="25" 
        v-model="vueCode" 
        readonly 
        class="code-input"
        resize="none"
      />
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('example.formCreate.close') }}</el-button>
          <el-button type="primary" @click="copyCode">{{ t('example.formCreate.copy') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import FcDesigner from '@form-create/designer/src/index.js'
  import fcDesignerEn from '@form-create/designer/src/locale/en.js'
  import fcDesignerZhCn from '@form-create/designer/src/locale/zh-cn.js'
  import { useI18n } from 'vue-i18n'

  defineOptions({
    name: 'FormGenerator'
  })

  const { t, locale } = useI18n()
  const designer = ref(null)

  /** 与站点语言同步；设计器仅内置 en / zh-cn，其余语言使用英文包 */
  const fcDesignerLocale = computed(() => {
    const code = locale.value
    if (code === 'zh-CN' || code === 'zh-TW') return fcDesignerZhCn
    return fcDesignerEn
  })
  const dialogVisible = ref(false)
  const vueCode = ref('')

  const config = {
    fieldReadonly: false,
    useTemplate: true
  }

  const kebabCase = (str) => {
    return str.replace(/([A-Z])/g, '-$1').toLowerCase()
  }

  const generateVueCode = (rules, options, tr) => {
    let formDataInit = []
    let formRules = []

    const parseRule = (rule) => {
      if (rule.type === 'row') {
        const propsStr = rule.props ? Object.entries(rule.props).map(([k, v]) => `:${k}="${v}"`).join(' ') : ''
        let childrenStr = rule.children ? rule.children.map(c => parseRule(c)).join('\n') : ''
        return `\n    <el-row ${propsStr}>${childrenStr}\n    </el-row>`
      }
      if (rule.type === 'col') {
        const propsStr = rule.props ? Object.entries(rule.props).map(([k, v]) => `:${k}="${v}"`).join(' ') : ''
        let childrenStr = rule.children ? rule.children.map(c => parseRule(c)).join('\n') : ''
        return `\n      <el-col ${propsStr}>${childrenStr}\n      </el-col>`
      }

      if (!rule.field) return ''

      let tag = rule.type
      
      const typeMap = {
        input: 'el-input',
        inputNumber: 'el-input-number',
        select: 'el-select',
        radio: 'el-radio-group',
        checkbox: 'el-checkbox-group',
        switch: 'el-switch',
        timePicker: 'el-time-picker',
        datePicker: 'el-date-picker',
        slider: 'el-slider',
        rate: 'el-rate',
        colorPicker: 'el-color-picker',
        cascader: 'el-cascader',
        upload: 'el-upload'
      }

      const elTag = typeMap[tag] || (tag.startsWith('el-') ? tag : `el-${tag}`)

      let propsStr = ''
      if (rule.props) {
        for (const [key, value] of Object.entries(rule.props)) {
          if (value === null || value === undefined) continue
          if (typeof value === 'boolean') {
            propsStr += value ? ` ${kebabCase(key)}` : ` :${kebabCase(key)}="false"`
          } else if (typeof value === 'string') {
            propsStr += ` ${kebabCase(key)}="${value}"`
          } else {
            propsStr += ` :${kebabCase(key)}='${JSON.stringify(value)}'`
          }
        }
      }

      let innerContent = ''
      if (rule.options && Array.isArray(rule.options)) {
        if (tag === 'select') {
          innerContent = rule.options.map(opt => `\n        <el-option label="${opt.label}" value="${opt.value}" />`).join('') + '\n      '
        } else if (tag === 'radio') {
          innerContent = rule.options.map(opt => `\n        <el-radio label="${opt.value}">${opt.label}</el-radio>`).join('') + '\n      '
        } else if (tag === 'checkbox') {
          innerContent = rule.options.map(opt => `\n        <el-checkbox label="${opt.value}">${opt.label}</el-checkbox>`).join('') + '\n      '
        }
      }

      let initVal = rule.value !== undefined ? rule.value : (tag === 'checkbox' ? [] : null)
      formDataInit.push(`  ${rule.field}: ${JSON.stringify(initVal)}`)

      if (rule.$required || (rule.effect && rule.effect.required)) {
        const reqMsg = tr('example.formCreate.genRequired', { field: rule.title }).replace(/'/g, "\\'")
        formRules.push(`  ${rule.field}: [{ required: true, message: '${reqMsg}', trigger: 'blur' }]`)
      } else if (rule.validate) {
        formRules.push(`  ${rule.field}: ${JSON.stringify(rule.validate)}`)
      }

      return `
    <el-form-item label="${rule.title}" prop="${rule.field}">
      <${elTag} v-model="formData.${rule.field}"${propsStr}>${innerContent}</${elTag}>
    </el-form-item>`
    }

    const formItems = rules.map(parseRule).join('')

    const formConfig = options.form || {}
    let formPropsStr = []
    if (formConfig.labelWidth) formPropsStr.push(`label-width="${formConfig.labelWidth}"`)
    if (formConfig.size) formPropsStr.push(`size="${formConfig.size}"`)
    if (formConfig.labelPosition) formPropsStr.push(`label-position="${formConfig.labelPosition}"`)
    if (formConfig.hideRequiredAsterisk) formPropsStr.push(`hide-required-asterisk`)

    // 8. 拼装成标准的 <template> 和 <script setup> 闭环代码
    return `<template>
  <div>
    <el-form ref="formRef" :model="formData" :rules="rules" ${formPropsStr.join(' ')}>
${formItems}
      <el-form-item>
        <el-button type="primary" @click="submitForm">${tr('example.formCreate.genSubmit')}</el-button>
        <el-button @click="resetForm">${tr('example.formCreate.genReset')}</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

const formRef = ref(null)

const formData = reactive({
${formDataInit.join(',\n')}
})

const rules = reactive({
${formRules.join(',\n')}
})

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      ElMessage.success(${JSON.stringify(tr('example.formCreate.genValidateOk'))})
      console.log('submit:', formData)
    } else {
      ElMessage.error(${JSON.stringify(tr('example.formCreate.genValidateFail'))})
    }
  })
}

const resetForm = () => {
  if (!formRef.value) return
  formRef.value.resetFields()
}
<\/script>
`
  }

  const exportVueTemplate = () => {
    const rules = designer.value.getRule()
    const options = designer.value.getOption()
    
    vueCode.value = generateVueCode(rules, options, t)
    dialogVisible.value = true
  }

  const copyCode = async () => {
    try {
      await navigator.clipboard.writeText(vueCode.value)
      ElMessage.success(t('example.formCreate.copyOk'))
      dialogVisible.value = false
    } catch (err) {
      ElMessage.error(t('example.formCreate.copyFail'))
    }
  }
</script>

<style scoped>

</style>

