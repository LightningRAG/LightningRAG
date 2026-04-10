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

type bitbucketProvider struct{}

func init() {
	Register(&bitbucketProvider{})
}

func (bitbucketProvider) Kind() string               { return "bitbucket" }
func (bitbucketProvider) DefaultDisplayName() string { return "Bitbucket" }
func (bitbucketProvider) DefaultScopes() []string {
	return []string{"account", "email"}
}

func (bitbucketProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://bitbucket.org/site/oauth2/authorize",
			TokenURL: "https://bitbucket.org/site/oauth2/access_token",
		},
	}
}

func (bitbucketProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.bitbucket.org/2.0/user", nil)
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
		return nil, fmt.Errorf("bitbucket user: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		UUID        string `json:"uuid"`
		DisplayName string `json:"display_name"`
		AccountID   string `json:"account_id"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
		} `json:"links"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	sub := strings.TrimSpace(u.AccountID)
	if sub == "" {
		sub = strings.Trim(u.UUID, "{}")
	}
	if sub == "" {
		return nil, fmt.Errorf("bitbucket: empty subject")
	}
	email, _ := bitbucketPrimaryEmail(ctx, tok.AccessToken)
	name := strings.TrimSpace(u.DisplayName)
	if name == "" {
		name = sub
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     email,
		Name:      name,
		AvatarURL: strings.TrimSpace(u.Links.Avatar.Href),
	}, nil
}

func bitbucketPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.bitbucket.org/2.0/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil
	}
	var wrap struct {
		Values []struct {
			Email     string `json:"email"`
			IsPrimary bool   `json:"is_primary"`
			IsConfirmed bool `json:"is_confirmed"`
		} `json:"values"`
	}
	if json.Unmarshal(body, &wrap) != nil {
		return "", nil
	}
	for _, e := range wrap.Values {
		if e.IsPrimary && e.IsConfirmed {
			return e.Email, nil
		}
	}
	for _, e := range wrap.Values {
		if e.IsConfirmed {
			return e.Email, nil
		}
	}
	for _, e := range wrap.Values {
		if e.Email != "" {
			return e.Email, nil
		}
	}
	return "", nil
}
