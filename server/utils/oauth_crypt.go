package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// 加密材料：优先库表 sys_oauth_settings.secret_key；未配置时回退 JWT 签名密钥（与登录签发共用）。
// 仅被 service/oauthapp（提供商密钥）等调用。
func oauthSecretMaterial() []byte {
	s := strings.TrimSpace(global.OAuthSecretKeyFromDB())
	if s == "" {
		s = global.LRAG_CONFIG.JWT.SigningKey
	}
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}

// EncryptOAuthSecret AES-256-GCM，输出 base64(nonce||ciphertext)
func EncryptOAuthSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	key := oauthSecretMaterial()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptOAuthSecret 解密 EncryptOAuthSecret 的结果；空串返回空
func DecryptOAuthSecret(enc string) (string, error) {
	if enc == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	key := oauthSecretMaterial()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", errors.New("oauth secret: ciphertext too short")
	}
	nonce, ct := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
