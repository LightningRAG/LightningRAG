-- RAG 模块 MySQL 表结构
-- 在项目根目录执行: mysql -u root -p lrag < server/scripts/rag_tables_mysql.sql
-- 或使用你的数据库名替换 lrag

CREATE TABLE IF NOT EXISTS `rag_knowledge_bases` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uuid` varchar(36) DEFAULT NULL,
  `name` varchar(128) DEFAULT NULL COMMENT '知识库名称',
  `description` text COMMENT '描述',
  `owner_id` bigint unsigned DEFAULT NULL COMMENT '所有者用户ID',
  `embedding_id` bigint unsigned DEFAULT NULL COMMENT '嵌入模型配置ID',
  `vector_store_id` bigint unsigned DEFAULT NULL COMMENT '向量存储配置ID',
  `file_storage_id` bigint unsigned DEFAULT NULL COMMENT '文件存储配置ID',
  `retriever_type` varchar(32) DEFAULT 'vector' COMMENT '检索类型',
  `chunk_size` int DEFAULT 500 COMMENT '切片大小',
  `chunk_overlap` int DEFAULT 50 COMMENT '切片重叠',
  PRIMARY KEY (`id`),
  KEY `idx_rag_knowledge_bases_deleted_at` (`deleted_at`),
  KEY `idx_rag_knowledge_bases_owner_id` (`owner_id`),
  KEY `idx_rag_knowledge_bases_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_documents` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uuid` varchar(36) DEFAULT NULL,
  `knowledge_base_id` bigint unsigned DEFAULT NULL COMMENT '所属知识库ID',
  `name` varchar(256) DEFAULT NULL COMMENT '文件名',
  `file_type` varchar(32) DEFAULT NULL COMMENT '文件类型',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小',
  `storage_path` varchar(512) DEFAULT NULL COMMENT '存储路径',
  `status` varchar(32) DEFAULT 'processing' COMMENT '状态',
  `chunk_count` int DEFAULT 0 COMMENT '切片数量',
  `error_msg` text COMMENT '错误信息',
  PRIMARY KEY (`id`),
  KEY `idx_rag_documents_deleted_at` (`deleted_at`),
  KEY `idx_rag_documents_knowledge_base_id` (`knowledge_base_id`),
  KEY `idx_rag_documents_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_chunks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uuid` varchar(36) DEFAULT NULL,
  `document_id` bigint unsigned DEFAULT NULL COMMENT '所属文档ID',
  `content` text COMMENT '切片文本内容',
  `vector_store_id` varchar(64) DEFAULT NULL COMMENT '向量库中的ID',
  `page_index` int DEFAULT NULL COMMENT '页码',
  `chunk_index` int DEFAULT NULL COMMENT '切片序号',
  `metadata` text COMMENT 'JSON元数据',
  PRIMARY KEY (`id`),
  KEY `idx_rag_chunks_deleted_at` (`deleted_at`),
  KEY `idx_rag_chunks_document_id` (`document_id`),
  KEY `idx_rag_chunks_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_llm_providers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL COMMENT '提供商名称',
  `model_name` varchar(128) DEFAULT NULL COMMENT '模型名称',
  `model_types` text COMMENT '适用场景 JSON 如 ["chat","rerank"]',
  `base_url` varchar(256) DEFAULT NULL COMMENT 'API Base URL',
  `api_key` varchar(512) DEFAULT NULL COMMENT 'API Key',
  `config` text COMMENT '额外配置 JSON',
  `max_context_tokens` bigint unsigned DEFAULT 0 COMMENT '最大上下文token数，0表示不限制',
  `supports_deep_thinking` tinyint(1) DEFAULT 0 COMMENT '是否支持深度思考',
  `supports_tool_call` tinyint(1) DEFAULT 1 COMMENT '是否支持工具调用',
  `share_scope` varchar(32) DEFAULT 'private' COMMENT '共享范围',
  `share_target` text COMMENT '共享目标 JSON',
  `enabled` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`),
  KEY `idx_rag_llm_providers_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_embedding_providers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `model_name` varchar(128) DEFAULT NULL,
  `base_url` varchar(256) DEFAULT NULL,
  `api_key` varchar(512) DEFAULT NULL,
  `config` text,
  `dimensions` int DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_rag_embedding_providers_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_vector_store_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `provider` varchar(32) DEFAULT NULL,
  `config` text,
  `enabled` tinyint(1) DEFAULT 1,
  `allow_all` tinyint(1) DEFAULT 1 COMMENT '是否所有角色可选用',
  `allowed_authority_ids` text COMMENT 'JSON 数组：指定角色 ID',
  PRIMARY KEY (`id`),
  KEY `idx_rag_vector_store_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_file_storage_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `provider` varchar(32) DEFAULT NULL,
  `config` text,
  `enabled` tinyint(1) DEFAULT 1,
  `allow_all` tinyint(1) DEFAULT 1 COMMENT '是否所有角色可选用',
  `allowed_authority_ids` text COMMENT 'JSON 数组：指定角色 ID',
  PRIMARY KEY (`id`),
  KEY `idx_rag_file_storage_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_user_llms` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `provider` varchar(64) DEFAULT NULL,
  `model_name` varchar(128) DEFAULT NULL,
  `model_types` text COMMENT '适用场景 JSON 如 ["chat","rerank"]',
  `base_url` varchar(256) DEFAULT NULL,
  `api_key` varchar(512) DEFAULT NULL,
  `config` text,
  `max_context_tokens` bigint unsigned DEFAULT 0 COMMENT '最大上下文token数，0表示不限制',
  `supports_deep_thinking` tinyint(1) DEFAULT 0 COMMENT '是否支持深度思考',
  `supports_tool_call` tinyint(1) DEFAULT 1 COMMENT '是否支持工具调用',
  `enabled` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_rag_user_llms_deleted_at` (`deleted_at`),
  KEY `idx_rag_user_llms_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_knowledge_base_shares` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `knowledge_base_id` bigint unsigned DEFAULT NULL COMMENT '知识库ID',
  `share_type` varchar(32) DEFAULT NULL COMMENT '分享类型',
  `target_type` varchar(32) DEFAULT NULL COMMENT '目标类型',
  `target_id` bigint unsigned DEFAULT NULL COMMENT '目标ID',
  `permission` varchar(32) DEFAULT 'read' COMMENT '权限',
  PRIMARY KEY (`id`),
  KEY `idx_rag_knowledge_base_shares_deleted_at` (`deleted_at`),
  KEY `idx_rag_knowledge_base_shares_knowledge_base_id` (`knowledge_base_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_conversations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uuid` varchar(36) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `title` varchar(256) DEFAULT '新对话' COMMENT '对话标题',
  `llm_provider_id` bigint unsigned DEFAULT NULL COMMENT '使用的LLM配置ID',
  `llm_source` varchar(16) DEFAULT 'user' COMMENT '模型来源 admin|user',
  `source_type` varchar(32) DEFAULT NULL COMMENT '来源类型',
  `source_ids` text COMMENT '来源ID JSON',
  `enabled_tool_names` text COMMENT '启用的工具 JSON 数组如 ["web_search"]，空则不使用任何工具',
  PRIMARY KEY (`id`),
  KEY `idx_rag_conversations_deleted_at` (`deleted_at`),
  KEY `idx_rag_conversations_user_id` (`user_id`),
  KEY `idx_rag_conversations_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_messages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uuid` varchar(36) DEFAULT NULL,
  `conversation_id` bigint unsigned DEFAULT NULL COMMENT '会话ID',
  `role` varchar(16) DEFAULT NULL COMMENT '角色',
  `content` text COMMENT '消息内容',
  `token_count` int DEFAULT NULL COMMENT 'Token 数',
  `references` text COMMENT '引用来源 JSON 数组',
  PRIMARY KEY (`id`),
  KEY `idx_rag_messages_deleted_at` (`deleted_at`),
  KEY `idx_rag_messages_conversation_id` (`conversation_id`),
  KEY `idx_rag_messages_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rag_agents` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `owner_id` bigint unsigned DEFAULT NULL COMMENT '创建者用户ID',
  `name` varchar(128) DEFAULT NULL COMMENT 'Agent 名称',
  `desc` varchar(512) DEFAULT NULL COMMENT '描述',
  `dsl` longtext COMMENT 'DSL JSON 工作流定义',
  PRIMARY KEY (`id`),
  KEY `idx_rag_agents_deleted_at` (`deleted_at`),
  KEY `idx_rag_agents_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 若 rag_knowledge_bases 已存在，添加 file_storage_id 列：
-- ALTER TABLE `rag_knowledge_bases` ADD COLUMN `file_storage_id` bigint unsigned DEFAULT NULL COMMENT '文件存储配置ID' AFTER `vector_store_id`;

-- 若 rag_conversations 已存在，可单独执行以下语句添加 llm_source 列：
-- ALTER TABLE `rag_conversations` ADD COLUMN `llm_source` varchar(16) DEFAULT 'user' COMMENT '模型来源 admin|user' AFTER `llm_provider_id`;

-- 若 rag_conversations 已存在，添加 enabled_tool_names 列（用户选中的工具，默认全不启用）：
-- ALTER TABLE `rag_conversations` ADD COLUMN `enabled_tool_names` text COMMENT '启用的工具 JSON 数组如 ["web_search"]，空则不使用任何工具' AFTER `source_ids`;

-- 若表已存在，添加 model_types 列（适用于已有数据库升级）：
-- ALTER TABLE `rag_llm_providers` ADD COLUMN `model_types` text COMMENT '适用场景 JSON 如 ["chat","rerank"]' AFTER `model_name`;
-- ALTER TABLE `rag_user_llms` ADD COLUMN `model_types` text COMMENT '适用场景 JSON 如 ["chat","rerank"]' AFTER `model_name`;

-- 若表已存在，添加 max_context_tokens 列（用于上下文长度限制）：
-- ALTER TABLE `rag_llm_providers` ADD COLUMN `max_context_tokens` bigint unsigned DEFAULT 0 COMMENT '最大上下文token数，0表示不限制' AFTER `config`;
-- ALTER TABLE `rag_user_llms` ADD COLUMN `max_context_tokens` bigint unsigned DEFAULT 0 COMMENT '最大上下文token数，0表示不限制' AFTER `config`;

-- 若表已存在，添加 supports_deep_thinking、supports_tool_call 列（模型能力配置）：
-- ALTER TABLE `rag_llm_providers` ADD COLUMN `supports_deep_thinking` tinyint(1) DEFAULT 0 COMMENT '是否支持深度思考' AFTER `max_context_tokens`;
-- ALTER TABLE `rag_llm_providers` ADD COLUMN `supports_tool_call` tinyint(1) DEFAULT 1 COMMENT '是否支持工具调用' AFTER `supports_deep_thinking`;
-- ALTER TABLE `rag_user_llms` ADD COLUMN `supports_deep_thinking` tinyint(1) DEFAULT 0 COMMENT '是否支持深度思考' AFTER `max_context_tokens`;
-- ALTER TABLE `rag_user_llms` ADD COLUMN `supports_tool_call` tinyint(1) DEFAULT 1 COMMENT '是否支持工具调用' AFTER `supports_deep_thinking`;

CREATE TABLE IF NOT EXISTS `rag_user_web_search_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `provider` varchar(32) DEFAULT 'duckduckgo' COMMENT '搜索引擎 duckduckgo|baidu',
  `config` text COMMENT '引擎配置 JSON 如 {"apiKey":"xxx"}',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_web_search` (`user_id`),
  KEY `idx_rag_user_web_search_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户互联网搜索配置';

-- ========== 添加 Agent 菜单（若添加了 Agent 功能但看不到菜单，执行以下 SQL） ==========
-- 说明：菜单初始化仅在首次建库时执行，后续新增的菜单需手动插入
-- 执行前请确认 sys_base_menus 中已有 name='rag' 的父菜单

-- 1. 插入 Agent 编排菜单（若已存在则跳过）
INSERT INTO `sys_base_menus` (`created_at`,`updated_at`,`deleted_at`,`menu_level`,`parent_id`,`path`,`name`,`hidden`,`component`,`sort`,`title`,`icon`,`keep_alive`,`default_menu`,`close_tab`)
SELECT NOW(), NOW(), NULL, 1, r.id, 'agent', 'ragAgent', 0, 'view/rag/agent/index.vue', 4, 'Agent 编排', 'connection', 0, 0, 0
FROM `sys_base_menus` r
WHERE r.name = 'rag' AND r.parent_id = 0 AND r.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_base_menus` x WHERE x.name = 'ragAgent' AND x.deleted_at IS NULL)
LIMIT 1;

-- 2. 插入 Agent 编辑器菜单（隐藏，若已存在则跳过）
INSERT INTO `sys_base_menus` (`created_at`,`updated_at`,`deleted_at`,`menu_level`,`parent_id`,`path`,`name`,`hidden`,`component`,`sort`,`title`,`icon`,`keep_alive`,`default_menu`,`close_tab`)
SELECT NOW(), NOW(), NULL, 1, r.id, 'agentEditor', 'ragAgentEditor', 1, 'view/rag/agent/editor.vue', 5, 'Agent 编辑器', 'edit', 0, 0, 0
FROM `sys_base_menus` r
WHERE r.name = 'rag' AND r.parent_id = 0 AND r.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_base_menus` x WHERE x.name = 'ragAgentEditor' AND x.deleted_at IS NULL)
LIMIT 1;

-- 3. 为超级管理员角色(888)分配 Agent 菜单权限
INSERT IGNORE INTO `sys_authority_menus` (`sys_base_menu_id`, `sys_authority_authority_id`)
SELECT m.id, 888 FROM `sys_base_menus` m
WHERE m.name IN ('ragAgent', 'ragAgentEditor') AND m.deleted_at IS NULL;

-- 若 rag_messages 已存在，添加 references 列（用于持久化对话引用来源）：
-- ALTER TABLE `rag_messages` ADD COLUMN `references` text COMMENT '引用来源 JSON 数组' AFTER `token_count`;
