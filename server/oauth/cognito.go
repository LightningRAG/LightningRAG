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

type cognitoProvider struct{}

func init() {
	Register(&cognitoProvider{})
}

// cognitoDomain 托管域主机名，如 your-prefix.auth.us-east-1.amazoncognito.com（勿含 https://）
func cognitoDomain(extra map[string]any) string {
	if extra == nil {
		return ""
	}
	h, _ := extra["cognito_domain"].(string)
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	return strings.TrimSuffix(h, "/")
}

func (cognitoProvider) Kind() string               { return "cognito" }
func (cognitoProvider) DefaultDisplayName() string { return "Amazon Cognito" }
func (cognitoProvider) DefaultScopes() []string {
	return []string{"openid", "email", "profile"}
}

func (cognitoProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	host := cognitoDomain(rc.Extra)
	if host == "" {
		host = "configure-cognito-domain.invalid"
	}
	base := "https://" + host
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  base + "/oauth2/authorize",
			TokenURL: base + "/oauth2/token",
		},
	}
}

func (cognitoProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	host := cognitoDomain(rc.Extra)
	if host == "" {
		return nil, fmt.Errorf("cognito: missing extra.cognito_domain")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/oauth2/userInfo", nil)
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
		return nil, fmt.Errorf("cognito userInfo: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		Sub         string `json:"sub"`
		Email       string `json:"email"`
		Username    string `json:"username"`
		Name        string `json:"name"`
		Picture     string `json:"picture"`
		GivenName   string `json:"given_name"`
		FamilyName  string `json:"family_name"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	sub := strings.TrimSpace(u.Sub)
	if sub == "" {
		sub = strings.TrimSpace(u.Username)
	}
	if sub == "" {
		return nil, fmt.Errorf("cognito: empty subject")
	}
	name := strings.TrimSpace(u.Name)
	if name == "" {
		name = strings.TrimSpace(strings.TrimSpace(u.GivenName) + " " + strings.TrimSpace(u.FamilyName))
	}
	if name == "" {
		name = strings.TrimSpace(u.Email)
	}
	if name == "" {
		name = sub
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     strings.TrimSpace(u.Email),
		Name:      name,
		AvatarURL: strings.TrimSpace(u.Picture),
	}, nil
}
