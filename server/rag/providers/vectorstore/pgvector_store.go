package vectorstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// PgVectorStore 基于 PostgreSQL + pgvector 的向量存储
// 需要数据库启用 pgvector 扩展: CREATE EXTENSION IF NOT EXISTS vector;
type PgVectorStore struct {
	db             *gorm.DB
	embedder       interfaces.Embedder
	collectionName string
	tableName      string
	vectorDims     int
}

// pgVectorRow 存储行
type pgVectorRow struct {
	ID        string          `gorm:"primaryKey;column:id;size:36"`
	Embedding pgvector.Vector `gorm:"type:vector;column:embedding"`
	Document  string          `gorm:"column:document"`
	Namespace string          `gorm:"column:namespace;index"`
	Metadata  string          `gorm:"type:jsonb;column:metadata"`
}

func (pgVectorRow) TableName() string {
	return "rag_vector_embeddings"
}

// NewPgVectorStore 创建 PgVector 存储
func NewPgVectorStore(db *gorm.DB, embedder interfaces.Embedder, collectionName string, vectorDims int) (*PgVectorStore, error) {
	if vectorDims <= 0 && embedder != nil {
		vectorDims = embedder.Dimensions()
	}
	if vectorDims <= 0 {
		vectorDims = 1536 // OpenAI 默认维度
	}
	s := &PgVectorStore{
		db:             db,
		embedder:       embedder,
		collectionName: collectionName,
		tableName:      "rag_vector_embeddings",
		vectorDims:     vectorDims,
	}
	return s, s.ensureTable(context.Background())
}

func (s *PgVectorStore) ensureTable(ctx context.Context) error {
	// 需要 PostgreSQL 且已安装 pgvector: CREATE EXTENSION IF NOT EXISTS vector;
	s.db.WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS vector")
	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id varchar(36) PRIMARY KEY,
			embedding vector(%d),
			document text,
			namespace text,
			metadata jsonb
		);
	`, s.tableName, s.vectorDims)
	if err := s.db.WithContext(ctx).Exec(sql).Error; err != nil {
		return err
	}
	// 创建 namespace 索引（IF NOT EXISTS 需 PostgreSQL 9.5+）
	s.db.WithContext(ctx).Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_namespace ON %s(namespace)", s.tableName, s.tableName))
	return nil
}

func (s *PgVectorStore) AddDocuments(ctx context.Context, docs []schema.Document, options ...interfaces.VectorStoreOption) ([]string, error) {
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

		vec := pgvector.NewVector(vectors[i])
		row := pgVectorRow{
			ID:        id,
			Embedding: vec,
			Document:  doc.PageContent,
			Namespace: ns,
			Metadata:  meta,
		}

		if err := s.db.WithContext(ctx).Table(s.tableName).Create(&row).Error; err != nil {
			return nil, fmt.Errorf("insert doc %d: %w", i, err)
		}
	}
	return ids, nil
}

func (s *PgVectorStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...interfaces.VectorStoreOption) ([]schema.Document, error) {
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
	vec := pgvector.NewVector(queryVec)

	// 使用余弦相似度 <=> 或 内积 <#>，这里用 <-> 欧氏距离，1-distance 近似相似度
	// pgvector 的 <=> 是余弦距离，1 - <=> 为相似度
	var rows []struct {
		Document string  `gorm:"column:document"`
		Metadata string  `gorm:"column:metadata"`
		Score    float64 `gorm:"column:score"`
	}

	// pgvector: <=> 为余弦距离，1 - distance 为相似度
	sql := fmt.Sprintf(`
		SELECT document, metadata, (1 - (embedding <=> $1::vector))::float as score
		FROM %s
		WHERE namespace = $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`, s.tableName)

	// GORM Raw: $1=vector, $2=namespace, $3=limit. pgvector.Vector 实现 driver.Valuer
	err = s.db.WithContext(ctx).Raw(sql, vec, ns, numDocuments).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	docs := make([]schema.Document, len(rows))
	for i, r := range rows {
		docs[i] = schema.Document{
			PageContent: r.Document,
			Metadata:    parseMetadata(r.Metadata),
			Score:       float32(r.Score),
		}
	}
	ApplyConfiguredRankBoostToScores(docs)
	return docs, nil
}

// keywordRelaxedTokensPG 宽松全文重试：按空白切词，至少命中一词的 ILIKE 即召回（对齐 Ragflow 降低 min_match 思路）
func keywordRelaxedTokensPG(q string) []string {
	parts := strings.Fields(q)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 && q != "" {
		return []string{q}
	}
	return out
}

func escapeILikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (s *PgVectorStore) KeywordSearch(ctx context.Context, query string, numDocuments int, options ...interfaces.VectorStoreOption) ([]schema.Document, error) {
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
	var rows []struct {
		Document string  `gorm:"column:document"`
		Metadata string  `gorm:"column:metadata"`
		Score    float64 `gorm:"column:score"`
	}
	var err error
	if opts.RelaxedKeywordSearch {
		toks := keywordRelaxedTokensPG(q)
		if len(toks) == 0 {
			return nil, nil
		}
		args := []any{ns}
		clauses := make([]string, 0, len(toks))
		for _, t := range toks {
			pat := "%" + escapeILikePattern(t) + "%"
			args = append(args, pat)
			clauses = append(clauses, fmt.Sprintf("document ILIKE $%d", len(args)))
		}
		args = append(args, numDocuments)
		lim := len(args)
		sql := fmt.Sprintf(`
		SELECT document, metadata, 0.5::float AS score
		FROM %s
		WHERE namespace = $1 AND (%s)
		ORDER BY length(document) ASC
		LIMIT $%d
	`, s.tableName, strings.Join(clauses, " OR "), lim)
		err = s.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error
	} else {
		pattern := "%" + escapeILikePattern(q) + "%"
		sql := fmt.Sprintf(`
		SELECT document, metadata, (
			GREATEST(
				COALESCE(ts_rank(to_tsvector('simple', document), plainto_tsquery('simple', $1)), 0),
				CASE WHEN document ILIKE $2 THEN 0.05 ELSE 0 END
			)
		)::float AS score
		FROM %s
		WHERE namespace = $3 AND (
			(to_tsvector('simple', document) @@ plainto_tsquery('simple', $1))
			OR (document ILIKE $2)
		)
		ORDER BY score DESC
		LIMIT $4
	`, s.tableName)
		err = s.db.WithContext(ctx).Raw(sql, q, pattern, ns, numDocuments).Scan(&rows).Error
	}
	if err != nil {
		return nil, err
	}
	docs := make([]schema.Document, len(rows))
	for i, r := range rows {
		docs[i] = schema.Document{
			PageContent: r.Document,
			Metadata:    parseMetadata(r.Metadata),
			Score:       float32(r.Score),
		}
	}
	ApplyConfiguredRankBoostToScores(docs)
	return docs, nil
}

func (s *PgVectorStore) DeleteByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Table(s.tableName).Where("id IN ?", ids).Delete(nil).Error
}

func (s *PgVectorStore) DeleteByNamespace(ctx context.Context, namespace string) error {
	return s.db.WithContext(ctx).Table(s.tableName).Where("namespace = ?", namespace).Delete(nil).Error
}

func (s *PgVectorStore) DeleteByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) error {
	// metadata 为 jsonb，仅支持 document_id 键（防注入）
	if metaKey != "document_id" {
		metaKey = "document_id"
	}
	val := fmt.Sprintf("%v", metaValue)
	q := s.db.WithContext(ctx).Table(s.tableName).Where("namespace = ?", namespace)
	q = q.Where("metadata->>'"+metaKey+"' = ?", val)
	return q.Delete(nil).Error
}

func (s *PgVectorStore) ListByMetadata(ctx context.Context, namespace, metaKey string, metaValue any) ([]schema.Document, error) {
	val := fmt.Sprintf("%v", metaValue)
	var rows []struct {
		Document string `gorm:"column:document"`
		Metadata string `gorm:"column:metadata"`
	}
	sql := fmt.Sprintf(`
		SELECT document, metadata FROM %s
		WHERE namespace = $1 AND metadata->>'%s' = $2
		ORDER BY (metadata->>'chunk_index')::int ASC
	`, s.tableName, metaKey)
	if err := s.db.WithContext(ctx).Raw(sql, namespace, val).Scan(&rows).Error; err != nil {
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

func (s *PgVectorStore) ProviderName() string { return "postgresql" }

func parseMetadata(s string) map[string]any {
	if s == "" || s == "{}" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	return m
}
