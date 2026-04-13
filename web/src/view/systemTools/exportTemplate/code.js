import { i18n } from '@/locale'

export const getCode = (templateID) => {
  const t = i18n.global.t
  return `<template>
  <!-- ${t('tools.exportTemplate.codeCommentExportComp')} -->
  <ExportExcel templateId="${templateID}" :condition="condition" :limit="limit" :offset="offset" :order="order" />

  <!-- ${t('tools.exportTemplate.codeCommentImportComp')} -->
  <ImportExcel templateId="${templateID}" @on-success="handleSuccess" />

  <!-- ${t('tools.exportTemplate.codeCommentExportTplComp')} -->
  <ExportTemplate templateId="${templateID}" />
</template>

<script setup>
import { ref } from 'vue';
import ExportExcel from '@/components/exportExcel/exportExcel.vue';
import ImportExcel from '@/components/exportExcel/importExcel.vue';
import ExportTemplate from '@/components/exportExcel/exportTemplate.vue';

const condition = ref({}); // ${t('tools.exportTemplate.codeCommentCondition')}
const limit = ref(10); // ${t('tools.exportTemplate.codeCommentLimit')}
const offset = ref(0); // ${t('tools.exportTemplate.codeCommentOffset')}
const order = ref('id desc'); // ${t('tools.exportTemplate.codeCommentOrder')}

const handleSuccess = (res) => {
  // ${t('tools.exportTemplate.codeCommentImportCallback')}
};
</script>`
}
