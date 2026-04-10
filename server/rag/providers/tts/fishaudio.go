package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// FishAudio Fish Audio TTS（MsgPack + api-key 头），与上游参考 FishAudioTTS 行为一致
type FishAudio struct {
	apiKey  string
	baseURL string
	model   string
	refID   string
	client  *http.Client
}

type fishTTSRequest struct {
	Text        string        `msgpack:"text"`
	ChunkLength int           `msgpack:"chunk_length"`
	Format      string        `msgpack:"format"`
	MP3Bitrate  int           `msgpack:"mp3_bitrate"`
	References  []interface{} `msgpack:"references"`
	ReferenceID *string       `msgpack:"reference_id,omitempty"`
	Normalize   bool          `msgpack:"normalize"`
	Latency     string        `msgpack:"latency"`
}

// ParseFishAudioCredentials 解析 API Key：支持 JSON {"fish_audio_ak":"","fish_audio_refid":""}，或纯字符串 ak
func ParseFishAudioCredentials(keyStr string) (apiKey, refID string, err error) {
	keyStr = strings.TrimSpace(keyStr)
	if keyStr == "" {
		return "", "", fmt.Errorf("fish audio: 缺少 API Key")
	}
	if keyStr[0] == '{' || strings.Contains(keyStr, "fish_audio_ak") {
		var m map[string]any
		if e := json.Unmarshal([]byte(keyStr), &m); e != nil {
			return "", "", fmt.Errorf("fish audio: 解析密钥 JSON: %w", e)
		}
		if v, ok := m["fish_audio_ak"].(string); ok {
			apiKey = v
		}
		if v, ok := m["fish_audio_refid"].(string); ok {
			refID = v
		}
		if apiKey == "" {
			return "", "", fmt.Errorf("fish audio: JSON 中缺少 fish_audio_ak")
		}
		return apiKey, refID, nil
	}
	return keyStr, "", nil
}

// NewFishAudio 创建 Fish Audio TTS；baseURL 为空时使用官方默认
func NewFishAudio(apiKey, baseURL, model string) (*FishAudio, error) {
	ak, ref, err := ParseFishAudioCredentials(apiKey)
	if err != nil {
		return nil, err
	}
	if baseURL == "" {
		baseURL = "https://api.fish.audio/v1/tts"
	}
	return &FishAudio{
		apiKey:  ak,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		refID:   ref,
		client:  &http.Client{},
	}, nil
}

func (f *FishAudio) Synthesize(ctx context.Context, text string, voice string) (io.ReadCloser, error) {
	text = normalizeFishText(text)
	ref := f.refID
	if strings.TrimSpace(voice) != "" {
		ref = strings.TrimSpace(voice)
	}
	var refPtr *string
	if ref != "" {
		refPtr = &ref
	}
	reqBody := fishTTSRequest{
		Text:        text,
		ChunkLength: 200,
		Format:      "mp3",
		MP3Bitrate:  128,
		References:  []interface{}{},
		ReferenceID: refPtr,
		Normalize:   true,
		Latency:     "normal",
	}
	payload, err := msgpack.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("api-key", f.apiKey)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("fish audio tts: %s", string(b))
	}
	return resp.Body, nil
}

func normalizeFishText(s string) string {
	// 与上游参考 normalize_text 类似：去掉 Markdown 噪声
	s = strings.ReplaceAll(s, "**", "")
	for _, ch := range []string{"#"} {
		s = strings.ReplaceAll(s, ch, "")
	}
	return strings.TrimSpace(s)
}

func (f *FishAudio) ProviderName() string { return "fishaudio" }
func (f *FishAudio) ModelName() string    { return f.model }
