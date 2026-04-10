package system

// postInitDBCallbacks 在 InitDB 成功提交后执行（例如补全 RAG Casbin，见 initialize 包注册）。
var postInitDBCallbacks []func()

// RegisterPostInitDBCallback 注册 InitDB 完成后的回调。由 initialize 包 init 注册，避免 service 依赖 initialize 产生 import cycle。
func RegisterPostInitDBCallback(fn func()) {
	if fn == nil {
		return
	}
	postInitDBCallbacks = append(postInitDBCallbacks, fn)
}

// RunPostInitDBCallbacks 执行已注册的 InitDB 后回调。
func RunPostInitDBCallbacks() {
	for _, fn := range postInitDBCallbacks {
		fn()
	}
}
