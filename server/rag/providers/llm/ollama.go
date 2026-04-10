package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// Ollama LLM 实现，支持本地部署
type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllama 创建 Ollama LLM
func NewOllama(baseURL, model string) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type ollamaRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaMsg            `json:"messages"`
	Stream   bool                   `json:"stream"`
	Think    any                    `json:"think,omitempty"` // 与 WithReasoningEffort 对齐；支持 thinking 类模型
	Tools    []map[string]any       `json:"tools,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ollamaMsg struct {
	Role      string           `json:"role"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
	ToolName  string           `json:"tool_name,omitempty"`
}

type ollamaToolCall struct {
	Type     string `json:"type"`
	Function struct {
		Index     int    `json:"index,omitempty"`
		Name      string `json:"name"`
		Arguments any    `json:"arguments"` // 对象或字符串
	} `json:"function"`
}

type ollamaResponse struct {
	Message struct {
		Content   string           `json:"content"`
		Thinking  string           `json:"thinking,omitempty"` // Ollama thinking 模型；与 OpenAI reasoning_content 一样包装为 think 标签供前端解析
		ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
	} `json:"message"`
	Done       bool   `json:"done"`
	DoneReason string `json:"done_reason,omitempty"` // stop | length | …；length 时追加与 OpenAI 一致的截断提示
}

func (o *Ollama) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{Model: o.model}
	for _, opt := range options {
		opt(opts)
	}
	model := opts.Model
	if model == "" {
		model = o.model
	}

	msgs := make([]ollamaMsg, 0, len(messages))
	for _, m := range messages {
		om := ollamaMsg{Role: string(m.Role)}
		if m.Role == interfaces.MessageRoleTool {
			om.ToolName = m.ToolName
			if om.ToolName == "" {
				om.ToolName = m.ToolCallID
			}
			for _, p := range m.Parts {
				if t, ok := p.(interfaces.TextContent); ok {
					om.Content += t.Text
				}
			}
		} else if len(m.ToolCalls) > 0 {
			for _, p := range m.Parts {
				if t, ok := p.(interfaces.TextContent); ok {
					om.Content += t.Text
				}
			}
			om.ToolCalls = make([]ollamaToolCall, 0, len(m.ToolCalls))
			for i, tc := range m.ToolCalls {
				var args any = tc.Arguments
				if tc.Arguments != "" {
					var obj map[string]any
					if json.Unmarshal([]byte(tc.Arguments), &obj) == nil {
						args = obj
					}
				}
				om.ToolCalls = append(om.ToolCalls, ollamaToolCall{
					Type: "function",
					Function: struct {
						Index     int    `json:"index,omitempty"`
						Name      string `json:"name"`
						Arguments any    `json:"arguments"`
					}{Index: i, Name: tc.Name, Arguments: args},
				})
			}
		} else {
			for _, p := range m.Parts {
				if t, ok := p.(interfaces.TextContent); ok {
					om.Content += t.Text
				}
			}
		}
		msgs = append(msgs, om)
	}

	stream := opts.Stream && opts.StreamCallback != nil
	reqBody := ollamaRequest{
		Model:    model,
		Messages: msgs,
		Stream:   stream,
	}
	if len(opts.Tools) > 0 {
		reqBody.Tools = opts.Tools
		if stream {
			stream = false
			reqBody.Stream = false
		}
	}
	if v, ok := ollamaThinkFromReasoningEffort(opts.ReasoningEffort); ok {
		reqBody.Think = v
	}
	if opts.Temperature > 0 || opts.TopP > 0 || opts.MaxTokens > 0 {
		reqBody.Options = make(map[string]interface{})
		if opts.Temperature > 0 {
			reqBody.Options["temperature"] = opts.Temperature
		}
		if opts.TopP > 0 {
			reqBody.Options["top_p"] = opts.TopP
		}
		if opts.MaxTokens > 0 {
			reqBody.Options["num_predict"] = opts.MaxTokens
		}
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama api error: %s", string(b))
	}

	if stream {
		var fullContent strings.Builder
		reasoningStart := false
		lengthNotified := false
		scanner := newStreamLineScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			line := scanner.Text()
			if line == "" {
				continue
			}
			var chunk ollamaResponse
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}
			if th := chunk.Message.Thinking; th != "" {
				out := ""
				if !reasoningStart {
					reasoningStart = true
					out = AssistantReasoningOpenTag
				}
				out += th
				fullContent.WriteString(out)
				opts.StreamCallback(out)
			}
			if co := chunk.Message.Content; co != "" {
				if reasoningStart {
					reasoningStart = false
					fullContent.WriteString(AssistantReasoningCloseTag)
					opts.StreamCallback(AssistantReasoningCloseTag)
				}
				fullContent.WriteString(co)
				opts.StreamCallback(co)
			}
			if chunk.Done {
				if reasoningStart {
					reasoningStart = false
					fullContent.WriteString(AssistantReasoningCloseTag)
					opts.StreamCallback(AssistantReasoningCloseTag)
				}
				if !lengthNotified && chunk.DoneReason == "length" {
					lengthNotified = true
					fullContent.WriteString(AssistantOutputTruncationNotice)
					opts.StreamCallback(AssistantOutputTruncationNotice)
				}
				break
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("ollama stream: %w", err)
		}
		return &interfaces.ContentResponse{
			Choices: []interfaces.Choice{{Content: fullContent.String()}},
		}, nil
	}

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := result.Message.Content
	if tk := strings.TrimSpace(result.Message.Thinking); tk != "" {
		content = AssistantReasoningOpenTag + tk + AssistantReasoningCloseTag + content
	}
	if result.DoneReason == "length" {
		content += AssistantOutputTruncationNotice
	}
	res := &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: content}},
	}
	if len(result.Message.ToolCalls) > 0 {
		tcs := make([]interfaces.ToolCall, 0, len(result.Message.ToolCalls))
		for _, tc := range result.Message.ToolCalls {
			args := ""
			switch v := tc.Function.Arguments.(type) {
			case string:
				args = v
			default:
				if b, err := json.Marshal(tc.Function.Arguments); err == nil {
					args = string(b)
				}
			}
			tcs = append(tcs, interfaces.ToolCall{
				ID:        tc.Function.Name,
				Name:      tc.Function.Name,
				Arguments: args,
			})
		}
		res.Choices[0].ToolCalls = tcs
	}
	return res, nil
}

func (o *Ollama) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
	msg := interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: prompt}},
	}
	resp, err := o.GenerateContent(ctx, []interfaces.MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty response")
	}
	return resp.Choices[0].Content, nil
}

func (o *Ollama) ProviderName() string { return "ollama" }
func (o *Ollama) ModelName() string    { return o.model }
