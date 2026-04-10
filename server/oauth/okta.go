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

type oktaProvider struct{}

func init() {
	Register(&oktaProvider{})
}

// oktaIssuer 优先 extra.okta_issuer（完整 Issuer URL，如 https://dev-xxx.okta.com/oauth2/default）；
// 否则用 okta_domain + okta_auth_server（默认 default）：https://{domain}/oauth2/{server}
func oktaIssuer(extra map[string]any) string {
	if extra == nil {
		return ""
	}
	if v, _ := extra["okta_issuer"].(string); strings.TrimSpace(v) != "" {
		s := strings.TrimSpace(v)
		s = strings.TrimSuffix(s, "/")
		if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
			s = "https://" + s
		}
		return s
	}
	d, _ := extra["okta_domain"].(string)
	d = strings.TrimSpace(d)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimSuffix(d, "/")
	if d == "" {
		return ""
	}
	srv := "default"
	if s, ok := extra["okta_auth_server"].(string); ok && strings.TrimSpace(s) != "" {
		srv = strings.TrimSpace(s)
	}
	return "https://" + d + "/oauth2/" + srv
}

func (oktaProvider) Kind() string               { return "okta" }
func (oktaProvider) DefaultDisplayName() string { return "Okta" }
func (oktaProvider) DefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}

func (oktaProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	iss := oktaIssuer(rc.Extra)
	if iss == "" {
		iss = "https://configure-okta-issuer.invalid/oauth2/default"
	}
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  iss + "/v1/authorize",
			TokenURL: iss + "/v1/token",
		},
	}
}

func (oktaProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, rc RuntimeConfig) (*NormalizedProfile, error) {
	iss := oktaIssuer(rc.Extra)
	if iss == "" {
		return nil, fmt.Errorf("okta: set extra.okta_issuer or extra.okta_domain")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iss+"/v1/userinfo", nil)
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
		return nil, fmt.Errorf("okta userinfo: HTTP %d %s", resp.StatusCode, string(body))
	}
	var u struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	if strings.TrimSpace(u.Sub) == "" {
		return nil, fmt.Errorf("okta: empty sub")
	}
	return &NormalizedProfile{
		Subject:   strings.TrimSpace(u.Sub),
		Email:     strings.TrimSpace(u.Email),
		Name:      strings.TrimSpace(u.Name),
		AvatarURL: strings.TrimSpace(u.Picture),
	}, nil
}
