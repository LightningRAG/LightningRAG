package response

import "github.com/LightningRAG/LightningRAG/server/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}
