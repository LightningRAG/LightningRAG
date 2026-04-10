package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// Bedrock 使用 Amazon Bedrock Converse / ConverseStream API
type Bedrock struct {
	client  *bedrockruntime.Client
	modelID string
}

type bedrockKeyJSON struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	SessionToken    string `json:"aws_session_token"`
	Region          string `json:"aws_region"`
}

// NewBedrock apiKey 可为空（使用环境变量/实例角色），或为 JSON 静态凭证；region 为空时尝试 extra["aws_region"] 或 us-east-1
func NewBedrock(apiKey, region, modelID string, extra map[string]any) (*Bedrock, error) {
	if strings.TrimSpace(modelID) == "" {
		return nil, errors.New("bedrock: 需要配置模型 ID（如 anthropic.claude-3-5-sonnet-20240620-v1:0）")
	}
	cfg, err := loadBedrockAWSConfig(apiKey, region, extra)
	if err != nil {
		return nil, err
	}
	return &Bedrock{
		client:  bedrockruntime.NewFromConfig(cfg),
		modelID: modelID,
	}, nil
}

func loadBedrockAWSConfig(apiKey, region string, extra map[string]any) (aws.Config, error) {
	if extra != nil {
		if r, ok := extra["aws_region"].(string); ok && strings.TrimSpace(r) != "" {
			region = strings.TrimSpace(r)
		}
	}
	if strings.TrimSpace(region) == "" {
		region = "us-east-1"
	}
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(region))

	apiKey = strings.TrimSpace(apiKey)
	if apiKey != "" && (apiKey[0] == '{' || strings.Contains(apiKey, "aws_access_key_id")) {
		var j bedrockKeyJSON
		if err := json.Unmarshal([]byte(apiKey), &j); err != nil {
			return aws.Config{}, fmt.Errorf("bedrock: 解析 aws 凭证 JSON: %w", err)
		}
		if j.Region != "" {
			opts = append(opts, config.WithRegion(j.Region))
		}
		if j.AccessKeyID != "" && j.SecretAccessKey != "" {
			opts = append(opts, config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(j.AccessKeyID, j.SecretAccessKey, j.SessionToken),
			))
		}
	}

	return config.LoadDefaultConfig(context.Background(), opts...)
}

func (b *Bedrock) GenerateContent(ctx context.Context, messages []interfaces.MessageContent, options ...interfaces.CallOption) (*interfaces.ContentResponse, error) {
	opts := &interfaces.CallOptions{Model: b.modelID}
	for _, o := range options {
		o(opts)
	}
	if len(opts.Tools) > 0 {
		return nil, errors.New("bedrock: 当前实现不支持工具调用，请改用无工具流程")
	}
	modelID := opts.Model
	if modelID == "" {
		modelID = b.modelID
	}

	sys, msgs, err := toBedrockMessages(messages)
	if err != nil {
		return nil, err
	}

	mt := opts.MaxTokens
	var thinkBlock map[string]any
	if bedrockModelSupportsAnthropicThinking(modelID) {
		if tb := anthropicThinkingBlockFromReasoningEffort(opts.ReasoningEffort, &mt); tb != nil {
			thinkBlock = tb
		}
	}
	inf := &types.InferenceConfiguration{}
	if mt > 0 {
		inf.MaxTokens = aws.Int32(int32(mt))
	}
	if opts.Temperature > 0 {
		inf.Temperature = aws.Float32(opts.Temperature)
	}
	if opts.TopP > 0 {
		inf.TopP = aws.Float32(opts.TopP)
	}

	in := &bedrockruntime.ConverseInput{
		ModelId:         aws.String(modelID),
		Messages:        msgs,
		InferenceConfig: inf,
	}
	if len(sys) > 0 {
		in.System = sys
	}
	if thinkBlock != nil {
		in.AdditionalModelRequestFields = document.NewLazyDocument(map[string]any{
			"thinking": thinkBlock,
		})
	}

	if opts.Stream && opts.StreamCallback != nil {
		return b.converseStream(ctx, in, opts.StreamCallback)
	}

	out, err := b.client.Converse(ctx, in)
	if err != nil {
		return nil, err
	}
	text, err := extractConverseText(out.Output)
	if err != nil {
		return nil, err
	}
	res := &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: text}},
	}
	if out.Usage != nil {
		u := &interfaces.Usage{}
		if out.Usage.InputTokens != nil {
			u.PromptTokens = int(*out.Usage.InputTokens)
		}
		if out.Usage.OutputTokens != nil {
			u.CompletionTokens = int(*out.Usage.OutputTokens)
		}
		if out.Usage.TotalTokens != nil {
			u.TotalTokens = int(*out.Usage.TotalTokens)
		}
		res.Usage = u
	}
	return res, nil
}

func (b *Bedrock) converseStream(ctx context.Context, in *bedrockruntime.ConverseInput, cb func(string)) (*interfaces.ContentResponse, error) {
	sout, err := b.client.ConverseStream(ctx, &bedrockruntime.ConverseStreamInput{
		ModelId:                      in.ModelId,
		Messages:                     in.Messages,
		System:                       in.System,
		InferenceConfig:              in.InferenceConfig,
		AdditionalModelRequestFields: in.AdditionalModelRequestFields,
	})
	if err != nil {
		return nil, err
	}
	stream := sout.GetStream()
	defer stream.Close()

	var full strings.Builder
	reasoningStart := false
	for ev := range stream.Events() {
		switch v := ev.(type) {
		case *types.ConverseStreamOutputMemberContentBlockDelta:
			switch d := v.Value.Delta.(type) {
			case *types.ContentBlockDeltaMemberReasoningContent:
				var piece string
				switch rd := d.Value.(type) {
				case *types.ReasoningContentBlockDeltaMemberText:
					piece = rd.Value
				}
				if piece == "" {
					continue
				}
				out := ""
				if !reasoningStart {
					reasoningStart = true
					out = AssistantReasoningOpenTag
				}
				out += piece
				full.WriteString(out)
				cb(out)
			case *types.ContentBlockDeltaMemberText:
				if d.Value == "" {
					continue
				}
				if reasoningStart {
					reasoningStart = false
					full.WriteString(AssistantReasoningCloseTag)
					cb(AssistantReasoningCloseTag)
				}
				full.WriteString(d.Value)
				cb(d.Value)
			}
		}
	}
	if reasoningStart {
		full.WriteString(AssistantReasoningCloseTag)
		cb(AssistantReasoningCloseTag)
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}
	return &interfaces.ContentResponse{
		Choices: []interfaces.Choice{{Content: full.String()}},
	}, nil
}

func extractConverseText(out types.ConverseOutput) (string, error) {
	msg, ok := out.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("bedrock: 非文本消息响应")
	}
	var sb strings.Builder
	reasoningStart := false
	closeReasoning := func() {
		if reasoningStart {
			reasoningStart = false
			sb.WriteString(AssistantReasoningCloseTag)
		}
	}
	for _, block := range msg.Value.Content {
		switch b := block.(type) {
		case *types.ContentBlockMemberReasoningContent:
			switch rc := b.Value.(type) {
			case *types.ReasoningContentBlockMemberReasoningText:
				if rc.Value.Text == nil {
					continue
				}
				t := strings.TrimSpace(*rc.Value.Text)
				if t == "" {
					continue
				}
				if !reasoningStart {
					reasoningStart = true
					sb.WriteString(AssistantReasoningOpenTag)
				}
				sb.WriteString(*rc.Value.Text)
			}
		case *types.ContentBlockMemberText:
			closeReasoning()
			sb.WriteString(b.Value)
		}
	}
	closeReasoning()
	if sb.Len() == 0 {
		return "", errors.New("bedrock: 模型未返回文本内容")
	}
	return sb.String(), nil
}

func toBedrockMessages(messages []interfaces.MessageContent) (sys []types.SystemContentBlock, out []types.Message, err error) {
	var systemText strings.Builder
	var pending []types.Message

	for _, m := range messages {
		text := partsToPlainText(m.Parts)
		switch m.Role {
		case interfaces.MessageRoleSystem:
			if text != "" {
				systemText.WriteString(text)
				systemText.WriteByte('\n')
			}
		case interfaces.MessageRoleHuman:
			pending = append(pending, types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: text},
				},
			})
		case interfaces.MessageRoleAssistant:
			if len(m.ToolCalls) > 0 {
				var lines []string
				for _, tc := range m.ToolCalls {
					lines = append(lines, fmt.Sprintf("[tool %s] %s", tc.Name, tc.Arguments))
				}
				text = strings.TrimSpace(text + "\n" + strings.Join(lines, "\n"))
			}
			pending = append(pending, types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: text},
				},
			})
		case interfaces.MessageRoleTool:
			// 作为 user 侧上下文，避免 ToolResult 结构依赖 tool_use id
			tid := m.ToolName
			if tid == "" {
				tid = m.ToolCallID
			}
			if tid == "" {
				tid = "result"
			}
			pending = append(pending, types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: fmt.Sprintf("[tool %s output]\n%s", tid, text)},
				},
			})
		default:
			pending = append(pending, types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: text},
				},
			})
		}
	}

	if systemText.Len() > 0 {
		sys = append(sys, &types.SystemContentBlockMemberText{Value: strings.TrimSpace(systemText.String())})
	}
	return sys, pending, nil
}

func partsToPlainText(parts []interfaces.ContentPart) string {
	var sb strings.Builder
	for _, p := range parts {
		if t, ok := p.(interfaces.TextContent); ok {
			sb.WriteString(t.Text)
		}
	}
	return strings.TrimSpace(sb.String())
}

func (b *Bedrock) Call(ctx context.Context, prompt string, options ...interfaces.CallOption) (string, error) {
	msg := interfaces.MessageContent{
		Role:  interfaces.MessageRoleHuman,
		Parts: []interfaces.ContentPart{interfaces.TextContent{Text: prompt}},
	}
	resp, err := b.GenerateContent(ctx, []interfaces.MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("bedrock: 空响应")
	}
	return resp.Choices[0].Content, nil
}

func (b *Bedrock) ProviderName() string { return "bedrock" }
func (b *Bedrock) ModelName() string    { return b.modelID }

// bedrockModelSupportsAnthropicThinking 仅对 Claude / Anthropic 系模型附加 thinking 参数，避免 Llama 等返回 400。
func bedrockModelSupportsAnthropicThinking(modelID string) bool {
	m := strings.ToLower(strings.TrimSpace(modelID))
	if m == "" {
		return false
	}
	return strings.Contains(m, "anthropic") || strings.Contains(m, "claude")
}
