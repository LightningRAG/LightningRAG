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

type zoomProvider struct{}

func init() {
	Register(&zoomProvider{})
}

func (zoomProvider) Kind() string               { return "zoom" }
func (zoomProvider) DefaultDisplayName() string { return "Zoom" }
func (zoomProvider) DefaultScopes() []string {
	return []string{"user:read:user"}
}

func (zoomProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://zoom.us/oauth/authorize",
			TokenURL: "https://zoom.us/oauth/token",
		},
	}
}

func (zoomProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.zoom.us/v2/users/me", nil)
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
		return nil, fmt.Errorf("zoom users/me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		PicURL    string `json:"pic_url"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(strings.TrimSpace(u.FirstName + " " + u.LastName))
	if name == "" {
		name = strings.TrimSpace(u.Email)
	}
	if name == "" {
		name = u.ID
	}
	sub := strings.TrimSpace(u.ID)
	if sub == "" {
		sub = strings.TrimSpace(u.Email)
	}
	if sub == "" {
		return nil, fmt.Errorf("zoom: empty subject")
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     strings.TrimSpace(u.Email),
		Name:      name,
		AvatarURL: strings.TrimSpace(u.PicURL),
	}, nil
}
