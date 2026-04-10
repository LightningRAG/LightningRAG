package speech2text

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	asr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// TencentASR 腾讯云录音文件识别（异步任务 + 轮询），与上游参考 TencentCloudSeq2txt 行为一致
type TencentASR struct {
	client       *asr.Client
	model        string
	channelNum   uint64
	pollInterval time.Duration
	maxPolls     int
}

var tencentTimestampStrip = regexp.MustCompile(`\[\d+:\d+\.\d+,\d+:\d+\.\d+\]\s*`)

// NewTencentASR apiKey 为 JSON：{"tencent_cloud_sid":"...","tencent_cloud_sk":"..."}，也支持 secret_id/secret_key
// extra：tencent_region（如 ap-guangzhou，可空）、channel_num（uint64，默认 1）
// BaseURL：可选，自定义接入点主机名，如 asr.tencentcloudapi.com
func NewTencentASR(apiKey, baseURL, model string, extra map[string]any) (*TencentASR, error) {
	sid, sk, err := parseTencentCredentials(apiKey)
	if err != nil {
		return nil, err
	}
	region := ""
	if extra != nil {
		if v, ok := extra["tencent_region"].(string); ok {
			region = strings.TrimSpace(v)
		}
	}
	cred := common.NewCredential(sid, sk)
	cpf := profile.NewClientProfile()
	if ep := strings.TrimSpace(baseURL); ep != "" {
		ep = strings.TrimPrefix(strings.TrimPrefix(ep, "https://"), "http://")
		ep = strings.TrimSuffix(ep, "/")
		cpf.HttpProfile.Endpoint = ep
	}
	client, err := asr.NewClient(cred, region, cpf)
	if err != nil {
		return nil, err
	}
	if model == "" {
		model = "16k_zh"
	}
	poll := 5 * time.Second
	maxP := 60
	chNum := uint64(1)
	if extra != nil {
		switch v := extra["channel_num"].(type) {
		case float64:
			if v == 1 || v == 2 {
				chNum = uint64(v)
			}
		case int:
			if v == 1 || v == 2 {
				chNum = uint64(v)
			}
		case int64:
			if v == 1 || v == 2 {
				chNum = uint64(v)
			}
		}
		if v, ok := extra["tencent_poll_interval_sec"].(float64); ok && v > 0 {
			poll = time.Duration(v) * time.Second
		}
		if v, ok := extra["tencent_poll_max"].(float64); ok && v > 0 {
			maxP = int(v)
		}
	}
	return &TencentASR{
		client:       client,
		model:        model,
		channelNum:   chNum,
		pollInterval: poll,
		maxPolls:     maxP,
	}, nil
}

func parseTencentCredentials(apiKey string) (sid, sk string, err error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", "", fmt.Errorf("tencent asr: 需要密钥 JSON")
	}
	var m map[string]any
	if e := json.Unmarshal([]byte(apiKey), &m); e != nil {
		return "", "", fmt.Errorf("tencent asr: 密钥须为 JSON（tencent_cloud_sid / tencent_cloud_sk）: %w", e)
	}
	sid = firstString(m, "tencent_cloud_sid", "secret_id", "SecretId", "TENCENTCLOUD_SECRET_ID")
	sk = firstString(m, "tencent_cloud_sk", "secret_key", "SecretKey", "TENCENTCLOUD_SECRET_KEY")
	if sid == "" || sk == "" {
		return "", "", fmt.Errorf("tencent asr: JSON 中需同时包含 tencent_cloud_sid 与 tencent_cloud_sk（或 secret_id/secret_key）")
	}
	return sid, sk, nil
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if s := strings.TrimSpace(t); s != "" {
					return s
				}
			case fmt.Stringer:
				return strings.TrimSpace(t.String())
			}
		}
	}
	return ""
}

func (t *TencentASR) Transcribe(ctx context.Context, audio interface{}) (string, error) {
	raw, err := readAudioRaw(audio)
	if err != nil {
		return "", err
	}
	const maxBytes = 5 * 1024 * 1024
	if len(raw) > maxBytes {
		return "", fmt.Errorf("tencent asr: 音频超过 5MB 限制（当前 %d 字节）", len(raw))
	}
	b64 := base64.StdEncoding.EncodeToString(raw)
	dataLen := uint64(len(raw))
	resFmt := uint64(0)
	srcType := uint64(1)
	ch := t.channelNum
	req := asr.NewCreateRecTaskRequest()
	req.EngineModelType = &t.model
	req.ChannelNum = &ch
	req.ResTextFormat = &resFmt
	req.SourceType = &srcType
	req.Data = &b64
	req.DataLen = &dataLen

	createResp, err := t.client.CreateRecTaskWithContext(ctx, req)
	if err != nil {
		return "", fmt.Errorf("tencent asr CreateRecTask: %w", err)
	}
	if createResp.Response == nil || createResp.Response.Data == nil || createResp.Response.Data.TaskId == nil {
		return "", fmt.Errorf("tencent asr: CreateRecTask 无 TaskId")
	}
	taskID := *createResp.Response.Data.TaskId

	for i := 0; i < t.maxPolls; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		if i > 0 {
			time.Sleep(t.pollInterval)
		}
		q := asr.NewDescribeTaskStatusRequest()
		q.TaskId = &taskID
		stResp, err := t.client.DescribeTaskStatusWithContext(ctx, q)
		if err != nil {
			return "", fmt.Errorf("tencent asr DescribeTaskStatus: %w", err)
		}
		if stResp.Response == nil || stResp.Response.Data == nil {
			continue
		}
		data := stResp.Response.Data
		statusStr := ""
		if data.StatusStr != nil {
			statusStr = *data.StatusStr
		}
		if data.Status != nil && *data.Status == 2 {
			statusStr = "success"
		}
		if data.Status != nil && *data.Status == 3 {
			msg := ""
			if data.ErrorMsg != nil {
				msg = *data.ErrorMsg
			}
			return "", fmt.Errorf("tencent asr 任务失败: %s", msg)
		}
		switch statusStr {
		case "success":
			text := ""
			if data.Result != nil {
				text = *data.Result
			}
			text = tencentTimestampStrip.ReplaceAllString(text, "")
			return strings.TrimSpace(text), nil
		case "failed":
			msg := ""
			if data.ErrorMsg != nil {
				msg = *data.ErrorMsg
			}
			return "", fmt.Errorf("tencent asr 失败: %s", msg)
		}
	}
	return "", fmt.Errorf("tencent asr: 轮询超时（%d 次）", t.maxPolls)
}

func readAudioRaw(audio interface{}) ([]byte, error) {
	switch v := audio.(type) {
	case string:
		return os.ReadFile(v)
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("audio 需为文件路径或 []byte")
	}
}

func (t *TencentASR) ProviderName() string { return "tencent" }
func (t *TencentASR) ModelName() string    { return t.model }
