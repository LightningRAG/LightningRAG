package rag

import (
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
)

// 仅缓存「存在实体」的阳性结果，避免 COUNT 为 0 时写入缓存导致新抽取后长期误判（对齐常见 KV 缓存用法）。

func kgEntityPresenceCacheKey(kbID uint) string {
	return fmt.Sprintf("lrag_kghas1:%d", kbID)
}

// InvalidateKgEntityPresenceCache 在图谱数据被删除或瘦身时调用，使检索器重新走 DB 判断
func InvalidateKgEntityPresenceCache(kbID uint) {
	global.BlackCache.Delete(kgEntityPresenceCacheKey(kbID))
}
