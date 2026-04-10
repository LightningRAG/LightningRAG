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

type stravaProvider struct{}

func init() {
	Register(&stravaProvider{})
}

func (stravaProvider) Kind() string               { return "strava" }
func (stravaProvider) DefaultDisplayName() string { return "Strava" }
func (stravaProvider) DefaultScopes() []string {
	return []string{"read"}
}

func (stravaProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.strava.com/oauth/authorize",
			TokenURL: "https://www.strava.com/oauth/token",
		},
	}
}

func (stravaProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.strava.com/api/v3/athlete", nil)
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
		return nil, fmt.Errorf("strava athlete: HTTP %d %s", resp.StatusCode, string(body))
	}
	var a struct {
		ID        int64  `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		Profile   string `json:"profile"`
		Email     string `json:"email"`
	}
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, err
	}
	if a.ID == 0 {
		return nil, fmt.Errorf("strava: empty athlete id")
	}
	sub := fmt.Sprintf("%d", a.ID)
	name := strings.TrimSpace(strings.TrimSpace(a.FirstName) + " " + strings.TrimSpace(a.LastName))
	if name == "" {
		name = sub
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     strings.TrimSpace(a.Email),
		Name:      name,
		AvatarURL: strings.TrimSpace(a.Profile),
	}, nil
}
