package request

import (
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
)

// KnowledgeBaseCreate 创建知识库
type KnowledgeBaseCreate struct {
	Name                string `json:"name" binding:"required"`
	Description         string `json:"description"`
	EmbeddingID         uint   `json:"embeddingId" binding:"required"`
	EmbeddingSource     string `json:"embeddingSource"`
	VectorStoreID       uint   `json:"vectorStoreId" binding:"required"`
	FileStorageID       uint   `json:"fileStorageId" binding:"required"`
	RetrieverType       string `json:"retrieverType"`
	ChunkMethod         string `json:"chunkMethod"`
	ChunkSize           int    `json:"chunkSize"`
	ChunkOverlap        int    `json:"chunkOverlap"`
	ConcurrentSliceJobs int    `json:"concurrentSliceJobs"`
	Delimiter           string `json:"delimiter"`
	AutoKeywords        int    `json:"autoKeywords"`
	AutoQuestions       int    `json:"autoQuestions"`
	UseRerank           bool   `json:"useRerank"`
	RerankID            uint   `json:"rerankId"`
	RerankSource        string `json:"rerankSource"`
	RerankTopK          int    `json:"rerankTopK"`
	// PageIndex LLM
	PageIndexLLMID     uint   `json:"pageIndexLlmId"`
	PageIndexLLMSource string `json:"pageIndexLlmSource"`
	// OCR
	UseOCR    bool   `json:"useOcr"`
	OCRID     uint   `json:"ocrId"`
	OCRSource string `json:"ocrSource"`
	// CV 图片描述
	UseImageDescription    bool   `json:"useImageDescription"`
	ImageDescriptionID     uint   `json:"imageDescriptionId"`
	ImageDescriptionSource string `json:"imageDescriptionSource"`
	// Speech2Text
	UseSpeech2Text    bool   `json:"useSpeech2Text"`
	Speech2TextID     uint   `json:"speech2TextId"`
	Speech2TextSource string `json:"speech2TextSource"`
	// 知识图谱（LightRAG 风格实体/关系抽取与检索）；省略时服务端默认开启
	EnableKnowledgeGraph *bool `json:"enableKnowledgeGraph"`
}

// KnowledgeBaseList 知识库列表
type KnowledgeBaseList struct {
	request.PageInfo
}

// KnowledgeBaseGet 获取知识库
type KnowledgeBaseGet struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// KnowledgeBaseUpdate 更新知识库（向量存储、文件存储创建后不可更改；嵌入模型仅可切换同维度模型）
type KnowledgeBaseUpdate struct {
	ID                  uint    `json:"id" binding:"required"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	EmbeddingID         uint    `json:"embeddingId"` // 仅允许切换为同维度的嵌入模型
	EmbeddingSource     *string `json:"embeddingSource"`
	RetrieverType       string  `json:"retrieverType"`
	ChunkMethod         string  `json:"chunkMethod"`
	ChunkSize           int     `json:"chunkSize"`
	ChunkOverlap        int     `json:"chunkOverlap"`
	ConcurrentSliceJobs int     `json:"concurrentSliceJobs"`
	Delimiter           string  `json:"delimiter"`
	AutoKeywords        int     `json:"autoKeywords"`
	AutoQuestions       int     `json:"autoQuestions"`
	UseRerank           *bool   `json:"useRerank"`
	RerankID            *uint   `json:"rerankId"`
	RerankSource        *string `json:"rerankSource"`
	RerankTopK          *int    `json:"rerankTopK"`
	// PageIndex LLM
	PageIndexLLMID     *uint   `json:"pageIndexLlmId"`
	PageIndexLLMSource *string `json:"pageIndexLlmSource"`
	// OCR
	UseOCR    *bool   `json:"useOcr"`
	OCRID     *uint   `json:"ocrId"`
	OCRSource *string `json:"ocrSource"`
	// CV 图片描述
	UseImageDescription    *bool   `json:"useImageDescription"`
	ImageDescriptionID     *uint   `json:"imageDescriptionId"`
	ImageDescriptionSource *string `json:"imageDescriptionSource"`
	// Speech2Text
	UseSpeech2Text    *bool   `json:"useSpeech2Text"`
	Speech2TextID     *uint   `json:"speech2TextId"`
	Speech2TextSource *string `json:"speech2TextSource"`
	// 知识图谱
	EnableKnowledgeGraph *bool `json:"enableKnowledgeGraph"`
}

// KnowledgeBaseDelete 删除知识库
type KnowledgeBaseDelete struct {
	ID uint `json:"id" binding:"required"`
}

// KnowledgeBaseKnowledgeGraph 获取知识库图谱可视化数据（子集，避免超大库拖垮浏览器）
type KnowledgeBaseKnowledgeGraph struct {
	ID uint `json:"id" binding:"required"`
	// MaxEntities 最多返回实体数；0 表示使用服务端默认
	MaxEntities int `json:"maxEntities"`
	// MaxRelationships 最多返回关系数（两端均在已选实体集合内）；0 表示使用服务端默认
	MaxRelationships int `json:"maxRelationships"`
	// DocumentID 仅返回与该文档切片关联的实体/关系（须属于该知识库）；0 表示不按文档筛选
	DocumentID uint `json:"documentId"`
}

// KnowledgeBaseShare 分享知识库
type KnowledgeBaseShare struct {
	ID         uint   `json:"id" binding:"required"`
	TargetType string `json:"targetType" binding:"required"` // user|role|org
	TargetID   uint   `json:"targetId" binding:"required"`
	Permission string `json:"permission"` // read|write|admin
}

// KnowledgeBaseTransfer 转让知识库
type KnowledgeBaseTransfer struct {
	ID       uint `json:"id" binding:"required"`
	TargetID uint `json:"targetId" binding:"required"` // 目标用户ID
}

// KnowledgeBaseRetrieve 在知识库中测试检索切片（选库 + query + topN）
type KnowledgeBaseRetrieve struct {
	KnowledgeBaseID  uint   `json:"knowledgeBaseId"`  // 单库；与 knowledgeBaseIds 二选一或同时传（合并去重）
	KnowledgeBaseIDs []uint `json:"knowledgeBaseIds"` // 多库
	Query            string `json:"query" binding:"required"`
	TopN             int    `json:"topN"` // 最终返回条数，对齐 chunk_top_k；默认 8，最大由服务配置限制
	// ChunkTopK 与 TopN 同义（对齐 LightRAG chunk_top_k 字段名）；仅当 topN<=0 时生效
	ChunkTopK int `json:"chunkTopK"`
	// TopK 候选池上限，对齐 LightningRAG top_k；大于 topN 时先按 topK 检索再截断到 topN
	TopK *int `json:"topK"`
	// Mode 单次请求覆盖各库的 RetrieverType（LightningRAG API 的 mode；naive 与 vector 等价）；空则使用知识库配置
	Mode string `json:"mode"`
	// EnableRerank 单次请求是否 Rerank（对齐 LightningRAG enable_rerank）；nil=按知识库
	EnableRerank *bool `json:"enableRerank"`
	// HlKeywords / LlKeywords 对齐 LightningRAG，拼入检索 query
	HlKeywords          []string `json:"hlKeywords"`
	LlKeywords          []string `json:"llKeywords"`
	IncludeReferences   *bool    `json:"includeReferences"`
	IncludeChunkContent *bool    `json:"includeChunkContent"`
	CosineThreshold     *float32 `json:"cosineThreshold"`
	MinRerankScore      *float32 `json:"minRerankScore"`
	// MaxRagContextTokens 返回前按粗估 token 裁剪切片正文；0 或未传可用服务端 default-max-rag-context-tokens
	MaxRagContextTokens *uint `json:"maxRagContextTokens"`
	// TocEnhance 对齐 references/ragflow toc_enhance：PageIndex 知识库是否启用「向量召回 + TOC」混合；nil=默认启用（与 Ragflow 默认关闭不同，保持兼容）；false=纯目录/树推理
	TocEnhance *bool `json:"tocEnhance,omitempty"`
	// TocEnhanceRagflow 与 Ragflow toc_enhance 同义；与 TocEnhance 同时传时以 TocEnhance 为准
	TocEnhanceRagflow *bool `json:"toc_enhance,omitempty"`
}

// DocumentList 文档列表
type DocumentList struct {
	KnowledgeBaseID uint `json:"knowledgeBaseId" form:"knowledgeBaseId" binding:"required"`
	request.PageInfo
}

// DocumentGet 获取文档详情
type DocumentGet struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DocumentDelete 删除文档
type DocumentDelete struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DocumentRetry 重试解析文档
type DocumentRetry struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DocumentDownload 下载/预览文档
type DocumentDownload struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// ChunkList 文档切片列表
type ChunkList struct {
	DocumentID uint `json:"documentId" form:"documentId" binding:"required"`
	request.PageInfo
}

// ChunkUpdate 更新切片内容
type ChunkUpdate struct {
	ID      uint   `json:"id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// DocumentBatchByIDs 批量操作文档（需属于同一知识库）
type DocumentBatchByIDs struct {
	KnowledgeBaseID uint   `json:"knowledgeBaseId" binding:"required"`
	DocumentIDs     []uint `json:"documentIds" binding:"required"`
}

// DocumentBatchRetrieval 批量设置文档是否参与 RAG 检索
type DocumentBatchRetrieval struct {
	KnowledgeBaseID uint   `json:"knowledgeBaseId" binding:"required"`
	DocumentIDs     []uint `json:"documentIds" binding:"required"`
	Enabled         bool   `json:"enabled"`
}

// DocumentBatchPriority 批量设置文档检索权重（0~1），已向量化的文档会异步重写入向量 metadata
type DocumentBatchPriority struct {
	KnowledgeBaseID uint    `json:"knowledgeBaseId" binding:"required"`
	DocumentIDs     []uint  `json:"documentIds" binding:"required"`
	Priority        float64 `json:"priority" binding:"gte=0,lte=1"`
}
