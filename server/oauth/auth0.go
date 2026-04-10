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

type auth0Provider struct{}

func init() {
	Register(&auth0Provider{})
}

// auth0Domain 必填，如 tenant.eu.auth0.com（勿含 https://）
func auth0Domain(extra map[string]any) string {
	if extra == nil {
		return ""
	}
	d, _ := extra["auth0_domain"].(string)
	d = strings.TrimSpace(d)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	return strings.TrimSuffix(d, "/")
}

func (auth0Provider) Kind() string               { return "auth0" }
func (auth0Provider) DefaultDisplayName() string { return "Auth0" }
func (auth0Provider) DefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}

func (auth0Provider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	d := auth0Domain(rc.Extra)
	if d == "" {
		d = "configure-auth0-domain.invalid"
	}
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + d + "/authorize",
			TokenURL: "https://" + d + "/oauth/token",
		},
	}
}

func (auth0Provider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	d := auth0Domain(rc.Extra)
	if d == "" {
		return nil, fmt.Errorf("auth0: missing extra.auth0_domain")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+d+"/userinfo", nil)
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
		return nil, fmt.Errorf("auth0 userinfo: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	if strings.TrimSpace(u.Sub) == "" {
		return nil, fmt.Errorf("auth0: empty sub")
	}
	return &NormalizedProfile{
		Subject:   strings.TrimSpace(u.Sub),
		Email:     strings.TrimSpace(u.Email),
		Name:      strings.TrimSpace(u.Name),
		AvatarURL: strings.TrimSpace(u.Picture),
	}, nil
}
