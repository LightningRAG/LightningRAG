package component

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	ragretriever "github.com/LightningRAG/LightningRAG/server/rag/providers/retriever"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

func init() {
	Register("Retrieval", NewRetrieval)
}

// Retrieval 知识库检索组件
type Retrieval struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewRetrieval 创建 Retrieval 组件
func NewRetrieval(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &Retrieval{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (r *Retrieval) ComponentName() string {
	return "Retrieval"
}

// Invoke 执行
func (r *Retrieval) Invoke(inputs map[string]any) error {
	r.mu.Lock()
	r.err = ""
	r.mu.Unlock()

	queryVar := getStrParam(r.params, "query")
	if queryVar == "" {
		queryVar = "sys.query"
	}
	query := r.canvas.ResolveString(NormalizeSingleRefForResolve(queryVar))
	if query == "" {
		query = getStrParam(r.params, "query")
	}
	if query == "" {
		r.mu.Lock()
		r.err = "检索 query 为空"
		r.mu.Unlock()
		return fmt.Errorf("检索 query 为空")
	}

	topN := getIntParam(r.params, "top_n", 6)
	emptyResp := getStrParam(r.params, "empty_response")
	if emptyResp == "" {
		emptyResp = "No relevant knowledge found."
	}

	factory := r.canvas.GetRetrieverFactory()
	if factory == nil {
		r.mu.Lock()
		r.output["formalized_content"] = emptyResp
		r.mu.Unlock()
		return nil
	}

	kbIDs := getUintSliceParam(r.params, "kb_ids")
	if len(kbIDs) == 0 {
		r.mu.Lock()
		r.output["formalized_content"] = emptyResp
		r.mu.Unlock()
		return nil
	}

	retriever, err := factory(context.Background(), kbIDs, r.canvas.GetTenantID(), topN, RetrieverFactoryOptions{
		PageIndexTocEnhance: getOptionalBoolPtrParam(r.params, "toc_enhance", "tocEnhance"),
	})
	if err != nil {
		r.mu.Lock()
		r.err = err.Error()
		r.mu.Unlock()
		return err
	}
	if retriever == nil {
		r.mu.Lock()
		r.err = "无法创建知识库检索器，请检查向量存储配置"
		r.output["formalized_content"] = emptyResp
		r.mu.Unlock()
		return fmt.Errorf("无法创建知识库检索器，请检查向量存储配置")
	}

	docs, err := retriever.GetRelevantDocuments(context.Background(), query, topN)
	if err != nil {
		r.mu.Lock()
		r.err = err.Error()
		r.mu.Unlock()
		return err
	}

	formalized := formatChunks(docs)
	if formalized == "" {
		formalized = emptyResp
	}

	r.mu.Lock()
	r.output["formalized_content"] = formalized
	r.mu.Unlock()
	return nil
}

func formatChunks(docs []Document) string {
	if len(docs) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, d := range docs {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, d.PageContent))
	}
	return strings.TrimSpace(sb.String())
}

// Output 获取输出
func (r *Retrieval) Output(key string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.output[key]
}

// OutputAll 获取所有输出
func (r *Retrieval) OutputAll() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range r.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (r *Retrieval) SetOutput(key string, value any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.output[key] = value
}

// Error 返回错误
func (r *Retrieval) Error() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.err
}

// Reset 重置
func (r *Retrieval) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.output = make(map[string]any)
	r.err = ""
}

// RAGRetrieverAdapter 将 interfaces.Retriever 适配为 component.Retriever
type RAGRetrieverAdapter struct {
	R interfaces.Retriever
}

func (a *RAGRetrieverAdapter) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]Document, error) {
	docs, err := a.R.GetRelevantDocuments(ctx, query, numDocs)
	if err != nil {
		return nil, err
	}
	docs = ragretriever.DeduplicateRetrievedDocuments(docs)
	out := make([]Document, len(docs))
	for i, d := range docs {
		out[i] = schemaDocToComponentDoc(d)
	}
	return out, nil
}

func schemaDocToComponentDoc(d schema.Document) Document {
	return Document{
		PageContent: d.PageContent,
		Metadata:    d.Metadata,
		Score:       d.Score,
	}
}
