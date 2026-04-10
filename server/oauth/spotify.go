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

type spotifyProvider struct{}

func init() {
	Register(&spotifyProvider{})
}

func (spotifyProvider) Kind() string               { return "spotify" }
func (spotifyProvider) DefaultDisplayName() string { return "Spotify" }
func (spotifyProvider) DefaultScopes() []string {
	return []string{"user-read-email", "user-read-private"}
}

func (spotifyProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}
}

func (spotifyProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me", nil)
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
		return nil, fmt.Errorf("spotify me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		ID           string `json:"id"`
		DisplayName  string `json:"display_name"`
		Email        string `json:"email"`
		Images       []struct {
			URL string `json:"url"`
		} `json:"images"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	avatar := ""
	if len(u.Images) > 0 {
		avatar = strings.TrimSpace(u.Images[0].URL)
	}
	name := strings.TrimSpace(u.DisplayName)
	if name == "" {
		name = u.ID
	}
	return &NormalizedProfile{
		Subject:   u.ID,
		Email:     strings.TrimSpace(u.Email),
		Name:      name,
		AvatarURL: avatar,
	}, nil
}
