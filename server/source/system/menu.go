package system

import (
	"context"

	sysmodel "github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/LightningRAG/LightningRAG/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderMenu = initOrderAuthority + 1

type initMenu struct{}

// auto run
func init() {
	system.RegisterInit(initOrderMenu, &initMenu{})
}

func (i *initMenu) InitializerName() string {
	return sysmodel.SysBaseMenu{}.TableName()
}

func (i *initMenu) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(
		&sysmodel.SysBaseMenu{},
		&sysmodel.SysBaseMenuParameter{},
		&sysmodel.SysBaseMenuBtn{},
	)
}

func (i *initMenu) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	m := db.Migrator()
	return m.HasTable(&sysmodel.SysBaseMenu{}) &&
		m.HasTable(&sysmodel.SysBaseMenuParameter{}) &&
		m.HasTable(&sysmodel.SysBaseMenuBtn{})
}

func (i *initMenu) InitializeData(ctx context.Context) (next context.Context, err error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}

	// 定义所有菜单（Title 为英文缺省，与前端 en.js menu.names 及 DefaultMenuTitleEnglish 一致）
	allMenus := []sysmodel.SysBaseMenu{
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "dashboard", Name: "dashboard", Component: "view/dashboard/index.vue", Sort: 1, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("dashboard"), TitleKey: sysmodel.MenuTitleKeyForRouteName("dashboard"), Icon: "odometer"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "about", Name: "about", Component: "view/about/index.vue", Sort: 9, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("about"), TitleKey: sysmodel.MenuTitleKeyForRouteName("about"), Icon: "info-filled"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "admin", Name: "superAdmin", Component: "view/superAdmin/index.vue", Sort: 3, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("superAdmin"), TitleKey: sysmodel.MenuTitleKeyForRouteName("superAdmin"), Icon: "user"}},
		{MenuLevel: 0, Hidden: true, ParentId: 0, Path: "person", Name: "person", Component: "view/person/person.vue", Sort: 4, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("person"), TitleKey: sysmodel.MenuTitleKeyForRouteName("person"), Icon: "message"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "rag", Name: "rag", Component: "view/rag/index.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("rag"), TitleKey: sysmodel.MenuTitleKeyForRouteName("rag"), Icon: "reading"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "example", Name: "example", Component: "view/example/index.vue", Sort: 7, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("example"), TitleKey: sysmodel.MenuTitleKeyForRouteName("example"), Icon: "management"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "systemTools", Name: "systemTools", Component: "view/systemTools/index.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("systemTools"), TitleKey: sysmodel.MenuTitleKeyForRouteName("systemTools"), Icon: "tools"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "https://lightningrag.com", Name: "https://lightningrag.com", Component: "/", Sort: 0, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("https://lightningrag.com"), TitleKey: sysmodel.MenuTitleKeyForRouteName("https://lightningrag.com"), Icon: "customer-lrag"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "state", Name: "state", Component: "view/system/state.vue", Sort: 8, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("state"), TitleKey: sysmodel.MenuTitleKeyForRouteName("state"), Icon: "cloudy"}},
		{MenuLevel: 0, Hidden: false, ParentId: 0, Path: "plugin", Name: "plugin", Component: "view/routerHolder.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("plugin"), TitleKey: sysmodel.MenuTitleKeyForRouteName("plugin"), Icon: "cherry"}},
	}

	// 先创建父级菜单（ParentId = 0 的菜单）
	if err = db.Create(&allMenus).Error; err != nil {
		return ctx, errors.Wrap(err, sysmodel.SysBaseMenu{}.TableName()+"父级菜单初始化失败!")
	}

	// 建立菜单映射 - 通过Name查找已创建的菜单及其ID
	menuNameMap := make(map[string]uint)
	for _, menu := range allMenus {
		menuNameMap[menu.Name] = menu.ID
	}

	// 定义子菜单，并设置正确的ParentId
	childMenus := []sysmodel.SysBaseMenu{
		// superAdmin子菜单
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "authority", Name: "authority", Component: "view/superAdmin/authority/authority.vue", Sort: 1, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("authority"), TitleKey: sysmodel.MenuTitleKeyForRouteName("authority"), Icon: "avatar"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "menu", Name: "menu", Component: "view/superAdmin/menu/menu.vue", Sort: 2, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("menu"), TitleKey: sysmodel.MenuTitleKeyForRouteName("menu"), Icon: "tickets", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "api", Name: "api", Component: "view/superAdmin/api/api.vue", Sort: 3, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("api"), TitleKey: sysmodel.MenuTitleKeyForRouteName("api"), Icon: "platform", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "user", Name: "user", Component: "view/superAdmin/user/user.vue", Sort: 4, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("user"), TitleKey: sysmodel.MenuTitleKeyForRouteName("user"), Icon: "coordinate"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "dictionary", Name: "dictionary", Component: "view/superAdmin/dictionary/sysDictionary.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("dictionary"), TitleKey: sysmodel.MenuTitleKeyForRouteName("dictionary"), Icon: "notebook"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "operation", Name: "operation", Component: "view/superAdmin/operation/sysOperationRecord.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("operation"), TitleKey: sysmodel.MenuTitleKeyForRouteName("operation"), Icon: "pie-chart"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["superAdmin"], Path: "sysParams", Name: "sysParams", Component: "view/superAdmin/params/sysParams.vue", Sort: 7, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("sysParams"), TitleKey: sysmodel.MenuTitleKeyForRouteName("sysParams"), Icon: "compass"}},

		// rag子菜单
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["rag"], Path: "knowledgeBase", Name: "knowledgeBase", Component: "view/rag/knowledgeBase/knowledgeBase.vue", Sort: 1, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("knowledgeBase"), TitleKey: sysmodel.MenuTitleKeyForRouteName("knowledgeBase"), Icon: "folder"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["rag"], Path: "retrieval", Name: "ragDocumentRetrieval", Component: "view/rag/knowledgeBase/retrieval.vue", Sort: 2, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragDocumentRetrieval"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragDocumentRetrieval"), Icon: "search"}},
		{MenuLevel: 1, Hidden: true, ParentId: menuNameMap["rag"], Path: "ragDocuments", Name: "ragDocuments", Component: "view/rag/knowledgeBase/documents.vue", Sort: 3, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragDocuments"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragDocuments"), Icon: "document"}},
		{MenuLevel: 1, Hidden: true, ParentId: menuNameMap["rag"], Path: "ragKnowledgeGraph", Name: "ragKnowledgeGraph", Component: "view/rag/knowledgeBase/knowledgeGraph.vue", Sort: 3, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragKnowledgeGraph"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragKnowledgeGraph"), Icon: "histogram"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["rag"], Path: "conversation", Name: "conversation", Component: "view/rag/conversation/conversation.vue", Sort: 4, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("conversation"), TitleKey: sysmodel.MenuTitleKeyForRouteName("conversation"), Icon: "chat-dot-round"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["rag"], Path: "agent", Name: "ragAgent", Component: "view/rag/agent/index.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragAgent"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragAgent"), Icon: "connection"}},
		{MenuLevel: 1, Hidden: true, ParentId: menuNameMap["rag"], Path: "agentEditor", Name: "ragAgentEditor", Component: "view/rag/agent/editor.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragAgentEditor"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragAgentEditor"), Icon: "edit"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["rag"], Path: "model", Name: "ragModel", Component: "view/rag/model/model.vue", Sort: 7, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragModel"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragModel"), Icon: "cpu"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["rag"], Path: "channelConnector", Name: "ragChannelConnectors", Component: "view/rag/channelConnector/channelConnector.vue", Sort: 8, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("ragChannelConnectors"), TitleKey: sysmodel.MenuTitleKeyForRouteName("ragChannelConnectors"), Icon: "message"}},

		// example子菜单
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["example"], Path: "upload", Name: "upload", Component: "view/example/upload/upload.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("upload"), TitleKey: sysmodel.MenuTitleKeyForRouteName("upload"), Icon: "upload"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["example"], Path: "breakpoint", Name: "breakpoint", Component: "view/example/breakpoint/breakpoint.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("breakpoint"), TitleKey: sysmodel.MenuTitleKeyForRouteName("breakpoint"), Icon: "upload-filled"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["example"], Path: "customer", Name: "customer", Component: "view/example/customer/customer.vue", Sort: 7, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("customer"), TitleKey: sysmodel.MenuTitleKeyForRouteName("customer"), Icon: "avatar"}},

		// systemTools子菜单
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "autoCode", Name: "autoCode", Component: "view/systemTools/autoCode/index.vue", Sort: 1, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("autoCode"), TitleKey: sysmodel.MenuTitleKeyForRouteName("autoCode"), Icon: "cpu", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "formCreate", Name: "formCreate", Component: "view/systemTools/formCreate/index.vue", Sort: 3, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("formCreate"), TitleKey: sysmodel.MenuTitleKeyForRouteName("formCreate"), Icon: "magic-stick", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "system", Name: "system", Component: "view/systemTools/system/system.vue", Sort: 4, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("system"), TitleKey: sysmodel.MenuTitleKeyForRouteName("system"), Icon: "operation"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "autoCodeAdmin", Name: "autoCodeAdmin", Component: "view/systemTools/autoCodeAdmin/index.vue", Sort: 2, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("autoCodeAdmin"), TitleKey: sysmodel.MenuTitleKeyForRouteName("autoCodeAdmin"), Icon: "magic-stick"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "loginLog", Name: "loginLog", Component: "view/systemTools/loginLog/index.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("loginLog"), TitleKey: sysmodel.MenuTitleKeyForRouteName("loginLog"), Icon: "monitor"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "apiToken", Name: "apiToken", Component: "view/systemTools/apiToken/index.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("apiToken"), TitleKey: sysmodel.MenuTitleKeyForRouteName("apiToken"), Icon: "key"}},
		{MenuLevel: 1, Hidden: true, ParentId: menuNameMap["systemTools"], Path: "autoCodeEdit/:id", Name: "autoCodeEdit", Component: "view/systemTools/autoCode/index.vue", Sort: 0, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("autoCodeEdit"), TitleKey: sysmodel.MenuTitleKeyForRouteName("autoCodeEdit"), Icon: "magic-stick"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "autoPkg", Name: "autoPkg", Component: "view/systemTools/autoPkg/autoPkg.vue", Sort: 0, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("autoPkg"), TitleKey: sysmodel.MenuTitleKeyForRouteName("autoPkg"), Icon: "folder"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "exportTemplate", Name: "exportTemplate", Component: "view/systemTools/exportTemplate/exportTemplate.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("exportTemplate"), TitleKey: sysmodel.MenuTitleKeyForRouteName("exportTemplate"), Icon: "reading"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "skills", Name: "skills", Component: "view/systemTools/skills/index.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("skills"), TitleKey: sysmodel.MenuTitleKeyForRouteName("skills"), Icon: "document"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "picture", Name: "picture", Component: "view/systemTools/autoCode/picture.vue", Sort: 6, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("picture"), TitleKey: sysmodel.MenuTitleKeyForRouteName("picture"), Icon: "picture-filled"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "mcpTool", Name: "mcpTool", Component: "view/systemTools/autoCode/mcp.vue", Sort: 7, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("mcpTool"), TitleKey: sysmodel.MenuTitleKeyForRouteName("mcpTool"), Icon: "magnet"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "mcpTest", Name: "mcpTest", Component: "view/systemTools/autoCode/mcpTest.vue", Sort: 7, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("mcpTest"), TitleKey: sysmodel.MenuTitleKeyForRouteName("mcpTest"), Icon: "partly-cloudy"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "sysVersion", Name: "sysVersion", Component: "view/systemTools/version/version.vue", Sort: 8, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("sysVersion"), TitleKey: sysmodel.MenuTitleKeyForRouteName("sysVersion"), Icon: "server"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "sysError", Name: "sysError", Component: "view/systemTools/sysError/sysError.vue", Sort: 9, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("sysError"), TitleKey: sysmodel.MenuTitleKeyForRouteName("sysError"), Icon: "warn"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["systemTools"], Path: "oauthProvider", Name: "oauthProvider", Component: "view/oauth/settings.vue", Sort: 10, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("oauthProvider"), TitleKey: sysmodel.MenuTitleKeyForRouteName("oauthProvider"), Icon: "key"}},

		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["plugin"], Path: "installPlugin", Name: "installPlugin", Component: "view/systemTools/installPlugin/index.vue", Sort: 1, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("installPlugin"), TitleKey: sysmodel.MenuTitleKeyForRouteName("installPlugin"), Icon: "box"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["plugin"], Path: "pubPlug", Name: "pubPlug", Component: "view/systemTools/pubPlug/pubPlug.vue", Sort: 3, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("pubPlug"), TitleKey: sysmodel.MenuTitleKeyForRouteName("pubPlug"), Icon: "files"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["plugin"], Path: "plugin-email", Name: "plugin-email", Component: "plugin/email/view/index.vue", Sort: 4, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("plugin-email"), TitleKey: sysmodel.MenuTitleKeyForRouteName("plugin-email"), Icon: "message"}},
		{MenuLevel: 1, Hidden: false, ParentId: menuNameMap["plugin"], Path: "anInfo", Name: "anInfo", Component: "plugin/announcement/view/info.vue", Sort: 5, Meta: sysmodel.Meta{Title: sysmodel.DefaultMenuTitleEnglish("anInfo"), TitleKey: sysmodel.MenuTitleKeyForRouteName("anInfo"), Icon: "scaleToOriginal"}},
	}

	// 创建子菜单
	if err = db.Create(&childMenus).Error; err != nil {
		return ctx, errors.Wrap(err, sysmodel.SysBaseMenu{}.TableName()+"子菜单初始化失败!")
	}

	// 组合所有菜单作为返回结果
	allEntities := append(allMenus, childMenus...)
	next = context.WithValue(ctx, i.InitializerName(), allEntities)
	return next, nil
}

func (i *initMenu) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where("path = ?", "autoPkg").First(&sysmodel.SysBaseMenu{}).Error, gorm.ErrRecordNotFound) { // 判断是否存在数据
		return false
	}
	return true
}
