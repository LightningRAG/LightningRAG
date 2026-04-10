// Package canvas Agent 流程编排执行引擎
package canvas

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/agent/component"
	"github.com/LightningRAG/LightningRAG/server/agent/dsl"
)

// 确保 Canvas 实现 RunContext
var _ component.RunContext = (*runContext)(nil)

type runContext struct {
	onChunk func(string)
	history []component.HistoryMessage
}

func (r *runContext) GetStreamCallback() func(string)        { return r.onChunk }
func (r *runContext) GetHistory() []component.HistoryMessage { return r.history }

// RunInput 运行输入
type RunInput struct {
	Query           string
	Files           []string
	UserID          uint
	OnChunk         func(string)               // 流式输出回调，非空时 LLM 组件将流式返回
	History         []component.HistoryMessage // 多轮对话历史，将注入 LLM 上下文
	ConversationID  uint                       // 对话 ID，用于保存消息（由 service 层处理）
	WorkflowGlobals map[string]any             // 与 DSL globals 合并，用于多轮补全 AwaitResponse 等变量
}

// Canvas 画布执行引擎
type Canvas struct {
	dsl               *dsl.DSL
	components        map[string]component.Component
	path              []string
	globals           map[string]any
	history           [][]any
	retrieval         []map[string]any
	tenantID          uint
	retrieverFactory  component.RetrieverFactory
	llmConfigResolver component.LLMConfigResolver
	runCtx            *runContext // 当前运行的上下文（流式、历史）
}

// New 创建画布
func New(d *dsl.DSL, tenantID uint, retrieverFactory component.RetrieverFactory, llmConfigResolver component.LLMConfigResolver) (*Canvas, error) {
	c := &Canvas{
		dsl:               d,
		components:        make(map[string]component.Component),
		path:              d.Path,
		globals:           d.Globals,
		history:           d.History,
		retrieval:         d.Retrieval,
		tenantID:          tenantID,
		retrieverFactory:  retrieverFactory,
		llmConfigResolver: llmConfigResolver,
	}
	if c.globals == nil {
		c.globals = dsl.DefaultGlobals()
	}
	if c.path == nil {
		c.path = []string{}
	}
	if c.retrieval == nil {
		c.retrieval = []map[string]any{}
	}
	if err := c.loadComponents(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Canvas) loadComponents() error {
	for id, def := range c.dsl.Components {
		comp, err := component.Create(c, id, def)
		if err != nil {
			return err
		}
		c.components[id] = comp
	}
	return nil
}

// GetComponent 获取组件
func (c *Canvas) GetComponent(cpnID string) (component.Component, bool) {
	comp, ok := c.components[cpnID]
	return comp, ok
}

// GetComponentObj 获取组件对象
func (c *Canvas) GetComponentObj(cpnID string) component.Component {
	return c.components[cpnID]
}

// GetVariableValue 解析变量引用，如 {sys.query} 或 {retrieval_0@formalized_content}
func (c *Canvas) GetVariableValue(exp string) (any, bool) {
	exp = strings.Trim(exp, "{} \t")
	if exp == "" {
		return nil, false
	}
	if strings.Index(exp, "@") < 0 {
		if v, ok := c.globals[exp]; ok {
			return v, true
		}
		return nil, false
	}
	parts := strings.SplitN(exp, "@", 2)
	if len(parts) != 2 {
		return nil, false
	}
	cpnID, varPath := parts[0], parts[1]
	comp, ok := c.components[cpnID]
	if !ok {
		return nil, false
	}
	val := comp.Output(varPath)
	return val, true
}

// SetVariableValue 设置变量
func (c *Canvas) SetVariableValue(exp string, value any) {
	exp = strings.Trim(exp, "{} \t")
	if exp == "" {
		return
	}
	if strings.Index(exp, "@") < 0 {
		c.globals[exp] = value
		return
	}
	parts := strings.SplitN(exp, "@", 2)
	if len(parts) != 2 {
		return
	}
	cpnID, varPath := parts[0], parts[1]
	comp, ok := c.components[cpnID]
	if !ok {
		return
	}
	comp.SetOutput(varPath, value)
}

// IsVariableRef 是否为变量引用
func (c *Canvas) IsVariableRef(exp string) bool {
	exp = strings.Trim(exp, "{} \t")
	if exp == "" {
		return false
	}
	if strings.Index(exp, "@") < 0 {
		_, ok := c.globals[exp]
		return ok
	}
	parts := strings.SplitN(exp, "@", 2)
	if len(parts) != 2 {
		return false
	}
	_, ok := c.components[parts[0]]
	return ok
}

// GetGlobals 获取全局变量
func (c *Canvas) GetGlobals() map[string]any {
	return c.globals
}

// GetTenantID 获取租户 ID
func (c *Canvas) GetTenantID() uint {
	return c.tenantID
}

// ResolveString 解析字符串中的变量引用，替换为实际值
func (c *Canvas) ResolveString(s string) string {
	return component.VariableRefPattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := regexp.MustCompile(`\{([^}]+)\}`).FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := strings.TrimSpace(sub[1])
		if v, ok := c.GetVariableValue(key); ok && v != nil {
			if str, ok := v.(string); ok {
				return str
			}
			return fmt.Sprint(v)
		}
		return ""
	})
}

// GetRetrieverFactory 获取检索器工厂
func (c *Canvas) GetRetrieverFactory() component.RetrieverFactory {
	return c.retrieverFactory
}

// GetLLMConfigResolver 获取 LLM 配置解析器
func (c *Canvas) GetLLMConfigResolver() component.LLMConfigResolver {
	return c.llmConfigResolver
}

// RunContext 返回当前运行上下文，供 LLM 等组件使用
func (c *Canvas) RunContext() component.RunContext {
	if c.runCtx != nil {
		return c.runCtx
	}
	return &runContext{}
}

// InvokeComponent 执行指定组件（供 Iteration 等循环组件调用）
func (c *Canvas) InvokeComponent(cpnID string) error {
	comp, ok := c.components[cpnID]
	if !ok {
		return fmt.Errorf("未找到组件: %s", cpnID)
	}
	if err := comp.Invoke(nil); err != nil {
		return err
	}
	if comp.Error() != "" {
		return fmt.Errorf("组件 %s 执行失败: %s", cpnID, comp.Error())
	}
	return nil
}

// GetComponentDownstream 获取组件的下游 ID 列表
func (c *Canvas) GetComponentDownstream(cpnID string) []string {
	def, ok := c.dsl.Components[cpnID]
	if !ok {
		return nil
	}
	return def.Downstream
}
