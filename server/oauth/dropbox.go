package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type dropboxProvider struct{}

func init() {
	Register(&dropboxProvider{})
}

func (dropboxProvider) Kind() string               { return "dropbox" }
func (dropboxProvider) DefaultDisplayName() string { return "Dropbox" }
func (dropboxProvider) DefaultScopes() []string {
	return []string{"account_info.read"}
}

func (dropboxProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/oauth2/authorize",
			TokenURL: "https://api.dropboxapi.com/oauth2/token",
		},
	}
}

func (dropboxProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/users/get_current_account", bytes.NewBufferString("null"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("dropbox get_current_account: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		AccountID string `json:"account_id"`
		Email     string `json:"email"`
		Name      struct {
			DisplayName string `json:"display_name"`
		} `json:"name"`
		ProfilePhotoURL string `json:"profile_photo_url"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	sub := strings.TrimSpace(u.AccountID)
	if sub == "" {
		return nil, fmt.Errorf("dropbox: empty account_id")
	}
	name := strings.TrimSpace(u.Name.DisplayName)
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
		AvatarURL: strings.TrimSpace(u.ProfilePhotoURL),
	}, nil
}
