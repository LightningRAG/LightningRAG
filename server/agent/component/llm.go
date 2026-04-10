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
	Register("LLM", NewLLM)
}

// LLM 大模型生成组件
type LLM struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewLLM 创建 LLM 组件
func NewLLM(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &LLM{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (l *LLM) ComponentName() string {
	return "LLM"
}

// Invoke 执行
func (l *LLM) Invoke(inputs map[string]any) error {
	l.mu.Lock()
	l.err = ""
	l.mu.Unlock()

	sysPrompt := l.canvas.ResolveString(getStrParam(l.params, "sys_prompt"))
	prompts := getPrompts(l.params)
	if len(prompts) == 0 {
		prompts = []string{"{sys.query}"}
	}
	// 解析 prompts 中的变量
	var userParts []string
	for _, p := range prompts {
		userParts = append(userParts, l.canvas.ResolveString(p))
	}
	userContent := strings.Join(userParts, "\n")

	llmInst, err := l.createLLM()
	if err != nil {
		l.mu.Lock()
		l.err = err.Error()
		l.mu.Unlock()
		return err
	}

	msgs := l.buildMessages(sysPrompt, userContent)

	ctx := context.Background()
	var opts []interfaces.CallOption
	if rc := l.canvas.RunContext(); rc != nil && rc.GetStreamCallback() != nil {
		opts = append(opts, interfaces.WithStreamCallback(rc.GetStreamCallback()))
	}
	if t := l.getTemperature(); t >= 0 {
		opts = append(opts, interfaces.WithTemperature(float32(t)))
	}
	if p := l.getTopP(); p > 0 && p <= 1 {
		opts = append(opts, interfaces.WithTopP(float32(p)))
	}
	if n := getIntParam(l.params, "max_tokens", 0); n > 0 {
		opts = append(opts, interfaces.WithMaxTokens(n))
	}
	resp, err := llmInst.GenerateContent(ctx, msgs, opts...)
	if err != nil {
		l.mu.Lock()
		l.err = err.Error()
		l.mu.Unlock()
		return err
	}

	l.mu.Lock()
	if len(resp.Choices) > 0 {
		l.output["content"] = resp.Choices[0].Content
	} else {
		l.output["content"] = ""
	}
	l.mu.Unlock()
	return nil
}

// buildMessages 构建消息列表：系统提示 + 历史消息 + 当前用户消息
func (l *LLM) buildMessages(sysPrompt, userContent string) []interfaces.MessageContent {
	var msgs []interfaces.MessageContent
	if sysPrompt != "" {
		msgs = append(msgs, interfaces.MessageContent{
			Role:  interfaces.MessageRoleSystem,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: sysPrompt}},
		})
	}
	if rc := l.canvas.RunContext(); rc != nil && len(rc.GetHistory()) > 0 {
		for _, h := range rc.GetHistory() {
			role := interfaces.MessageRoleHuman
			if h.Role == "assistant" {
				role = interfaces.MessageRoleAssistant
			} else if h.Role == "system" {
				role = interfaces.MessageRoleSystem
			}
			msgs = append(msgs, interfaces.MessageContent{
				Role:  role,
				Parts: []interfaces.ContentPart{interfaces.TextContent{Text: h.Content}},
			})
		}
	}
	msgs = append(msgs, interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: userContent}},
	})
	return msgs
}

func (l *LLM) createLLM() (interfaces.LLM, error) {
	provider, modelName, baseURL, apiKey := resolveComponentLLMConfig(l.canvas, l.params)

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

// Output 获取输出
func (l *LLM) Output(key string) any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.output[key]
}

// OutputAll 获取所有输出
func (l *LLM) OutputAll() map[string]any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range l.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (l *LLM) SetOutput(key string, value any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output[key] = value
}

// Error 返回错误
func (l *LLM) Error() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.err
}

// Reset 重置
func (l *LLM) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = make(map[string]any)
	l.err = ""
}

// getTemperature 根据 creativity 预设或 temperature 参数返回温度
func (l *LLM) getTemperature() float64 {
	creativity := getStrParam(l.params, "creativity")
	switch strings.ToLower(creativity) {
	case "improvise":
		return 0.8
	case "balance":
		return 0.5
	case "precise":
		return 0.1
	}
	return getFloatParam(l.params, "temperature", 0.1)
}

func (l *LLM) getTopP() float64 {
	return getFloatParam(l.params, "top_p", 0)
}

func getPrompts(m map[string]any) []string {
	v, ok := m["prompts"]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			if c, ok := m["content"].(string); ok {
				out = append(out, c)
			}
		}
	}
	return out
}
