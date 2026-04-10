package request

// AgentRun 运行 Agent 请求（支持 agent_id 或 dsl）
type AgentRun struct {
	AgentID         uint           `json:"agentId"` // 可选，有则从数据库加载 DSL
	DSL             map[string]any `json:"dsl"`     // 可选，直接传入 DSL；agentId 优先
	Query           string         `json:"query" binding:"required"`
	ConversationID  uint           `json:"conversationId"`  // 可选，多轮对话时传入，用于记忆上下文
	WorkflowGlobals map[string]any `json:"workflowGlobals"` // 可选，合并进画布 globals，供 AwaitResponse 等组件读取
}

// AgentCreate 创建 Agent
type AgentCreate struct {
	Name string         `json:"name" binding:"required"`
	Desc string         `json:"desc"`
	DSL  map[string]any `json:"dsl" binding:"required"`
}

// AgentUpdate 更新 Agent
type AgentUpdate struct {
	ID   uint           `json:"id" binding:"required"`
	Name string         `json:"name"`
	Desc string         `json:"desc"`
	DSL  map[string]any `json:"dsl"`
}

// AgentList 列表请求
type AgentList struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Name     string `json:"name" form:"name"`
}

// AgentGet 获取单个
type AgentGet struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// AgentDelete 删除
type AgentDelete struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// AgentCreateFromTemplate 从模板创建
type AgentCreateFromTemplate struct {
	TemplateName string `json:"templateName" binding:"required"` // 如 retrieval_and_generate
	Name         string `json:"name" binding:"required"`
	Desc         string `json:"desc"`
}

// AgentLoadTemplate 加载模板
type AgentLoadTemplate struct {
	TemplateName string `json:"templateName" binding:"required"`
}
