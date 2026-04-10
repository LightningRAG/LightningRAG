package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
)

type gitlabProvider struct{}

func init() {
	Register(&gitlabProvider{})
}

// extra.gitlab_base_url 自建实例根地址，默认 https://gitlab.com（勿尾斜杠）
func gitlabRoot(extra map[string]any) string {
	const def = "https://gitlab.com"
	if extra == nil {
		return def
	}
	v, _ := extra["gitlab_base_url"].(string)
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	return strings.TrimSuffix(v, "/")
}

func (gitlabProvider) Kind() string               { return "gitlab" }
func (gitlabProvider) DefaultDisplayName() string { return "GitLab" }
func (gitlabProvider) DefaultScopes() []string {
	// read_user 即可访问 /api/v4/user；自建实例可在后台追加 openid 等 Scope
	return []string{"read_user"}
}

func (gitlabProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	root := gitlabRoot(rc.Extra)
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  root + "/oauth/authorize",
			TokenURL: root + "/oauth/token",
		},
	}
}

func (gitlabProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	root := gitlabRoot(rc.Extra)
	u, err := url.Parse(root + "/api/v4/user")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
		return nil, fmt.Errorf("gitlab user: HTTP %d %s", resp.StatusCode, string(body))
	}
	var raw struct {
		ID             int    `json:"id"`
		Username       string `json:"username"`
		Name           string `json:"name"`
		AvatarURL      string `json:"avatar_url"`
		Email          string `json:"email"`
		PublicEmail    string `json:"public_email"`
		CommitEmail    string `json:"commit_email"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	sub := strconv.Itoa(raw.ID)
	name := strings.TrimSpace(raw.Name)
	if name == "" {
		name = raw.Username
	}
	email := strings.TrimSpace(raw.Email)
	if email == "" {
		email = strings.TrimSpace(raw.PublicEmail)
	}
	if email == "" {
		email = strings.TrimSpace(raw.CommitEmail)
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     email,
		Name:      name,
		AvatarURL: raw.AvatarURL,
	}, nil
}
