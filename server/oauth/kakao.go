package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
)

type kakaoProvider struct{}

func init() {
	Register(&kakaoProvider{})
}

func (kakaoProvider) Kind() string               { return "kakao" }
func (kakaoProvider) DefaultDisplayName() string { return "Kakao" }
func (kakaoProvider) DefaultScopes() []string {
	return []string{"profile_nickname", "account_email"}
}

func (kakaoProvider) OAuth2Config(rc RuntimeConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     rc.ClientID,
		ClientSecret: rc.ClientSecret,
		RedirectURL:  rc.RedirectURI,
		Scopes:       rc.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://kauth.kakao.com/oauth/authorize",
			TokenURL: "https://kauth.kakao.com/oauth/token",
		},
	}
}

func (kakaoProvider) FetchProfile(ctx context.Context, tok *oauth2.Token, _ RuntimeConfig) (*NormalizedProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://kapi.kakao.com/v2/user/me", nil)
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
		return nil, fmt.Errorf("kakao user/me: HTTP %d %s", resp.StatusCode, string(body))
	}
	var raw struct {
		ID           int64 `json:"id"`
		KakaoAccount struct {
			Email   string `json:"email"`
			Profile struct {
				Nickname         string `json:"nickname"`
				ProfileImageURL  string `json:"profile_image_url"`
				ThumbnailURL     string `json:"thumbnail_image_url"`
			} `json:"profile"`
		} `json:"kakao_account"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.ID == 0 {
		return nil, fmt.Errorf("kakao: empty id")
	}
	sub := strconv.FormatInt(raw.ID, 10)
	name := strings.TrimSpace(raw.KakaoAccount.Profile.Nickname)
	if name == "" {
		name = sub
	}
	avatar := strings.TrimSpace(raw.KakaoAccount.Profile.ProfileImageURL)
	if avatar == "" {
		avatar = strings.TrimSpace(raw.KakaoAccount.Profile.ThumbnailURL)
	}
	return &NormalizedProfile{
		Subject:   sub,
		Email:     strings.TrimSpace(raw.KakaoAccount.Email),
		Name:      name,
		AvatarURL: avatar,
	}, nil
}
