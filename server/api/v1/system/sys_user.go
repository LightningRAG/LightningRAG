package system

import (
	"strconv"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common"
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Login
// @Tags     Base
// @Summary  用户登录
// @Produce   application/json
// @Param    data  body      systemReq.Login                                             true  "用户名, 密码, 验证码"
// @Success  200   {object}  response.Response{data=systemRes.LoginResponse,msg=string}  "返回包括用户信息,token,过期时间"
// @Router   /base/login [post]
func (b *BaseApi) Login(c *gin.Context) {
	var l systemReq.Login
	err := c.ShouldBindJSON(&l)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(l, utils.LoginVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}

	key := c.ClientIP()
	// 判断验证码是否开启
	openCaptcha := global.LRAG_CONFIG.Captcha.OpenCaptcha               // 是否开启防爆次数
	openCaptchaTimeOut := global.LRAG_CONFIG.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	v, ok := global.BlackCache.Get(key)
	if !ok {
		global.BlackCache.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}

	var oc bool = openCaptcha == 0 || openCaptcha < interfaceToInt(v)
	if oc && (l.Captcha == "" || l.CaptchaId == "" || !store.Verify(l.CaptchaId, l.Captcha, true)) {
		// 验证码次数+1
		global.BlackCache.Increment(key, 1)
		msg := i18n.Msg(c, "sys.login.captcha_wrong")
		response.FailWithMessage(msg, c)
		// 记录登录失败日志（存 i18n 键，便于日志页按界面语言展示）
		loginLogService.CreateLoginLog(system.SysLoginLog{
			Username:     l.Username,
			Ip:           c.ClientIP(),
			Agent:        c.Request.UserAgent(),
			Status:       false,
			ErrorMessage: "sys.login.captcha_wrong",
		})
		return
	}

	u := &system.SysUser{Username: l.Username, Password: l.Password}
	user, err := userService.Login(u)
	if err != nil {
		global.LRAG_LOG.Error("登陆失败! 用户名不存在或者密码错误!", zap.Error(err))
		// 验证码次数+1
		global.BlackCache.Increment(key, 1)
		msg := i18n.Msg(c, "sys.login.bad_credentials")
		response.FailWithMessage(msg, c)
		loginLogService.CreateLoginLog(system.SysLoginLog{
			Username:     l.Username,
			Ip:           c.ClientIP(),
			Agent:        c.Request.UserAgent(),
			Status:       false,
			ErrorMessage: "sys.login.bad_credentials",
		})
		return
	}
	if user.Enable != 1 {
		global.LRAG_LOG.Error("登陆失败! 用户被禁止登录!")
		// 验证码次数+1
		global.BlackCache.Increment(key, 1)
		msg := i18n.Msg(c, "sys.login.user_disabled")
		response.FailWithMessage(msg, c)
		loginLogService.CreateLoginLog(system.SysLoginLog{
			Username:     l.Username,
			Ip:           c.ClientIP(),
			Agent:        c.Request.UserAgent(),
			Status:       false,
			ErrorMessage: "sys.login.user_disabled",
			UserID:       user.ID,
		})
		return
	}
	b.TokenNext(c, *user)
}

// TokenNext 登录以后签发jwt
func (b *BaseApi) TokenNext(c *gin.Context, user system.SysUser) {
	loginResp, failKey := userService.FinishLoginSession(c, &user)
	if failKey != "" {
		response.FailWithMessage(i18n.Msg(c, failKey), c)
		return
	}
	response.OkWithDetailed(loginResp, i18n.Msg(c, "common.login_success"), c)
}

// Register
// @Tags     SysUser
// @Summary  用户注册账号
// @Produce   application/json
// @Param    data  body      systemReq.Register                                            true  "用户名, 昵称, 密码, 角色ID"
// @Success  200   {object}  response.Response{data=systemRes.SysUserResponse,msg=string}  "用户注册账号,返回包括用户信息"
// @Router   /user/admin_register [post]
func (b *BaseApi) Register(c *gin.Context) {
	var r systemReq.Register
	err := c.ShouldBindJSON(&r)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(r, utils.RegisterVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	var authorities []system.SysAuthority
	for _, v := range r.AuthorityIds {
		authorities = append(authorities, system.SysAuthority{
			AuthorityId: v,
		})
	}
	user := &system.SysUser{Username: r.Username, NickName: r.NickName, Password: r.Password, HeaderImg: r.HeaderImg, AuthorityId: r.AuthorityId, Authorities: authorities, Enable: r.Enable, Phone: r.Phone, Email: r.Email}
	userReturn, err := userService.Register(*user)
	if err != nil {
		global.LRAG_LOG.Error("注册失败!", zap.Error(err))
		response.FailWithDetailed(systemRes.SysUserResponse{User: userReturn}, i18n.Msg(c, "common.register_failed"), c)
		return
	}
	response.OkWithDetailed(systemRes.SysUserResponse{User: userReturn}, i18n.Msg(c, "common.register_success"), c)
}

// ChangePassword
// @Tags      SysUser
// @Summary   用户修改密码
// @Security  ApiKeyAuth
// @Produce  application/json
// @Param     data  body      systemReq.ChangePasswordReq    true  "用户名, 原密码, 新密码"
// @Success   200   {object}  response.Response{msg=string}  "用户修改密码"
// @Router    /user/changePassword [post]
func (b *BaseApi) ChangePassword(c *gin.Context) {
	var req systemReq.ChangePasswordReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(req, utils.ChangePasswordVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	u := &system.SysUser{LRAG_MODEL: global.LRAG_MODEL{ID: uid}, Password: req.Password}
	err = userService.ChangePassword(u, req.NewPassword)
	if err != nil {
		global.LRAG_LOG.Error("修改失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.modify_failed_wrong_password"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.modify_success"), c)
}

// GetUserList
// @Tags      SysUser
// @Summary   分页获取用户列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.GetUserList                                        true  "页码, 每页大小"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页获取用户列表,返回包括列表,总数,页码,每页数量"
// @Router    /user/getUserList [post]
func (b *BaseApi) GetUserList(c *gin.Context) {
	var pageInfo systemReq.GetUserList
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(pageInfo, utils.PageInfoVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	list, total, err := userService.GetUserInfoList(pageInfo)
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
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// SetUserAuthority
// @Tags      SysUser
// @Summary   更改用户权限
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SetUserAuth          true  "用户UUID, 角色ID"
// @Success   200   {object}  response.Response{msg=string}  "设置用户权限"
// @Router    /user/setUserAuthority [post]
func (b *BaseApi) SetUserAuthority(c *gin.Context) {
	var sua systemReq.SetUserAuth
	err := c.ShouldBindJSON(&sua)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if UserVerifyErr := utils.Verify(sua, utils.SetUserAuthorityVerify); UserVerifyErr != nil {
		response.FailWithError(c, UserVerifyErr)
		return
	}
	userID := utils.GetUserID(c)
	err = userService.SetUserAuthority(userID, sua.AuthorityId)
	if err != nil {
		global.LRAG_LOG.Error("修改失败!", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	claims := utils.GetUserInfo(c)
	claims.AuthorityId = sua.AuthorityId
	token, err := utils.NewJWT().CreateToken(*claims)
	if err != nil {
		global.LRAG_LOG.Error("修改失败!", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	c.Header("new-token", token)
	c.Header("new-expires-at", strconv.FormatInt(claims.ExpiresAt.Unix(), 10))
	utils.SetToken(c, token, int(claims.ExpiresAt.Unix()-time.Now().Unix()))
	response.OkWithMessage(i18n.Msg(c, "common.modify_success"), c)
}

// SetUserAuthorities
// @Tags      SysUser
// @Summary   设置用户权限
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SetUserAuthorities   true  "用户UUID, 角色ID"
// @Success   200   {object}  response.Response{msg=string}  "设置用户权限"
// @Router    /user/setUserAuthorities [post]
func (b *BaseApi) SetUserAuthorities(c *gin.Context) {
	var sua systemReq.SetUserAuthorities
	err := c.ShouldBindJSON(&sua)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	authorityID := utils.GetUserAuthorityId(c)
	err = userService.SetUserAuthorities(authorityID, sua.ID, sua.AuthorityIds)
	if err != nil {
		global.LRAG_LOG.Error("修改失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.modify_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.modify_success"), c)
}

// DeleteUser
// @Tags      SysUser
// @Summary   删除用户
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                true  "用户ID"
// @Success   200   {object}  response.Response{msg=string}  "删除用户"
// @Router    /user/deleteUser [delete]
func (b *BaseApi) DeleteUser(c *gin.Context) {
	var reqId request.GetById
	err := c.ShouldBindJSON(&reqId)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(reqId, utils.IdVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	jwtId := utils.GetUserID(c)
	if jwtId == uint(reqId.ID) {
		response.FailWithMessage(i18n.Msg(c, "common.cannot_delete_self"), c)
		return
	}
	err = userService.DeleteUser(reqId.ID)
	if err != nil {
		global.LRAG_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.delete_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// SetUserInfo
// @Tags      SysUser
// @Summary   设置用户信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysUser                                             true  "ID, 用户名, 昵称, 头像链接"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "设置用户信息"
// @Router    /user/setUserInfo [put]
func (b *BaseApi) SetUserInfo(c *gin.Context) {
	var user systemReq.ChangeUserInfo
	err := c.ShouldBindJSON(&user)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = utils.Verify(user, utils.IdVerify)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	if len(user.AuthorityIds) != 0 {
		authorityID := utils.GetUserAuthorityId(c)
		err = userService.SetUserAuthorities(authorityID, user.ID, user.AuthorityIds)
		if err != nil {
			global.LRAG_LOG.Error("设置失败!", zap.Error(err))
			response.FailWithMessage(i18n.Msg(c, "common.set_failed"), c)
			return
		}
	}
	err = userService.SetUserInfo(system.SysUser{
		LRAG_MODEL: global.LRAG_MODEL{
			ID: user.ID,
		},
		NickName:  user.NickName,
		HeaderImg: user.HeaderImg,
		Phone:     user.Phone,
		Email:     user.Email,
		Enable:    user.Enable,
	})
	if err != nil {
		global.LRAG_LOG.Error("设置失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.set_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// SetSelfInfo
// @Tags      SysUser
// @Summary   设置用户信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysUser                                             true  "ID, 用户名, 昵称, 头像链接"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "设置用户信息"
// @Router    /user/SetSelfInfo [put]
func (b *BaseApi) SetSelfInfo(c *gin.Context) {
	var user systemReq.ChangeUserInfo
	err := c.ShouldBindJSON(&user)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	user.ID = utils.GetUserID(c)
	err = userService.SetSelfInfo(system.SysUser{
		LRAG_MODEL: global.LRAG_MODEL{
			ID: user.ID,
		},
		NickName:  user.NickName,
		HeaderImg: user.HeaderImg,
		Phone:     user.Phone,
		Email:     user.Email,
		Enable:    user.Enable,
	})
	if err != nil {
		global.LRAG_LOG.Error("设置失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.set_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// SetSelfSetting
// @Tags      SysUser
// @Summary   设置用户配置
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      map[string]interface{}  true  "用户配置数据"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "设置用户配置"
// @Router    /user/SetSelfSetting [put]
func (b *BaseApi) SetSelfSetting(c *gin.Context) {
	var req common.JSONMap
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(c, err)
		return
	}

	err = userService.SetSelfSetting(req, utils.GetUserID(c))
	if err != nil {
		global.LRAG_LOG.Error("设置失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.set_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.set_success"), c)
}

// GetUserInfo
// @Tags      SysUser
// @Summary   获取用户信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=map[string]interface{},msg=string}  "获取用户信息"
// @Router    /user/getUserInfo [get]
func (b *BaseApi) GetUserInfo(c *gin.Context) {
	uuid := utils.GetUserUuid(c)
	ReqUser, err := userService.GetUserInfo(uuid)
	if err != nil {
		global.LRAG_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithDetailed(gin.H{"userInfo": ReqUser}, i18n.Msg(c, "common.fetch_success"), c)
}

// ResetPassword
// @Tags      SysUser
// @Summary   重置用户密码
// @Security  ApiKeyAuth
// @Produce  application/json
// @Param     data  body      system.SysUser                 true  "ID"
// @Success   200   {object}  response.Response{msg=string}  "重置用户密码"
// @Router    /user/resetPassword [post]
func (b *BaseApi) ResetPassword(c *gin.Context) {
	var rps systemReq.ResetPassword
	err := c.ShouldBindJSON(&rps)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = userService.ResetPassword(rps.ID, rps.Password)
	if err != nil {
		global.LRAG_LOG.Error("重置失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msgf(c, "common.reset_failed_detail", err.Error()), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.reset_success"), c)
}
