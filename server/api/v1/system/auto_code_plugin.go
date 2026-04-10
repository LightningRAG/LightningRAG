package system

import (
	"os"
	"path/filepath"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	"github.com/LightningRAG/LightningRAG/server/plugin/plugin-tool/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AutoCodePluginApi struct{}

// Install
// @Tags      AutoCodePlugin
// @Summary   安装插件
// @Security  ApiKeyAuth
// @accept    multipart/form-data
// @Produce   application/json
// @Param     plug  formData  file                                              true  "this is a test file"
// @Success   200   {object}  response.Response{data=[]interface{},msg=string}  "安装插件成功"
// @Router    /autoCode/installPlugin [post]
func (a *AutoCodePluginApi) Install(c *gin.Context) {
	header, err := c.FormFile("plug")
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	web, server, err := autoCodePluginService.Install(header)
	webStr := "web插件安装成功"
	serverStr := "server插件安装成功"
	if web == -1 {
		webStr = "web端插件未成功安装，请按照文档自行解压安装，如果为纯后端插件请忽略此条提示"
	}
	if server == -1 {
		serverStr = "server端插件未成功安装，请按照文档自行解压安装，如果为纯前端插件请忽略此条提示"
	}
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OkWithData([]interface{}{
		gin.H{
			"code": web,
			"msg":  webStr,
		},
		gin.H{
			"code": server,
			"msg":  serverStr,
		}}, c)
}

// Packaged
// @Tags      AutoCodePlugin
// @Summary   打包插件
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     plugName  query    string  true  "插件名称"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "打包插件成功"
// @Router    /autoCode/pubPlug [post]
func (a *AutoCodePluginApi) Packaged(c *gin.Context) {
	plugName := c.Query("plugName")
	zipPath, err := autoCodePluginService.PubPlug(plugName)
	if err != nil {
		global.LRAG_LOG.Error("打包失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.pack_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msgf(c, "common.pack_success_path", zipPath), c)
}

// InitMenu
// @Tags      AutoCodePlugin
// @Summary   打包插件
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "打包插件成功"
// @Router    /autoCode/initMenu [post]
func (a *AutoCodePluginApi) InitMenu(c *gin.Context) {
	var menuInfo request.InitMenu
	err := c.ShouldBindJSON(&menuInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = autoCodePluginService.InitMenu(menuInfo)
	if err != nil {
		global.LRAG_LOG.Error("创建初始化Menu失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.init_menu_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.files_updated_success"), c)
}

// InitAPI
// @Tags      AutoCodePlugin
// @Summary   打包插件
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "打包插件成功"
// @Router    /autoCode/initAPI [post]
func (a *AutoCodePluginApi) InitAPI(c *gin.Context) {
	var apiInfo request.InitApi
	err := c.ShouldBindJSON(&apiInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = autoCodePluginService.InitAPI(apiInfo)
	if err != nil {
		global.LRAG_LOG.Error("创建初始化API失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.init_api_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.files_updated_success"), c)
}

// InitDictionary
// @Tags      AutoCodePlugin
// @Summary   打包插件
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "打包插件成功"
// @Router    /autoCode/initDictionary [post]
func (a *AutoCodePluginApi) InitDictionary(c *gin.Context) {
	var dictInfo request.InitDictionary
	err := c.ShouldBindJSON(&dictInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = autoCodePluginService.InitDictionary(dictInfo)
	if err != nil {
		global.LRAG_LOG.Error("创建初始化Dictionary失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.init_dictionary_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.files_updated_success"), c)
}

// GetPluginList
// @Tags      AutoCodePlugin
// @Summary   获取插件列表
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200   {object}  response.Response{data=[]systemRes.PluginInfo}  "获取插件列表成功"
// @Router    /autoCode/getPluginList [get]
func (a *AutoCodePluginApi) GetPluginList(c *gin.Context) {
	serverDir := filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Server, "plugin")
	webDir := filepath.Join(global.LRAG_CONFIG.AutoCode.Root, global.LRAG_CONFIG.AutoCode.Web, "plugin")

	serverEntries, _ := os.ReadDir(serverDir)
	webEntries, _ := os.ReadDir(webDir)

	configMap := make(map[string]string)

	for _, entry := range serverEntries {
		if entry.IsDir() {
			configMap[entry.Name()] = "server"
		}
	}

	for _, entry := range webEntries {
		if entry.IsDir() {
			if val, ok := configMap[entry.Name()]; ok {
				if val == "server" {
					configMap[entry.Name()] = "full"
				}
			} else {
				configMap[entry.Name()] = "web"
			}
		}
	}

	var list []systemRes.PluginInfo
	for k, v := range configMap {
		apis, menus, dicts := utils.GetPluginData(k)
		list = append(list, systemRes.PluginInfo{
			PluginName:   k,
			PluginType:   v,
			Apis:         apis,
			Menus:        menus,
			Dictionaries: dicts,
		})
	}

	response.OkWithDetailed(list, i18n.Msg(c, "common.fetch_success"), c)
}

// Remove
// @Tags      AutoCodePlugin
// @Summary   删除插件
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     pluginName  query    string  true  "插件名称"
// @Param     pluginType  query    string  true  "插件类型"
// @Success   200   {object}  response.Response{msg=string}  "删除插件成功"
// @Router    /autoCode/removePlugin [post]
func (a *AutoCodePluginApi) Remove(c *gin.Context) {
	pluginName := c.Query("pluginName")
	pluginType := c.Query("pluginType")
	err := autoCodePluginService.Remove(pluginName, pluginType)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.delete_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}
