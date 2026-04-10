package request

// EffectiveTocEnhance 合并 REST 的 tocEnhance 与 Ragflow / Python SDK 常用的 toc_enhance；camelCase 优先。
func EffectiveTocEnhance(camel, snake *bool) *bool {
	if camel != nil {
		return camel
	}
	return snake
}
