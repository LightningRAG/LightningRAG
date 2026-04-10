package response

import (
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
)

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}
