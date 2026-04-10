package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const feishuDefaultAPIBase = "https://open.feishu.cn"

const (
	feishuTenantTokenPath = "/open-apis/auth/v3/tenant_access_token/internal"
	feishuMessagesPath    = "/open-apis/im/v1/messages"
)

type feishuTokenResp struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

type feishuMsgResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var feishuTokenMu sync.Mutex
var feishuTokenCache = map[string]feishuCachedToken{} // key: apiBase + "\x00" + appID

type feishuCachedToken struct {
	token     string
	expiresAt time.Time
}

func feishuExtraString(extra map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := extra[k]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

// FeishuAPIBaseFromExtra 开放平台根 URL，默认国内 https://open.feishu.cn；国际 Lark 填 https://open.larksuite.com（或 feishu_api_base / lark_api_base）
func FeishuAPIBaseFromExtra(extra map[string]any) string {
	s := feishuExtraString(extra, "feishu_api_base", "lark_api_base")
	s = strings.TrimSpace(strings.TrimRight(s, "/"))
	if s == "" {
		return feishuDefaultAPIBase
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	return s
}

func feishuGetTenantAccessToken(ctx context.Context, appID, appSecret, apiBase string) (string, error) {
	if appID == "" || appSecret == "" {
		return "", errors.New("feishu: extra 需配置 app_id 与 app_secret")
	}
	cacheKey := apiBase + "\x00" + appID
	now := time.Now()
	feishuTokenMu.Lock()
	if ent, ok := feishuTokenCache[cacheKey]; ok && ent.expiresAt.After(now.Add(2*time.Minute)) {
		tok := ent.token
		feishuTokenMu.Unlock()
		return tok, nil
	}
	feishuTokenMu.Unlock()

	body, err := json.Marshal(map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	})
	if err != nil {
		return "", err
	}
	tokenURL := apiBase + feishuTenantTokenPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var tr feishuTokenResp
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", fmt.Errorf("feishu token: parse: %w", err)
	}
	if tr.Code != 0 {
		return "", fmt.Errorf("feishu token: %s (%d)", tr.Msg, tr.Code)
	}
	if tr.TenantAccessToken == "" {
		return "", errors.New("feishu token: empty")
	}
	expSec := tr.Expire
	if expSec < 60 {
		expSec = 7000
	}
	feishuTokenMu.Lock()
	feishuTokenCache[cacheKey] = feishuCachedToken{
		token:     tr.TenantAccessToken,
		expiresAt: now.Add(time.Duration(expSec) * time.Second),
	}
	feishuTokenMu.Unlock()
	return tr.TenantAccessToken, nil
}

func feishuSendChatText(ctx context.Context, tenantToken, apiBase, chatID, text string) error {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return errors.New("feishu: empty chat_id")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if len(text) > 20000 {
		text = text[:20000] + "…"
	}
	inner, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{
		"receive_id": chatID,
		"msg_type":   "text",
		"content":    string(inner),
	})
	if err != nil {
		return err
	}
	u := apiBase + feishuMessagesPath + "?receive_id_type=chat_id"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tenantToken)

	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var mr feishuMsgResp
	_ = json.Unmarshal(raw, &mr)
	if mr.Code != 0 {
		return fmt.Errorf("feishu send message: %s (%d) body=%s", mr.Msg, mr.Code, string(raw))
	}
	return nil
}

func (feishuAdapter) SendReply(ctx context.Context, _ string, cfg *ConnectorConfig, thread ThreadRef, text string) error {
	if cfg == nil || cfg.Extra == nil {
		return errors.New("feishu SendReply: no config")
	}
	appID := feishuExtraString(cfg.Extra, "app_id")
	secret := feishuExtraString(cfg.Extra, "app_secret")
	var chatID string
	if thread.Opaque != nil {
		if v, ok := thread.Opaque["chat_id"].(string); ok {
			chatID = v
		}
	}
	if chatID == "" {
		return errors.New("feishu SendReply: missing chat_id in ReplyRef")
	}
	apiBase := FeishuAPIBaseFromExtra(cfg.Extra)
	tok, err := feishuGetTenantAccessToken(ctx, appID, secret, apiBase)
	if err != nil {
		return err
	}
	return feishuSendChatText(ctx, tok, apiBase, chatID, text)
}
