package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// ReplicateEmbed 通过 Replicate Predictions API 调用嵌入模型（与上游参考 ReplicateEmbed 一致：input 默认键 texts）
type ReplicateEmbed struct {
	token      string
	model      string
	baseAPI    string
	client     *http.Client
	dimensions int
	inputKey   string
}

// NewReplicateEmbed model 为 owner/name；inputKey 为空时使用 "texts"
func NewReplicateEmbed(token, model string, dimensions int, inputKey string) *ReplicateEmbed {
	if inputKey == "" {
		inputKey = "texts"
	}
	return &ReplicateEmbed{
		token:      strings.TrimSpace(token),
		model:      strings.TrimSpace(model),
		baseAPI:    "https://api.replicate.com/v1",
		client:     &http.Client{Timeout: 180 * time.Second},
		dimensions: dimensions,
		inputKey:   inputKey,
	}
}

type replicatePred struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

func (e *ReplicateEmbed) predictionURL() string {
	return fmt.Sprintf("%s/models/%s/predictions", e.baseAPI, strings.Trim(e.model, "/"))
}

func (e *ReplicateEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	if e.token == "" {
		return nil, fmt.Errorf("replicate embed: 需要 API Token")
	}
	if e.model == "" {
		return nil, fmt.Errorf("replicate embed: 需要模型名 owner/name")
	}
	const batchSize = 16
	var all [][]float32
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		vecs, err := e.runBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
		if len(vecs) != len(batch) {
			return nil, fmt.Errorf("replicate embed: 期望 %d 条向量，得到 %d", len(batch), len(vecs))
		}
		all = append(all, vecs...)
	}
	return all, nil
}

func (e *ReplicateEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	vecs, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("replicate embed: 空结果")
	}
	return vecs[0], nil
}

func (e *ReplicateEmbed) runBatch(ctx context.Context, batch []string) ([][]float32, error) {
	input := map[string]any{e.inputKey: batch}
	body, _ := json.Marshal(map[string]any{"input": input})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.predictionURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait=120")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("replicate embed: HTTP %d %s", resp.StatusCode, string(respBody))
	}
	var pred replicatePred
	if err := json.Unmarshal(respBody, &pred); err != nil {
		return nil, fmt.Errorf("replicate embed: %w", err)
	}
	out, err := e.resolveOutput(ctx, &pred)
	if err != nil {
		return nil, err
	}
	return decodeReplicateEmbeddingVectors(out)
}

func (e *ReplicateEmbed) resolveOutput(ctx context.Context, pred *replicatePred) (json.RawMessage, error) {
	p := pred
	for range 90 {
		switch p.Status {
		case "succeeded":
			if len(p.Output) == 0 || string(p.Output) == "null" {
				return nil, fmt.Errorf("replicate embed: 空输出")
			}
			return p.Output, nil
		case "failed", "canceled":
			if p.Error != "" {
				return nil, fmt.Errorf("replicate embed: %s", p.Error)
			}
			return nil, fmt.Errorf("replicate embed: 状态 %s", p.Status)
		case "starting", "processing":
			if p.URLs.Get == "" {
				return nil, fmt.Errorf("replicate embed: 缺少轮询 URL")
			}
			time.Sleep(2 * time.Second)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URLs.Get, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+e.token)
			resp, err := e.client.Do(req)
			if err != nil {
				return nil, err
			}
			b, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(b, &p); err != nil {
				return nil, err
			}
			continue
		default:
			if len(p.Output) > 0 && string(p.Output) != "null" {
				return p.Output, nil
			}
			if p.URLs.Get != "" {
				time.Sleep(2 * time.Second)
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URLs.Get, nil)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Authorization", "Bearer "+e.token)
				resp, err := e.client.Do(req)
				if err != nil {
					return nil, err
				}
				b, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					return nil, err
				}
				if err := json.Unmarshal(b, &p); err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("replicate embed: 未预期状态 %q", p.Status)
		}
	}
	return nil, fmt.Errorf("replicate embed: 等待结果超时")
}

func decodeReplicateEmbeddingVectors(raw json.RawMessage) ([][]float32, error) {
	raw = unwrapEmbeddingObject(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("replicate embed: 无法解析向量")
	}
	// 单层向量（单条文本）
	if v, err := floatVecFromJSON(raw); err == nil {
		return [][]float32{v}, nil
	}
	// 多条：JSON 数组，每项为一向量
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("replicate embed: 解析输出: %w", err)
	}
	out := make([][]float32, 0, len(rows))
	for _, row := range rows {
		v, err := floatVecFromJSON(row)
		if err != nil {
			return nil, fmt.Errorf("replicate embed: 向量行: %w", err)
		}
		out = append(out, v)
	}
	return out, nil
}

func unwrapEmbeddingObject(raw json.RawMessage) json.RawMessage {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil || len(obj) == 0 {
		return raw
	}
	for _, k := range []string{"embeddings", "embedding", "vectors", "output", "data"} {
		if v, ok := obj[k]; ok && len(v) > 0 && string(v) != "null" {
			return unwrapEmbeddingObject(v)
		}
	}
	return raw
}

func floatVecFromJSON(data json.RawMessage) ([]float32, error) {
	var xs []float64
	if err := json.Unmarshal(data, &xs); err == nil {
		out := make([]float32, len(xs))
		for i, v := range xs {
			out[i] = float32(v)
		}
		return out, nil
	}
	var xs32 []float32
	if err := json.Unmarshal(data, &xs32); err == nil {
		return xs32, nil
	}
	return nil, fmt.Errorf("非数值向量")
}

func (e *ReplicateEmbed) ProviderName() string { return "replicate" }
func (e *ReplicateEmbed) ModelName() string    { return e.model }
func (e *ReplicateEmbed) Dimensions() int      { return e.dimensions }

var _ interfaces.Embedder = (*ReplicateEmbed)(nil)
