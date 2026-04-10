package channel

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const botFrameworkJWKSURL = "https://login.botframework.com/v1/keys"

type botFrameworkJWKS struct {
	Keys []struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

var (
	bfJWKSBuf     []byte
	bfJWKSExpires time.Time
	bfJWKSMu      sync.Mutex
)

func fetchBotFrameworkJWKS() ([]byte, error) {
	bfJWKSMu.Lock()
	defer bfJWKSMu.Unlock()
	if len(bfJWKSBuf) > 0 && time.Now().Before(bfJWKSExpires) {
		return bfJWKSBuf, nil
	}
	req, err := http.NewRequest(http.MethodGet, botFrameworkJWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := ExternalHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("teams jwks: %s", resp.Status)
	}
	bfJWKSBuf = raw
	bfJWKSExpires = time.Now().Add(1 * time.Hour)
	return bfJWKSBuf, nil
}

func rsaPublicKeyFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nb)
	ei := new(big.Int).SetBytes(eb)
	if ei.Sign() <= 0 {
		return nil, errors.New("teams jwks: invalid exponent")
	}
	if !ei.IsInt64() {
		return nil, errors.New("teams jwks: exponent too large")
	}
	return &rsa.PublicKey{N: n, E: int(ei.Int64())}, nil
}

func botFrameworkPublicKeyForKID(kid string) (*rsa.PublicKey, error) {
	kid = strings.TrimSpace(kid)
	if kid == "" {
		return nil, errors.New("teams jwt: missing kid")
	}
	raw, err := fetchBotFrameworkJWKS()
	if err != nil {
		return nil, err
	}
	var doc botFrameworkJWKS
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	for _, k := range doc.Keys {
		if strings.TrimSpace(k.Kid) != kid || !strings.EqualFold(k.Kty, "RSA") {
			continue
		}
		return rsaPublicKeyFromJWK(k.N, k.E)
	}
	bfJWKSMu.Lock()
	bfJWKSBuf = nil
	bfJWKSExpires = time.Time{}
	bfJWKSMu.Unlock()
	return nil, fmt.Errorf("teams jwks: unknown kid %q", kid)
}

func teamsAudienceMatch(claims jwt.MapClaims, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return false
	}
	switch v := claims["aud"].(type) {
	case string:
		return strings.TrimSpace(v) == want
	case []any:
		for _, x := range v {
			if s, ok := x.(string); ok && strings.TrimSpace(s) == want {
				return true
			}
		}
	}
	return false
}

func isAllowedBotFrameworkIssuer(iss string) bool {
	iss = strings.TrimSpace(iss)
	if iss == "" {
		return false
	}
	if strings.HasPrefix(iss, "https://api.botframework.com") {
		return true
	}
	if strings.Contains(iss, "sts.windows.net") {
		return true
	}
	return false
}

// TeamsVerifyBearerToken 校验 Bot Framework 发往机器人的 Bearer JWT（签名、aud、iss、exp）。
func TeamsVerifyBearerToken(rawToken, expectAppID string) error {
	expectAppID = strings.TrimSpace(expectAppID)
	if expectAppID == "" {
		return errors.New("teams jwt: empty app id")
	}
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return errors.New("teams jwt: empty token")
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithLeeway(2*time.Minute),
	)
	token, err := parser.Parse(rawToken, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		return botFrameworkPublicKeyForKID(kid)
	})
	if err != nil {
		return fmt.Errorf("teams jwt: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("teams jwt: claims type")
	}
	iss, _ := claims["iss"].(string)
	if !isAllowedBotFrameworkIssuer(iss) {
		return errors.New("teams jwt: issuer not allowed")
	}
	if !teamsAudienceMatch(claims, expectAppID) {
		return errors.New("teams jwt: audience mismatch")
	}
	return nil
}
