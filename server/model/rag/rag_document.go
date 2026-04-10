package rag

import (
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/google/uuid"
)

// RagDocument 知识库内的文档/文件
type RagDocument struct {
	global.LRAG_MODEL
	UUID               uuid.UUID  `json:"uuid" gorm:"index;comment:文档UUID"`
	KnowledgeBaseID    uint       `json:"knowledgeBaseId" gorm:"index;comment:所属知识库ID"`
	Name               string     `json:"name" gorm:"size:256;comment:文件名"`
	FileType           string     `json:"fileType" gorm:"size:32;comment:文件类型 pdf|docx|txt|md 等"`
	FileSize           int64      `json:"fileSize" gorm:"comment:文件大小(字节)"`
	StoragePath        string     `json:"storagePath" gorm:"size:512;comment:存储路径"`
	Status             string     `json:"status" gorm:"size:32;default:processing;comment:状态 processing|completed|failed|cancelled"`
	RetrievalEnabled   bool       `json:"retrievalEnabled" gorm:"default:true;comment:是否参与RAG检索"`
	Priority           float64    `json:"priority" gorm:"default:0;comment:文档检索权重 0~1，索引入库时作为切片 rank_boost 下限"`
	ChunkCount         int        `json:"chunkCount" gorm:"default:0;comment:切片数量"`
	TokenCount         int        `json:"tokenCount" gorm:"default:0;comment:token数量(估算)"`
	ErrorMsg           string     `json:"errorMsg" gorm:"type:text;comment:错误信息"`
	Thumbnail          string     `json:"thumbnail" gorm:"type:longtext;comment:缩略图 base64"`
	PageIndexStructure string     `json:"pageIndexStructure" gorm:"type:longtext;comment:PageIndex 树结构 JSON，用于推理检索"`
	IndexingClaimOwner string     `json:"-" gorm:"size:64;index;comment:切片任务占用实例ID"`
	IndexingClaimUntil *time.Time `json:"-" gorm:"index;comment:切片租约到期时间"`
}

func (RagDocument) TableName() string {
	return "rag_documents"
}
