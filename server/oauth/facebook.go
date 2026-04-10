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

// facebookGraphAPIVersion Graph API 版本，可在 extra.facebook_graph_version 覆盖（如 "v21.0"）
func facebookGraphVersion(extra map[string]any) string {
	const def = "v20.0"
	if extra == nil {
		return def
	}
	v, _ := extra["facebook_graph_version"].(string)
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

type facebookProvider struct{}

func init() {
	Register(&facebookProvider{})
}

func (facebookProvider) Kind() string               { return "facebook" }
func (facebookProvider) DefaultDisplayName() string { return "Facebook" }
func (facebookProvider) DefaultScopes() []string {
	return []string{"email", "public_profile"}
}

func (facebookProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	ver := facebookGraphVersion(rc.Extra)
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/" + ver + "/dialog/oauth",
			TokenURL: "https://graph.facebook.com/" + ver + "/oauth/access_token",
		},
	}
}

func (facebookProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	ver := facebookGraphVersion(rc.Extra)
	q := url.Values{}
	q.Set("fields", "id,name,email,picture.type(large)")
	q.Set("access_token", tok.AccessToken)
	u := "https://graph.facebook.com/" + ver + "/me?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("facebook me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var raw struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return &NormalizedProfile{
		Subject:   raw.ID,
		Email:     strings.TrimSpace(raw.Email),
		Name:      strings.TrimSpace(raw.Name),
		AvatarURL: strings.TrimSpace(raw.Picture.Data.URL),
	}, nil
}
