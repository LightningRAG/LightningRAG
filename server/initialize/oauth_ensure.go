package initialize

import (
	"errors"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	adapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// oauthApis 与 source/system/api.go 中 OAuth 条目一致，供已有库补全 sys_apis
var oauthApis = []system.SysApi{
	{ApiGroup: "OAuth provider", Method: "POST", Path: "/sysOAuthProvider/createOAuthProvider", Description: "Create OAuth provider"},
	{ApiGroup: "OAuth provider", Method: "DELETE", Path: "/sysOAuthProvider/deleteOAuthProvider", Description: "Delete OAuth provider"},
	{ApiGroup: "OAuth provider", Method: "DELETE", Path: "/sysOAuthProvider/deleteOAuthProviderByIds", Description: "Batch delete OAuth providers"},
	{ApiGroup: "OAuth provider", Method: "PUT", Path: "/sysOAuthProvider/updateOAuthProvider", Description: "Update OAuth provider"},
	{ApiGroup: "OAuth provider", Method: "GET", Path: "/sysOAuthProvider/findOAuthProvider", Description: "Get OAuth provider by ID"},
	{ApiGroup: "OAuth provider", Method: "GET", Path: "/sysOAuthProvider/getOAuthProviderList", Description: "OAuth provider list"},
	{ApiGroup: "OAuth provider", Method: "GET", Path: "/sysOAuthProvider/getRegisteredOAuthKinds", Description: "Registered OAuth kinds"},
	{ApiGroup: "OAuth setting", Method: "GET", Path: "/sysOAuthSetting/getOAuthSetting", Description: "Get OAuth global settings"},
	{ApiGroup: "OAuth setting", Method: "PUT", Path: "/sysOAuthSetting/updateOAuthSetting", Description: "Update OAuth global settings"},
}

// oauthCasbinRules 与 source/system/casbin.go 中 OAuth 策略一致，并补全测试角色 9528（其菜单含系统工具子项时应对应 API 权限）
var oauthCasbinRules = []adapter.CasbinRule{
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/createOAuthProvider", V2: "POST"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/deleteOAuthProvider", V2: "DELETE"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/deleteOAuthProviderByIds", V2: "DELETE"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/updateOAuthProvider", V2: "PUT"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/findOAuthProvider", V2: "GET"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/getOAuthProviderList", V2: "GET"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthProvider/getRegisteredOAuthKinds", V2: "GET"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthSetting/getOAuthSetting", V2: "GET"},
	{Ptype: "p", V0: "888", V1: "/sysOAuthSetting/updateOAuthSetting", V2: "PUT"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/createOAuthProvider", V2: "POST"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/deleteOAuthProvider", V2: "DELETE"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/deleteOAuthProviderByIds", V2: "DELETE"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/updateOAuthProvider", V2: "PUT"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/findOAuthProvider", V2: "GET"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/getOAuthProviderList", V2: "GET"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthProvider/getRegisteredOAuthKinds", V2: "GET"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthSetting/getOAuthSetting", V2: "GET"},
	{Ptype: "p", V0: "9528", V1: "/sysOAuthSetting/updateOAuthSetting", V2: "PUT"},
}

// EnsureOAuthApis 确保 OAuth 管理 API 已写入 sys_apis（已有库未跑过含 OAuth 的 init 时补齐）
func EnsureOAuthApis() {
	if global.LRAG_DB == nil {
		return
	}
	for _, api := range oauthApis {
		apiCopy := api
		if err := global.LRAG_DB.Where("path = ? AND method = ?", apiCopy.Path, apiCopy.Method).
			FirstOrCreate(&apiCopy).Error; err != nil {
			global.LRAG_LOG.Warn("EnsureOAuthApis: FirstOrCreate failed",
				zap.String("path", apiCopy.Path),
				zap.String("method", apiCopy.Method),
				zap.Error(err))
		}
	}
}

// EnsureOAuthCasbin 确保 OAuth 相关 Casbin 规则存在（888 与 9528）
func EnsureOAuthCasbin() {
	if global.LRAG_DB == nil {
		return
	}
	deduplicateCasbinRules()
	for _, rule := range oauthCasbinRules {
		ruleCopy := rule
		if err := global.LRAG_DB.
			Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).
			FirstOrCreate(&ruleCopy).Error; err != nil {
			global.LRAG_LOG.Error("EnsureOAuthCasbin: create rule failed", zap.String("path", rule.V1), zap.Error(err))
		}
	}
}

// EnsureOAuthQuickLoginMenu 确保「第三方快捷登录」菜单挂在系统工具下、组件路径正确，并关联超级管理员(888)；测试角色(9528)若存在则一并关联。
// 解决：代码已合并但历史库未插入菜单、或菜单仍挂在超级管理员下、或 component 仍为旧路径。
func EnsureOAuthQuickLoginMenu() {
	if global.LRAG_DB == nil {
		return
	}
	db := global.LRAG_DB
	deduplicateAuthorityMenus(db, 888)

	var tools system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "systemTools").First(&tools).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.LRAG_LOG.Debug("EnsureOAuthQuickLoginMenu: systemTools parent missing, skip")
		}
		return
	}

	meta := system.Meta{
		Title:    system.DefaultMenuTitleEnglish("oauthProvider"),
		TitleKey: system.MenuTitleKeyForRouteName("oauthProvider"),
		Icon:     "key",
	}

	var m system.SysBaseMenu
	err := db.Where("name = ?", "oauthProvider").First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m = system.SysBaseMenu{
			MenuLevel: 1,
			ParentId:  tools.ID,
			Path:      "oauthProvider",
			Name:      "oauthProvider",
			Hidden:    false,
			Component: "view/oauth/settings.vue",
			Sort:      10,
			Meta:      meta,
		}
		if err := db.Create(&m).Error; err != nil {
			global.LRAG_LOG.Warn("EnsureOAuthQuickLoginMenu: create menu failed", zap.Error(err))
			return
		}
	} else if err != nil {
		global.LRAG_LOG.Warn("EnsureOAuthQuickLoginMenu: query menu failed", zap.Error(err))
		return
	} else {
		up := map[string]interface{}{
			"parent_id":  tools.ID,
			"path":       "oauthProvider",
			"name":       "oauthProvider",
			"hidden":     false,
			"component":  "view/oauth/settings.vue",
			"sort":       10,
			"title":      meta.Title,
			"title_key":  meta.TitleKey,
			"icon":       meta.Icon,
		}
		if err := db.Model(&m).Updates(up).Error; err != nil {
			global.LRAG_LOG.Warn("EnsureOAuthQuickLoginMenu: update menu failed", zap.Error(err))
		}
	}

	appendMenuToAuthorities(db, m, []uint{888, 9528})
}

func appendMenuToAuthorities(db *gorm.DB, menu system.SysBaseMenu, authorityIDs []uint) {
	for _, aid := range authorityIDs {
		var authority system.SysAuthority
		if err := db.Where("authority_id = ?", aid).Preload("SysBaseMenus").First(&authority).Error; err != nil {
			continue
		}
		has := false
		for _, existing := range authority.SysBaseMenus {
			if existing.ID == menu.ID {
				has = true
				break
			}
		}
		if has {
			continue
		}
		if err := db.Model(&authority).Association("SysBaseMenus").Append(&menu); err != nil {
			global.LRAG_LOG.Warn("EnsureOAuthQuickLoginMenu: append menu to authority failed",
				zap.Uint("authority_id", aid), zap.Error(err))
		}
	}
}
