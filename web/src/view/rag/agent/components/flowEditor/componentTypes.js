/**
 * Agent 组件类型定义（与后端 DSL component_name 对应）
 * paletteGroup: 侧栏展示分组，与本产品「核心 / 工具」侧栏分组一致
 *   core — 编排主干（流程、生成、分支、变量等）
 *   tool — 外部调用与扩展能力（HTTP、搜索、MCP、SQL 等）
 */
export const COMPONENT_TYPES = {
  Begin: {
    componentName: 'Begin',
    label: '开始',
    icon: 'play-circle',
    color: '#22c55e',
    defaultParams: { prologue: '你好！' },
    category: 'input',
    paletteGroup: 'core'
  },
  Retrieval: {
    componentName: 'Retrieval',
    label: '检索',
    icon: 'search',
    color: '#3b82f6',
    defaultParams: { query: '{sys.query}', top_n: 20, empty_response: '未找到相关内容', kb_ids: [] },
    category: 'retrieval',
    paletteGroup: 'core'
  },
  LLM: {
    componentName: 'LLM',
    label: '生成',
    icon: 'chat-dot-round',
    color: '#8b5cf6',
    defaultParams: {
      llm_id: 'ollama@llama3.2',
      creativity: 'precise',
      temperature: 0.1,
      sys_prompt: '根据知识库内容回答。\n\n知识库：\n{retrieval_0@formalized_content}',
      prompts: [{ role: 'user', content: '用户问题：{sys.query}' }]
    },
    category: 'generate',
    paletteGroup: 'core'
  },
  Agent: {
    componentName: 'Agent',
    label: '智能体',
    icon: 'avatar',
    color: '#0ea5e9',
    defaultParams: {
      llm_id: 'ollama@llama3.2',
      creativity: 'balance',
      sys_prompt: '你是一个智能助手，请根据用户需求完成任务。',
      user_prompt: '{sys.query}',
      max_retries: 1,
      delay_after_error: 1
    },
    category: 'generate',
    paletteGroup: 'core'
  },
  Message: {
    componentName: 'Message',
    label: '输出',
    icon: 'message',
    color: '#f59e0b',
    defaultParams: { content: ['{generate_0@content}'] },
    category: 'output',
    paletteGroup: 'core'
  },
  Switch: {
    componentName: 'Switch',
    label: '条件分支',
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
    label: '意图分类',
    icon: 'collection-tag',
    color: '#ec4899',
    defaultParams: {
      input: '{sys.query}',
      llm_id: 'ollama@llama3.2',
      categories: [
        { name: 'qa', description: '知识问答', examples: ['什么是X?'], downstream: '' },
        { name: 'chat', description: '闲聊', examples: ['你好'], downstream: '' }
      ]
    },
    category: 'control',
    paletteGroup: 'core'
  },
  HTTPRequest: {
    componentName: 'HTTPRequest',
    label: 'HTTP 请求',
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
    label: '迭代',
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
    label: '文本处理',
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
    label: '执行 SQL',
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
    label: '文档生成',
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
    label: 'MCP 工具',
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
    label: '设置变量',
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
    label: 'LLM 转换',
    icon: 'magic-stick',
    color: '#d946ef',
    defaultParams: {
      input: '{retrieval_0@formalized_content}',
      llm_id: 'ollama@llama3.2',
      creativity: 'balance',
      temperature: 0.2,
      instruction: '请对输入文本做简要摘要，保留关键事实，使用与用户相同的语言输出。'
    },
    category: 'generate',
    paletteGroup: 'core'
  },
  AwaitResponse: {
    componentName: 'AwaitResponse',
    label: '等待回复',
    icon: 'chat-line-round',
    color: '#f97316',
    defaultParams: {
      message: '请补充更多信息（例如具体日期、订单号）：',
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
    label: '维基百科',
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
    label: 'Tavily 搜索',
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
    label: '变量赋值',
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
    label: '变量聚合',
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
    label: '列表操作',
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
    label: '字符串变换',
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
    label: 'Invoke 请求',
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

/** 侧栏：核心组件（编排主干） */
export const PALETTE_CORE_LIST = COMPONENT_LIST.filter((c) => c.paletteGroup === 'core')

/** 侧栏：工具（外部能力） */
export const PALETTE_TOOL_LIST = COMPONENT_LIST.filter((c) => c.paletteGroup === 'tool')
