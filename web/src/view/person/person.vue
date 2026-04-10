<template>
  <div class="profile-container">
    <!-- 顶部个人信息卡片 -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-sm mb-8">
      <!-- 顶部背景图 -->
      <div class="h-48 bg-blue-50 dark:bg-slate-600 relative">
        <div class="absolute inset-0 bg-pattern opacity-7"></div>
      </div>

      <!-- 个人信息区 -->
      <div class="px-8 -mt-20 pb-8">
        <div class="flex flex-col lg:flex-row items-start gap-8">
          <!-- 左侧头像 -->
          <div class="profile-avatar-wrapper flex-shrink-0 mx-auto lg:mx-0">
            <SelectImage
                v-model="userStore.userInfo.headerImg"
                file-type="image"
                rounded
            />
          </div>

          <!-- 右侧信息 -->
          <div class="flex-1 pt-12 lg:pt-20 w-full">
            <div
              class="flex flex-col lg:flex-row items-start lg:items-start justify-between gap-4"
            >
              <div class="lg:mt-4">
                <div class="flex items-center gap-4 mb-4">
                  <div
                    v-if="!editFlag"
                    class="text-2xl font-bold flex items-center gap-3 text-gray-800 dark:text-gray-100"
                  >
                    {{ userStore.userInfo.nickName }}
                    <el-icon
                      class="cursor-pointer text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors duration-200"
                      @click="openEdit"
                    >
                      <edit />
                    </el-icon>
                  </div>
                  <div v-else class="flex items-center">
                    <el-input v-model="nickName" class="w-48 mr-4" />
                    <el-button type="primary" plain @click="enterEdit">
                      {{ $t('profile.confirm') }}
                    </el-button>
                    <el-button type="danger" plain @click="closeEdit">
                      {{ $t('settings.general.cancel') }}
                    </el-button>
                  </div>
                </div>

                <div
                  class="flex flex-col lg:flex-row items-start lg:items-center gap-4 lg:gap-8 text-gray-500 dark:text-gray-400"
                >
                  <div class="flex items-center gap-2">
                    <el-icon><location /></el-icon>
                    <span>{{ $t('profile.demoLocation') }}</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <el-icon><office-building /></el-icon>
                    <span>{{ $t('profile.demoCompany') }}</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <el-icon><user /></el-icon>
                    <span>{{ $t('profile.demoDept') }}</span>
                  </div>
                </div>
              </div>

              <div class="flex gap-4 mt-4">
                <el-button type="primary" plain icon="message">
                  {{ $t('profile.sendMessage') }}
                </el-button>
                <el-button icon="share">{{ $t('profile.shareProfile') }}</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主要内容区 -->
    <div class="grid lg:grid-cols-12 md:grid-cols-1 gap-8">
      <!-- 左侧信息栏 -->
      <div class="lg:col-span-4">
        <div
          class="bg-white dark:bg-slate-800 rounded-xl p-6 mb-6 profile-card"
        >
          <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <el-icon class="text-blue-500"><info-filled /></el-icon>
            {{ $t('profile.basicInfo') }}
          </h2>
          <div class="space-y-4">
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-blue-500"><phone /></el-icon>
              <span class="font-medium">{{ $t('profile.phoneLabel') }}</span>
              <span>{{ userStore.userInfo.phone || $t('profile.notSet') }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="changePhoneFlag = true"
              >
                {{ $t('profile.edit') }}
              </el-button>
            </div>
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-green-500"><message /></el-icon>
              <span class="font-medium flex-shrink-0">{{ $t('profile.emailLabel') }}</span>
              <span>{{ userStore.userInfo.email || $t('profile.notSet') }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="changeEmailFlag = true"
              >
                {{ $t('profile.edit') }}
              </el-button>
            </div>
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-purple-500"><lock /></el-icon>
              <span class="font-medium">{{ $t('profile.passwordAccount') }}</span>
              <span>{{ $t('profile.passwordSet') }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="showPassword = true"
              >
                {{ $t('profile.edit') }}
              </el-button>
            </div>
          </div>
        </div>

        <div class="bg-white dark:bg-slate-800 rounded-xl p-6 profile-card">
          <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <el-icon class="text-blue-500"><medal /></el-icon>
            {{ $t('profile.skills') }}
          </h2>
          <div class="flex flex-wrap gap-2">
            <el-tag effect="plain" type="success">GoLang</el-tag>
            <el-tag effect="plain" type="warning">JavaScript</el-tag>
            <el-tag effect="plain" type="danger">Vue</el-tag>
            <el-tag effect="plain" type="info">Gorm</el-tag>
            <el-button link class="text-sm">
              <el-icon><plus /></el-icon>
              {{ $t('profile.addSkill') }}
            </el-button>
          </div>
        </div>
      </div>

      <!-- 右侧内容区 -->
      <div class="lg:col-span-8">
        <div class="bg-white dark:bg-slate-800 rounded-xl p-6 profile-card">
          <el-tabs class="custom-tabs">
            <el-tab-pane>
              <template #label>
                <div class="flex items-center gap-2">
                  <el-icon><data-line /></el-icon>
                  {{ $t('profile.statsTab') }}
                </div>
              </template>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 lg:gap-6 py-6">
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-blue-500 mb-2"
                  >
                    138
                  </div>
                  <div class="text-gray-500 text-sm">{{ $t('profile.statProjects') }}</div>
                </div>
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-green-500 mb-2"
                  >
                    2.3k
                  </div>
                  <div class="text-gray-500 text-sm">{{ $t('profile.statCommits') }}</div>
                </div>
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-purple-500 mb-2"
                  >
                    95%
                  </div>
                  <div class="text-gray-500 text-sm">{{ $t('profile.statTasks') }}</div>
                </div>
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-yellow-500 mb-2"
                  >
                    12
                  </div>
                  <div class="text-gray-500 text-sm">{{ $t('profile.statMedals') }}</div>
                </div>
              </div>
            </el-tab-pane>
            <el-tab-pane>
              <template #label>
                <div class="flex items-center gap-2">
                  <el-icon><calendar /></el-icon>
                  {{ $t('profile.activityTab') }}
                </div>
              </template>
              <div class="py-6">
                <el-timeline>
                  <el-timeline-item
                    v-for="(activity, index) in activities"
                    :key="index"
                    :type="activity.type"
                    :timestamp="activity.timestamp"
                    :hollow="true"
                    class="pb-6"
                  >
                    <h3 class="text-base font-medium mb-1">
                      {{ activity.title }}
                    </h3>
                    <p class="text-gray-500 text-sm">{{ activity.content }}</p>
                  </el-timeline-item>
                </el-timeline>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </div>

    <!-- 弹窗 -->
    <el-dialog
      v-model="showPassword"
      :title="$t('profile.dialogPwdTitle')"
      width="400px"
      class="custom-dialog"
      @close="clearPassword"
    >
      <el-form
        ref="modifyPwdForm"
        :model="pwdModify"
        :rules="rules"
        label-width="90px"
        class="py-4"
      >
        <el-form-item :minlength="6" :label="$t('profile.oldPassword')" prop="password">
          <el-input v-model="pwdModify.password" show-password />
        </el-form-item>
        <el-form-item :minlength="6" :label="$t('profile.newPassword')" prop="newPassword">
          <el-input v-model="pwdModify.newPassword" show-password />
        </el-form-item>
        <el-form-item :minlength="6" :label="$t('profile.confirmPassword')" prop="confirmPassword">
          <el-input v-model="pwdModify.confirmPassword" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showPassword = false">{{ $t('settings.general.cancel') }}</el-button>
          <el-button type="primary" @click="savePassword">{{ $t('settings.general.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="changePhoneFlag"
      :title="$t('profile.dialogPhoneTitle')"
      width="400px"
      class="custom-dialog"
    >
      <el-form :model="phoneForm" label-width="80px" class="py-4">
        <el-form-item :label="$t('profile.phoneField')">
          <el-input v-model="phoneForm.phone" :placeholder="$t('profile.phonePh')">
            <template #prefix>
              <el-icon><phone /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item :label="$t('profile.codeField')">
          <div class="flex gap-4">
            <el-input
              v-model="phoneForm.code"
              :placeholder="$t('profile.codePh')"
              class="flex-1"
            >
              <template #prefix>
                <el-icon><key /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              :disabled="time > 0"
              class="w-32"
              @click="getCode"
            >
              {{ time > 0 ? `${time}s` : $t('profile.getCode') }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangePhone">{{ $t('settings.general.cancel') }}</el-button>
          <el-button type="primary" @click="changePhone">{{ $t('settings.general.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="changeEmailFlag"
      :title="$t('profile.dialogEmailTitle')"
      width="400px"
      class="custom-dialog"
    >
      <el-form :model="emailForm" label-width="80px" class="py-4">
        <el-form-item :label="$t('profile.emailField')">
          <el-input v-model="emailForm.email" :placeholder="$t('profile.emailPh')">
            <template #prefix>
              <el-icon><message /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item :label="$t('profile.codeField')">
          <div class="flex gap-4">
            <el-input
              v-model="emailForm.code"
              :placeholder="$t('profile.codePh')"
              class="flex-1"
            >
              <template #prefix>
                <el-icon><key /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              :disabled="emailTime > 0"
              class="w-32"
              @click="getEmailCode"
            >
              {{ emailTime > 0 ? `${emailTime}s` : $t('profile.getCode') }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangeEmail">{{ $t('settings.general.cancel') }}</el-button>
          <el-button type="primary" @click="changeEmail">{{ $t('settings.general.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { setSelfInfo, changePassword } from '@/api/user.js'
  import { reactive, ref, watch, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/pinia/modules/user'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  defineOptions({
    name: 'Person'
  })

  const { t } = useI18n()
  const userStore = useUserStore()
  const modifyPwdForm = ref(null)
  const showPassword = ref(false)
  const pwdModify = ref({})
  const nickName = ref('')
  const editFlag = ref(false)

  const rules = computed(() => ({
    password: [
      { required: true, message: t('profile.rulesPasswordRequired'), trigger: 'blur' },
      { min: 6, message: t('profile.rulesMin6'), trigger: 'blur' }
    ],
    newPassword: [
      { required: true, message: t('profile.rulesNewPwdRequired'), trigger: 'blur' },
      { min: 6, message: t('profile.rulesMin6'), trigger: 'blur' }
    ],
    confirmPassword: [
      { required: true, message: t('profile.rulesConfirmPwdRequired'), trigger: 'blur' },
      { min: 6, message: t('profile.rulesMin6'), trigger: 'blur' },
      {
        validator: (rule, value, callback) => {
          if (value !== pwdModify.value.newPassword) {
            callback(new Error(t('profile.rulesPwdMismatch')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  }))

  const savePassword = async () => {
    modifyPwdForm.value.validate((valid) => {
      if (valid) {
        changePassword({
          password: pwdModify.value.password,
          newPassword: pwdModify.value.newPassword
        }).then((res) => {
          if (res.code === 0) {
            ElMessage.success(t('profile.pwdChangeOk'))
          }
          showPassword.value = false
        })
      }
    })
  }

  const clearPassword = () => {
    pwdModify.value = {
      password: '',
      newPassword: '',
      confirmPassword: ''
    }
    modifyPwdForm.value?.clearValidate()
  }

  const openEdit = () => {
    nickName.value = userStore.userInfo.nickName
    editFlag.value = true
  }

  const closeEdit = () => {
    nickName.value = ''
    editFlag.value = false
  }

  const enterEdit = async () => {
    const res = await setSelfInfo({
      nickName: nickName.value
    })
    if (res.code === 0) {
      userStore.ResetUserInfo({ nickName: nickName.value })
      ElMessage.success(t('profile.updateOk'))
    }
    nickName.value = ''
    editFlag.value = false
  }

  const changePhoneFlag = ref(false)
  const time = ref(0)
  const phoneForm = reactive({
    phone: '',
    code: ''
  })

  const getCode = async () => {
    time.value = 60
    let timer = setInterval(() => {
      time.value--
      if (time.value <= 0) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  }

  const closeChangePhone = () => {
    changePhoneFlag.value = false
    phoneForm.phone = ''
    phoneForm.code = ''
  }

  const changePhone = async () => {
    const res = await setSelfInfo({ phone: phoneForm.phone })
    if (res.code === 0) {
      ElMessage.success(t('profile.updateOk'))
      userStore.ResetUserInfo({ phone: phoneForm.phone })
      closeChangePhone()
    }
  }

  const changeEmailFlag = ref(false)
  const emailTime = ref(0)
  const emailForm = reactive({
    email: '',
    code: ''
  })

  const getEmailCode = async () => {
    emailTime.value = 60
    let timer = setInterval(() => {
      emailTime.value--
      if (emailTime.value <= 0) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  }

  const closeChangeEmail = () => {
    changeEmailFlag.value = false
    emailForm.email = ''
    emailForm.code = ''
  }

  const changeEmail = async () => {
    const res = await setSelfInfo({ email: emailForm.email })
    if (res.code === 0) {
      ElMessage.success(t('profile.updateOk'))
      userStore.ResetUserInfo({ email: emailForm.email })
      closeChangeEmail()
    }
  }

  watch(() => userStore.userInfo.headerImg, async(val) => {
    const res = await setSelfInfo({ headerImg: val })
    if (res.code === 0) {
      userStore.ResetUserInfo({ headerImg: val })
      ElMessage({
        type: 'success',
        message: t('profile.avatarOk'),
      })
    }
  })

  const activities = computed(() => [
    {
      timestamp: '2024-01-10',
      title: t('profile.activity1Title'),
      content: t('profile.activity1Content'),
      type: 'primary'
    },
    {
      timestamp: '2024-01-11',
      title: t('profile.activity2Title'),
      content: t('profile.activity2Content'),
      type: 'success'
    },
    {
      timestamp: '2024-01-12',
      title: t('profile.activity3Title'),
      content: t('profile.activity3Content'),
      type: 'warning'
    },
    {
      timestamp: '2024-01-13',
      title: t('profile.activity4Title'),
      content: t('profile.activity4Content'),
      type: 'danger'
    }
  ])
</script>

<style lang="scss">
  .profile-container {
    @apply p-4 lg:p-6 min-h-screen bg-gray-50 dark:bg-slate-900;

    .bg-pattern {
      background-image: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
    }

    .profile-card {
      @apply shadow-sm hover:shadow-md transition-shadow duration-300;
    }

    .profile-action-btn {
      @apply bg-white/10 hover:bg-white/20 border-white/20;
      .el-icon {
        @apply mr-1;
      }
    }

    .stat-card {
      @apply p-4 lg:p-6 rounded-lg bg-gray-50 dark:bg-slate-700/50 text-center hover:shadow-md transition-all duration-300;
    }

    .custom-tabs {
      :deep(.el-tabs__nav-wrap::after) {
        @apply h-0.5 bg-gray-100 dark:bg-gray-700;
      }
      :deep(.el-tabs__active-bar) {
        @apply h-0.5 bg-blue-500;
      }
      :deep(.el-tabs__item) {
        @apply text-base font-medium px-6;
        .el-icon {
          @apply mr-1 text-lg;
        }
        &.is-active {
          @apply text-blue-500;
        }
      }
      :deep(.el-timeline-item__node--normal) {
        @apply left-[-2px];
      }
      :deep(.el-timeline-item__wrapper) {
        @apply pl-8;
      }
      :deep(.el-timeline-item__timestamp) {
        @apply text-gray-400 text-sm;
      }
    }

    .custom-dialog {
      :deep(.el-dialog__header) {
        @apply mb-0 pb-4 border-b border-gray-100 dark:border-gray-700;
      }
      :deep(.el-dialog__footer) {
        @apply mt-0 pt-4 border-t border-gray-100 dark:border-gray-700;
      }
      :deep(.el-input__wrapper) {
        @apply shadow-none;
      }
      :deep(.el-input__prefix) {
        @apply mr-2;
      }
    }

    .edit-input {
      :deep(.el-input__wrapper) {
        @apply bg-white/10 border-white/20 shadow-none;
        input {
          @apply text-white;
          &::placeholder {
            @apply text-white/60;
          }
        }
      }
    }
  }
</style>
