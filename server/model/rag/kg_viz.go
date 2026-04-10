package rag

// KnowledgeGraphVizEntity 知识图谱可视化中的实体节点
type KnowledgeGraphVizEntity struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	EntityType  string `json:"entityType"`
	Description string `json:"description"`
}

// KnowledgeGraphVizRelationship 知识图谱可视化中的关系边
type KnowledgeGraphVizRelationship struct {
	ID             uint   `json:"id"`
	SourceEntityID uint   `json:"sourceEntityId"`
	TargetEntityID uint   `json:"targetEntityId"`
	Keywords       string `json:"keywords"`
	Description    string `json:"description"`
}

// KnowledgeGraphViz 供前端力导向/图编辑组件渲染的图谱子集（可能截断）
type KnowledgeGraphViz struct {
	KnowledgeBaseID           uint                            `json:"knowledgeBaseId"`
	EnableKnowledgeGraph      bool                            `json:"enableKnowledgeGraph"`
	EntityCount               int64                           `json:"entityCount"`
	RelationshipCount         int64                           `json:"relationshipCount"`
	Truncated                 bool                            `json:"truncated"`
	Entities                  []KnowledgeGraphVizEntity       `json:"entities"`
	Relationships             []KnowledgeGraphVizRelationship `json:"relationships"`
	MaxEntitiesRequested      int                             `json:"maxEntitiesRequested"`
	MaxRelationshipsRequested int                             `json:"maxRelationshipsRequested"`
	ScopeDocumentID           uint                            `json:"scopeDocumentId,omitempty"`
}
