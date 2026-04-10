package docparse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ParseTOMLText 将 TOML 转为缩进 JSON 文本，与 YAML/JSON 共用下游分块。
func ParseTOMLText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("TOML 数据为空")
	}
	var v any
	if err := toml.Unmarshal(data, &v); err != nil {
		return "", fmt.Errorf("解析 TOML 失败: %w", err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("TOML 转 JSON 失败: %w", err)
	}
	out := strings.TrimSpace(string(b))
	if out == "" {
		return "", fmt.Errorf("TOML 未产生文本")
	}
	return out, nil
}
