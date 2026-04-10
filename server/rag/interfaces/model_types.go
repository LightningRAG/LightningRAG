// Package interfaces 定义 RAG 模型类型常量，参考 references 目录内 llm_factories.json
// 各类型对应不同接口：chat->LLM, embedding->Embedder, rerank->Reranker
package interfaces

// ModelType 模型类型，与 LightningRAG model_type 对应
const (
	ModelTypeChat       = "chat"        // 对话/生成，对应 LLM 接口
	ModelTypeEmbedding  = "embedding"   // 文本嵌入，对应 Embedder 接口
	ModelTypeRerank     = "rerank"      // 重排序，对应 Reranker 接口
	ModelTypeSpeech2Txt = "speech2text" // 语音转文字
	ModelTypeTTS        = "tts"         // 文字转语音
	ModelTypeOCR        = "ocr"         // 光学字符识别
	ModelTypeCV         = "cv"          // 计算机视觉
)

// AllModelTypes 所有支持的模型类型
var AllModelTypes = []string{
	ModelTypeChat,
	ModelTypeEmbedding,
	ModelTypeRerank,
	ModelTypeSpeech2Txt,
	ModelTypeTTS,
	ModelTypeOCR,
	ModelTypeCV,
}
