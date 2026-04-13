<template>
  <div class="flex h-[calc(100vh-120px)] gap-4">
    <!-- 左侧会话列表 -->
    <div class="w-64 flex-shrink-0 bg-white dark:bg-slate-900 rounded p-4 flex flex-col min-h-0">
      <div class="flex justify-between items-center mb-2 shrink-0">
        <span class="font-medium">{{ $t('rag.conv.convListTitle') }}</span>
        <el-button type="primary" size="small" icon="plus" @click="createConv">{{ $t('rag.conv.listNewBtn') }}</el-button>
      </div>
      <div v-if="conversations.length" class="flex items-center gap-2 mb-2 shrink-0 flex-wrap">
        <el-checkbox
          :indeterminate="selectAllState.indeterminate"
          :model-value="selectAllState.checked"
          @change="onSelectAllChange"
        >
          <span class="text-xs">{{ $t('rag.conv.selectAll') }}</span>
        </el-checkbox>
        <el-button
          type="danger"
          size="small"
          :disabled="!selectedConvIds.length"
          @click="batchDeleteConvs"
        >
          {{ $t('rag.conv.batchDelete') }}
        </el-button>
      </div>
      <el-scrollbar class="flex-1 min-h-0">
        <div
          v-for="c in conversations"
          :key="c.ID ?? c.id"
          class="py-2 px-3 rounded mb-1 group flex items-start gap-2 justify-between hover:bg-slate-100 dark:hover:bg-slate-800"
          :class="{ 'bg-blue-50 dark:bg-blue-900/30': Number(currentConv?.ID ?? currentConv?.id) === Number(c.ID ?? c.id) }"
        >
          <el-checkbox
            class="mt-0.5 shrink-0"
            :model-value="isConvRowSelected(c)"
            @click.stop
            @change="(val) => onConvRowCheckChange(c, val)"
          />
          <div class="flex-1 min-w-0 cursor-pointer" @click="selectConv(c)">
            <div class="truncate text-sm">{{ c.title || $t('rag.conv.newChat') }}</div>
            <div class="text-xs text-gray-500 mt-1">{{ formatDate(c.CreatedAt) }}</div>
          </div>
          <el-button
            type="danger"
            link
            size="small"
            icon="delete"
            class="opacity-0 group-hover:opacity-100 shrink-0"
            @click.stop="deleteConv(c)"
          />
        </div>
      </el-scrollbar>
    </div>

    <!-- 右侧对话区 -->
    <div class="flex-1 flex flex-col bg-white dark:bg-slate-900 rounded p-4">
      <!-- 引用悬浮面板 -->
      <Teleport to="body">
        <div
          v-show="citationPopover.visible"
          ref="citationPopoverEl"
          class="citation-popover"
          :style="{ top: citationPopover.top + 'px', left: citationPopover.left + 'px' }"
          @mouseenter="onPopoverEnter"
          @mouseleave="onPopoverLeave"
        >
          <div v-if="citationPopover.ref" class="text-xs">
            <div class="font-semibold text-slate-800 dark:text-slate-100 mb-1.5 flex items-center gap-1.5">
              <span class="inline-flex items-center justify-center min-w-5 h-5 px-0.5 rounded-full bg-blue-500 text-white text-[10px] font-bold shrink-0">
                {{ citationPopover.figDisplay != null ? citationPopover.figDisplay : citationPopover.ref.index }}
              </span>
              <span class="truncate font-medium">{{ citationPopover.ref.sourceLabel || citationPopover.ref.docName || citationPopover.ref.title || $t('rag.conv.refMaterial') }}</span>
            </div>
            <div
              v-if="citationPopover.ref.docName && citationPopover.ref.title && String(citationPopover.ref.title) !== String(citationPopover.ref.docName)"
              class="text-[11px] text-slate-500 dark:text-slate-400 mb-1.5 pl-6 leading-snug"
            >
              {{ citationPopover.ref.title }}
            </div>
            <div class="citation-popover-content text-slate-600 dark:text-slate-300 leading-relaxed whitespace-pre-wrap text-[12px] bg-slate-50 dark:bg-slate-700/50 rounded-md p-2.5">{{ citationPopover.ref.content || '' }}</div>
            <div class="mt-1 text-[10px] text-slate-500 dark:text-slate-400 pl-0.5">{{ $t('rag.conv.bodyText') }} [ID:{{ citationPopover.ref.index }}]</div>
            <div
              v-if="citationPopover.ref.score != null && citationPopover.ref.score !== ''"
              class="mt-1.5 text-[10px] text-blue-500 dark:text-blue-400 font-medium pl-0.5"
            >
              {{ $t('rag.conv.relevance') }}: {{ formatRagRelevanceScore(citationPopover.ref.score) }}
            </div>
          </div>
        </div>
      </Teleport>
      <template v-if="currentConv">
        <div ref="messagesEl" class="flex-1 overflow-y-auto mb-4 space-y-4">
          <div
            v-for="(msg, idx) in messages"
            :key="msg.ID || 'stream-' + idx"
            class="flex"
            :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div
              class="max-w-[80%] rounded-lg px-4 py-2"
              :class="msg.role === 'user' ? 'bg-blue-500 text-white' : 'bg-slate-100 dark:bg-slate-800'"
            >
              <div v-if="msg.role === 'user'" class="whitespace-pre-wrap">{{ msg.content }}</div>
              <div v-else>
                <div
                  v-if="idx === messages.length - 1 && toolCallStatus"
                  class="text-sm text-amber-600 dark:text-amber-400 mb-2 flex items-center gap-1"
                >
                  <el-icon class="is-loading"><Loading /></el-icon>
                  {{ toolCallStatus }}
                </div>
                <details
                  v-if="assistantThinkingPartsByIdx[idx]?.think"
                  class="mb-2 think-block rounded border border-slate-200 dark:border-slate-600 bg-slate-50/80 dark:bg-slate-900/40"
                  :open="assistantThinkingPartsByIdx[idx]?.streaming ? true : undefined"
                >
                  <summary
                    class="px-3 py-2 cursor-pointer text-sm font-medium text-slate-600 dark:text-slate-300 select-none list-none marker:content-none flex items-center gap-1"
                  >
                    <span class="opacity-60 think-chevron inline-block transition-transform">▸</span>
                    {{
                      assistantThinkingPartsByIdx[idx]?.streaming
                        ? $t('rag.conv.reasoningStreaming')
                        : $t('rag.conv.reasoningBlock')
                    }}
                  </summary>
                  <div
                    class="think-body px-3 pb-3 whitespace-pre-wrap text-sm text-slate-600 dark:text-slate-400 max-h-72 overflow-y-auto leading-relaxed border-t border-slate-200/80 dark:border-slate-600/80"
                  >
                    {{ assistantThinkingPartsByIdx[idx]?.think }}
                  </div>
                </details>
                <div
                  class="msg-markdown"
                  v-html="renderMarkdownWithCitations(
                    assistantThinkingPartsByIdx[idx]?.main ?? '',
                    getMessageRefs(msg, idx)
                  )"
                />
              </div>
            </div>
          </div>
        </div>
        <div class="flex flex-col gap-2">
          <div class="text-xs text-gray-500 flex items-center gap-2 flex-wrap">
            <span>{{ $t('rag.conv.toolsEnabled') }}</span>
            <template v-if="enabledToolsForCurrentConv.length">
              <span
                v-for="tool in enabledToolsForCurrentConv"
                :key="tool.name"
                class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-700"
                :title="formatRagToolDescription(tool, t)"
              >{{ formatRagToolDisplayName(tool, t) }}</span>
              <el-button type="primary" link size="small" @click="openEditTools">{{ $t('rag.conv.editTools') }}</el-button>
            </template>
            <template v-else>
              <span class="text-slate-400">{{ $t('rag.conv.noToolsHint') }}</span>
              <el-button type="primary" link size="small" @click="openEditTools">{{ $t('rag.conv.selectTools') }}</el-button>
            </template>
          </div>
          <el-collapse v-model="advPanel" accordion class="conv-rag-adv mb-2 rounded border border-slate-100 dark:border-slate-800 overflow-hidden">
            <el-collapse-item name="adv" :title="$t('rag.conv.advancedRetrieval')">
              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pb-1">
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advQueryMode') }}</div>
                  <el-select v-model="advQueryMode" clearable class="w-full" size="small" :placeholder="$t('rag.retrieval.modeKbDefault')">
                    <el-option :label="$t('rag.retrieval.modeKbDefault')" value="" />
                    <el-option :label="$t('rag.kb.retrieverVector')" value="vector" />
                    <el-option :label="$t('rag.kb.retrieverKeyword')" value="keyword" />
                    <el-option :label="$t('rag.kb.retrieverLocal')" value="local" />
                    <el-option :label="$t('rag.kb.retrieverGlobal')" value="global" />
                    <el-option :label="$t('rag.kb.retrieverHybrid')" value="hybrid" />
                    <el-option :label="$t('rag.kb.retrieverMix')" value="mix" />
                    <el-option :label="$t('rag.kb.retrieverBypass')" value="bypass" />
                    <el-option :label="$t('rag.kb.retrieverPageindex')" value="pageindex" />
                  </el-select>
                </div>
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advChunkTopK') }}</div>
                  <el-input-number v-model="advChunkTopK" :min="0" :max="50" :step="1" size="small" class="w-full" />
                </div>
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advPoolTopK') }}</div>
                  <el-input-number v-model="advPoolTopK" :min="0" :max="50" :step="1" size="small" class="w-full" />
                </div>
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advRerank') }}</div>
                  <el-select v-model="advRerank" class="w-full" size="small">
                    <el-option :label="$t('rag.conv.rerankDefault')" value="default" />
                    <el-option :label="$t('rag.conv.rerankOn')" value="on" />
                    <el-option :label="$t('rag.conv.rerankOff')" value="off" />
                  </el-select>
                </div>
              </div>
              <div class="grid grid-cols-1 pb-1">
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.retrieval.labelTocEnhance') }}</div>
                  <el-select v-model="advTocEnhance" clearable class="w-full" size="small">
                    <el-option :label="$t('rag.retrieval.tocEnhanceDefault')" value="" />
                    <el-option :label="$t('rag.retrieval.tocEnhanceOn')" value="on" />
                    <el-option :label="$t('rag.retrieval.tocEnhanceOff')" value="off" />
                  </el-select>
                  <p class="text-[11px] text-gray-400 mt-1 mb-0">{{ $t('rag.retrieval.tocEnhanceHint') }}</p>
                </div>
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pb-1">
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advHlKeywords') }}</div>
                  <el-input
                    v-model="advHlKeywords"
                    size="small"
                    clearable
                    :placeholder="$t('rag.conv.advKeywordsPlaceholder')"
                  />
                </div>
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advLlKeywords') }}</div>
                  <el-input
                    v-model="advLlKeywords"
                    size="small"
                    clearable
                    :placeholder="$t('rag.conv.advKeywordsPlaceholder')"
                  />
                </div>
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pb-1">
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advCosineThreshold') }}</div>
                  <el-input-number
                    v-model="advCosineThreshold"
                    :min="0"
                    :max="1"
                    :step="0.05"
                    :precision="2"
                    size="small"
                    class="w-full"
                  />
                </div>
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advMinRerankScore') }}</div>
                  <el-input-number
                    v-model="advMinRerankScore"
                    :min="0"
                    :max="100"
                    :step="0.05"
                    :precision="4"
                    size="small"
                    class="w-full"
                  />
                </div>
                <div>
                  <div class="text-xs text-gray-500 mb-1">{{ $t('rag.conv.advMaxRagContextTokens') }}</div>
                  <el-input-number v-model="advMaxRagContextTokens" :min="0" :max="200000" :step="256" size="small" class="w-full" />
                </div>
              </div>
              <p class="text-[11px] text-gray-400 mt-1 mb-0">{{ $t('rag.conv.advancedHint') }}</p>
              <div class="mt-2">
                <el-button type="info" plain size="small" :loading="queryDataLoading" @click="runQueryDataDebug">
                  {{ $t('rag.conv.debugQueryData') }}
                </el-button>
              </div>
            </el-collapse-item>
          </el-collapse>
          <div class="flex gap-2 items-center flex-wrap">
            <el-select
              v-model="chatModelDisplayValue"
              :placeholder="$t('rag.conv.modelPlaceholder')"
              clearable
              size="small"
              style="width: 240px"
            >
              <el-option
                v-for="m in providers"
                :key="(m.source || 'user') + '-' + m.id"
                :label="`${m.name} / ${m.modelName}`"
                :value="(m.source || 'user') + ':' + m.id"
              />
            </el-select>
            <el-switch
              v-if="currentModel?.supportsDeepThinking"
              v-model="useDeepThinking"
              :active-text="$t('rag.conv.deepThink')"
              size="small"
              class="shrink-0"
            />
            <el-input
              v-model="inputText"
              type="textarea"
              :rows="2"
              :placeholder="$t('rag.conv.inputPlaceholder')"
              class="flex-1"
              @keydown.enter.exact.prevent="sendMsg"
            />
            <el-button type="primary" icon="promotion" :loading="sending" @click="sendMsg">{{ $t('rag.conv.send') }}</el-button>
          </div>
        </div>
      </template>
      <template v-else>
        <div class="flex-1 flex items-center justify-center text-gray-500">
          <div class="text-center">
            <el-icon class="text-6xl mb-4"><ChatDotRound /></el-icon>
            <p>{{ $t('rag.conv.emptyHint1') }}</p>
            <p class="text-sm mt-2">{{ $t('rag.conv.emptyHint2') }}</p>
          </div>
        </div>
      </template>
    </div>

    <!-- 新建对话弹窗 -->
    <el-dialog v-model="createVisible" :title="$t('rag.conv.dialogNewTitle')" width="450px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item :label="$t('rag.conv.labelTitle')">
          <el-input v-model="createForm.title" :placeholder="$t('rag.conv.titlePh')" />
        </el-form-item>
        <el-form-item :label="$t('rag.conv.labelModel')" required>
          <el-select v-model="createForm.llmProviderRef" :placeholder="$t('rag.conv.modelPh')" style="width: 100%">
            <el-option
              v-for="m in providers"
              :key="(m.source || 'user') + '-' + m.id"
              :label="formatModelOptionLabel(m)"
              :value="(m.source || 'user') + ':' + m.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('rag.conv.labelKb')">
          <el-select v-model="createForm.selectedKbIds" multiple :placeholder="$t('rag.conv.kbPh')" style="width: 100%">
            <el-option
              v-for="kb in knowledgeBases"
              :key="kb.ID"
              :label="kb.name"
              :value="kb.ID"
            />
          </el-select>
          <div class="text-xs text-gray-500 mt-1">{{ $t('rag.conv.kbHint') }}</div>
          <div v-if="globalKbs.length" class="text-xs mt-2 p-2 bg-blue-50 dark:bg-blue-900/20 rounded">
            <span class="text-blue-600 dark:text-blue-400 font-medium">{{ $t('rag.conv.globalKbLabel') }}</span>
            <el-tag v-for="gkb in globalKbs" :key="gkb.knowledgeBaseId" size="small" type="info" class="ml-1 mt-1">
              {{ gkb.knowledgeBaseName || $t('rag.conv.kbFallbackName', { id: gkb.knowledgeBaseId }) }}
            </el-tag>
          </div>
        </el-form-item>
        <el-form-item v-if="availableTools.length" :label="$t('rag.conv.labelTools')">
          <el-checkbox-group v-model="createForm.selectedToolNames" class="flex flex-wrap gap-2">
            <el-checkbox
              v-for="tool in availableTools"
              :key="tool.name"
              :value="tool.name"
              :title="formatRagToolDescription(tool, t)"
            >
              {{ formatRagToolDisplayName(tool, t) }}
            </el-checkbox>
          </el-checkbox-group>
          <div class="text-xs text-gray-500 mt-1">{{ $t('rag.conv.toolsHint') }}</div>
          <div v-if="selectedModelInCreate && !selectedModelInCreate.supportsToolCall" class="text-xs text-amber-600 dark:text-amber-400 mt-1">{{ $t('rag.conv.toolsNoToolCallWarning') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="doCreate">{{ $t('rag.conv.create') }}</el-button>
      </template>
    </el-dialog>

    <!-- queryData 调试结果（无 LLM 生成） -->
    <el-dialog
      v-model="queryDataVisible"
      :title="$t('rag.conv.debugQueryDataTitle')"
      width="720px"
      class="conv-query-data-dialog"
      destroy-on-close
    >
      <el-scrollbar max-height="70vh">
        <pre class="text-xs whitespace-pre-wrap break-words p-2 bg-slate-50 dark:bg-slate-900 rounded">{{ queryDataJson }}</pre>
      </el-scrollbar>
      <template #footer>
        <el-button @click="copyQueryDataResult">{{ $t('rag.conv.copyQueryDataJson') }}</el-button>
        <el-button type="primary" @click="queryDataVisible = false">{{ $t('rag.conv.debugQueryDataClose') }}</el-button>
      </template>
    </el-dialog>

    <!-- 编辑对话工具弹窗 -->
    <el-dialog v-model="editToolsVisible" :title="$t('rag.conv.dialogToolsTitle')" width="450px">
      <el-checkbox-group v-model="editToolsForm.selectedToolNames" class="flex flex-wrap gap-2">
        <el-checkbox
          v-for="tool in availableTools"
          :key="tool.name"
          :value="tool.name"
          :title="formatRagToolDescription(tool, t)"
        >
          {{ formatRagToolDisplayName(tool, t) }}
        </el-checkbox>
      </el-checkbox-group>
      <div class="text-xs text-gray-500 mt-2">{{ $t('rag.conv.toolsHint') }}</div>
      <template #footer>
        <el-button @click="editToolsVisible = false">{{ $t('settings.general.cancel') }}</el-button>
        <el-button type="primary" @click="doUpdateTools">{{ $t('rag.conv.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    createConversation,
    chatConversationStream,
    queryConversationData,
    getConversationList,
    updateConversation,
    deleteConversation,
    getConversationMessages,
    listConversationTools,
    listLLMProviders,
    getKnowledgeBaseList,
    listGlobalKnowledgeBases
  } from '@/api/rag'
  import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate, formatRagRelevanceScore } from '@/utils/format'
  import {
    formatRagToolDisplayName,
    formatRagToolDescription,
    formatRagToolNameOnly
  } from '@/utils/ragToolUi'
  import { splitModelThinking } from '@/utils/ragThinking'
  import { ChatDotRound, Loading } from '@element-plus/icons-vue'
  import { Marked } from 'marked'
  import { markedHighlight } from 'marked-highlight'
  import hljs from 'highlight.js'
  import DOMPurify from 'dompurify'

  defineOptions({ name: 'Conversation' })

  const { t } = useI18n()

  const formatModelOptionLabel = (m) => {
    if (!m) return ''
    const src =
      m.source === 'admin' ? t('rag.conv.sourceAdmin') : t('rag.conv.sourceCustom')
    return `${m.name} / ${m.modelName} (${src})${m.isDefault ? t('rag.conv.defaultTag') : ''}`
  }

  const marked = new Marked(
    markedHighlight({
      langPrefix: 'hljs language-',
      highlight(code, lang) {
        const language = hljs.getLanguage(lang) ? lang : 'plaintext'
        return hljs.highlight(code, { language }).value
      }
    })
  )

  // 与 references 目录内 citation-utils 思路类似：部分文种数字归一为 ASCII，便于匹配模型输出
  const normalizeCitationDigits = (s) => {
    if (!s) return s
    return s.replace(/[٠-٩۰-۹]/g, (ch) => {
      const c = ch.charCodeAt(0)
      if (c >= 0x0660 && c <= 0x0669) return String.fromCharCode(c - 0x0660 + 0x30)
      if (c >= 0x06f0 && c <= 0x06f9) return String.fromCharCode(c - 0x06f0 + 0x30)
      return ch
    })
  }

  // LightningRAG 旧式 ##n## → [ID:n]
  const replaceLegacyCitationMarkers = (s) =>
    s.replace(/##\s*(\d+)\s*##/g, '[ID:$1]')

  const citationMarkerRe = /\[(?:ID:)?(\d+)\]/g

  /** 旧版接口 references.index 为 1..n；新版与 LightningRAG 一致为 0..n-1（与 prompt 中 ID:0 对齐） */
  const isLegacyOneBasedRefs = (refs) => {
    if (!Array.isArray(refs) || refs.length === 0) return false
    return refs.every((r, i) => Number(r.index) === i + 1)
  }

  /** 仅在非代码区域将 [ID:n]/[n] 替换为内联角标 HTML，再交给 marked，避免代码块与行内代码内误替换（对齐 LightningRAG 思路） */
  const injectCitationSpansBeforeMarkdown = (text, references) => {
    if (!text || !references?.length) return text
    const legacy = isLegacyOneBasedRefs(references)
    const refByIndex = new Map(references.map((r) => [Number(r.index), r]))
    const toSpan = (match, idStr) => {
      const idx = parseInt(idStr, 10)
      const ref = refByIndex.get(idx)
      if (!ref) return match
      // LightningRAG UI：Fig. 为 1 起算；0-based 的 [ID:0] 显示为 1
      const display = legacy ? idx : idx + 1
      return `<span class="citation-ref" data-ref-idx="${idx}">${display}</span>`
    }
    const replaceInPlain = (segment) => segment.replace(citationMarkerRe, toSpan)
    const protectInlineCode = (segment) => {
      const ilRe = /`[^`\n]+`/g
      const parts = []
      let last = 0
      let m
      while ((m = ilRe.exec(segment)) !== null) {
        parts.push(replaceInPlain(segment.slice(last, m.index)))
        parts.push(m[0])
        last = m.index + m[0].length
      }
      parts.push(replaceInPlain(segment.slice(last)))
      return parts.join('')
    }

    const pieces = []
    let pos = 0
    while (pos < text.length) {
      const fence = text.indexOf('```', pos)
      if (fence === -1) {
        pieces.push(protectInlineCode(text.slice(pos)))
        break
      }
      if (fence > pos) {
        pieces.push(protectInlineCode(text.slice(pos, fence)))
      }
      const close = text.indexOf('```', fence + 3)
      if (close === -1) {
        pieces.push(text.slice(fence))
        break
      }
      pieces.push(text.slice(fence, close + 3))
      pos = close + 3
    }
    return pieces.join('')
  }

  const renderMarkdownWithCitations = (text, references) => {
    if (!text || !text.trim()) return ''
    try {
      let md = normalizeCitationDigits(text)
      md = replaceLegacyCitationMarkers(md)
      if (references?.length) {
        md = injectCitationSpansBeforeMarkdown(md, references)
      }
      return DOMPurify.sanitize(marked.parse(md))
    } catch {
      return text
    }
  }

  // 获取消息的引用列表（统一处理当前流式消息和历史消息）
  // 优先使用流式期间的 pendingRefs（仅对最后一条消息），
  // 其次使用消息自身的 references（流结束后或历史消息）
  const getMessageRefs = (msg, idx) => {
    const isLast = idx === messages.value.length - 1
    const pending = isLast ? pendingRefs.value : []
    if (pending.length > 0) return pending
    const msgRefs = msg?.references
    if (Array.isArray(msgRefs) && msgRefs.length > 0) return msgRefs
    return []
  }

  // 引用悬浮面板
  const citationPopoverEl = ref(null)
  const citationPopover = ref({ visible: false, top: 0, left: 0, ref: null, figDisplay: null })
  let citationHideTimer = null
  const pendingRefs = ref([])

  const findRefForElement = (el) => {
    const idx = parseInt(el.dataset.refIdx, 10)
    if (isNaN(idx)) return null
    const msgEl = el.closest('.flex.justify-start')
    if (!msgEl) return null
    const msgIdx = Array.from(messagesEl.value?.children || []).indexOf(msgEl)
    if (msgIdx < 0) return null
    const msg = messages.value[msgIdx]
    return getMessageRefs(msg, msgIdx).find(r => Number(r.index) === idx) || null
  }

  const isInsideCitationPopover = (el) => {
    if (!el || typeof el.closest !== 'function') return false
    return !!el.closest('.citation-popover')
  }

  /** 与 LightningRAG HoverCard 一致：仅悬停在引用角标上打开；可移入浮层内滚动；离开浮层即关闭 */
  const CITATION_SIDE_OFFSET = 4
  let lastCitationEl = null
  let citationHoverBoundEl = null

  const showPopoverForCitationEl = (target) => {
    const refData = findRefForElement(target)
    if (!refData) return
    const msgEl = target.closest('.flex.justify-start')
    const msgIdx = msgEl ? Array.from(messagesEl.value?.children || []).indexOf(msgEl) : -1
    const msg = msgIdx >= 0 ? messages.value[msgIdx] : null
    const list = msg ? getMessageRefs(msg, msgIdx) : []
    const legacy = isLegacyOneBasedRefs(list)
    const figDisplay = legacy ? Number(refData.index) : Number(refData.index) + 1
    const rect = target.getBoundingClientRect()
    const popoverHeight = 260
    const spaceBelow = window.innerHeight - rect.bottom
    const top =
      spaceBelow < popoverHeight
        ? rect.top - popoverHeight - CITATION_SIDE_OFFSET
        : rect.bottom + CITATION_SIDE_OFFSET
    citationPopover.value = {
      visible: true,
      top: Math.max(8, top),
      left: Math.min(rect.left, window.innerWidth - 400),
      ref: refData,
      figDisplay
    }
  }

  const scheduleCitationHide = (delayMs) => {
    clearTimeout(citationHideTimer)
    citationHideTimer = setTimeout(() => {
      citationPopover.value = { ...citationPopover.value, visible: false }
      lastCitationEl = null
    }, delayMs)
  }

  const onCitationMouseOverCapture = (e) => {
    const ref = e.target.closest?.('.citation-ref')
    if (!ref || !messagesEl.value?.contains(ref)) return
    clearTimeout(citationHideTimer)
    if (ref !== lastCitationEl) {
      lastCitationEl = ref
      showPopoverForCitationEl(ref)
    }
  }

  const onCitationMouseOutCapture = (e) => {
    const fromRef = e.target?.closest?.('.citation-ref')
    if (!fromRef || !messagesEl.value?.contains(fromRef)) return
    const to = e.relatedTarget
    if (to && fromRef.contains(to)) return
    if (isInsideCitationPopover(to)) return
    const popEl = citationPopoverEl.value
    if (popEl && to && popEl.contains(to)) return
    scheduleCitationHide(280)
  }

  const bindCitationHoverDelegates = () => {
    const el = messagesEl.value
    if (!el) return
    if (citationHoverBoundEl === el) return
    if (citationHoverBoundEl) {
      unbindCitationHoverDelegates()
    }
    el.addEventListener('mouseover', onCitationMouseOverCapture, true)
    el.addEventListener('mouseout', onCitationMouseOutCapture, true)
    el.dataset.citationHoverBound = '1'
    citationHoverBoundEl = el
  }

  const unbindCitationHoverDelegates = () => {
    const el = citationHoverBoundEl
    if (!el || el.dataset.citationHoverBound !== '1') {
      citationHoverBoundEl = null
      return
    }
    el.removeEventListener('mouseover', onCitationMouseOverCapture, true)
    el.removeEventListener('mouseout', onCitationMouseOutCapture, true)
    delete el.dataset.citationHoverBound
    citationHoverBoundEl = null
  }

  const onPopoverEnter = () => {
    clearTimeout(citationHideTimer)
  }

  const onPopoverLeave = (e) => {
    const to = e.relatedTarget
    const popEl = citationPopoverEl.value
    if (popEl && to && popEl.contains(to)) return
    const backRef = to?.closest?.('.citation-ref')
    if (backRef && messagesEl.value?.contains(backRef)) return
    scheduleCitationHide(0)
  }

  const conversations = ref([])
  /** 批量删除：对话主键 ID 列表 */
  const selectedConvIds = ref([])
  const currentConv = ref(null)

  const convRowId = (c) => {
    const v = c?.ID ?? c?.id
    const n = Number(v)
    return Number.isNaN(n) ? NaN : n
  }

  const allSelectableConvIds = computed(() =>
    conversations.value.map(convRowId).filter((id) => !Number.isNaN(id))
  )

  const selectAllState = computed(() => {
    const ids = allSelectableConvIds.value
    const sel = selectedConvIds.value
    if (!ids.length) return { checked: false, indeterminate: false }
    const inPage = sel.filter((id) => ids.includes(id)).length
    if (inPage === 0) return { checked: false, indeterminate: false }
    if (inPage === ids.length) return { checked: true, indeterminate: false }
    return { checked: false, indeterminate: true }
  })

  const isConvRowSelected = (c) => {
    const id = convRowId(c)
    return !Number.isNaN(id) && selectedConvIds.value.includes(id)
  }

  const onSelectAllChange = (checked) => {
    if (checked) {
      selectedConvIds.value = [...allSelectableConvIds.value]
    } else {
      selectedConvIds.value = []
    }
  }

  const onConvRowCheckChange = (c, checked) => {
    const id = convRowId(c)
    if (Number.isNaN(id)) return
    const set = new Set(selectedConvIds.value)
    if (checked) set.add(id)
    else set.delete(id)
    selectedConvIds.value = [...set]
  }

  watch(conversations, (list) => {
    const valid = new Set(list.map(convRowId).filter((x) => !Number.isNaN(x)))
    selectedConvIds.value = selectedConvIds.value.filter((id) => valid.has(id))
  })
  const messages = ref([])
  const inputText = ref('')
  const providers = ref([])
  const createVisible = ref(false)
  const knowledgeBases = ref([])
  const createForm = ref({
    title: '',
    llmProviderRef: '', // 格式 "user:1" 或 "admin:1"
    selectedKbIds: [],
    selectedToolNames: [] // 用户选中的工具，默认全不选
  })
  const chatModelOverride = ref('') // 临时切换模型，格式 "user:1" 或 "admin:1"，空则用对话创建时的模型
  // 下拉框显示值：有覆盖时显示覆盖，否则显示当前对话的模型（便于未切换时也能看到在用哪个模型）
  const chatModelDisplayValue = computed({
    get: () => chatModelOverride.value || currentModelRef.value || null,
    set: (val) => { chatModelOverride.value = (val == null || val === '') ? '' : val }
  })
  const useDeepThinking = ref(false) // 是否启用深度思考（仅当模型支持时有效）
  /** 与 server/service/rag/lightrag_defaults.go DefaultConversationRAGTopK 一致；config 为 0 时服务端即用此值 */
  const RAG_CODE_DEFAULT_CHUNK_TOP_K = 20
  /** LightningRAG 风格可选检索参数（折叠面板） */
  const advPanel = ref('')
  const advQueryMode = ref('')
  const advChunkTopK = ref(0)
  const advPoolTopK = ref(0)
  const advRerank = ref('default')
  /** PageIndex：''=不传 tocEnhance；on/off → 与 Ragflow toc_enhance 同义 */
  const advTocEnhance = ref('')
  const advHlKeywords = ref('')
  const advLlKeywords = ref('')
  const advCosineThreshold = ref(0)
  const advMinRerankScore = ref(0)
  const advMaxRagContextTokens = ref(0)
  const queryDataVisible = ref(false)
  const queryDataResult = ref(null)
  const queryDataLoading = ref(false)
  const queryDataJson = computed(() => {
    try {
      return JSON.stringify(queryDataResult.value ?? {}, null, 2)
    } catch {
      return String(queryDataResult.value)
    }
  })

  const parseCommaKeywordList = (s) => {
    const raw = (s == null ? '' : String(s)).trim()
    if (!raw) return null
    const parts = raw.split(/[,，;；\n]+/).map((x) => x.trim()).filter(Boolean)
    return parts.length ? parts : null
  }

  const mergeChatAdvancedPayload = (payload) => {
    const m = (advQueryMode.value || '').trim()
    if (m) payload.queryMode = m
    const ck = Number(advChunkTopK.value)
    if (Number.isFinite(ck) && ck > 0) {
      payload.chunkTopK = Math.min(50, Math.floor(ck))
    }
    const pk = Number(advPoolTopK.value)
    const baseN =
      payload.chunkTopK && payload.chunkTopK > 0 ? payload.chunkTopK : RAG_CODE_DEFAULT_CHUNK_TOP_K
    if (Number.isFinite(pk) && pk > 0 && pk > baseN) {
      payload.topK = Math.min(50, Math.floor(pk))
    }
    if (advRerank.value === 'on') payload.enableRerank = true
    else if (advRerank.value === 'off') payload.enableRerank = false
    if (advTocEnhance.value === 'on') payload.tocEnhance = true
    else if (advTocEnhance.value === 'off') payload.tocEnhance = false
    const hl = parseCommaKeywordList(advHlKeywords.value)
    if (hl) payload.hlKeywords = hl
    const ll = parseCommaKeywordList(advLlKeywords.value)
    if (ll) payload.llKeywords = ll
    const ct = Number(advCosineThreshold.value)
    if (Number.isFinite(ct) && ct > 0) payload.cosineThreshold = ct
    const mr = Number(advMinRerankScore.value)
    if (Number.isFinite(mr) && mr > 0) payload.minRerankScore = mr
    const mrt = Number(advMaxRagContextTokens.value)
    if (Number.isFinite(mrt) && mrt > 0) payload.maxRagContextTokens = Math.min(200000, Math.floor(mrt))
  }

  const copyQueryDataResult = async () => {
    const text = queryDataJson.value
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success(t('rag.conv.copyQueryDataOk'))
    } catch {
      ElMessage.warning(t('rag.conv.copyQueryDataFail'))
    }
  }

  const runQueryDataDebug = async () => {
    const convId = currentConv.value?.ID || currentConv.value?.id
    if (!convId) return
    const q = (inputText.value || '').trim()
    if (q.length < 3) {
      ElMessage.warning(t('rag.conv.warnQueryDataShort'))
      return
    }
    const payload = { conversationId: convId, query: q }
    mergeChatAdvancedPayload(payload)
    const refM = chatModelOverride.value
    if (refM) {
      const [source, idStr] = refM.split(':')
      const id = parseInt(idStr, 10)
      if (!isNaN(id)) {
        payload.llmProviderId = id
        payload.llmSource = source || 'user'
      }
    }
    queryDataLoading.value = true
    try {
      const res = await queryConversationData(payload)
      if (res.code === 0) {
        queryDataResult.value = res.data
        queryDataVisible.value = true
      } else {
        ElMessage.error(res.msg || t('rag.conv.debugQueryDataFail'))
      }
    } catch (e) {
      ElMessage.error(e?.message || t('rag.conv.debugQueryDataFail'))
    } finally {
      queryDataLoading.value = false
    }
  }
  const availableTools = ref([]) // 全部可选的工具列表
  const editToolsVisible = ref(false)
  const editToolsForm = ref({ selectedToolNames: [] })
  // 当前对话启用的工具（仅展示用户选中的）
  const enabledToolsForCurrentConv = computed(() => {
    if (!currentConv.value?.enabledToolNames) return []
    try {
      const raw = currentConv.value.enabledToolNames
      const names = typeof raw === 'string' ? JSON.parse(raw || '[]') : (raw || [])
      return availableTools.value.filter((toolDef) => names.includes(toolDef.name))
    } catch {
      return []
    }
  })

  // 当前对话使用的模型（用于判断是否显示深度思考、工具选择）
  const currentModelRef = computed(() => {
    if (chatModelOverride.value) return chatModelOverride.value
    const c = currentConv.value
    if (!c || (c.llmProviderId == null || c.llmProviderId === undefined)) return ''
    return (c.llmSource || 'user') + ':' + c.llmProviderId
  })
  const currentModel = computed(() => {
    const ref = currentModelRef.value
    if (!ref) return null
    return providers.value.find(m => ((m.source || 'user') + ':' + m.id) === ref) || null
  })

  // 新建对话时选中的模型（用于判断是否显示工具选择）
  const selectedModelInCreate = computed(() => {
    const ref = createForm.value.llmProviderRef
    if (!ref) return null
    return providers.value.find(m => ((m.source || 'user') + ':' + m.id) === ref) || null
  })

  const RAG_LAST_CONV_KEY = 'rag_last_conversation_id'

  const loadConversations = async (opts = {}) => {
    const res = await getConversationList({ page: 1, pageSize: 50 })
    if (res.code === 0) {
      conversations.value = res.data?.list || []
      if (opts.skipRestoreSelection) return
      // 刷新后恢复上次选中的对话
      const lastId = localStorage.getItem(RAG_LAST_CONV_KEY)
      if (lastId) {
        const id = parseInt(lastId, 10)
        const found = conversations.value.find(c => Number(c.ID || c.id) === id)
        if (found) {
          currentConv.value = found
        }
      }
    }
  }

  const loadProviders = async () => {
    const res = await listLLMProviders({ scenarioType: 'chat' })
    if (res.code === 0) {
      providers.value = res.data || []
    }
  }

  const globalKbs = ref([])

  const loadKnowledgeBases = async () => {
    const res = await getKnowledgeBaseList({ page: 1, pageSize: 100 })
    if (res.code === 0 && res.data?.list) {
      knowledgeBases.value = res.data.list
    }
  }

  const loadGlobalKbs = async () => {
    const res = await listGlobalKnowledgeBases()
    if (res.code === 0) {
      globalKbs.value = res.data || []
    }
  }

  const loadTools = async () => {
    const res = await listConversationTools()
    if (res.code === 0 && res.data?.length) {
      availableTools.value = res.data
    }
  }

  onMounted(() => {
    loadConversations()
    loadProviders()
    loadKnowledgeBases()
    loadGlobalKbs()
    loadTools()
    nextTick(() => bindCitationHoverDelegates())
  })

  watch(currentConv, (c) => {
    if (!c) {
      clearTimeout(citationHideTimer)
      citationPopover.value = { ...citationPopover.value, visible: false }
      lastCitationEl = null
      unbindCitationHoverDelegates()
    }
    nextTick(() => bindCitationHoverDelegates())
  }, { flush: 'post' })

  onUnmounted(() => {
    clearTimeout(citationHideTimer)
    unbindCitationHoverDelegates()
  })

  const scrollToBottom = () => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  }

  const loadMessages = async (conv) => {
    const convId = conv?.ID || conv?.id
    if (!convId) return
    const res = await getConversationMessages({ conversationId: convId })
    if (res.code === 0 && res.data?.list) {
      messages.value = res.data.list.map(m => {
        let refs = m.references || []
        if (typeof refs === 'string') {
          try { refs = JSON.parse(refs) } catch { refs = [] }
        }
        return { ID: m.ID, role: m.role, content: m.content || '', references: refs }
      })
      nextTick(() => scrollToBottom())
    } else {
      messages.value = []
    }
  }

  watch(currentConv, (conv) => {
    if (conv) {
      loadMessages(conv)
    }
  })

  const selectConv = (c) => {
    currentConv.value = c
    if (c?.ID || c?.id) {
      localStorage.setItem(RAG_LAST_CONV_KEY, String(c.ID || c.id))
    }
  }

  const createConv = () => {
    const list = providers.value || []
    // 优先使用默认模型（用户默认 > 角色默认），否则 Ollama，否则第一个
    const defaultOne = list.find(m => m.isDefault)
    const ollama = list.find(m => (m.name || '').toLowerCase() === 'ollama')
    const first = defaultOne || ollama || list[0]
    createForm.value = {
      title: t('rag.conv.newChat'),
      llmProviderRef: first ? (first.source || 'user') + ':' + first.id : '',
      selectedKbIds: [],
      selectedToolNames: []
    }
    createVisible.value = true
  }

  const doCreate = async () => {
    const ref = createForm.value.llmProviderRef
    // 不选则使用默认（用户默认>角色默认），后端会解析
    let llmProviderId = 0
    let llmSource = ''
    if (ref) {
      const [source, idStr] = ref.split(':')
      const id = parseInt(idStr, 10)
      if (isNaN(id)) {
        ElMessage.warning(t('rag.conv.invalidModel'))
        return
      }
      llmProviderId = id
      llmSource = source || 'user'
    }
    const kbIds = createForm.value.selectedKbIds || []
    const sourceIds = JSON.stringify(kbIds.map(k => String(k)))
    const res = await createConversation({
      title: createForm.value.title,
      llmProviderId,
      llmSource: llmSource || undefined,
      sourceType: 'knowledge_base',
      sourceIds,
      enabledToolNames: createForm.value.selectedToolNames || []
    })
    if (res.code === 0) {
      createVisible.value = false
      const newId = res.data?.ID ?? res.data?.id
      const merged = {
        ...res.data,
        llmProviderId: res.data?.llmProviderId ?? llmProviderId,
        llmSource: res.data?.llmSource || llmSource || 'user'
      }
      // 勿用旧 lastId 恢复选中，否则会短暂或错误地留在上一会话；列表无排序/分页时可能不含新会话，需兜底插入
      await loadConversations({ skipRestoreSelection: true })
      const nid = newId != null && newId !== '' ? Number(newId) : NaN
      if (!Number.isNaN(nid)) {
        let fromList = conversations.value.find(c => Number(c.ID || c.id) === nid)
        if (!fromList) {
          conversations.value = [merged, ...conversations.value.filter(c => Number(c.ID || c.id) !== nid)]
          fromList = conversations.value[0]
        }
        currentConv.value = fromList
        localStorage.setItem(RAG_LAST_CONV_KEY, String(newId))
      }
      ElMessage.success(t('rag.conv.createOk'))
    }
  }

  const sending = ref(false)
  const streamingContent = ref('')
  /** 每条消息的 think/main 拆分：随 messages + streamingContent 一次计算，避免 v-for 内重复 splitModelThinking */
  const assistantThinkingPartsByIdx = computed(() => {
    const arr = messages.value
    const n = arr.length
    const tail = streamingContent.value || ''
    const out = []
    for (let idx = 0; idx < n; idx++) {
      const raw = (arr[idx].content || '') + (idx === n - 1 ? tail : '')
      out.push(splitModelThinking(raw))
    }
    return out
  })
  const toolCallStatus = ref('') // 工具调用状态，如 "正在搜索..."
  const messagesEl = ref(null)
  const sendMsg = async () => {
    const text = inputText.value?.trim()
    if (!text || !currentConv.value || sending.value) return
    messages.value.push({ role: 'user', content: text })
    messages.value.push({ role: 'assistant', content: '' })
    inputText.value = ''
    sending.value = true
    streamingContent.value = ''
    toolCallStatus.value = ''
    pendingRefs.value = []
    const payload = { conversationId: currentConv.value.ID || currentConv.value.id, content: text, useDeepThinking: useDeepThinking.value }
    const ref = chatModelOverride.value
    if (ref) {
      const [source, idStr] = ref.split(':')
      const id = parseInt(idStr, 10)
      if (!isNaN(id)) {
        payload.llmProviderId = id
        payload.llmSource = source || 'user'
      }
    }
    mergeChatAdvancedPayload(payload)
    try {
      await chatConversationStream(
        payload,
        {
          onChunk: (chunk) => {
            toolCallStatus.value = ''
            streamingContent.value += chunk
            nextTick(() => scrollToBottom())
          },
          onToolCall: (name, status, _result, toolCall) => {
            if (status === 'start') {
              const displayName = formatRagToolNameOnly(name, toolCall?.displayName, t)
              toolCallStatus.value = t('rag.conv.toolCalling', { name: displayName })
            } else {
              toolCallStatus.value = ''
            }
            nextTick(() => scrollToBottom())
          },
          onReferences: (refs) => {
            pendingRefs.value = refs || []
          },
          onDone: (references = [], _meta) => {
            const last = messages.value[messages.value.length - 1]
            if (last?.role === 'assistant') {
              last.content += streamingContent.value
              last.references = references?.length ? references : pendingRefs.value
            }
            streamingContent.value = ''
            pendingRefs.value = []
            sending.value = false
            nextTick(() => scrollToBottom())
          },
          onError: (err) => {
            const last = messages.value[messages.value.length - 1]
            if (last?.role === 'assistant') {
              last.content = (last.content || '') + (streamingContent.value || '') + t('rag.conv.errorPrefix') + err
            }
            streamingContent.value = ''
            toolCallStatus.value = ''
            sending.value = false
            ElMessage.error(err || t('rag.conv.streamNotifyFallback'))
          }
        }
      )
    } catch (e) {
      const last = messages.value[messages.value.length - 1]
      if (last?.role === 'assistant') {
        last.content = (last.content || '') + streamingContent.value + t('rag.conv.errorPrefix') + e.message
      }
      streamingContent.value = ''
      toolCallStatus.value = ''
      sending.value = false
      ElMessage.error(e.message || t('rag.conv.streamNotifyFallback'))
    }
  }

  const openEditTools = () => {
    if (!currentConv.value) return
    try {
      const raw = currentConv.value.enabledToolNames
      const names = typeof raw === 'string' ? JSON.parse(raw || '[]') : (raw || [])
      editToolsForm.value.selectedToolNames = Array.isArray(names) ? [...names] : []
    } catch {
      editToolsForm.value.selectedToolNames = []
    }
    editToolsVisible.value = true
  }

  const doUpdateTools = async () => {
    const conv = currentConv.value
    if (!conv?.ID && !conv?.id) return
    const res = await updateConversation({
      id: conv.ID || conv.id,
      enabledToolNames: editToolsForm.value.selectedToolNames || []
    })
    if (res.code === 0) {
      editToolsVisible.value = false
      // 更新当前对话的 enabledToolNames
      const names = editToolsForm.value.selectedToolNames || []
      currentConv.value = { ...conv, enabledToolNames: JSON.stringify(names) }
      // 同步更新列表中的对话
      const idx = conversations.value.findIndex(c => (c.ID || c.id) === (conv.ID || conv.id))
      if (idx >= 0) {
        conversations.value[idx] = { ...conversations.value[idx], enabledToolNames: JSON.stringify(names) }
      }
      ElMessage.success(t('rag.conv.updateOk'))
    }
  }

  const syncCurrentConvAfterListReload = (deletedIds) => {
    const curId = Number(currentConv.value?.ID ?? currentConv.value?.id)
    if (Number.isNaN(curId)) return
    if (!deletedIds.includes(curId)) {
      const still = conversations.value.find((x) => convRowId(x) === curId)
      if (still) currentConv.value = still
      return
    }
    localStorage.removeItem(RAG_LAST_CONV_KEY)
    currentConv.value = conversations.value[0] || null
    if (currentConv.value) {
      localStorage.setItem(RAG_LAST_CONV_KEY, String(currentConv.value.ID || currentConv.value.id))
    }
  }

  const deleteConv = (c) => {
    const cid = c.ID || c.id
    ElMessageBox.confirm(
      t('rag.conv.deleteConfirm', { title: c.title || t('rag.conv.newChat') }),
      t('rag.conv.deleteTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      const res = await deleteConversation({ id: cid })
      if (res.code === 0) {
        ElMessage.success(t('rag.conv.deleteOk'))
        const nid = Number(cid)
        selectedConvIds.value = selectedConvIds.value.filter((x) => x !== nid)
        const wasCurrent =
          Number(currentConv.value?.ID ?? currentConv.value?.id) === nid
        if (wasCurrent) {
          localStorage.removeItem(RAG_LAST_CONV_KEY)
        }
        await loadConversations()
        syncCurrentConvAfterListReload([nid])
      }
    }).catch(() => {})
  }

  const batchDeleteConvs = () => {
    if (!selectedConvIds.value.length) {
      ElMessage.warning(t('rag.conv.batchDeleteNone'))
      return
    }
    const ids = [...selectedConvIds.value]
    const count = ids.length
    ElMessageBox.confirm(
      t('rag.conv.batchDeleteConfirm', { count }),
      t('rag.conv.deleteTitle'),
      {
        confirmButtonText: t('settings.general.confirm'),
        cancelButtonText: t('settings.general.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      let fail = 0
      for (const id of ids) {
        const res = await deleteConversation({ id })
        if (res.code !== 0) fail++
      }
      selectedConvIds.value = []
      const curId = Number(currentConv.value?.ID ?? currentConv.value?.id)
      if (!Number.isNaN(curId) && ids.includes(curId)) {
        localStorage.removeItem(RAG_LAST_CONV_KEY)
      }
      await loadConversations()
      syncCurrentConvAfterListReload(ids)
      if (fail === 0) {
        ElMessage.success(t('rag.conv.batchDeleteOk', { count }))
      } else {
        ElMessage.warning(
          t('rag.conv.batchDeletePartial', { ok: count - fail, fail })
        )
      }
    }).catch(() => {})
  }
</script>

<style scoped>
  @import 'highlight.js/styles/github.css';
  .think-block[open] .think-chevron {
    transform: rotate(90deg);
  }
  .conv-rag-adv :deep(.el-collapse-item__header) {
    font-size: 12px;
    padding: 8px 12px;
    min-height: 36px;
    background: transparent;
  }
  .conv-rag-adv :deep(.el-collapse-item__wrap) {
    background: transparent;
  }
  .conv-rag-adv :deep(.el-collapse-item__content) {
    padding: 8px 12px 12px;
  }
  .msg-markdown :deep(p) { margin: 0.5em 0; }
  .msg-markdown :deep(p:first-child) { margin-top: 0; }
  .msg-markdown :deep(p:last-child) { margin-bottom: 0; }
  .msg-markdown :deep(ul), .msg-markdown :deep(ol) { margin: 0.5em 0; padding-left: 1.5em; }
  .msg-markdown :deep(li) { margin: 0.25em 0; }
  .msg-markdown :deep(h1), .msg-markdown :deep(h2), .msg-markdown :deep(h3) { margin: 0.75em 0 0.5em; font-weight: 600; }
  .msg-markdown :deep(blockquote) { margin: 0.5em 0; padding-left: 1em; border-left: 4px solid #94a3b8; opacity: 0.9; }
  .msg-markdown :deep(code) { padding: 0.2em 0.4em; border-radius: 4px; font-size: 0.9em; background: rgba(0,0,0,0.08); }
  .msg-markdown :deep(pre) { margin: 0.5em 0; padding: 0.75em; border-radius: 6px; overflow-x: auto; background: rgba(0,0,0,0.06); }
  .msg-markdown :deep(pre code) { padding: 0; background: none; }
  .msg-markdown :deep(a) { color: #3b82f6; text-decoration: underline; }
  .msg-markdown :deep(table) { border-collapse: collapse; margin: 0.5em 0; }
  .msg-markdown :deep(th), .msg-markdown :deep(td) { border: 1px solid rgba(0,0,0,0.1); padding: 0.25em 0.5em; }
  .dark .msg-markdown :deep(code) { background: rgba(255,255,255,0.1); }
  .dark .msg-markdown :deep(pre) { background: rgba(255,255,255,0.06); }
  .dark .msg-markdown :deep(blockquote) { border-left-color: #64748b; }

  /* 内联引用标记（上标数字徽标） */
  .msg-markdown :deep(.citation-ref) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 16px;
    height: 16px;
    padding: 0 3px;
    margin: 0 1px;
    font-size: 10px;
    font-weight: 700;
    line-height: 1;
    color: #3b82f6;
    background: #eff6ff;
    border: 1px solid #bfdbfe;
    border-radius: 8px;
    cursor: pointer;
    vertical-align: super;
    transition: all 0.15s ease;
    user-select: none;
  }
  .msg-markdown :deep(.citation-ref:hover) {
    color: #fff;
    background: #3b82f6;
    border-color: #3b82f6;
    transform: scale(1.15);
  }
  .dark .msg-markdown :deep(.citation-ref) {
    color: #93c5fd;
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.3);
  }
  .dark .msg-markdown :deep(.citation-ref:hover) {
    color: #fff;
    background: #3b82f6;
    border-color: #3b82f6;
  }
</style>

<style>
  /* el-dialog 挂到 body，需非 scoped；小屏不超出视口 */
  .conv-query-data-dialog.el-dialog {
    width: min(720px, 92vw) !important;
    max-width: 92vw;
  }
</style>

<!-- 浮动引用面板（全局样式，不受 scoped 限制） -->
<style>
  .citation-popover {
    position: fixed;
    z-index: 9999;
    max-width: 380px;
    min-width: 260px;
    padding: 14px 16px;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12), 0 4px 12px rgba(0, 0, 0, 0.06);
    pointer-events: auto;
    animation: citation-pop-in 0.15s ease-out;
  }
  .dark .citation-popover {
    background: #1e293b;
    border-color: #334155;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
  }
  @keyframes citation-pop-in {
    from { opacity: 0; transform: translateY(-6px) scale(0.97); }
    to { opacity: 1; transform: translateY(0) scale(1); }
  }

  /* 引用片段过长时纵向滚动，并显示滚动条（避免 macOS 上仅靠 overlay 不易察觉） */
  .citation-popover-content {
    max-height: 13rem;
    overflow-x: hidden;
    overflow-y: auto;
    word-break: break-word;
    scrollbar-width: thin;
    scrollbar-color: rgba(100, 116, 139, 0.5) rgba(0, 0, 0, 0.06);
  }
  .citation-popover-content::-webkit-scrollbar {
    width: 8px;
  }
  .citation-popover-content::-webkit-scrollbar-track {
    background: rgba(0, 0, 0, 0.06);
    border-radius: 4px;
    margin: 2px 0;
  }
  .citation-popover-content::-webkit-scrollbar-thumb {
    background: rgba(100, 116, 139, 0.45);
    border-radius: 4px;
  }
  .citation-popover-content::-webkit-scrollbar-thumb:hover {
    background: rgba(100, 116, 139, 0.65);
  }
  .dark .citation-popover-content {
    scrollbar-color: rgba(148, 163, 184, 0.45) rgba(255, 255, 255, 0.08);
  }
  .dark .citation-popover-content::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.08);
  }
  .dark .citation-popover-content::-webkit-scrollbar-thumb {
    background: rgba(148, 163, 184, 0.4);
  }
  .dark .citation-popover-content::-webkit-scrollbar-thumb:hover {
    background: rgba(148, 163, 184, 0.55);
  }
</style>
