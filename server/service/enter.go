package service

import (
	"github.com/LightningRAG/LightningRAG/server/service/example"
	"github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup
	RagServiceGroup     rag.ServiceGroup
}
