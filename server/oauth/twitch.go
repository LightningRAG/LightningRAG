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

type twitchProvider struct{}

func init() {
	Register(&twitchProvider{})
}

func (twitchProvider) Kind() string               { return "twitch" }
func (twitchProvider) DefaultDisplayName() string { return "Twitch" }
func (twitchProvider) DefaultScopes() []string {
	return []string{"user:read:email"}
}

func (twitchProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://id.twitch.tv/oauth2/authorize",
			TokenURL: "https://id.twitch.tv/oauth2/token",
		},
	}
}

func (twitchProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Client-Id", strings.TrimSpace(rc.ClientID))
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
		return nil, fmt.Errorf("twitch helix users: HTTP %d %s", resp.StatusCode, string(body))
	}
	var wrap struct {
		Data []struct {
			ID              string `json:"id"`
			Login           string `json:"login"`
			DisplayName     string `json:"display_name"`
			Email           string `json:"email"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	if len(wrap.Data) == 0 {
		return nil, fmt.Errorf("twitch: empty user data")
	}
	u := wrap.Data[0]
	name := strings.TrimSpace(u.DisplayName)
	if name == "" {
		name = u.Login
	}
	return &NormalizedProfile{
		Subject:   u.ID,
		Email:     strings.TrimSpace(u.Email),
		Name:      name,
		AvatarURL: strings.TrimSpace(u.ProfileImageURL),
	}, nil
}
