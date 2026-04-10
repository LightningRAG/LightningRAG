package rag

import "github.com/LightningRAG/LightningRAG/server/global"

// RagKgEntity 知识库级实体（对齐 LightRAG 知识图谱中的 entity 节点）
type RagKgEntity struct {
	global.LRAG_MODEL
	KnowledgeBaseID uint   `json:"knowledgeBaseId" gorm:"uniqueIndex:idx_kb_kg_entity_norm,priority:1;comment:知识库ID"`
	Name            string `json:"name" gorm:"size:512;comment:实体名称"`
	NormalizedName  string `json:"normalizedName" gorm:"size:512;uniqueIndex:idx_kb_kg_entity_norm,priority:2;comment:规范化名(小写去空格)用于去重"`
	EntityType      string `json:"entityType" gorm:"size:128;comment:实体类型"`
	Description     string `json:"description" gorm:"type:text;comment:实体描述"`
	VectorStoreID   string `json:"vectorStoreId" gorm:"size:64;comment:实体向量在向量库中的行ID"`
}

func (RagKgEntity) TableName() string { return "rag_kg_entities" }

// RagKgRelationship 实体间关系
type RagKgRelationship struct {
	global.LRAG_MODEL
	KnowledgeBaseID uint   `json:"knowledgeBaseId" gorm:"index;comment:知识库ID"`
	SourceEntityID  uint   `json:"sourceEntityId" gorm:"index;comment:源实体ID"`
	TargetEntityID  uint   `json:"targetEntityId" gorm:"index;comment:目标实体ID"`
	Keywords        string `json:"keywords" gorm:"type:text;comment:关系关键词(逗号分隔)"`
	Description     string `json:"description" gorm:"type:text;comment:关系描述"`
	VectorStoreID   string `json:"vectorStoreId" gorm:"size:64;comment:关系向量在向量库中的行ID"`
}

func (RagKgRelationship) TableName() string { return "rag_kg_relationships" }

// RagKgEntityChunk 实体与文本切片的关联（标明该实体从哪些 chunk 抽取）
type RagKgEntityChunk struct {
	EntityID uint `json:"entityId" gorm:"primaryKey;comment:实体ID"`
	ChunkID  uint `json:"chunkId" gorm:"primaryKey;index;comment:切片ID"`
}

func (RagKgEntityChunk) TableName() string { return "rag_kg_entity_chunks" }

// RagKgRelationshipChunk 关系与文本切片的关联
type RagKgRelationshipChunk struct {
	RelationshipID uint `json:"relationshipId" gorm:"primaryKey;comment:关系ID"`
	ChunkID        uint `json:"chunkId" gorm:"primaryKey;index;comment:切片ID"`
}

func (RagKgRelationshipChunk) TableName() string { return "rag_kg_relationship_chunks" }
