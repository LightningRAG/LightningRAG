package middleware

import (
	"errors"
	"strconv"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/golang-jwt/jwt/v5"

	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 我们这里jwt鉴权取头部信息 x-token 登录时回返回token信息 这里前端需要把token存储到cookie或者本地localStorage中 不过需要跟后端协商过期时间 可以约定刷新令牌或者重新登录
		token := utils.GetToken(c)
		if token == "" {
			response.NoAuth(i18n.Msg(c, "auth.not_logged_in"), c)
			c.Abort()
			return
		}
		if isBlacklist(token) {
			response.NoAuth(i18n.Msg(c, "auth.token_blacklisted"), c)
			utils.ClearToken(c)
			c.Abort()
			return
		}
		j := utils.NewJWT()
		// parseToken 解析token包含的信息
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.TokenExpired) {
				response.NoAuth(i18n.Msg(c, "auth.token_expired"), c)
				utils.ClearToken(c)
				c.Abort()
				return
			}
			response.NoAuth(jwtFailMessage(c, err), c)
			utils.ClearToken(c)
			c.Abort()
			return
		}

		// 已登录用户被管理员禁用 需要使该用户的jwt失效 此处比较消耗性能 如果需要 请自行打开
		// 用户被删除的逻辑 需要优化 此处比较消耗性能 如果需要 请自行打开

		//if user, err := userService.FindUserByUuid(claims.UUID.String()); err != nil || user.Enable == 2 {
		//	_ = jwtService.JsonInBlacklist(system.JwtBlacklist{Jwt: token})
		//	response.FailWithDetailed(gin.H{"reload": true}, err.Error(), c)
		//	c.Abort()
		//}
		c.Set("claims", claims)
		if claims.ExpiresAt != nil && claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, durErr := utils.ParseDuration(global.LRAG_CONFIG.JWT.ExpiresTime)
			if durErr == nil {
				claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(dr))
				newToken, ctErr := j.CreateTokenByOldToken(token, *claims)
				if ctErr == nil && newToken != "" {
					newClaims, ptErr := j.ParseToken(newToken)
					if ptErr == nil && newClaims != nil && newClaims.ExpiresAt != nil {
						c.Header("new-token", newToken)
						c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
						// http.Cookie MaxAge 单位为秒，与登录流程一致使用 dr 的总秒数
						utils.SetToken(c, newToken, int(dr.Seconds()))
						if global.LRAG_CONFIG.System.UseMultipoint {
							_ = utils.SetRedisJWT(newToken, newClaims.Username)
						}
					}
				}
			}
		}
		c.Next()

		if newToken, exists := c.Get("new-token"); exists {
			if s, ok := newToken.(string); ok {
				c.Header("new-token", s)
			}
		}
		if newExpiresAt, exists := c.Get("new-expires-at"); exists {
			if s, ok := newExpiresAt.(string); ok {
				c.Header("new-expires-at", s)
			}
		}
	}
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: IsBlacklist
//@description: 判断JWT是否在黑名单内部
//@param: jwt string
//@return: bool

func isBlacklist(jwt string) bool {
	_, ok := global.BlackCache.Get(jwt)
	return ok
}

func jwtFailMessage(c *gin.Context, err error) string {
	switch {
	case errors.Is(err, utils.TokenMalformed):
		return i18n.Msg(c, "auth.token_malformed")
	case errors.Is(err, utils.TokenSignatureInvalid):
		return i18n.Msg(c, "auth.token_signature_invalid")
	case errors.Is(err, utils.TokenNotValidYet):
		return i18n.Msg(c, "auth.token_not_valid_yet")
	case errors.Is(err, utils.TokenInvalid):
		return i18n.Msg(c, "auth.token_invalid")
	case errors.Is(err, utils.TokenValid):
		return i18n.Msg(c, "auth.token_unknown")
	default:
		return i18n.Msg(c, "auth.token_invalid")
	}
}
