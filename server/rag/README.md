# LightningRAG backend RAG module

**简体中文：** [README_zh.md](./README_zh.md)

This module turns LightningRAG into a RAG (retrieval-augmented generation) stack, aligned with patterns under `references`, with a modular and extensible layout.

## Architecture

### Core interfaces (`interfaces/`)

- **LLM:** `GenerateContent`, `Call`
- **Embedder:** `EmbedDocuments`, `EmbedQuery`
- **Reranker:** `Rerank` for reordering hits
- **VectorStore:** `AddDocuments`, `SimilaritySearch`, `DeleteByIDs`, `DeleteByNamespace`
- **Retriever:** `GetRelevantDocuments` — vector, pageindex, keyword

### Model kinds (`model_types.go`)

Supported kinds: `chat`, `embedding`, `rerank`, `speech2text`, `tts`, `ocr`, `cv` (same naming as `references`).

### Speech2Text / TTS / OCR / CV

- **Speech2Text:** OpenAI, Xinference, LocalAI, DeepSeek, StepFun, CometAPI, DeerAPI, GiteeAI, DeepInfra  
- **TTS:** OpenAI, Xinference, LocalAI, StepFun, CometAPI, DeerAPI, DeepInfra  
- **OCR:** OpenAI, DeepSeek, Xinference, LocalAI, Moonshot, ZHIPU, StepFun  
- **CV:** OpenAI, Xinference, LocalAI, DeepSeek, Moonshot, ZHIPU, StepFun, Tencent Hunyuan, Lingyi, SiliconFlow, OpenRouter

### Tables (`model/rag/`)

| Table | Purpose |
|-------|---------|
| rag_knowledge_bases | Knowledge bases |
| rag_documents | Documents in a KB |
| rag_chunks | Chunks (metadata in SQL, vectors in vector store) |
| rag_llm_providers | Admin-configured LLMs |
| rag_embedding_providers | Embedding configs |
| rag_vector_store_configs | Vector DB configs |
| rag_user_llms | User-defined models |
| rag_knowledge_base_shares | Share / transfer |
| rag_conversations | Chat sessions |
| rag_messages | Chat messages |

### Providers (summary)

**LLM:** OpenAI, Ollama, Azure, Anthropic; CN: Tongyi/DashScope, StepFun, MiniMax, Hunyuan, Lingyi/01ai, 302.AI, JiekouAI, GiteeAI, CometAPI, NovitaAI, ZHIPU, Moonshot, SiliconFlow; global: DeepSeek, DeepInfra, OpenRouter, Groq, Together, LongCat, PPIO, PerfXCloud, Upstage, DeerAPI, n1n, Avian; local: Xinference, LocalAI.

**Embedding:** OpenAI, Ollama; CN: Tongyi, StepFun, ZHIPU, GiteeAI, CometAPI, NovitaAI, Moonshot, SiliconFlow, Volcengine; global: DeepSeek, DeepInfra, Voyage, Jina, Cohere, NVIDIA; local: Xinference, LocalAI.

**Rerank:** Jina (default `jina-reranker-v2-base-multilingual`), SiliconFlow (`BAAI/bge-reranker-v2-m3`), NovitaAI, GiteeAI, Cohere, Tongyi, Voyage, NVIDIA, 302.AI, JiekouAI, Xinference, LocalAI (configure `baseURL`).

**VectorStore:** PostgreSQL + pgvector (`CREATE EXTENSION vector`); Elasticsearch 8.x with `dense_vector` + cosine; config example: `{"address":"http://localhost:9200","username":"","password":"","index_prefix":"rag_vectors"}`.

### Registry (`registry/`)

`RegisterLLM` / `CreateLLM`, `RegisterEmbedding` / `CreateEmbedding`, `RegisterVectorStore` / `CreateVectorStore`, `RegisterRerank` / `CreateRerank`, `RegisterSpeech2Text`, `RegisterTTS`, `RegisterOCR`, `RegisterCV`.

New vendors implement the interface and register.

## HTTP API

RAG routes require JWT.

### Knowledge base (`rag/knowledgeBase`)

create, list, get, update, delete, uploadDocument (form: knowledgeBaseId, file), share, transfer.

### Conversation (`rag/conversation`)

create, chat, chatStream (SSE; last frame may include `retrievalMode`, `retrievalQuery`, `searchQuery`, `references`), queryData (retrieval-only, Casbin `/rag/conversation/queryData`), list, get, listMessages, listTools, update, delete.

Optional body fields (similar to `references/LightRAG` `QueryRequest`): `queryMode`, `chunkTopK`, `topK`, `enableRerank`, `hlKeywords`, `llKeywords`, `conversationHistory`, `maxTotalTokens`, `maxRagContextTokens`, `responseLanguage`, `includeReferences`, `includeChunkContent`, `cosineThreshold`, `minRerankScore`, etc. — see `server/model/rag/request/conversation.go`.

### Models (`rag/llm`)

listProviders, listUserModels, addUserModel, deleteUserModel.

## Possible extensions

1. Richer parsing and chunking (PDF, Word, Markdown; langchaingo textsplitters).  
2. Async pipeline: upload → parse → chunk → embed → vector store.  
3. RAG chat: retrieve by `SourceType` / `SourceIDs`, inject into prompt.  
4. PageIndex retriever: `RetrieverTypePageIndex` tree reasoning.  
5. ~~Elasticsearch vector store~~ — done.  
6. Model sharing: `ShareScope`, `ShareTarget`, role/org filters.  
7. More embedding providers — many already listed above.

## Configuration

- Enable **pgvector** on PostgreSQL when used.  
- Admins configure `rag_llm_providers`, `rag_embedding_providers`, `rag_vector_store_configs`.  
- Users can add models under “user models” with their own API keys.  
- **Global defaults & ops**: the **`rag:`** block in **`server/config.yaml`** (or environment-specific YAML) centralizes conversation/KB top-k and candidate pools, hybrid fusion weights and score floors, cosine threshold, knowledge-graph tuning, and public **channel webhook** rate limits / outbound retry queues; **`0` usually means built-in defaults**. High-level overview: root **[README.md](../../README.md)** §6.
