package component

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	ragregistry "github.com/LightningRAG/LightningRAG/server/rag/registry"
)

func init() {
	Register("Transformer", NewTransformer)
}

// Transformer 使用 LLM 对输入文本做转换、摘要或结构化抽取（对齐上游编排 Transformer 编排语义）
type Transformer struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewTransformer 创建 Transformer 组件
func NewTransformer(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Transformer{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (t *Transformer) ComponentName() string { return "Transformer" }

func (t *Transformer) Invoke(map[string]any) error {
	t.mu.Lock()
	t.err = ""
	t.mu.Unlock()

	inputRef := strings.TrimSpace(getStrParam(t.params, "input"))
	if inputRef == "" {
		t.fail("input 不能为空")
		return fmt.Errorf("input 不能为空")
	}
	raw, ok := t.canvas.GetVariableValue(inputRef)
	if !ok || raw == nil {
		t.fail("无法解析输入变量: " + inputRef)
		return fmt.Errorf("无法解析输入变量: %s", inputRef)
	}
	inputText := strings.TrimSpace(fmt.Sprint(raw))
	if inputText == "" {
		t.mu.Lock()
		t.output["content"] = ""
		t.output["formalized_content"] = ""
		t.mu.Unlock()
		return nil
	}

	sysPrompt := t.canvas.ResolveString(getStrParam(t.params, "instruction"))
	if sysPrompt == "" {
		sysPrompt = "Organize, summarize, or extract key information from the following text as requested. Output only the result, without extra explanation."
	}
	userContent := "Text to process:\n---\n" + inputText + "\n---"

	llmInst, err := t.createLLM()
	if err != nil {
		t.fail(err.Error())
		return err
	}

	msgs := []interfaces.MessageContent{
		{
			Role:  interfaces.MessageRoleSystem,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: sysPrompt}},
		},
		{
			Role:  interfaces.MessageRoleHuman,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: userContent}},
		},
	}

	ctx := context.Background()
	var opts []interfaces.CallOption
	if rc := t.canvas.RunContext(); rc != nil && rc.GetStreamCallback() != nil {
		opts = append(opts, interfaces.WithStreamCallback(rc.GetStreamCallback()))
	}
	if temp := t.getTemperature(); temp >= 0 {
		opts = append(opts, interfaces.WithTemperature(float32(temp)))
	}
	if n := getIntParam(t.params, "max_tokens", 0); n > 0 {
		opts = append(opts, interfaces.WithMaxTokens(n))
	}

	resp, err := llmInst.GenerateContent(ctx, msgs, opts...)
	if err != nil {
		t.fail(err.Error())
		return err
	}

	text := ""
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Content
	}
	t.mu.Lock()
	t.output["content"] = text
	t.output["formalized_content"] = text
	t.mu.Unlock()
	return nil
}

func (t *Transformer) fail(msg string) {
	t.mu.Lock()
	t.err = msg
	t.mu.Unlock()
}

func (t *Transformer) createLLM() (interfaces.LLM, error) {
	provider, modelName, baseURL, apiKey := resolveComponentLLMConfig(t.canvas, t.params)
	llm, err := ragregistry.CreateLLM(ragregistry.LLMConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
	if err != nil || llm == nil {
		return nil, fmt.Errorf("创建 LLM 失败: %v", err)
	}
	return llm, nil
}

func (t *Transformer) getTemperature() float64 {
	creativity := getStrParam(t.params, "creativity")
	switch strings.ToLower(creativity) {
	case "improvise":
		return 0.8
	case "balance":
		return 0.5
	case "precise":
		return 0.1
	}
	return getFloatParam(t.params, "temperature", 0.2)
}

func (t *Transformer) Output(key string) any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.output[key]
}

func (t *Transformer) OutputAll() map[string]any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range t.output {
		out[k] = v
	}
	return out
}

func (t *Transformer) SetOutput(key string, value any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output[key] = value
}

func (t *Transformer) Error() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.err
}

func (t *Transformer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output = make(map[string]any)
	t.err = ""
}
