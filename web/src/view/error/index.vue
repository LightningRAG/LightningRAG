<template>
  <div>
    <div class="w-full h-screen bg-gray-50 flex items-center justify-center">
      <div class="flex flex-col items-center text-2xl gap-4">
        <img class="w-1/3" src="../../assets/404.png" />
        <p class="text-lg">{{ $t('error.title') }}</p>
        <p class="text-lg">
          {{ $t('error.hint') }}
        </p>
        <p>
          {{ $t('error.projectLabel') }}<a
            href="https://github.com/LightningRAG/LightningRAG"
            target="_blank"
            class="text-blue-600"
            >https://github.com/LightningRAG/LightningRAG</a
          >
        </p>
        <el-button @click="toDashboard">{{ $t('error.backHome') }}</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { useI18n } from 'vue-i18n'
  import { useUserStore } from '@/pinia/modules/user'
  import { useRouter } from 'vue-router'
  import { emitter } from '@/utils/bus'

  const { t } = useI18n()

  defineOptions({
    name: 'Error'
  })

  const userStore = useUserStore()
  const router = useRouter()
  const toDashboard = () => {
    try {
      router.push({ name: userStore.userInfo.authority.defaultRouter })
    } catch (error) {
        emitter.emit('show-error', {
        code: '401',
        message: t('error.routeChangedRelogin'),
        fn: () => {
          userStore.ClearStorage()
          router.push({ name: 'Login', replace: true })
        }
      })
    }
  }
</script>
