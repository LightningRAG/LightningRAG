package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/google/uuid"
)

// 切片方法常量
const (
	ChunkMethodGeneral      = "general"      // 通用：段落/句子感知分块
	ChunkMethodQA           = "qa"           // 问答对：按 Q/A 模式提取
	ChunkMethodBook         = "book"         // 书籍：按标题层级分块
	ChunkMethodPaper        = "paper"        // 论文：按章节结构分块
	ChunkMethodLaws         = "laws"         // 法律：按条款层级分块
	ChunkMethodPresentation = "presentation" // 演示文稿：按页/幻灯片分块
	ChunkMethodTable        = "table"        // 表格：按行分块
	ChunkMethodOne          = "one"          // 整文档：整篇文档作为一个切片
)

// RagKnowledgeBase 知识库
type RagKnowledgeBase struct {
	global.LRAG_MODEL
	UUID                 uuid.UUID `json:"uuid" gorm:"index;comment:知识库UUID"`
	Name                 string    `json:"name" gorm:"size:128;comment:知识库名称"`
	Description          string    `json:"description" gorm:"type:text;comment:描述"`
	OwnerID              uint      `json:"ownerId" gorm:"index;comment:所有者用户ID"`
	EmbeddingID          uint      `json:"embeddingId" gorm:"comment:嵌入模型配置ID"`
	EmbeddingSource      string    `json:"embeddingSource" gorm:"size:16;default:legacy;comment:嵌入模型来源 legacy|admin|user"`
	VectorStoreID        uint      `json:"vectorStoreId" gorm:"comment:向量存储配置ID"`
	FileStorageID        uint      `json:"fileStorageId" gorm:"comment:文件存储配置ID"`
	RetrieverType        string    `json:"retrieverType" gorm:"size:32;default:vector;comment:检索类型 vector|keyword|local|global|hybrid|mix|bypass|pageindex；naive 已并入 vector（兼容旧值）"`
	EnableKnowledgeGraph bool      `json:"enableKnowledgeGraph" gorm:"default:true;comment:启用后索引完成时抽取实体/关系；local/global/hybrid/mix 在存在图谱数据时走图谱检索"`
	ChunkMethod          string    `json:"chunkMethod" gorm:"size:32;default:general;comment:切片方法 general|qa|book|paper|laws|presentation|table|one"`
	ChunkSize            int       `json:"chunkSize" gorm:"default:500;comment:每块最大字符数"`
	ChunkOverlap         int       `json:"chunkOverlap" gorm:"default:50;comment:切片重叠字符数"`
	ConcurrentSliceJobs  int       `json:"concurrentSliceJobs" gorm:"default:1;comment:同时进行的文档切片任务数(每知识库)"`
	Delimiter            string    `json:"delimiter" gorm:"size:128;default:\\n!?。；！？;comment:文本分段标识符"`
	AutoKeywords         int       `json:"autoKeywords" gorm:"default:0;comment:自动生成关键词数量(0-30)"`
	AutoQuestions        int       `json:"autoQuestions" gorm:"default:0;comment:自动生成问题数量(0-10)"`
	UseRerank            bool      `json:"useRerank" gorm:"default:false;comment:是否启用 Rerank 重排序"`
	RerankID             uint      `json:"rerankId" gorm:"default:0;comment:Rerank 模型 ID"`
	RerankSource         string    `json:"rerankSource" gorm:"size:16;default:admin;comment:Rerank 模型来源 admin|user"`
	RerankTopK           int       `json:"rerankTopK" gorm:"default:0;comment:Rerank 候选文档数量(0=自动)"`

	// PageIndex 推理检索用 LLM
	PageIndexLLMID     uint   `json:"pageIndexLlmId" gorm:"default:0;comment:PageIndex 推理检索 LLM ID(0=自动选择)"`
	PageIndexLLMSource string `json:"pageIndexLlmSource" gorm:"size:16;default:admin;comment:PageIndex LLM 来源 admin|user"`

	// OCR（图片/扫描件 PDF 解析）
	UseOCR    bool   `json:"useOcr" gorm:"default:true;comment:是否启用 OCR"`
	OCRID     uint   `json:"ocrId" gorm:"default:0;comment:OCR 模型 ID(0=自动选择)"`
	OCRSource string `json:"ocrSource" gorm:"size:16;default:admin;comment:OCR 模型来源 admin|user"`

	// CV 图片理解/描述
	UseImageDescription    bool   `json:"useImageDescription" gorm:"default:false;comment:是否启用图片描述(CV 模型)"`
	ImageDescriptionID     uint   `json:"imageDescriptionId" gorm:"default:0;comment:CV 模型 ID(0=自动选择)"`
	ImageDescriptionSource string `json:"imageDescriptionSource" gorm:"size:16;default:admin;comment:CV 模型来源 admin|user"`

	// Speech2Text 音频转文字
	UseSpeech2Text    bool   `json:"useSpeech2Text" gorm:"default:false;comment:是否启用语音转文字"`
	Speech2TextID     uint   `json:"speech2TextId" gorm:"default:0;comment:Speech2Text 模型 ID(0=自动选择)"`
	Speech2TextSource string `json:"speech2TextSource" gorm:"size:16;default:admin;comment:Speech2Text 模型来源 admin|user"`
}

func (RagKnowledgeBase) TableName() string {
	return "rag_knowledge_bases"
}
