package channel

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DingTalkAESKeyFromExtra 钉钉 HTTP 回调 EncodingAESKey：43 字符，Base64 解码为 32 字节
func DingTalkAESKeyFromExtra(extra map[string]any) (key []byte, ok bool, err error) {
	s := extraString(extra, "dingtalk_encoding_aes_key")
	if s == "" {
		return nil, false, nil
	}
	if len(s) != 43 {
		return nil, true, fmt.Errorf("dingtalk_encoding_aes_key 须为 43 字符")
	}
	key, err = base64.StdEncoding.DecodeString(s + "=")
	if err != nil {
		return nil, true, fmt.Errorf("dingtalk_encoding_aes_key base64: %w", err)
	}
	if len(key) != 32 {
		return nil, true, fmt.Errorf("dingtalk aes key length want 32 got %d", len(key))
	}
	return key, true, nil
}

// DingTalkVerifyURLSignature 校验 URL 参数 signature：SHA1(字典序拼接 token、timestamp、nonce、encrypt)
func DingTalkVerifyURLSignature(token string, q url.Values, encrypt string) bool {
	token = strings.TrimSpace(token)
	encrypt = strings.TrimSpace(encrypt)
	if q == nil || token == "" || encrypt == "" {
		return false
	}
	sig := strings.TrimSpace(q.Get("signature"))
	if sig == "" {
		return false
	}
	ts := q.Get("timestamp")
	nonce := q.Get("nonce")
	arr := []string{encrypt, nonce, token, ts}
	sort.Strings(arr)
	hex := fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(arr, ""))))
	if len(hex) != len(sig) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hex), []byte(sig)) == 1
}

func dingTalkMsgSigHex(token, timestamp, nonce, encrypt string) string {
	arr := []string{encrypt, nonce, timestamp, token}
	sort.Strings(arr)
	return fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(arr, ""))))
}

func dingTalkRandomNonce() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// DingTalkBuildEncryptedSuccessResponse 开放平台校验 URL 后需返回的 JSON（encrypt 为 success 的密文）
func DingTalkBuildEncryptedSuccessResponse(token string, aesKey []byte, suiteKey string) ([]byte, error) {
	token = strings.TrimSpace(token)
	suiteKey = strings.TrimSpace(suiteKey)
	if token == "" || len(aesKey) != 32 {
		return nil, fmt.Errorf("dingtalk: token 或 aes key 无效")
	}
	if suiteKey == "" {
		return nil, fmt.Errorf("dingtalk: 需配置 dingtalk_suite_key（解密包尾与加密 success 一致）")
	}
	enc, err := WechatEncryptMessageForReply([]byte("success"), suiteKey, aesKey)
	if err != nil {
		return nil, err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := dingTalkRandomNonce()
	sig := dingTalkMsgSigHex(token, ts, nonce, enc)
	m := map[string]string{
		"msg_signature": sig,
		"timeStamp":     ts,
		"nonce":         nonce,
		"encrypt":       enc,
	}
	return json.Marshal(m)
}

// DingTalkSuiteKeyFromExtra 解密包尾 / 加密 success 尾缀（开放平台一般为 suiteKey 或企业 appKey）
func DingTalkSuiteKeyFromExtra(extra map[string]any) string {
	return extraString(extra, "dingtalk_suite_key")
}
