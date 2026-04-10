package global

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"

	"github.com/LightningRAG/LightningRAG/server/utils/timer"
	"github.com/songzhibin97/gkit/cache/local_cache"

	"golang.org/x/sync/singleflight"

	"go.uber.org/zap"

	"github.com/LightningRAG/LightningRAG/server/config"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	LRAG_DB        *gorm.DB
	LRAG_DBList    map[string]*gorm.DB
	LRAG_REDIS     redis.UniversalClient
	LRAG_REDISList map[string]redis.UniversalClient
	LRAG_MONGO     *qmgo.QmgoClient
	LRAG_CONFIG    config.Server
	LRAG_VP        *viper.Viper
	// LRAG_LOG    *oplogging.Logger
	LRAG_LOG                 *zap.Logger
	LRAG_Timer               timer.Timer = timer.NewTimerTask()
	LRAG_Concurrency_Control             = &singleflight.Group{}
	LRAG_ROUTERS             gin.RoutesInfo
	LRAG_ACTIVE_DBNAME       *string
	LRAG_MCP_SERVER          *server.MCPServer
	BlackCache               local_cache.Cache
	lock                     sync.RWMutex
)

// GetGlobalDBByDBName 通过名称获取db list中的db
func GetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	return LRAG_DBList[dbname]
}

// MustGetGlobalDBByDBName 通过名称获取db 如果不存在则panic
func MustGetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	db, ok := LRAG_DBList[dbname]
	if !ok || db == nil {
		panic("db no init")
	}
	return db
}

func GetRedis(name string) redis.UniversalClient {
	redis, ok := LRAG_REDISList[name]
	if !ok || redis == nil {
		panic(fmt.Sprintf("redis `%s` no init", name))
	}
	return redis
}
