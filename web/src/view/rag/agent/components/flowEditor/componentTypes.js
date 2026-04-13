/**
 * Agent component type definitions (aligned with backend DSL component_name).
 * paletteGroup: sidebar grouping
 *   core — orchestration backbone (flow, generation, branching, variables, etc.)
 *   tool — external calls & extensions (HTTP, search, MCP, SQL, etc.)
 *
 * label: English fallback only; the UI uses `rag.flowEditor.comp.<componentName>` i18n keys.
 */
export const COMPONENT_TYPES = {
  Begin: {
    componentName: 'Begin',
    label: 'Start',
    icon: 'play-circle',
    color: '#22c55e',
    defaultParams: { prologue: '' },
    category: 'input',
    paletteGroup: 'core'
  },
  Retrieval: {
    componentName: 'Retrieval',
    label: 'Retrieval',
    icon: 'search',
    color: '#3b82f6',
    defaultParams: { query: '{sys.query}', top_n: 20, empty_response: '', kb_ids: [] },
    category: 'retrieval',
    paletteGroup: 'core'
  },
  LLM: {
    componentName: 'LLM',
    label: 'Generate',
    icon: 'chat-dot-round',
    color: '#8b5cf6',
    defaultParams: {
      llm_id: 'ollama@llama3.2',
      creativity: 'precise',
      temperature: 0.1,
      sys_prompt: 'Answer based on the knowledge base.\n\nKnowledge base:\n{retrieval_0@formalized_content}',
      prompts: [{ role: 'user', content: '{sys.query}' }]
    },
    category: 'generate',
    paletteGroup: 'core'
  },
  Agent: {
    componentName: 'Agent',
    label: 'Agent',
    icon: 'avatar',
    color: '#0ea5e9',
    defaultParams: {
      llm_id: 'ollama@llama3.2',
      creativity: 'balance',
      sys_prompt: 'You are a helpful assistant. Complete the task based on user requirements.',
      user_prompt: '{sys.query}',
      max_retries: 1,
      delay_after_error: 1
    },
    category: 'generate',
    paletteGroup: 'core'
  },
  Message: {
    componentName: 'Message',
    label: 'Output',
    icon: 'message',
    color: '#f59e0b',
    defaultParams: { content: ['{generate_0@content}'] },
    category: 'output',
    paletteGroup: 'core'
  },
  Switch: {
    componentName: 'Switch',
    label: 'Branch',
    icon: 'switch',
    color: '#06b6d4',
    defaultParams: {
      cases: [
        {
          conditions: [{ ref: '{retrieval_0@formalized_content}', op: 'not_empty', value: '' }],
          logic: 'AND',
          downstream: ''
        },
        {
          conditions: [{ ref: '{retrieval_0@formalized_content}', op: 'is_empty', value: '' }],
          logic: 'AND',
          downstream: ''
        }
      ]
    },
    category: 'control',
    paletteGroup: 'core'
  },
  Categorize: {
    componentName: 'Categorize',
    label: 'Intent classify',
    icon: 'collection-tag',
    color: '#ec4899',
    defaultParams: {
      input: '{sys.query}',
      llm_id: 'ollama@llama3.2',
      categories: [
        { name: 'qa', description: 'Knowledge Q&A', examples: ['What is X?'], downstream: '' },
        { name: 'chat', description: 'Casual chat', examples: ['Hello'], downstream: '' }
      ]
    },
    category: 'control',
    paletteGroup: 'core'
  },
  HTTPRequest: {
    componentName: 'HTTPRequest',
    label: 'HTTP request',
    icon: 'link',
    color: '#14b8a6',
    defaultParams: {
      url: 'https://httpbin.org/get',
      method: 'GET',
      timeout: 60,
      headers: {},
      params: {}
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  Iteration: {
    componentName: 'Iteration',
    label: 'Iterate',
    icon: 'refresh',
    color: '#a855f7',
    defaultParams: {
      input: '{sys.query}',
      delimiter: 'comma',
      downstream: ''
    },
    category: 'control',
    paletteGroup: 'core'
  },
  TextProcessing: {
    componentName: 'TextProcessing',
    label: 'Text',
    icon: 'document',
    color: '#64748b',
    defaultParams: {
      method: 'split',
      split_ref: '{retrieval_0@formalized_content}',
      script: '',
      delimiter: 'newline'
    },
    category: 'tool',
    paletteGroup: 'core'
  },
  ExecuteSQL: {
    componentName: 'ExecuteSQL',
    label: 'Run SQL',
    icon: 'data-analysis',
    color: '#0d9488',
    defaultParams: {
      sql: 'SELECT 1',
      db_type: 'mysql',
      host: '127.0.0.1',
      port: 3306,
      username: 'root',
      password: '',
      database: 'test',
      max_records: 1024
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  DocsGenerator: {
    componentName: 'DocsGenerator',
    label: 'Doc export',
    icon: 'document-add',
    color: '#0891b2',
    defaultParams: {
      content: '{llm_0@content}',
      title: '',
      output_format: 'pdf',
      output_dir: '',
      filename: ''
    },
    category: 'output',
    paletteGroup: 'core'
  },
  MCP: {
    componentName: 'MCP',
    label: 'MCP tool',
    icon: 'share',
    color: '#6366f1',
    defaultParams: {
      server_url: '',
      server_name: '',
      tool_name: 'list_all_menus',
      arguments: {}
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  SetVariable: {
    componentName: 'SetVariable',
    label: 'Set variable',
    icon: 'setting',
    color: '#78716c',
    defaultParams: {
      assignments: [{ key: 'sys.custom', value: '{retrieval_0@formalized_content}' }]
    },
    category: 'tool',
    paletteGroup: 'core'
  },
  Transformer: {
    componentName: 'Transformer',
    label: 'LLM transform',
    icon: 'magic-stick',
    color: '#d946ef',
    defaultParams: {
      input: '{retrieval_0@formalized_content}',
      llm_id: 'ollama@llama3.2',
      creativity: 'balance',
      temperature: 0.2,
      instruction: 'Summarize the input text concisely, keeping key facts, in the same language as the user.'
    },
    category: 'generate',
    paletteGroup: 'core'
  },
  AwaitResponse: {
    componentName: 'AwaitResponse',
    label: 'Wait for reply',
    icon: 'chat-line-round',
    color: '#f97316',
    defaultParams: {
      message: '',
      variable_key: 'sys.await_reply',
      require_non_empty: true
    },
    category: 'input',
    paletteGroup: 'core'
  },
  DuckDuckGo: {
    componentName: 'DuckDuckGo',
    label: 'DuckDuckGo',
    icon: 'compass',
    color: '#de5833',
    defaultParams: {
      query: '{sys.query}',
      channel: 'general',
      top_n: 10,
      timeout: 15
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  Wikipedia: {
    componentName: 'Wikipedia',
    label: 'Wikipedia',
    icon: 'reading',
    color: '#1e40af',
    defaultParams: {
      query: '{sys.query}',
      language: 'zh',
      top_n: 5,
      timeout: 30
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  ArXiv: {
    componentName: 'ArXiv',
    label: 'arXiv',
    icon: 'notebook',
    color: '#b91c1c',
    defaultParams: {
      query: '{sys.query}',
      top_n: 10,
      sort_by: 'submittedDate',
      timeout: 20
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  TavilySearch: {
    componentName: 'TavilySearch',
    label: 'Tavily',
    icon: 'trend-charts',
    color: '#7c3aed',
    defaultParams: {
      query: '{sys.query}',
      api_key: '',
      topic: 'general',
      search_depth: 'basic',
      max_results: 6,
      include_answer: false,
      include_domains_csv: '',
      exclude_domains_csv: '',
      timeout: 30
    },
    category: 'tool',
    paletteGroup: 'tool'
  },
  VariableAssigner: {
    componentName: 'VariableAssigner',
    label: 'Var assign',
    icon: 'edit-pen',
    color: '#57534e',
    defaultParams: {
      variables: [
        { variable: 'sys.temp', operator: 'overwrite', parameter: '{sys.query}' }
      ]
    },
    category: 'logic',
    paletteGroup: 'core'
  },
  VariableAggregator: {
    componentName: 'VariableAggregator',
    label: 'Var merge',
    icon: 'connection',
    color: '#0f766e',
    defaultParams: {
      groups: [
        {
          group_name: 'picked',
          variables: [{ value: '{sys.query}' }]
        }
      ]
    },
    category: 'logic',
    paletteGroup: 'core'
  },
  ListOperations: {
    componentName: 'ListOperations',
    label: 'List ops',
    icon: 'sort',
    color: '#7c2d12',
    defaultParams: {
      input: '',
      input_literal: '[]',
      operation: 'topn',
      n: 10,
      field: '',
      value: '',
      filter_operator: '=',
      sort_by: 'letter',
      sort_order: 'asc',
      dedupe_key: ''
    },
    category: 'logic',
    paletteGroup: 'core'
  },
  StringTransform: {
    componentName: 'StringTransform',
    label: 'String transform',
    icon: 'copy-document',
    color: '#4d7c0f',
    defaultParams: {
      mode: 'split',
      input: '{sys.query}',
      input_literal: '',
      delimiters: [','],
      template: '{a}',
      merge_variables: { a: '{sys.query}' }
    },
    category: 'logic',
    paletteGroup: 'core'
  },
  Invoke: {
    componentName: 'Invoke',
    label: 'Invoke',
    icon: 'promotion',
    color: '#0369a1',
    defaultParams: {
      url: 'https://httpbin.org/get',
      method: 'GET',
      datatype: 'json',
      variables: [{ key: 'q', ref: '{sys.query}' }],
      headers: {},
      timeout: 30
    },
    category: 'tool',
    paletteGroup: 'tool'
  }
}

export const COMPONENT_LIST = Object.entries(COMPONENT_TYPES).map(([key, v]) => ({
  key,
  ...v
}))

export const PALETTE_CORE_LIST = COMPONENT_LIST.filter((c) => c.paletteGroup === 'core')

export const PALETTE_TOOL_LIST = COMPONENT_LIST.filter((c) => c.paletteGroup === 'tool')
