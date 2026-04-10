package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const wecomGetTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
const wecomMessageSendURL = "https://qyapi.weixin.qq.com/cgi-bin/message/send"

type wecomTokenCacheEntry struct {
	token string
	exp   time.Time
}

var (
	wecomTokenMu   sync.Mutex
	wecomTokenByID map[string]wecomTokenCacheEntry // key: corpID + "\x00" + corpSecret
)

func init() {
	wecomTokenByID = make(map[string]wecomTokenCacheEntry)
}

func wecomAccessToken(ctx context.Context, corpID, corpSecret string) (string, error) {
	corpID, corpSecret = strings.TrimSpace(corpID), strings.TrimSpace(corpSecret)
	if corpID == "" || corpSecret == "" {
		return "", fmt.Errorf("wecom: empty corpid/secret")
	}
	key := corpID + "\x00" + corpSecret
	now := time.Now()
	wecomTokenMu.Lock()
	if e, ok := wecomTokenByID[key]; ok && e.exp.After(now.Add(30*time.Second)) {
		tok := e.token
		wecomTokenMu.Unlock()
		return tok, nil
	}
	wecomTokenMu.Unlock()

	u, _ := url.Parse(wecomGetTokenURL)
	q := u.Query()
	q.Set("corpid", corpID)
	q.Set("corpsecret", corpSecret)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tokResp struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokResp); err != nil {
		return "", fmt.Errorf("wecom gettoken json: %w", err)
	}
	if tokResp.ErrCode != 0 || tokResp.AccessToken == "" {
		return "", fmt.Errorf("wecom gettoken: %d %s", tokResp.ErrCode, tokResp.ErrMsg)
	}
	ttl := tokResp.ExpiresIn
	if ttl < 60 {
		ttl = 120
	}
	if ttl > 7000 {
		ttl = 7000
	}
	wecomTokenMu.Lock()
	wecomTokenByID[key] = wecomTokenCacheEntry{token: tokResp.AccessToken, exp: now.Add(time.Duration(ttl) * time.Second)}
	wecomTokenMu.Unlock()
	return tokResp.AccessToken, nil
}

func wecomSendText(ctx context.Context, corpID, corpSecret string, agentID int, toUser, text string) error {
	toUser = strings.TrimSpace(toUser)
	text = strings.TrimSpace(text)
	if toUser == "" || text == "" {
		return nil
	}
	if len(text) > 2048 {
		text = text[:2042] + "…"
	}
	tok, err := wecomAccessToken(ctx, corpID, corpSecret)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"touser":  toUser,
		"msgtype": "text",
		"agentid": agentID,
		"text":    map[string]string{"content": text},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u, _ := url.Parse(wecomMessageSendURL)
	q := u.Query()
	q.Set("access_token", tok)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	var sendResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.Unmarshal(out, &sendResp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wecom message/send http: %s %s", resp.Status, string(out))
	}
	if sendResp.ErrCode != 0 {
		return fmt.Errorf("wecom message/send: %d %s", sendResp.ErrCode, sendResp.ErrMsg)
	}
	return nil
}
