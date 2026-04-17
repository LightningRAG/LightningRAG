package rag

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	ragService "github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type KnowledgeBaseApi struct{}

// Create 创建知识库
func (k *KnowledgeBaseApi) Create(c *gin.Context) {
	var req request.KnowledgeBaseCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	authorityID := utils.GetUserAuthorityId(c)
	kb, err := knowledgeBaseService.Create(c.Request.Context(), uid, authorityID, req)
	if err != nil {
		global.LRAG_LOG.Error("创建知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(kb, c)
}

// List 知识库列表
func (k *KnowledgeBaseApi) List(c *gin.Context) {
	var req request.KnowledgeBaseList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	list, total, err := knowledgeBaseService.List(c.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取知识库列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// Get 获取知识库详情
func (k *KnowledgeBaseApi) Get(c *gin.Context) {
	var req request.KnowledgeBaseGet
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	kb, err := knowledgeBaseService.Get(c.Request.Context(), uid, req.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(kb, c)
}

// Update 更新知识库
func (k *KnowledgeBaseApi) Update(c *gin.Context) {
	var req request.KnowledgeBaseUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := knowledgeBaseService.Update(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("更新知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

// Delete 删除知识库
func (k *KnowledgeBaseApi) Delete(c *gin.Context) {
	var req request.KnowledgeBaseDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := knowledgeBaseService.Delete(c.Request.Context(), uid, req.ID); err != nil {
		global.LRAG_LOG.Error("删除知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// ListDocuments 获取知识库下的文档列表
func (k *KnowledgeBaseApi) ListDocuments(c *gin.Context) {
	var req request.DocumentList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	list, total, err := knowledgeBaseService.ListDocuments(c.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取文档列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// UploadDocument 上传文档到知识库
func (k *KnowledgeBaseApi) UploadDocument(c *gin.Context) {
	kbID := c.PostForm("knowledgeBaseId")
	if kbID == "" {
		response.FailWithMessage(i18n.Msg(c, "rag.kb.knowledge_base_id_required"), c)
		return
	}
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		response.FailWithMessage(i18n.Msg(c, "rag.kb.receive_file_failed"), c)
		return
	}
	uid := utils.GetUserID(c)
	doc, err := knowledgeBaseService.UploadDocument(c.Request.Context(), uid, kbID, header)
	if err != nil {
		global.LRAG_LOG.Error("上传文档失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(doc, c)
}

// GetDocument 获取文档详情
func (k *KnowledgeBaseApi) GetDocument(c *gin.Context) {
	var req request.DocumentGet
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	doc, err := knowledgeBaseService.GetDocument(c.Request.Context(), uid, req.ID)
	if err != nil {
		global.LRAG_LOG.Error("获取文档失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(doc, c)
}

// DeleteDocument 删除文档
func (k *KnowledgeBaseApi) DeleteDocument(c *gin.Context) {
	var req request.DocumentDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	// 删除含向量与图谱清理，可能超过客户端/代理等待时间；不因客户端断开而取消，避免半删与 context canceled
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 30*time.Minute)
	defer cancel()
	if err := knowledgeBaseService.DeleteDocument(ctx, uid, req.ID); err != nil {
		global.LRAG_LOG.Error("删除文档失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// RetryDocument 重试解析文档
func (k *KnowledgeBaseApi) RetryDocument(c *gin.Context) {
	var req request.DocumentRetry
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	doc, err := knowledgeBaseService.RetryDocument(c.Request.Context(), uid, req.ID)
	if err != nil {
		global.LRAG_LOG.Error("重试文档失败", zap.Error(err))
		if errors.Is(err, ragService.ErrDocumentNoStoragePath) {
			response.FailWithMessage(i18n.Msg(c, "rag.kb.document_no_storage_path"), c)
			return
		}
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(doc, c)
}

// DownloadDocument 下载/预览文档（仅 Local OSS）
// 支持 query: id=文档ID, preview=1 时以内嵌方式返回（浏览器可预览），否则触发下载
func (k *KnowledgeBaseApi) DownloadDocument(c *gin.Context) {
	docID := c.Query("id")
	if docID == "" {
		response.FailWithMessage(i18n.Msg(c, "rag.kb.id_required"), c)
		return
	}
	id64, err := strconv.ParseUint(docID, 10, 32)
	if err != nil {
		response.FailWithMessage(i18n.Msg(c, "rag.kb.id_invalid"), c)
		return
	}
	id := uint(id64)
	uid := utils.GetUserID(c)
	localPath, filename, err2 := knowledgeBaseService.GetDocumentDownloadPath(c.Request.Context(), uid, id)
	if err2 != nil {
		global.LRAG_LOG.Error("获取文档路径失败", zap.Error(err2))
		if errors.Is(err2, ragService.ErrDocumentNoStoragePath) {
			response.FailWithMessage(i18n.Msg(c, "rag.kb.document_no_storage_path"), c)
			return
		}
		response.FailWithError(c, err2)
		return
	}
	// 防止路径遍历
	abs, _ := filepath.Abs(localPath)
	storeAbs, _ := filepath.Abs(global.LRAG_CONFIG.Local.StorePath)
	if !strings.HasPrefix(abs, storeAbs) {
		response.FailWithMessage(i18n.Msg(c, "rag.kb.illegal_path"), c)
		return
	}
	preview := c.Query("preview") == "1"
	if preview {
		c.Header("Content-Disposition", "inline; filename=\""+strings.ReplaceAll(filename, `"`, `\"`)+"\"")
		c.File(localPath)
	} else {
		c.FileAttachment(localPath, filename)
	}
}

// Share 分享知识库
func (k *KnowledgeBaseApi) Share(c *gin.Context) {
	var req request.KnowledgeBaseShare
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := knowledgeBaseService.Share(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("分享知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.share_success"), c)
}

// Transfer 转让知识库
func (k *KnowledgeBaseApi) Transfer(c *gin.Context) {
	var req request.KnowledgeBaseTransfer
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if err := knowledgeBaseService.Transfer(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("转让知识库失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.transfer_success"), c)
}

// ListChunks 获取文档的切片列表
func (k *KnowledgeBaseApi) ListChunks(c *gin.Context) {
	var req request.ChunkList
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	list, total, err := knowledgeBaseService.ListChunks(c.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取切片列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, i18n.Msg(c, "common.fetch_success"), c)
}

// UpdateChunk 更新切片内容
func (k *KnowledgeBaseApi) UpdateChunk(c *gin.Context) {
	var req request.ChunkUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	if err := knowledgeBaseService.UpdateChunk(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("更新切片失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.update_success"), c)
}

// BatchDeleteDocuments 批量删除文档
func (k *KnowledgeBaseApi) BatchDeleteDocuments(c *gin.Context) {
	var req request.DocumentBatchByIDs
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 30*time.Minute)
	defer cancel()
	if err := knowledgeBaseService.BatchDeleteDocuments(ctx, uid, req); err != nil {
		global.LRAG_LOG.Error("批量删除文档失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.delete_success"), c)
}

// BatchReindexDocuments 批量重新切片
func (k *KnowledgeBaseApi) BatchReindexDocuments(c *gin.Context) {
	var req request.DocumentBatchByIDs
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	if err := knowledgeBaseService.BatchReindexDocuments(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("批量重新切片失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "rag.kb.chunk_job_submitted"), c)
}

// BatchCancelDocumentIndexing 批量取消进行中的切片任务
func (k *KnowledgeBaseApi) BatchCancelDocumentIndexing(c *gin.Context) {
	var req request.DocumentBatchByIDs
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	if err := knowledgeBaseService.BatchCancelDocumentIndexing(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("批量取消切片失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.cancelled"), c)
}

// BatchSetDocumentRetrieval 批量启用/禁用文档参与检索
func (k *KnowledgeBaseApi) BatchSetDocumentRetrieval(c *gin.Context) {
	var req request.DocumentBatchRetrieval
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	if err := knowledgeBaseService.BatchSetDocumentRetrieval(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("批量设置检索开关失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.state_updated"), c)
}

// BatchSetDocumentPriority 批量设置文档检索权重（0~1）
func (k *KnowledgeBaseApi) BatchSetDocumentPriority(c *gin.Context) {
	var req request.DocumentBatchPriority
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	if err := knowledgeBaseService.BatchSetDocumentPriority(c.Request.Context(), uid, req); err != nil {
		global.LRAG_LOG.Error("批量设置文档 priority 失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "common.state_updated"), c)
}

// ListEmbeddingProviders 列出可用的嵌入模型（供创建知识库时选择）
func (k *KnowledgeBaseApi) ListEmbeddingProviders(c *gin.Context) {
	uid := utils.GetUserID(c)
	list, err := knowledgeBaseService.ListEmbeddingProviders(c.Request.Context(), uid)
	if err != nil {
		global.LRAG_LOG.Error("获取嵌入模型列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// ListVectorStoreConfigs 列出可用的向量存储配置（供创建知识库时选择）
func (k *KnowledgeBaseApi) ListVectorStoreConfigs(c *gin.Context) {
	authorityID := utils.GetUserAuthorityId(c)
	list, err := knowledgeBaseService.ListVectorStoreConfigs(c.Request.Context(), authorityID)
	if err != nil {
		global.LRAG_LOG.Error("获取向量存储列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}

// Retrieve 在选定知识库中检索文档切片（测试检索）
func (k *KnowledgeBaseApi) Retrieve(c *gin.Context) {
	var req request.KnowledgeBaseRetrieve
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	chunks, err := knowledgeBaseService.Retrieve(c.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("知识库检索失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(chunks, c)
}

// KnowledgeGraph 知识库图谱可视化数据（实体与关系子集，大库截断）
func (k *KnowledgeBaseApi) KnowledgeGraph(c *gin.Context) {
	var req request.KnowledgeBaseKnowledgeGraph
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, err)
		return
	}
	uid := utils.GetUserID(c)
	if uid == 0 {
		response.FailWithMessage(i18n.Msg(c, "common.not_logged_in"), c)
		return
	}
	data, err := knowledgeBaseService.GetKnowledgeGraphViz(c.Request.Context(), uid, req)
	if err != nil {
		global.LRAG_LOG.Error("获取知识图谱数据失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(data, c)
}

// ListFileStorageConfigs 列出可用的文件存储配置（供创建知识库时选择）
func (k *KnowledgeBaseApi) ListFileStorageConfigs(c *gin.Context) {
	authorityID := utils.GetUserAuthorityId(c)
	list, err := knowledgeBaseService.ListFileStorageConfigs(c.Request.Context(), authorityID)
	if err != nil {
		global.LRAG_LOG.Error("获取文件存储列表失败", zap.Error(err))
		response.FailWithError(c, err)
		return
	}
	response.OkWithData(list, c)
}
