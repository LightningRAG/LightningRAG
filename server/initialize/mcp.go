package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	mcpTool "github.com/LightningRAG/LightningRAG/server/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func McpRun() *server.SSEServer {
	config := global.LRAG_CONFIG.MCP

	s := server.NewMCPServer(
		config.Name,
		config.Version,
	)

	global.LRAG_MCP_SERVER = s

	mcpTool.RegisterAllTools(s)

	return server.NewSSEServer(s,
		server.WithSSEEndpoint(config.SSEPath),
		server.WithMessageEndpoint(config.MessagePath),
		server.WithBaseURL(config.UrlPrefix))
}
