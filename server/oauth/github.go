package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
)

type githubProvider struct{}

func init() {
	Register(&githubProvider{})
}

// githubURLs 默认 github.com；GitHub Enterprise 在 extra 中配置 github_enterprise_host（仅主机名，如 github.company.com）
func githubURLs(extra map[string]any) (authURL, tokenURL, apiBase string) {
	const (
		defAuth  = "https://github.com/login/oauth/authorize"
		defToken = "https://github.com/login/oauth/access_token"
		defAPI   = "https://api.github.com"
	)
	if extra == nil {
		return defAuth, defToken, defAPI
	}
	h, _ := extra["github_enterprise_host"].(string)
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	h = strings.TrimSuffix(h, "/")
	if h == "" {
		return defAuth, defToken, defAPI
	}
	return "https://" + h + "/login/oauth/authorize",
		"https://" + h + "/login/oauth/access_token",
		"https://" + h + "/api/v3"
}

func (githubProvider) Kind() string               { return "github" }
func (githubProvider) DefaultDisplayName() string { return "GitHub" }
func (githubProvider) DefaultScopes() []string {
	return []string{"read:user", "user:email"}
}

func (githubProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	authURL, tokenURL, _ := githubURLs(rc.Extra)
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
}

func (githubProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	_, _, apiBase := githubURLs(rc.Extra)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
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
		return nil, fmt.Errorf("github user: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		ID        float64 `json:"id"`
		Login     string  `json:"login"`
		Name      string  `json:"name"`
		AvatarURL string  `json:"avatar_url"`
		Email     string  `json:"email"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	sub := strconv.FormatInt(int64(u.ID), 10)
	email := u.Email
	if email == "" {
		email, _ = githubPrimaryEmail(ctx, tok.AccessToken, apiBase)
	}
	name := u.Name
	if name == "" {
		name = u.Login
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     email,
		Name:      name,
		AvatarURL: u.AvatarURL,
	}, nil
}

func githubPrimaryEmail(ctx context.Context, accessToken, apiBase string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if json.Unmarshal(body, &emails) != nil {
		return "", nil
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}
