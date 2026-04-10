package oauth

import (
	"net/http"
	"time"
)

// HTTPClient 拉取用户信息等出站请求的统一超时，避免挂死
var HTTPClient = &http.Client{Timeout: 25 * time.Second}
