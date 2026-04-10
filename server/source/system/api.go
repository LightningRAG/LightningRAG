package system

import (
	"context"

	sysModel "github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/LightningRAG/LightningRAG/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type initApi struct{}

const initOrderApi = system.InitOrderSystem + 1

// auto run
func init() {
	system.RegisterInit(initOrderApi, &initApi{})
}

func (i *initApi) InitializerName() string {
	return sysModel.SysApi{}.TableName()
}

func (i *initApi) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysApi{})
}

func (i *initApi) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&sysModel.SysApi{})
}

func (i *initApi) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	entities := []sysModel.SysApi{
		{ApiGroup: "jwt", Method: "POST", Path: "/jwt/jsonInBlacklist", Description: "Add JWT to blacklist (logout, required)"},

		{ApiGroup: "Login log", Method: "DELETE", Path: "/sysLoginLog/deleteLoginLog", Description: "Delete login log"},
		{ApiGroup: "Login log", Method: "DELETE", Path: "/sysLoginLog/deleteLoginLogByIds", Description: "Batch delete login logs"},
		{ApiGroup: "Login log", Method: "GET", Path: "/sysLoginLog/findLoginLog", Description: "Get login log by ID"},
		{ApiGroup: "Login log", Method: "GET", Path: "/sysLoginLog/getLoginLogList", Description: "Get login log list"},

		{ApiGroup: "API Token", Method: "POST", Path: "/sysApiToken/createApiToken", Description: "Issue API token"},
		{ApiGroup: "API Token", Method: "POST", Path: "/sysApiToken/getApiTokenList", Description: "Get API token list"},
		{ApiGroup: "API Token", Method: "POST", Path: "/sysApiToken/deleteApiToken", Description: "Revoke API token"},

		{ApiGroup: "System user", Method: "DELETE", Path: "/user/deleteUser", Description: "Delete user"},
		{ApiGroup: "System user", Method: "POST", Path: "/user/admin_register", Description: "Register user"},
		{ApiGroup: "System user", Method: "POST", Path: "/user/getUserList", Description: "Get user list"},
		{ApiGroup: "System user", Method: "PUT", Path: "/user/setUserInfo", Description: "Set user info"},
		{ApiGroup: "System user", Method: "PUT", Path: "/user/setSelfInfo", Description: "Set self info (required)"},
		{ApiGroup: "System user", Method: "GET", Path: "/user/getUserInfo", Description: "Get self info (required)"},
		{ApiGroup: "System user", Method: "POST", Path: "/user/setUserAuthorities", Description: "Set user authorities"},
		{ApiGroup: "System user", Method: "POST", Path: "/user/changePassword", Description: "Change password (recommended)"},
		{ApiGroup: "System user", Method: "POST", Path: "/user/setUserAuthority", Description: "Set user authority (required)"},
		{ApiGroup: "System user", Method: "POST", Path: "/user/resetPassword", Description: "Reset user password"},
		{ApiGroup: "System user", Method: "PUT", Path: "/user/setSelfSetting", Description: "User UI settings"},

		{ApiGroup: "api", Method: "POST", Path: "/api/createApi", Description: "Create API"},
		{ApiGroup: "api", Method: "POST", Path: "/api/deleteApi", Description: "Delete API"},
		{ApiGroup: "api", Method: "POST", Path: "/api/updateApi", Description: "Update API"},
		{ApiGroup: "api", Method: "POST", Path: "/api/getApiList", Description: "Get API list"},
		{ApiGroup: "api", Method: "POST", Path: "/api/getAllApis", Description: "Get all APIs"},
		{ApiGroup: "api", Method: "POST", Path: "/api/getApiById", Description: "Get API details"},
		{ApiGroup: "api", Method: "DELETE", Path: "/api/deleteApisByIds", Description: "Batch delete APIs"},
		{ApiGroup: "api", Method: "GET", Path: "/api/syncApi", Description: "Get APIs pending sync"},
		{ApiGroup: "api", Method: "GET", Path: "/api/getApiGroups", Description: "Get API route groups"},
		{ApiGroup: "api", Method: "POST", Path: "/api/enterSyncApi", Description: "Confirm API sync"},
		{ApiGroup: "api", Method: "POST", Path: "/api/ignoreApi", Description: "Ignore API"},

		{ApiGroup: "Role", Method: "POST", Path: "/authority/copyAuthority", Description: "Copy role"},
		{ApiGroup: "Role", Method: "POST", Path: "/authority/createAuthority", Description: "Create role"},
		{ApiGroup: "Role", Method: "POST", Path: "/authority/deleteAuthority", Description: "Delete role"},
		{ApiGroup: "Role", Method: "PUT", Path: "/authority/updateAuthority", Description: "Update role"},
		{ApiGroup: "Role", Method: "POST", Path: "/authority/getAuthorityList", Description: "Get role list"},
		{ApiGroup: "Role", Method: "POST", Path: "/authority/setDataAuthority", Description: "Set data authority"},
		{ApiGroup: "Role", Method: "GET", Path: "/authority/getUsersByAuthority", Description: "Get user IDs by authority"},
		{ApiGroup: "Role", Method: "POST", Path: "/authority/setRoleUsers", Description: "Replace role users"},

		{ApiGroup: "casbin", Method: "POST", Path: "/casbin/updateCasbin", Description: "Update role API policy"},
		{ApiGroup: "casbin", Method: "POST", Path: "/casbin/getPolicyPathByAuthorityId", Description: "Get policy list"},

		{ApiGroup: "Menu", Method: "POST", Path: "/menu/addBaseMenu", Description: "Add menu"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/getMenu", Description: "Get menu tree (required)"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/deleteBaseMenu", Description: "Delete menu"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/updateBaseMenu", Description: "Update menu"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/getBaseMenuById", Description: "Get menu by ID"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/getMenuList", Description: "Paginated base menu list"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/getBaseMenuTree", Description: "Get user dynamic routes"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/getMenuAuthority", Description: "Get menus by authority"},
		{ApiGroup: "Menu", Method: "POST", Path: "/menu/addMenuAuthority", Description: "Add menu authority"},

		{ApiGroup: "Chunked upload", Method: "GET", Path: "/fileUploadAndDownload/findFile", Description: "Find file (instant upload)"},
		{ApiGroup: "Chunked upload", Method: "POST", Path: "/fileUploadAndDownload/breakpointContinue", Description: "Resume upload"},
		{ApiGroup: "Chunked upload", Method: "POST", Path: "/fileUploadAndDownload/breakpointContinueFinish", Description: "Finish resume upload"},
		{ApiGroup: "Chunked upload", Method: "POST", Path: "/fileUploadAndDownload/removeChunk", Description: "Remove chunk after upload"},

		{ApiGroup: "File upload & download", Method: "POST", Path: "/fileUploadAndDownload/upload", Description: "Upload file (recommended)"},
		{ApiGroup: "File upload & download", Method: "POST", Path: "/fileUploadAndDownload/deleteFile", Description: "Delete file"},
		{ApiGroup: "File upload & download", Method: "POST", Path: "/fileUploadAndDownload/editFileName", Description: "Edit file name or remark"},
		{ApiGroup: "File upload & download", Method: "POST", Path: "/fileUploadAndDownload/getFileList", Description: "Get uploaded file list"},
		{ApiGroup: "File upload & download", Method: "POST", Path: "/fileUploadAndDownload/importURL", Description: "Import URL"},

		{ApiGroup: "System service", Method: "POST", Path: "/system/getServerInfo", Description: "Get server info"},
		{ApiGroup: "System service", Method: "POST", Path: "/system/getSystemConfig", Description: "Get system config"},
		{ApiGroup: "System service", Method: "POST", Path: "/system/setSystemConfig", Description: "Set system config"},

		{ApiGroup: "skills", Method: "GET", Path: "/skills/getTools", Description: "List skill tools"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getSkillList", Description: "List skills"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getSkillDetail", Description: "Get skill detail"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/saveSkill", Description: "Save skill definition"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/deleteSkill", Description: "Delete skill"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/createScript", Description: "Create skill script"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getScript", Description: "Get skill script"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/saveScript", Description: "Save skill script"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/createResource", Description: "Create skill resource"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getResource", Description: "Get skill resource"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/saveResource", Description: "Save skill resource"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/createReference", Description: "Create skill reference"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getReference", Description: "Get skill reference"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/saveReference", Description: "Save skill reference"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/createTemplate", Description: "Create skill template"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getTemplate", Description: "Get skill template"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/saveTemplate", Description: "Save skill template"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/getGlobalConstraint", Description: "Get global constraint"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/saveGlobalConstraint", Description: "Save global constraint"},
		{ApiGroup: "skills", Method: "POST", Path: "/skills/packageSkill", Description: "Package skill"},

		// RAG 相关 API 统一由 initialize/rag_api.go 的 EnsureRagApis() 注册，避免双源重复

		{ApiGroup: "Customer", Method: "PUT", Path: "/customer/customer", Description: "Update customer"},
		{ApiGroup: "Customer", Method: "POST", Path: "/customer/customer", Description: "Create customer"},
		{ApiGroup: "Customer", Method: "DELETE", Path: "/customer/customer", Description: "Delete customer"},
		{ApiGroup: "Customer", Method: "GET", Path: "/customer/customer", Description: "Get customer"},
		{ApiGroup: "Customer", Method: "GET", Path: "/customer/customerList", Description: "Get customer list"},

		{ApiGroup: "Code generator", Method: "GET", Path: "/autoCode/getDB", Description: "Get all databases"},
		{ApiGroup: "Code generator", Method: "GET", Path: "/autoCode/getTables", Description: "Get database tables"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/createTemp", Description: "Generate code"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/preview", Description: "Preview generated code"},
		{ApiGroup: "Code generator", Method: "GET", Path: "/autoCode/getColumn", Description: "Get table columns"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/installPlugin", Description: "Install plugin"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/pubPlug", Description: "Package plugin"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/removePlugin", Description: "Uninstall plugin"},
		{ApiGroup: "Code generator", Method: "GET", Path: "/autoCode/getPluginList", Description: "Get installed plugins"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/mcp", Description: "Generate MCP tool template"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/mcpTest", Description: "MCP tool test"},
		{ApiGroup: "Code generator", Method: "POST", Path: "/autoCode/mcpList", Description: "Get MCP tool list"},

		{ApiGroup: "Template config", Method: "POST", Path: "/autoCode/createPackage", Description: "Configure template"},
		{ApiGroup: "Template config", Method: "GET", Path: "/autoCode/getTemplates", Description: "Get template files"},
		{ApiGroup: "Template config", Method: "POST", Path: "/autoCode/getPackage", Description: "Get all templates"},
		{ApiGroup: "Template config", Method: "POST", Path: "/autoCode/delPackage", Description: "Delete template"},

		{ApiGroup: "Code generator history", Method: "POST", Path: "/autoCode/getMeta", Description: "Get meta info"},
		{ApiGroup: "Code generator history", Method: "POST", Path: "/autoCode/rollback", Description: "Rollback generated code"},
		{ApiGroup: "Code generator history", Method: "POST", Path: "/autoCode/getSysHistory", Description: "Query rollback history"},
		{ApiGroup: "Code generator history", Method: "POST", Path: "/autoCode/delSysHistory", Description: "Delete rollback history"},
		{ApiGroup: "Code generator history", Method: "POST", Path: "/autoCode/addFunc", Description: "Add template method"},

		{ApiGroup: "Dictionary detail", Method: "PUT", Path: "/sysDictionaryDetail/updateSysDictionaryDetail", Description: "Update dictionary detail"},
		{ApiGroup: "Dictionary detail", Method: "POST", Path: "/sysDictionaryDetail/createSysDictionaryDetail", Description: "Create dictionary detail"},
		{ApiGroup: "Dictionary detail", Method: "DELETE", Path: "/sysDictionaryDetail/deleteSysDictionaryDetail", Description: "Delete dictionary detail"},
		{ApiGroup: "Dictionary detail", Method: "GET", Path: "/sysDictionaryDetail/findSysDictionaryDetail", Description: "Get dictionary detail by ID"},
		{ApiGroup: "Dictionary detail", Method: "GET", Path: "/sysDictionaryDetail/getSysDictionaryDetailList", Description: "Get dictionary detail list"},

		{ApiGroup: "Dictionary detail", Method: "GET", Path: "/sysDictionaryDetail/getDictionaryTreeList", Description: "Get dictionary tree list"},
		{ApiGroup: "Dictionary detail", Method: "GET", Path: "/sysDictionaryDetail/getDictionaryTreeListByType", Description: "Get dictionary tree by type"},
		{ApiGroup: "Dictionary detail", Method: "GET", Path: "/sysDictionaryDetail/getDictionaryDetailsByParent", Description: "Get dictionary details by parent ID"},
		{ApiGroup: "Dictionary detail", Method: "GET", Path: "/sysDictionaryDetail/getDictionaryPath", Description: "Get dictionary detail full path"},

		{ApiGroup: "Dictionary", Method: "POST", Path: "/sysDictionary/createSysDictionary", Description: "Create dictionary"},
		{ApiGroup: "Dictionary", Method: "DELETE", Path: "/sysDictionary/deleteSysDictionary", Description: "Delete dictionary"},
		{ApiGroup: "Dictionary", Method: "PUT", Path: "/sysDictionary/updateSysDictionary", Description: "Update dictionary"},
		{ApiGroup: "Dictionary", Method: "GET", Path: "/sysDictionary/findSysDictionary", Description: "Get dictionary by ID (recommended)"},
		{ApiGroup: "Dictionary", Method: "GET", Path: "/sysDictionary/getSysDictionaryList", Description: "Get dictionary list"},
		{ApiGroup: "Dictionary", Method: "POST", Path: "/sysDictionary/importSysDictionary", Description: "Import dictionary JSON"},
		{ApiGroup: "Dictionary", Method: "GET", Path: "/sysDictionary/exportSysDictionary", Description: "Export dictionary JSON"},

		{ApiGroup: "Operation record", Method: "POST", Path: "/sysOperationRecord/createSysOperationRecord", Description: "Create operation record"},
		{ApiGroup: "Operation record", Method: "GET", Path: "/sysOperationRecord/findSysOperationRecord", Description: "Get operation record by ID"},
		{ApiGroup: "Operation record", Method: "GET", Path: "/sysOperationRecord/getSysOperationRecordList", Description: "Get operation record list"},
		{ApiGroup: "Operation record", Method: "DELETE", Path: "/sysOperationRecord/deleteSysOperationRecord", Description: "Delete operation record"},
		{ApiGroup: "Operation record", Method: "DELETE", Path: "/sysOperationRecord/deleteSysOperationRecordByIds", Description: "Batch delete operation records"},

		{ApiGroup: "Simple uploader", Method: "POST", Path: "/simpleUploader/upload", Description: "Plugin chunked upload"},
		{ApiGroup: "Simple uploader", Method: "GET", Path: "/simpleUploader/checkFileMd5", Description: "Verify file integrity"},
		{ApiGroup: "Simple uploader", Method: "GET", Path: "/simpleUploader/mergeFileMd5", Description: "Merge uploaded chunks"},

		{ApiGroup: "email", Method: "POST", Path: "/email/emailTest", Description: "Send test email"},
		{ApiGroup: "email", Method: "POST", Path: "/email/sendEmail", Description: "Send email"},

		{ApiGroup: "Authority button", Method: "POST", Path: "/authorityBtn/setAuthorityBtn", Description: "Set button authority"},
		{ApiGroup: "Authority button", Method: "POST", Path: "/authorityBtn/getAuthorityBtn", Description: "Get button authority"},
		{ApiGroup: "Authority button", Method: "POST", Path: "/authorityBtn/canRemoveAuthorityBtn", Description: "Remove button authority"},

		{ApiGroup: "Export template", Method: "POST", Path: "/sysExportTemplate/createSysExportTemplate", Description: "Create export template"},
		{ApiGroup: "Export template", Method: "DELETE", Path: "/sysExportTemplate/deleteSysExportTemplate", Description: "Delete export template"},
		{ApiGroup: "Export template", Method: "DELETE", Path: "/sysExportTemplate/deleteSysExportTemplateByIds", Description: "Batch delete export templates"},
		{ApiGroup: "Export template", Method: "PUT", Path: "/sysExportTemplate/updateSysExportTemplate", Description: "Update export template"},
		{ApiGroup: "Export template", Method: "GET", Path: "/sysExportTemplate/findSysExportTemplate", Description: "Get export template by ID"},
		{ApiGroup: "Export template", Method: "GET", Path: "/sysExportTemplate/getSysExportTemplateList", Description: "Get export template list"},
		{ApiGroup: "Export template", Method: "GET", Path: "/sysExportTemplate/exportExcel", Description: "Export Excel"},
		{ApiGroup: "Export template", Method: "GET", Path: "/sysExportTemplate/exportTemplate", Description: "Download template"},
		{ApiGroup: "Export template", Method: "GET", Path: "/sysExportTemplate/previewSQL", Description: "Preview SQL"},
		{ApiGroup: "Export template", Method: "POST", Path: "/sysExportTemplate/importExcel", Description: "Import Excel"},

		{ApiGroup: "Error log", Method: "POST", Path: "/sysError/createSysError", Description: "Create error log"},
		{ApiGroup: "Error log", Method: "DELETE", Path: "/sysError/deleteSysError", Description: "Delete error log"},
		{ApiGroup: "Error log", Method: "DELETE", Path: "/sysError/deleteSysErrorByIds", Description: "Batch delete error logs"},
		{ApiGroup: "Error log", Method: "PUT", Path: "/sysError/updateSysError", Description: "Update error log"},
		{ApiGroup: "Error log", Method: "GET", Path: "/sysError/findSysError", Description: "Get error log by ID"},
		{ApiGroup: "Error log", Method: "GET", Path: "/sysError/getSysErrorList", Description: "Get error log list"},
		{ApiGroup: "Error log", Method: "GET", Path: "/sysError/getSysErrorSolution", Description: "Trigger error handling (async)"},

		{ApiGroup: "Announcement", Method: "POST", Path: "/info/createInfo", Description: "Create announcement"},
		{ApiGroup: "Announcement", Method: "DELETE", Path: "/info/deleteInfo", Description: "Delete announcement"},
		{ApiGroup: "Announcement", Method: "DELETE", Path: "/info/deleteInfoByIds", Description: "Batch delete announcements"},
		{ApiGroup: "Announcement", Method: "PUT", Path: "/info/updateInfo", Description: "Update announcement"},
		{ApiGroup: "Announcement", Method: "GET", Path: "/info/findInfo", Description: "Get announcement by ID"},
		{ApiGroup: "Announcement", Method: "GET", Path: "/info/getInfoList", Description: "Get announcement list"},

		{ApiGroup: "System params", Method: "POST", Path: "/sysParams/createSysParams", Description: "Create parameter"},
		{ApiGroup: "System params", Method: "DELETE", Path: "/sysParams/deleteSysParams", Description: "Delete parameter"},
		{ApiGroup: "System params", Method: "DELETE", Path: "/sysParams/deleteSysParamsByIds", Description: "Batch delete parameters"},
		{ApiGroup: "System params", Method: "PUT", Path: "/sysParams/updateSysParams", Description: "Update parameter"},
		{ApiGroup: "System params", Method: "GET", Path: "/sysParams/findSysParams", Description: "Get parameter by ID"},
		{ApiGroup: "System params", Method: "GET", Path: "/sysParams/getSysParamsList", Description: "Get parameter list"},
		{ApiGroup: "System params", Method: "GET", Path: "/sysParams/getSysParam", Description: "Get parameter"},

		{ApiGroup: "OAuth provider", Method: "POST", Path: "/sysOAuthProvider/createOAuthProvider", Description: "Create OAuth provider"},
		{ApiGroup: "OAuth provider", Method: "DELETE", Path: "/sysOAuthProvider/deleteOAuthProvider", Description: "Delete OAuth provider"},
		{ApiGroup: "OAuth provider", Method: "DELETE", Path: "/sysOAuthProvider/deleteOAuthProviderByIds", Description: "Batch delete OAuth providers"},
		{ApiGroup: "OAuth provider", Method: "PUT", Path: "/sysOAuthProvider/updateOAuthProvider", Description: "Update OAuth provider"},
		{ApiGroup: "OAuth provider", Method: "GET", Path: "/sysOAuthProvider/findOAuthProvider", Description: "Get OAuth provider by ID"},
		{ApiGroup: "OAuth provider", Method: "GET", Path: "/sysOAuthProvider/getOAuthProviderList", Description: "OAuth provider list"},
		{ApiGroup: "OAuth provider", Method: "GET", Path: "/sysOAuthProvider/getRegisteredOAuthKinds", Description: "Registered OAuth kinds"},

		{ApiGroup: "OAuth setting", Method: "GET", Path: "/sysOAuthSetting/getOAuthSetting", Description: "Get OAuth global settings"},
		{ApiGroup: "OAuth setting", Method: "PUT", Path: "/sysOAuthSetting/updateOAuthSetting", Description: "Update OAuth global settings"},

		{ApiGroup: "Attachment category", Method: "GET", Path: "/attachmentCategory/getCategoryList", Description: "Category list"},
		{ApiGroup: "Attachment category", Method: "POST", Path: "/attachmentCategory/addCategory", Description: "Add or edit category"},
		{ApiGroup: "Attachment category", Method: "POST", Path: "/attachmentCategory/deleteCategory", Description: "Delete category"},

		{ApiGroup: "Version control", Method: "GET", Path: "/sysVersion/findSysVersion", Description: "Get version"},
		{ApiGroup: "Version control", Method: "GET", Path: "/sysVersion/getSysVersionList", Description: "Get version list"},
		{ApiGroup: "Version control", Method: "GET", Path: "/sysVersion/downloadVersionJson", Description: "Download version JSON"},
		{ApiGroup: "Version control", Method: "POST", Path: "/sysVersion/exportVersion", Description: "Create version export"},
		{ApiGroup: "Version control", Method: "POST", Path: "/sysVersion/importVersion", Description: "Import version"},
		{ApiGroup: "Version control", Method: "DELETE", Path: "/sysVersion/deleteSysVersion", Description: "Delete version"},
		{ApiGroup: "Version control", Method: "DELETE", Path: "/sysVersion/deleteSysVersionByIds", Description: "Batch delete versions"},
	}
	if err := db.Create(&entities).Error; err != nil {
		return ctx, errors.Wrap(err, sysModel.SysApi{}.TableName()+"表数据初始化失败!")
	}
	next := context.WithValue(ctx, i.InitializerName(), entities)
	return next, nil
}

func (i *initApi) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where("path = ? AND method = ?", "/authorityBtn/canRemoveAuthorityBtn", "POST").
		First(&sysModel.SysApi{}).Error, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
