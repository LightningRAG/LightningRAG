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
	"time"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// Replicate 通过 Replicate HTTP API 调用社区模型（与上游参考 ReplicateChat 输入习惯一致：prompt + system_prompt）
type Replicate struct {
	token   string
	model   string // owner/name，如 meta/llama-3-8b-instruct
	client  *http.Client
	baseAPI string
}

// NewReplicate model 为 Replicate 的 models 路径，如 meta/llama-3-8b-instruct
func NewReplicate(token, model string) *Replicate {
	return &Replicate{
		token:   strings.TrimSpace(token),
		model:   strings.TrimSpace(model),
		client:  &http.Client{Timeout: 180 * time.Second},
		baseAPI: "https://api.replicate.com/v1",
	}
}

func (r *Replicate) predictionURL() string {
	return fmt.Sprintf("%s/models/%s/predictions", r.baseAPI, strings.Trim(r.model, "/"))
}

type replicatePrediction struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

func (r *Replicate) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{}
	for _, o := range options {
		o(opts)
	}
	if len(opts.Tools) > 0 {
		return nil, errors.New("replicate: 当前实现不支持工具调用")
	}
	system, prompt := flattenReplicateMessages(messages)
	input := map[string]any{
		"prompt":         prompt,
		"system_prompt":  system,
		"max_new_tokens": opts.MaxTokens,
	}
	if opts.MaxTokens <= 0 {
		delete(input, "max_new_tokens")
	}
	if opts.Temperature > 0 {
		input["temperature"] = opts.Temperature
	}
	body, _ := json.Marshal(map[string]any{"input": input})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.predictionURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait=120")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("replicate: HTTP %d %s", resp.StatusCode, string(respBody))
	}
	var pred replicatePrediction
	if err := json.Unmarshal(respBody, &pred); err != nil {
		return nil, fmt.Errorf("replicate: 解析响应: %w", err)
	}
	out, err := r.resolveOutput(ctx, &pred)
	if err != nil {
		return nil, err
	}
	return &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: out}},
	}, nil
}

func (r *Replicate) resolveOutput(ctx context.Context, pred *replicatePrediction) (string, error) {
	p := pred
	for range 90 {
		switch p.Status {
		case "succeeded":
			return decodeReplicateOutput(p.Output)
		case "failed", "canceled":
			if p.Error != "" {
				return "", fmt.Errorf("replicate: %s", p.Error)
			}
			return "", fmt.Errorf("replicate: 状态 %s", p.Status)
		case "starting", "processing":
			if p.URLs.Get == "" {
				return "", fmt.Errorf("replicate: 异步任务缺少轮询 URL")
			}
			time.Sleep(2 * time.Second)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URLs.Get, nil)
			if err != nil {
				return "", err
			}
			req.Header.Set("Authorization", "Bearer "+r.token)
			resp, err := r.client.Do(req)
			if err != nil {
				return "", err
			}
			b, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return "", err
			}
			if err := json.Unmarshal(b, &p); err != nil {
				return "", err
			}
			continue
		default:
			if len(p.Output) > 0 && string(p.Output) != "null" {
				return decodeReplicateOutput(p.Output)
			}
			if p.URLs.Get != "" {
				time.Sleep(2 * time.Second)
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URLs.Get, nil)
				if err != nil {
					return "", err
				}
				req.Header.Set("Authorization", "Bearer "+r.token)
				resp, err := r.client.Do(req)
				if err != nil {
					return "", err
				}
				b, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					return "", err
				}
				if err := json.Unmarshal(b, &p); err != nil {
					return "", err
				}
				continue
			}
			return "", fmt.Errorf("replicate: 未预期状态 %q", p.Status)
		}
	}
	return "", fmt.Errorf("replicate: 等待结果超时")
}

func decodeReplicateOutput(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", errors.New("replicate: 空输出")
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var ss []string
	if err := json.Unmarshal(raw, &ss); err == nil {
		return strings.Join(ss, ""), nil
	}
	return string(raw), nil
}

func flattenReplicateMessages(messages []interfaces.MessageContent) (system, prompt string) {
	var lines []string
	for _, m := range messages {
		var text string
		for _, p := range m.Parts {
			if t, ok := p.(interfaces.TextContent); ok {
				text += t.Text
			}
		}
		switch m.Role {
		case interfaces.MessageRoleSystem:
			if text != "" {
				system = text
			}
		case interfaces.MessageRoleHuman:
			lines = append(lines, "User: "+text)
		case interfaces.MessageRoleAssistant:
			if strings.TrimSpace(text) != "" {
				lines = append(lines, "Assistant: "+text)
			}
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					lines = append(lines, fmt.Sprintf("Assistant tool %s(%s)", tc.Name, tc.Arguments))
				}
			}
			if strings.TrimSpace(text) == "" && len(m.ToolCalls) == 0 {
				lines = append(lines, "Assistant: ")
			}
		case interfaces.MessageRoleTool:
			toolID := m.ToolName
			if toolID == "" {
				toolID = m.ToolCallID
			}
			if toolID == "" {
				toolID = "result"
			}
			lines = append(lines, fmt.Sprintf("Tool(%s): %s", toolID, text))
		}
	}
	return system, strings.Join(lines, "\n")
}

func (r *Replicate) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
	msg := interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: prompt}},
	}
	out, err := r.GenerateContent(ctx, []interfaces.MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("replicate: 空响应")
	}
	return out.Choices[0].Content, nil
}

func (r *Replicate) ProviderName() string { return "replicate" }
func (r *Replicate) ModelName() string    { return r.model }
