package utils

import (
	"sync"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
)

var (
	casbinMu             sync.Mutex
	syncedCachedEnforcer *casbin.SyncedCachedEnforcer
)

const casbinModelText = `
		[request_definition]
		r = sub, obj, act
		
		[policy_definition]
		p = sub, obj, act
		
		[role_definition]
		g = _, _
		
		[policy_effect]
		e = some(where (p.eft == allow))
		
		[matchers]
		m = r.sub == p.sub && keyMatch2(r.obj,p.obj) && r.act == p.act
		`

// ResetCasbinEnforcer 丢弃当前 Enforcer，使下次 GetCasbin 按当前 LRAG_DB 重建。
// 典型场景：启动时 LRAG_DB 为 nil 导致首次 sync.Once 失败后无法恢复（已改为互斥懒加载仍须在 InitDB/换库后重置适配器）。
func ResetCasbinEnforcer() {
	casbinMu.Lock()
	defer casbinMu.Unlock()
	syncedCachedEnforcer = nil
}

// GetCasbin 获取 casbin 实例；LRAG_DB 未就绪或适配失败时返回 nil。
func GetCasbin() *casbin.SyncedCachedEnforcer {
	casbinMu.Lock()
	defer casbinMu.Unlock()
	if syncedCachedEnforcer != nil {
		return syncedCachedEnforcer
	}
	if global.LRAG_DB == nil {
		return nil
	}
	a, err := gormadapter.NewAdapterByDB(global.LRAG_DB)
	if err != nil {
		global.LRAG_LOG.Error("适配数据库失败请检查casbin表是否为InnoDB引擎!", zap.Error(err))
		return nil
	}
	m, err := model.NewModelFromString(casbinModelText)
	if err != nil {
		global.LRAG_LOG.Error("字符串加载模型失败!", zap.Error(err))
		return nil
	}
	enforcer, err := casbin.NewSyncedCachedEnforcer(m, a)
	if err != nil {
		global.LRAG_LOG.Error("创建 Casbin Enforcer 失败", zap.Error(err))
		return nil
	}
	enforcer.SetExpireTime(60 * 60)
	if err = enforcer.LoadPolicy(); err != nil {
		global.LRAG_LOG.Error("Casbin LoadPolicy 失败", zap.Error(err))
	}
	syncedCachedEnforcer = enforcer
	return syncedCachedEnforcer
}
