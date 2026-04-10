package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

type twitterProvider struct{}

func init() {
	Register(&twitterProvider{})
}

func (twitterProvider) Kind() string               { return "twitter" }
func (twitterProvider) DefaultDisplayName() string { return "X (Twitter)" }
func (twitterProvider) DefaultScopes() []string {
	// users.read：基本资料；users.email：邮箱（需开发者应用具备该权限且用户已验证邮箱）
	return []string{"users.read", "users.email", "tweet.read"}
}

func (twitterProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
	}
}

func (twitterProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	q := url.Values{}
	// confirmed_email：用户已在 X 侧确认的邮箱（需 users.email scope，且开发者应用允许请求邮箱）
	q.Set("user.fields", "id,name,username,profile_image_url,confirmed_email")
	u := "https://api.twitter.com/2/users/me?" + q.Encode()
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
		return nil, fmt.Errorf("twitter users/me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var wrap struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ProfileImageURL string `json:"profile_image_url"`
			ConfirmedEmail  string `json:"confirmed_email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	d := wrap.Data
	if d.ID == "" {
		return nil, fmt.Errorf("twitter: empty user id")
	}
	name := strings.TrimSpace(d.Name)
	if name == "" {
		name = d.Username
	}
	if name == "" {
		name = d.ID
	}
	return &NormalizedProfile{
		Subject:   d.ID,
		Email:     strings.TrimSpace(d.ConfirmedEmail),
		Name:      name,
		AvatarURL: strings.TrimSpace(d.ProfileImageURL),
	}, nil
}
