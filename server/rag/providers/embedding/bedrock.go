package embedding

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
)

// BedrockEmbed 使用 Amazon Bedrock InvokeModel 调用 Titan 等嵌入模型
type BedrockEmbed struct {
	client     *bedrockruntime.Client
	model      string
	dimensions int
}

type bedrockEmbedCreds struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
	SessionToken    string `json:"aws_session_token"`
	Region          string `json:"aws_region"`
}

func loadBedrockConfigForEmbed(apiKey, region string, extra map[string]any) (aws.Config, error) {
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
		var j bedrockEmbedCreds
		if err := json.Unmarshal([]byte(apiKey), &j); err != nil {
			return aws.Config{}, fmt.Errorf("bedrock embed: 解析凭证 JSON: %w", err)
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

// NewBedrockEmbed model 如 amazon.titan-embed-text-v2:0；dimensions 仅部分模型支持，0 表示省略
func NewBedrockEmbed(apiKey, region, model string, extra map[string]any, dimensions int) (*BedrockEmbed, error) {
	if strings.TrimSpace(model) == "" {
		model = "amazon.titan-embed-text-v2:0"
	}
	cfg, err := loadBedrockConfigForEmbed(apiKey, region, extra)
	if err != nil {
		return nil, err
	}
	return &BedrockEmbed{
		client:     bedrockruntime.NewFromConfig(cfg),
		model:      model,
		dimensions: dimensions,
	}, nil
}

type titanEmbedRequest struct {
	InputText  string `json:"inputText"`
	Dimensions int    `json:"dimensions,omitempty"`
	Normalize  bool   `json:"normalize,omitempty"`
}

type titanEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (e *BedrockEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	out := make([][]float32, 0, len(texts))
	for _, t := range texts {
		v, err := e.embedOne(ctx, t)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (e *BedrockEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.embedOne(ctx, text)
}

func (e *BedrockEmbed) embedOne(ctx context.Context, text string) ([]float32, error) {
	reqBody := titanEmbedRequest{InputText: text, Normalize: true}
	if e.dimensions > 0 {
		reqBody.Dimensions = e.dimensions
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	out, err := e.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(e.model),
		ContentType: aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return nil, err
	}
	body := out.Body
	var resp titanEmbedResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("bedrock embed: 解析响应: %w; body=%s", err, truncateForErr(body))
	}
	if len(resp.Embedding) == 0 {
		return nil, errors.New("bedrock embed: 空向量")
	}
	v := make([]float32, len(resp.Embedding))
	for i, x := range resp.Embedding {
		v[i] = float32(x)
	}
	return v, nil
}

func truncateForErr(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "..."
	}
	return string(b)
}

func (e *BedrockEmbed) ProviderName() string { return "bedrock" }
func (e *BedrockEmbed) ModelName() string    { return e.model }
func (e *BedrockEmbed) Dimensions() int      { return e.dimensions }
