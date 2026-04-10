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

type linkedinProvider struct{}

func init() {
	Register(&linkedinProvider{})
}

func (linkedinProvider) Kind() string               { return "linkedin" }
func (linkedinProvider) DefaultDisplayName() string { return "LinkedIn" }
func (linkedinProvider) DefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}

func (linkedinProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
		},
	}
}

func (linkedinProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.linkedin.com/v2/userinfo", nil)
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
		return nil, fmt.Errorf("linkedin userinfo: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	return &NormalizedProfile{
		Subject:   strings.TrimSpace(u.Sub),
		Email:     strings.TrimSpace(u.Email),
		Name:      strings.TrimSpace(u.Name),
		AvatarURL: strings.TrimSpace(u.Picture),
	}, nil
}
