package vectorstore

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MysqlVectorStore 基于 MySQL 的向量存储
// 向量以 JSON 数组形式存储，相似度检索在应用层计算
type MysqlVectorStore struct {
	db             *gorm.DB
	embedder       interfaces.Embedder
	collectionName string
	tableName      string
	vectorDims     int
}

type mysqlVectorRow struct {
	ID        string `gorm:"primaryKey;column:id;size:36"`
	Embedding string `gorm:"type:LONGTEXT;column:embedding"`
	Document  string `gorm:"type:LONGTEXT;column:document"`
	Namespace string `gorm:"column:namespace;size:255;index"`
	Metadata  string `gorm:"type:JSON;column:metadata"`
}

func (mysqlVectorRow) TableName() string {
	return "rag_vector_embeddings"
}

func NewMysqlVectorStore(db *gorm.DB, embedder interfaces.Embedder, collectionName string, vectorDims int) (*MysqlVectorStore, error) {
	if vectorDims <= 0 && embedder != nil {
		vectorDims = embedder.Dimensions()
	}
	if vectorDims <= 0 {
		vectorDims = 1536
	}
	s := &MysqlVectorStore{
		db:             db,
		embedder:       embedder,
		collectionName: collectionName,
		tableName:      "rag_vector_embeddings",
		vectorDims:     vectorDims,
	}
	return s, s.ensureTable(context.Background())
}

func (s *MysqlVectorStore) ensureTable(_ context.Context) error {
	return s.db.AutoMigrate(&mysqlVectorRow{})
}

func (s *MysqlVectorStore) AddDocuments(ctx context.Context, docs []schema.Document, options ...interfaces.VectorStoreOption) ([]string, error) {
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

	ids := make([]string, len(docs))
	for i, doc := range docs {
		id := uuid.New().String()
		ids[i] = id

		meta := "{}"
		if doc.Metadata != nil {
			if b, err := json.Marshal(doc.Metadata); err == nil {
				meta = string(b)
			}
		}

		embJSON, err := json.Marshal(vectors[i])
		if err != nil {
			return nil, fmt.Errorf("marshal embedding %d: %w", i, err)
		}

		row := mysqlVectorRow{
			ID:        id,
			Embedding: string(embJSON),
			Document:  doc.PageContent,
			Namespace: ns,
			Metadata:  meta,
		}

		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, fmt.Errorf("insert doc %d: %w", i, err)
		}
	}
	return ids, nil
}

func (s *MysqlVectorStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...interfaces.VectorStoreOption) ([]schema.Document, error) {
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

	var rows []mysqlVectorRow
	if err := s.db.WithContext(ctx).Where("namespace = ?", ns).Find(&rows).Error; err != nil {
		return nil, err
	}

	type scored struct {
		doc   schema.Document
		score float64
	}
	results := make([]scored, 0, len(rows))
	for _, r := range rows {
		var vec []float32
		if err := json.Unmarshal([]byte(r.Embedding), &vec); err != nil {
			continue
		}
		sim := cosineSimilarity(queryVec, vec)
		results = append(results, scored{
			doc: schema.Document{
				PageContent: r.Document,
				Metadata:    parseMetadata(r.Metadata),
				Score:       float32(sim),
			},
			score: sim,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if len(results) > numDocuments {
		results = results[:numDocuments]
	}

	docs := make([]schema.Document, len(results))
	for i, r := range results {
		docs[i] = r.doc
	}
	ApplyConfiguredRankBoostToScores(docs)
	return docs, nil
}

func keywordQueryTokens(q string) []string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return []string{q}
	}
	return parts
}

func countKeywordHits(docLower string, tokens []string) int {
	n := 0
	for _, t := range tokens {
		tl := strings.ToLower(strings.TrimSpace(t))
		if tl == "" {
			continue
		}
		if strings.Contains(docLower, tl) {
			n++
		}
	}
	return n
}

func (s *MysqlVectorStore) KeywordSearch(ctx context.Context, query string, numDocuments int, options ...interfaces.VectorStoreOption) ([]schema.Document, error) {
	opts := &interfaces.VectorStoreOptions{Namespace: s.collectionName}
	for _, o := range options {
		o(opts)
	}
	ns := opts.Namespace
	if ns == "" {
		ns = s.collectionName
	}
	tokens := keywordQueryTokens(query)
	if len(tokens) == 0 {
		return nil, nil
	}
	var rows []mysqlVectorRow
	if err := s.db.WithContext(ctx).Where("namespace = ?", ns).Find(&rows).Error; err != nil {
		return nil, err
	}
	type scored struct {
		doc   schema.Document
		score float64
	}
	results := make([]scored, 0, len(rows))
	denom := float64(len(tokens))
	queryLower := strings.ToLower(strings.TrimSpace(query))
	for _, r := range rows {
		dl := strings.ToLower(r.Document)
		hits := countKeywordHits(dl, tokens)
		if opts.RelaxedKeywordSearch && hits == 0 && queryLower != "" && strings.Contains(dl, queryLower) {
			hits = 1
		}
		if hits == 0 {
			continue
		}
		results = append(results, scored{
			doc: schema.Document{
				PageContent: r.Document,
				Metadata:    parseMetadata(r.Metadata),
				Score:       float32(float64(hits) / denom),
			},
			score: float64(hits) / denom,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		li, lj := len(results[i].doc.PageContent), len(results[j].doc.PageContent)
		if li != lj {
			return li > lj
		}
		return results[i].doc.PageContent < results[j].doc.PageContent
	})
	if len(results) > numDocuments {
		results = results[:numDocuments]
	}
	docs := make([]schema.Document, len(results))
	for i, r := range results {
		docs[i] = r.doc
	}
	ApplyConfiguredRankBoostToScores(docs)
	return docs, nil
}

func (s *MysqlVectorStore) DeleteByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&mysqlVectorRow{}).Error
}

func (s *MysqlVectorStore) DeleteByNamespace(ctx context.Context, namespace string) error {
	return s.db.WithContext(ctx).Where("namespace = ?", namespace).Delete(&mysqlVectorRow{}).Error
}

func (s *MysqlVectorStore) DeleteByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) error {
	if metaKey != "document_id" {
		metaKey = "document_id"
	}
	val := fmt.Sprintf("%v", metaValue)
	return s.db.WithContext(ctx).
		Where("namespace = ?", namespace).
		Where("JSON_EXTRACT(metadata, ?) = ?", "$."+metaKey, val).
		Delete(&mysqlVectorRow{}).Error
}

func (s *MysqlVectorStore) ListByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) ([]schema.Document, error) {
	val := fmt.Sprintf("%v", metaValue)
	var rows []mysqlVectorRow
	err := s.db.WithContext(ctx).
		Where("namespace = ?", namespace).
		Where("JSON_EXTRACT(metadata, ?) = CAST(? AS JSON)", "$."+metaKey, val).
		Order("CAST(JSON_EXTRACT(metadata, '$.chunk_index') AS UNSIGNED) ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	docs := make([]schema.Document, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, schema.Document{
			PageContent: r.Document,
			Metadata:    parseMetadata(r.Metadata),
		})
	}
	return docs, nil
}

func (s *MysqlVectorStore) ProviderName() string { return "mysql" }

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		normA += fa * fa
		normB += fb * fb
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
