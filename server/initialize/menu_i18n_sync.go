package initialize

import (
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	"go.uber.org/zap"
)

// SyncMenuTitleKeys 按路由 Name 回写 meta.title_key，使已有数据库在启动后也能获得 i18n 键（与前端 menu.names.* 对齐）
func SyncMenuTitleKeys() {
	if global.LRAG_DB == nil {
		return
	}
	db := global.LRAG_DB
	var menus []system.SysBaseMenu
	if err := db.Find(&menus).Error; err != nil {
		global.LRAG_LOG.Debug("SyncMenuTitleKeys: list menus failed", zap.Error(err))
		return
	}
	for _, m := range menus {
		key := system.MenuTitleKeyForRouteName(m.Name)
		tk := strings.TrimSpace(m.Meta.TitleKey)
		// 补全空 title_key；或修正历史错误（把整段 URL 拼进 i18n 键名，导致前端显示 menu.names.https://...）
		legacyBroken := strings.Contains(tk, "://")
		if tk != "" && !legacyBroken && tk == key {
			continue
		}
		if tk != "" && !legacyBroken {
			continue
		}
		if err := db.Model(&system.SysBaseMenu{}).Where("id = ?", m.ID).Update("title_key", key).Error; err != nil {
			global.LRAG_LOG.Debug("SyncMenuTitleKeys: update failed", zap.Uint("id", m.ID), zap.Error(err))
		}
	}
}

// SyncMenuEnglishDefaultTitles 将 meta.title 回写为英文缺省值（与前端 en.js menu.names 一致），
// 仅当 title_key 与按 Name 推导的标准键一致时更新，避免覆盖自定义 title_key 的菜单。
func SyncMenuEnglishDefaultTitles() {
	if global.LRAG_DB == nil {
		return
	}
	db := global.LRAG_DB
	var menus []system.SysBaseMenu
	if err := db.Find(&menus).Error; err != nil {
		global.LRAG_LOG.Debug("SyncMenuEnglishDefaultTitles: list menus failed", zap.Error(err))
		return
	}
	for _, m := range menus {
		wantKey := system.MenuTitleKeyForRouteName(m.Name)
		if strings.TrimSpace(m.Meta.TitleKey) != wantKey {
			continue
		}
		en := system.DefaultMenuTitleEnglish(m.Name)
		if en == "" {
			continue
		}
		if m.Meta.Title == en {
			continue
		}
		if err := db.Model(&system.SysBaseMenu{}).Where("id = ?", m.ID).Update("title", en).Error; err != nil {
			global.LRAG_LOG.Debug("SyncMenuEnglishDefaultTitles: update failed", zap.Uint("id", m.ID), zap.Error(err))
		}
	}
}
