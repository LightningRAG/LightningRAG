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

// Anthropic Claude LLM 实现
type Anthropic struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewAnthropic 创建 Anthropic LLM
func NewAnthropic(apiKey, baseURL, model string) *Anthropic {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &Anthropic{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	Messages  []anthropicMsg `json:"messages"`
	Stream    bool           `json:"stream,omitempty"`
	System    string         `json:"system,omitempty"`
	Thinking  map[string]any `json:"thinking,omitempty"` // extended thinking；与 WithReasoningEffort 对齐
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type     string `json:"type"`
		Text     string `json:"text"`
		Thinking string `json:"thinking,omitempty"` // extended thinking 内容块
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type anthropicStreamDelta struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

func (a *Anthropic) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{Model: a.model, MaxTokens: 4096}
	for _, opt := range options {
		opt(opts)
	}
	model := opts.Model
	if model == "" {
		model = a.model
	}
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	var system string
	msgs := make([]anthropicMsg, 0, len(messages))
	for _, m := range messages {
		content := ""
		for _, p := range m.Parts {
			if t, ok := p.(interfaces.TextContent); ok {
				content += t.Text
			}
		}
		if m.Role == interfaces.MessageRoleSystem {
			system = content
			continue
		}
		role := string(m.Role)
		if m.Role == interfaces.MessageRoleTool {
			// Messages API 无独立 tool 角色：并入 user 侧文本（与 Bedrock 路径一致）
			role = "user"
			tid := m.ToolName
			if tid == "" {
				tid = m.ToolCallID
			}
			if tid == "" {
				tid = "result"
			}
			content = fmt.Sprintf("[tool %s output]\n%s", tid, content)
		} else if m.Role == interfaces.MessageRoleAssistant && len(m.ToolCalls) > 0 {
			var lines []string
			for _, tc := range m.ToolCalls {
				lines = append(lines, fmt.Sprintf("[tool %s] %s", tc.Name, tc.Arguments))
			}
			content = strings.TrimSpace(content + "\n" + strings.Join(lines, "\n"))
		}
		if role == "user" {
			role = "user"
		} else if role == "assistant" {
			role = "assistant"
		}
		msgs = append(msgs, anthropicMsg{Role: role, Content: content})
	}

	stream := opts.Stream && opts.StreamCallback != nil
	var thinkBlock map[string]any
	if tb := anthropicThinkingBlockFromReasoningEffort(opts.ReasoningEffort, &maxTokens); tb != nil {
		thinkBlock = tb
	}
	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  msgs,
		Stream:    stream,
		System:    system,
	}
	if thinkBlock != nil {
		reqBody.Thinking = thinkBlock
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic api error: %s", string(b))
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
			for strings.HasPrefix(data, "data: ") {
				data = strings.TrimPrefix(data, "data: ")
			}
			if data == "[DONE]" || data == "[done]" {
				break
			}
			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type       string `json:"type"`
					Text       string `json:"text"`
					Thinking   string `json:"thinking"`
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			d := event.Delta
			// message_delta：stop_reason 为 max_tokens 时表示输出被截断（与 OpenAI finish_reason length 提示对齐）
			if event.Type == "message_delta" && d.StopReason == "max_tokens" && !lengthNotified {
				lengthNotified = true
				if reasoningStart {
					reasoningStart = false
					fullContent.WriteString(AssistantReasoningCloseTag)
					opts.StreamCallback(AssistantReasoningCloseTag)
				}
				notice := AssistantOutputTruncationNotice
				fullContent.WriteString(notice)
				opts.StreamCallback(notice)
			}
			// Extended thinking：SSE delta.thinking（官方多为 type thinking_delta；以字段为准避免漏解析）
			if d.Thinking != "" {
				out := ""
				if !reasoningStart {
					reasoningStart = true
					out = AssistantReasoningOpenTag
				}
				out += d.Thinking
				fullContent.WriteString(out)
				opts.StreamCallback(out)
			} else if d.Text != "" {
				if reasoningStart {
					reasoningStart = false
					fullContent.WriteString(AssistantReasoningCloseTag)
					opts.StreamCallback(AssistantReasoningCloseTag)
				}
				fullContent.WriteString(d.Text)
				opts.StreamCallback(d.Text)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("anthropic stream: %w", err)
		}
		if reasoningStart {
			fullContent.WriteString(AssistantReasoningCloseTag)
			opts.StreamCallback(AssistantReasoningCloseTag)
		}
		return &interfaces.ContentResponse{
			Choices: []interfaces.Choice{{Content: fullContent.String()}},
		}, nil
	}

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Content) == 0 {
		return nil, errors.New("empty response from anthropic")
	}

	var thinkBuf strings.Builder
	text := ""
	for _, c := range result.Content {
		switch c.Type {
		case "thinking":
			thinkBuf.WriteString(c.Thinking)
		case "text":
			text += c.Text
		default:
			if c.Text != "" {
				text += c.Text
			}
		}
	}
	out := text
	if ts := strings.TrimSpace(thinkBuf.String()); ts != "" {
		out = AssistantReasoningOpenTag + thinkBuf.String() + AssistantReasoningCloseTag + text
	}
	if result.StopReason == "max_tokens" {
		out += AssistantOutputTruncationNotice
	}

	res := &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: out}},
		Usage: &interfaces.Usage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}
	return res, nil
}

func (a *Anthropic) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
	msg := interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: prompt}},
	}
	resp, err := a.GenerateContent(ctx, []interfaces.MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty response")
	}
	return resp.Choices[0].Content, nil
}

func (a *Anthropic) ProviderName() string { return "anthropic" }
func (a *Anthropic) ModelName() string    { return a.model }
