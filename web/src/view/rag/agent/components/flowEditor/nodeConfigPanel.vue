<template>
  <div v-if="node" class="node-config-panel">
    <div class="panel-header">
      <span>{{ $t('rag.flowEditor.configTitle', { name: node.data?.label || node.id }) }}</span>
      <el-button link icon="Close" size="small" @click="$emit('close')" />
    </div>
    <div class="panel-body">
      <p class="panel-ref-tip">{{ $t('rag.flowEditor.panel.refPanelTip') }}</p>
      <el-form label-width="90px" size="small">
        <!-- Begin -->
        <template v-if="node.data?.componentName === 'Begin'">
          <el-form-item :label="$t('rag.flowEditor.panel.prologue')">
            <el-input :model-value="params.prologue" type="textarea" :rows="2" :placeholder="$t('rag.flowEditor.panel.phPrologue')" @update:model-value="(v) => updateParam('prologue', v)" />
          </el-form-item>
        </template>

        <!-- Retrieval -->
        <template v-else-if="node.data?.componentName === 'Retrieval'">
          <el-form-item :label="$t('rag.flowEditor.panel.queryVar')">
            <FlowParamInput
              :model-value="params.query"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputSysQuery')"

              @update:model-value="(v) => updateParam('query', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.topN')">
            <el-input-number :model-value="params.top_n" :min="1" :max="20" @update:model-value="(v) => updateParam('top_n', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.emptyResultHint')">
            <FlowParamInput
              :model-value="params.empty_response"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phEmptyResponse')"

              @update:model-value="(v) => updateParam('empty_response', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.kb')">
            <el-select :model-value="params.kb_ids || []" multiple :placeholder="$t('rag.flowEditor.panel.phSelectKb')" style="width: 100%" @update:model-value="(v) => updateParam('kb_ids', v)">
              <el-option v-for="kb in knowledgeBases" :key="kb.ID" :label="kb.name" :value="String(kb.ID)" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.tocEnhance')">
            <el-select :model-value="retrievalTocEnhanceSelect" clearable style="width: 100%" @update:model-value="updateTocEnhanceSelect">
              <el-option :label="$t('rag.flowEditor.panel.tocEnhanceDefault')" value="" />
              <el-option :label="$t('rag.flowEditor.panel.tocEnhanceOn')" value="on" />
              <el-option :label="$t('rag.flowEditor.panel.tocEnhanceOff')" value="off" />
            </el-select>
            <div class="text-xs text-gray-400 mt-1">{{ $t('rag.flowEditor.panel.tocEnhanceHint') }}</div>
          </el-form-item>
        </template>

        <!-- LLM -->
        <template v-else-if="node.data?.componentName === 'LLM'">
          <el-form-item :label="$t('rag.flowEditor.panel.model')">
            <el-select :model-value="params.llm_id" :placeholder="$t('rag.flowEditor.panel.phSelectModel')" style="width: 100%" filterable @update:model-value="(v) => updateParam('llm_id', v)">
              <el-option v-for="m in llmModels" :key="m.id" :label="`${m.name} / ${m.modelName}`" :value="m.id" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.creativity')">
            <el-select :model-value="params.creativity || 'precise'" style="width: 100%" @update:model-value="(v) => updateParam('creativity', v)">
              <el-option :label="$t('rag.flowEditor.panel.creativityOptPrecise')" value="precise" />
              <el-option :label="$t('rag.flowEditor.panel.creativityOptBalance')" value="balance" />
              <el-option :label="$t('rag.flowEditor.panel.creativityOptImprovise')" value="improvise" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.temperature')">
            <el-input-number :model-value="params.temperature ?? 0.1" :min="0" :max="2" :step="0.1" @update:model-value="(v) => updateParam('temperature', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.sysPrompt')">
            <FlowParamInput
              :model-value="params.sys_prompt"
              input-type="textarea"
              :rows="4"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phSysPromptLlm')"

              @update:model-value="(v) => updateParam('sys_prompt', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.userPrompt')">
            <FlowParamInput
              :model-value="userPromptContent"
              input-type="textarea"
              :rows="2"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phUserPromptLlm')"

              @update:model-value="updateUserPrompt"
            />
          </el-form-item>
        </template>

        <!-- Agent -->
        <template v-else-if="node.data?.componentName === 'Agent'">
          <el-form-item :label="$t('rag.flowEditor.panel.model')">
            <el-select :model-value="params.llm_id" :placeholder="$t('rag.flowEditor.panel.phSelectModel')" style="width: 100%" filterable @update:model-value="(v) => updateParam('llm_id', v)">
              <el-option v-for="m in llmModels" :key="m.id" :label="`${m.name} / ${m.modelName}`" :value="m.id" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.creativity')">
            <el-select :model-value="params.creativity || 'balance'" style="width: 100%" @update:model-value="(v) => updateParam('creativity', v)">
              <el-option :label="$t('rag.flowEditor.panel.creativityOptPrecise')" value="precise" />
              <el-option :label="$t('rag.flowEditor.panel.creativityOptBalance')" value="balance" />
              <el-option :label="$t('rag.flowEditor.panel.creativityOptImprovise')" value="improvise" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.sysPrompt')">
            <FlowParamInput
              :model-value="params.sys_prompt"
              input-type="textarea"
              :rows="3"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phSysPromptAgent')"

              @update:model-value="(v) => updateParam('sys_prompt', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.userPrompt')">
            <FlowParamInput
              :model-value="params.user_prompt"
              input-type="textarea"
              :rows="2"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phUserPromptAgent')"

              @update:model-value="(v) => updateParam('user_prompt', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.maxRetries')">
            <el-input-number :model-value="params.max_retries ?? 1" :min="0" :max="5" @update:model-value="(v) => updateParam('max_retries', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.retryIntervalSec')">
            <el-input-number :model-value="params.delay_after_error ?? 1" :min="0" :max="60" @update:model-value="(v) => updateParam('delay_after_error', v)" />
          </el-form-item>
        </template>

        <!-- Message -->
        <template v-else-if="node.data?.componentName === 'Message'">
          <el-form-item :label="$t('rag.flowEditor.panel.outputContent')">
            <FlowParamInput
              :model-value="messageContentStr"
              input-type="textarea"
              :rows="3"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phOutputContent')"

              @update:model-value="updateMessageContent"
            />
          </el-form-item>
        </template>

        <!-- Switch -->
        <template v-else-if="node.data?.componentName === 'Switch'">
          <div class="switch-cases">
            <div v-for="(c, idx) in (params.cases || [])" :key="idx" class="switch-case-block">
              <div class="case-header">{{ $t('rag.flowEditor.panel.branchTitle', { n: idx + 1 }) }}</div>
              <el-form-item :label="$t('rag.flowEditor.panel.downstreamNode')">
                <el-select :model-value="c.downstream" :placeholder="$t('rag.flowEditor.panel.phDownstream')" style="width: 100%" clearable @update:model-value="(v) => updateCaseDownstream(idx, v)">
                  <el-option v-for="n in downstreamNodeOptions" :key="n.id" :label="n.data?.label || n.id" :value="n.id" />
                </el-select>
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.logic')">
                <el-radio-group :model-value="c.logic || 'AND'" size="small" @update:model-value="(v) => updateCaseField(idx, 'logic', v)">
                  <el-radio-button label="AND">{{ $t('rag.flowEditor.panel.logicAnd') }}</el-radio-button>
                  <el-radio-button label="OR">{{ $t('rag.flowEditor.panel.logicOr') }}</el-radio-button>
                </el-radio-group>
              </el-form-item>
              <div v-for="(cond, cidx) in (c.conditions || [])" :key="cidx" class="condition-row">
                <FlowParamInput
                  :model-value="cond.ref"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phVarRef')"

                  small
                  class="condition-row__ref"
                  @update:model-value="(v) => updateCondition(idx, cidx, 'ref', v)"
                />
                <el-select :model-value="cond.op" :placeholder="$t('rag.flowEditor.panel.phOperator')" size="small" style="flex: 1" @update:model-value="(v) => updateCondition(idx, cidx, 'op', v)">
                  <el-option :label="$t('rag.flowEditor.panel.opNotEmpty')" value="not_empty" />
                  <el-option :label="$t('rag.flowEditor.panel.opEmpty')" value="is_empty" />
                  <el-option :label="$t('rag.flowEditor.panel.opEquals')" value="equals" />
                  <el-option :label="$t('rag.flowEditor.panel.opNotEqual')" value="not_equal" />
                  <el-option :label="$t('rag.flowEditor.panel.opContains')" value="contains" />
                  <el-option :label="$t('rag.flowEditor.panel.opNotContains')" value="not_contains" />
                  <el-option :label="$t('rag.flowEditor.panel.opStarts')" value="starts_with" />
                  <el-option :label="$t('rag.flowEditor.panel.opEnds')" value="ends_with" />
                  <el-option :label="$t('rag.flowEditor.panel.opLt')" value="less_than" />
                  <el-option :label="$t('rag.flowEditor.panel.opLe')" value="less_equal" />
                  <el-option :label="$t('rag.flowEditor.panel.opGt')" value="greater_than" />
                  <el-option :label="$t('rag.flowEditor.panel.opGe')" value="greater_equal" />
                </el-select>
                <FlowParamInput
                  v-if="!['is_empty','not_empty'].includes(cond.op)"
                  :model-value="cond.value"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phValue')"

                  small
                  class="condition-row__val"
                  @update:model-value="(v) => updateCondition(idx, cidx, 'value', v)"
                />
              </div>
              <el-button type="primary" link size="small" @click="addCondition(idx)">{{ $t('rag.flowEditor.panel.addCondition') }}</el-button>
            </div>
          </div>
          <el-button type="primary" link size="small" @click="addCase">{{ $t('rag.flowEditor.panel.addBranch') }}</el-button>
        </template>

        <!-- Categorize -->
        <template v-else-if="node.data?.componentName === 'Categorize'">
          <el-form-item :label="$t('rag.flowEditor.panel.inputVar')">
            <FlowParamInput
              :model-value="params.input"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputSysQuery')"

              @update:model-value="(v) => updateParam('input', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.model')">
            <el-select :model-value="params.llm_id" :placeholder="$t('rag.flowEditor.panel.phSelectModel')" style="width: 100%" filterable @update:model-value="(v) => updateParam('llm_id', v)">
              <el-option v-for="m in llmModels" :key="m.id" :label="`${m.name} / ${m.modelName}`" :value="m.id" />
            </el-select>
          </el-form-item>
          <div class="categorize-categories">
            <div v-for="(cat, idx) in (params.categories || [])" :key="idx" class="category-block">
              <div class="case-header">{{ $t('rag.flowEditor.panel.categoryTitle', { n: idx + 1 }) }}</div>
              <el-form-item :label="$t('rag.flowEditor.panel.name')">
                <el-input :model-value="cat.name" :placeholder="$t('rag.flowEditor.panel.phNameQa')" @update:model-value="(v) => updateCategory(idx, 'name', v)" />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.labelDescription')">
                <el-input :model-value="cat.description" type="textarea" :rows="2" :placeholder="$t('rag.flowEditor.panel.phCatDesc')" @update:model-value="(v) => updateCategory(idx, 'description', v)" />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.examples')">
                <el-input :model-value="(cat.examples || []).join(', ')" :placeholder="$t('rag.flowEditor.panel.phExamples')" @update:model-value="(v) => updateCategoryExamples(idx, v)" />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.downstreamNode')">
                <el-select :model-value="cat.downstream" :placeholder="$t('rag.flowEditor.panel.phDownstream')" style="width: 100%" clearable @update:model-value="(v) => updateCategory(idx, 'downstream', v)">
                  <el-option v-for="n in downstreamNodeOptions" :key="n.id" :label="n.data?.label || n.id" :value="n.id" />
                </el-select>
              </el-form-item>
            </div>
          </div>
          <el-button type="primary" link size="small" @click="addCategory">{{ $t('rag.flowEditor.panel.addCategory') }}</el-button>
        </template>

        <!-- HTTPRequest -->
        <template v-else-if="node.data?.componentName === 'HTTPRequest'">
          <el-form-item :label="$t('rag.flowEditor.panel.url')">
            <FlowParamInput
              :model-value="params.url"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phUrlEx')"

              @update:model-value="(v) => updateParam('url', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.method')">
            <el-select :model-value="params.method || 'GET'" style="width: 100%" @update:model-value="(v) => updateParam('method', v)">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.timeoutSec')">
            <el-input-number :model-value="params.timeout || 60" :min="1" :max="300" @update:model-value="(v) => updateParam('timeout', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.proxy')">
            <el-input :model-value="params.proxy" :placeholder="$t('rag.flowEditor.panel.phProxy')" @update:model-value="(v) => updateParam('proxy', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.headersJson')">
            <FlowParamInput
              :model-value="httpHeadersStr"
              input-type="textarea"
              :rows="2"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phHeadersJson')"

              @update:model-value="updateHttpHeaders"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.paramsJson')">
            <FlowParamInput
              :model-value="httpParamsStr"
              input-type="textarea"
              :rows="2"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phParamsJson')"

              @update:model-value="updateHttpParams"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.stripHtml')">
            <el-switch :model-value="params.clean_html" @update:model-value="(v) => updateParam('clean_html', v)" />
          </el-form-item>
        </template>

        <!-- Iteration -->
        <template v-else-if="node.data?.componentName === 'Iteration'">
          <el-form-item :label="$t('rag.flowEditor.panel.inputVar')">
            <FlowParamInput
              :model-value="params.input"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputIter')"

              @update:model-value="(v) => updateParam('input', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.delimiter')">
            <el-select :model-value="params.delimiter || 'comma'" style="width: 100%" @update:model-value="(v) => updateParam('delimiter', v)">
              <el-option :label="$t('rag.flowEditor.panel.delimComma')" value="comma" />
              <el-option :label="$t('rag.flowEditor.panel.delimNewline')" value="newline" />
              <el-option :label="$t('rag.flowEditor.panel.delimSemicolon')" value="semicolon" />
              <el-option :label="$t('rag.flowEditor.panel.delimTab')" value="tab" />
              <el-option :label="$t('rag.flowEditor.panel.delimDash')" value="dash" />
              <el-option :label="$t('rag.flowEditor.panel.delimCustom')" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="params.delimiter === 'custom'" :label="$t('rag.flowEditor.panel.customDelimiter')">
            <el-input :model-value="params.delimiter_text" :placeholder="$t('rag.flowEditor.panel.phDelimText')" @update:model-value="(v) => updateParam('delimiter_text', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.downstreamNode')">
            <el-select :model-value="params.downstream" :placeholder="$t('rag.flowEditor.panel.phDownstreamWire')" style="width: 100%" clearable @update:model-value="(v) => updateParam('downstream', v)">
              <el-option v-for="n in downstreamNodeOptions" :key="n.id" :label="n.data?.label || n.id" :value="n.id" />
            </el-select>
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintIteration') }}</div>
        </template>

        <!-- DocsGenerator -->
        <template v-else-if="node.data?.componentName === 'DocsGenerator'">
          <el-form-item :label="$t('rag.flowEditor.panel.content')">
            <FlowParamInput
              :model-value="params.content"
              input-type="textarea"
              :rows="4"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phMdContent')"

              @update:model-value="(v) => updateParam('content', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.docTitle')">
            <FlowParamInput
              :model-value="params.title"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phDocTitle')"

              @update:model-value="(v) => updateParam('title', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.outputFormat')">
            <el-select :model-value="params.output_format || 'pdf'" style="width: 100%" @update:model-value="(v) => updateParam('output_format', v)">
              <el-option label="PDF" value="pdf" />
              <el-option label="DOCX" value="docx" />
              <el-option label="TXT" value="txt" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.outputDir')">
            <el-input :model-value="params.output_dir" :placeholder="$t('rag.flowEditor.panel.phOutputDir')" @update:model-value="(v) => updateParam('output_dir', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.fileName')">
            <el-input :model-value="params.filename" :placeholder="$t('rag.flowEditor.panel.phFileNameAuto')" @update:model-value="(v) => updateParam('filename', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintDocsOut') }}</div>
        </template>

        <!-- ExecuteSQL -->
        <template v-else-if="node.data?.componentName === 'ExecuteSQL'">
          <el-form-item :label="$t('rag.flowEditor.panel.sql')">
            <FlowParamInput
              :model-value="params.sql"
              input-type="textarea"
              :rows="3"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phSql')"

              @update:model-value="(v) => updateParam('sql', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.dbType')">
            <el-select :model-value="params.db_type || 'mysql'" style="width: 100%" @update:model-value="(v) => updateParam('db_type', v)">
              <el-option label="MySQL" value="mysql" />
              <el-option label="MariaDB" value="mariadb" />
              <el-option label="PostgreSQL" value="postgresql" />
              <el-option label="SQL Server" value="mssql" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.host')">
            <el-input :model-value="params.host" :placeholder="$t('rag.flowEditor.panel.phHost')" @update:model-value="(v) => updateParam('host', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.port')">
            <el-input-number :model-value="params.port || 3306" :min="1" :max="65535" @update:model-value="(v) => updateParam('port', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.username')">
            <el-input :model-value="params.username" :placeholder="$t('rag.flowEditor.panel.phUsername')" @update:model-value="(v) => updateParam('username', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.password')">
            <el-input :model-value="params.password" type="password" :placeholder="$t('rag.flowEditor.panel.phDbPassword')" show-password @update:model-value="(v) => updateParam('password', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.database')">
            <el-input :model-value="params.database" :placeholder="$t('rag.flowEditor.panel.phDatabase')" @update:model-value="(v) => updateParam('database', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.maxRows')">
            <el-input-number :model-value="params.max_records ?? 1024" :min="1" :max="10000" @update:model-value="(v) => updateParam('max_records', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintSqlOut') }}</div>
        </template>

        <!-- MCP -->
        <template v-else-if="node.data?.componentName === 'MCP'">
          <el-form-item :label="$t('rag.flowEditor.panel.mcpServerUrl')">
            <el-input :model-value="params.server_url" :placeholder="$t('rag.flowEditor.panel.phMcpUrl')" @update:model-value="(v) => updateParam('server_url', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.serverName')">
            <el-input :model-value="params.server_name" :placeholder="$t('rag.flowEditor.panel.phServerName')" @update:model-value="(v) => updateParam('server_name', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.toolName')">
            <el-input :model-value="params.tool_name" :placeholder="$t('rag.flowEditor.panel.phToolExamples')" @update:model-value="(v) => updateParam('tool_name', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.argsJson')">
            <FlowParamInput
              :model-value="mcpArgumentsStr"
              input-type="textarea"
              :rows="3"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phMcpArgsJson')"

              @update:model-value="updateMcpArguments"
            />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintMcpOut') }}</div>
        </template>

        <!-- SetVariable -->
        <template v-else-if="node.data?.componentName === 'SetVariable'">
          <div class="categorize-categories">
            <div v-for="(row, idx) in (params.assignments || [])" :key="idx" class="category-block">
              <div class="case-header">{{ $t('rag.flowEditor.panel.assignTitle', { n: idx + 1 }) }}</div>
              <el-form-item :label="$t('rag.flowEditor.panel.globalKey')">
                <FlowParamInput
                  :model-value="row.key"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phSysSummary')"

                  @update:model-value="(v) => updateAssignment(idx, 'key', v)"
                />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.valueTemplate')">
                <FlowParamInput
                  :model-value="row.value"
                  input-type="textarea"
                  :rows="2"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phValueTpl')"

                  @update:model-value="(v) => updateAssignment(idx, 'value', v)"
                />
              </el-form-item>
            </div>
          </div>
          <el-button type="primary" link size="small" @click="addAssignment">{{ $t('rag.flowEditor.panel.addAssignment') }}</el-button>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintSetVar') }}</div>
        </template>

        <!-- VariableAssigner -->
        <template v-else-if="node.data?.componentName === 'VariableAssigner'">
          <div class="categorize-categories">
            <div v-for="(row, idx) in (params.variables || [])" :key="idx" class="category-block">
              <div class="case-header">{{ $t('rag.flowEditor.panel.ruleTitle', { n: idx + 1 }) }}</div>
              <el-form-item :label="$t('rag.flowEditor.panel.targetGlobalKey')">
                <FlowParamInput
                  :model-value="row.variable"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phSysItems')"

                  @update:model-value="(v) => updateVaItem(idx, 'variable', v)"
                />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.operator')">
                <el-select :model-value="row.operator || 'overwrite'" style="width: 100%" @update:model-value="(v) => updateVaItem(idx, 'operator', v)">
                  <el-option :label="$t('rag.flowEditor.panel.vaOpOverwrite')" value="overwrite" />
                  <el-option :label="$t('rag.flowEditor.panel.vaOpClear')" value="clear" />
                  <el-option :label="$t('rag.flowEditor.panel.vaOpSet')" value="set" />
                  <el-option :label="$t('rag.flowEditor.panel.vaOpAppend')" value="append" />
                  <el-option :label="$t('rag.flowEditor.panel.vaOpExtend')" value="extend" />
                  <el-option label="remove_first" value="remove_first" />
                  <el-option label="remove_last" value="remove_last" />
                  <el-option label="+=" :value="'+='" />
                  <el-option label="-=" :value="'-='" />
                  <el-option label="*=" :value="'*='" />
                  <el-option label="/=" :value="'/='" />
                </el-select>
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.paramRefTpl')">
                <FlowParamInput
                  :model-value="row.parameter"
                  input-type="textarea"
                  :rows="2"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phVaParam')"

                  @update:model-value="(v) => updateVaItem(idx, 'parameter', v)"
                />
              </el-form-item>
            </div>
          </div>
          <el-button type="primary" link size="small" @click="addVaItem">{{ $t('rag.flowEditor.panel.addRule') }}</el-button>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintVa') }}</div>
        </template>

        <!-- VariableAggregator -->
        <template v-else-if="node.data?.componentName === 'VariableAggregator'">
          <div class="categorize-categories">
            <div v-for="(g, gidx) in (params.groups || [])" :key="gidx" class="category-block">
              <div class="case-header">{{ $t('rag.flowEditor.panel.groupTitle', { n: gidx + 1 }) }}</div>
              <el-form-item :label="$t('rag.flowEditor.panel.outputGroupName')">
                <FlowParamInput
                  :model-value="g.group_name"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phGroupRef')"

                  @update:model-value="(v) => updateAggGroupField(gidx, 'group_name', v)"
                />
              </el-form-item>
              <div v-for="(sel, vidx) in (g.variables || [])" :key="vidx" class="condition-row">
                <FlowParamInput
                  :model-value="sel.value"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phVarPick')"

                  class="condition-row__agg"
                  @update:model-value="(v) => updateAggSelector(gidx, vidx, v)"
                />
              </div>
              <el-button type="primary" link size="small" @click="addAggSelector(gidx)">{{ $t('rag.flowEditor.panel.addCandidateVar') }}</el-button>
            </div>
          </div>
          <el-button type="primary" link size="small" @click="addAggGroup">{{ $t('rag.flowEditor.panel.addGroup') }}</el-button>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintAgg') }}</div>
        </template>

        <!-- ListOperations -->
        <template v-else-if="node.data?.componentName === 'ListOperations'">
          <el-form-item :label="$t('rag.flowEditor.panel.inputVar')">
            <FlowParamInput
              :model-value="params.input"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phListUpstream')"

              @update:model-value="(v) => updateParam('input', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.orJsonLiteral')">
            <FlowParamInput
              :model-value="params.input_literal"
              input-type="textarea"
              :rows="2"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phJsonLiteralArr')"

              @update:model-value="(v) => updateParam('input_literal', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.operation')">
            <el-select :model-value="params.operation || 'topn'" style="width: 100%" @update:model-value="(v) => updateParam('operation', v)">
              <el-option label="topN" value="topn" />
              <el-option label="head" value="head" />
              <el-option label="tail" value="tail" />
              <el-option label="filter" value="filter" />
              <el-option label="sort" value="sort" />
              <el-option label="drop_duplicates" value="drop_duplicates" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="['topn','head','tail'].includes(params.operation || 'topn')" :label="$t('rag.flowEditor.panel.nLabel')">
            <el-input-number :model-value="params.n ?? 10" :min="0" :max="9999" @update:model-value="(v) => updateParam('n', v)" />
          </el-form-item>
          <template v-if="(params.operation || '') === 'filter'">
            <el-form-item :label="$t('rag.flowEditor.panel.fieldField')">
              <FlowParamInput
                :model-value="params.field"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phFieldObj')"

                @update:model-value="(v) => updateParam('field', v)"
              />
            </el-form-item>
            <el-form-item :label="$t('rag.flowEditor.panel.compareOp')">
              <el-select :model-value="params.filter_operator || '='" style="width: 100%" @update:model-value="(v) => updateParam('filter_operator', v)">
                <el-option :label="$t('rag.flowEditor.panel.cmpEq')" value="=" />
                <el-option :label="$t('rag.flowEditor.panel.cmpNe')" value="!=" />
                <el-option :label="$t('rag.flowEditor.panel.cmpContains')" value="contains" />
                <el-option :label="$t('rag.flowEditor.panel.cmpStarts')" value="startswith" />
                <el-option :label="$t('rag.flowEditor.panel.cmpEnds')" value="endswith" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('rag.flowEditor.panel.compareValue')">
              <FlowParamInput
                :model-value="params.value"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"

                @update:model-value="(v) => updateParam('value', v)"
              />
            </el-form-item>
          </template>
          <template v-if="(params.operation || '') === 'sort'">
            <el-form-item :label="$t('rag.flowEditor.panel.sortBy')">
              <el-select :model-value="params.sort_by || 'letter'" style="width: 100%" @update:model-value="(v) => updateParam('sort_by', v)">
                <el-option :label="$t('rag.flowEditor.panel.sortLetter')" value="letter" />
                <el-option :label="$t('rag.flowEditor.panel.sortNumeric')" value="numeric" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('rag.flowEditor.panel.order')">
              <el-select :model-value="params.sort_order || 'asc'" style="width: 100%" @update:model-value="(v) => updateParam('sort_order', v)">
                <el-option :label="$t('rag.flowEditor.panel.orderAsc')" value="asc" />
                <el-option :label="$t('rag.flowEditor.panel.orderDesc')" value="desc" />
              </el-select>
            </el-form-item>
          </template>
          <el-form-item v-if="(params.operation || '') === 'drop_duplicates'" :label="$t('rag.flowEditor.panel.dedupeKey')">
            <FlowParamInput
              :model-value="params.dedupe_key"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phDedupe')"

              @update:model-value="(v) => updateParam('dedupe_key', v)"
            />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintListOut') }}</div>
        </template>

        <!-- StringTransform -->
        <template v-else-if="node.data?.componentName === 'StringTransform'">
          <el-form-item :label="$t('rag.flowEditor.panel.mode')">
            <el-radio-group :model-value="params.mode || 'split'" @update:model-value="(v) => updateParam('mode', v)">
              <el-radio-button label="split">{{ $t('rag.flowEditor.panel.split') }}</el-radio-button>
              <el-radio-button label="merge">{{ $t('rag.flowEditor.panel.merge') }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <template v-if="(params.mode || 'split') === 'split'">
            <el-form-item :label="$t('rag.flowEditor.panel.inputVar')">
              <FlowParamInput
                :model-value="params.input"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phEgSysQuery')"

                @update:model-value="(v) => updateParam('input', v)"
              />
            </el-form-item>
            <el-form-item :label="$t('rag.flowEditor.panel.directText')">
              <FlowParamInput
                :model-value="params.input_literal"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"

                @update:model-value="(v) => updateParam('input_literal', v)"
              />
            </el-form-item>
            <el-form-item :label="$t('rag.flowEditor.panel.delimJsonArray')">
              <FlowParamInput
                :model-value="stringDelimitersStr"
                input-type="textarea"
                :rows="2"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phDelimJson')"

                @update:model-value="updateStringDelimiters"
              />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item :label="$t('rag.flowEditor.panel.template')">
              <FlowParamInput
                :model-value="params.template"
                input-type="textarea"
                :rows="2"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phTitleBodyTpl')"

                @update:model-value="(v) => updateParam('template', v)"
              />
            </el-form-item>
            <el-form-item :label="$t('rag.flowEditor.panel.placeholderVarJson')">
              <FlowParamInput
                :model-value="stringMergeVarsStr"
                input-type="textarea"
                :rows="4"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phPlcJson')"

                @update:model-value="updateStringMergeVars"
              />
            </el-form-item>
          </template>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintStrOut') }}</div>
        </template>

        <!-- Invoke -->
        <template v-else-if="node.data?.componentName === 'Invoke'">
          <el-form-item :label="$t('rag.flowEditor.panel.url')">
            <FlowParamInput
              :model-value="params.url"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phUrlEx')"

              @update:model-value="(v) => updateParam('url', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.method')">
            <el-select :model-value="(params.method || 'GET').toUpperCase()" style="width: 100%" @update:model-value="(v) => updateParam('method', v)">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.bodyType')">
            <el-select :model-value="params.datatype || 'json'" style="width: 100%" @update:model-value="(v) => updateParam('datatype', v)">
              <el-option :label="$t('rag.flowEditor.panel.bodyOptJson')" value="json" />
              <el-option :label="$t('rag.flowEditor.panel.bodyOptFormdata')" value="formdata" />
              <el-option :label="$t('rag.flowEditor.panel.bodyOptUrlencoded')" value="urlencoded" />
            </el-select>
          </el-form-item>
          <div class="categorize-categories">
            <div v-for="(row, idx) in (params.variables || [])" :key="idx" class="category-block">
              <div class="case-header">{{ $t('rag.flowEditor.panel.invokeArgTitle', { n: idx + 1 }) }}</div>
              <el-form-item :label="$t('rag.flowEditor.panel.key')">
                <el-input :model-value="row.key" @update:model-value="(v) => updateInvokeVar(idx, 'key', v)" />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.refVar')">
                <FlowParamInput
                  :model-value="row.ref"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"
                  :placeholder="$t('rag.flowEditor.panel.phRefLlm')"

                  @update:model-value="(v) => updateInvokeVar(idx, 'ref', v)"
                />
              </el-form-item>
              <el-form-item :label="$t('rag.flowEditor.panel.orValueStr')">
                <FlowParamInput
                  :model-value="row.value_str"
                  :nodes="nodes"
                  :edges="edges"
                  :current-node-id="node.id"

                  @update:model-value="(v) => updateInvokeVar(idx, 'value_str', v)"
                />
              </el-form-item>
            </div>
          </div>
          <el-button type="primary" link size="small" @click="addInvokeVar">{{ $t('rag.flowEditor.panel.addArg') }}</el-button>
          <el-form-item :label="$t('rag.flowEditor.panel.headersJson')">
            <FlowParamInput
              :model-value="invokeHeadersStr"
              input-type="textarea"
              :rows="3"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"

              @update:model-value="updateInvokeHeaders"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.timeoutSec')">
            <el-input-number :model-value="params.timeout ?? 30" :min="5" :max="120" @update:model-value="(v) => updateParam('timeout', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintInvoke') }}</div>
        </template>

        <!-- Transformer -->
        <template v-else-if="node.data?.componentName === 'Transformer'">
          <el-form-item :label="$t('rag.flowEditor.panel.inputVar')">
            <FlowParamInput
              :model-value="params.input"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputTf')"

              @update:model-value="(v) => updateParam('input', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.model')">
            <el-select :model-value="params.llm_id" :placeholder="$t('rag.flowEditor.panel.phSelectModel')" style="width: 100%" filterable @update:model-value="(v) => updateParam('llm_id', v)">
              <el-option v-for="m in llmModels" :key="m.id" :label="`${m.name} / ${m.modelName}`" :value="m.id" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.creativity')">
            <el-select :model-value="params.creativity || 'balance'" style="width: 100%" @update:model-value="(v) => updateParam('creativity', v)">
              <el-option :label="$t('rag.flowEditor.panel.creativityOptPrecise')" value="precise" />
              <el-option :label="$t('rag.flowEditor.panel.creativityOptBalance')" value="balance" />
              <el-option :label="$t('rag.flowEditor.panel.creativityOptImprovise')" value="improvise" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.temperature')">
            <el-input-number :model-value="params.temperature ?? 0.2" :min="0" :max="2" :step="0.1" @update:model-value="(v) => updateParam('temperature', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.sysInstruction')">
            <FlowParamInput
              :model-value="params.instruction"
              input-type="textarea"
              :rows="4"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInstruct')"

              @update:model-value="(v) => updateParam('instruction', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.maxTokens')">
            <el-input-number :model-value="params.max_tokens || 0" :min="0" :max="8192" :placeholder="$t('rag.flowEditor.panel.phMaxTokens0')" @update:model-value="(v) => updateParam('max_tokens', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintTfOut') }}</div>
        </template>

        <!-- AwaitResponse -->
        <template v-else-if="node.data?.componentName === 'AwaitResponse'">
          <el-form-item :label="$t('rag.flowEditor.panel.promptCopy')">
            <FlowParamInput
              :model-value="params.message"
              input-type="textarea"
              :rows="3"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phUserFacing')"

              @update:model-value="(v) => updateParam('message', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.readGlobalKey')">
            <FlowParamInput
              :model-value="params.variable_key || 'sys.await_reply'"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              placeholder="sys.await_reply"

              @update:model-value="(v) => updateParam('variable_key', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.requireNonEmpty')">
            <el-switch :model-value="params.require_non_empty !== false" @update:model-value="(v) => updateParam('require_non_empty', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintAwait') }}</div>
        </template>

        <!-- DuckDuckGo -->
        <template v-else-if="node.data?.componentName === 'DuckDuckGo'">
          <el-form-item :label="$t('rag.flowEditor.panel.query')">
            <FlowParamInput
              :model-value="params.query"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phQueryBrace')"

              @update:model-value="(v) => updateParam('query', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.channel')">
            <el-select :model-value="params.channel || 'general'" style="width: 100%" @update:model-value="(v) => updateParam('channel', v)">
              <el-option :label="$t('rag.flowEditor.panel.chGeneral')" value="general" />
              <el-option :label="$t('rag.flowEditor.panel.chNews')" value="news" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.topN')">
            <el-input-number :model-value="params.top_n ?? 10" :min="1" :max="25" @update:model-value="(v) => updateParam('top_n', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.timeoutSec')">
            <el-input-number :model-value="params.timeout ?? 15" :min="5" :max="60" @update:model-value="(v) => updateParam('timeout', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintDdg') }}</div>
        </template>

        <!-- Wikipedia -->
        <template v-else-if="node.data?.componentName === 'Wikipedia'">
          <el-form-item :label="$t('rag.flowEditor.panel.query')">
            <FlowParamInput
              :model-value="params.query"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputSysQuery')"

              @update:model-value="(v) => updateParam('query', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.langCode')">
            <el-input :model-value="params.language || 'zh'" :placeholder="$t('rag.flowEditor.panel.phLang')" @update:model-value="(v) => updateParam('language', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.topN')">
            <el-input-number :model-value="params.top_n ?? 5" :min="1" :max="15" @update:model-value="(v) => updateParam('top_n', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.timeoutSec')">
            <el-input-number :model-value="params.timeout ?? 30" :min="5" :max="90" @update:model-value="(v) => updateParam('timeout', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintWiki') }}</div>
        </template>

        <!-- ArXiv -->
        <template v-else-if="node.data?.componentName === 'ArXiv'">
          <el-form-item :label="$t('rag.flowEditor.panel.query')">
            <FlowParamInput
              :model-value="params.query"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputSysQuery')"

              @update:model-value="(v) => updateParam('query', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.topN')">
            <el-input-number :model-value="params.top_n ?? 10" :min="1" :max="50" @update:model-value="(v) => updateParam('top_n', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.sortByArxiv')">
            <el-select :model-value="params.sort_by || 'submittedDate'" style="width: 100%" @update:model-value="(v) => updateParam('sort_by', v)">
              <el-option :label="$t('rag.flowEditor.panel.sortSubmitted')" value="submittedDate" />
              <el-option :label="$t('rag.flowEditor.panel.sortUpdated')" value="lastUpdatedDate" />
              <el-option :label="$t('rag.flowEditor.panel.sortRelevance')" value="relevance" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.timeoutSec')">
            <el-input-number :model-value="params.timeout ?? 20" :min="5" :max="90" @update:model-value="(v) => updateParam('timeout', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintArxiv') }}</div>
        </template>

        <!-- TavilySearch -->
        <template v-else-if="node.data?.componentName === 'TavilySearch'">
          <el-form-item :label="$t('rag.flowEditor.panel.query')">
            <FlowParamInput
              :model-value="params.query"
              :nodes="nodes"
              :edges="edges"
              :current-node-id="node.id"
              :placeholder="$t('rag.flowEditor.panel.phInputSysQuery')"

              @update:model-value="(v) => updateParam('query', v)"
            />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.apiKey')">
            <el-input :model-value="params.api_key" type="password" :placeholder="$t('rag.flowEditor.panel.phTavilyKey')" show-password autocomplete="off" @update:model-value="(v) => updateParam('api_key', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.topic')">
            <el-select :model-value="params.topic || 'general'" style="width: 100%" @update:model-value="(v) => updateParam('topic', v)">
              <el-option :label="$t('rag.flowEditor.panel.topicGeneral')" value="general" />
              <el-option :label="$t('rag.flowEditor.panel.topicNews')" value="news" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.depth')">
            <el-select :model-value="params.search_depth || 'basic'" style="width: 100%" @update:model-value="(v) => updateParam('search_depth', v)">
              <el-option :label="$t('rag.flowEditor.panel.depthBasic')" value="basic" />
              <el-option :label="$t('rag.flowEditor.panel.depthAdvanced')" value="advanced" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.maxResults')">
            <el-input-number :model-value="params.max_results ?? 6" :min="1" :max="20" @update:model-value="(v) => updateParam('max_results', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.includeAnswer')">
            <el-switch :model-value="params.include_answer" @update:model-value="(v) => updateParam('include_answer', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.domainsInclude')">
            <el-input :model-value="params.include_domains_csv" :placeholder="$t('rag.flowEditor.panel.phDomainsCsv')" @update:model-value="(v) => updateParam('include_domains_csv', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.domainsExclude')">
            <el-input :model-value="params.exclude_domains_csv" :placeholder="$t('rag.flowEditor.panel.phCommaSep')" @update:model-value="(v) => updateParam('exclude_domains_csv', v)" />
          </el-form-item>
          <el-form-item :label="$t('rag.flowEditor.panel.timeoutSec')">
            <el-input-number :model-value="params.timeout ?? 30" :min="5" :max="120" @update:model-value="(v) => updateParam('timeout', v)" />
          </el-form-item>
          <div class="text-gray-500 text-xs mt-1">{{ $t('rag.flowEditor.panel.hintTavily') }}</div>
        </template>

        <!-- TextProcessing -->
        <template v-else-if="node.data?.componentName === 'TextProcessing'">
          <el-form-item :label="$t('rag.flowEditor.panel.textMode')">
            <el-radio-group :model-value="params.method || 'split'" @update:model-value="(v) => updateParam('method', v)">
              <el-radio-button label="split">{{ $t('rag.flowEditor.panel.split') }}</el-radio-button>
              <el-radio-button label="merge">{{ $t('rag.flowEditor.panel.merge') }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <template v-if="(params.method || 'split') === 'split'">
            <el-form-item :label="$t('rag.flowEditor.panel.inputVar')">
              <FlowParamInput
                :model-value="params.split_ref"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phSplitRef')"

                @update:model-value="(v) => updateParam('split_ref', v)"
              />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item :label="$t('rag.flowEditor.panel.template')">
              <FlowParamInput
                :model-value="params.script"
                input-type="textarea"
                :rows="3"
                :nodes="nodes"
                :edges="edges"
                :current-node-id="node.id"
                :placeholder="$t('rag.flowEditor.panel.phScriptEx')"

                @update:model-value="(v) => updateParam('script', v)"
              />
            </el-form-item>
          </template>
          <el-form-item :label="$t('rag.flowEditor.panel.delimTp')">
            <el-select :model-value="params.delimiter || 'newline'" style="width: 100%" @update:model-value="(v) => updateParam('delimiter', v)">
              <el-option :label="$t('rag.flowEditor.panel.delimNewline')" value="newline" />
              <el-option :label="$t('rag.flowEditor.panel.delimComma')" value="comma" />
              <el-option :label="$t('rag.flowEditor.panel.delimSemicolon')" value="semicolon" />
              <el-option :label="$t('rag.flowEditor.panel.delimTab')" value="tab" />
              <el-option :label="$t('rag.flowEditor.panel.delimSpace')" value="space" />
              <el-option :label="$t('rag.flowEditor.panel.delimCustom')" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="params.delimiter === 'custom'" :label="$t('rag.flowEditor.panel.customDelimiter')">
            <el-input :model-value="params.delimiter_text" :placeholder="$t('rag.flowEditor.panel.phDelimText')" @update:model-value="(v) => updateParam('delimiter_text', v)" />
          </el-form-item>
        </template>

        <template v-else>
          <div class="text-gray-500 text-sm">{{ $t('rag.flowEditor.panel.notSupported') }}</div>
        </template>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import FlowParamInput from './FlowParamInput.vue'

const props = defineProps({
  node: Object,
  nodes: { type: Array, default: () => [] },
  edges: { type: Array, default: () => [] },
  knowledgeBases: { type: Array, default: () => [] },
  llmModels: { type: Array, default: () => [] }
})

const emit = defineEmits(['update', 'close'])

const params = computed(() => ({ ...(props.node?.data?.params || {}) }))

/** Retrieval：toc_enhance / tocEnhance → 下拉 '' | on | off */
const retrievalTocEnhanceSelect = computed(() => {
  const p = params.value
  const v = p.toc_enhance !== undefined ? p.toc_enhance : p.tocEnhance
  if (v === true) return 'on'
  if (v === false) return 'off'
  return ''
})

const updateTocEnhanceSelect = (v) => {
  const next = { ...params.value }
  delete next.toc_enhance
  delete next.tocEnhance
  if (v === 'on') next.toc_enhance = true
  else if (v === 'off') next.toc_enhance = false
  emit('update', next)
}

const userPromptContent = computed(() => {
  const prompts = params.value.prompts || []
  const user = prompts.find((p) => p.role === 'user')
  return user?.content || ''
})

const messageContentStr = computed(() => {
  const c = params.value.content
  return Array.isArray(c) ? c.join('\n') : (c || '')
})

const httpHeadersStr = computed(() => {
  const h = params.value.headers
  if (!h || typeof h !== 'object') return ''
  try {
    return JSON.stringify(h, null, 2)
  } catch {
    return ''
  }
})

const httpParamsStr = computed(() => {
  const p = params.value.params
  if (!p || typeof p !== 'object') return ''
  try {
    return JSON.stringify(p, null, 2)
  } catch {
    return ''
  }
})

const mcpArgumentsStr = computed(() => {
  const a = params.value.arguments
  if (!a || typeof a !== 'object') return ''
  try {
    return JSON.stringify(a, null, 2)
  } catch {
    return ''
  }
})

const stringDelimitersStr = computed(() => {
  const d = params.value.delimiters
  if (!d) return '[]'
  try {
    return JSON.stringify(Array.isArray(d) ? d : [d], null, 2)
  } catch {
    return '[]'
  }
})

const stringMergeVarsStr = computed(() => {
  const m = params.value.merge_variables
  if (!m || typeof m !== 'object') return '{}'
  try {
    return JSON.stringify(m, null, 2)
  } catch {
    return '{}'
  }
})

const invokeHeadersStr = computed(() => {
  const h = params.value.headers
  if (!h || typeof h !== 'object') return ''
  try {
    return JSON.stringify(h, null, 2)
  } catch {
    return ''
  }
})

// Switch 下游节点选项：当前节点通过 edges 连接到的目标节点
const downstreamNodeOptions = computed(() => {
  if (!props.node?.id || !props.edges?.length) return []
  const targets = props.edges.filter((e) => e.source === props.node.id).map((e) => e.target)
  return props.nodes.filter((n) => targets.includes(n.id))
})

const updateParam = (key, value) => {
  emit('update', { ...params.value, [key]: value })
}

const updateUserPrompt = (val) => {
  const prompts = [...(params.value.prompts || [])]
  const idx = prompts.findIndex((p) => p.role === 'user')
  if (idx >= 0) {
    prompts[idx] = { ...prompts[idx], content: val }
  } else {
    prompts.push({ role: 'user', content: val })
  }
  emit('update', { ...params.value, prompts })
}

const updateMessageContent = (val) => {
  const arr = val ? val.split('\n').map((s) => s.trim()).filter(Boolean) : []
  emit('update', { ...params.value, content: arr.length ? arr : [''] })
}

// Switch 分支配置
const updateCaseDownstream = (caseIdx, value) => {
  const cases = [...(params.value.cases || [])]
  if (!cases[caseIdx]) cases[caseIdx] = { conditions: [], logic: 'AND', downstream: '' }
  cases[caseIdx] = { ...cases[caseIdx], downstream: value }
  emit('update', { ...params.value, cases })
}

const updateCaseField = (caseIdx, field, value) => {
  const cases = [...(params.value.cases || [])]
  if (!cases[caseIdx]) cases[caseIdx] = { conditions: [], logic: 'AND', downstream: '' }
  cases[caseIdx] = { ...cases[caseIdx], [field]: value }
  emit('update', { ...params.value, cases })
}

const updateCondition = (caseIdx, condIdx, field, value) => {
  const cases = JSON.parse(JSON.stringify(params.value.cases || []))
  if (!cases[caseIdx]) cases[caseIdx] = { conditions: [], logic: 'AND', downstream: '' }
  if (!cases[caseIdx].conditions) cases[caseIdx].conditions = []
  if (!cases[caseIdx].conditions[condIdx]) cases[caseIdx].conditions[condIdx] = { ref: '', op: 'equals', value: '' }
  cases[caseIdx].conditions[condIdx][field] = value
  emit('update', { ...params.value, cases })
}

const addCondition = (caseIdx) => {
  const cases = JSON.parse(JSON.stringify(params.value.cases || []))
  if (!cases[caseIdx]) cases[caseIdx] = { conditions: [], logic: 'AND', downstream: '' }
  cases[caseIdx].conditions = cases[caseIdx].conditions || []
  cases[caseIdx].conditions.push({ ref: '', op: 'equals', value: '' })
  emit('update', { ...params.value, cases })
}

const addCase = () => {
  const cases = [...(params.value.cases || []), { conditions: [{ ref: '', op: 'equals', value: '' }], logic: 'AND', downstream: '' }]
  emit('update', { ...params.value, cases })
}

// Categorize 分类配置
const updateCategory = (idx, field, value) => {
  const categories = JSON.parse(JSON.stringify(params.value.categories || []))
  if (!categories[idx]) categories[idx] = { name: '', description: '', examples: [], downstream: '' }
  categories[idx][field] = value
  emit('update', { ...params.value, categories })
}

const updateCategoryExamples = (idx, val) => {
  const arr = val ? val.split(',').map((s) => s.trim()).filter(Boolean) : []
  updateCategory(idx, 'examples', arr)
}

const addCategory = () => {
  const categories = [...(params.value.categories || []), { name: '', description: '', examples: [], downstream: '' }]
  emit('update', { ...params.value, categories })
}

const updateAssignment = (idx, field, value) => {
  const list = JSON.parse(JSON.stringify(params.value.assignments || []))
  if (!list[idx]) list[idx] = { key: '', value: '' }
  list[idx][field] = value
  emit('update', { ...params.value, assignments: list })
}

const addAssignment = () => {
  const list = [...(params.value.assignments || []), { key: '', value: '' }]
  emit('update', { ...params.value, assignments: list })
}

const updateVaItem = (idx, field, value) => {
  const list = JSON.parse(JSON.stringify(params.value.variables || []))
  if (!list[idx]) list[idx] = { variable: '', operator: 'overwrite', parameter: '' }
  list[idx][field] = value
  emit('update', { ...params.value, variables: list })
}

const addVaItem = () => {
  const list = [...(params.value.variables || []), { variable: '', operator: 'overwrite', parameter: '' }]
  emit('update', { ...params.value, variables: list })
}

const updateAggGroupField = (gidx, field, value) => {
  const groups = JSON.parse(JSON.stringify(params.value.groups || []))
  if (!groups[gidx]) groups[gidx] = { group_name: '', variables: [{ value: '' }] }
  groups[gidx][field] = value
  emit('update', { ...params.value, groups })
}

const updateAggSelector = (gidx, vidx, value) => {
  const groups = JSON.parse(JSON.stringify(params.value.groups || []))
  if (!groups[gidx]) groups[gidx] = { group_name: '', variables: [] }
  const vars = groups[gidx].variables || []
  if (!vars[vidx]) vars[vidx] = { value: '' }
  vars[vidx].value = value
  groups[gidx].variables = vars
  emit('update', { ...params.value, groups })
}

const addAggSelector = (gidx) => {
  const groups = JSON.parse(JSON.stringify(params.value.groups || []))
  if (!groups[gidx]) groups[gidx] = { group_name: '', variables: [] }
  groups[gidx].variables = [...(groups[gidx].variables || []), { value: '' }]
  emit('update', { ...params.value, groups })
}

const addAggGroup = () => {
  const groups = [...(params.value.groups || []), { group_name: '', variables: [{ value: '' }] }]
  emit('update', { ...params.value, groups })
}

const updateStringDelimiters = (val) => {
  try {
    const parsed = val ? JSON.parse(val) : []
    const arr = Array.isArray(parsed) ? parsed.map(String) : [String(parsed)]
    emit('update', { ...params.value, delimiters: arr })
  } catch {
    // 忽略无效 JSON
  }
}

const updateStringMergeVars = (val) => {
  try {
    const o = val ? JSON.parse(val) : {}
    emit('update', { ...params.value, merge_variables: typeof o === 'object' && o !== null ? o : {} })
  } catch {
    // 忽略无效 JSON
  }
}

const updateInvokeHeaders = (val) => {
  try {
    const h = val ? JSON.parse(val) : {}
    emit('update', { ...params.value, headers: typeof h === 'object' && h !== null ? h : {} })
  } catch {
    // 忽略无效 JSON
  }
}

const updateInvokeVar = (idx, field, value) => {
  const list = JSON.parse(JSON.stringify(params.value.variables || []))
  if (!list[idx]) list[idx] = { key: '', ref: '', value_str: '' }
  list[idx][field] = value
  emit('update', { ...params.value, variables: list })
}

const addInvokeVar = () => {
  const list = [...(params.value.variables || []), { key: '', ref: '' }]
  emit('update', { ...params.value, variables: list })
}

const updateHttpHeaders = (val) => {
  try {
    const h = val ? JSON.parse(val) : {}
    emit('update', { ...params.value, headers: typeof h === 'object' ? h : {} })
  } catch {
    // 忽略无效 JSON
  }
}

const updateHttpParams = (val) => {
  try {
    const p = val ? JSON.parse(val) : {}
    emit('update', { ...params.value, params: typeof p === 'object' ? p : {} })
  } catch {
    // 忽略无效 JSON
  }
}

const updateMcpArguments = (val) => {
  try {
    const a = val ? JSON.parse(val) : {}
    emit('update', { ...params.value, arguments: typeof a === 'object' ? a : {} })
  } catch {
    // 忽略无效 JSON
  }
}
</script>

<style scoped>
.node-config-panel {
  width: 280px;
  height: 100%;
  background: var(--el-bg-color);
  border-left: 1px solid var(--el-border-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color);
  font-weight: 600;
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.switch-cases {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.switch-case-block {
  padding: 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
}

.case-header {
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
}

.categorize-categories {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.category-block {
  padding: 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
}

.condition-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: flex-start;
}

.condition-row__ref {
  flex: 2;
  min-width: 0;
}

.condition-row__val,
.condition-row__agg {
  flex: 1;
  min-width: 0;
}

.panel-ref-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.45;
  margin: 0 0 12px;
  padding: 8px 10px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
  border-left: 3px solid var(--el-color-primary);
}
</style>
