package config

// Rag 可选 RAG 默认参数（借鉴 references/LightRAG 中 TOP_K / CHUNK_TOP_K 等集中配置思想）
// 零值表示使用代码内硬编码默认，与旧行为一致
type Rag struct {
	// DefaultConversationChunkTopK 对话单轮注入知识库的默认切片条数（对应 LightningRAG chunk_top_k 量级，默认由代码常量给出）
	DefaultConversationChunkTopK int `mapstructure:"default-conversation-chunk-top-k" json:"default-conversation-chunk-top-k" yaml:"default-conversation-chunk-top-k"`
	// DefaultConversationRetrievePoolTopK 对话、queryData、Agent/Canvas 检索未传 topK 时，将检索候选池扩至此值（仅当大于 chunkTopK 时生效，且受 max-retrieve-candidate-top-k 限制）。0 表示与 references/ragflow 一致使用宽召回（见 EffectiveDefaultConversationRetrievePoolTopK）；设为 1 且不大于 chunkTopK 时可近似关闭扩展
	DefaultConversationRetrievePoolTopK int `mapstructure:"default-conversation-retrieve-pool-top-k" json:"default-conversation-retrieve-pool-top-k" yaml:"default-conversation-retrieve-pool-top-k"`
	// DefaultKnowledgeBaseRetrieveTopN 「文档检索」接口默认返回条数（与 LightRAG QueryParam.chunk_top_k 同量级概念；上游默认 20，本仓库历史默认 8）
	DefaultKnowledgeBaseRetrieveTopN int `mapstructure:"default-knowledge-base-retrieve-top-n" json:"default-knowledge-base-retrieve-top-n" yaml:"default-knowledge-base-retrieve-top-n"`
	// DefaultKnowledgeBaseRetrievePoolTopK 「文档检索」请求未传 topK 时，将向量/融合检索候选池扩至此值（仅当大于最终 topN 时生效，且受 max-retrieve-candidate-top-k 限制）。0 表示 Ragflow 式宽召回默认（与 DefaultConversationRetrievePoolTopK 语义一致）
	DefaultKnowledgeBaseRetrievePoolTopK int `mapstructure:"default-knowledge-base-retrieve-pool-top-k" json:"default-knowledge-base-retrieve-pool-top-k" yaml:"default-knowledge-base-retrieve-pool-top-k"`
	// MaxRetrieveTopN 最终返回/注入上下文的切片条数全局上限（0 表示 50）；与向量索引侧宽召回上限 max-retrieve-candidate-top-k 解耦，对齐 Ragflow 的 page_size vs topk=1024
	MaxRetrieveTopN int `mapstructure:"max-retrieve-top-n" json:"max-retrieve-top-n" yaml:"max-retrieve-top-n"`
	// MaxRetrieveCandidateTopK 向量/融合/Rerank 前候选池条数上限（0 表示 1024，与 references/ragflow/rag/nlp/search.py Dealer.retrieval top=1024 同量级）；不超过 lightragconst.MaxSimilarityFetchK
	MaxRetrieveCandidateTopK int `mapstructure:"max-retrieve-candidate-top-k" json:"max-retrieve-candidate-top-k" yaml:"max-retrieve-candidate-top-k"`
	// HybridFusionTermWeight hybrid/mix（无图谱）全文路在融合分中的权重；与 HybridFusionVectorWeight 均为 0 时使用 Ragflow 默认 0.05 / 0.95
	HybridFusionTermWeight float64 `mapstructure:"hybrid-fusion-term-weight" json:"hybrid-fusion-term-weight" yaml:"hybrid-fusion-term-weight"`
	// HybridFusionVectorWeight hybrid/mix 向量路权重；均为 0 时用 0.05/0.95
	HybridFusionVectorWeight float64 `mapstructure:"hybrid-fusion-vector-weight" json:"hybrid-fusion-vector-weight" yaml:"hybrid-fusion-vector-weight"`
	// HybridFusionMinScore 融合分下限（各路 min-max 归一化后的加权和），0 表示不过滤；0.2 近似 Ragflow retrieval similarity_threshold 对融合分的裁剪强度（尺度仍因存储实现而异）
	HybridFusionMinScore float64 `mapstructure:"hybrid-fusion-min-score" json:"hybrid-fusion-min-score" yaml:"hybrid-fusion-min-score"`
	// HybridFusionSkipEmptyRetry true 时关闭 hybrid/mix 零命中后的纯向量宽召回重试（Ragflow search 第二次放宽）
	HybridFusionSkipEmptyRetry bool `mapstructure:"hybrid-fusion-skip-empty-retry" json:"hybrid-fusion-skip-empty-retry" yaml:"hybrid-fusion-skip-empty-retry"`
	// VectorSkipEmptyRetry true 时关闭纯 vector/global（无图谱）及 PageIndex 向量腿在零命中后的无阈值宽池重试
	VectorSkipEmptyRetry bool `mapstructure:"vector-skip-empty-retry" json:"vector-skip-empty-retry" yaml:"vector-skip-empty-retry"`
	// KeywordSkipEmptyRetry true 时关闭 keyword/local（无图谱）在零命中后的宽松全文重试（各存储 RelaxedKeywordSearch）
	KeywordSkipEmptyRetry bool `mapstructure:"keyword-skip-empty-retry" json:"keyword-skip-empty-retry" yaml:"keyword-skip-empty-retry"`
	// ElasticsearchScoreRankBoostWeight 各向量存储检索：score *= 1 + weight*clamp(rank_boost,0,1)；rank_boost 来自切片 metadata；0 关闭（配置名历史保留，PG/MySQL 同样生效）
	ElasticsearchScoreRankBoostWeight float64 `mapstructure:"elasticsearch-score-rank-boost-weight" json:"elasticsearch-score-rank-boost-weight" yaml:"elasticsearch-score-rank-boost-weight"`
	// DefaultChunkRankBoost 索引入库时写入 metadata.rank_boost（0~1），仅当该键尚未存在且未启用按位置衰减时填充；0 表示不写常量默认
	DefaultChunkRankBoost float64 `mapstructure:"default-chunk-rank-boost" json:"default-chunk-rank-boost" yaml:"default-chunk-rank-boost"`
	// ChunkRankBoostByPosition true 时按切片在文档内序号生成 rank_boost（首块 1、末块趋近 chunk-rank-boost-position-floor），与 default-chunk-rank-boost 二选一优先位置；图谱实体/关系单条向量视为 1 段
	ChunkRankBoostByPosition bool `mapstructure:"chunk-rank-boost-by-position" json:"chunk-rank-boost-by-position" yaml:"chunk-rank-boost-by-position"`
	// ChunkRankBoostPositionFloor 位置衰减时下限（0~1）；0 表示内置 0.35
	ChunkRankBoostPositionFloor float64 `mapstructure:"chunk-rank-boost-position-floor" json:"chunk-rank-boost-position-floor" yaml:"chunk-rank-boost-position-floor"`
	// AutoExtractQueryKeywords 为 true 且请求未带 hl/ll 关键词时，用 LLM 抽取高层/低层关键词（对齐 references/LightRAG extract_keywords_only；会额外消耗一次模型调用）
	AutoExtractQueryKeywords bool `mapstructure:"auto-extract-query-keywords" json:"auto-extract-query-keywords" yaml:"auto-extract-query-keywords"`
	// KeywordExtractCacheTTLSeconds 抽词结果缓存秒数，0 表示关闭（与 LightRAG 对 LLM 响应做 KV 缓存的思路一致，复用全局 BlackCache）
	KeywordExtractCacheTTLSeconds int `mapstructure:"keyword-extract-cache-ttl-seconds" json:"keyword-extract-cache-ttl-seconds" yaml:"keyword-extract-cache-ttl-seconds"`
	// KgExtractLLMCacheTTLSeconds 知识图谱按批次抽取时缓存 LLM 原始输出秒数，0 关闭；相同文本批次重复索引时可减少调用（提示词变更需清缓存或调大版本常量）
	KgExtractLLMCacheTTLSeconds int `mapstructure:"kg-extract-llm-cache-ttl-seconds" json:"kg-extract-llm-cache-ttl-seconds" yaml:"kg-extract-llm-cache-ttl-seconds"`
	// KgExtractMaxGleaning 首轮抽取成功后是否再做一轮补抽（对齐 references/LightRAG entity_extract_max_gleaning；>0 启用，默认 0 避免加倍 LLM 成本）
	KgExtractMaxGleaning int `mapstructure:"kg-extract-max-gleaning" json:"kg-extract-max-gleaning" yaml:"kg-extract-max-gleaning"`
	// KgExtractGleaningMaxInputTokens 补抽请求的粗算输入 token 上限（system+两轮 user+首轮 JSON）；超出则跳过补抽并打日志。0 表示使用内置默认 20480（与 LightRAG DEFAULT_MAX_EXTRACT_INPUT_TOKENS 同量级）
	KgExtractGleaningMaxInputTokens int `mapstructure:"kg-extract-gleaning-max-input-tokens" json:"kg-extract-gleaning-max-input-tokens" yaml:"kg-extract-gleaning-max-input-tokens"`
	// RetrieveCacheTTLSeconds 向量/融合检索结果缓存秒数，0 关闭；开启后 TTL 内可能返回略旧结果，适合重复问句降延迟（键含 userId 与知识库集合）
	RetrieveCacheTTLSeconds int `mapstructure:"retrieve-cache-ttl-seconds" json:"retrieve-cache-ttl-seconds" yaml:"retrieve-cache-ttl-seconds"`
	// KgEntityPresenceCacheTTLSeconds 「知识库是否已有图谱实体」阳性结果缓存秒数，0 关闭；仅缓存 true，删除/瘦身图谱时会主动失效
	KgEntityPresenceCacheTTLSeconds int `mapstructure:"kg-entity-presence-cache-ttl-seconds" json:"kg-entity-presence-cache-ttl-seconds" yaml:"kg-entity-presence-cache-ttl-seconds"`
	// DefaultMaxEntityContextTokens 未传 maxEntityTokens 时的默认图谱实体摘要 token 预算（0 表示不注入，与旧行为一致）
	DefaultMaxEntityContextTokens int `mapstructure:"default-max-entity-context-tokens" json:"default-max-entity-context-tokens" yaml:"default-max-entity-context-tokens"`
	// DefaultMaxRelationContextTokens 未传 maxRelationTokens 时的默认关系摘要 token 预算（0 表示不注入）
	DefaultMaxRelationContextTokens int `mapstructure:"default-max-relation-context-tokens" json:"default-max-relation-context-tokens" yaml:"default-max-relation-context-tokens"`
	// DefaultMaxRagContextTokens 未传 maxRagContextTokens 或非正时的默认「切片正文」token 粗算上限（0 表示不裁剪，与旧行为一致）
	DefaultMaxRagContextTokens int `mapstructure:"default-max-rag-context-tokens" json:"default-max-rag-context-tokens" yaml:"default-max-rag-context-tokens"`
	// ConversationHistoryMaxMessages 从 DB 加载的最近消息条数上限（单条 role 计 1）；0 表示使用内置默认 20，最大 200
	ConversationHistoryMaxMessages int `mapstructure:"conversation-history-max-messages" json:"conversation-history-max-messages" yaml:"conversation-history-max-messages"`
	// KgPromptNeighborRelLimit 构建图谱注入上下文时，在切片直连实体/关系之外，再拉取至多 N 条与当前实体集相邻的关系（一跳，0 表示关闭，与旧行为一致）
	KgPromptNeighborRelLimit int `mapstructure:"kg-prompt-neighbor-rel-limit" json:"kg-prompt-neighbor-rel-limit" yaml:"kg-prompt-neighbor-rel-limit"`
	// DefaultCosineThreshold 请求未传 cosineThreshold 时使用的向量相似度下限（0~1）；0 表示不施加服务端默认（保持旧行为）。借鉴 LightRAG constants.DEFAULT_COSINE_THRESHOLD=0.2
	DefaultCosineThreshold float64 `mapstructure:"default-cosine-threshold" json:"default-cosine-threshold" yaml:"default-cosine-threshold"`
	// KgEntityNameMaxRunes 写入图谱前实体名最大字符数（runes）；0 表示使用内置 256（对齐 LightRAG DEFAULT_ENTITY_NAME_MAX_LENGTH）；最大 512 与 DB 字段一致
	KgEntityNameMaxRunes int `mapstructure:"kg-entity-name-max-runes" json:"kg-entity-name-max-runes" yaml:"kg-entity-name-max-runes"`
	// KgStoredDescriptionMaxRunes 实体/关系 description 入库前上限（合并后截断）；0 表示内置 16384，防止描述无限膨胀影响向量与检索
	KgStoredDescriptionMaxRunes int `mapstructure:"kg-stored-description-max-runes" json:"kg-stored-description-max-runes" yaml:"kg-stored-description-max-runes"`
	// KgStoredKeywordsMaxRunes 关系 keywords 字段入库前上限；0 表示内置 4096
	KgStoredKeywordsMaxRunes int `mapstructure:"kg-stored-keywords-max-runes" json:"kg-stored-keywords-max-runes" yaml:"kg-stored-keywords-max-runes"`
	// KgMaxChunksPerEntity 单实体最多关联多少个切片（0=不限制）。对齐 LightRAG max_source_ids_per_entity（默认 300 可在 yaml 中配置）
	KgMaxChunksPerEntity int `mapstructure:"kg-max-chunks-per-entity" json:"kg-max-chunks-per-entity" yaml:"kg-max-chunks-per-entity"`
	// KgMaxChunksPerRelationship 单关系最多关联多少个切片（0=不限制）。对齐 LightRAG max_source_ids_per_relation
	KgMaxChunksPerRelationship int `mapstructure:"kg-max-chunks-per-relationship" json:"kg-max-chunks-per-relationship" yaml:"kg-max-chunks-per-relationship"`
	// KgChunkLinkLimitMethod 超出上限时保留策略：fifo 保留「按 rag_chunks.id 较新」的关联（与 LightRAG SOURCE_IDS_LIMIT_METHOD_FIFO 一致）；keep 保留较旧。空字符串视为 fifo
	KgChunkLinkLimitMethod string `mapstructure:"kg-chunk-link-limit-method" json:"kg-chunk-link-limit-method" yaml:"kg-chunk-link-limit-method"`
	// KgMergeSummarizeMinSegments 实体或关系 description 经换行合并后，非空段数 ≥ 此值时用 LLM 压成一段摘要（对齐 references/LightRAG operate._summarize_descriptions / force_llm_summary_on_merge）；0 关闭（默认）
	KgMergeSummarizeMinSegments int `mapstructure:"kg-merge-summarize-min-segments" json:"kg-merge-summarize-min-segments" yaml:"kg-merge-summarize-min-segments"`
	// KgMergeSummaryTargetRunes 摘要提示中建议的输出长度上限（runes）；0 表示内置 1200
	KgMergeSummaryTargetRunes int `mapstructure:"kg-merge-summary-target-runes" json:"kg-merge-summary-target-runes" yaml:"kg-merge-summary-target-runes"`
	// KgMergeSummarizeLLMCacheTTLSeconds 实体/关系「合并描述」LLM 摘要结果缓存秒数，0 关闭；相同类型+名称+描述段列表重复抽取时可降本（键内嵌提示版本，与 kg-extract 缓存同理）
	KgMergeSummarizeLLMCacheTTLSeconds int `mapstructure:"kg-merge-summarize-llm-cache-ttl-seconds" json:"kg-merge-summarize-llm-cache-ttl-seconds" yaml:"kg-merge-summarize-llm-cache-ttl-seconds"`
	// KgRetrieverRelatedChunkNumber local/global/hybrid/mix 下图谱向量命中后，每个实体或关系最多展开多少条关联切片参与候选（按 chunk_id 升序取前 N；0=不限制，与旧行为一致）。对齐 LightRAG related_chunk_number / DEFAULT_RELATED_CHUNK_NUMBER=5
	KgRetrieverRelatedChunkNumber int `mapstructure:"kg-retriever-related-chunk-number" json:"kg-retriever-related-chunk-number" yaml:"kg-retriever-related-chunk-number"`
	// ChannelWebhookIPLimitPerMinute 公开渠道 Webhook：同一 connectorId + 客户端 IP 在滑动 1 分钟窗口内最大请求数；0 表示不限制。依赖 Redis；Redis 不可用时中间件放行（fail-open）
	ChannelWebhookIPLimitPerMinute int `mapstructure:"channel-webhook-ip-limit-per-minute" json:"channel-webhook-ip-limit-per-minute" yaml:"channel-webhook-ip-limit-per-minute"`
	// OAuthIPLimitPerMinute 公开 OAuth（/base/oauth/*）：同一客户端 IP 在滑动 1 分钟窗口内最大请求数；0 表示不限制。依赖 Redis；Redis 不可用时中间件放行（fail-open）
	OAuthIPLimitPerMinute int `mapstructure:"oauth-ip-limit-per-minute" json:"oauth-ip-limit-per-minute" yaml:"oauth-ip-limit-per-minute"`
	// ChannelWebhookEventRetentionDays rag_channel_webhook_events 物理删除早于此时长的记录；0 表示默认 7；-1 关闭每日清理（表可能持续增长）
	ChannelWebhookEventRetentionDays int `mapstructure:"channel-webhook-event-retention-days" json:"channel-webhook-event-retention-days" yaml:"channel-webhook-event-retention-days"`
	// ChannelOutboundPollSeconds 定时扫描 rag_channel_outbounds 的间隔秒数；0 表示默认 30；-1 表示不注册定时任务（仍允许入队，可后续改配置或手动清队列）
	ChannelOutboundPollSeconds int `mapstructure:"channel-outbound-poll-seconds" json:"channel-outbound-poll-seconds" yaml:"channel-outbound-poll-seconds"`
	// ChannelOutboundMaxAttempts 单条出站任务最大执行次数（超过则丢弃并打日志）；0 表示默认 8
	ChannelOutboundMaxAttempts int `mapstructure:"channel-outbound-max-attempts" json:"channel-outbound-max-attempts" yaml:"channel-outbound-max-attempts"`
	// ChannelOutboundBatchSize 每轮定时任务最多处理条数；0 表示默认 32
	ChannelOutboundBatchSize int `mapstructure:"channel-outbound-batch-size" json:"channel-outbound-batch-size" yaml:"channel-outbound-batch-size"`
	// ChannelOutboundClaimLeaseSeconds 多实例出站认领租约秒数；0 表示默认 180；过小易重复投递，过大则崩溃后恢复慢
	ChannelOutboundClaimLeaseSeconds int `mapstructure:"channel-outbound-claim-lease-seconds" json:"channel-outbound-claim-lease-seconds" yaml:"channel-outbound-claim-lease-seconds"`
}
