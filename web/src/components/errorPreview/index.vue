<template>
  <div 
    class="fixed inset-0 bg-black/40 dark:bg-black/60 flex items-center justify-center z-[999]"
    @click.self="closeModal"
  >
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-dialog dark:shadow-lg w-full max-w-md mx-4 transform transition-all duration-300 ease-in-out border border-transparent dark:border-gray-700">
      <div class="p-5 border-b border-gray-100 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">{{ displayData.title }}</h3>
        <div class="text-gray-400 dark:text-gray-300 hover:text-gray-600 dark:hover:text-gray-200 transition-colors cursor-pointer" @click="closeModal">
          <close class="h-6 w-6" />
        </div>
      </div>
      
      <div class="p-6 pt-0">
        <div class="mb-4">
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">{{ $t('errorPreview.labelErrorType') }}</div>
          <div class="flex items-center gap-2">
            <lock v-if="displayData.icon === 'lock'" :class="['w-5 h-5', displayData.color]" />
            <warn v-if="displayData.icon === 'warn'" :class="['w-5 h-5', displayData.color]" />
            <server v-if="displayData.icon === 'server'" :class="['w-5 h-5', displayData.color]" />
            <span class="font-medium text-gray-800 dark:text-gray-100">{{ displayData.type }}</span>
          </div>
        </div>
        
        <div class="mb-6">
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">{{ $t('errorPreview.labelDetails') }}</div>
          <div class="bg-gray-100 dark:bg-gray-900/40 rounded-lg p-3 text-sm text-gray-700 dark:text-gray-200 leading-relaxed">
            {{ displayData.message }}
          </div>
        </div>
        
        <div v-if="displayData.tips">
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">{{ $t('errorPreview.labelTips') }}</div>
          <div class="flex items-center gap-2">
            <idea class="text-blue-500 dark:text-blue-400 w-5 h-5" />
            <p class="text-sm text-gray-600 dark:text-gray-300">{{ displayData.tips }}</p>
          </div>
        </div>
      </div>
      
      <div class="py-2 px-4 border-t border-gray-100 dark:border-gray-700 flex justify-end">
        <div class="px-4 py-2 bg-blue-600 dark:bg-blue-500 text-white dark:text-gray-100 rounded-lg hover:bg-blue-700 dark:hover:bg-blue-600 transition-colors font-medium text-sm shadow-sm cursor-pointer" @click="handleConfirm">
          {{ $t('errorPreview.ok') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps({
  errorData: {
    type: Object,
    required: true
  }
});

const emits = defineEmits(['close', 'confirm']);

const presetErrors = computed(() => ({
  500: {
    title: t('errorPreview.title500'),
    type: t('errorPreview.type500'),
    icon: 'server',
    color: 'text-red-500 dark:text-red-400',
    tips: t('errorPreview.tips500')
  },
  404: {
    title: t('errorPreview.title404'),
    type: 'Not Found',
    icon: 'warn',
    color: 'text-orange-500 dark:text-orange-400',
    tips: t('errorPreview.tips404')
  },
  401: {
    title: t('errorPreview.title401'),
    type: t('errorPreview.type401'),
    icon: 'lock',
    color: 'text-purple-500 dark:text-purple-400',
    tips: t('errorPreview.tips401')
  },
  'network': {
    title: t('errorPreview.titleNetwork'),
    type: 'Network Error',
    icon: 'fa-wifi-slash',
    color: 'text-gray-500 dark:text-gray-400',
    tips: t('errorPreview.tipsNetwork')
  }
}));

const displayData = computed(() => {
  const preset = presetErrors.value[props.errorData.code];
  if (preset) {
    return {
      ...preset,
      message: props.errorData.message || t('errorPreview.msgDefault')
    };
  }

  return {
    title: t('errorPreview.titleUnknown'),
    type: t('errorPreview.typeUnknown'),
    icon: 'fa-question-circle',
    color: 'text-gray-400 dark:text-gray-300',
    message: props.errorData.message || t('errorPreview.msgUnknown'),
    tips: t('errorPreview.tipsUnknown')
  };
});

const closeModal = () => {
   emits('close')
};

const handleConfirm = () => {
  emits('confirm', props.errorData.code);
  closeModal();
};
</script>
