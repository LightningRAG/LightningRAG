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

type AuthorityMenuApi struct{}

// GetMenu
// @Tags      AuthorityMenu
// @Summary   获取用户动态路由
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      request.Empty                                                  true  "空"
// @Success   200   {object}  response.Response{data=systemRes.SysMenusResponse,msg=string}  "获取用户动态路由,返回包括系统菜单详情列表"
// @Router    /menu/getMenu [post]
func (a *AuthorityMenuApi) GetMenu(c *gin.Context) {
	menus, err := menuService.GetMenuTree(utils.GetUserAuthorityId(c))
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	if menus == nil {
		menus = []system.SysMenu{}
	}
	response.OkWithDetailed(systemRes.SysMenusResponse{Menus: menus}, i18n.Msg(c, "common.fetch_success"), c)
}

// GetBaseMenuTree
// @Tags      AuthorityMenu
// @Summary   获取用户动态路由
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      request.Empty                                                      true  "空"
// @Success   200   {object}  response.Response{data=systemRes.SysBaseMenusResponse,msg=string}  "获取用户动态路由,返回包括系统菜单列表"
// @Router    /menu/getBaseMenuTree [post]
func (a *AuthorityMenuApi) GetBaseMenuTree(c *gin.Context) {
	authority := utils.GetUserAuthorityId(c)
	menus, err := menuService.GetBaseMenuTree(authority)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(systemRes.SysBaseMenusResponse{Menus: menus}, i18n.Msg(c, "common.fetch_success"), c)
}

// AddMenuAuthority
// @Tags      AuthorityMenu
// @Summary   增加menu和角色关联关系
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.AddMenuAuthorityInfo  true  "角色ID"
// @Success   200   {object}  response.Response{msg=string}   "增加menu和角色关联关系"
// @Router    /menu/addMenuAuthority [post]
func (a *AuthorityMenuApi) AddMenuAuthority(c *gin.Context) {
	var authorityMenu systemReq.AddMenuAuthorityInfo
	err := c.ShouldBindJSON(&authorityMenu)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if err := utils.Verify(authorityMenu, utils.AuthorityIdVerify); err != nil {
		response.FailWithError(c, err)
		return
	}
	adminAuthorityID := utils.GetUserAuthorityId(c)
	if err := menuService.AddMenuAuthority(authorityMenu.Menus, adminAuthorityID, authorityMenu.AuthorityId); err != nil {
		global.LRAG_LOG.Error("添加失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.create_failed"), c)
	} else {
		response.OkWithMessage(i18n.Msg(c, "common.add_success"), c)
	}
}

// GetMenuAuthority
// @Tags      AuthorityMenu
// @Summary   获取指定角色menu
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetAuthorityId                                     true  "角色ID"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "获取指定角色menu"
// @Router    /menu/getMenuAuthority [post]
func (a *AuthorityMenuApi) GetMenuAuthority(c *gin.Context) {
	var param request.GetAuthorityId
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(param, utils.AuthorityIdVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	menus, err := menuService.GetMenuAuthority(&param)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithDetailed(systemRes.SysMenusResponse{Menus: menus}, i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(gin.H{"menus": menus}, i18n.Msg(c, "common.fetch_success"), c)
}

// AddBaseMenu
// @Tags      Menu
// @Summary   新增菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysBaseMenu             true  "路由path, 父菜单ID, 路由name, 对应前端文件路径, 排序标记"
// @Success   200   {object}  response.Response{msg=string}  "新增菜单"
// @Router    /menu/addBaseMenu [post]
func (a *AuthorityMenuApi) AddBaseMenu(c *gin.Context) {
	var menu system.SysBaseMenu
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(menu, utils.MenuVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(menu.Meta, utils.MenuMetaVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = menuService.AddBaseMenu(menu)
	if err != nil {
		global.LRAG_LOG.Error("添加失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.add_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.add_success"), c)
}

// DeleteBaseMenu
// @Tags      Menu
// @Summary   删除菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                true  "菜单id"
// @Success   200   {object}  response.Response{msg=string}  "删除菜单"
// @Router    /menu/deleteBaseMenu [post]
func (a *AuthorityMenuApi) DeleteBaseMenu(c *gin.Context) {
	var menu request.GetById
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(menu, utils.IdVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = baseMenuService.DeleteBaseMenu(menu.ID)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.delete_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// UpdateBaseMenu
// @Tags      Menu
// @Summary   更新菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysBaseMenu             true  "路由path, 父菜单ID, 路由name, 对应前端文件路径, 排序标记"
// @Success   200   {object}  response.Response{msg=string}  "更新菜单"
// @Router    /menu/updateBaseMenu [post]
func (a *AuthorityMenuApi) UpdateBaseMenu(c *gin.Context) {
	var menu system.SysBaseMenu
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(menu, utils.MenuVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(menu.Meta, utils.MenuMetaVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = baseMenuService.UpdateBaseMenu(menu)
	if err != nil {
		global.LRAG_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.update_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

// GetBaseMenuById
// @Tags      Menu
// @Summary   根据id获取菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                                                   true  "菜单id"
// @Success   200   {object}  response.Response{data=systemRes.SysBaseMenuResponse,msg=string}  "根据id获取菜单,返回包括系统菜单列表"
// @Router    /menu/getBaseMenuById [post]
func (a *AuthorityMenuApi) GetBaseMenuById(c *gin.Context) {
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
	menu, err := baseMenuService.GetBaseMenuById(idInfo.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(systemRes.SysBaseMenuResponse{Menu: menu}, i18n.Msg(c, "common.fetch_success"), c)
}

// GetMenuRoles
// @Tags      AuthorityMenu
// @Summary   获取拥有指定菜单的角色ID列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     menuId  query     uint                                                         true  "菜单ID"
// @Success   200     {object}  response.Response{data=map[string]interface{},msg=string}    "获取成功"
// @Router    /menu/getMenuRoles [get]
func (a *AuthorityMenuApi) GetMenuRoles(c *gin.Context) {
	var req systemReq.SetMenuAuthorities
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if req.MenuId == 0 {
		response.FailWithMessage(i18n.Msg(c, "validation.menu_id_required"), c)
		return
	}
	authorityIds, err := menuService.GetAuthoritiesByMenuId(req.MenuId)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.get_failed_detail", err.Error()), c)
		return
	}
	if authorityIds == nil {
		authorityIds = []uint{}
	}
	defaultRouterAuthorityIds, err := menuService.GetDefaultRouterAuthorityIds(req.MenuId)
	if err != nil {
		global.LRAG_LOG.Error("获取首页角色失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.get_failed_detail", err.Error()), c)
		return
	}
	if defaultRouterAuthorityIds == nil {
		defaultRouterAuthorityIds = []uint{}
	}
	response.OkWithDetailed(gin.H{
		"authorityIds":              authorityIds,
		"defaultRouterAuthorityIds": defaultRouterAuthorityIds,
	}, "获取成功", c)
}

// SetMenuRoles
// @Tags      AuthorityMenu
// @Summary   全量覆盖某菜单关联的角色列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SetMenuAuthorities   true  "菜单ID和角色ID列表"
// @Success   200   {object}  response.Response{msg=string}  "设置成功"
// @Router    /menu/setMenuRoles [post]
func (a *AuthorityMenuApi) SetMenuRoles(c *gin.Context) {
	var req systemReq.SetMenuAuthorities
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	if req.MenuId == 0 {
		response.FailWithMessage(i18n.Msg(c, "validation.menu_id_required"), c)
		return
	}
	if err := menuService.SetMenuAuthorities(req.MenuId, req.AuthorityIds); err != nil {
		global.LRAG_LOG.Error("设置失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.set_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// GetMenuList
// @Tags      Menu
// @Summary   分页获取基础menu列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.PageInfo                                        true  "页码, 每页大小"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页获取基础menu列表,返回包括列表,总数,页码,每页数量"
// @Router    /menu/getMenuList [post]
func (a *AuthorityMenuApi) GetMenuList(c *gin.Context) {
	authorityID := utils.GetUserAuthorityId(c)
	menuList, err := menuService.GetInfoList(authorityID)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(menuList, i18n.Msg(c, "common.fetch_success"), c)
}
