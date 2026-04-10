package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

func paypalSandbox(extra map[string]any) bool {
	if extra == nil {
		return false
	}
	if b, ok := extra["paypal_sandbox"].(bool); ok && b {
		return true
	}
	if s, ok := extra["paypal_sandbox"].(string); ok && strings.EqualFold(strings.TrimSpace(s), "true") {
		return true
	}
	if f, ok := extra["paypal_sandbox"].(float64); ok && f != 0 {
		return true
	}
	return false
}

func paypalEndpoints(extra map[string]any) (authURL, tokenURL, apiBase string) {
	if paypalSandbox(extra) {
		return "https://www.sandbox.paypal.com/signin/authorize",
			"https://api.sandbox.paypal.com/v1/oauth2/token",
			"https://api.sandbox.paypal.com"
	}
	return "https://www.paypal.com/signin/authorize",
		"https://api.paypal.com/v1/oauth2/token",
		"https://api.paypal.com"
}

type paypalProvider struct{}

func init() {
	Register(&paypalProvider{})
}

func (paypalProvider) Kind() string               { return "paypal" }
func (paypalProvider) DefaultDisplayName() string { return "PayPal" }
func (paypalProvider) DefaultScopes() []string {
	return []string{"openid", "email", "profile"}
}

func (paypalProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	authURL, tokenURL, _ := paypalEndpoints(rc.Extra)
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
}

func (paypalProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	_, _, apiBase := paypalEndpoints(rc.Extra)
	u := apiBase + "/v1/identity/openidconnect/userinfo?schema=openid"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("paypal userinfo: HTTP %d %s", resp.StatusCode, string(body))
	}
	var raw struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Payer struct {
			PayerID string `json:"payer_id"`
		} `json:"payer"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	sub := strings.TrimSpace(raw.Sub)
	if sub == "" {
		sub = strings.TrimSpace(raw.Payer.PayerID)
	}
	if sub == "" {
		return nil, fmt.Errorf("paypal: empty subject")
	}
	name := strings.TrimSpace(raw.Name)
	if name == "" {
		name = sub
	}
	return &NormalizedProfile{
		Subject: sub,
		Email:   strings.TrimSpace(raw.Email),
		Name:    name,
	}, nil
}
