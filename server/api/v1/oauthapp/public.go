package oauthapp

import (
	"net/url"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	sysmodel "github.com/LightningRAG/LightningRAG/server/model/system"
	idpoauth "github.com/LightningRAG/LightningRAG/server/oauth"
	"github.com/LightningRAG/LightningRAG/server/service/oauthapp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OAuthPublicApi struct{}

// OAuthPublicProviders 登录页：已启用的第三方登录列表，并返回回调路径模板（含 router-prefix），便于前端拼授权 URL。
func (a *OAuthPublicApi) OAuthPublicProviders(c *gin.Context) {
	list, err := oauthapp.SysOAuthProviderServiceApp.ListPublicEnabled()
	if err != nil {
		global.LRAG_LOG.Error("oauth list public", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "common.get_failed"), c)
		return
	}
	response.OkWithData(gin.H{
		"providers":                list,
		"callbackPathPattern":      oauthapp.OAuthCallbackPathPattern(),
		"defaultButtonIconsByKind": idpoauth.CloneDefaultButtonIcons(),
	}, c)
}

// OAuthAuthorize 跳转授权页
func (a *OAuthPublicApi) OAuthAuthorize(c *gin.Context) {
	kind := c.Param("kind")
	ret := c.Query("redirect")
	target, err := oauthapp.OAuthFlowServiceApp.OAuthAuthorizeRedirectURL(c, kind, ret)
	if err != nil {
		global.LRAG_LOG.Warn("oauth authorize", zap.String("kind", kind), zap.Error(err))
		fe := oauthapp.EffectiveOAuthFrontendRedirect()
		sep := "?"
		if strings.Contains(fe, "?") {
			sep = "&"
		}
		c.Redirect(302, fe+sep+"oauth_err="+url.QueryEscape("authorize"))
		return
	}
	c.Redirect(302, target)
}

// OAuthCallback IdP 回调：换票后重定向回前端
func (a *OAuthPublicApi) OAuthCallback(c *gin.Context) {
	kind := c.Param("kind")
	fe := oauthapp.EffectiveOAuthFrontendRedirect()
	sep := "?"
	if strings.Contains(fe, "?") {
		sep = "&"
	}
	if idpErr := strings.TrimSpace(c.Query("error")); idpErr != "" {
		reason := "denied"
		low := strings.ToLower(idpErr)
		if low != "access_denied" && low != "user_cancelled" && low != "cancelled" && low != "interaction_required" {
			reason = "idp"
		}
		global.LRAG_LOG.Info("oauth callback idp error", zap.String("kind", kind), zap.String("error", idpErr))
		loginLogService.CreateLoginLog(sysmodel.SysLoginLog{
			Username:     "oauth:" + kind,
			Ip:           c.ClientIP(),
			Agent:        c.Request.UserAgent(),
			Status:       false,
			ErrorMessage: idpErr,
		})
		c.Redirect(302, fe+sep+"oauth_err="+url.QueryEscape(reason))
		return
	}
	code := c.Query("code")
	state := c.Query("state")
	redir, reason, err := oauthapp.OAuthFlowServiceApp.OAuthHandleCallback(c, kind, code, state)
	if err != nil {
		global.LRAG_LOG.Warn("oauth callback", zap.String("kind", kind), zap.String("reason", reason), zap.Error(err))
		loginLogService.CreateLoginLog(sysmodel.SysLoginLog{
			Username:     "oauth:" + kind,
			Ip:           c.ClientIP(),
			Agent:        c.Request.UserAgent(),
			Status:       false,
			ErrorMessage: "sys.oauth.callback_failed",
		})
		if reason == "" {
			reason = "unknown"
		}
		c.Redirect(302, fe+sep+"oauth_err="+url.QueryEscape(reason))
		return
	}
	c.Redirect(302, redir)
}

// OAuthExchange 前端用一次性 oauth_ex 换取登录 JSON
func (a *OAuthPublicApi) OAuthExchange(c *gin.Context) {
	exID := strings.TrimSpace(c.Query("oauth_ex"))
	if exID == "" {
		response.FailWithMessage(i18n.Msg(c, "sys.oauth.exchange_missing"), c)
		return
	}
	if !oauthapp.ValidOAuthExchangeToken(exID) {
		response.FailWithMessage(i18n.Msg(c, "sys.oauth.exchange_invalid"), c)
		return
	}
	data, ok := oauthapp.OAuthFlowServiceApp.OAuthExchangeForJSON(exID)
	if !ok {
		response.FailWithMessage(i18n.Msg(c, "sys.oauth.exchange_invalid"), c)
		return
	}
	response.OkWithDetailed(data, i18n.Msg(c, "common.login_success"), c)
}
