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

// Azure OpenAI LLM 实现，使用 Azure 专用 endpoint 格式
type Azure struct {
	apiKey     string
	baseURL    string
	model      string // deployment name
	apiVersion string
	client     *http.Client
}

// NewAzure 创建 Azure OpenAI LLM
func NewAzure(apiKey, baseURL, model, apiVersion string) *Azure {
	if baseURL == "" {
		baseURL = "https://YOUR_RESOURCE.openai.azure.com"
	}
	if apiVersion == "" {
		apiVersion = "2024-02-01"
	}
	return &Azure{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		apiVersion: apiVersion,
		client:     &http.Client{},
	}
}

func (a *Azure) chatURL() string {
	return fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		a.baseURL, a.model, a.apiVersion)
}

func (a *Azure) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{Model: a.model}
	for _, opt := range options {
		opt(opts)
	}
	model := opts.Model
	if model == "" {
		model = a.model
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

	req, err := http.NewRequestWithContext(ctx, "POST", a.chatURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure openai error: %s", string(b))
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
			if !lengthNotified && chunk.Choices[0].FinishReason == "length" {
				lengthNotified = true
				notice := AssistantOutputTruncationNotice
				fullContent.WriteString(notice)
				opts.StreamCallback(notice)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("azure openai stream: %w", err)
		}
		return &interfaces.ContentResponse{
			Choices: []interfaces.Choice{{Content: fullContent.String()}},
		}, nil
	}

	// 非流式：部分部署可能返回 SSE 格式，需兼容（参考 openai.go）
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyStr := string(bodyBytes)
	if strings.HasPrefix(bodyStr, "data: ") {
		var lastMsg string
		reasoningStart := false
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
		return nil, errors.New("empty response from azure openai")
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

func (a *Azure) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
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

func (a *Azure) ProviderName() string { return "azure" }
func (a *Azure) ModelName() string    { return a.model }
