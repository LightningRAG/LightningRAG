package rag

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const maxConcurrentSliceJobsPerKB = 32

// indexingClaimLeaseDuration 多实例下切片任务数据库租约时长；向量化前会续期，避免超大文档未跑完租约失效
const indexingClaimLeaseDuration = 45 * time.Minute

var (
	indexingServerInstanceID = uuid.New().String()
	docIndexingInflight      sync.Map // docID -> struct{}，避免同一进程内重复入队
	kbConcurrencyGates       sync.Map // kbID -> *kbConcurrencyGate
)

// tryClaimDocumentIndexing 原子抢占切片任务：仅当无租约或租约已过期时可成功（防多实例重复处理同一文档）
func tryClaimDocumentIndexing(ctx context.Context, docID uint) bool {
	if docID == 0 {
		return false
	}
	now := time.Now()
	until := now.Add(indexingClaimLeaseDuration)
	res := global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("id = ? AND status = ?", docID, "processing").
		Where("(indexing_claim_until IS NULL OR indexing_claim_until < ?)", now).
		Updates(map[string]any{
			"indexing_claim_owner": indexingServerInstanceID,
			"indexing_claim_until": until,
		})
	return res.Error == nil && res.RowsAffected == 1
}

func releaseDocumentIndexingClaim(ctx context.Context, docID uint) {
	if docID == 0 {
		return
	}
	global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("id = ? AND indexing_claim_owner = ?", docID, indexingServerInstanceID).
		Updates(map[string]any{
			"indexing_claim_owner": "",
			"indexing_claim_until": gorm.Expr("NULL"),
		})
}

// renewDocumentIndexingLease 长耗时步骤前续租（仅持有者与 processing 状态）
func renewDocumentIndexingLease(ctx context.Context, docID uint) {
	if docID == 0 {
		return
	}
	until := time.Now().Add(indexingClaimLeaseDuration)
	global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
		Where("id = ? AND status = ? AND indexing_claim_owner = ?", docID, "processing", indexingServerInstanceID).
		Update("indexing_claim_until", until)
}

type kbConcurrencyGate struct {
	mu     sync.Mutex
	cond   *sync.Cond
	active int
}

func newKbConcurrencyGate() *kbConcurrencyGate {
	g := &kbConcurrencyGate{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

func kbConcurrentLimit(kb *rag.RagKnowledgeBase) int {
	n := kb.ConcurrentSliceJobs
	if n < 1 {
		n = 1
	}
	if n > maxConcurrentSliceJobsPerKB {
		n = maxConcurrentSliceJobsPerKB
	}
	return n
}

func getKbConcurrencyGate(kbID uint) *kbConcurrencyGate {
	if v, ok := kbConcurrencyGates.Load(kbID); ok {
		return v.(*kbConcurrencyGate)
	}
	g := newKbConcurrencyGate()
	actual, _ := kbConcurrencyGates.LoadOrStore(kbID, g)
	return actual.(*kbConcurrencyGate)
}

func (g *kbConcurrencyGate) acquire(limit int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if limit < 1 {
		limit = 1
	}
	for g.active >= limit {
		g.cond.Wait()
	}
	g.active++
}

func (g *kbConcurrencyGate) release() {
	g.mu.Lock()
	g.active--
	g.cond.Signal()
	g.mu.Unlock()
}

// safeJoinStorePath validates that the joined path stays within the storage root,
// preventing path traversal via crafted StoragePath values (e.g. "../../etc/passwd").
func safeJoinStorePath(basePath, sub string) (string, error) {
	joined := filepath.Join(basePath, sub)
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", err
	}
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	// strings.HasPrefix is more reliable than the deprecated filepath.HasPrefix
	prefix := absBase + string(filepath.Separator)
	if absJoined != absBase && !strings.HasPrefix(absJoined, prefix) {
		return "", errors.New("path traversal detected")
	}
	return absJoined, nil
}

func loadDocumentFileBytes(doc *rag.RagDocument) ([]byte, error) {
	path := doc.StoragePath
	if path == "" {
		return nil, errors.New("文档无存储路径")
	}
	actualPath, err := safeJoinStorePath(global.LRAG_CONFIG.Local.StorePath, path)
	if err != nil {
		return nil, err
	}
	fileData, readErr := os.ReadFile(actualPath)
	if readErr != nil {
		fileData2, err2 := os.ReadFile(path)
		if err2 != nil {
			return nil, readErr
		}
		fileData = fileData2
	}
	return fileData, nil
}

// runDocumentIndexingFromStorage 在已持有并发槽位的前提下：读文件、解析、切片与向量化
func runDocumentIndexingFromStorage(ctx context.Context, docID uint, ownerUID uint) {
	var doc rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).First(&doc, docID).Error; err != nil {
		return
	}
	if doc.Status != "processing" {
		return
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).First(&kb, doc.KnowledgeBaseID).Error; err != nil {
		return
	}
	if kb.OwnerID != ownerUID {
		global.LRAG_LOG.Warn("文档切片任务跳过：知识库所有者不匹配",
			zap.Uint("docID", docID), zap.Uint("kbOwner", kb.OwnerID), zap.Uint("expectOwner", ownerUID))
		return
	}
	if !tryClaimDocumentIndexing(ctx, docID) {
		return
	}
	defer releaseDocumentIndexingClaim(ctx, docID)

	fileData, err := loadDocumentFileBytes(&doc)
	if err != nil {
		global.LRAG_DB.WithContext(ctx).Model(&doc).Updates(map[string]any{
			"status":    "failed",
			"error_msg": "读取文件失败: " + err.Error(),
		})
		return
	}
	content, parseErr := parseDocumentContent(ctx, fileData, doc.FileType, doc.Name, ownerUID, &kb)
	if parseErr != nil {
		global.LRAG_DB.WithContext(ctx).Model(&doc).Updates(map[string]any{
			"status":    "failed",
			"error_msg": "解析失败: " + parseErr.Error(),
		})
		return
	}
	if doc.Thumbnail == "" {
		thumbnail := GenerateThumbnail(fileData, doc.FileType)
		if thumbnail != "" {
			global.LRAG_DB.WithContext(ctx).Model(&doc).Update("thumbnail", thumbnail)
		}
	}
	tokenCount := EstimateTokenCount(content)
	global.LRAG_DB.WithContext(ctx).Model(&doc).Update("token_count", tokenCount)
	if err := global.LRAG_DB.WithContext(ctx).First(&doc, docID).Error; err != nil {
		return
	}
	if doc.Status != "processing" {
		return
	}
	_ = ProcessDocument(ctx, &doc, content, &kb, ownerUID)
}

// EnqueueDocumentIndexing 将文档加入切片队列（按知识库并发上限）；同一文档同时仅允许一个任务
func EnqueueDocumentIndexing(docID uint, ownerUID uint) {
	if _, loaded := docIndexingInflight.LoadOrStore(docID, struct{}{}); loaded {
		return
	}
	go func() {
		defer docIndexingInflight.Delete(docID)
		ctx := context.Background()
		var kb rag.RagKnowledgeBase
		var doc rag.RagDocument
		if err := global.LRAG_DB.WithContext(ctx).First(&doc, docID).Error; err != nil {
			return
		}
		if doc.Status != "processing" {
			return
		}
		if err := global.LRAG_DB.WithContext(ctx).First(&kb, doc.KnowledgeBaseID).Error; err != nil {
			return
		}
		limit := kbConcurrentLimit(&kb)
		gate := getKbConcurrencyGate(kb.ID)
		gate.acquire(limit)
		defer gate.release()
		runDocumentIndexingFromStorage(ctx, docID, ownerUID)
	}()
}

// ResumeIncompleteDocumentJobs 服务启动后恢复 status=processing 且已落盘的文档切片任务
func ResumeIncompleteDocumentJobs() {
	if global.LRAG_DB == nil {
		return
	}
	ctx := context.Background()
	var docs []rag.RagDocument
	now := time.Now()
	err := global.LRAG_DB.WithContext(ctx).
		Where("status = ? AND storage_path != ''", "processing").
		Where("(indexing_claim_until IS NULL OR indexing_claim_until < ?)", now).
		Find(&docs).Error
	if err != nil {
		global.LRAG_LOG.Error("恢复未完成切片任务查询失败", zap.Error(err))
		return
	}
	for _, d := range docs {
		var kb rag.RagKnowledgeBase
		if global.LRAG_DB.WithContext(ctx).First(&kb, d.KnowledgeBaseID).Error != nil {
			continue
		}
		EnqueueDocumentIndexing(d.ID, kb.OwnerID)
	}
	if len(docs) > 0 {
		global.LRAG_LOG.Info("已重新入队未完成的文档切片任务（已排除有效租约中的任务，避免多实例重复切片）",
			zap.Int("count", len(docs)))
	}
}
