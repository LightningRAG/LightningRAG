package oauthapp

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	sysmodel "github.com/LightningRAG/LightningRAG/server/model/system"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	idpoauth "github.com/LightningRAG/LightningRAG/server/oauth"
	syssvc "github.com/LightningRAG/LightningRAG/server/service/system"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type OAuthFlowService struct{}

var OAuthFlowServiceApp = new(OAuthFlowService)

const (
	oauthStateTTL    = 10 * time.Minute
	oauthExchangeTTL = 2 * time.Minute
	oauthHTTPTimeout = 45 * time.Second
	oauthExTokenMin  = 20
	oauthExTokenMax  = 128
)

// ValidOAuthExchangeToken 与 randomURLState 生成格式一致（base64url），拒绝明显非法入参，减轻缓存键探测。
func ValidOAuthExchangeToken(s string) bool {
	n := len(s)
	if n < oauthExTokenMin || n > oauthExTokenMax {
		return false
	}
	for i := 0; i < n; i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '-', c == '_':
		default:
			return false
		}
	}
	return true
}

type oauthStatePayload struct {
	Kind       string `json:"k"`
	Verifier   string `json:"v"`
	ReturnPath string `json:"rp,omitempty"`
}

type oauthExchangePayload struct {
	Login    systemRes.LoginResponse `json:"login"`
	Redirect string                  `json:"rp,omitempty"`
}

func randomURLState() (string, error) {
	b := make([]byte, 24)
	if _, err := io.ReadFull(crand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func publicBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if xfp := c.GetHeader("X-Forwarded-Proto"); xfp == "https" || xfp == "http" {
		scheme = xfp
	}
	return scheme + "://" + c.Request.Host
}

func oauthCallbackURL(c *gin.Context, kind string) string {
	prefix := strings.TrimSuffix(global.LRAG_CONFIG.System.RouterPrefix, "/")
	return publicBaseURL(c) + prefix + "/base/oauth/callback/" + kind
}

func sanitizeOAuthReturnPath(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "/") || strings.Contains(s, "//") || strings.ContainsAny(s, "?\n\r") {
		return ""
	}
	return s
}

func oauthRuntimeConfig(row sysmodel.SysOAuthProvider, secret, redirect string) idpoauth.RuntimeConfig {
	ex := map[string]any(nil)
	if row.Extra != nil {
		ex = row.Extra
	}
	p, _ := idpoauth.Lookup(row.Kind)
	defScopes := []string(nil)
	if p != nil {
		defScopes = p.DefaultScopes()
	}
	return idpoauth.RuntimeConfig{
		ClientID:     row.ClientID,
		ClientSecret: secret,
		RedirectURI:  redirect,
		Scopes:       idpoauth.MergeScopes(row.Scopes, defScopes),
		Extra:        ex,
	}
}

// OAuthAuthorizeRedirectURL 构造跳转 IdP 的 URL（不写 response）
func (s *OAuthFlowService) OAuthAuthorizeRedirectURL(c *gin.Context, kind, returnQuery string) (string, error) {
	row, sec, err := SysOAuthProviderServiceApp.GetByKindForFlow(kind)
	if err != nil {
		return "", err
	}
	canon := strings.ToLower(strings.TrimSpace(row.Kind))
	p, err := idpoauth.Lookup(canon)
	if err != nil {
		return "", err
	}
	state, err := randomURLState()
	if err != nil {
		return "", err
	}
	var verifier string
	var authURL string
	rc := oauthRuntimeConfig(row, sec, oauthCallbackURL(c, canon))
	conf := p.OAuth2Config(rc)
	if ab, ok := p.(idpoauth.AuthorizeURLBuilder); ok {
		authURL, err = ab.BuildAuthorizeURL(rc, state)
		if err != nil {
			return "", err
		}
	} else {
		verifier = oauth2.GenerateVerifier()
		authURL = conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	}
	payload := oauthStatePayload{
		Kind:       canon,
		Verifier:   verifier,
		ReturnPath: sanitizeOAuthReturnPath(returnQuery),
	}
	b, _ := json.Marshal(payload)
	global.BlackCache.Set("oauth:st:"+state, string(b), oauthStateTTL)
	return authURL, nil
}

// OAuthHandleCallback 换票并返回前端重定向绝对 URL（带 oauth_ex）。
// 失败时 redirect 为空；reason 为简短英文码，供登录页 query oauth_err 展示差异化提示。
func (s *OAuthFlowService) OAuthHandleCallback(c *gin.Context, kind, code, state string) (redirect string, reason string, err error) {
	if code == "" || state == "" {
		return "", "missing", errors.New("missing code or state")
	}
	raw, ok := global.BlackCache.Get("oauth:st:" + state)
	if !ok {
		return "", "state", errors.New("state expired or invalid")
	}
	global.BlackCache.Delete("oauth:st:" + state)

	var st oauthStatePayload
	if err := json.Unmarshal([]byte(raw.(string)), &st); err != nil {
		return "", "state", err
	}
	flowKind := strings.ToLower(strings.TrimSpace(st.Kind))
	if flowKind == "" {
		return "", "state", errors.New("invalid state")
	}
	if strings.ToLower(strings.TrimSpace(kind)) != flowKind {
		return "", "state", errors.New("state kind mismatch")
	}

	row, sec, err := SysOAuthProviderServiceApp.GetByKindForFlow(flowKind)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "provider", err
		}
		return "", "provider", err
	}
	p, err := idpoauth.Lookup(flowKind)
	if err != nil {
		return "", "provider", err
	}
	rc := oauthRuntimeConfig(row, sec, oauthCallbackURL(c, flowKind))
	conf := p.OAuth2Config(rc)
	ctx, cancel := context.WithTimeout(c.Request.Context(), oauthHTTPTimeout)
	defer cancel()
	var tok *oauth2.Token
	if cx, ok := p.(idpoauth.CodeExchanger); ok {
		tok, err = cx.ExchangeCode(ctx, code, rc)
	} else {
		var exOpts []oauth2.AuthCodeOption
		if strings.TrimSpace(st.Verifier) != "" {
			exOpts = append(exOpts, oauth2.VerifierOption(st.Verifier))
		}
		tok, err = conf.Exchange(ctx, code, exOpts...)
	}
	if err != nil {
		return "", "token", err
	}
	prof, err := p.FetchProfile(ctx, tok, rc)
	if err != nil {
		return "", "token", err
	}
	if prof.Subject == "" {
		return "", "token", errors.New("empty subject from provider")
	}

	user, err := s.ensureOAuthUser(c, flowKind, prof, row)
	if err != nil {
		return "", "user", err
	}
	if user.Enable != 1 {
		return "", "user", errors.New("user disabled")
	}

	var full sysmodel.SysUser
	if err := global.LRAG_DB.Preload("Authorities").Preload("Authority").First(&full, user.ID).Error; err != nil {
		return "", "user", err
	}
	syssvc.MenuServiceApp.UserAuthorityDefaultRouter(&full)

	loginResp, failKey := syssvc.UserServiceApp.FinishLoginSession(c, &full)
	if failKey != "" {
		return "", "session", fmt.Errorf("%s", failKey)
	}

	exID, err := randomURLState()
	if err != nil {
		return "", "server", err
	}
	exBody := oauthExchangePayload{Login: loginResp, Redirect: st.ReturnPath}
	exBytes, _ := json.Marshal(exBody)
	global.BlackCache.Set("oauth:ex:"+exID, string(exBytes), oauthExchangeTTL)

	fe := EffectiveOAuthFrontendRedirect()
	sep := "?"
	if strings.Contains(fe, "?") {
		sep = "&"
	}
	return fe + sep + "oauth_ex=" + exID, "", nil
}

func (s *OAuthFlowService) ensureOAuthUser(c *gin.Context, kind string, prof *idpoauth.NormalizedProfile, prov sysmodel.SysOAuthProvider) (sysmodel.SysUser, error) {
	var bind sysmodel.SysUserOAuthBinding
	err := global.LRAG_DB.Where("LOWER(provider_kind) = ? AND subject = ?", strings.ToLower(strings.TrimSpace(kind)), prof.Subject).First(&bind).Error
	if err == nil {
		var u sysmodel.SysUser
		if err := global.LRAG_DB.First(&u, bind.UserID).Error; err != nil {
			return u, err
		}
		return u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return sysmodel.SysUser{}, err
	}

	username := oauthUsernameCandidate(kind, prof.Subject)
	for i := 0; i < 8; i++ {
		var cnt int64
		global.LRAG_DB.Model(&sysmodel.SysUser{}).Where("username = ?", username).Count(&cnt)
		if cnt == 0 {
			break
		}
		username = oauthUsernameCandidate(kind, prof.Subject) + "_" + randomSuffix4()
	}

	aid := prov.DefaultAuthorityID
	if aid == 0 {
		aid = 888
	}
	u := sysmodel.SysUser{
		UUID:        uuid.New(),
		Username:    username,
		Password:    utils.BcryptHash(uuid.New().String()),
		NickName:    firstNonEmpty(prof.Name, username),
		HeaderImg:   prof.AvatarURL,
		Email:       prof.Email,
		AuthorityId: aid,
		Authorities: []sysmodel.SysAuthority{{AuthorityId: aid}},
		Enable:      1,
	}

	err = global.LRAG_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		b := sysmodel.SysUserOAuthBinding{UserID: u.ID, ProviderKind: strings.ToLower(strings.TrimSpace(kind)), Subject: prof.Subject}
		return tx.Create(&b).Error
	})
	return u, err
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return strings.TrimSpace(a)
	}
	return b
}

func oauthUsernameCandidate(kind, subject string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	sb := strings.Builder{}
	sb.WriteString("oauth_")
	sb.WriteString(k)
	sb.WriteString("_")
	for _, r := range subject {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else if r == '-' || r == '_' {
			sb.WriteRune('_')
		}
	}
	out := sb.String()
	if len(out) > 60 {
		out = out[:60]
	}
	if len(out) < len("oauth_x_")+1 {
		out = "oauth_" + k + "_user"
	}
	return out
}

func randomSuffix4() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 4)
	_, _ = crand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// PopOAuthExchange 一次性取出换票数据
func (s *OAuthFlowService) PopOAuthExchange(exID string) (oauthExchangePayload, bool) {
	key := "oauth:ex:" + exID
	raw, ok := global.BlackCache.Get(key)
	if !ok {
		return oauthExchangePayload{}, false
	}
	global.BlackCache.Delete(key)
	str, _ := raw.(string)
	var p oauthExchangePayload
	if json.Unmarshal([]byte(str), &p) != nil {
		return oauthExchangePayload{}, false
	}
	return p, true
}

// OAuthExchangeForJSON 供 API 返回（包装 Redirect）
func (s *OAuthFlowService) OAuthExchangeForJSON(exID string) (systemRes.OAuthExchangeData, bool) {
	p, ok := s.PopOAuthExchange(exID)
	if !ok {
		return systemRes.OAuthExchangeData{}, false
	}
	return systemRes.OAuthExchangeData{
		LoginResponse: p.Login,
		Redirect:      p.Redirect,
	}, true
}
