package registry

import (
	"sort"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

// providerDisplayNames maps provider IDs to English UI labels (API / dropdowns; frontend may localize further).
var providerDisplayNames = map[string]string{
	"openai": "OpenAI", "ollama": "Ollama", "azure": "Azure", "azure_openai": "Azure OpenAI", "anthropic": "Anthropic",
	"deepseek": "DeepSeek", "xinference": "Xinference", "localai": "LocalAI",
	"siliconflow": "SiliconFlow", "moonshot": "Moonshot", "zhipu": "Zhipu AI", "openrouter": "OpenRouter",
	"groq": "Groq", "together": "Together", "tongyi": "Tongyi (Qwen)", "dashscope": "DashScope",
	"stepfun": "StepFun", "minimax": "Minimax", "hunyuan": "Hunyuan", "lingyi": "Lingyi (01.AI)", "01ai": "01.AI",
	"ai302": "302.AI", "jiekouai": "Jiekou AI", "giteeai": "Gitee AI", "cometapi": "Comet",
	"novitaai": "Novita", "deepinfra": "DeepInfra", "longcat": "LongCat", "ppio": "PPIO",
	"perfxcloud": "PerfxCloud", "upstage": "Upstage", "deerapi": "DeerAPI", "n1n": "N1N",
	"avian": "Avian", "xai": "xAI", "grok": "Grok (xAI)", "mistral": "Mistral", "gemini": "Gemini",
	"googlecloud": "Google Cloud (Gemini API)",
	"lmstudio":    "LM Studio", "baichuan": "Baichuan", "cohere": "Cohere", "volcengine": "Volcengine",
	"baiduyiyan": "ERNIE (Baidu)", "voyageai": "Voyage AI", "jina": "Jina", "jinaai": "Jina AI", "nvidia": "NVIDIA",
	"gpustack": "GPUStack", "vllm": "vLLM", "openai_api_compatible": "OpenAI-compatible API",
	"tokenpony": "TokenPony", "ragcon": "RAGcon", "xunfei": "iFlytek Spark", "huggingface": "HuggingFace",
	"modelscope": "ModelScope", "leptonai": "Lepton AI",
	"mineru": "MinerU", "paddleocr": "PaddleOCR",
	"bedrock": "Amazon Bedrock", "replicate": "Replicate", "fishaudio": "Fish Audio",
	"tencent": "Tencent Cloud ASR", "tencentcloud": "Tencent Cloud ASR",
	"builtin": "Builtin (TEI / OpenAI-compatible)", "youdao": "Youdao", "local": "Local (OpenAI-compatible CV)",
	"baai": "BAAI (TEI / OpenAI-compatible)", "fastembed": "FastEmbed (OpenAI-compatible)",
	"nomicai": "Nomic AI (OpenAI-compatible)", "sentence_transformers": "sentence-transformers (OpenAI-compatible)",
}

// ProviderOption 提供商选项，用于前端下拉
type ProviderOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// GetProviderLabel 获取提供商显示名称
func GetProviderLabel(provider string) string {
	if label, ok := providerDisplayNames[strings.ToLower(provider)]; ok {
		return label
	}
	// 首字母大写
	if len(provider) > 0 {
		return strings.ToUpper(provider[:1]) + strings.ToLower(provider[1:])
	}
	return provider
}

// listProvidersForScenario 获取单场景的提供商名称列表
func listProvidersForScenario(scenarioType string) []string {
	switch scenarioType {
	case interfaces.ModelTypeChat:
		return ListLLMProviders()
	case interfaces.ModelTypeEmbedding:
		return ListEmbeddingProviders()
	case interfaces.ModelTypeRerank:
		return ListRerankProviders()
	case interfaces.ModelTypeSpeech2Txt:
		return ListSpeech2TextProviders()
	case interfaces.ModelTypeTTS:
		return ListTTSProviders()
	case interfaces.ModelTypeOCR:
		return ListOCRProviders()
	case interfaces.ModelTypeCV:
		return ListCVProviders()
	default:
		return ListLLMProviders()
	}
}

// ListProvidersByScenario 根据场景类型返回可用提供商列表
// scenarioType 单场景；scenarioTypes 多场景时取并集
func ListProvidersByScenario(scenarioType string, scenarioTypes []string) []ProviderOption {
	types := scenarioTypes
	if len(types) == 0 && scenarioType != "" {
		types = []string{scenarioType}
	}
	if len(types) == 0 {
		types = []string{interfaces.ModelTypeChat}
	}
	seen := make(map[string]bool)
	for _, t := range types {
		for _, n := range listProvidersForScenario(t) {
			seen[n] = true
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	opts := make([]ProviderOption, 0, len(names))
	for _, n := range names {
		opts = append(opts, ProviderOption{Value: n, Label: GetProviderLabel(n)})
	}
	return opts
}
