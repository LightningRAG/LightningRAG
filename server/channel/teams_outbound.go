package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const teamsTokenURL = "https://login.microsoftonline.com/botframework/oauth2/v2.0/token"

type teamsTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type teamsTokenCacheEnt struct {
	token     string
	expiresAt time.Time
}

var teamsTokMu sync.Mutex
var teamsTokCache = map[string]teamsTokenCacheEnt{} // key: appID

func teamsGetAccessToken(ctx context.Context, appID, appPassword string) (string, error) {
	appID = strings.TrimSpace(appID)
	appPassword = strings.TrimSpace(appPassword)
	if appID == "" || appPassword == "" {
		return "", errors.New("teams: extra 需配置 teams_microsoft_app_id 与 teams_microsoft_app_password")
	}
	now := time.Now()
	teamsTokMu.Lock()
	if ent, ok := teamsTokCache[appID]; ok && ent.expiresAt.After(now.Add(2*time.Minute)) {
		t := ent.token
		teamsTokMu.Unlock()
		return t, nil
	}
	teamsTokMu.Unlock()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", appID)
	form.Set("client_secret", appPassword)
	form.Set("scope", "https://api.botframework.com/.default")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, teamsTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("teams oauth: %s %s", resp.Status, string(raw))
	}
	var tr teamsTokenResp
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", fmt.Errorf("teams oauth: parse: %w", err)
	}
	if strings.TrimSpace(tr.AccessToken) == "" {
		return "", errors.New("teams oauth: empty access_token")
	}
	expSec := tr.ExpiresIn
	if expSec < 60 {
		expSec = 3500
	}
	teamsTokMu.Lock()
	teamsTokCache[appID] = teamsTokenCacheEnt{
		token:     tr.AccessToken,
		expiresAt: now.Add(time.Duration(expSec) * time.Second),
	}
	teamsTokMu.Unlock()
	return tr.AccessToken, nil
}

func teamsSendMessage(ctx context.Context, bearer, serviceURL, conversationID, fromBotID, text string) error {
	serviceURL = strings.TrimSpace(serviceURL)
	conversationID = strings.TrimSpace(conversationID)
	fromBotID = strings.TrimSpace(fromBotID)
	if serviceURL == "" || conversationID == "" || fromBotID == "" {
		return errors.New("teams SendReply: missing serviceUrl / conversation / bot id")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	const maxRunes = 27000
	if len(text) > maxRunes {
		text = text[:maxRunes] + "…"
	}
	base := strings.TrimRight(serviceURL, "/")
	u := base + "/v3/conversations/" + url.PathEscape(conversationID) + "/activities"
	body := map[string]any{
		"type": "message",
		"text": text,
		"from": map[string]string{"id": fromBotID},
		"conversation": map[string]string{
			"id": conversationID,
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams send: %s %s", resp.Status, string(raw))
	}
	return nil
}
