package component

import (
	"fmt"
	"strings"
	"sync"
)

func init() {
	Register("AwaitResponse", NewAwaitResponse)
}

// AwaitResponse 向用户展示提示并从全局变量读取用户补充输入（对齐上游编排 Await Response）。
// 客户端应在后续请求的 workflowGlobals 中传入 variable_key 对应的值（默认 sys.await_reply）。
type AwaitResponse struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewAwaitResponse 创建 AwaitResponse 组件
func NewAwaitResponse(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &AwaitResponse{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

func (a *AwaitResponse) ComponentName() string { return "AwaitResponse" }

func (a *AwaitResponse) Invoke(map[string]any) error {
	a.mu.Lock()
	a.err = ""
	a.mu.Unlock()

	msg := a.canvas.ResolveString(getStrParam(a.params, "message"))
	if rc := a.canvas.RunContext(); rc != nil && msg != "" && rc.GetStreamCallback() != nil {
		rc.GetStreamCallback()(msg)
	}

	varKey := strings.TrimSpace(getStrParam(a.params, "variable_key"))
	if varKey == "" {
		varKey = "sys.await_reply"
	}
	if strings.Contains(varKey, "@") {
		a.fail("variable_key 应为全局键名，如 sys.await_reply")
		return fmt.Errorf("非法 variable_key")
	}

	g := a.canvas.GetGlobals()
	var raw any
	if g != nil {
		raw = g[varKey]
	}
	text := strings.TrimSpace(fmt.Sprint(raw))

	require := getBoolParamDefault(a.params, "require_non_empty", true)
	if require && text == "" {
		hint := fmt.Sprintf("请在请求体 workflowGlobals 中设置 %q（可与用户当前 query 配合使用）", varKey)
		a.fail(hint)
		return fmt.Errorf("%s", hint)
	}

	a.mu.Lock()
	a.output["user_input"] = text
	a.output["formalized_content"] = text
	a.output["content"] = text
	a.mu.Unlock()
	return nil
}

func (a *AwaitResponse) fail(msg string) {
	a.mu.Lock()
	a.err = msg
	a.mu.Unlock()
}

func (a *AwaitResponse) Output(key string) any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.output[key]
}

func (a *AwaitResponse) OutputAll() map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range a.output {
		out[k] = v
	}
	return out
}

func (a *AwaitResponse) SetOutput(key string, value any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output[key] = value
}

func (a *AwaitResponse) Error() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.err
}

func (a *AwaitResponse) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output = make(map[string]any)
	a.err = ""
}
