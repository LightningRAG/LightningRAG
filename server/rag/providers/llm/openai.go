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

// OpenAI LLM 实现
type OpenAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAI 创建 OpenAI LLM
func NewOpenAI(apiKey, baseURL, model string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAI{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

// openaiRequest OpenAI API 请求结构
type openaiRequest struct {
	Model           string           `json:"model"`
	Messages        []openaiMsg      `json:"messages"`
	Stream          bool             `json:"stream,omitempty"`
	Tools           []map[string]any `json:"tools,omitempty"`
	ToolChoice      string           `json:"tool_choice,omitempty"`      // "auto" | "none"
	ReasoningEffort string           `json:"reasoning_effort,omitempty"` // "low"|"medium"|"high"，仅 o1 等推理模型支持
}

type openaiMsg struct {
	Role       string           `json:"role"`
	Content    interface{}      `json:"content"` // string or null
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Index    int    `json:"index"` // 流式响应中标识 tool_call 序号
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openaiResponse struct {
	Choices []struct {
		Delta struct {
			Content          string                      `json:"content"`
			ReasoningContent flexOpenAIReasoningFragment `json:"reasoning_content,omitempty"`
			Reasoning        flexOpenAIReasoningFragment `json:"reasoning,omitempty"`
			Thought          flexOpenAIReasoningFragment `json:"thought,omitempty"`
			Thinking         flexOpenAIReasoningFragment `json:"thinking,omitempty"`
			ToolCalls        []openaiToolCall            `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
		Message      struct {
			Content          string                      `json:"content"`
			ReasoningContent flexOpenAIReasoningFragment `json:"reasoning_content,omitempty"`
			Reasoning        flexOpenAIReasoningFragment `json:"reasoning,omitempty"`
			Thought          flexOpenAIReasoningFragment `json:"thought,omitempty"`
			Thinking         flexOpenAIReasoningFragment `json:"thinking,omitempty"`
			ToolCalls        []openaiToolCall            `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// pickOpenAIReasoning 按顺序取第一个非空的推理相关字段（reasoning_content / reasoning / thought / thinking 等）。
func pickOpenAIReasoning(candidates ...string) string {
	for _, s := range candidates {
		if s != "" {
			return s
		}
	}
	return ""
}

// messageContentsToOpenAIMessages 将内部消息转为 OpenAI Chat Completions 的 messages（Azure 与 OpenAI 共用）。
func messageContentsToOpenAIMessages(messages []interfaces.MessageContent) []openaiMsg {
	msgs := make([]openaiMsg, 0, len(messages))
	for _, m := range messages {
		om := openaiMsg{Role: string(m.Role)}
		if m.Role == interfaces.MessageRoleTool {
			om.ToolCallID = m.ToolCallID
			if om.ToolCallID == "" {
				om.ToolCallID = m.ToolName
			}
			content := ""
			for _, p := range m.Parts {
				if t, ok := p.(interfaces.TextContent); ok {
					content += t.Text
				}
			}
			om.Content = content
		} else if len(m.ToolCalls) > 0 {
			toolText := ""
			for _, p := range m.Parts {
				if t, ok := p.(interfaces.TextContent); ok {
					toolText += t.Text
				}
			}
			if strings.TrimSpace(toolText) != "" {
				om.Content = toolText
			} else {
				om.Content = nil
			}
			om.ToolCalls = make([]openaiToolCall, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				om.ToolCalls = append(om.ToolCalls, openaiToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{Name: tc.Name, Arguments: tc.Arguments},
				})
			}
		} else {
			content := ""
			for _, p := range m.Parts {
				if t, ok := p.(interfaces.TextContent); ok {
					content += t.Text
				}
			}
			om.Content = content
		}
		msgs = append(msgs, om)
	}
	return msgs
}

func (o *OpenAI) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{Model: o.model}
	for _, opt := range options {
		opt(opts)
	}
	model := opts.Model
	if model == "" {
		model = o.model
	}

	msgs := messageContentsToOpenAIMessages(messages)

	stream := opts.Stream && opts.StreamCallback != nil
	reqBody := openaiRequest{
		Model:    model,
		Messages: msgs,
		Stream:   stream,
	}
	if len(opts.Tools) > 0 {
		reqBody.Tools = opts.Tools
		reqBody.ToolChoice = "auto"
		if stream {
			stream = false
			reqBody.Stream = false
		}
	}
	if opts.ReasoningEffort != "" {
		reqBody.ReasoningEffort = opts.ReasoningEffort
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error: %s", string(b))
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
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			// 兼容部分 API（如 DeepSeek）返回的 "data: data: {...}" 双重前缀
			for strings.HasPrefix(data, "data: ") {
				data = strings.TrimPrefix(data, "data: ")
			}
			if data == "[DONE]" {
				if reasoningStart {
					fullContent.WriteString(AssistantReasoningCloseTag)
					opts.StreamCallback(AssistantReasoningCloseTag)
				}
				break
			}
			var chunk openaiResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta
			// DeepSeek R1 等推理模型：优先输出 reasoning_content（思考链）
			if rText := pickOpenAIReasoning(string(delta.ReasoningContent), string(delta.Reasoning), string(delta.Thought), string(delta.Thinking)); rText != "" {
				out := ""
				if !reasoningStart {
					reasoningStart = true
					out = AssistantReasoningOpenTag
				}
				out += rText
				fullContent.WriteString(out)
				opts.StreamCallback(out)
			} else if delta.Content != "" {
				if reasoningStart {
					reasoningStart = false
					fullContent.WriteString(AssistantReasoningCloseTag)
					opts.StreamCallback(AssistantReasoningCloseTag)
				}
				fullContent.WriteString(delta.Content)
				opts.StreamCallback(delta.Content)
			}
			// 上下文截断时追加提示（与 references 目录内约定一致）
			if !lengthNotified && chunk.Choices[0].FinishReason == "length" {
				lengthNotified = true
				notice := AssistantOutputTruncationNotice
				fullContent.WriteString(notice)
				opts.StreamCallback(notice)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("openai stream: %w", err)
		}
		return &interfaces.ContentResponse{
			Choices: []interfaces.Choice{{Content: fullContent.String()}},
		}, nil
	}

	// 非流式：部分 API（如 DeepSeek）在请求 stream=false 时仍可能返回 SSE 格式，需兼容
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyStr := string(bodyBytes)
	// 若响应以 "data: " 开头，按 SSE 解析并取最后一条完整消息；同时收集 tool_calls（部分 API 如 DeepSeek 在 stream=false 时仍返回 SSE）
	if strings.HasPrefix(bodyStr, "data: ") {
		var lastMsg string
		reasoningStart := false
		// 按 index 累积 tool_calls（流式时 arguments 可能分多块）
		accums := make(map[int]*struct {
			id   string
			name string
			args strings.Builder
		})
		for _, line := range strings.Split(bodyStr, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			for strings.HasPrefix(data, "data: ") {
				data = strings.TrimPrefix(data, "data: ")
			}
			if data == "[DONE]" {
				if reasoningStart {
					lastMsg += AssistantReasoningCloseTag
					reasoningStart = false
				}
				break
			}
			var chunk openaiResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta
			msg := chunk.Choices[0].Message
			if c := pickOpenAIReasoning(string(delta.ReasoningContent), string(delta.Reasoning), string(delta.Thought), string(delta.Thinking)); c != "" {
				if !reasoningStart {
					reasoningStart = true
					lastMsg += AssistantReasoningOpenTag
				}
				lastMsg += c
			} else if c := delta.Content; c != "" {
				if reasoningStart {
					reasoningStart = false
					lastMsg += AssistantReasoningCloseTag
				}
				lastMsg += c
			} else if c := pickOpenAIReasoning(string(msg.ReasoningContent), string(msg.Reasoning), string(msg.Thought), string(msg.Thinking)); c != "" {
				if !reasoningStart {
					reasoningStart = true
					lastMsg += AssistantReasoningOpenTag
				}
				lastMsg += c
			} else if c := msg.Content; c != "" {
				if reasoningStart {
					reasoningStart = false
					lastMsg += AssistantReasoningCloseTag
				}
				lastMsg += c
			}
			for _, tc := range delta.ToolCalls {
				idx := tc.Index
				if idx < 0 {
					idx = 0
				}
				if accums[idx] == nil {
					accums[idx] = &struct {
						id   string
						name string
						args strings.Builder
					}{}
				}
				a := accums[idx]
				if tc.ID != "" {
					a.id = tc.ID
				}
				if tc.Function.Name != "" {
					a.name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					a.args.WriteString(tc.Function.Arguments)
				}
			}
			for _, tc := range msg.ToolCalls {
				idx := tc.Index
				if idx < 0 {
					idx = 0
				}
				if accums[idx] == nil {
					accums[idx] = &struct {
						id   string
						name string
						args strings.Builder
					}{}
				}
				a := accums[idx]
				if tc.ID != "" {
					a.id = tc.ID
				}
				if tc.Function.Name != "" {
					a.name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					a.args.WriteString(tc.Function.Arguments)
				}
			}
		}
		if reasoningStart {
			lastMsg += AssistantReasoningCloseTag
		}
		res := &interfaces.ContentResponse{
			Choices: []interfaces.Choice{{Content: lastMsg}},
		}
		if len(accums) > 0 {
			// 按 index 排序后转为 ToolCalls
			maxIdx := -1
			for k := range accums {
				if k > maxIdx {
					maxIdx = k
				}
			}
			tcs := make([]interfaces.ToolCall, 0, maxIdx+1)
			for i := 0; i <= maxIdx; i++ {
				if a, ok := accums[i]; ok && (a.id != "" || a.name != "") {
					tcs = append(tcs, interfaces.ToolCall{
						ID:        a.id,
						Name:      a.name,
						Arguments: a.args.String(),
					})
				}
			}
			if len(tcs) > 0 {
				res.Choices[0].ToolCalls = tcs
			}
		}
		return res, nil
	}

	var result openaiResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("empty response from openai")
	}

	choice := result.Choices[0]
	content := choice.Message.Content
	if rc := strings.TrimSpace(pickOpenAIReasoning(string(choice.Message.ReasoningContent), string(choice.Message.Reasoning), string(choice.Message.Thought), string(choice.Message.Thinking))); rc != "" {
		content = AssistantReasoningOpenTag + rc + AssistantReasoningCloseTag + content
	}
	res := &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: content}},
	}
	if len(choice.Message.ToolCalls) > 0 {
		tcs := make([]interfaces.ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			tcs = append(tcs, interfaces.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
		res.Choices[0].ToolCalls = tcs
	}
	if result.Usage != nil {
		res.Usage = &interfaces.Usage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		}
	}
	return res, nil
}

func (o *OpenAI) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
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

func (o *OpenAI) ProviderName() string { return "openai" }
func (o *OpenAI) ModelName() string    { return o.model }

// ProviderNameAdapter 包装 LLM 以自定义 ProviderName，用于同一实现多厂商注册
type ProviderNameAdapter struct {
	interfaces.LLM
	DisplayName string
}

func (p *ProviderNameAdapter) ProviderName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.LLM.ProviderName()
}
