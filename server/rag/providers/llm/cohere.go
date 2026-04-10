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

// Cohere LLM 实现，对齐 Cohere API v2 与 references 目录内对应实现
type Cohere struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewCohere 创建 Cohere LLM
func NewCohere(apiKey, baseURL, model string) *Cohere {
	if baseURL == "" {
		baseURL = "https://api.cohere.ai"
	}
	if model == "" {
		model = "command-r-plus"
	}
	return &Cohere{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type cohereMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type cohereRequest struct {
	Model    string         `json:"model"`
	Messages []cohereMsg    `json:"messages"`
	Stream   bool           `json:"stream,omitempty"`
	Thinking map[string]any `json:"thinking,omitempty"` // reasoning 模型；与 WithReasoningEffort 对齐
}

type cohereContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Thinking string `json:"thinking,omitempty"` // command-a-reasoning 等模型的思考块
}

type cohereMessage struct {
	Role    string               `json:"role"`
	Content []cohereContentBlock `json:"content"`
}

type cohereResponse struct {
	Message struct {
		Content []cohereContentBlock `json:"content"`
	} `json:"message"`
}

type cohereStreamDelta struct {
	Message *struct {
		Content *struct {
			Text     string `json:"text,omitempty"`
			Thinking string `json:"thinking,omitempty"`
		} `json:"content,omitempty"`
	} `json:"message,omitempty"`
}

type cohereStreamEvent struct {
	Type    string             `json:"type"`
	Text    string             `json:"text,omitempty"`
	Delta   *cohereStreamDelta `json:"delta,omitempty"`
	Message *struct {
		Content []cohereContentBlock `json:"content"`
	} `json:"message,omitempty"`
}

// cohereStripSSEDataLine 解析 Cohere Chat 流式的 text/event-stream 行（data: {...}）；纯 JSON 行亦兼容。
func cohereStripSSEDataLine(line string) (payload string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, ":") {
		return "", false
	}
	if strings.HasPrefix(line, "event:") {
		return "", false
	}
	for strings.HasPrefix(line, "data:") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	}
	if line == "" || line == "[DONE]" {
		return "", false
	}
	return line, true
}

// cohereEmitStreamThinkingAndText 处理 Cohere v2 content-delta / message 中的 thinking 与 text，与前端 think 标签约定一致。
func cohereEmitStreamThinkingAndText(reasoningStart *bool, full *strings.Builder, cb func(string), thinking, text string) {
	if thinking != "" {
		out := ""
		if !*reasoningStart {
			*reasoningStart = true
			out = AssistantReasoningOpenTag
		}
		out += thinking
		full.WriteString(out)
		cb(out)
	}
	if text != "" {
		if *reasoningStart {
			*reasoningStart = false
			full.WriteString(AssistantReasoningCloseTag)
			cb(AssistantReasoningCloseTag)
		}
		full.WriteString(text)
		cb(text)
	}
}

func cohereBuildAssistantContent(blocks []cohereContentBlock) string {
	var out strings.Builder
	reasoningStart := false
	closeR := func() {
		if reasoningStart {
			reasoningStart = false
			out.WriteString(AssistantReasoningCloseTag)
		}
	}
	for _, block := range blocks {
		switch block.Type {
		case "thinking":
			t := block.Thinking
			if strings.TrimSpace(t) == "" {
				continue
			}
			if !reasoningStart {
				reasoningStart = true
				out.WriteString(AssistantReasoningOpenTag)
			}
			out.WriteString(t)
		case "text":
			closeR()
			out.WriteString(block.Text)
		default:
			if block.Text != "" {
				closeR()
				out.WriteString(block.Text)
			}
		}
	}
	closeR()
	return out.String()
}

// cohereThinkingFromReasoningEffort 将 CallOptions.ReasoningEffort 转为 Cohere v2 的 thinking 对象；无需覆盖默认时不返回 nil。
func cohereThinkingFromReasoningEffort(effort string) map[string]any {
	e := strings.ToLower(strings.TrimSpace(effort))
	switch e {
	case "":
		return nil
	case "disabled", "none", "off":
		return map[string]any{"type": "disabled"}
	case "low":
		return map[string]any{"type": "enabled", "token_budget": 2048}
	case "medium":
		return map[string]any{"type": "enabled", "token_budget": 10240}
	case "high":
		// 与 Cohere 文档建议一致：不限定 budget 时由模型充分推理
		return map[string]any{"type": "enabled"}
	default:
		return map[string]any{"type": "enabled"}
	}
}

func (c *Cohere) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{Model: c.model}
	for _, opt := range options {
		opt(opts)
	}
	model := opts.Model
	if model == "" {
		model = c.model
	}

	msgs := make([]cohereMsg, 0, len(messages))
	for _, m := range messages {
		content := ""
		for _, p := range m.Parts {
			if t, ok := p.(interfaces.TextContent); ok {
				content += t.Text
			}
		}
		role := string(m.Role)
		if m.Role == interfaces.MessageRoleTool {
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
		msgs = append(msgs, cohereMsg{
			Role:    role,
			Content: content,
		})
	}

	stream := opts.Stream && opts.StreamCallback != nil
	reqBody := cohereRequest{
		Model:    model,
		Messages: msgs,
		Stream:   stream,
	}
	if th := cohereThinkingFromReasoningEffort(opts.ReasoningEffort); th != nil {
		reqBody.Thinking = th
	}
	body, _ := json.Marshal(reqBody)

	url := c.baseURL + "/v2/chat"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cohere api error: %s", string(b))
	}

	if stream {
		var fullContent strings.Builder
		reasoningStart := false
		scanner := newStreamLineScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			payload, ok := cohereStripSSEDataLine(scanner.Text())
			if !ok {
				continue
			}
			var event cohereStreamEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				continue
			}
			if event.Type == "content-delta" && event.Delta != nil && event.Delta.Message != nil && event.Delta.Message.Content != nil {
				c := event.Delta.Message.Content
				cohereEmitStreamThinkingAndText(&reasoningStart, &fullContent, opts.StreamCallback, c.Thinking, c.Text)
				continue
			}
			if event.Type == "message" && event.Message != nil {
				for _, block := range event.Message.Content {
					switch block.Type {
					case "thinking":
						cohereEmitStreamThinkingAndText(&reasoningStart, &fullContent, opts.StreamCallback, block.Thinking, "")
					case "text":
						cohereEmitStreamThinkingAndText(&reasoningStart, &fullContent, opts.StreamCallback, "", block.Text)
					default:
						if block.Text != "" {
							cohereEmitStreamThinkingAndText(&reasoningStart, &fullContent, opts.StreamCallback, "", block.Text)
						}
					}
				}
				continue
			}
			if event.Text != "" {
				cohereEmitStreamThinkingAndText(&reasoningStart, &fullContent, opts.StreamCallback, "", event.Text)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("cohere stream: %w", err)
		}
		if reasoningStart {
			fullContent.WriteString(AssistantReasoningCloseTag)
			opts.StreamCallback(AssistantReasoningCloseTag)
		}
		return &interfaces.ContentResponse{
			Choices: []interfaces.Choice{{Content: fullContent.String()}},
		}, nil
	}

	var result cohereResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	out := cohereBuildAssistantContent(result.Message.Content)
	if strings.TrimSpace(out) == "" {
		return nil, errors.New("empty response from cohere")
	}

	return &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: out}},
	}, nil
}

func (c *Cohere) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
	msg := interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: prompt}},
	}
	resp, err := c.GenerateContent(ctx, []interfaces.MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty response")
	}
	return resp.Choices[0].Content, nil
}

func (c *Cohere) ProviderName() string { return "cohere" }
func (c *Cohere) ModelName() string    { return c.model }
