package docparse

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseYAMLText 将 YAML 转为缩进 JSON 文本，便于下游与 JSON 文档走同一套切片逻辑。
func ParseYAMLText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("YAML 数据为空")
	}
	var v any
	if err := yaml.Unmarshal(data, &v); err != nil {
		return "", fmt.Errorf("解析 YAML 失败: %w", err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("YAML 转 JSON 失败: %w", err)
	}
	out := strings.TrimSpace(string(b))
	if out == "" {
		return "", fmt.Errorf("YAML 未产生文本")
	}
	return out, nil
}
