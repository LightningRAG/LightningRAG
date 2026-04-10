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

type discordProvider struct{}

func init() {
	Register(&discordProvider{})
}

func (discordProvider) Kind() string               { return "discord" }
func (discordProvider) DefaultDisplayName() string { return "Discord" }
func (discordProvider) DefaultScopes() []string {
	return []string{"identify", "email"}
}

func (discordProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}

func discordAvatarURL(userID, avatarHash string) string {
	avatarHash = strings.TrimSpace(avatarHash)
	if avatarHash == "" || userID == "" {
		return ""
	}
	// 动画头像是 a_ 前缀的 gif
	ext := "png"
	if strings.HasPrefix(avatarHash, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s", userID, avatarHash, ext)
}

func (discordProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/users/@me", nil)
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
		return nil, fmt.Errorf("discord @me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Discriminator string `json:"discriminator"`
		GlobalName    string `json:"global_name"`
		Avatar        string `json:"avatar"`
		Email         string `json:"email"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(u.GlobalName)
	if name == "" {
		name = u.Username
		if u.Discriminator != "" && u.Discriminator != "0" {
			name = u.Username + "#" + u.Discriminator
		}
	}
	return &NormalizedProfile{
		Subject:   u.ID,
		Email:     strings.TrimSpace(u.Email),
		Name:      name,
		AvatarURL: discordAvatarURL(u.ID, u.Avatar),
	}, nil
}
