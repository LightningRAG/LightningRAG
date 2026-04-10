package system

import (
	"errors"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// FinishLoginSession 签发 JWT、处理多点登录与 Cookie。failMsgKey 非空时为 i18n 键，与原有 TokenNext 行为一致。
func (userService *UserService) FinishLoginSession(c *gin.Context, user *system.SysUser) (loginResp systemRes.LoginResponse, failMsgKey string) {
	token, claims, err := utils.LoginToken(user)
	if err != nil {
		global.LRAG_LOG.Error("获取token失败!", zap.Error(err))
		return loginResp, "sys.login.token_failed"
	}
	LoginLogServiceApp.CreateLoginLog(system.SysLoginLog{
		Username:     user.Username,
		Ip:           c.ClientIP(),
		Agent:        c.Request.UserAgent(),
		Status:       true,
		UserID:       user.ID,
		ErrorMessage: "common.login_success",
	})
	if !global.LRAG_CONFIG.System.UseMultipoint {
		utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		return systemRes.LoginResponse{
			User:      *user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, ""
	}

	jwtSvc := JwtServiceApp
	if jwtStr, err := jwtSvc.GetRedisJWT(user.Username); err == redis.Nil {
		if err := utils.SetRedisJWT(token, user.Username); err != nil {
			global.LRAG_LOG.Error("设置登录状态失败!", zap.Error(err))
			return loginResp, "sys.login.session_setup_failed"
		}
		utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		return systemRes.LoginResponse{
			User:      *user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, ""
	} else if err != nil {
		global.LRAG_LOG.Error("设置登录状态失败!", zap.Error(err))
		return loginResp, "sys.login.session_setup_failed"
	} else {
		var blackJWT system.JwtBlacklist
		blackJWT.Jwt = jwtStr
		if err := jwtSvc.JsonInBlacklist(blackJWT); err != nil {
			return loginResp, "sys.login.jwt_revoke_failed"
		}
		if err := utils.SetRedisJWT(token, user.GetUsername()); err != nil {
			return loginResp, "sys.login.session_setup_failed"
		}
		utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		return systemRes.LoginResponse{
			User:      *user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, ""
	}
}

// FinishLoginSessionErr OAuth 等场景需要 error 返回值时使用
func (userService *UserService) FinishLoginSessionErr(c *gin.Context, user *system.SysUser) (systemRes.LoginResponse, error) {
	resp, key := userService.FinishLoginSession(c, user)
	if key != "" {
		return resp, errors.New(key)
	}
	return resp, nil
}
