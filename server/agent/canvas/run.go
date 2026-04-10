package canvas

import (
	"context"
	"fmt"
)

func bumpConversationTurns(g map[string]any) {
	if g == nil {
		return
	}
	n := 0
	if v, ok := g["sys.conversation_turns"]; ok && v != nil {
		switch x := v.(type) {
		case int:
			n = x
		case int64:
			n = int(x)
		case float64:
			n = int(x)
		case float32:
			n = int(x)
		}
	}
	g["sys.conversation_turns"] = n + 1
}

// Run 执行工作流（支持分支组件 Switch/Categorize 动态选择下游）
func (c *Canvas) Run(ctx context.Context, input RunInput) (map[string]any, error) {
	for k, v := range input.WorkflowGlobals {
		if k != "" {
			c.globals[k] = v
		}
	}
	c.globals["sys.query"] = input.Query
	c.globals["sys.files"] = input.Files
	c.globals["sys.user_id"] = input.UserID
	// 与 ragflow Canvas.run 一致：每次执行递增轮次，供提示词或分支使用
	bumpConversationTurns(c.globals)

	// 设置运行上下文（流式回调、历史消息），供 LLM 组件使用
	c.runCtx = &runContext{onChunk: input.OnChunk, history: input.History}
	defer func() { c.runCtx = nil }()

	path := c.buildPath()
	if len(path) == 0 {
		return nil, fmt.Errorf("工作流 path 为空")
	}

	// 重置所有组件
	for _, comp := range c.components {
		comp.Reset()
	}

	// 执行 begin（入口 id 按组件类型解析，有环时 path 顺序不可靠，不能仅用 path[0]）
	beginID := c.entryComponentID(path)
	if beginID == "" {
		return nil, fmt.Errorf("未找到 Begin 入口组件")
	}
	begin, ok := c.components[beginID]
	if !ok {
		return nil, fmt.Errorf("未找到 begin 组件: %s", beginID)
	}
	if err := begin.Invoke(map[string]any{
		"query":   input.Query,
		"request": input.Query,
	}); err != nil {
		return nil, err
	}

	// 图式执行：从 begin 开始，按 downstream 或分支组件的 next_id 推进
	var lastExecuted string = beginID
	pausedAtEntry := false
	currentID := c.getNextComponent(beginID, begin)
	for currentID != "" {
		// 连回入口（Begin）：ragflow 中回到开始等价于等待下一轮对话输入；同一次 Run 内不再重入 Begin，
		// 避免与 Message→Begin 闭环时在同一条用户消息里死循环输出。
		if currentID == beginID {
			pausedAtEntry = true
			break
		}
		comp, ok := c.components[currentID]
		if !ok {
			return nil, fmt.Errorf("未找到组件: %s", currentID)
		}
		if err := comp.Invoke(nil); err != nil {
			return nil, err
		}
		if comp.Error() != "" {
			return nil, fmt.Errorf("组件 %s 执行失败: %s", currentID, comp.Error())
		}
		lastExecuted = currentID
		currentID = c.getNextComponent(currentID, comp)
	}

	base := c.components[lastExecuted].OutputAll()
	out := make(map[string]any, len(base)+2)
	for k, v := range base {
		out[k] = v
	}
	if pausedAtEntry {
		out["workflowPausedAtEntry"] = true
	}
	return out, nil
}

// entryComponentID 返回 component_name 为 Begin 的节点 id，用于入口执行与「连回开始」判断。
func (c *Canvas) entryComponentID(path []string) string {
	for id, def := range c.dsl.Components {
		if def.Obj.ComponentName == "Begin" {
			return id
		}
	}
	if len(path) > 0 {
		return path[0]
	}
	return ""
}

// getNextComponent 获取下一个应执行的组件 ID
// 分支组件（Switch/Categorize）通过 Output("next_id") 指定；否则取 downstream[0]
func (c *Canvas) getNextComponent(currentID string, comp interface {
	Output(key string) any
}) string {
	if next := comp.Output("next_id"); next != nil {
		if s, ok := next.(string); ok && s != "" {
			if _, exists := c.components[s]; exists {
				return s
			}
		}
	}
	def, ok := c.dsl.Components[currentID]
	if !ok || len(def.Downstream) == 0 {
		return ""
	}
	return def.Downstream[0]
}

// buildPath 构建执行路径：从 begin 开始按 downstream 拓扑排序
func (c *Canvas) buildPath() []string {
	if len(c.path) > 0 {
		return c.path
	}
	// 从 DSL 计算 path
	visited := make(map[string]bool)
	var path []string
	var dfs func(id string)
	dfs = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		def, ok := c.dsl.Components[id]
		if !ok {
			return
		}
		for _, up := range def.Upstream {
			dfs(up)
		}
		path = append(path, id)
	}
	// 找到 Begin 入口（与 entryComponentID 一致，不假定节点 id 为 "begin"）
	entry := ""
	for id, def := range c.dsl.Components {
		if def.Obj.ComponentName == "Begin" {
			entry = id
			break
		}
	}
	if entry != "" {
		dfs(entry)
	} else if len(path) == 0 {
		for id := range c.dsl.Components {
			dfs(id)
			break
		}
	}
	return path
}
