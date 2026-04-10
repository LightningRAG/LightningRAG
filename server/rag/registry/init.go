package registry

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/cv"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/embedding"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/llm"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/ocr"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/rerank"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/speech2text"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/tts"
	"github.com/LightningRAG/LightningRAG/server/rag/providers/vectorstore"
)

func strFromExtra(extra map[string]any, key string) string {
	if extra == nil {
		return ""
	}
	if v, ok := extra[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func parseVolcEngineKey(keyStr string) (arkAPIKey, epID, endpointID string) {
	var m map[string]any
	if err := json.Unmarshal([]byte(keyStr), &m); err != nil {
		return keyStr, "", ""
	}
	if v, ok := m["ark_api_key"].(string); ok {
		arkAPIKey = v
	}
	if v, ok := m["ep_id"].(string); ok {
		epID = v
	}
	if v, ok := m["endpoint_id"].(string); ok {
		endpointID = v
	}
	return arkAPIKey, epID, endpointID
}

func init() {
	// ========== LLM (Chat) 提供商 ==========
	RegisterLLM("openai", func(config LLMConfig) (interfaces.LLM, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return llm.NewOpenAI(config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterLLM("ollama", func(config LLMConfig) (interfaces.LLM, error) {
		return llm.NewOllama(config.BaseURL, config.ModelName), nil
	})
	RegisterLLM("azure", func(config LLMConfig) (interfaces.LLM, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return llm.NewAzure(config.APIKey, config.BaseURL, config.ModelName, apiVer), nil
	})
	RegisterLLM("anthropic", func(config LLMConfig) (interfaces.LLM, error) {
		return llm.NewAnthropic(config.APIKey, config.BaseURL, config.ModelName), nil
	})
	// OpenAI 兼容 API 的厂商（同一实现，不同默认 baseURL）
	registerOpenAICompatLLM("deepseek", "https://api.deepseek.com")
	registerOpenAICompatLLM("xinference", "")
	registerOpenAICompatLLM("localai", "")
	registerOpenAICompatLLM("siliconflow", "https://api.siliconflow.cn/v1")
	registerOpenAICompatLLM("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatLLM("zhipu", "https://open.bigmodel.cn/api/paas/v4")
	registerOpenAICompatLLM("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatLLM("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatLLM("together", "https://api.together.xyz/v1")
	// 高优先级：国内常用 OpenAI 兼容
	registerOpenAICompatLLM("tongyi", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatLLM("dashscope", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatLLM("stepfun", "https://api.stepfun.com/v1")
	registerOpenAICompatLLM("minimax", "https://api.minimaxi.com/v1")
	registerOpenAICompatLLM("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatLLM("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatLLM("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatLLM("ai302", "https://api.302.ai/v1")
	registerOpenAICompatLLM("jiekouai", "https://api.jiekou.ai/openai")
	registerOpenAICompatLLM("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatLLM("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatLLM("novitaai", "https://api.novita.ai/v3/openai")
	// 中优先级：国际/其他 OpenAI 兼容
	registerOpenAICompatLLM("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatLLM("longcat", "https://api.longcat.chat/openai")
	registerOpenAICompatLLM("ppio", "https://api.ppinfra.com/v3/openai")
	registerOpenAICompatLLM("perfxcloud", "https://cloud.perfxlab.cn/v1")
	registerOpenAICompatLLM("upstage", "https://api.upstage.ai/v1/solar")
	registerOpenAICompatLLM("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatLLM("n1n", "https://api.n1n.ai/v1")
	registerOpenAICompatLLM("avian", "https://api.avian.io/v1")
	// 未覆盖厂商：xAI/Mistral/Gemini 等（OpenAI 兼容）
	registerOpenAICompatLLM("xai", "https://api.x.ai/v1")
	// 工厂别名 grok：与 xAI Grok 同一 API
	registerOpenAICompatLLM("grok", "https://api.x.ai/v1")
	registerOpenAICompatLLM("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatLLM("gemini", "https://generativelanguage.googleapis.com/v1beta")
	// 「Google Cloud」侧 Gemini API Key 路径（非 Vertex 服务账号）
	registerOpenAICompatLLM("googlecloud", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatLLM("lmstudio", "")
	registerOpenAICompatLLM("baichuan", "https://api.baichuan-ai.com/v1")
	// Cohere 自有 API
	RegisterLLM("cohere", func(config LLMConfig) (interfaces.LLM, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.cohere.ai"
		}
		model := config.ModelName
		if model == "" {
			model = "command-r-plus"
		}
		return llm.NewCohere(config.APIKey, baseURL, model), nil
	})
	// 百度文心：新版 API 使用 Bearer API_Key，OpenAI 兼容
	registerOpenAICompatLLM("baiduyiyan", "https://qianfan.baidubce.com/v2")
	// VolcEngine 火山引擎：key 可为 JSON {"ark_api_key":"","ep_id":"","endpoint_id":""}，model 为 ep_id+endpoint_id
	RegisterLLM("volcengine", func(config LLMConfig) (interfaces.LLM, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://ark.cn-beijing.volces.com/api/v3"
		}
		apiKey := config.APIKey
		model := config.ModelName
		if apiKey != "" && (apiKey[0] == '{' || strings.Contains(apiKey, "ark_api_key")) {
			arkKey, epID, epEndID := parseVolcEngineKey(apiKey)
			apiKey = arkKey
			if model == "" && epID != "" {
				model = epID + epEndID
			}
		}
		if model == "" {
			model = strFromExtra(config.Extra, "ep_id") + strFromExtra(config.Extra, "endpoint_id")
		}
		if apiKey == "" {
			apiKey = strFromExtra(config.Extra, "ark_api_key")
		}
		impl := llm.NewOpenAI(apiKey, baseURL, model)
		return &llm.ProviderNameAdapter{LLM: impl, DisplayName: "volcengine"}, nil
	})
	// 与 references 目录内上游 llm 注册对齐：GPUStack、NVIDIA Chat、VLLM、OpenAI-API-Compatible、TokenPony、RAGcon、讯飞星火、HF/ModelScope、LeptonAI
	registerOpenAICompatLLM("gpustack", "")
	RegisterLLM("nvidia", func(config LLMConfig) (interfaces.LLM, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://integrate.api.nvidia.com/v1"
		}
		impl := llm.NewOpenAI(config.APIKey, baseURL, config.ModelName)
		return &llm.ProviderNameAdapter{LLM: impl, DisplayName: "nvidia"}, nil
	})
	registerOpenAICompatLLM("vllm", "")
	registerOpenAICompatLLM("openai_api_compatible", "")
	registerOpenAICompatLLM("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatLLM("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatLLM("xunfei", "https://spark-api-open.xf-yun.com/v1")
	registerOpenAICompatLLM("huggingface", "")
	registerOpenAICompatLLM("modelscope", "")
	RegisterLLM("leptonai", func(config LLMConfig) (interfaces.LLM, error) {
		baseURL := strings.TrimSpace(config.BaseURL)
		if baseURL == "" {
			m := strings.TrimSpace(config.ModelName)
			if m == "" {
				return nil, fmt.Errorf("leptonai: 请填写 Base URL，或填写模型名以使用 https://{模型名}.lepton.run/api/v1")
			}
			baseURL = fmt.Sprintf("https://%s.lepton.run/api/v1", m)
		}
		impl := llm.NewOpenAI(config.APIKey, baseURL, config.ModelName)
		return &llm.ProviderNameAdapter{LLM: impl, DisplayName: "leptonai"}, nil
	})
	RegisterLLM("bedrock", func(config LLMConfig) (interfaces.LLM, error) {
		region := strFromExtra(config.Extra, "aws_region")
		return llm.NewBedrock(config.APIKey, region, config.ModelName, config.Extra)
	})
	RegisterLLM("replicate", func(config LLMConfig) (interfaces.LLM, error) {
		if strings.TrimSpace(config.APIKey) == "" {
			return nil, fmt.Errorf("replicate: 需要 API Token")
		}
		if strings.TrimSpace(config.ModelName) == "" {
			return nil, fmt.Errorf("replicate: 模型名需为 Replicate 的 owner/name，例如 meta/llama-3-8b-instruct")
		}
		return llm.NewReplicate(config.APIKey, config.ModelName), nil
	})

	// ========== Embedding 提供商 ==========
	RegisterEmbedding("openai", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return embedding.NewOpenAIEmbed(config.APIKey, baseURL, config.ModelName, config.Dimensions), nil
	})
	RegisterEmbedding("azure", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return embedding.NewAzureEmbed(config.APIKey, config.BaseURL, config.ModelName, apiVer, config.Dimensions), nil
	})
	RegisterEmbedding("ollama", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		return embedding.NewOllamaEmbed(config.BaseURL, config.ModelName), nil
	})
	registerOpenAICompatEmbedding("deepseek", "https://api.deepseek.com")
	registerOpenAICompatEmbedding("xinference", "")
	registerOpenAICompatEmbedding("localai", "")
	registerOpenAICompatEmbedding("siliconflow", "https://api.siliconflow.cn/v1")
	registerOpenAICompatEmbedding("tongyi", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatEmbedding("dashscope", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatEmbedding("stepfun", "https://api.stepfun.com/v1")
	registerOpenAICompatEmbedding("zhipu", "https://open.bigmodel.cn/api/paas/v4")
	registerOpenAICompatEmbedding("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatEmbedding("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatEmbedding("novitaai", "https://api.novita.ai/v3/openai")
	registerOpenAICompatEmbedding("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatEmbedding("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatEmbedding("xai", "https://api.x.ai/v1")
	registerOpenAICompatEmbedding("grok", "https://api.x.ai/v1")
	registerOpenAICompatEmbedding("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatEmbedding("gemini", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatEmbedding("googlecloud", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatEmbedding("baiduyiyan", "https://qianfan.baidubce.com/v2")
	registerOpenAICompatEmbedding("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatEmbedding("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatEmbedding("baichuan", "https://api.baichuan-ai.com/v1")
	registerOpenAICompatEmbedding("longcat", "https://api.longcat.chat/openai")
	registerOpenAICompatEmbedding("ppio", "https://api.ppinfra.com/v3/openai")
	registerOpenAICompatEmbedding("perfxcloud", "https://cloud.perfxlab.cn/v1")
	registerOpenAICompatEmbedding("upstage", "https://api.upstage.ai/v1/solar")
	registerOpenAICompatEmbedding("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatEmbedding("n1n", "https://api.n1n.ai/v1")
	registerOpenAICompatEmbedding("avian", "https://api.avian.io/v1")
	registerOpenAICompatEmbedding("ai302", "https://api.302.ai/v1")
	registerOpenAICompatEmbedding("jiekouai", "https://api.jiekou.ai/openai")
	registerOpenAICompatEmbedding("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatEmbedding("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatEmbedding("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatEmbedding("minimax", "https://api.minimaxi.com/v1")
	registerOpenAICompatEmbedding("lmstudio", "")
	// 参照 references 目录内 补充：Voyage AI、Jina、Cohere、VolcEngine
	RegisterEmbedding("voyageai", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.voyageai.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "voyage-3"
		}
		return embedding.NewVoyageEmbed(config.APIKey, baseURL, model, config.Dimensions), nil
	})
	RegisterEmbedding("jina", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.jina.ai/v1/embeddings"
		}
		model := config.ModelName
		if model == "" {
			model = "jina-embeddings-v3"
		}
		return embedding.NewJinaEmbed(config.APIKey, baseURL, model, config.Dimensions), nil
	})
	// 工厂别名 jinaai：与 Jina 相同 HTTP 实现
	RegisterEmbedding("jinaai", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.jina.ai/v1/embeddings"
		}
		model := config.ModelName
		if model == "" {
			model = "jina-embeddings-v3"
		}
		return embedding.NewJinaEmbed(config.APIKey, baseURL, model, config.Dimensions), nil
	})
	RegisterEmbedding("cohere", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.cohere.ai"
		}
		model := config.ModelName
		if model == "" {
			model = "embed-english-v3.0"
		}
		return embedding.NewCohereEmbed(config.APIKey, baseURL, model, config.Dimensions), nil
	})
	RegisterEmbedding("volcengine", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://ark.cn-beijing.volces.com/api/v3"
		}
		apiKey := config.APIKey
		if apiKey != "" && (apiKey[0] == '{' || strings.Contains(apiKey, "ark_api_key")) {
			arkKey, _, _ := parseVolcEngineKey(apiKey)
			apiKey = arkKey
		}
		if apiKey == "" {
			apiKey = strFromExtra(config.Extra, "ark_api_key")
		}
		model := config.ModelName
		if model == "" {
			model = strFromExtra(config.Extra, "ep_id") + strFromExtra(config.Extra, "endpoint_id")
		}
		return embedding.NewVolcEngineEmbed(apiKey, baseURL, model, config.Dimensions), nil
	})
	RegisterEmbedding("nvidia", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://integrate.api.nvidia.com/v1/embeddings"
		}
		model := config.ModelName
		if model == "" {
			model = "nvidia/nv-embed-qa-4"
		}
		return embedding.NewNvidiaEmbed(config.APIKey, baseURL, model, config.Dimensions), nil
	})
	registerOpenAICompatEmbedding("gpustack", "")
	registerOpenAICompatEmbedding("together", "https://api.together.xyz/v1")
	registerOpenAICompatEmbedding("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatEmbedding("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatEmbedding("xunfei", "https://spark-api-open.xf-yun.com/v1")
	registerOpenAICompatEmbedding("huggingface", "")
	registerOpenAICompatEmbedding("modelscope", "")
	// BuiltinEmbed：对接本地 TEI / OpenAI 兼容嵌入服务，需填写 Base URL
	registerOpenAICompatEmbedding("builtin", "")
	// Youdao：官方云 API 为 OpenAI 兼容 /v1/embeddings；本地 FlagEmbedding 请自建兼容网关并填 Base URL
	registerOpenAICompatEmbedding("youdao", "")
	// 本地嵌入类工厂（界面侧常隐藏默认 URL）：对接 TEI / 自建 OpenAI 兼容服务时需填 Base URL
	registerOpenAICompatEmbedding("baai", "")
	registerOpenAICompatEmbedding("fastembed", "")
	registerOpenAICompatEmbedding("nomicai", "")
	registerOpenAICompatEmbedding("sentence_transformers", "")
	RegisterEmbedding("bedrock", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		region := strFromExtra(config.Extra, "aws_region")
		return embedding.NewBedrockEmbed(config.APIKey, region, config.ModelName, config.Extra, config.Dimensions)
	})
	RegisterEmbedding("replicate", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		if strings.TrimSpace(config.APIKey) == "" {
			return nil, fmt.Errorf("replicate embed: 需要 API Token")
		}
		if strings.TrimSpace(config.ModelName) == "" {
			return nil, fmt.Errorf("replicate embed: 需要模型名 owner/name")
		}
		inputKey := strFromExtra(config.Extra, "replicate_embed_input_key")
		return embedding.NewReplicateEmbed(config.APIKey, config.ModelName, config.Dimensions, inputKey), nil
	})

	// ========== VectorStore 向量存储 ==========
	RegisterVectorStore("postgresql", func(config VectorStoreConfig, embedder interfaces.Embedder, namespace string, vectorDims int) (interfaces.VectorStore, error) {
		return vectorstore.NewPgVectorStore(global.LRAG_DB, embedder, namespace, vectorDims)
	})
	RegisterVectorStore("mysql", func(config VectorStoreConfig, embedder interfaces.Embedder, namespace string, vectorDims int) (interfaces.VectorStore, error) {
		return vectorstore.NewMysqlVectorStore(global.LRAG_DB, embedder, namespace, vectorDims)
	})
	RegisterVectorStore("elasticsearch", func(config VectorStoreConfig, embedder interfaces.Embedder, namespace string, vectorDims int) (interfaces.VectorStore, error) {
		return vectorstore.NewElasticsearchStore(config.Config, embedder, namespace, vectorDims)
	})

	// ========== Rerank 提供商 ==========
	RegisterRerank("jina", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.jina.ai/v1/rerank"
		}
		model := config.ModelName
		if model == "" {
			model = "jina-reranker-v2-base-multilingual"
		}
		return rerank.NewJina(config.APIKey, baseURL, model), nil
	})
	RegisterRerank("jinaai", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.jina.ai/v1/rerank"
		}
		model := config.ModelName
		if model == "" {
			model = "jina-reranker-v2-base-multilingual"
		}
		impl := rerank.NewJina(config.APIKey, baseURL, model)
		return &rerank.ProviderNameAdapter{Reranker: impl, DisplayName: "jinaai"}, nil
	})
	RegisterRerank("xinference", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			return nil, nil
		}
		return rerank.NewOpenAICompat("xinference", config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterRerank("localai", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			return nil, nil
		}
		return rerank.NewOpenAICompat("localai", config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterRerank("siliconflow", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.siliconflow.cn/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "BAAI/bge-reranker-v2-m3"
		}
		return rerank.NewOpenAICompat("siliconflow", config.APIKey, baseURL, model), nil
	})
	RegisterRerank("novitaai", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.novita.ai/v3/openai"
		}
		return rerank.NewOpenAICompat("novitaai", config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterRerank("giteeai", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://ai.gitee.com/v1"
		}
		return rerank.NewOpenAICompat("giteeai", config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterRerank("cohere", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.cohere.ai"
		}
		model := config.ModelName
		if model == "" {
			model = "rerank-v3.5"
		}
		return rerank.NewCohere(config.APIKey, baseURL, model), nil
	})
	RegisterRerank("tongyi", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://dashscope.aliyuncs.com/api/v1/services/rerank/text-rerank"
		}
		model := config.ModelName
		if model == "" {
			model = "gte-rerank"
		}
		return rerank.NewTongyi(config.APIKey, baseURL, model), nil
	})
	RegisterRerank("dashscope", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://dashscope.aliyuncs.com/api/v1/services/rerank/text-rerank"
		}
		model := config.ModelName
		if model == "" {
			model = "gte-rerank"
		}
		impl := rerank.NewTongyi(config.APIKey, baseURL, model)
		return &rerank.ProviderNameAdapter{Reranker: impl, DisplayName: "dashscope"}, nil
	})
	// 参照 references 目录内 补充：Voyage AI、NVIDIA、302.AI、Jiekou.AI
	RegisterRerank("voyageai", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.voyageai.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "rerank-2"
		}
		return rerank.NewVoyage(config.APIKey, baseURL, model), nil
	})
	RegisterRerank("nvidia", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://ai.api.nvidia.com/v1/retrieval/nvidia"
		}
		model := config.ModelName
		if model == "" {
			model = "nv-rerankqa-mistral-4b-v3"
		}
		return rerank.NewNvidia(config.APIKey, baseURL, model), nil
	})
	RegisterRerank("ai302", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.302.ai/v1"
		}
		return rerank.NewOpenAICompat("ai302", config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterRerank("jiekouai", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.jiekou.ai/openai/v1"
		}
		return rerank.NewOpenAICompat("jiekouai", config.APIKey, baseURL, config.ModelName), nil
	})
	RegisterRerank("baiduyiyan", func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://qianfan.baidubce.com/v2/rerank"
		}
		return rerank.NewBaiduQianfan(config.APIKey, baseURL, config.ModelName)
	})
	RegisterRerank("huggingface", func(config RerankConfig) (interfaces.Reranker, error) {
		return rerank.NewHuggingFaceHTTP(config.BaseURL, config.ModelName), nil
	})
	registerOpenAICompatRerank("gpustack", "")
	registerOpenAICompatRerank("vllm", "")
	registerOpenAICompatRerank("openai_api_compatible", "")
	registerOpenAICompatRerank("lmstudio", "")
	registerOpenAICompatRerank("together", "https://api.together.xyz/v1")
	registerOpenAICompatRerank("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatRerank("deepseek", "https://api.deepseek.com")
	registerOpenAICompatRerank("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatRerank("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatRerank("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatRerank("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatRerank("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatRerank("zhipu", "https://open.bigmodel.cn/api/paas/v4")
	registerOpenAICompatRerank("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatRerank("xai", "https://api.x.ai/v1")
	registerOpenAICompatRerank("grok", "https://api.x.ai/v1")
	registerOpenAICompatRerank("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatRerank("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatRerank("gemini", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatRerank("googlecloud", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatRerank("longcat", "https://api.longcat.chat/openai")
	registerOpenAICompatRerank("ppio", "https://api.ppinfra.com/v3/openai")
	registerOpenAICompatRerank("perfxcloud", "https://cloud.perfxlab.cn/v1")
	registerOpenAICompatRerank("upstage", "https://api.upstage.ai/v1/solar")
	registerOpenAICompatRerank("n1n", "https://api.n1n.ai/v1")
	registerOpenAICompatRerank("avian", "https://api.avian.io/v1")
	registerOpenAICompatRerank("baichuan", "https://api.baichuan-ai.com/v1")
	registerOpenAICompatRerank("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatRerank("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatRerank("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatRerank("minimax", "https://api.minimaxi.com/v1")
	registerOpenAICompatRerank("modelscope", "")
	registerOpenAICompatRerank("youdao", "")
	registerOpenAICompatRerank("baai", "")
	registerOpenAICompatRerank("fastembed", "")
	registerOpenAICompatRerank("nomicai", "")
	registerOpenAICompatRerank("sentence_transformers", "")

	// ========== Speech2Text 提供商 ==========
	RegisterSpeech2Text("openai", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "whisper-1"
		}
		return speech2text.NewOpenAI(config.APIKey, baseURL, model), nil
	})
	RegisterSpeech2Text("azure", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return speech2text.NewAzure(config.APIKey, config.BaseURL, config.ModelName, apiVer), nil
	})
	RegisterSpeech2Text("tongyi", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		impl, err := speech2text.NewDashScopeASR(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
		if err != nil {
			return nil, err
		}
		return &speech2text.ProviderNameAdapter{Speech2Text: impl, DisplayName: "tongyi"}, nil
	})
	RegisterSpeech2Text("dashscope", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		return speech2text.NewDashScopeASR(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
	})
	RegisterSpeech2Text("tencent", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		return speech2text.NewTencentASR(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
	})
	RegisterSpeech2Text("tencentcloud", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		impl, err := speech2text.NewTencentASR(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
		if err != nil {
			return nil, err
		}
		return &speech2text.ProviderNameAdapter{Speech2Text: impl, DisplayName: "tencentcloud"}, nil
	})
	registerOpenAICompatSpeech2Text("xinference", "")
	registerOpenAICompatSpeech2Text("localai", "")
	registerOpenAICompatSpeech2Text("deepseek", "https://api.deepseek.com")
	registerOpenAICompatSpeech2Text("stepfun", "https://api.stepfun.com/v1")
	registerOpenAICompatSpeech2Text("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatSpeech2Text("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatSpeech2Text("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatSpeech2Text("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatSpeech2Text("gpustack", "")
	registerOpenAICompatSpeech2Text("nvidia", "https://integrate.api.nvidia.com/v1")
	registerOpenAICompatSpeech2Text("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatSpeech2Text("siliconflow", "https://api.siliconflow.cn/v1")
	registerOpenAICompatSpeech2Text("novitaai", "https://api.novita.ai/v3/openai")
	registerOpenAICompatSpeech2Text("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatSpeech2Text("together", "https://api.together.xyz/v1")
	registerOpenAICompatSpeech2Text("minimax", "https://api.minimaxi.com/v1")
	RegisterSpeech2Text("zhipu", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://open.bigmodel.cn/api/paas/v4"
		}
		model := config.ModelName
		if model == "" {
			model = "glm-asr"
		}
		impl := speech2text.NewOpenAI(config.APIKey, baseURL, model)
		return &speech2text.ProviderNameAdapter{Speech2Text: impl, DisplayName: "zhipu"}, nil
	})
	registerOpenAICompatSpeech2Text("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatSpeech2Text("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatSpeech2Text("xunfei", "https://spark-api-open.xf-yun.com/v1")
	registerOpenAICompatSpeech2Text("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatSpeech2Text("baichuan", "https://api.baichuan-ai.com/v1")
	registerOpenAICompatSpeech2Text("xai", "https://api.x.ai/v1")
	registerOpenAICompatSpeech2Text("grok", "https://api.x.ai/v1")
	registerOpenAICompatSpeech2Text("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatSpeech2Text("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatSpeech2Text("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatSpeech2Text("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatSpeech2Text("ai302", "https://api.302.ai/v1")
	registerOpenAICompatSpeech2Text("jiekouai", "https://api.jiekou.ai/openai")
	registerOpenAICompatSpeech2Text("vllm", "")
	registerOpenAICompatSpeech2Text("openai_api_compatible", "")
	registerOpenAICompatSpeech2Text("lmstudio", "")
	registerOpenAICompatSpeech2Text("baiduyiyan", "https://qianfan.baidubce.com/v2")
	registerOpenAICompatSpeech2Text("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatSpeech2Text("gemini", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatSpeech2Text("googlecloud", "https://generativelanguage.googleapis.com/v1beta")

	// ========== TTS 提供商 ==========
	RegisterTTS("openai", func(config TTSConfig) (interfaces.TTS, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "tts-1"
		}
		return tts.NewOpenAI(config.APIKey, baseURL, model), nil
	})
	RegisterTTS("azure", func(config TTSConfig) (interfaces.TTS, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return tts.NewAzure(config.APIKey, config.BaseURL, config.ModelName, apiVer), nil
	})
	RegisterTTS("fishaudio", func(config TTSConfig) (interfaces.TTS, error) {
		return tts.NewFishAudio(config.APIKey, config.BaseURL, config.ModelName)
	})
	RegisterTTS("tongyi", func(config TTSConfig) (interfaces.TTS, error) {
		impl, err := tts.NewDashScopeTTS(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
		if err != nil {
			return nil, err
		}
		return &tts.ProviderNameAdapter{TTS: impl, DisplayName: "tongyi"}, nil
	})
	RegisterTTS("dashscope", func(config TTSConfig) (interfaces.TTS, error) {
		return tts.NewDashScopeTTS(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
	})
	registerOpenAICompatTTS("xinference", "")
	registerOpenAICompatTTS("localai", "")
	registerOpenAICompatTTS("stepfun", "https://api.stepfun.com/v1")
	registerOpenAICompatTTS("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatTTS("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatTTS("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatTTS("gpustack", "")
	registerOpenAICompatTTS("nvidia", "https://integrate.api.nvidia.com/v1")
	registerOpenAICompatTTS("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatTTS("siliconflow", "https://api.siliconflow.cn/v1")
	registerOpenAICompatTTS("novitaai", "https://api.novita.ai/v3/openai")
	registerOpenAICompatTTS("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatTTS("together", "https://api.together.xyz/v1")
	registerOpenAICompatTTS("minimax", "https://api.minimaxi.com/v1")
	registerOpenAICompatTTS("zhipu", "https://open.bigmodel.cn/api/paas/v4")
	registerOpenAICompatTTS("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatTTS("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatTTS("xunfei", "https://spark-api-open.xf-yun.com/v1")
	registerOpenAICompatTTS("deepseek", "https://api.deepseek.com")
	registerOpenAICompatTTS("xai", "https://api.x.ai/v1")
	registerOpenAICompatTTS("grok", "https://api.x.ai/v1")
	registerOpenAICompatTTS("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatTTS("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatTTS("gemini", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatTTS("googlecloud", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatTTS("longcat", "https://api.longcat.chat/openai")
	registerOpenAICompatTTS("ppio", "https://api.ppinfra.com/v3/openai")
	registerOpenAICompatTTS("perfxcloud", "https://cloud.perfxlab.cn/v1")
	registerOpenAICompatTTS("upstage", "https://api.upstage.ai/v1/solar")
	registerOpenAICompatTTS("n1n", "https://api.n1n.ai/v1")
	registerOpenAICompatTTS("avian", "https://api.avian.io/v1")
	registerOpenAICompatTTS("baichuan", "https://api.baichuan-ai.com/v1")
	registerOpenAICompatTTS("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatTTS("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatTTS("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatTTS("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatTTS("ai302", "https://api.302.ai/v1")
	registerOpenAICompatTTS("jiekouai", "https://api.jiekou.ai/openai")
	registerOpenAICompatTTS("vllm", "")
	registerOpenAICompatTTS("openai_api_compatible", "")
	registerOpenAICompatTTS("lmstudio", "")
	registerOpenAICompatTTS("baiduyiyan", "https://qianfan.baidubce.com/v2")

	// ========== OCR 提供商 ==========
	RegisterOCR("ollama", func(config OCRConfig) (interfaces.OCR, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "llava"
		}
		return ocr.NewOpenAI("", baseURL, model), nil
	})
	RegisterOCR("openai", func(config OCRConfig) (interfaces.OCR, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return ocr.NewOpenAI(config.APIKey, baseURL, model), nil
	})
	RegisterOCR("azure", func(config OCRConfig) (interfaces.OCR, error) {
		apiKey := cv.ParseAzureAPIKey(config.APIKey)
		apiVer := strFromExtra(config.Extra, "api_version")
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return ocr.NewAzure(apiKey, config.BaseURL, model, apiVer), nil
	})
	RegisterOCR("anthropic", func(config OCRConfig) (interfaces.OCR, error) {
		baseURL := strings.TrimSpace(config.BaseURL)
		if baseURL == "" {
			baseURL = "https://api.anthropic.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
		return ocr.NewAnthropic(config.APIKey, baseURL, model), nil
	})
	registerOpenAICompatOCR("deepseek", "https://api.deepseek.com")
	registerOpenAICompatOCR("xinference", "")
	registerOpenAICompatOCR("localai", "")
	registerOpenAICompatOCR("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatOCR("zhipu", "https://open.bigmodel.cn/api/paas/v4")
	registerOpenAICompatOCR("stepfun", "https://api.stepfun.com/v1")
	registerOpenAICompatOCR("tongyi", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatOCR("dashscope", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatOCR("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatOCR("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatOCR("novitaai", "https://api.novita.ai/v3/openai")
	registerOpenAICompatOCR("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatOCR("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatOCR("siliconflow", "https://api.siliconflow.cn/v1")
	registerOpenAICompatOCR("gpustack", "")
	registerOpenAICompatOCR("vllm", "")
	registerOpenAICompatOCR("openai_api_compatible", "")
	registerOpenAICompatOCR("nvidia", "https://integrate.api.nvidia.com/v1")
	registerOpenAICompatOCR("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatOCR("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatOCR("xunfei", "https://spark-api-open.xf-yun.com/v1")
	registerOpenAICompatOCR("huggingface", "")
	registerOpenAICompatOCR("modelscope", "")
	registerOpenAICompatOCR("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatOCR("xai", "https://api.x.ai/v1")
	registerOpenAICompatOCR("grok", "https://api.x.ai/v1")
	registerOpenAICompatOCR("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatOCR("gemini", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatOCR("googlecloud", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatOCR("longcat", "https://api.longcat.chat/openai")
	registerOpenAICompatOCR("ppio", "https://api.ppinfra.com/v3/openai")
	registerOpenAICompatOCR("perfxcloud", "https://cloud.perfxlab.cn/v1")
	registerOpenAICompatOCR("upstage", "https://api.upstage.ai/v1/solar")
	registerOpenAICompatOCR("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatOCR("n1n", "https://api.n1n.ai/v1")
	registerOpenAICompatOCR("avian", "https://api.avian.io/v1")
	registerOpenAICompatOCR("baichuan", "https://api.baichuan-ai.com/v1")
	registerOpenAICompatOCR("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatOCR("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatOCR("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatOCR("minimax", "https://api.minimaxi.com/v1")
	registerOpenAICompatOCR("ai302", "https://api.302.ai/v1")
	registerOpenAICompatOCR("jiekouai", "https://api.jiekou.ai/openai")
	registerOpenAICompatOCR("lmstudio", "")
	registerOpenAICompatOCR("baiduyiyan", "https://qianfan.baidubce.com/v2")
	RegisterOCR("mineru", func(config OCRConfig) (interfaces.OCR, error) {
		return ocr.NewMinerUOCR(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
	})
	RegisterOCR("paddleocr", func(config OCRConfig) (interfaces.OCR, error) {
		return ocr.NewPaddleOCRHTTP(config.APIKey, config.BaseURL, config.ModelName, config.Extra)
	})

	// ========== CV 提供商 ==========
	RegisterCV("ollama", func(config CVConfig) (interfaces.CV, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "llava"
		}
		return cv.NewOpenAI("", baseURL, model), nil
	})
	RegisterCV("openai", func(config CVConfig) (interfaces.CV, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return cv.NewOpenAI(config.APIKey, baseURL, model), nil
	})
	RegisterCV("azure", func(config CVConfig) (interfaces.CV, error) {
		apiKey := cv.ParseAzureAPIKey(config.APIKey)
		apiVer := strFromExtra(config.Extra, "api_version")
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return cv.NewAzure(apiKey, config.BaseURL, model, apiVer), nil
	})
	RegisterCV("anthropic", func(config CVConfig) (interfaces.CV, error) {
		baseURL := strings.TrimSpace(config.BaseURL)
		if baseURL == "" {
			baseURL = "https://api.anthropic.com/v1"
		}
		model := config.ModelName
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
		return cv.NewAnthropic(config.APIKey, baseURL, model), nil
	})
	RegisterCV("volcengine", func(config CVConfig) (interfaces.CV, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://ark.cn-beijing.volces.com/api/v3"
		}
		apiKey := config.APIKey
		model := config.ModelName
		if apiKey != "" && (apiKey[0] == '{' || strings.Contains(apiKey, "ark_api_key")) {
			arkKey, epID, epEndID := parseVolcEngineKey(apiKey)
			apiKey = arkKey
			if model == "" && epID != "" {
				model = epID + epEndID
			}
		}
		if model == "" {
			model = strFromExtra(config.Extra, "ep_id") + strFromExtra(config.Extra, "endpoint_id")
		}
		if apiKey == "" {
			apiKey = strFromExtra(config.Extra, "ark_api_key")
		}
		if model == "" {
			model = "gpt-4o"
		}
		impl := cv.NewOpenAI(apiKey, baseURL, model)
		return &cv.ProviderNameAdapter{CV: impl, DisplayName: "volcengine"}, nil
	})
	registerOpenAICompatCV("xinference", "")
	registerOpenAICompatCV("localai", "")
	registerOpenAICompatCV("deepseek", "https://api.deepseek.com")
	registerOpenAICompatCV("moonshot", "https://api.moonshot.cn/v1")
	registerOpenAICompatCV("zhipu", "https://open.bigmodel.cn/api/paas/v4")
	registerOpenAICompatCV("stepfun", "https://api.stepfun.com/v1")
	registerOpenAICompatCV("hunyuan", "https://api.hunyuan.cloud.tencent.com/v1")
	registerOpenAICompatCV("lingyi", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatCV("siliconflow", "https://api.siliconflow.cn/v1")
	registerOpenAICompatCV("openrouter", "https://openrouter.ai/api/v1")
	registerOpenAICompatCV("tongyi", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatCV("dashscope", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	registerOpenAICompatCV("cometapi", "https://api.cometapi.com/v1")
	registerOpenAICompatCV("giteeai", "https://ai.gitee.com/v1/")
	registerOpenAICompatCV("novitaai", "https://api.novita.ai/v3/openai")
	registerOpenAICompatCV("groq", "https://api.groq.com/openai/v1")
	registerOpenAICompatCV("baiduyiyan", "https://qianfan.baidubce.com/v2")
	registerOpenAICompatCV("gemini", "https://generativelanguage.googleapis.com/v1beta")
	registerOpenAICompatCV("xai", "https://api.x.ai/v1")
	registerOpenAICompatCV("grok", "https://api.x.ai/v1")
	registerOpenAICompatCV("gpustack", "")
	registerOpenAICompatCV("vllm", "")
	registerOpenAICompatCV("openai_api_compatible", "")
	registerOpenAICompatCV("nvidia", "https://integrate.api.nvidia.com/v1")
	registerOpenAICompatCV("tokenpony", "https://ragflow.vip-api.tokenpony.cn/v1")
	registerOpenAICompatCV("ragcon", "https://connect.ragcon.com/v1")
	registerOpenAICompatCV("xunfei", "https://spark-api-open.xf-yun.com/v1")
	registerOpenAICompatCV("huggingface", "")
	registerOpenAICompatCV("modelscope", "")
	registerOpenAICompatCV("lmstudio", "")
	registerOpenAICompatCV("together", "https://api.together.xyz/v1")
	registerOpenAICompatCV("minimax", "https://api.minimaxi.com/v1")
	registerOpenAICompatCV("01ai", "https://api.lingyiwanwu.com/v1")
	registerOpenAICompatCV("deepinfra", "https://api.deepinfra.com/v1/openai")
	registerOpenAICompatCV("longcat", "https://api.longcat.chat/openai")
	registerOpenAICompatCV("ppio", "https://api.ppinfra.com/v3/openai")
	registerOpenAICompatCV("perfxcloud", "https://cloud.perfxlab.cn/v1")
	registerOpenAICompatCV("upstage", "https://api.upstage.ai/v1/solar")
	registerOpenAICompatCV("deerapi", "https://api.deerapi.com/v1")
	registerOpenAICompatCV("n1n", "https://api.n1n.ai/v1")
	registerOpenAICompatCV("avian", "https://api.avian.io/v1")
	registerOpenAICompatCV("baichuan", "https://api.baichuan-ai.com/v1")
	registerOpenAICompatCV("mistral", "https://api.mistral.ai/v1")
	registerOpenAICompatCV("ai302", "https://api.302.ai/v1")
	registerOpenAICompatCV("jiekouai", "https://api.jiekou.ai/openai")
	registerOpenAICompatCV("googlecloud", "https://generativelanguage.googleapis.com/v1beta")
	// 上游 local 工厂无默认实现；此处提供 OpenAI 兼容视觉接口，需配置 Base URL（如本地 vLLM）
	registerOpenAICompatCV("local", "")
	RegisterCV("leptonai", func(config CVConfig) (interfaces.CV, error) {
		baseURL := strings.TrimSpace(config.BaseURL)
		if baseURL == "" {
			m := strings.TrimSpace(config.ModelName)
			if m == "" {
				return nil, fmt.Errorf("leptonai: 请填写 Base URL，或填写模型名以使用 https://{模型名}.lepton.run/api/v1")
			}
			baseURL = fmt.Sprintf("https://%s.lepton.run/api/v1", m)
		}
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		impl := cv.NewOpenAI(config.APIKey, baseURL, model)
		return &cv.ProviderNameAdapter{CV: impl, DisplayName: "leptonai"}, nil
	})

	registerAzureOpenAIAliases()
}

// registerAzureOpenAIAliases 与「Azure-OpenAI」工厂别名对齐，行为与 provider「azure」一致。
func registerAzureOpenAIAliases() {
	RegisterLLM("azure_openai", func(config LLMConfig) (interfaces.LLM, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return llm.NewAzure(config.APIKey, config.BaseURL, config.ModelName, apiVer), nil
	})
	RegisterEmbedding("azure_openai", func(config EmbeddingConfig) (interfaces.Embedder, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return embedding.NewAzureEmbed(config.APIKey, config.BaseURL, config.ModelName, apiVer, config.Dimensions), nil
	})
	RegisterSpeech2Text("azure_openai", func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return speech2text.NewAzure(config.APIKey, config.BaseURL, config.ModelName, apiVer), nil
	})
	RegisterTTS("azure_openai", func(config TTSConfig) (interfaces.TTS, error) {
		apiVer := strFromExtra(config.Extra, "api_version")
		return tts.NewAzure(config.APIKey, config.BaseURL, config.ModelName, apiVer), nil
	})
	RegisterOCR("azure_openai", func(config OCRConfig) (interfaces.OCR, error) {
		apiKey := cv.ParseAzureAPIKey(config.APIKey)
		apiVer := strFromExtra(config.Extra, "api_version")
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return ocr.NewAzure(apiKey, config.BaseURL, model, apiVer), nil
	})
	RegisterCV("azure_openai", func(config CVConfig) (interfaces.CV, error) {
		apiKey := cv.ParseAzureAPIKey(config.APIKey)
		apiVer := strFromExtra(config.Extra, "api_version")
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return cv.NewAzure(apiKey, config.BaseURL, model, apiVer), nil
	})
}

func registerOpenAICompatSpeech2Text(provider, defaultBaseURL string) {
	RegisterSpeech2Text(provider, func(config Speech2TextConfig) (interfaces.Speech2Text, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		model := config.ModelName
		if model == "" {
			model = "whisper-1"
		}
		impl := speech2text.NewOpenAI(config.APIKey, baseURL, model)
		return &speech2text.ProviderNameAdapter{Speech2Text: impl, DisplayName: provider}, nil
	})
}

func registerOpenAICompatTTS(provider, defaultBaseURL string) {
	RegisterTTS(provider, func(config TTSConfig) (interfaces.TTS, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		model := config.ModelName
		if model == "" {
			model = "tts-1"
		}
		impl := tts.NewOpenAI(config.APIKey, baseURL, model)
		return &tts.ProviderNameAdapter{TTS: impl, DisplayName: provider}, nil
	})
}

func registerOpenAICompatCV(provider, defaultBaseURL string) {
	RegisterCV(provider, func(config CVConfig) (interfaces.CV, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		impl := cv.NewOpenAI(config.APIKey, baseURL, model)
		return &cv.ProviderNameAdapter{CV: impl, DisplayName: provider}, nil
	})
}

func registerOpenAICompatLLM(provider, defaultBaseURL string) {
	RegisterLLM(provider, func(config LLMConfig) (interfaces.LLM, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		impl := llm.NewOpenAI(config.APIKey, baseURL, config.ModelName)
		return &llm.ProviderNameAdapter{LLM: impl, DisplayName: provider}, nil
	})
}

func registerOpenAICompatOCR(provider, defaultBaseURL string) {
	RegisterOCR(provider, func(config OCRConfig) (interfaces.OCR, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		model := config.ModelName
		if model == "" {
			model = "gpt-4o"
		}
		return ocr.NewOpenAI(config.APIKey, baseURL, model), nil
	})
}

func registerOpenAICompatEmbedding(provider, defaultBaseURL string) {
	RegisterEmbedding(provider, func(config EmbeddingConfig) (interfaces.Embedder, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		return embedding.NewOpenAICompatEmbed(provider, config.APIKey, baseURL, config.ModelName, config.Dimensions), nil
	})
}

func registerOpenAICompatRerank(provider, defaultBaseURL string) {
	RegisterRerank(provider, func(config RerankConfig) (interfaces.Reranker, error) {
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		if baseURL == "" {
			return nil, nil
		}
		return rerank.NewOpenAICompat(provider, config.APIKey, baseURL, config.ModelName), nil
	})
}
