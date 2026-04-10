package router

import (
	"github.com/LightningRAG/LightningRAG/server/router/example"
	oauthapprouter "github.com/LightningRAG/LightningRAG/server/router/oauthapp"
	"github.com/LightningRAG/LightningRAG/server/router/rag"
	"github.com/LightningRAG/LightningRAG/server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System   system.RouterGroup
	OAuthApp oauthapprouter.RouterGroup
	Example  example.RouterGroup
	Rag      rag.RouterGroup
}
