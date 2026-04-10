package component

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	ragregistry "github.com/LightningRAG/LightningRAG/server/rag/registry"
)

func init() {
	Register("Agent", NewAgent)
}

// Agent 智能体组件：LLM + 推理能力，支持重试（后续可扩展 Tools）
// 参照 references 目录内 Agent，当前为 standalone 模式，与 LLM 类似但支持 max_retries、delay_after_error
type Agent struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewAgent 创建 Agent 组件
func NewAgent(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Agent{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (a *Agent) ComponentName() string {
	return "Agent"
}

// Invoke 执行：与 LLM 类似，支持重试
func (a *Agent) Invoke(inputs map[string]any) error {
	a.mu.Lock()
	a.err = ""
	a.mu.Unlock()

	sysPrompt := a.canvas.ResolveString(getStrParam(a.params, "sys_prompt"))
	userPrompt := getStrParam(a.params, "user_prompt")
	if userPrompt == "" {
		userPrompt = "{sys.query}"
	}
	userContent := a.canvas.ResolveString(userPrompt)

	llmInst, err := a.createLLM()
	if err != nil {
		a.mu.Lock()
		a.err = err.Error()
		a.mu.Unlock()
		return err
	}

	msgs := a.buildMessages(sysPrompt, userContent)
	maxRetries := getIntParam(a.params, "max_retries", 1)
	delaySec := getFloatParam(a.params, "delay_after_error", 1)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(delaySec * float64(time.Second)))
		}

		ctx := context.Background()
		var opts []interfaces.CallOption
		if rc := a.canvas.RunContext(); rc != nil && rc.GetStreamCallback() != nil {
			opts = append(opts, interfaces.WithStreamCallback(rc.GetStreamCallback()))
		}
		if t := a.getTemperature(); t >= 0 {
			opts = append(opts, interfaces.WithTemperature(float32(t)))
		}
		if p := getFloatParam(a.params, "top_p", 0); p > 0 && p <= 1 {
			opts = append(opts, interfaces.WithTopP(float32(p)))
		}
		if n := getIntParam(a.params, "max_tokens", 0); n > 0 {
			opts = append(opts, interfaces.WithMaxTokens(n))
		}

		resp, err := llmInst.GenerateContent(ctx, msgs, opts...)
		if err != nil {
			lastErr = err
			continue
		}

		a.mu.Lock()
		if len(resp.Choices) > 0 {
			a.output["content"] = resp.Choices[0].Content
		} else {
			a.output["content"] = ""
		}
		a.mu.Unlock()
		return nil
	}

	a.mu.Lock()
	a.err = lastErr.Error()
	a.mu.Unlock()
	return fmt.Errorf("Agent 重试 %d 次后失败: %w", maxRetries+1, lastErr)
}

func (a *Agent) buildMessages(sysPrompt, userContent string) []interfaces.MessageContent {
	var msgs []interfaces.MessageContent
	if sysPrompt != "" {
		msgs = append(msgs, interfaces.MessageContent{
			Role:  interfaces.MessageRoleSystem,
			Parts: []interfaces.ContentPart{interfaces.TextContent{Text: sysPrompt}},
		})
	}
	if rc := a.canvas.RunContext(); rc != nil && len(rc.GetHistory()) > 0 {
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

func (a *Agent) getTemperature() float64 {
	creativity := getStrParam(a.params, "creativity")
	switch strings.ToLower(creativity) {
	case "improvise":
		return 0.8
	case "balance":
		return 0.5
	case "precise":
		return 0.1
	}
	return getFloatParam(a.params, "temperature", 0.1)
}

func (a *Agent) createLLM() (interfaces.LLM, error) {
	provider, modelName, baseURL, apiKey := resolveComponentLLMConfig(a.canvas, a.params)

	return ragregistry.CreateLLM(ragregistry.LLMConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
}

// Output 获取输出
func (a *Agent) Output(key string) any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.output[key]
}

// OutputAll 获取所有输出
func (a *Agent) OutputAll() map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range a.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (a *Agent) SetOutput(key string, value any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output[key] = value
}

// Error 返回错误
func (a *Agent) Error() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.err
}

// Reset 重置
func (a *Agent) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output = make(map[string]any)
	a.err = ""
}
