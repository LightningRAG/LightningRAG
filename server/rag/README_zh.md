# LightningRAG 后端 RAG 模块

**English:** [README.md](./README.md)

本模块将 LightningRAG 升级为 RAG（Retrieval-augmented Generation）项目，在 langchaingo 等实践基础上扩展，并与 `references` 目录内上游参考设计对齐，支持模块化和可扩展架构。

## 架构概览

### 核心接口 (interfaces/)

- **LLM**: 大语言模型接口，支持 `GenerateContent` 和 `Call`
- **Embedder**: 向量嵌入接口，支持 `EmbedDocuments` 和 `EmbedQuery`
- **Reranker**: 重排序接口，支持 `Rerank`，对检索结果按相关性排序
- **VectorStore**: 向量存储接口，支持 `AddDocuments`、`SimilaritySearch`、`DeleteByIDs`、`DeleteByNamespace`
- **Retriever**: 检索器接口，支持 `GetRelevantDocuments`，类型包括 vector、pageindex、keyword

### 模型类型 (model_types.go)

本模块支持多种模型类型：`chat`、`embedding`、`rerank`、`speech2text`、`tts`、`ocr`、`cv`（命名与 `references` 目录内 model_type 约定一致）

### Speech2Text / TTS / OCR / CV

**Speech2Text（语音转文字）**
- OpenAI (openai)、Xinference、LocalAI、DeepSeek、StepFun、CometAPI、DeerAPI、GiteeAI、DeepInfra

**TTS（文字转语音）**
- OpenAI (openai)、Xinference、LocalAI、StepFun、CometAPI、DeerAPI、DeepInfra

**OCR（光学字符识别）**
- OpenAI (openai)、DeepSeek、Xinference、LocalAI、Moonshot、ZHIPU、StepFun

**CV（计算机视觉）**
- OpenAI (openai)、Xinference、LocalAI、DeepSeek、Moonshot、ZHIPU、StepFun、腾讯混元 (hunyuan)、零一万物 (lingyi)、硅基流动 (siliconflow)、OpenRouter

### 数据模型 (model/rag/)

| 表名 | 说明 |
|------|------|
| rag_knowledge_bases | 知识库 |
| rag_documents | 知识库内的文档 |
| rag_chunks | 文档切片（元数据存关系库，向量存向量库） |
| rag_llm_providers | 管理员配置的大模型 |
| rag_embedding_providers | 嵌入模型配置 |
| rag_vector_store_configs | 向量存储配置 |
| rag_user_llms | 用户自定义模型 |
| rag_knowledge_base_shares | 知识库分享/转让 |
| rag_conversations | 对话会话 |
| rag_messages | 对话消息 |

### 已实现提供商

**LLM (Chat)**
- OpenAI (openai)、Ollama (ollama)、Azure (azure)、Anthropic (anthropic)
- 国内：通义 (tongyi/dashscope)、StepFun (stepfun)、MiniMax (minimax)、腾讯混元 (hunyuan)、零一万物 (lingyi/01ai)、302.AI (ai302)、接口AI (jiekouai)、GiteeAI (giteeai)、CometAPI (cometapi)、NovitaAI (novitaai)、智谱 (zhipu)、月之暗面 (moonshot)、硅基流动 (siliconflow)
- 国际：DeepSeek (deepseek)、DeepInfra (deepinfra)、OpenRouter (openrouter)、Groq (groq)、Together (together)、LongCat (longcat)、PPIO (ppio)、PerfXCloud (perfxcloud)、Upstage (upstage)、DeerAPI (deerapi)、n1n (n1n)、Avian (avian)
- 本地：Xinference (xinference)、LocalAI (localai)

**Embedding**
- OpenAI (openai)、Ollama (ollama)
- 国内：通义 (tongyi/dashscope)、StepFun (stepfun)、智谱 (zhipu)、GiteeAI (giteeai)、CometAPI (cometapi)、NovitaAI (novitaai)、月之暗面 (moonshot)、硅基流动 (siliconflow)、火山引擎 (volcengine)
- 国际：DeepSeek (deepseek)、DeepInfra (deepinfra)、Voyage AI (voyageai)、Jina (jina)、Cohere (cohere)、NVIDIA (nvidia)
- 本地：Xinference (xinference)、LocalAI (localai)

**Rerank**
- Jina (jina)，默认模型 jina-reranker-v2-base-multilingual
- SiliconFlow (siliconflow)，默认 BAAI/bge-reranker-v2-m3
- NovitaAI (novitaai)、GiteeAI (giteeai)
- Cohere (cohere)、通义 (tongyi)
- Voyage AI (voyageai)、NVIDIA (nvidia)
- 302.AI (ai302)、接口AI (jiekouai)
- Xinference (xinference)、LocalAI (localai)，需配置 baseURL

**VectorStore**（基于 interface，可扩展更多向量库）
- PostgreSQL + pgvector (postgresql)，需数据库启用 `CREATE EXTENSION vector`
- Elasticsearch (elasticsearch)，支持 8.x，dense_vector + cosine 相似度，Config 示例：`{"address":"http://localhost:9200","username":"","password":"","index_prefix":"rag_vectors"}`

### 注册表 (registry/)

- `RegisterLLM(provider, factory)` / `CreateLLM(config)` 
- `RegisterEmbedding(provider, factory)` / `CreateEmbedding(config)`
- `RegisterVectorStore(provider, factory)` / `CreateVectorStore(config, embedder, namespace, vectorDims)`
- `RegisterRerank(provider, factory)` / `CreateRerank(config)`
- `RegisterSpeech2Text(provider, factory)` / `CreateSpeech2Text(config)`
- `RegisterTTS(provider, factory)` / `CreateTTS(config)`
- `RegisterOCR(provider, factory)` / `CreateOCR(config)`
- `RegisterCV(provider, factory)` / `CreateCV(config)`

新提供商只需实现对应接口并注册即可接入。

## API 路由

所有 RAG 接口需 JWT 鉴权。

### 知识库 (rag/knowledgeBase)
- POST create - 创建知识库
- POST list - 列表
- POST get - 详情
- POST update - 更新
- POST delete - 删除
- POST uploadDocument - 上传文档（form: knowledgeBaseId, file）
- POST share - 分享
- POST transfer - 转让

### 对话 (rag/conversation)
- POST create - 创建对话
- POST chat - 发送消息（非流式）
- POST chatStream - 流式 SSE（末帧可含 `retrievalMode`、`retrievalQuery`、`searchQuery`、`references`）
- POST queryData - 纯检索结构化结果（对齐 LightningRAG `/query/data`，无 LLM、不写消息；需 Casbin 注册 `/rag/conversation/queryData`）
- POST list - 列表
- POST get - 详情
- POST listMessages / listTools / update / delete

可选请求体字段（与 `references/LightRAG` QueryRequest 相近）：`queryMode`、`chunkTopK`、`topK`、`enableRerank`、`hlKeywords`、`llKeywords`、`conversationHistory`、`maxTotalTokens`、`maxRagContextTokens`、`responseLanguage`、`includeReferences`、`includeChunkContent`、`cosineThreshold`、`minRerankScore` 等，详见 `server/model/rag/request/conversation.go`。

### 模型 (rag/llm)
- POST listProviders - 列出可用模型（管理员+用户）
- POST listUserModels - 用户自定义模型列表
- POST addUserModel - 添加用户模型
- POST deleteUserModel - 删除用户模型

## 后续扩展建议

1. **文档解析与切片**：接入 PDF、Word、Markdown 等解析，实现文本分块（可参考 langchaingo textsplitter）
2. **向量化流水线**：文档上传后异步解析→切片→向量化→写入向量库
3. **检索增强对话**：Chat 时根据 SourceType/SourceIDs 从知识库检索相关切片，拼入 prompt
4. **PageIndex 检索**：实现 `RetrieverTypePageIndex` 的 Retriever，支持树状推理检索
5. ~~**Elasticsearch 向量存储**~~：已实现，支持 kNN 向量检索
6. **模型共享**：完善 RagLLMProvider 的 ShareScope、ShareTarget，按角色/组织过滤
7. **Embedding 提供商**：已支持 Ollama、DeepSeek、Xinference、LocalAI、SiliconFlow 等

## 配置说明

- 使用 PostgreSQL 时需安装 pgvector 扩展
- 管理员在 `rag_llm_providers`、`rag_embedding_providers`、`rag_vector_store_configs` 中配置
- 用户可在「添加用户模型」中配置自己的 API Key
- **全局 RAG 默认值与运维**：根配置 **`server/config.yaml`（或各环境对应文件）中的 `rag:` 段** 可统一调整对话/知识库检索的 TopK 与候选池、混合检索权重与分数阈值、向量相似度下限、知识图谱抽取与检索参数、以及 **公开 Webhook 渠道** 的限流与出站重试队列等；多数数值为 **`0` 表示使用代码内置默认**。项目级中文总览见仓库根目录 **[README_zh.md](../../README_zh.md)** 第六节。
