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

type figmaProvider struct{}

func init() {
	Register(&figmaProvider{})
}

func (figmaProvider) Kind() string               { return "figma" }
func (figmaProvider) DefaultDisplayName() string { return "Figma" }
func (figmaProvider) DefaultScopes() []string {
	return []string{"file_read"}
}

func (figmaProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.figma.com/oauth",
			TokenURL: "https://api.figma.com/v1/oauth/token",
		},
	}
}

func (figmaProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.figma.com/v1/me", nil)
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
		return nil, fmt.Errorf("figma me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		ID     string `json:"id"`
		Email  string `json:"email"`
		Handle string `json:"handle"`
		ImgURL string `json:"img_url"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	sub := strings.TrimSpace(u.ID)
	if sub == "" {
		sub = strings.TrimSpace(u.Handle)
	}
	if sub == "" {
		return nil, fmt.Errorf("figma: empty subject")
	}
	name := strings.TrimSpace(u.Handle)
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
		AvatarURL: strings.TrimSpace(u.ImgURL),
	}, nil
}
