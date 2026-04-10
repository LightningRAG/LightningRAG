package schema

// Document 文档切片，用于 RAG 检索和存储
// 参考 langchaingo schema.Document
type Document struct {
	PageContent string         // 文本内容
	Metadata    map[string]any // 元数据，如 source, chunk_id, doc_id 等
	Score       float32        // 相似度分数（检索时填充）
}
