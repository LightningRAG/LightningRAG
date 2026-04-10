package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

type JwtBlacklist struct {
	global.LRAG_MODEL
	Jwt string `gorm:"type:text;comment:jwt"`
}
