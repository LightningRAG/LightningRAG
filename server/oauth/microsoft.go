package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type microsoftProvider struct{}

func init() {
	Register(&microsoftProvider{})
}

func (microsoftProvider) Kind() string               { return "microsoft" }
func (microsoftProvider) DefaultDisplayName() string { return "Microsoft" }
func (microsoftProvider) DefaultScopes() []string {
	return []string{"openid", "profile", "email", "User.Read"}
}

func tenantFromExtra(extra map[string]any) string {
	if extra == nil {
		return "common"
	}
	v, _ := extra["tenant"].(string)
	v = strings.TrimSpace(v)
	if v == "" {
		return "common"
	}
	return v
}

func (microsoftProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	tenant := tenantFromExtra(rc.Extra)
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint:     microsoft.AzureADEndpoint(tenant),
	}
}

func (microsoftProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
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
		return nil, fmt.Errorf("microsoft graph me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		ID                string `json:"id"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	email := u.Mail
	if email == "" {
		email = u.UserPrincipalName
	}
	return &NormalizedProfile{
		Subject: u.ID,
		Email:   email,
		Name:    u.DisplayName,
	}, nil
}
