/**
 * DSL 与 VueFlow 节点/边 双向转换
 * DSL 结构参考 server/agent/dsl/types.go
 */

/**
 * 从 DSL 转换为 VueFlow 的 nodes 和 edges
 * @param {Object} dsl - { components, path, globals, ... }
 * @returns {{ nodes: Array, edges: Array }}
 */
export function dslToFlow(dsl) {
  if (!dsl?.components) {
    return { nodes: [], edges: [] }
  }
  const components = dsl.components
  const path = dsl.path || Object.keys(components)

  const nodeWidth = 180
  const nodeHeight = 80
  const gapX = 80
  const gapY = 120

  const nodes = []
  const edges = []
  const posMap = {}

  path.forEach((id, idx) => {
    const comp = components[id]
    if (!comp) return

    const row = Math.floor(idx / 3)
    const col = idx % 3
    const x = 100 + col * (nodeWidth + gapX)
    const y = 80 + row * (nodeHeight + gapY)
    posMap[id] = { x, y }

    const obj = comp.obj || {}
    const params = obj.params || {}
    const componentName = obj.component_name || ''

    const labelMap = {
      Begin: '开始',
      Retrieval: '检索',
      LLM: '生成',
      Agent: '智能体',
      Message: '输出',
      Switch: '条件分支',
      Categorize: '意图分类',
      HTTPRequest: 'HTTP 请求',
      Iteration: '迭代',
      TextProcessing: '文本处理',
      ExecuteSQL: '执行 SQL',
      DocsGenerator: '文档生成',
      MCP: 'MCP 工具',
      SetVariable: '设置变量',
      Transformer: 'LLM 转换',
      AwaitResponse: '等待回复',
      DuckDuckGo: 'DuckDuckGo',
      Wikipedia: '维基百科',
      ArXiv: 'arXiv',
      TavilySearch: 'Tavily 搜索',
      VariableAssigner: '变量赋值',
      VariableAggregator: '变量聚合',
      ListOperations: '列表操作',
      StringTransform: '字符串变换',
      Invoke: 'Invoke 请求'
    }
    const colorMap = {
      Begin: '#22c55e',
      Retrieval: '#3b82f6',
      LLM: '#8b5cf6',
      Agent: '#0ea5e9',
      Message: '#f59e0b',
      Switch: '#06b6d4',
      Categorize: '#ec4899',
      HTTPRequest: '#14b8a6',
      Iteration: '#a855f7',
      TextProcessing: '#64748b',
      ExecuteSQL: '#0d9488',
      DocsGenerator: '#0891b2',
      MCP: '#6366f1',
      SetVariable: '#78716c',
      Transformer: '#d946ef',
      AwaitResponse: '#f97316',
      DuckDuckGo: '#de5833',
      Wikipedia: '#1e40af',
      ArXiv: '#b91c1c',
      TavilySearch: '#7c3aed',
      VariableAssigner: '#57534e',
      VariableAggregator: '#0f766e',
      ListOperations: '#7c2d12',
      StringTransform: '#4d7c0f',
      Invoke: '#0369a1'
    }

    nodes.push({
      id,
      type: 'agent',
      position: { x, y },
      data: {
        componentName,
        label: labelMap[componentName] || componentName,
        color: colorMap[componentName] || '#6366f1',
        params: { ...params }
      }
    })

    ;(comp.downstream || []).forEach((targetId) => {
      if (components[targetId]) {
        edges.push({
          id: `e-${id}-${targetId}`,
          source: id,
          target: targetId
        })
      }
    })
  })

  return { nodes, edges }
}

/**
 * 从 VueFlow 的 nodes 和 edges 转换为 DSL
 * @param {Array} nodes - VueFlow nodes
 * @param {Array} edges - VueFlow edges
 * @returns {Object} DSL
 */
export function flowToDsl(nodes, edges) {
  const components = {}
  const edgeMap = {}
  const nodesArr = Array.isArray(nodes) ? nodes : (nodes ? Array.from(nodes) : [])
  const edgesArr = Array.isArray(edges) ? edges : (edges ? Array.from(edges) : [])
  edgesArr.forEach((e) => {
    edgeMap[e.source] = edgeMap[e.source] || []
    edgeMap[e.source].push(e.target)
  })

  const defaultGlobals = {
    'sys.query': '',
    'sys.user_id': '',
    'sys.conversation_turns': 0,
    'sys.files': [],
    'sys.await_reply': '',
    'env.tavily_api_key': ''
  }

  nodesArr.forEach((n) => {
    const id = n.id
    const d = n.data || {}
    const params = d.params || {}

    components[id] = {
      obj: {
        component_name: d.componentName || 'Begin',
        params: { ...params }
      },
      downstream: edgeMap[id] || [],
      upstream: []
    }
  })

  // 计算 upstream
  Object.keys(components).forEach((id) => {
    components[id].downstream.forEach((targetId) => {
      if (components[targetId]) {
        components[targetId].upstream = components[targetId].upstream || []
        if (!components[targetId].upstream.includes(id)) {
          components[targetId].upstream.push(id)
        }
      }
    })
  })

  // 拓扑排序得到 path
  const path = topologicalSort(components)

  return {
    components,
    path,
    globals: defaultGlobals,
    history: [],
    retrieval: []
  }
}

function topologicalSort(components) {
  const inDegree = {}
  const compIds = Object.keys(components)
  compIds.forEach((id) => { inDegree[id] = 0 })
  Object.entries(components).forEach(([id, comp]) => {
    (comp.downstream || []).forEach((targetId) => {
      if (components[targetId]) inDegree[targetId]++
    })
  })

  const queue = compIds.filter((id) => inDegree[id] === 0)
  const sorted = []
  while (queue.length) {
    const id = queue.shift()
    sorted.push(id)
    const downstream = components[id]?.downstream || []
    downstream.forEach((targetId) => {
      if (components[targetId] !== undefined) {
        inDegree[targetId]--
        if (inDegree[targetId] === 0) queue.push(targetId)
      }
    })
  }
  return sorted.length === compIds.length ? [...sorted] : [...compIds]
}
