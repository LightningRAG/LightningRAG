// Package dsl 定义 Agent 流程编排的 DSL 数据结构
// 参考 references 目录内 Agent 参考实现
package dsl

// ComponentObj 组件对象定义
type ComponentObj struct {
	ComponentName string         `json:"component_name"`
	Params        map[string]any `json:"params"`
}

// ComponentDef 单个组件定义，含上下游
type ComponentDef struct {
	Obj        ComponentObj `json:"obj"`
	Downstream []string     `json:"downstream"`
	Upstream   []string     `json:"upstream"`
	ParentID   string       `json:"parent_id,omitempty"`
}

// DSL Agent 工作流 DSL
type DSL struct {
	Components map[string]ComponentDef `json:"components"`
	Path       []string                `json:"path"`
	Globals    map[string]any          `json:"globals"`
	History    [][]any                 `json:"history"`
	Retrieval  []map[string]any        `json:"retrieval"`
	Memory     []any                   `json:"memory"`
	Graph      *GraphDef               `json:"graph,omitempty"`
}

// GraphDef 前端画布图结构（节点、边）
type GraphDef struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode 画布节点
type GraphNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Position map[string]any `json:"position"`
	Data     map[string]any `json:"data"`
}

// GraphEdge 画布边
type GraphEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// DefaultGlobals 默认全局变量
func DefaultGlobals() map[string]any {
	return map[string]any{
		"sys.query":              "",
		"sys.user_id":            "",
		"sys.conversation_turns": 0,
		"sys.files":              []any{},
		"sys.history":            []any{},
		"sys.await_reply":        "",
		"env.tavily_api_key":     "",
	}
}
