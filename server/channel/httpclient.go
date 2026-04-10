package channel

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

// 出站调用第三方平台 API 的统一超时，避免 goroutine 长期挂起
var externalHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
}

// ExternalHTTPClient 用于飞书 / 钉钉 / Discord 等外呼
func ExternalHTTPClient() *http.Client {
	return externalHTTPClient
}

// ExternalHTTPDo 对同一请求体执行最多 3 次 Do：遇 502/503/504 或网络超时等可重试错误时退避重试。
func ExternalHTTPDo(req *http.Request) (*http.Response, error) {
	client := externalHTTPClient
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(200*attempt) * time.Millisecond)
		}
		r := req.Clone(req.Context())
		if len(body) > 0 {
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		resp, err := client.Do(r)
		if err != nil {
			lastErr = err
			if attempt < 2 && shouldRetryOutboundErr(err) {
				continue
			}
			return nil, err
		}
		if attempt < 2 && shouldRetryOutboundStatus(resp.StatusCode) {
			_ = resp.Body.Close()
			continue
		}
		return resp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("channel: http retry exhausted")
}

func shouldRetryOutboundStatus(code int) bool {
	switch code {
	case 502, 503, 504:
		return true
	default:
		return false
	}
}

func shouldRetryOutboundErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	return errors.Is(err, io.EOF)
}
