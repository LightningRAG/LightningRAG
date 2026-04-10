package rag

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/google/uuid"
)

// RagChunk 文档切片，存储于向量库或关系库
// 向量库中存 embedding，关系库中存元数据
type RagChunk struct {
	global.LRAG_MODEL
	UUID          uuid.UUID `json:"uuid" gorm:"index;comment:切片UUID"`
	DocumentID    uint      `json:"documentId" gorm:"index;comment:所属文档ID"`
	Content       string    `json:"content" gorm:"type:text;comment:切片文本内容"`
	VectorStoreID string    `json:"vectorStoreId" gorm:"size:64;comment:向量库中的ID"`
	PageIndex     int       `json:"pageIndex" gorm:"comment:页码(如支持)"`
	ChunkIndex    int       `json:"chunkIndex" gorm:"comment:切片序号"`
	Metadata      string    `json:"metadata" gorm:"type:text;comment:JSON元数据"`
}

func (RagChunk) TableName() string {
	return "rag_chunks"
}
