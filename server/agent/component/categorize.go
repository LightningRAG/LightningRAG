package component

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	ragregistry "github.com/LightningRAG/LightningRAG/server/rag/registry"
)

func init() {
	Register("Categorize", NewCategorize)
}

// CategorizeCategory 分类项
type CategorizeCategory struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	Downstream  string   `json:"downstream"`
}

// Categorize LLM 意图分类组件，根据分类结果选择下游
type Categorize struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewCategorize 创建 Categorize 组件
func NewCategorize(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Categorize{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (c *Categorize) ComponentName() string {
	return "Categorize"
}

// Invoke 执行：调用 LLM 分类，输出 selected_category 和 next_id
func (c *Categorize) Invoke(inputs map[string]any) error {
	c.mu.Lock()
	c.err = ""
	c.mu.Unlock()

	categories := c.getCategories()
	if len(categories) == 0 {
		c.mu.Lock()
		c.err = "Categorize 至少需要两个分类"
		c.mu.Unlock()
		return fmt.Errorf("Categorize 至少需要两个分类")
	}

	inputVar := getStrParam(c.params, "input")
	if inputVar == "" {
		inputVar = "sys.query"
	}
	inputText := c.canvas.ResolveString(NormalizeSingleRefForResolve(inputVar))
	if inputText == "" {
		if v, ok := c.canvas.GetVariableValue(inputVar); ok && v != nil {
			inputText = fmt.Sprint(v)
		}
	}

	sysPrompt := c.buildCategorizePrompt(categories)
	userContent := "User input:\n" + inputText + "\n\nOutput only the category name, nothing else."

	llm, err := c.createLLM()
	if err != nil {
		c.mu.Lock()
		c.err = err.Error()
		c.mu.Unlock()
		return err
	}

	msgs := []interfaces.MessageContent{
		{Role: interfaces.MessageRoleSystem, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: sysPrompt}}},
		{Role: interfaces.MessageRoleHuman, Parts: []interfaces.ContentPart{interfaces.TextContent{Text: userContent}}},
	}

	resp, err := llm.GenerateContent(context.Background(), msgs)
	if err != nil {
		c.mu.Lock()
		c.err = err.Error()
		c.mu.Unlock()
		return err
	}

	raw := ""
	if len(resp.Choices) > 0 {
		raw = strings.TrimSpace(resp.Choices[0].Content)
	}

	// 解析 LLM 输出，匹配分类名，得到 downstream 组件 ID
	nextID := c.matchCategory(raw, categories)
	if nextID == "" {
		nextID = categories[0].Downstream
	}

	c.mu.Lock()
	c.output["selected_category"] = raw
	c.output["next_id"] = nextID
	c.mu.Unlock()
	return nil
}

func (c *Categorize) getCategories() []CategorizeCategory {
	v, ok := c.params["categories"]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []CategorizeCategory
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cat := CategorizeCategory{
			Name:        getStrParam(m, "name"),
			Description: getStrParam(m, "description"),
			Downstream:  getStrParam(m, "downstream"),
		}
		if ex, ok := m["examples"].([]any); ok {
			for _, e := range ex {
				if s, ok := e.(string); ok {
					cat.Examples = append(cat.Examples, s)
				}
			}
		}
		out = append(out, cat)
	}
	return out
}

func (c *Categorize) buildCategorizePrompt(categories []CategorizeCategory) string {
	var sb strings.Builder
	sb.WriteString("You are an intent classifier. Pick the single best-matching category for the user input.\n\n")
	sb.WriteString("Categories:\n")
	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("- %s: %s", cat.Name, cat.Description))
		if len(cat.Examples) > 0 {
			sb.WriteString(" Examples: " + strings.Join(cat.Examples, ", "))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\nOutput only the category name (e.g. " + categories[0].Name + "), nothing else.")
	return sb.String()
}

func (c *Categorize) matchCategory(raw string, categories []CategorizeCategory) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	for _, cat := range categories {
		name := strings.ToLower(cat.Name)
		if raw == name || strings.Contains(raw, name) {
			return cat.Downstream
		}
	}
	// 尝试模糊匹配
	for _, cat := range categories {
		if regexp.MustCompile(`(?i)` + regexp.QuoteMeta(cat.Name)).MatchString(raw) {
			return cat.Downstream
		}
	}
	return ""
}

func (c *Categorize) createLLM() (interfaces.LLM, error) {
	provider, modelName, baseURL, apiKey := resolveComponentLLMConfig(c.canvas, c.params)

	return ragregistry.CreateLLM(ragregistry.LLMConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
}

// Output 获取输出
func (c *Categorize) Output(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.output[key]
}

// OutputAll 获取所有输出
func (c *Categorize) OutputAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range c.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (c *Categorize) SetOutput(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.output[key] = value
}

// Error 返回错误
func (c *Categorize) Error() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.err
}

// Reset 重置
func (c *Categorize) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.output = make(map[string]any)
	c.err = ""
}
