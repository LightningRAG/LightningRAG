package v1

import (
	"github.com/LightningRAG/LightningRAG/server/api/v1/example"
	oauthappapi "github.com/LightningRAG/LightningRAG/server/api/v1/oauthapp"
	"github.com/LightningRAG/LightningRAG/server/api/v1/rag"
	"github.com/LightningRAG/LightningRAG/server/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup  system.ApiGroup
	OAuthAppGroup   oauthappapi.ApiPack
	ExampleApiGroup example.ApiGroup
	RagApiGroup     rag.ApiGroup
}
