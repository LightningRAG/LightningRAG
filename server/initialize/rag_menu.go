package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemService "github.com/LightningRAG/LightningRAG/server/service/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ragAgentMenus 需要确保存在的 RAG Agent 相关菜单（与 source/system/menu.go 保持一致）
// 每次启动时自动执行，幂等设计：使用 FirstOrCreate 避免重复创建，关联前检查避免重复菜单
var ragAgentMenus = []struct {
	parentName string
	menus      []system.SysBaseMenu
}{
	{
		parentName: "rag",
		menus: []system.SysBaseMenu{
			{MenuLevel: 1, Hidden: false, Path: "retrieval", Name: "ragDocumentRetrieval", Component: "view/rag/knowledgeBase/retrieval.vue", Sort: 2, Meta: system.Meta{Title: system.DefaultMenuTitleEnglish("ragDocumentRetrieval"), TitleKey: system.MenuTitleKeyForRouteName("ragDocumentRetrieval"), Icon: "search"}},
		},
	},
	{
		parentName: "rag",
		menus: []system.SysBaseMenu{
			{MenuLevel: 1, Hidden: true, Path: "ragKnowledgeGraph", Name: "ragKnowledgeGraph", Component: "view/rag/knowledgeBase/knowledgeGraph.vue", Sort: 3, Meta: system.Meta{Title: system.DefaultMenuTitleEnglish("ragKnowledgeGraph"), TitleKey: system.MenuTitleKeyForRouteName("ragKnowledgeGraph"), Icon: "histogram"}},
		},
	},
	{
		parentName: "rag",
		menus: []system.SysBaseMenu{
			{MenuLevel: 1, Hidden: false, Path: "agent", Name: "ragAgent", Component: "view/rag/agent/index.vue", Sort: 5, Meta: system.Meta{Title: system.DefaultMenuTitleEnglish("ragAgent"), TitleKey: system.MenuTitleKeyForRouteName("ragAgent"), Icon: "connection"}},
			{MenuLevel: 1, Hidden: true, Path: "agentEditor", Name: "ragAgentEditor", Component: "view/rag/agent/editor.vue", Sort: 6, Meta: system.Meta{Title: system.DefaultMenuTitleEnglish("ragAgentEditor"), TitleKey: system.MenuTitleKeyForRouteName("ragAgentEditor"), Icon: "edit"}},
			{MenuLevel: 1, Hidden: false, Path: "systemModel", Name: "ragSystemModel", Component: "view/rag/systemModel/systemModel.vue", Sort: 7, Meta: system.Meta{Title: system.DefaultMenuTitleEnglish("ragSystemModel"), TitleKey: system.MenuTitleKeyForRouteName("ragSystemModel"), Icon: "setting"}},
			{MenuLevel: 1, Hidden: false, Path: "channelConnector", Name: "ragChannelConnectors", Component: "view/rag/channelConnector/channelConnector.vue", Sort: 8, Meta: system.Meta{Title: system.DefaultMenuTitleEnglish("ragChannelConnectors"), TitleKey: system.MenuTitleKeyForRouteName("ragChannelConnectors"), Icon: "message"}},
		},
	},
}

// deduplicateAuthorityMenus 去重 sys_authority_menus 中同一 authority 下重复的 menu 关联
// 跨库兼容：查询 authority 下所有行，在 Go 中按 menu_id 分组，每组保留一条，删除其余
func deduplicateAuthorityMenus(db *gorm.DB, authorityID uint) {
	var rows []struct {
		MenuID uint `gorm:"column:sys_base_menu_id"`
	}
	if err := db.Table("sys_authority_menus").Where("sys_authority_authority_id = ?", authorityID).Select("sys_base_menu_id").Find(&rows).Error; err != nil {
		return
	}
	menuCounts := make(map[uint]int)
	for _, r := range rows {
		menuCounts[r.MenuID]++
	}
	for menuID, count := range menuCounts {
		if count <= 1 {
			continue
		}
		if err := db.Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", menuID, authorityID).Delete(&system.SysAuthorityMenu{}).Error; err != nil {
			continue
		}
		db.Table("sys_authority_menus").Create(map[string]interface{}{
			"sys_base_menu_id":           menuID,
			"sys_authority_authority_id": authorityID,
		})
	}
}

// EnsureRagMenus 确保 RAG Agent 相关菜单已注册到 sys_base_menus 表
// 每次启动时执行，幂等：FirstOrCreate 避免重复创建，关联前检查避免重复分配
func EnsureRagMenus() {
	if global.LRAG_DB == nil {
		return
	}
	db := global.LRAG_DB
	deduplicateAuthorityMenus(db, 888)

	// 移除已废弃的「知识库设置」菜单（向量存储与文件存储均在「系统配置」）
	var ragParent system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "rag").First(&ragParent).Error; err == nil {
		var oldSettings system.SysBaseMenu
		if err := db.Where("name = ? AND parent_id = ?", "ragSettings", ragParent.ID).First(&oldSettings).Error; err == nil {
			if delErr := systemService.BaseMenuServiceApp.DeleteBaseMenu(int(oldSettings.ID)); delErr != nil {
				global.LRAG_LOG.Debug("EnsureRagMenus: remove deprecated ragSettings menu", zap.Error(delErr))
			}
		}
	}

	for _, group := range ragAgentMenus {
		var parentMenu system.SysBaseMenu
		// parent_id = 0 确保取到顶层 rag 菜单，避免误匹配子菜单
		if err := db.Where("name = ? AND parent_id = 0", group.parentName).First(&parentMenu).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue // 父菜单不存在，说明尚未完成完整初始化
			}
			continue
		}

		var ensuredMenus []system.SysBaseMenu
		for _, menu := range group.menus {
			menu.ParentId = parentMenu.ID
			menuCopy := menu
			// 使用 (name, parent_id) 精确匹配，确保只复用 rag 父级下的菜单
			if err := db.Where("name = ? AND parent_id = ?", menuCopy.Name, parentMenu.ID).FirstOrCreate(&menuCopy).Error; err != nil {
				continue
			}
			ensuredMenus = append(ensuredMenus, menuCopy)
		}

		if len(ensuredMenus) == 0 {
			continue
		}

		var authority system.SysAuthority
		if err := db.Where("authority_id = ?", 888).Preload("SysBaseMenus").First(&authority).Error; err != nil {
			continue
		}

		existingIDs := make(map[uint]bool)
		for _, m := range authority.SysBaseMenus {
			existingIDs[m.ID] = true
		}
		var toAppend []system.SysBaseMenu
		for _, m := range ensuredMenus {
			if !existingIDs[m.ID] {
				toAppend = append(toAppend, m)
			}
		}
		if len(toAppend) > 0 {
			if err := db.Model(&authority).Association("SysBaseMenus").Append(toAppend); err != nil {
				global.LRAG_LOG.Error("EnsureRagMenus: append menus to authority failed", zap.Error(err))
			}
		}
	}
}
