<template>
  <div class="h-full">
    <warning-bar
        href="https://plugin.LightningRAG.com/license"
        :title="$t('tools.skills.warnBar')"
    />
    <el-row :gutter="12" class="h-full">
      <el-col :xs="24" :sm="8" :md="6" :lg="5" class="flex flex-col gap-4 h-full">
        <el-card shadow="never" class="!border-none shrink-0">
          <div class="font-bold mb-2">{{ $t('tools.skills.aiTools') }}</div>
          <div class="flex flex-wrap gap-2">
            <div
              v-for="tool in tools"
              :key="tool.key"
              class="px-3 py-1.5 rounded-md text-sm cursor-pointer transition-all border select-none"
              :class="activeTool === tool.key
                ? 'bg-[var(--el-color-primary)] text-white border-[var(--el-color-primary)] shadow-sm'
                : 'bg-white hover:bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700'"
              @click="handleToolSelect(tool.key)"
            >
              {{ tool.label }}
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="!border-none shrink-0">
          <div class="flex justify-between items-center mb-2">
            <span class="font-bold">{{ $t('tools.skills.globalConstraint') }}</span>
            <el-button type="primary" link icon="Edit" @click="openGlobalConstraint">{{ $t('tools.skills.edit') }}</el-button>
          </div>
          <div class="text-xs text-gray-500">{{ $t('tools.skills.pathPrefix') }}: {{ globalConstraintPath }}</div>
        </el-card>

        <el-card shadow="never" class="!border-none flex-1 mt-2 flex flex-col min-h-0">
          <div class="flex justify-between items-center mb-2">
            <span class="font-bold">{{ $t('tools.skills.skillsList') }}</span>
            <div class="flex gap-1">
              <el-button type="primary" link icon="Plus" @click="openCreateDialog">{{ $t('tools.skills.add') }}</el-button>
            </div>
          </div>
          <el-input
            v-model="skillFilter"
            size="small"
            clearable
            :placeholder="$t('tools.skills.searchPh')"
            class="mb-2"
            prefix-icon="Search"
          />
          <el-scrollbar class="h-[calc(100vh-380px)]">
            <el-menu :default-active="activeSkill" class="!border-none" @select="handleSkillSelect">
              <el-menu-item
                v-for="skill in filteredSkills"
                :key="skill"
                :index="skill"
                class="!h-10 !leading-10 !my-1 !mx-1 !rounded-[4px]"
              >
                <div class="w-full flex items-center justify-between min-w-0">
                  <div class="flex items-center min-w-0 gap-1">
                    <el-icon><Document /></el-icon>
                    <span class="truncate" :title="skill">{{ skill }}</span>
                  </div>
                  <el-button
                    type="danger"
                    link
                    icon="Delete"
                    @click.stop="handleDeleteSkill(skill)"
                  >
                    {{ $t('tools.skills.delete') }}
                  </el-button>
                </div>
              </el-menu-item>
            </el-menu>
          </el-scrollbar>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="16" :md="18" :lg="19" class="h-full">
        <el-card shadow="never" class="!border-none h-full flex flex-col">
          <template v-if="!activeSkill">
            <div class="h-full flex items-center justify-center">
              <el-empty :description="$t('tools.skills.emptyPickSkill')" />
            </div>
          </template>
          <template v-else>
            <div class="flex justify-between items-center mb-4 pb-4 border-b border-gray-100 dark:border-gray-800">
              <div class="text-lg font-bold flex items-center gap-2">
                <span>{{ activeSkill }}</span>
                <el-tag size="small" type="info">Skill</el-tag>
              </div>
              <div class="flex items-center gap-2">
                <el-button icon="Download" @click="packageCurrentSkill">{{ $t('tools.skills.package') }}</el-button>
                <el-button type="primary" icon="Check" @click="saveCurrentSkill">{{ $t('tools.skills.saveConfig') }}</el-button>
              </div>
            </div>

            <el-tabs v-model="activeTab" class="h-full">
              <el-tab-pane :label="$t('tools.skills.tabConfig')" name="config">
                <div
                  class="mt-4 mb-4 rounded-md border border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/30 p-3 text-xs text-gray-600 dark:text-gray-300"
                >
                  <div class="font-medium text-gray-700 dark:text-gray-200 mb-2">{{ $t('tools.skills.guideTitle') }}</div>
                  <ul class="list-disc pl-4 space-y-1">
                    <li>{{ $t('tools.skills.guideLi1') }}</li>
                    <li>{{ $t('tools.skills.guideLi2') }}</li>
                    <li>{{ $t('tools.skills.guideLi3') }}</li>
                    <li>{{ $t('tools.skills.guideLi4') }}</li>
                  </ul>
                </div>
                <el-form :model="form" label-width="160px">
                  <el-form-item>
                    <template #label>
                      <div class="flex items-center">
                        {{ $t('tools.skills.nameLabel') }}
                        <el-tooltip :content="$t('tools.skills.tipName')" placement="top">
                          <el-icon class="ml-1 cursor-pointer"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </template>
                    <el-input v-model="form.name" :placeholder="$t('tools.skills.phNameExample')" />
                    <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintNameWords') }}</div>
                  </el-form-item>
                  <el-form-item>
                    <template #label>
                      <div class="flex items-center">
                        {{ $t('tools.skills.descLabel') }}
                        <el-tooltip :content="$t('tools.skills.tipDesc')" placement="top">
                          <el-icon class="ml-1 cursor-pointer"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </template>
                    <el-input
                      v-model="form.description"
                      :placeholder="$t('tools.skills.phDescExample')"
                    />
                    <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintDesc') }}</div>
                  </el-form-item>
                  <el-form-item>
                    <template #label>
                      <div class="flex items-center">
                        {{ $t('tools.skills.allowedToolsLabel') }}
                        <el-tooltip :content="$t('tools.skills.tipAllowedTools')" placement="top">
                          <el-icon class="ml-1 cursor-pointer"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </template>
                    <el-input v-model="form.allowedTools" :placeholder="$t('tools.skills.phAllowedTools')" />
                    <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintOptionalClear') }}</div>
                  </el-form-item>
                  <el-form-item>
                    <template #label>
                      <div class="flex items-center">
                        {{ $t('tools.skills.contextLabel') }}
                        <el-tooltip :content="$t('tools.skills.tipContext')" placement="top">
                          <el-icon class="ml-1 cursor-pointer"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </template>
                    <el-input v-model="form.context" :placeholder="$t('tools.skills.phContext')" />
                    <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintOptionalClear') }}</div>
                  </el-form-item>
                  <el-form-item>
                    <template #label>
                      <div class="flex items-center">
                        {{ $t('tools.skills.agentLabel') }}
                        <el-tooltip :content="$t('tools.skills.tipAgent')" placement="top">
                          <el-icon class="ml-1 cursor-pointer"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </template>
                    <el-input v-model="form.agent" :placeholder="$t('tools.skills.phAgent')" />
                    <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintOptionalClear') }}</div>
                  </el-form-item>
                  <el-form-item>
                    <template #label>
                      <div class="flex items-center">
                        {{ $t('tools.skills.markdownLabel') }}
                        <el-tooltip :content="$t('tools.skills.tipMarkdown')" placement="top">
                          <el-icon class="ml-1 cursor-pointer"><QuestionFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </template>
                    <div class="mb-2 flex flex-wrap gap-2">
                      <el-button
                        v-for="block in quickBlocks"
                        :key="block.label"
                        size="small"
                        @click="appendMarkdown(block.content)"
                      >
                        {{ block.label }}
                      </el-button>
                      <el-button size="small" @click="insertFullTemplate">{{ $t('tools.skills.btnInsertFullTpl') }}</el-button>
                    </div>
                    <el-input
                      v-model="form.markdown"
                      type="textarea"
                      :rows="20"
                      :placeholder="markdownPlaceholder"
                    />
                    <div class="text-xs text-gray-400 mt-1">
                      {{ $t('tools.skills.hintMarkdownShort') }}
                    </div>
                  </el-form-item>
                </el-form>
              </el-tab-pane>

              <el-tab-pane :label="$t('tools.skills.tabScripts')" name="scripts" class="mt-4">
                <div class="flex justify-between items-center mb-4">
                  <div class="text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded">{{ $t('tools.skills.pathScripts') }}</div>
                  <el-button type="primary" icon="Plus" size="small" @click="openScriptDialog">{{ $t('tools.skills.btnCreateScript') }}</el-button>
                </div>
                <div class="text-xs text-gray-500 mb-3">
                  {{ $t('tools.skills.hintScripts') }}
                </div>
                <el-table :data="scriptRows" style="width: 100%">
                  <el-table-column prop="name" :label="$t('common.colFileName')">
                    <template #default="scope">
                      <div class="flex items-center gap-2">
                        <el-icon><Document /></el-icon>
                        <span>{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column :label="$t('common.colActions')" width="180">
                    <template #default="scope">
                      <el-button type="primary" link icon="Edit" @click="openScriptEditor(scope.row.name)">{{ $t('tools.skills.btnEdit') }}</el-button>
                      <el-button type="primary" link @click="insertFileSnippet('script', scope.row.name)">{{ $t('tools.skills.btnInvoke') }}</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="scriptRows.length === 0" :description="$t('tools.skills.emptyScripts')" />
              </el-tab-pane>

              <el-tab-pane :label="$t('tools.skills.tabResources')" name="resources">
                <div class="flex justify-between items-center mb-4 mt-4">
                  <div class="text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded">{{ $t('tools.skills.pathResources') }}</div>
                  <el-button type="primary" icon="Plus" size="small" @click="openResourceDialog">{{ $t('tools.skills.btnCreateResource') }}</el-button>
                </div>
                <div class="text-xs text-gray-500 mb-3">
                  {{ $t('tools.skills.hintResources') }}
                </div>
                <el-table :data="resourceRows" style="width: 100%">
                  <el-table-column prop="name" :label="$t('common.colFileName')">
                    <template #default="scope">
                      <div class="flex items-center gap-2">
                        <el-icon><Document /></el-icon>
                        <span>{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column :label="$t('common.colActions')" width="180">
                    <template #default="scope">
                      <el-button type="primary" link icon="Edit" @click="openResourceEditor(scope.row.name)">{{ $t('tools.skills.btnEdit') }}</el-button>
                      <el-button type="primary" link @click="insertFileSnippet('resource', scope.row.name)">{{ $t('tools.skills.btnCite') }}</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="resourceRows.length === 0" :description="$t('tools.skills.emptyResources')" />
              </el-tab-pane>

              <el-tab-pane :label="$t('tools.skills.tabReferences')" name="references">
                <div class="flex justify-between items-center mb-4 mt-4">
                  <div class="text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded">{{ $t('tools.skills.pathReferences') }}</div>
                  <el-button type="primary" icon="Plus" size="small" @click="openReferenceDialog">{{ $t('tools.skills.btnCreateReference') }}</el-button>
                </div>
                <div class="text-xs text-gray-500 mb-3">
                  {{ $t('tools.skills.hintReferences') }}
                </div>
                <el-table :data="referenceRows" style="width: 100%">
                  <el-table-column prop="name" :label="$t('common.colFileName')">
                    <template #default="scope">
                      <div class="flex items-center gap-2">
                        <el-icon><Document /></el-icon>
                        <span>{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column :label="$t('common.colActions')" width="180">
                    <template #default="scope">
                      <el-button type="primary" link icon="Edit" @click="openReferenceEditor(scope.row.name)">{{ $t('tools.skills.btnEdit') }}</el-button>
                      <el-button type="primary" link @click="insertFileSnippet('reference', scope.row.name)">{{ $t('tools.skills.btnCite') }}</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="referenceRows.length === 0" :description="$t('tools.skills.emptyReferences')" />
              </el-tab-pane>

              <el-tab-pane :label="$t('tools.skills.tabTemplates')" name="templates">
                <div class="flex justify-between items-center mb-4 mt-4">
                  <div class="text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded">{{ $t('tools.skills.pathTemplates') }}</div>
                  <el-button type="primary" icon="Plus" size="small" @click="openTemplateDialog">{{ $t('tools.skills.btnCreateTemplate') }}</el-button>
                </div>
                <div class="text-xs text-gray-500 mb-3">
                  {{ $t('tools.skills.hintTemplates') }}
                </div>
                <el-table :data="templateRows" style="width: 100%">
                  <el-table-column prop="name" :label="$t('common.colFileName')">
                    <template #default="scope">
                      <div class="flex items-center gap-2">
                        <el-icon><Document /></el-icon>
                        <span>{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column :label="$t('common.colActions')" width="180">
                    <template #default="scope">
                      <el-button type="primary" link icon="Edit" @click="openTemplateEditor(scope.row.name)">{{ $t('tools.skills.btnEdit') }}</el-button>
                      <el-button type="primary" link @click="insertFileSnippet('template', scope.row.name)">{{ $t('tools.skills.btnCiteTpl') }}</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="templateRows.length === 0" :description="$t('tools.skills.emptyTemplates')" />
              </el-tab-pane>
            </el-tabs>
          </template>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="createDialogVisible" :title="$t('tools.skills.dlgNewSkill')" width="420px">
      <el-form :model="newSkill" label-width="100px">
        <el-form-item :label="$t('tools.skills.skillName')">
          <el-input v-model="newSkill.name" :placeholder="$t('tools.skills.phNameExample')" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintKebab') }}</div>
        </el-form-item>
        <el-form-item :label="$t('tools.skills.labelDesc')">
          <el-input v-model="newSkill.description" :placeholder="$t('tools.skills.phNewSkillDesc')" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintNewDesc') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="createSkill">{{ $t('tools.skills.btnCreate') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scriptDialogVisible" :title="$t('tools.skills.dlgCreateScript')" width="420px">
      <el-form :model="newScript" label-width="100px">
        <el-form-item :label="$t('tools.skills.scriptType')">
          <el-select v-model="newScript.type" :placeholder="$t('tools.skills.pickType')">
            <el-option label="Python (.py)" value="py" />
            <el-option label="JavaScript (.js)" value="js" />
            <el-option label="Shell (.sh)" value="sh" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('tools.skills.fileName')">
          <el-input v-model="newScript.name" :placeholder="$t('tools.skills.phScriptName')" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintScriptExt') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scriptDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="createScript">{{ $t('tools.skills.btnCreate') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resourceDialogVisible" :title="$t('tools.skills.dlgCreateResource')" width="420px">
      <el-form :model="newResource" label-width="100px">
        <el-form-item :label="$t('tools.skills.fileName')">
          <el-input v-model="newResource.name" :placeholder="$t('tools.skills.phResourceName')" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintAutoMd') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resourceDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="createResource">{{ $t('tools.skills.btnCreate') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="referenceDialogVisible" :title="$t('tools.skills.dlgCreateReference')" width="420px">
      <el-form :model="newReference" label-width="100px">
        <el-form-item :label="$t('tools.skills.fileName')">
          <el-input v-model="newReference.name" :placeholder="$t('tools.skills.phRefName')" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintAutoMd') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="referenceDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="createReference">{{ $t('tools.skills.btnCreate') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="templateDialogVisible" :title="$t('tools.skills.dlgCreateTemplate')" width="420px">
      <el-form :model="newTemplate" label-width="100px">
        <el-form-item :label="$t('tools.skills.fileName')">
          <el-input v-model="newTemplate.name" :placeholder="$t('tools.skills.phTplName')" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('tools.skills.hintAutoMd') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="templateDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="createTemplate">{{ $t('tools.skills.btnCreate') }}</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="editorVisible" size="70%" destroy-on-close :with-header="false">
      <div class="h-full flex flex-col p-4">
        <div class="flex justify-between items-center mb-4">
          <div class="text-lg font-bold flex items-center gap-2">
            <el-icon><Edit /></el-icon>
            {{ editorTitle }}
          </div>
          <div class="flex gap-2">
            <el-button @click="editorVisible = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" icon="Check" @click="saveEditor">{{ $t('tools.skills.editorSave') }}</el-button>
          </div>
        </div>
        <div class="flex-1 overflow-hidden border border-gray-200 dark:border-gray-700 rounded-md shadow-inner">
          <v-ace-editor
            v-model:value="editorContent"
            :lang="editorLang"
            theme="github_dark"
            class="w-full h-full"
            :options="{ showPrintMargin: false, fontSize: 14 }"
          />
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { QuestionFilled, Document, Plus, Search, Check, Edit } from '@element-plus/icons-vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import {
    getSkillTools,
    getSkillList,
    getSkillDetail,
    saveSkill,
    deleteSkill,
    createSkillScript,
    getSkillScript,
    saveSkillScript,
    createSkillResource,
    getSkillResource,
    saveSkillResource,
    createSkillReference,
    getSkillReference,
    saveSkillReference,
    createSkillTemplate,
    getSkillTemplate,
    saveSkillTemplate,
    getGlobalConstraint,
    saveGlobalConstraint,
    packageSkill
  } from '@/api/skills'
  import { VAceEditor } from 'vue3-ace-editor'
  import 'ace-builds/src-noconflict/mode-javascript'
  import 'ace-builds/src-noconflict/mode-python'
  import 'ace-builds/src-noconflict/mode-sh'
  import 'ace-builds/src-noconflict/mode-markdown'
  import 'ace-builds/src-noconflict/theme-github_dark'

  defineOptions({
    name: 'Skills'
  })

  const { t } = useI18n()

  const tools = ref([
    { key: 'copilot', label: 'Copilot' },
    { key: 'claude', label: 'Claude' },
    { key: 'cursor', label: 'Cursor' },
    { key: 'trae', label: 'Trae' },
    { key: 'codex', label: 'Codex' }
  ])
  const activeTool = ref('claude')
  const skills = ref([])
  const activeSkill = ref('')
  const skillFilter = ref('')
  const activeTab = ref('config')
  const globalConstraintExists = ref(false)

  const toolDirMap = {
    copilot: '.aone_copilot',
    claude: '.claude',
    cursor: '.cursor',
    trae: '.trae',
    codex: '.codex'
  }

  const globalConstraintPath = computed(() => {
    if (!activeTool.value) return 'skills/README.md'
    const toolDir = toolDirMap[activeTool.value] || `.${activeTool.value}`
    return `${toolDir}/skills/README.md`
  })

  const form = reactive({
    name: '',
    description: '',
    allowedTools: '',
    context: '',
    agent: '',
    markdown: ''
  })

  const markdownPlaceholder = computed(() => t('tools.skills.markdownPlaceholder'))

  const quickBlocks = computed(() => [
    { label: t('tools.skills.blockTitle'), content: t('tools.skills.quickTitle') },
    { label: t('tools.skills.blockInstr'), content: t('tools.skills.quickInstr') },
    { label: t('tools.skills.blockEx'), content: t('tools.skills.quickEx') },
    { label: t('tools.skills.blockGuide'), content: t('tools.skills.quickGuide') },
    { label: t('tools.skills.blockOut'), content: t('tools.skills.quickOut') },
    { label: t('tools.skills.blockTpl'), content: t('tools.skills.quickTpl') },
    { label: t('tools.skills.blockRef'), content: t('tools.skills.quickRef') },
    {
      label: t('tools.skills.blockScript'),
      content: t('tools.skills.quickScript', { input: t('tools.skills.snippetInputToken') })
    }
  ])

  const scripts = ref([])
  const resources = ref([])
  const references = ref([])
  const templates = ref([])

  const scriptRows = computed(() => skillsFilesToRows(scripts.value))
  const resourceRows = computed(() => skillsFilesToRows(resources.value))
  const referenceRows = computed(() => skillsFilesToRows(references.value))
  const templateRows = computed(() => skillsFilesToRows(templates.value))

  const createDialogVisible = ref(false)
  const scriptDialogVisible = ref(false)
  const resourceDialogVisible = ref(false)
  const referenceDialogVisible = ref(false)
  const templateDialogVisible = ref(false)

  const newSkill = reactive({
    name: '',
    description: ''
  })

  const newScript = reactive({
    name: '',
    type: 'py'
  })

  const newResource = reactive({
    name: ''
  })

  const newReference = reactive({
    name: ''
  })

  const newTemplate = reactive({
    name: ''
  })

  const editorVisible = ref(false)
  const editorContent = ref('')
  const editorFileName = ref('')
  const editorType = ref('script')
  const editorLang = ref('text')

  const editorTitle = computed(() => {
    if (!editorFileName.value) {
      return editorType.value === 'constraint'
        ? t('tools.skills.editorGlobal')
        : t('tools.skills.editorFile')
    }
    if (editorType.value === 'script') return t('tools.skills.editorScript', { name: editorFileName.value })
    if (editorType.value === 'resource') return t('tools.skills.editorResource', { name: editorFileName.value })
    if (editorType.value === 'reference') return t('tools.skills.editorReference', { name: editorFileName.value })
    if (editorType.value === 'template') return t('tools.skills.editorTemplate', { name: editorFileName.value })
    if (editorType.value === 'constraint') return t('tools.skills.editorConstraint', { name: editorFileName.value })
    return t('tools.skills.editorGeneric', { name: editorFileName.value })
  })

  const filteredSkills = computed(() => {
    if (!skillFilter.value) return skills.value
    return skills.value.filter((item) => item.toLowerCase().includes(skillFilter.value.toLowerCase()))
  })

  onMounted(async () => {
    await loadTools()
    await loadSkills()
  })

  async function loadTools() {
    try {
      const res = await getSkillTools()
      if (res.code === 0 && res.data?.tools?.length) {
        tools.value = res.data.tools
        if (!tools.value.find((item) => item.key === activeTool.value)) {
          activeTool.value = tools.value[0]?.key || 'claude'
        }
      }
    } catch (e) {
      ElMessage.warning(t('tools.skills.warnLoadTools'))
    }
  }

  async function loadSkills() {
    if (!activeTool.value) return
    try {
      const res = await getSkillList({ tool: activeTool.value })
      if (res.code === 0) {
        skills.value = res.data?.skills || []
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errLoadSkills'))
    }
  }

  async function loadSkillDetail(skillName) {
    if (!activeTool.value || !skillName) return
    try {
      const res = await getSkillDetail({ tool: activeTool.value, skill: skillName })
      if (res.code === 0) {
        const detail = res.data?.detail
        activeSkill.value = detail?.skill || skillName
        form.name = detail?.meta?.name || skillName
        form.description = detail?.meta?.description || ''
        form.allowedTools = detail?.meta?.allowedTools || ''
        form.context = detail?.meta?.context || ''
        form.agent = detail?.meta?.agent || ''
        form.markdown = detail?.markdown || ''
        scripts.value = detail?.scripts || []
        resources.value = detail?.resources || []
        references.value = detail?.references || []
        templates.value = detail?.templates || []
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errLoadDetail'))
    }
  }

  async function openGlobalConstraint() {
    if (!activeTool.value) {
      ElMessage.warning(t('tools.skills.warnPickTool'))
      return
    }
    try {
      const res = await getGlobalConstraint({ tool: activeTool.value })
      if (res.code === 0) {
        globalConstraintExists.value = !!res.data?.exists
        if (!globalConstraintExists.value) {
          ElMessage.info(t('tools.skills.infoNoReadme'))
        }
        openEditor('constraint', 'README.md', res.data?.content || '')
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errReadConstraint'))
    }
  }

  function resetDetail() {
    activeSkill.value = ''
    form.name = ''
    form.description = ''
    form.allowedTools = ''
    form.context = ''
    form.agent = ''
    form.markdown = ''
    scripts.value = []
    resources.value = []
    references.value = []
    templates.value = []
    activeTab.value = 'config'
  }

  function handleToolSelect(key) {
    activeTool.value = key
    resetDetail()
    globalConstraintExists.value = false
    loadSkills()
  }

  function handleSkillSelect(skillName) {
    loadSkillDetail(skillName)
  }

  async function handleDeleteSkill(skillName) {
    if (!activeTool.value || !skillName) return
    try {
      await ElMessageBox.confirm(
        t('tools.skills.confirmDeleteSkill', { name: skillName }),
        t('tools.skills.confirmDeleteTitle'),
        {
          confirmButtonText: t('tools.skills.btnConfirmDelete'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
    } catch (e) {
      return
    }

    try {
      const res = await deleteSkill({ tool: activeTool.value, skill: skillName })
      if (res.code !== 0) {
        return
      }
      if (activeSkill.value === skillName) {
        resetDetail()
      }
      await loadSkills()
      ElMessage.success(t('tools.skills.deleteOk'))
    } catch (e) {
      ElMessage.error(t('tools.skills.deleteFail'))
    }
  }

  function openCreateDialog() {
    newSkill.name = ''
    newSkill.description = ''
    createDialogVisible.value = true
  }

  async function createSkill() {
    if (!newSkill.name.trim()) {
      ElMessage.warning(t('tools.skills.warnSkillName'))
      return
    }
    const payload = {
      tool: activeTool.value,
      skill: newSkill.name.trim(),
      meta: {
        name: newSkill.name.trim(),
        description: newSkill.description.trim() || t('tools.skills.defaultDesc'),
        allowedTools: 'Bash(gh *)',
        context: 'fork',
        agent: 'Explore'
      },
      markdown: defaultSkillTemplate()
    }
    try {
      const res = await saveSkill(payload)
      if (res.code === 0) {
        ElMessage.success(t('tools.skills.createOk'))
        createDialogVisible.value = false
        await loadSkills()
        await loadSkillDetail(payload.skill)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.createFail'))
    }
  }

  async function saveCurrentSkill() {
    if (!activeSkill.value) return
    if (!form.name.trim()) {
      ElMessage.warning(t('tools.skills.warnNameEmpty'))
      return
    }
    const payload = {
      tool: activeTool.value,
      skill: activeSkill.value,
      meta: {
        name: form.name.trim(),
        description: form.description.trim(),
        allowedTools: form.allowedTools.trim(),
        context: form.context.trim(),
        agent: form.agent.trim()
      },
      markdown: form.markdown
    }

    let syncTools = []
    try {
      await ElMessageBox.confirm(t('tools.skills.syncBody'), t('tools.skills.syncTitle'), {
        confirmButtonText: t('tools.skills.syncConfirm'),
        cancelButtonText: t('tools.skills.syncOnlyCurrent'),
        type: 'warning'
      })
      syncTools = tools.value
        .map((item) => item.key)
        .filter((key) => key && key !== activeTool.value)
    } catch (e) {
      syncTools = []
    }

    if (syncTools.length) {
      payload.syncTools = syncTools
    }

    try {
      const res = await saveSkill(payload)
      if (res.code === 0) {
        ElMessage.success(t('tools.skills.saveOk'))
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.saveFail'))
    }
  }

  function extractFileNameFromDisposition(disposition) {
    if (!disposition) return ''
    const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/i)
    if (utf8Match?.[1]) {
      try {
        return decodeURIComponent(utf8Match[1])
      } catch (e) {
        return utf8Match[1]
      }
    }
    const normalMatch = disposition.match(/filename="?([^";]+)"?/i)
    return normalMatch?.[1] || ''
  }

  async function packageCurrentSkill() {
    if (!activeTool.value || !activeSkill.value) {
      ElMessage.warning(t('tools.skills.warnPickSkill'))
      return
    }
    try {
      const res = await packageSkill({ tool: activeTool.value, skill: activeSkill.value })
      const blob = res instanceof Blob ? res : (res?.data instanceof Blob ? res.data : null)
      if (!blob) {
        ElMessage.error(t('tools.skills.packageFail'))
        return
      }
      const contentType = String(res?.headers?.['content-type'] || blob.type || '').toLowerCase()
      const disposition = String(res?.headers?.['content-disposition'] || '')
      const isZipResponse = contentType.includes('application/zip') || disposition.toLowerCase().includes('filename=')
      const isErrorBlob = contentType.includes('application/json') || contentType.includes('text/plain')
      if (!isZipResponse || isErrorBlob) {
        let msg = t('tools.skills.packageFail')
        try {
          const text = await blob.text()
          if (text) {
            try {
              const json = JSON.parse(text)
              msg = json?.msg || msg
            } catch (e) {
              msg = text
            }
          }
        } catch (e) {
          // ignore parse error
        }
        ElMessage.error(msg)
        return
      }
      const fileName = extractFileNameFromDisposition(disposition) || `${activeSkill.value}.zip`
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = fileName
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      ElMessage.success(t('tools.skills.packageOk'))
    } catch (e) {
      ElMessage.error(t('tools.skills.packageFail'))
    }
  }

  function appendMarkdown(content) {
    form.markdown = `${form.markdown || ''}${content}`
  }

  function insertFileSnippet(kind, fileName) {
    if (!fileName) return
    let snippet = ''
    switch (kind) {
      case 'script':
        snippet = t('tools.skills.snippetScript', {
          file: fileName,
          input: t('tools.skills.snippetInputToken')
        })
        break
      case 'resource':
        snippet = t('tools.skills.snippetResource', { file: fileName })
        break
      case 'reference':
        snippet = t('tools.skills.snippetReference', { file: fileName })
        break
      case 'template':
        snippet = t('tools.skills.snippetTemplate', { file: fileName })
        break
      default:
        snippet = ''
    }
    if (!snippet) return
    appendMarkdown(`\n${snippet}\n`)
    ElMessage.success(t('tools.skills.inserted'))
    activeTab.value = 'config'
  }

  function insertFullTemplate() {
    if (!form.markdown.trim()) {
      form.markdown = defaultSkillTemplate()
      return
    }
    form.markdown = `${form.markdown}\n${defaultSkillTemplate()}`
  }

  function openScriptDialog() {
    if (!activeSkill.value) {
      ElMessage.warning(t('tools.skills.warnPickSkill'))
      return
    }
    newScript.name = ''
    newScript.type = 'py'
    scriptDialogVisible.value = true
  }

  async function createScript() {
    if (!newScript.name.trim()) {
      ElMessage.warning(t('tools.skills.warnScriptName'))
      return
    }
    try {
      const res = await createSkillScript({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName: newScript.name.trim(),
        scriptType: newScript.type
      })
      if (res.code === 0) {
        scriptDialogVisible.value = false
        await loadSkillDetail(activeSkill.value)
        openEditor('script', res.data.fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errCreateScript'))
    }
  }

  async function openScriptEditor(fileName) {
    if (!fileName) return
    try {
      const res = await getSkillScript({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName
      })
      if (res.code === 0) {
        openEditor('script', fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errReadScript'))
    }
  }

  function openResourceDialog() {
    if (!activeSkill.value) {
      ElMessage.warning(t('tools.skills.warnPickSkill'))
      return
    }
    newResource.name = ''
    resourceDialogVisible.value = true
  }

  async function createResource() {
    if (!newResource.name.trim()) {
      ElMessage.warning(t('tools.skills.warnResourceName'))
      return
    }
    try {
      const res = await createSkillResource({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName: newResource.name.trim()
      })
      if (res.code === 0) {
        resourceDialogVisible.value = false
        await loadSkillDetail(activeSkill.value)
        openEditor('resource', res.data.fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errCreateResource'))
    }
  }

  async function openResourceEditor(fileName) {
    if (!fileName) return
    try {
      const res = await getSkillResource({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName
      })
      if (res.code === 0) {
        openEditor('resource', fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errReadResource'))
    }
  }

  function openReferenceDialog() {
    if (!activeSkill.value) {
      ElMessage.warning(t('tools.skills.warnPickSkill'))
      return
    }
    newReference.name = ''
    referenceDialogVisible.value = true
  }

  async function createReference() {
    if (!newReference.name.trim()) {
      ElMessage.warning(t('tools.skills.warnRefName'))
      return
    }
    try {
      const res = await createSkillReference({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName: newReference.name.trim()
      })
      if (res.code === 0) {
        referenceDialogVisible.value = false
        await loadSkillDetail(activeSkill.value)
        openEditor('reference', res.data.fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errCreateReference'))
    }
  }

  async function openReferenceEditor(fileName) {
    if (!fileName) return
    try {
      const res = await getSkillReference({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName
      })
      if (res.code === 0) {
        openEditor('reference', fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errReadReference'))
    }
  }

  function openTemplateDialog() {
    if (!activeSkill.value) {
      ElMessage.warning(t('tools.skills.warnPickSkill'))
      return
    }
    newTemplate.name = ''
    templateDialogVisible.value = true
  }

  async function createTemplate() {
    if (!newTemplate.name.trim()) {
      ElMessage.warning(t('tools.skills.warnTplName'))
      return
    }
    try {
      const res = await createSkillTemplate({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName: newTemplate.name.trim()
      })
      if (res.code === 0) {
        templateDialogVisible.value = false
        await loadSkillDetail(activeSkill.value)
        openEditor('template', res.data.fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errCreateTemplate'))
    }
  }

  async function openTemplateEditor(fileName) {
    if (!fileName) return
    try {
      const res = await getSkillTemplate({
        tool: activeTool.value,
        skill: activeSkill.value,
        fileName
      })
      if (res.code === 0) {
        openEditor('template', fileName, res.data.content)
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.errReadTemplate'))
    }
  }

  function openEditor(type, fileName, content) {
    editorType.value = type
    editorFileName.value = fileName
    editorContent.value = content || ''
    editorLang.value = detectLang(fileName)
    editorVisible.value = true
  }

  async function saveEditor() {
    if (!editorFileName.value) return
    try {
      if (editorType.value === 'script') {
        const res = await saveSkillScript({
          tool: activeTool.value,
          skill: activeSkill.value,
          fileName: editorFileName.value,
          content: editorContent.value
        })
        if (res.code === 0) {
          ElMessage.success(t('tools.skills.saveOk'))
        }
      } else if (editorType.value === 'resource') {
        const res = await saveSkillResource({
          tool: activeTool.value,
          skill: activeSkill.value,
          fileName: editorFileName.value,
          content: editorContent.value
        })
        if (res.code === 0) {
          ElMessage.success(t('tools.skills.saveOk'))
        }
      } else if (editorType.value === 'reference') {
        const res = await saveSkillReference({
          tool: activeTool.value,
          skill: activeSkill.value,
          fileName: editorFileName.value,
          content: editorContent.value
        })
        if (res.code === 0) {
          ElMessage.success(t('tools.skills.saveOk'))
        }
      } else if (editorType.value === 'template') {
        const res = await saveSkillTemplate({
          tool: activeTool.value,
          skill: activeSkill.value,
          fileName: editorFileName.value,
          content: editorContent.value
        })
        if (res.code === 0) {
          ElMessage.success(t('tools.skills.saveOk'))
        }
      } else if (editorType.value === 'constraint') {
        let syncTools = []
        if (tools.value.length > 1) {
          try {
            await ElMessageBox.confirm(t('tools.skills.syncBody'), t('tools.skills.syncTitle'), {
              confirmButtonText: t('tools.skills.syncConfirm'),
              cancelButtonText: t('tools.skills.syncOnlyCurrent'),
              type: 'warning'
            })
            syncTools = tools.value
              .map((item) => item.key)
              .filter((key) => key && key !== activeTool.value)
          } catch (e) {
            syncTools = []
          }
        }

        const res = await saveGlobalConstraint({
          tool: activeTool.value,
          content: editorContent.value,
          syncTools
        })
        if (res.code !== 0) {
          ElMessage.error(t('tools.skills.saveFail'))
          return
        }
        globalConstraintExists.value = true
        ElMessage.success(syncTools.length ? t('tools.skills.saveSyncedOk') : t('tools.skills.saveOk'))
      }
    } catch (e) {
      ElMessage.error(t('tools.skills.saveFail'))
    }
  }

  function detectLang(fileName) {
    if (!fileName) return 'text'
    const lower = fileName.toLowerCase()
    if (lower.endsWith('.py')) return 'python'
    if (lower.endsWith('.js')) return 'javascript'
    if (lower.endsWith('.sh')) return 'sh'
    if (lower.endsWith('.md')) return 'markdown'
    return 'text'
  }

  function defaultSkillTemplate() {
    return t('tools.skills.defaultTemplateBody')
  }

  function skillsFilesToRows(list) {
    return (list || []).map((name) => ({ name }))
  }
</script>

