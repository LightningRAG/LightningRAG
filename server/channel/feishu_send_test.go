package channel

import "testing"

func TestFeishuAPIBaseFromExtra(t *testing.T) {
	if got := FeishuAPIBaseFromExtra(nil); got != feishuDefaultAPIBase {
		t.Fatalf("default: want %q got %q", feishuDefaultAPIBase, got)
	}
	m := map[string]any{"lark_api_base": "https://open.larksuite.com"}
	if got := FeishuAPIBaseFromExtra(m); got != "https://open.larksuite.com" {
		t.Fatalf("lark: %q", got)
	}
	m2 := map[string]any{"feishu_api_base": "open.feishu.cn"}
	if got := FeishuAPIBaseFromExtra(m2); got != "https://open.feishu.cn" {
		t.Fatalf("scheme prepend: %q", got)
	}
}
