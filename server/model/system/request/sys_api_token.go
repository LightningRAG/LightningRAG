package request

import (
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/system"
)

type SysApiTokenSearch struct {
	system.SysApiToken
	request.PageInfo
	Status *bool `json:"status" form:"status"`
}
