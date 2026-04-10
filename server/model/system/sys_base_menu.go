package system

import (
	"net/url"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
)

type SysBaseMenu struct {
	global.LRAG_MODEL
	MenuLevel     uint                          `json:"-"`
	ParentId      uint                          `json:"parentId" gorm:"comment:父菜单ID"`     // 父菜单ID
	Path          string                        `json:"path" gorm:"comment:路由path"`        // 路由path
	Name          string                        `json:"name" gorm:"comment:路由name"`        // 路由name
	Hidden        bool                          `json:"hidden" gorm:"comment:是否在列表隐藏"`     // 是否在列表隐藏
	Component     string                        `json:"component" gorm:"comment:对应前端文件路径"` // 对应前端文件路径
	Sort          int                           `json:"sort" gorm:"comment:排序标记"`          // 排序标记
	Meta          `json:"meta" gorm:"embedded"` // 附加属性
	SysAuthoritys []SysAuthority                `json:"authoritys" gorm:"many2many:sys_authority_menus;"`
	Children      []SysBaseMenu                 `json:"children" gorm:"-"`
	Parameters    []SysBaseMenuParameter        `json:"parameters"`
	MenuBtn       []SysBaseMenuBtn              `json:"menuBtn"`
}

type Meta struct {
	ActiveName     string `json:"activeName" gorm:"comment:高亮菜单"`
	KeepAlive      bool   `json:"keepAlive" gorm:"comment:是否缓存"`                        // 是否缓存
	DefaultMenu    bool   `json:"defaultMenu" gorm:"comment:是否是基础路由（开发中）"`              // 是否是基础路由（开发中）
	Title          string `json:"title" gorm:"comment:菜单名(英文缺省,与 en.js menu.names 一致)"` // 菜单名；有 titleKey 时前端优先 i18n
	TitleKey       string `json:"titleKey" gorm:"comment:前端i18n键(menu.names.*)"`        // 非空时前端用 vue-i18n 展示
	Icon           string `json:"icon" gorm:"comment:菜单图标"`                             // 菜单图标
	CloseTab       bool   `json:"closeTab" gorm:"comment:自动关闭tab"`                      // 自动关闭tab
	TransitionType string `json:"transitionType" gorm:"comment:路由切换动画"`                 // 路由切换动画
}

type SysBaseMenuParameter struct {
	global.LRAG_MODEL
	SysBaseMenuID uint
	Type          string `json:"type" gorm:"comment:地址栏携带参数为params还是query"` // 地址栏携带参数为params还是query
	Key           string `json:"key" gorm:"comment:地址栏携带参数的key"`            // 地址栏携带参数的key
	Value         string `json:"value" gorm:"comment:地址栏携带参数的值"`            // 地址栏携带参数的值
}

func (SysBaseMenu) TableName() string {
	return "sys_base_menus"
}

// MenuTitleKeyForRouteName 与 web 端 locale 中 menu.names.* 键一致（按路由 Name 生成 titleKey）。
// 外链菜单的 Name 常为完整 URL（含 www、大小写、尾斜杠），不能拼成 i18n 键，需按主机名归一化。
func MenuTitleKeyForRouteName(routeName string) string {
	switch routeName {
	case "plugin-email":
		return "menu.names.pluginEmail"
	default:
		if key := menuTitleKeyForExternalURL(routeName); key != "" {
			return key
		}
		return "menu.names." + routeName
	}
}

func menuTitleKeyForExternalURL(routeName string) string {
	s := strings.TrimSpace(routeName)
	if !strings.Contains(s, "://") {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	switch host {
	case "lightningrag.com", "www.lightningrag.com":
		return "menu.names.officialWebsite"
	case "plugin.lightningrag.com":
		return "menu.names.pluginMarket"
	default:
		return ""
	}
}
