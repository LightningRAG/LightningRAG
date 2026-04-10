package channel

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EncodingAESKeyFromExtra 从 extra 中指定字段读取 43 字符 EncodingAESKey（微信公众平台/企业微信一致）
func EncodingAESKeyFromExtra(extra map[string]any, field string) (key []byte, ok bool, err error) {
	s := extraString(extra, field)
	if s == "" {
		return nil, false, nil
	}
	if len(s) != 43 {
		return nil, true, fmt.Errorf("%s 须为 43 字符", field)
	}
	key, err = base64.StdEncoding.DecodeString(s + "=")
	if err != nil {
		return nil, true, fmt.Errorf("%s base64: %w", field, err)
	}
	if len(key) != 32 {
		return nil, true, fmt.Errorf("%s decoded length want 32 got %d", field, len(key))
	}
	return key, true, nil
}

// WechatAESKeyFromExtra 从 extra.wechat_encoding_aes_key（43 字符）解码 32 字节 AES 密钥；未配置时 ok=false
func WechatAESKeyFromExtra(extra map[string]any) (key []byte, ok bool, err error) {
	return EncodingAESKeyFromExtra(extra, "wechat_encoding_aes_key")
}

// WechatAppIDFromExtra 安全模式下解密包尾校验与加密回复需要
func WechatAppIDFromExtra(extra map[string]any) string {
	return extraString(extra, "wechat_app_id")
}

// WechatMsgSignatureMatch 校验 msg_signature：SHA1(字典序 token、timestamp、nonce、encrypt 拼接)
func WechatMsgSignatureMatch(token, timestamp, nonce, encrypt, msgSignature string) bool {
	token = strings.TrimSpace(token)
	encrypt = strings.TrimSpace(encrypt)
	msgSignature = strings.ToLower(strings.TrimSpace(msgSignature))
	if token == "" || encrypt == "" || msgSignature == "" {
		return false
	}
	arr := []string{encrypt, nonce, timestamp, token}
	sort.Strings(arr)
	hex := fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(arr, ""))))
	if len(hex) != len(msgSignature) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hex), []byte(msgSignature)) == 1
}

func wechatPKCS7Unpad(p []byte, blockSize int) ([]byte, error) {
	if len(p) == 0 || len(p)%blockSize != 0 {
		return nil, errors.New("wechat: bad pad block")
	}
	pad := int(p[len(p)-1])
	if pad <= 0 || pad > blockSize {
		return nil, errors.New("wechat: bad pad byte")
	}
	return p[:len(p)-pad], nil
}

func wechatPKCS7Pad(p []byte, blockSize int) []byte {
	pad := blockSize - len(p)%blockSize
	if pad == 0 {
		pad = blockSize
	}
	out := make([]byte, len(p)+pad)
	copy(out, p)
	for i := len(p); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

// WechatDecryptEncryptBase64 解密 Encrypt 字段（Base64）
func WechatDecryptEncryptBase64(encryptB64 string, aesKey []byte) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encryptB64))
	if err != nil {
		return nil, err
	}
	if len(raw)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("wechat: cipher len")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	if len(aesKey) < 16 {
		return nil, fmt.Errorf("wechat: aes key")
	}
	iv := aesKey[:16]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(raw, raw)
	return wechatPKCS7Unpad(raw, aes.BlockSize)
}

// WechatUnpackDecryptedPayload random(16) + msg_len(4 BE) + msg + appid
func WechatUnpackDecryptedPayload(p []byte, expectAppID string) (msg []byte, err error) {
	if len(p) < 20 {
		return nil, fmt.Errorf("wechat: payload short")
	}
	msgLen := int(binary.BigEndian.Uint32(p[16:20]))
	if msgLen < 0 || len(p) < 20+msgLen {
		return nil, fmt.Errorf("wechat: bad msg len")
	}
	msg = make([]byte, msgLen)
	copy(msg, p[20:20+msgLen])
	tail := string(p[20+msgLen:])
	if expectAppID != "" && tail != expectAppID {
		return nil, fmt.Errorf("wechat: appid tail mismatch")
	}
	return msg, nil
}

// WechatEncryptMessageForReply 将被动回复 XML 明文打包为 Encrypt Base64
func WechatEncryptMessageForReply(plainXML []byte, appID string, aesKey []byte) (encryptB64 string, err error) {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return "", fmt.Errorf("wechat: need wechat_app_id for encrypt")
	}
	rnd := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	buf.Write(rnd)
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(plainXML)))
	buf.Write(lenBuf)
	buf.Write(plainXML)
	buf.WriteString(appID)
	padded := wechatPKCS7Pad(buf.Bytes(), aes.BlockSize)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, aesKey[:16])
	mode.CryptBlocks(padded, padded)
	return base64.StdEncoding.EncodeToString(padded), nil
}

// WechatEncryptedPassiveXML 安全/兼容模式下被动回复外层 XML
func WechatEncryptedPassiveXML(plainInnerXML []byte, token string, aesKey []byte, appID string) ([]byte, error) {
	enc, err := WechatEncryptMessageForReply(plainInnerXML, appID, aesKey)
	if err != nil {
		return nil, err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := wechatRandomNonce()
	sig := wechatMsgSigHex(token, ts, nonce, enc)
	return fmt.Appendf(nil, `<xml>
<Encrypt><![CDATA[%s]]></Encrypt>
<MsgSignature><![CDATA[%s]]></MsgSignature>
<TimeStamp>%s</TimeStamp>
<Nonce><![CDATA[%s]]></Nonce>
</xml>`, enc, sig, ts, nonce), nil
}

func wechatMsgSigHex(token, timestamp, nonce, encrypt string) string {
	arr := []string{encrypt, nonce, timestamp, token}
	sort.Strings(arr)
	return fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(arr, ""))))
}

func wechatRandomNonce() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// WechatOuterEncrypt 仅含 Encrypt 的外层（用于从 POST 体取密文）
type WechatOuterEncrypt struct {
	Encrypt string `xml:"Encrypt"`
}

// WechatDecryptInboundXML 若存在 Encrypt 则校验 msg_signature 并解密得到内层 XML 字节；否则返回原文且 usedCrypto=false
func WechatDecryptInboundXML(rawBody []byte, token, timestamp, nonce, msgSig string, aesKey []byte, appID string) (innerXML []byte, usedCrypto bool, err error) {
	var outer WechatOuterEncrypt
	if err := xml.Unmarshal(rawBody, &outer); err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(outer.Encrypt) == "" {
		return rawBody, false, nil
	}
	if strings.TrimSpace(msgSig) == "" {
		return nil, true, fmt.Errorf("wechat: 密文消息需 URL 参数 msg_signature")
	}
	if !WechatMsgSignatureMatch(token, timestamp, nonce, outer.Encrypt, msgSig) {
		return nil, true, fmt.Errorf("wechat msg_signature mismatch")
	}
	dec, err := WechatDecryptEncryptBase64(outer.Encrypt, aesKey)
	if err != nil {
		return nil, true, err
	}
	inner, err := WechatUnpackDecryptedPayload(dec, appID)
	if err != nil {
		return nil, true, err
	}
	return inner, true, nil
}
