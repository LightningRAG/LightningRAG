package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	"github.com/LightningRAG/LightningRAG/server/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SystemApiApi struct{}

// CreateApi
// @Tags      SysApi
// @Summary   创建基础api
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysApi                  true  "api路径, api中文描述, api组, 方法"
// @Success   200   {object}  response.Response{msg=string}  "创建基础api"
// @Router    /api/createApi [post]
func (s *SystemApiApi) CreateApi(c *gin.Context) {
	var api system.SysApi
	err := c.ShouldBindJSON(&api)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(api, utils.ApiVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiService.CreateApi(api)
	if err != nil {
		global.LRAG_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.create_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.create_success"), c)
}

// SyncApi
// @Tags      SysApi
// @Summary   同步API
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "同步API"
// @Router    /api/syncApi [get]
func (s *SystemApiApi) SyncApi(c *gin.Context) {
	newApis, deleteApis, ignoreApis, err := apiService.SyncApi()
	if err != nil {
		global.LRAG_LOG.Error("同步失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.sync_failed"), c)
		return
	}
	response.OkWithData(gin.H{
		"newApis":    newApis,
		"deleteApis": deleteApis,
		"ignoreApis": ignoreApis,
	}, c)
}

// GetApiGroups
// @Tags      SysApi
// @Summary   获取API分组
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "获取API分组"
// @Router    /api/getApiGroups [get]
func (s *SystemApiApi) GetApiGroups(c *gin.Context) {
	groups, apiGroupMap, err := apiService.GetApiGroups()
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithData(gin.H{
		"groups":      groups,
		"apiGroupMap": apiGroupMap,
	}, c)
}

// IgnoreApi
// @Tags      IgnoreApi
// @Summary   忽略API
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "同步API"
// @Router    /api/ignoreApi [post]
func (s *SystemApiApi) IgnoreApi(c *gin.Context) {
	var ignoreApi system.SysIgnoreApi
	err := c.ShouldBindJSON(&ignoreApi)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiService.IgnoreApi(ignoreApi)
	if err != nil {
		global.LRAG_LOG.Error("忽略失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.ignore_failed"), c)
		return
	}
	response.Ok(c)
}

// EnterSyncApi
// @Tags      SysApi
// @Summary   确认同步API
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "确认同步API"
// @Router    /api/enterSyncApi [post]
func (s *SystemApiApi) EnterSyncApi(c *gin.Context) {
	var syncApi systemRes.SysSyncApis
	err := c.ShouldBindJSON(&syncApi)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiService.EnterSyncApi(syncApi)
	if err != nil {
		global.LRAG_LOG.Error("忽略失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.ignore_failed"), c)
		return
	}
	response.Ok(c)
}

// DeleteApi
// @Tags      SysApi
// @Summary   删除api
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysApi                  true  "ID"
// @Success   200   {object}  response.Response{msg=string}  "删除api"
// @Router    /api/deleteApi [post]
func (s *SystemApiApi) DeleteApi(c *gin.Context) {
	var api system.SysApi
	err := c.ShouldBindJSON(&api)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(api.LRAG_MODEL, utils.IdVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiService.DeleteApi(api)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// GetApiList
// @Tags      SysApi
// @Summary   分页获取API列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SearchApiParams                               true  "分页获取API列表"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页获取API列表,返回包括列表,总数,页码,每页数量"
// @Router    /api/getApiList [post]
func (s *SystemApiApi) GetApiList(c *gin.Context) {
	var pageInfo systemReq.SearchApiParams
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(pageInfo.PageInfo, utils.PageInfoVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := apiService.GetAPIInfoList(pageInfo.SysApi, pageInfo.PageInfo, pageInfo.OrderKey, pageInfo.Desc)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetApiById
// @Tags      SysApi
// @Summary   根据id获取api
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                                   true  "根据id获取api"
// @Success   200   {object}  response.Response{data=systemRes.SysAPIResponse}  "根据id获取api,返回包括api详情"
// @Router    /api/getApiById [post]
func (s *SystemApiApi) GetApiById(c *gin.Context) {
	var idInfo request.GetById
	err := c.ShouldBindJSON(&idInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(idInfo, utils.IdVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	api, err := apiService.GetApiById(idInfo.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(systemRes.SysAPIResponse{Api: api}, i18n.Msg(c, "common.fetch_success"), c)
}

// UpdateApi
// @Tags      SysApi
// @Summary   修改基础api
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysApi                  true  "api路径, api中文描述, api组, 方法"
// @Success   200   {object}  response.Response{msg=string}  "修改基础api"
// @Router    /api/updateApi [post]
func (s *SystemApiApi) UpdateApi(c *gin.Context) {
	var api system.SysApi
	err := c.ShouldBindJSON(&api)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(api, utils.ApiVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiService.UpdateApi(api)
	if err != nil {
		global.LRAG_LOG.Error("修改失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.modify_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.modify_success"), c)
}

// GetAllApis
// @Tags      SysApi
// @Summary   获取所有的Api 不分页
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=systemRes.SysAPIListResponse,msg=string}  "获取所有的Api 不分页,返回包括api列表"
// @Router    /api/getAllApis [post]
func (s *SystemApiApi) GetAllApis(c *gin.Context) {
	authorityID := utils.GetUserAuthorityId(c)
	apis, err := apiService.GetAllApis(authorityID)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(systemRes.SysAPIListResponse{Apis: apis}, i18n.Msg(c, "common.fetch_success"), c)
}

// DeleteApisByIds
// @Tags      SysApi
// @Summary   删除选中Api
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.IdsReq                 true  "ID"
// @Success   200   {object}  response.Response{msg=string}  "删除选中Api"
// @Router    /api/deleteApisByIds [delete]
func (s *SystemApiApi) DeleteApisByIds(c *gin.Context) {
	var ids request.IdsReq
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = apiService.DeleteApisByIds(ids)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// FreshCasbin
// @Tags      SysApi
// @Summary   刷新casbin缓存
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "刷新成功"
// @Router    /api/freshCasbin [get]
func (s *SystemApiApi) FreshCasbin(c *gin.Context) {
	err := casbinService.FreshCasbin()
	if err != nil {
		global.LRAG_LOG.Error("刷新失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.refresh_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.refresh_success"), c)
}

// GetApiRoles
// @Tags      SysApi
// @Summary   获取拥有指定API权限的角色ID列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     path    query     string                                                    true  "API路径"
// @Param     method  query     string                                                    true  "请求方法"
// @Success   200     {object}  response.Response{data=map[string]interface{},msg=string}  "获取成功"
// @Router    /api/getApiRoles [get]
func (s *SystemApiApi) GetApiRoles(c *gin.Context) {
	path := c.Query("path")
	method := c.Query("method")
	if path == "" || method == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.api_path_method_required"), c)
		return
	}
	authorityIds, err := casbinService.GetAuthoritiesByApi(path, method)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.get_failed_detail", err.Error()), c)
		return
	}
	if authorityIds == nil {
		authorityIds = []uint{}
	}
	response.OkWithDetailed(authorityIds, i18n.Msg(c, "common.fetch_success"), c)
}

// SetApiRoles
// @Tags      SysApi
// @Summary   全量覆盖某API关联的角色列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SetApiAuthorities    true  "API路径、请求方法和角色ID列表"
// @Success   200   {object}  response.Response{msg=string}  "设置成功"
// @Router    /api/setApiRoles [post]
func (s *SystemApiApi) SetApiRoles(c *gin.Context) {
	var req systemReq.SetApiAuthorities
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if req.Path == "" || req.Method == "" {
		response.FailWithMessage(i18n.Msg(c, "validation.api_path_method_required"), c)
		return
	}
	if err := casbinService.SetApiAuthorities(req.Path, req.Method, req.AuthorityIds); err != nil {
		global.LRAG_LOG.Error("设置失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.set_failed_detail", err.Error()), c)
		return
	}
	// 刷新casbin缓存使策略立即生效
	_ = casbinService.FreshCasbin()
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}
