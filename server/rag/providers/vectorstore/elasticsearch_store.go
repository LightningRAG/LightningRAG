package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

// ElasticsearchStore 基于 Elasticsearch 8.x 的向量存储
// 参考 references 目录内 的 ES 设计，使用 dense_vector + cosine 相似度
// Config 字段: address(必填), username, password, index_prefix(可选，默认 rag_vectors)
type ElasticsearchStore struct {
	client         *elasticsearch.Client
	embedder       interfaces.Embedder
	collectionName string
	indexName      string
	vectorDims     int
}

func strFromConfig(cfg map[string]any, key string) string {
	if cfg == nil {
		return ""
	}
	if v, ok := cfg[key]; ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// NewElasticsearchStore 创建 Elasticsearch 向量存储
func NewElasticsearchStore(cfg map[string]any, embedder interfaces.Embedder, collectionName string, vectorDims int) (*ElasticsearchStore, error) {
	address := strFromConfig(cfg, "address")
	if address == "" {
		address = "http://localhost:9200"
	}
	if vectorDims <= 0 && embedder != nil {
		vectorDims = embedder.Dimensions()
	}
	if vectorDims <= 0 {
		vectorDims = 1536
	}

	prefix := strFromConfig(cfg, "index_prefix")
	if prefix == "" {
		prefix = "rag_vectors"
	}
	// 索引名：rag_vectors 或 rag_vectors_{dims}，按维度分索引便于不同 embedding 模型
	indexName := prefix + "_" + strconv.Itoa(vectorDims)

	esCfg := elasticsearch.Config{
		Addresses: []string{address},
	}
	if u := strFromConfig(cfg, "username"); u != "" {
		esCfg.Username = u
	}
	if p := strFromConfig(cfg, "password"); p != "" {
		esCfg.Password = p
	}
	if caCert := strFromConfig(cfg, "ca_cert"); caCert != "" {
		// 可扩展 TLS 配置
		_ = caCert
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 Elasticsearch 客户端失败: %w", err)
	}

	s := &ElasticsearchStore{
		client:         client,
		embedder:       embedder,
		collectionName: collectionName,
		indexName:      indexName,
		vectorDims:     vectorDims,
	}
	if err := s.ensureIndex(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ElasticsearchStore) ensureIndex(ctx context.Context) error {
	mapping := fmt.Sprintf(`{
		"mappings": {
			"properties": {
				"document": { "type": "text" },
				"namespace": { "type": "keyword" },
				"metadata": {
					"properties": {
						"document_id": { "type": "long" },
						"chunk_index": { "type": "integer" },
						"doc_name": { "type": "keyword" },
						"rank_boost": { "type": "float" }
					}
				},
				"embedding": {
					"type": "dense_vector",
					"dims": %d,
					"index": true,
					"similarity": "cosine"
				}
			}
		}
	}`, s.vectorDims)

	req := esapi.IndicesCreateRequest{
		Index: s.indexName,
		Body:  strings.NewReader(mapping),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 400 && strings.Contains(res.String(), "resource_already_exists") {
			// 索引已存在，尝试更新 mapping 以添加 metadata 子字段索引
			s.tryUpdateMetadataMapping(ctx)
			return nil
		}
		return fmt.Errorf("创建索引失败: %s", res.String())
	}
	return nil
}

// tryUpdateMetadataMapping 尝试为已有索引更新 metadata 字段映射（向后兼容旧索引）
func (s *ElasticsearchStore) tryUpdateMetadataMapping(ctx context.Context) {
	mapping := `{
		"properties": {
			"metadata": {
				"properties": {
					"document_id": { "type": "long" },
					"chunk_index": { "type": "integer" },
					"doc_name": { "type": "keyword" },
					"rank_boost": { "type": "float" }
				}
			}
		}
	}`
	putReq := esapi.IndicesPutMappingRequest{
		Index: []string{s.indexName},
		Body:  strings.NewReader(mapping),
	}
	res, err := putReq.Do(ctx, s.client)
	if err != nil {
		return
	}
	defer res.Body.Close()
}

func (s *ElasticsearchStore) AddDocuments(ctx context.Context, docs []schema.Document, options ...interfaces.VectorStoreOption) ([]string, error) {
	opts := &interfaces.VectorStoreOptions{Namespace: s.collectionName}
	for _, o := range options {
		o(opts)
	}
	ns := opts.Namespace
	if ns == "" {
		ns = s.collectionName
	}

	texts := make([]string, len(docs))
	for i := range docs {
		texts[i] = docs[i].PageContent
	}
	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	if len(vectors) != len(docs) {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(vectors), len(docs))
	}

	var buf bytes.Buffer
	ids := make([]string, len(docs))
	for i, doc := range docs {
		id := uuid.New().String()
		ids[i] = id

		meta := doc.Metadata
		if meta == nil {
			meta = map[string]any{}
		}

		// bulk 格式：action + doc
		action := map[string]any{
			"index": map[string]string{"_index": s.indexName, "_id": id},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		docBody := map[string]any{
			"document":  doc.PageContent,
			"namespace": ns,
			"metadata":  meta,
			"embedding": vectors[i],
		}
		docJSON, _ := json.Marshal(docBody)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	res, err := s.client.Bulk(bytes.NewReader(buf.Bytes()), s.client.Bulk.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("bulk 写入失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("bulk 写入失败: %s", res.String())
	}
	var bulkResp struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Index *struct {
				Error *struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"error"`
			} `json:"index"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err == nil && bulkResp.Errors {
		for _, item := range bulkResp.Items {
			if item.Index != nil && item.Index.Error != nil {
				return nil, fmt.Errorf("bulk 写入错误: %s", item.Index.Error.Reason)
			}
		}
	}
	return ids, nil
}

func (s *ElasticsearchStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...interfaces.VectorStoreOption) ([]schema.Document, error) {
	opts := &interfaces.VectorStoreOptions{Namespace: s.collectionName}
	for _, o := range options {
		o(opts)
	}
	ns := opts.Namespace
	if ns == "" {
		ns = s.collectionName
	}

	queryVec, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	k := numDocuments
	if k < 1 {
		k = 1
	}
	// Ragflow 对向量腿用 topk≈1024；ES kNN 的 num_candidates 过小会导致近似索引漏召（官方建议远大于 k）
	numCandidates := k * 4
	if numCandidates < 200 {
		numCandidates = 200
	}
	const maxESKNNCandidates = 10000
	if numCandidates > maxESKNNCandidates {
		numCandidates = maxESKNNCandidates
	}
	knnQuery := map[string]any{
		"knn": map[string]any{
			"field":          "embedding",
			"query_vector":   queryVec,
			"k":              k,
			"num_candidates": numCandidates,
			"filter": map[string]any{
				"term": map[string]string{"namespace": ns},
			},
		},
	}
	queryBody, _ := json.Marshal(knnQuery)

	req := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryBody),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("kNN 检索失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("kNN 检索失败: %s", res.String())
	}

	var searchResp struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Document string         `json:"document"`
					Metadata map[string]any `json:"metadata"`
				} `json:"_source"`
				Score *float64 `json:"_score"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析检索结果失败: %w", err)
	}

	docs := make([]schema.Document, 0, len(searchResp.Hits.Hits))
	for _, h := range searchResp.Hits.Hits {
		score := float32(0)
		if h.Score != nil {
			score = float32(*h.Score)
		}
		docs = append(docs, schema.Document{
			PageContent: h.Source.Document,
			Metadata:    h.Source.Metadata,
			Score:       score,
		})
	}
	ApplyConfiguredRankBoostToScores(docs)
	return docs, nil
}

func (s *ElasticsearchStore) KeywordSearch(ctx context.Context, query string, numDocuments int, options ...interfaces.VectorStoreOption) ([]schema.Document, error) {
	opts := &interfaces.VectorStoreOptions{Namespace: s.collectionName}
	for _, o := range options {
		o(opts)
	}
	ns := opts.Namespace
	if ns == "" {
		ns = s.collectionName
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}
	multi := map[string]any{
		"query":    q,
		"fields":   []string{"document"},
		"type":     "best_fields",
		"operator": "or",
	}
	if opts.RelaxedKeywordSearch {
		// Ragflow 第二次 search：更宽松词项匹配 + 模糊（近似降低 min_match）
		multi["fuzziness"] = "AUTO"
		multi["minimum_should_match"] = "10%"
	}
	body := map[string]any{
		"size": numDocuments,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{
					map[string]any{"term": map[string]string{"namespace": ns}},
					map[string]any{"multi_match": multi},
				},
			},
		},
	}
	queryBody, _ := json.Marshal(body)
	req := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryBody),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("全文检索失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("全文检索失败: %s", res.String())
	}
	var searchResp struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Document string         `json:"document"`
					Metadata map[string]any `json:"metadata"`
				} `json:"_source"`
				Score *float64 `json:"_score"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析全文检索结果失败: %w", err)
	}
	docs := make([]schema.Document, 0, len(searchResp.Hits.Hits))
	for _, h := range searchResp.Hits.Hits {
		score := float32(0)
		if h.Score != nil {
			score = float32(*h.Score)
		}
		docs = append(docs, schema.Document{
			PageContent: h.Source.Document,
			Metadata:    h.Source.Metadata,
			Score:       score,
		})
	}
	// BM25 的 _score 无上界，与向量余弦（约 0~1）在 hybrid/mix 合并时同列展示会混用「百分比」与「原始分」。
	// 按本批命中最高分缩放到 (0,1]，保留批内排序，便于与向量分统一展示与阈值语义。
	if len(docs) > 0 {
		var maxS float32
		for i := range docs {
			if docs[i].Score > maxS {
				maxS = docs[i].Score
			}
		}
		if maxS > 0 {
			for i := range docs {
				docs[i].Score = docs[i].Score / maxS
			}
		}
	}
	ApplyConfiguredRankBoostToScores(docs)
	return docs, nil
}

func (s *ElasticsearchStore) DeleteByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for _, id := range ids {
		action := map[string]any{
			"delete": map[string]string{"_index": s.indexName, "_id": id},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')
	}
	res, err := s.client.Bulk(bytes.NewReader(buf.Bytes()), s.client.Bulk.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("删除失败: %s", res.String())
	}
	return nil
}

func (s *ElasticsearchStore) DeleteByNamespace(ctx context.Context, namespace string) error {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]string{"namespace": namespace},
		},
	}
	queryBody, _ := json.Marshal(query)
	req := esapi.DeleteByQueryRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryBody),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("按命名空间删除失败: %s", res.String())
	}
	return nil
}

func (s *ElasticsearchStore) DeleteByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) error {
	if metaKey != "document_id" {
		metaKey = "document_id"
	}
	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]string{"namespace": namespace}},
					{"term": map[string]any{"metadata." + metaKey: metaValue}},
				},
			},
		},
	}
	queryBody, _ := json.Marshal(query)
	req := esapi.DeleteByQueryRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryBody),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("按 metadata 删除失败: %s", res.String())
	}
	return nil
}

func (s *ElasticsearchStore) ListByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) ([]schema.Document, error) {
	metaValueStr := fmt.Sprintf("%v", metaValue)

	// 使用 runtime_mappings 在查询时动态创建虚拟字段，兼容旧索引（metadata 未索引）和新索引
	query := map[string]any{
		"runtime_mappings": map[string]any{
			"meta_filter_field": map[string]any{
				"type": "keyword",
				"script": map[string]string{
					"source": fmt.Sprintf(
						"if (params._source.metadata != null && params._source.metadata.containsKey('%s')) { def val = params._source.metadata.%s; if (val instanceof Number) { emit(String.valueOf(((Number)val).longValue())); } else { emit(String.valueOf(val)); } }",
						metaKey, metaKey,
					),
				},
			},
			"meta_chunk_index": map[string]any{
				"type": "long",
				"script": map[string]string{
					"source": "if (params._source.metadata != null && params._source.metadata.containsKey('chunk_index')) { emit(((Number)params._source.metadata.chunk_index).longValue()); } else { emit(0); }",
				},
			},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]string{"namespace": namespace}},
					{"term": map[string]string{"meta_filter_field": metaValueStr}},
				},
			},
		},
		"sort":    []map[string]string{{"meta_chunk_index": "asc"}},
		"size":    10000,
		"_source": []string{"document", "metadata"},
	}
	queryBody, _ := json.Marshal(query)

	req := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryBody),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("ES ListByMetadata 查询失败: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("ES ListByMetadata 查询失败: %s", res.String())
	}

	var searchResp struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Document string         `json:"document"`
					Metadata map[string]any `json:"metadata"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析 ES 响应失败: %w", err)
	}

	docs := make([]schema.Document, 0, len(searchResp.Hits.Hits))
	for _, h := range searchResp.Hits.Hits {
		docs = append(docs, schema.Document{
			PageContent: h.Source.Document,
			Metadata:    h.Source.Metadata,
		})
	}
	return docs, nil
}

func (s *ElasticsearchStore) ProviderName() string { return "elasticsearch" }
