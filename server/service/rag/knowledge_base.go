package rag

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
	"github.com/LightningRAG/LightningRAG/server/service/rag/docparse"
	"github.com/LightningRAG/LightningRAG/server/utils/upload"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"slices"
)

// ragflowStemCounterPattern 对齐 references/ragflow api/db/services._split_name_counter
var ragflowStemCounterPattern = regexp.MustCompile(`^(.*?)\((\d+)\)$`)

// splitRagflowStemCounter 将文件名主干中末尾的 (n) 拆出，若无则 counter<0
func splitRagflowStemCounter(stem string) (mainPart string, counter int, hasCounter bool) {
	m := ragflowStemCounterPattern.FindStringSubmatch(stem)
	if m == nil {
		return stem, 0, false
	}
	mainPart = strings.TrimRight(m[1], " \t\n\r")
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return stem, 0, false
	}
	return mainPart, n, true
}

// dedupeDocumentNameInKnowledgeBase 对齐 ragflow duplicate_name：同名则在扩展名前插入 (1)、(2)…
func dedupeDocumentNameInKnowledgeBase(ctx context.Context, kbID uint, original string) (string, error) {
	const maxRetries = 1000
	current := original
	for retries := 0; retries < maxRetries; retries++ {
		var n int64
		err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).
			Where("knowledge_base_id = ? AND name = ?", kbID, current).
			Count(&n).Error
		if err != nil {
			return "", err
		}
		if n == 0 {
			return current, nil
		}
		base := filepath.Base(current)
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		mainPart, prev, ok := splitRagflowStemCounter(stem)
		next := 1
		if ok {
			next = prev + 1
		}
		current = fmt.Sprintf("%s(%d)%s", mainPart, next, ext)
	}
	return "", i18n.NewErrorf("svc.kb.unique_filename_failed", maxRetries, original)
}

// Create 创建知识库
func (s *KnowledgeBaseService) Create(ctx context.Context, uid uint, authorityID uint, req request.KnowledgeBaseCreate) (*rag.RagKnowledgeBase, error) {
	if err := s.EnsureVectorStoreUsableByAuthority(ctx, authorityID, req.VectorStoreID); err != nil {
		return nil, err
	}
	if err := s.EnsureFileStorageUsableByAuthority(ctx, authorityID, req.FileStorageID); err != nil {
		return nil, err
	}
	enableKG := true
	if req.EnableKnowledgeGraph != nil {
		enableKG = *req.EnableKnowledgeGraph
	}
	kb := &rag.RagKnowledgeBase{
		UUID:                uuid.New(),
		Name:                req.Name,
		Description:         req.Description,
		OwnerID:             uid,
		EmbeddingID:         req.EmbeddingID,
		EmbeddingSource:     req.EmbeddingSource,
		VectorStoreID:       req.VectorStoreID,
		FileStorageID:       req.FileStorageID,
		RetrieverType:       req.RetrieverType,
		ChunkMethod:         req.ChunkMethod,
		ChunkSize:           req.ChunkSize,
		ChunkOverlap:        req.ChunkOverlap,
		ConcurrentSliceJobs: req.ConcurrentSliceJobs,
		Delimiter:           req.Delimiter,
		AutoKeywords:        req.AutoKeywords,
		AutoQuestions:       req.AutoQuestions,
		UseRerank:           req.UseRerank,
		RerankID:            req.RerankID,
		RerankSource:        req.RerankSource,
		RerankTopK:          req.RerankTopK,
		// PageIndex LLM
		PageIndexLLMID:     req.PageIndexLLMID,
		PageIndexLLMSource: req.PageIndexLLMSource,
		// OCR
		UseOCR:    req.UseOCR,
		OCRID:     req.OCRID,
		OCRSource: req.OCRSource,
		// CV 图片描述
		UseImageDescription:    req.UseImageDescription,
		ImageDescriptionID:     req.ImageDescriptionID,
		ImageDescriptionSource: req.ImageDescriptionSource,
		// Speech2Text
		UseSpeech2Text:       req.UseSpeech2Text,
		Speech2TextID:        req.Speech2TextID,
		Speech2TextSource:    req.Speech2TextSource,
		EnableKnowledgeGraph: enableKG,
	}
	if kb.RetrieverType == "" {
		kb.RetrieverType = "vector"
	}
	if strings.ToLower(strings.TrimSpace(kb.RetrieverType)) == "naive" {
		kb.RetrieverType = "vector"
	}
	if kb.ChunkMethod == "" {
		kb.ChunkMethod = rag.ChunkMethodGeneral
	}
	if kb.ChunkSize <= 0 {
		kb.ChunkSize = 500
	}
	if kb.ChunkOverlap <= 0 {
		kb.ChunkOverlap = 50
	}
	if kb.ConcurrentSliceJobs < 1 {
		kb.ConcurrentSliceJobs = 1
	}
	if kb.ConcurrentSliceJobs > maxConcurrentSliceJobsPerKB {
		kb.ConcurrentSliceJobs = maxConcurrentSliceJobsPerKB
	}
	if kb.Delimiter == "" {
		kb.Delimiter = `\n!?。；！？`
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(kb).Error; err != nil {
		return nil, err
	}
	return kb, nil
}

// List 知识库列表
func (s *KnowledgeBaseService) List(ctx context.Context, uid uint, req request.KnowledgeBaseList) ([]rag.RagKnowledgeBase, int64, error) {
	var list []rag.RagKnowledgeBase
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagKnowledgeBase{}).Where("owner_id = ?", uid)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Scopes(req.Paginate()).Find(&list).Error
	return list, total, err
}

// Get 获取知识库
func (s *KnowledgeBaseService) Get(ctx context.Context, uid uint, id uint) (*rag.RagKnowledgeBase, error) {
	var kb rag.RagKnowledgeBase
	err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", id, uid).First(&kb).Error
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// Update 更新知识库（向量存储、文件存储创建后不可更改；嵌入模型仅可切换同维度）
func (s *KnowledgeBaseService) Update(ctx context.Context, uid uint, req request.KnowledgeBaseUpdate) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.EmbeddingID > 0 {
		updates["embedding_id"] = req.EmbeddingID
		if req.EmbeddingSource != nil {
			updates["embedding_source"] = *req.EmbeddingSource
		}
	}
	if req.RetrieverType != "" {
		rt := req.RetrieverType
		if strings.ToLower(strings.TrimSpace(rt)) == "naive" {
			rt = "vector"
		}
		updates["retriever_type"] = rt
	}
	if req.ChunkMethod != "" {
		updates["chunk_method"] = req.ChunkMethod
	}
	if req.ChunkSize > 0 {
		updates["chunk_size"] = req.ChunkSize
	}
	if req.ChunkOverlap >= 0 {
		updates["chunk_overlap"] = req.ChunkOverlap
	}
	if req.ConcurrentSliceJobs > 0 {
		n := req.ConcurrentSliceJobs
		if n > maxConcurrentSliceJobsPerKB {
			n = maxConcurrentSliceJobsPerKB
		}
		if n < 1 {
			n = 1
		}
		updates["concurrent_slice_jobs"] = n
	}
	if req.Delimiter != "" {
		updates["delimiter"] = req.Delimiter
	}
	if req.AutoKeywords >= 0 {
		updates["auto_keywords"] = req.AutoKeywords
	}
	if req.AutoQuestions >= 0 {
		updates["auto_questions"] = req.AutoQuestions
	}
	if req.UseRerank != nil {
		updates["use_rerank"] = *req.UseRerank
	}
	if req.RerankID != nil {
		updates["rerank_id"] = *req.RerankID
	}
	if req.RerankSource != nil {
		updates["rerank_source"] = *req.RerankSource
	}
	if req.RerankTopK != nil {
		updates["rerank_top_k"] = *req.RerankTopK
	}
	// PageIndex LLM
	if req.PageIndexLLMID != nil {
		updates["page_index_llm_id"] = *req.PageIndexLLMID
	}
	if req.PageIndexLLMSource != nil {
		updates["page_index_llm_source"] = *req.PageIndexLLMSource
	}
	// OCR
	if req.UseOCR != nil {
		updates["use_ocr"] = *req.UseOCR
	}
	if req.OCRID != nil {
		updates["ocr_id"] = *req.OCRID
	}
	if req.OCRSource != nil {
		updates["ocr_source"] = *req.OCRSource
	}
	// CV 图片描述
	if req.UseImageDescription != nil {
		updates["use_image_description"] = *req.UseImageDescription
	}
	if req.ImageDescriptionID != nil {
		updates["image_description_id"] = *req.ImageDescriptionID
	}
	if req.ImageDescriptionSource != nil {
		updates["image_description_source"] = *req.ImageDescriptionSource
	}
	// Speech2Text
	if req.UseSpeech2Text != nil {
		updates["use_speech2_text"] = *req.UseSpeech2Text
	}
	if req.Speech2TextID != nil {
		updates["speech2_text_id"] = *req.Speech2TextID
	}
	if req.Speech2TextSource != nil {
		updates["speech2_text_source"] = *req.Speech2TextSource
	}
	if req.EnableKnowledgeGraph != nil {
		updates["enable_knowledge_graph"] = *req.EnableKnowledgeGraph
	}
	if len(updates) == 0 {
		return nil
	}
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagKnowledgeBase{}).
		Where("id = ? AND owner_id = ?", req.ID, uid).Updates(updates).Error
}

// Delete 删除知识库
func (s *KnowledgeBaseService) Delete(ctx context.Context, uid uint, id uint) error {
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", id, uid).First(&kb).Error; err != nil {
		return err
	}
	DeleteKnowledgeGraphForKnowledgeBase(ctx, &kb, uid)
	var kbDocs []rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("knowledge_base_id = ?", id).Find(&kbDocs).Error; err != nil {
		return err
	}
	for i := range kbDocs {
		deleteDocumentSliceStorage(ctx, &kb, uid, kbDocs[i].ID)
	}
	err := global.LRAG_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("knowledge_base_id = ?", id).Delete(&rag.RagDocument{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND owner_id = ?", id, uid).Delete(&rag.RagKnowledgeBase{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		BumpRetrieveCacheEpochForKnowledgeBase(id)
	}
	return err
}

// UploadDocument 上传文档到知识库，先落盘 OSS 再解析、切片与向量化
func (s *KnowledgeBaseService) UploadDocument(ctx context.Context, uid uint, kbIDStr string, header *multipart.FileHeader) (*rag.RagDocument, error) {
	kbID, err := strconv.ParseUint(kbIDStr, 10, 32)
	if err != nil {
		return nil, err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", kbID, uid).First(&kb).Error; err != nil {
		return nil, err
	}

	filename := header.Filename
	var dedupeErr error
	filename, dedupeErr = dedupeDocumentNameInKnowledgeBase(ctx, uint(kbID), filename)
	if dedupeErr != nil {
		return nil, dedupeErr
	}
	doc := &rag.RagDocument{
		UUID:             uuid.New(),
		KnowledgeBaseID:  uint(kbID),
		Name:             filename,
		FileType:         inferFileType(filename),
		FileSize:         header.Size,
		Status:           "processing",
		RetrievalEnabled: true,
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(doc).Error; err != nil {
		return nil, err
	}

	// 1. 先上传到 OSS 落盘，写入 StoragePath
	oss := upload.NewOss()
	_, storageKey, uploadErr := oss.UploadFile(header)
	if uploadErr != nil {
		global.LRAG_DB.WithContext(ctx).Model(doc).Updates(map[string]any{
			"status":    "failed",
			"error_msg": i18n.Tf(i18n.DefaultLocale, "svc.kb.file_storage_failed", uploadErr.Error()),
		})
		return doc, nil
	}
	// 存储路径：使用 storageKey 便于 OSS 删除；Local 时实际文件在 StorePath/storageKey
	doc.StoragePath = storageKey
	global.LRAG_DB.WithContext(ctx).Model(doc).Update("storage_path", doc.StoragePath)

	// 2. 打开文件流进行解析（header.Open 可多次调用）
	file, openErr := header.Open()
	if openErr != nil {
		global.LRAG_DB.WithContext(ctx).Model(doc).Updates(map[string]any{
			"status":    "failed",
			"error_msg": i18n.Tf(i18n.DefaultLocale, "svc.kb.file_read_failed", openErr.Error()),
		})
		return doc, nil
	}
	defer file.Close()

	// 多格式解析：见 parseDocumentContent 与 docs/DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md（中文全文见 DOCUMENT_PARSE_RAGFLOW_ALIGNMENT_zh.md）
	fileData, readErr := readFileHeaderData(file)
	if readErr != nil {
		global.LRAG_DB.WithContext(ctx).Model(doc).Updates(map[string]any{
			"status":    "failed",
			"error_msg": i18n.Tf(i18n.DefaultLocale, "svc.kb.file_read_failed", readErr.Error()),
		})
		return doc, nil
	}

	// 生成缩略图
	thumbnail := GenerateThumbnail(fileData, doc.FileType)
	if thumbnail != "" {
		global.LRAG_DB.WithContext(ctx).Model(doc).Update("thumbnail", thumbnail)
	}

	// 解析、切片与向量化在后台按知识库并发限制执行；进程重启后会自动恢复未完成的 processing 任务
	EnqueueDocumentIndexing(doc.ID, uid)
	return doc, nil
}

// inferExtPair 扩展名 → fileType；在 init 中按 ext 长度降序排序，避免 map 遍历不确定，且使 .jsonl 优先于 .json 等更长后缀优先匹配。
type inferExtPair struct {
	ext string
	ft  string
}

var inferExtPairs []inferExtPair

func init() {
	inferExtPairs = []inferExtPair{
		{".markdown", "md"},
		{".ndjson", "jsonl"}, {".ldjson", "jsonl"}, {".jsonl", "jsonl"},
		{".mhtml", "html"}, {".xhtml", "html"},
		{".ipynb", "ipynb"},
		{".xlsx", "xlsx"}, {".docx", "docx"}, {".pptx", "pptx"},
		// OOXML 宏与模板：ZIP 布局与 docx/xlsx/pptx 相同，走同一解析器
		{".docm", "docx"}, {".dotm", "docx"}, {".dotx", "docx"},
		{".xlsm", "xlsx"}, {".xltm", "xlsx"}, {".xltx", "xlsx"},
		{".pptm", "pptx"}, {".potm", "pptx"}, {".potx", "pptx"},
		{".tiff", "tiff"}, {".jpeg", "jpg"}, {".jfif", "jpg"}, {".jpe", "jpg"},
		{".icon", "ico"}, {".ico", "ico"}, {".apng", "apng"},
		{".epub", "epub"}, {".numbers", "numbers"}, {".pages", "pages"},
		{".yaml", "yaml"},
		{".html", "html"}, {".geojson", "json"}, {".topojson", "json"},
		{".json", "json"}, {".har", "json"}, {".toml", "toml"},
		{".svgz", "gz"},
		{".tgz", "gz"},
		{".tbz2", "bz2"},
		{".bz2", "bz2"},
		{".pdf", "pdf"}, {".doc", "doc"}, {".txt", "txt"},
		{".mdx", "mdx"}, {".mkd", "md"}, {".md", "md"},
		{".jpg", "jpg"}, {".png", "png"}, {".gif", "gif"}, {".webp", "webp"},
		{".avif", "avif"}, {".heic", "heic"}, {".heif", "heic"},
		{".svg", "svg"},
		{".xls", "xls"}, {".xlsb", "xlsb"}, {".csv", "csv"}, {".tsv", "tsv"},
		{".htm", "html"}, {".mht", "html"}, {".xht", "html"}, {".smi", "html"},
		{".ppt", "ppt"},
		{".key", "key"},
		{".yml", "yaml"},
		{".rtf", "rtf"},
		{".odt", "odt"}, {".ods", "ods"}, {".odp", "odp"}, {".odg", "odg"},
		{".xliff", "xml"}, {".xlf", "xml"},
		{".xslt", "xml"}, {".xsl", "xml"},
		{".rss", "xml"}, {".atom", "xml"}, {".plist", "xml"},
		{".gpx", "xml"}, {".kml", "xml"},
		{".drawio", "xml"}, {".dio", "xml"},
		{".wsdl", "xml"}, {".xsd", "xml"},
		{".rng", "xml"}, {".sch", "xml"}, {".fo", "xml"},
		{".xml", "xml"},
		{".bmp", "bmp"}, {".tif", "tiff"},
		{".eml", "eml"}, {".msg", "msg"},
		{".wps", "wps"}, {".et", "et"}, {".dps", "dps"}, {".hlp", "hlp"},
		{".gz", "gz"},
		{".vtt", "txt"}, {".srt", "txt"}, {".sbv", "txt"}, {".ics", "txt"}, {".ical", "txt"},
		{".ass", "txt"}, {".ssa", "txt"},
	}
	slices.SortFunc(inferExtPairs, func(a, b inferExtPair) int {
		if la, lb := len(a.ext), len(b.ext); la != lb {
			return lb - la
		}
		return strings.Compare(a.ext, b.ext)
	})
}

func inferFileType(filename string) string {
	lower := strings.ToLower(filename)
	for _, suf := range []string{
		".mp4", ".avi", ".mkv", ".mov", ".webm", ".wmv", ".m4v", ".mpeg", ".mpg", ".3gp", ".flv",
		".ogv", ".f4v", ".asf", ".rm", ".rmvb", ".mpe", ".mpa",
	} {
		if strings.HasSuffix(lower, suf) {
			return "video"
		}
	}
	for _, suf := range []string{
		".mp3", ".wav", ".wave", ".m4a", ".ogg", ".flac", ".aac", ".wma",
		".aiff", ".aif", ".au", ".midi", ".mid", ".ape", ".da",
		".realaudio", ".vqf", ".oggvorbis", ".opus", ".spx", ".alac", ".wv",
	} {
		if strings.HasSuffix(lower, suf) {
			return "audio"
		}
	}
	for _, p := range inferExtPairs {
		if strings.HasSuffix(lower, p.ext) {
			return p.ft
		}
	}
	// 对齐 references/ragflow/rag/app/naive.py 中按纯文本解析的源码后缀，并扩展常见工程文本
	for _, suf := range []string{
		".py", ".js", ".mjs", ".cjs", ".java", ".cpp", ".php", ".go", ".ts", ".tsx", ".jsx", ".sh", ".cs", ".kt", ".sql",
		".c", ".h", ".hpp", ".hh", ".hxx", ".swift", ".rb", ".rs", ".lua", ".dart", ".vue", ".svelte", ".scala", ".gradle", ".kts",
		".graphql", ".gql", ".proto", ".ps1", ".psm1", ".vim", ".dockerfile",
		".bat", ".cmd", ".glsl", ".vert", ".frag", ".geom", ".comp", ".hlsl", ".metal",
		".bib", ".sty", ".cls", ".f90", ".f95", ".f03", ".f08", ".for", ".f", ".pm", ".pl",
		".jl", ".r", ".nim", ".ex", ".exs", ".erl", ".hrl", ".clj", ".cljs", ".edn", ".zig",
		".v", ".vh", ".sv", ".vhd", ".vhdl", ".cmake", ".mk", ".patch", ".diff",
		".tf", ".tfvars", ".hcl",
	} {
		if strings.HasSuffix(lower, suf) {
			return "txt"
		}
	}
	// 配置 / 日志 / 文档类：按纯文本入库（与 Ragflow naive 文本管线同类）
	for _, suf := range []string{
		".ini", ".cfg", ".conf", ".properties", ".env", ".log", ".rst", ".tex", ".adoc", ".asciidoc", ".org",
		".css", ".less", ".scss", ".sass", ".styl", ".mdc", ".fish", ".zsh",
		".gitignore", ".gitattributes", ".dockerignore", ".jsonc",
		".j2", ".jinja", ".jinja2",
		".nix", ".bzl", ".bazel", ".bazelrc", ".gn", ".gni", ".cmake.in",
		".pem", ".crt", ".cer",
		".lrc", ".ejs", ".hbs", ".mustache",
		".pug", ".jade", ".liquid", ".twig", ".erb", ".eex", ".coffee", ".cue", ".url", ".mjml",
	} {
		if strings.HasSuffix(lower, suf) {
			return "txt"
		}
	}
	return "txt"
}

// emlAttachmentParseChain 对齐 Ragflow 对邮件附件再解析；限制嵌套 .eml 深度，避免递归过深。
func emlAttachmentParseChain(ctx context.Context, uid uint, kb *rag.RagKnowledgeBase) docparse.AttachmentExtractFunc {
	var depth int
	var fn docparse.AttachmentExtractFunc
	fn = func(ctx context.Context, filename, contentType string, data []byte) (string, error) {
		if depth >= 2 {
			return "", nil
		}
		if strings.TrimSpace(filename) == "" {
			return "", nil
		}
		ft := inferFileType(filename)
		if mimeFT := docparse.FileTypeFromMIME(contentType); docparse.ShouldPreferMIMEOverExt(ft, mimeFT) {
			ft = mimeFT
		}
		ft = docparse.RefineFileTypeByContent(filename, ft, data)
		switch ft {
		case "video":
			return "", nil
		case "eml":
			depth++
			s, err := docparse.ParseEMLTextWithAttachments(ctx, data, fn)
			depth--
			return s, err
		default:
			return parseDocumentContent(ctx, data, ft, filename, uid, kb)
		}
	}
	return fn
}

// stripGzipFilename 去掉 .tgz / .gzip / .gz 后缀（大小写不敏感），供推断内层类型名使用。.tgz 规范化为 .tar 以便识别 tar 包并拒绝。
// .svgz 规范化为 .svg（gzip 压缩的 SVG，解压后走 svg 文本解析）。
func stripGzipFilename(filename string) string {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".svgz") && len(filename) >= len(".svgz") {
		return filename[:len(filename)-len(".svgz")] + ".svg"
	}
	if strings.HasSuffix(lower, ".tgz") && len(filename) >= len(".tgz") {
		return filename[:len(filename)-len(".tgz")] + ".tar"
	}
	if strings.HasSuffix(lower, ".gzip") && len(filename) >= len(".gzip") {
		return filename[:len(filename)-len(".gzip")]
	}
	if strings.HasSuffix(lower, ".gz") && len(filename) >= len(".gz") {
		return filename[:len(filename)-len(".gz")]
	}
	return filename
}

// stripBzip2Filename 去掉 .tbz2 / .bz2 后缀；.tbz2 规范为 .tar 以便拒绝 tar 包。
func stripBzip2Filename(filename string) string {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".tbz2") && len(filename) >= len(".tbz2") {
		return filename[:len(filename)-len(".tbz2")] + ".tar"
	}
	if strings.HasSuffix(lower, ".bz2") && len(filename) >= len(".bz2") {
		return filename[:len(filename)-len(".bz2")]
	}
	return filename
}

// readFileHeaderData 从 multipart 文件中读取全部数据
func readFileHeaderData(file multipart.File) ([]byte, error) {
	return io.ReadAll(file)
}

// parseDocumentContent 根据文件类型解析文档内容，使用知识库配置的模型
func parseDocumentContent(ctx context.Context, data []byte, fileType, filename string, uid uint, kb *rag.RagKnowledgeBase) (string, error) {
	ft := strings.ToLower(fileType)
	ft = docparse.RefineFileTypeByContent(filename, ft, data)
	switch ft {
	case "txt", "md", "mdx":
		return ParseTextContent(bytes.NewReader(data))
	case "svg":
		return ParseTextContent(bytes.NewReader(data))
	case "yaml", "yml":
		return docparse.ParseYAMLText(data)
	case "toml":
		return docparse.ParseTOMLText(data)
	case "rtf":
		return docparse.ParseRTFPlainText(data)
	case "odt":
		return docparse.ParseODTText(data)
	case "ods":
		return docparse.ParseODSText(data)
	case "odp":
		return docparse.ParseODPText(data)
	case "odg":
		return docparse.ParseODGText(data)
	case "epub":
		return docparse.ParseEPUBText(data)
	case "ipynb":
		return docparse.ParseIPynbText(data)
	case "xml":
		return ParseTextContent(bytes.NewReader(data))
	case "json", "jsonl":
		return docparse.ParseJSONText(data)
	case "eml":
		return docparse.ParseEMLTextWithAttachments(ctx, data, emlAttachmentParseChain(ctx, uid, kb))
	case "msg":
		return docparse.ParseMSGText(data)
	case "gz":
		if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
			return "", i18n.NewError("svc.kb.gzip_invalid")
		}
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("%s: %w", i18n.T(i18n.DefaultLocale, "svc.kb.gzip_decompress_failed"), err)
		}
		inner, err := io.ReadAll(gr)
		_ = gr.Close()
		if err != nil {
			return "", fmt.Errorf("%s: %w", i18n.T(i18n.DefaultLocale, "svc.kb.gzip_read_failed"), err)
		}
		base := stripGzipFilename(filename)
		if strings.HasSuffix(strings.ToLower(base), ".tar") {
			return "", i18n.NewError("svc.kb.tar_unsupported")
		}
		innerFT := inferFileType(base)
		innerFT = docparse.RefineFileTypeByContent(base, innerFT, inner)
		if innerFT == "gz" {
			return "", i18n.NewError("svc.kb.gzip_inner_still_compressed")
		}
		return parseDocumentContent(ctx, inner, innerFT, base, uid, kb)
	case "bz2":
		if !docparse.IsBzip2Magic(data) {
			return "", i18n.NewError("svc.kb.bzip2_invalid")
		}
		br := bzip2.NewReader(bytes.NewReader(data))
		inner, err := io.ReadAll(br)
		if err != nil {
			return "", fmt.Errorf("%s: %w", i18n.T(i18n.DefaultLocale, "svc.kb.bz2_decompress_failed"), err)
		}
		base := stripBzip2Filename(filename)
		if strings.HasSuffix(strings.ToLower(base), ".tar") {
			return "", i18n.NewError("svc.kb.tar_bz2_unsupported")
		}
		innerFT := inferFileType(base)
		innerFT = docparse.RefineFileTypeByContent(base, innerFT, inner)
		if innerFT == "bz2" {
			return "", i18n.NewError("svc.kb.bz2_inner_still_compressed")
		}
		return parseDocumentContent(ctx, inner, innerFT, base, uid, kb)
	case "video":
		return "", i18n.NewError("svc.kb.video_unsupported")
	case "wps":
		return "", i18n.NewError("svc.kb.wps_unsupported")
	case "et":
		return "", i18n.NewError("svc.kb.et_unsupported")
	case "dps":
		return "", i18n.NewError("svc.kb.dps_unsupported")
	case "hlp":
		return "", i18n.NewError("svc.kb.hlp_unsupported")
	case "doc":
		return "", i18n.NewError("svc.kb.doc_legacy_unsupported")
	case "ppt":
		return "", i18n.NewError("svc.kb.ppt_legacy_unsupported")
	case "pptx":
		return docparse.ParsePPTXText(data)
	case "pages", "numbers", "key":
		pdfData, err := extractIWorkPreviewPDF(data)
		if err != nil {
			return "", err
		}
		return parseDocumentContent(ctx, pdfData, "pdf", filename, uid, kb)
	case "pdf":
		content, err := ParsePDFContent(bytes.NewReader(data))
		if err != nil {
			// PDF 文本提取失败，尝试 OCR（图片型/扫描件 PDF）
			if kb != nil && !kb.UseOCR {
				return "", i18n.NewErrorf("svc.kb.pdf_failed_ocr_off", err)
			}
			ocrContent, ocrErr := ParseImageWithOCRFromKB(ctx, data, filename, uid, kb)
			if ocrErr != nil {
				return "", i18n.NewErrorf("svc.kb.pdf_and_ocr_failed", err, ocrErr)
			}
			if ocrContent != "" {
				return ocrContent, nil
			}
			return "", err
		}
		return content, nil
	case "docx":
		return ParseDOCXContent(bytes.NewReader(data))
	case "xlsb":
		return "", i18n.NewError("svc.kb.xlsb_unsupported")
	case "xlsx", "xls":
		switch docparse.NormalizeExcelFileType(ft, data) {
		case "xls":
			return docparse.ParseXLSText(data)
		default:
			return ParseXLSXContent(bytes.NewReader(data))
		}
	case "csv":
		return ParseCSVContent(bytes.NewReader(data))
	case "tsv":
		return ParseTSVContent(bytes.NewReader(data))
	case "html", "htm":
		return ParseHTMLContent(bytes.NewReader(data))
	case "jpg", "jpeg", "png", "gif", "webp", "bmp", "tif", "tiff", "avif", "heic", "ico", "apng":
		return ParseImageContentFromKB(ctx, data, filename, uid, kb)
	case "audio", "mp3", "wav", "m4a", "ogg", "flac", "aac", "wma", "wv", "alac":
		return ParseAudioContent(ctx, data, filename, uid, kb)
	default:
		content, err := ParseTextContent(bytes.NewReader(data))
		if err != nil {
			return "", i18n.NewErrorf("svc.kb.unsupported_format", ft)
		}
		return content, nil
	}
}

// GetDocument 获取文档详情
func (s *KnowledgeBaseService) GetDocument(ctx context.Context, uid uint, docID uint) (*rag.RagDocument, error) {
	var doc rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", docID).First(&doc).Error; err != nil {
		return nil, err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", doc.KnowledgeBaseID, uid).First(&kb).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// deleteDocumentSliceStorage 删除文档的切片：向量库中按 vector_store_id 精确删行，再按 document_id 扫尾，最后删 rag_chunks
func deleteDocumentSliceStorage(ctx context.Context, kb *rag.RagKnowledgeBase, uid uint, docID uint) {
	db := global.LRAG_DB.WithContext(ctx)
	var vectorIDs []string
	_ = db.Model(&rag.RagChunk{}).Where("document_id = ?", docID).Pluck("vector_store_id", &vectorIDs).Error
	uniq := make([]string, 0, len(vectorIDs))
	seen := make(map[string]struct{}, len(vectorIDs))
	for _, id := range vectorIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	withVectorStoreForDeleteOps(ctx, kb, uid, func(opCtx context.Context, store interfaces.VectorStore) {
		ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
		if len(uniq) > 0 {
			if err := store.DeleteByIDs(opCtx, uniq); err != nil {
				global.LRAG_LOG.Warn("删除文档向量行失败（已跳过）", zap.Uint("documentId", docID), zap.Error(err))
			}
		}
		if err := store.DeleteByMetadata(opCtx, ns, "document_id", docID); err != nil {
			global.LRAG_LOG.Warn("按 document_id 清理向量失败（已跳过）", zap.Uint("documentId", docID), zap.Error(err))
		}
	})
	db.Where("document_id = ?", docID).Delete(&rag.RagChunk{})
}

// DeleteDocument 删除文档（含向量/PageIndex 数据及 OSS 文件）
func (s *KnowledgeBaseService) DeleteDocument(ctx context.Context, uid uint, docID uint) error {
	var doc rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", docID).First(&doc).Error; err != nil {
		return err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", doc.KnowledgeBaseID, uid).First(&kb).Error; err != nil {
		return err
	}
	CleanupKnowledgeGraphLinksForDocument(ctx, &kb, uid, docID)
	// 向量库 + rag_chunks：先按切片记录的 vector_store_id 删除向量行，再 DeleteByMetadata 清理无主/旧索引，最后删 DB 切片
	deleteDocumentSliceStorage(ctx, &kb, uid, docID)
	// 删除 OSS 文件
	if doc.StoragePath != "" {
		oss := upload.NewOss()
		_ = oss.DeleteFile(doc.StoragePath)
	}
	if err := global.LRAG_DB.WithContext(ctx).Delete(&doc).Error; err != nil {
		return err
	}
	BumpRetrieveCacheEpochForKnowledgeBase(kb.ID)
	return nil
}

// createVectorStoreForDocDelete 为删除文档创建向量存储（仅用于 DeleteByMetadata）
func createVectorStoreForDocDelete(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) (interfaces.VectorStore, error) {
	if kb.VectorStoreID > 0 {
		var vsCfg rag.RagVectorStoreConfig
		if err := global.LRAG_DB.WithContext(ctx).Select("enabled").Where("id = ?", kb.VectorStoreID).First(&vsCfg).Error; err != nil {
			return nil, fmt.Errorf("向量存储配置不存在: %w", err)
		}
		if !vsCfg.Enabled {
			return nil, fmt.Errorf("向量存储配置已禁用")
		}
	}
	emb, err := resolveEmbeddingConfig(ctx, kb, userID)
	if err != nil {
		return nil, err
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		return nil, err
	}
	ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
	return createVectorStoreFromKB(ctx, kb, embedder, ns, emb.Dimensions)
}

// vectorStoreDeleteOpTimeout 删除路径上初始化/操作向量库的时限；配置错误或外部不可达时避免长时间阻塞 HTTP。
const vectorStoreDeleteOpTimeout = 20 * time.Second

// openVectorStoreForDeleteOrNil 限时创建向量存储；失败或超时返回 nil store，调用方应跳过向量删除并继续删库内数据。若 store 非 nil，必须调用 cancel 释放定时器。
func openVectorStoreForDeleteOrNil(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) (store interfaces.VectorStore, opCtx context.Context, cancel context.CancelFunc) {
	opCtx, cancel = context.WithTimeout(ctx, vectorStoreDeleteOpTimeout)
	s, err := createVectorStoreForDocDelete(opCtx, kb, userID)
	if err != nil {
		global.LRAG_LOG.Warn("跳过向量存储删除（初始化失败或超时，可能为嵌入/向量配置错误）", zap.Uint("knowledgeBaseId", kb.ID), zap.Error(err))
		cancel()
		return nil, ctx, func() {}
	}
	if s == nil {
		global.LRAG_LOG.Warn("跳过向量存储删除（向量存储未就绪或不支持的类型）", zap.Uint("knowledgeBaseId", kb.ID))
		cancel()
		return nil, ctx, func() {}
	}
	return s, opCtx, cancel
}

// withVectorStoreForDeleteOps 在限时内创建向量存储并执行 fn；失败时仅打日志，不执行 fn。
func withVectorStoreForDeleteOps(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint, fn func(opCtx context.Context, store interfaces.VectorStore)) {
	store, opCtx, cancel := openVectorStoreForDeleteOrNil(ctx, kb, userID)
	defer cancel()
	if store == nil {
		return
	}
	fn(opCtx, store)
}

// RetryDocument 重试解析文档（仅限 failed 状态）
func (s *KnowledgeBaseService) RetryDocument(ctx context.Context, uid uint, docID uint) (*rag.RagDocument, error) {
	var doc rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", docID).First(&doc).Error; err != nil {
		return nil, err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", doc.KnowledgeBaseID, uid).First(&kb).Error; err != nil {
		return nil, err
	}
	if doc.Status != "failed" && doc.Status != "completed" && doc.Status != "cancelled" {
		return &doc, fmt.Errorf("retry not allowed for status %s (only failed, cancelled, or completed)", doc.Status)
	}
	if doc.StoragePath == "" {
		return nil, ErrDocumentNoStoragePath
	}
	doc.Status = "processing"
	doc.ErrorMsg = ""
	global.LRAG_DB.WithContext(ctx).Model(&doc).Updates(map[string]any{"status": "processing", "error_msg": ""})
	EnqueueDocumentIndexing(doc.ID, uid)
	return &doc, nil
}

// GetDocumentDownloadPath 获取文档本地路径（用于下载/预览，仅 Local OSS）
func (s *KnowledgeBaseService) GetDocumentDownloadPath(ctx context.Context, uid uint, docID uint) (localPath, filename string, err error) {
	doc, err := s.GetDocument(ctx, uid, docID)
	if err != nil {
		return "", "", err
	}
	if doc.StoragePath == "" {
		return "", "", ErrDocumentNoStoragePath
	}
	safePath, err := safeJoinStorePath(global.LRAG_CONFIG.Local.StorePath, doc.StoragePath)
	if err != nil {
		return "", "", err
	}
	localPath = safePath
	return localPath, doc.Name, nil
}

// readFileContent 从本地路径读取内容
func readFileContent(actualPath, fallbackPath string) (string, error) {
	if b, err := os.ReadFile(actualPath); err == nil {
		return strings.TrimSpace(string(b)), nil
	}
	if b, err := os.ReadFile(fallbackPath); err == nil {
		return strings.TrimSpace(string(b)), nil
	}
	return "", i18n.NewErrorf("svc.kb.file_read_failed", actualPath)
}

// ListDocuments 获取知识库下的文档列表
func (s *KnowledgeBaseService) ListDocuments(ctx context.Context, uid uint, req request.DocumentList) ([]rag.RagDocument, int64, error) {
	// 校验用户是否有该知识库权限
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", req.KnowledgeBaseID, uid).First(&kb).Error; err != nil {
		return nil, 0, err
	}
	var list []rag.RagDocument
	var total int64
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagDocument{}).Where("knowledge_base_id = ?", req.KnowledgeBaseID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Scopes(req.Paginate()).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

// Share 分享知识库
func (s *KnowledgeBaseService) Share(ctx context.Context, uid uint, req request.KnowledgeBaseShare) error {
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", req.ID, uid).First(&kb).Error; err != nil {
		return err
	}
	share := &rag.RagKnowledgeBaseShare{
		KnowledgeBaseID: req.ID,
		ShareType:       "share",
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		Permission:      req.Permission,
	}
	if share.Permission == "" {
		share.Permission = "read"
	}
	return global.LRAG_DB.WithContext(ctx).Create(share).Error
}

// Transfer 转让知识库
func (s *KnowledgeBaseService) Transfer(ctx context.Context, uid uint, req request.KnowledgeBaseTransfer) error {
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagKnowledgeBase{}).
		Where("id = ? AND owner_id = ?", req.ID, uid).
		Update("owner_id", req.TargetID).Error
}

// ListEmbeddingProviders 列出可用的嵌入模型（供创建知识库时选择）
// 返回管理员模型（rag_llm_providers embedding 类型）、用户模型（rag_user_llm embedding 类型）
// 以及旧版 rag_embedding_providers 中的遗留数据
func (s *KnowledgeBaseService) ListEmbeddingProviders(ctx context.Context, uid uint) ([]map[string]any, error) {
	result := make([]map[string]any, 0)

	// 管理员模型（embedding 场景）
	var adminLLMs []rag.RagLLMProvider
	if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&adminLLMs).Error; err != nil {
		return nil, err
	}
	for _, p := range adminLLMs {
		if !slices.Contains(p.ModelTypes, "embedding") {
			continue
		}
		result = append(result, map[string]any{
			"id":        p.ID,
			"label":     p.Name + " / " + p.ModelName,
			"name":      p.Name,
			"source":    "admin",
			"modelName": p.ModelName,
		})
	}

	// 用户模型（embedding 场景）
	var userLLMs []rag.RagUserLLM
	if err := global.LRAG_DB.WithContext(ctx).Where("user_id = ? AND enabled = ?", uid, true).Find(&userLLMs).Error; err != nil {
		return nil, err
	}
	for _, u := range userLLMs {
		if !slices.Contains(u.ModelTypes, "embedding") {
			continue
		}
		result = append(result, map[string]any{
			"id":        u.ID,
			"label":     u.Provider + " / " + u.ModelName,
			"name":      u.Provider,
			"source":    "user",
			"modelName": u.ModelName,
		})
	}

	// 兼容旧版：若以上均为空，从 rag_embedding_providers 返回遗留数据
	if len(result) == 0 {
		var legacyList []rag.RagEmbeddingProvider
		if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&legacyList).Error; err != nil {
			return nil, err
		}
		for _, p := range legacyList {
			label := p.Name + " / " + p.ModelName
			if p.Dimensions > 0 {
				label += " (" + strconv.Itoa(p.Dimensions) + "维)"
			}
			result = append(result, map[string]any{
				"id":        p.ID,
				"label":     label,
				"name":      p.Name,
				"source":    "legacy",
				"modelName": p.ModelName,
			})
		}
	}

	return result, nil
}

// syncEmbeddingFromLLM 从 rag_llm_providers（含 embedding 场景）同步到 rag_embedding_providers
func (s *KnowledgeBaseService) syncEmbeddingFromLLM(ctx context.Context) error {
	var llms []rag.RagLLMProvider
	if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&llms).Error; err != nil {
		return err
	}
	for _, p := range llms {
		if !slices.Contains(p.ModelTypes, "embedding") {
			continue
		}
		emb := &rag.RagEmbeddingProvider{
			Name:       strings.ToLower(p.Name),
			ModelName:  p.ModelName,
			BaseURL:    p.BaseURL,
			APIKey:     p.APIKey,
			Dimensions: 0,
			Enabled:    true,
		}
		if emb.Name == "" {
			emb.Name = "openai"
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(emb).Error; err != nil {
			continue
		}
	}
	return nil
}

// embeddingResolved 嵌入模型解析结果
type embeddingResolved struct {
	Name       string
	ModelName  string
	BaseURL    string
	APIKey     string
	Dimensions int
}

// resolveEmbeddingConfig 根据知识库的 EmbeddingSource + EmbeddingID 解析嵌入模型配置
// 支持 admin（rag_llm_providers）、user（rag_user_llm）和 legacy（rag_embedding_providers）三种来源
func resolveEmbeddingConfig(ctx context.Context, kb *rag.RagKnowledgeBase, userID uint) (*embeddingResolved, error) {
	switch kb.EmbeddingSource {
	case "admin":
		var m rag.RagLLMProvider
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND enabled = ?", kb.EmbeddingID, true).First(&m).Error; err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.Tf(i18n.DefaultLocale, "svc.kb.embedding_admin_not_found", kb.EmbeddingID), err)
		}
		return &embeddingResolved{Name: m.Name, ModelName: m.ModelName, BaseURL: m.BaseURL, APIKey: m.APIKey}, nil
	case "user":
		var m rag.RagUserLLM
		if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND user_id = ?", kb.EmbeddingID, userID).First(&m).Error; err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.Tf(i18n.DefaultLocale, "svc.kb.embedding_user_not_found", kb.EmbeddingID), err)
		}
		return &embeddingResolved{Name: m.Provider, ModelName: m.ModelName, BaseURL: m.BaseURL, APIKey: m.APIKey}, nil
	default:
		// legacy: 从旧版 rag_embedding_providers 表读取
		var emb rag.RagEmbeddingProvider
		err := global.LRAG_DB.WithContext(ctx).Where("id = ?", kb.EmbeddingID).First(&emb).Error
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			err = global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).First(&emb).Error
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.Tf(i18n.DefaultLocale, "svc.kb.embedding_not_found", kb.EmbeddingID), err)
		}
		return &embeddingResolved{Name: emb.Name, ModelName: emb.ModelName, BaseURL: emb.BaseURL, APIKey: emb.APIKey, Dimensions: emb.Dimensions}, nil
	}
}

// ListVectorStoreConfigs 列出当前角色可选用的已启用向量存储（供创建知识库时选择）
// 若表为空且主库为 PostgreSQL，则插入默认 PostgreSQL 配置；否则需用户手动添加向量存储
func (s *KnowledgeBaseService) ListVectorStoreConfigs(ctx context.Context, authorityID uint) ([]map[string]any, error) {
	var list []rag.RagVectorStoreConfig
	if err := global.LRAG_DB.WithContext(ctx).Where("enabled = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		// 仅当主库为 PostgreSQL 时插入默认配置，因 pgvector 使用主库连接
		dbType := strings.ToLower(global.LRAG_CONFIG.System.DbType)
		if dbType == "pgsql" || dbType == "postgresql" {
			defaultCfg := &rag.RagVectorStoreConfig{
				Name:                "默认 (PostgreSQL)",
				Provider:            "postgresql",
				Enabled:             true,
				AllowAll:            true,
				AllowedAuthorityIDs: []uint{},
			}
			if err := global.LRAG_DB.WithContext(ctx).Create(defaultCfg).Error; err != nil {
				return nil, err
			}
			list = []rag.RagVectorStoreConfig{*defaultCfg}
		}
	}
	result := make([]map[string]any, 0, len(list))
	for _, vs := range list {
		if !vectorStoreUsableByAuthority(&vs, authorityID) {
			continue
		}
		label := vs.Name
		if vs.Provider != "" {
			label += " (" + vs.Provider + ")"
		}
		result = append(result, map[string]any{
			"id":       vs.ID,
			"label":    label,
			"name":     vs.Name,
			"provider": vs.Provider,
		})
	}
	return result, nil
}

// Retrieve 在选定知识库中按查询检索文档切片，结果结构与对话引用一致
func (s *KnowledgeBaseService) Retrieve(ctx context.Context, uid uint, req request.KnowledgeBaseRetrieve) ([]map[string]any, error) {
	q := strings.TrimSpace(req.Query)
	if q == "" {
		return nil, i18n.NewError("svc.kb.query_empty")
	}
	seen := make(map[uint]struct{})
	var kbIDs []uint
	for _, id := range req.KnowledgeBaseIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		kbIDs = append(kbIDs, id)
	}
	if req.KnowledgeBaseID > 0 {
		if _, ok := seen[req.KnowledgeBaseID]; !ok {
			kbIDs = append(kbIDs, req.KnowledgeBaseID)
		}
	}
	if len(kbIDs) == 0 {
		return nil, i18n.NewError("svc.kb.select_at_least_one")
	}
	topN := req.TopN
	if topN <= 0 && req.ChunkTopK > 0 {
		topN = req.ChunkTopK
	}
	topN = ClampRetrieveTopN(topN, DefaultKnowledgeBaseRetrieveTopNFromConfig())
	session := RetrieverSessionFromLightningRAGParams(req.Mode, req.TopK, req.EnableRerank, req.CosineThreshold, req.MinRerankScore)
	ApplyPageIndexTocEnhanceFromRequest(&session, request.EffectiveTocEnhance(req.TocEnhance, req.TocEnhanceRagflow))
	if session.RetrievePoolTopK == nil {
		if pool := EffectiveDefaultKnowledgeBaseRetrievePoolTopK(); pool > topN {
			p := pool
			session.RetrievePoolTopK = &p
		}
	}
	var llmKw interfaces.LLM
	if global.LRAG_CONFIG.Rag.AutoExtractQueryKeywords && len(req.HlKeywords) == 0 && len(req.LlKeywords) == 0 {
		llmKw = resolveExtractLLM(ctx, uid)
	}
	combined, kgEnt, kgRel, _, _ := PrepareLightningRAGSearchQueries(ctx, uid, llmKw, q, req.HlKeywords, req.LlKeywords, "")
	session.KgEntitySearchQuery = kgEnt
	session.KgRelSearchQuery = kgRel
	searchQ := combined
	docs, err := fetchRelevantDocumentsForKnowledgeBases(ctx, kbIDs, uid, nil, searchQ, topN, session)
	if err != nil {
		return nil, err
	}
	docs = trimDocsToRagTokenBudget(docs, EffectiveMaxRagContextTokens(req.MaxRagContextTokens))
	return ExposeReferencesForAPI(ragDocumentsToRefMaps(docs), req.IncludeReferences, req.IncludeChunkContent), nil
}

// ListChunks 获取文档的切片列表
// 优先从 rag_chunks 关系表查询；若为空但文档有切片（旧数据仅在向量库），则从向量库恢复并回填
func (s *KnowledgeBaseService) ListChunks(ctx context.Context, uid uint, req request.ChunkList) ([]rag.RagChunk, int64, error) {
	var doc rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", req.DocumentID).First(&doc).Error; err != nil {
		return nil, 0, err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", doc.KnowledgeBaseID, uid).First(&kb).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	baseDB := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChunk{}).Where("document_id = ?", req.DocumentID)
	if err := baseDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 && doc.ChunkCount > 0 {
		recovered, err := s.recoverChunksFromVectorStore(ctx, &doc, &kb, uid)
		if err != nil {
			global.LRAG_LOG.Warn("从向量库恢复切片失败", zap.Uint("docID", doc.ID), zap.Error(err))
		}
		if len(recovered) > 0 {
			if err := global.LRAG_DB.WithContext(ctx).CreateInBatches(recovered, 100).Error; err != nil {
				global.LRAG_LOG.Warn("回写切片到数据库失败", zap.Uint("docID", doc.ID), zap.Error(err))
			} else {
				baseDB = global.LRAG_DB.WithContext(ctx).Model(&rag.RagChunk{}).Where("document_id = ?", req.DocumentID)
				baseDB.Count(&total)
			}
		}
	}

	var chunks []rag.RagChunk
	err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChunk{}).
		Where("document_id = ?", req.DocumentID).
		Scopes(req.Paginate()).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, total, err
}

// recoverChunksFromVectorStore 从向量库中按 document_id 恢复切片数据（用于旧文档未持久化到 rag_chunks 的情况）
func (s *KnowledgeBaseService) recoverChunksFromVectorStore(ctx context.Context, doc *rag.RagDocument, kb *rag.RagKnowledgeBase, uid uint) ([]rag.RagChunk, error) {
	emb, err := resolveEmbeddingConfig(ctx, kb, uid)
	if err != nil {
		return nil, err
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		return nil, i18n.NewErrorf("svc.kb.create_embedder_failed", err)
	}

	ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
	store, err := createVectorStoreFromKB(ctx, kb, embedder, ns, emb.Dimensions)
	if err != nil {
		return nil, err
	}

	schemaDocs, err := store.ListByMetadata(ctx, ns, "document_id", doc.ID)
	if err != nil {
		return nil, err
	}

	chunks := make([]rag.RagChunk, 0, len(schemaDocs))
	for _, d := range schemaDocs {
		chunkIndex := 0
		if ci, ok := d.Metadata["chunk_index"]; ok {
			switch v := ci.(type) {
			case float64:
				chunkIndex = int(v)
			case int:
				chunkIndex = v
			}
		}
		chunks = append(chunks, rag.RagChunk{
			UUID:       uuid.New(),
			DocumentID: doc.ID,
			Content:    d.PageContent,
			ChunkIndex: chunkIndex,
		})
	}
	return chunks, nil
}

// UpdateChunk 更新切片内容（同时异步更新向量库）
func (s *KnowledgeBaseService) UpdateChunk(ctx context.Context, uid uint, req request.ChunkUpdate) error {
	var chunk rag.RagChunk
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", req.ID).First(&chunk).Error; err != nil {
		return err
	}
	var doc rag.RagDocument
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", chunk.DocumentID).First(&doc).Error; err != nil {
		return err
	}
	var kb rag.RagKnowledgeBase
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", doc.KnowledgeBaseID, uid).First(&kb).Error; err != nil {
		return err
	}
	if err := global.LRAG_DB.WithContext(ctx).Model(&chunk).Update("content", req.Content).Error; err != nil {
		return err
	}
	// 异步更新向量库：删除旧的再重新写入所有切片（PageIndex 知识库节点也已入向量，与 references/ragflow 行为一致）
	if chunk.VectorStoreID != "" {
		go s.reindexDocumentChunks(doc, kb, uid)
	}
	return nil
}

// reindexDocumentChunks 重新索引文档的所有切片到向量库
func (s *KnowledgeBaseService) reindexDocumentChunks(doc rag.RagDocument, kb rag.RagKnowledgeBase, uid uint) {
	ctx := context.Background()
	emb, err := resolveEmbeddingConfig(ctx, &kb, uid)
	if err != nil {
		return
	}
	embedder, err := registry.CreateEmbedding(registry.EmbeddingConfig{
		Provider:   emb.Name,
		ModelName:  emb.ModelName,
		BaseURL:    emb.BaseURL,
		APIKey:     emb.APIKey,
		Dimensions: emb.Dimensions,
	})
	if err != nil || embedder == nil {
		return
	}
	ns := "kb_" + strconv.FormatUint(uint64(kb.ID), 10)
	store, err := createVectorStoreFromKB(ctx, &kb, embedder, ns, emb.Dimensions)
	if err != nil || store == nil {
		return
	}
	_ = store.DeleteByMetadata(ctx, ns, "document_id", doc.ID)
	var allChunks []rag.RagChunk
	global.LRAG_DB.Where("document_id = ?", doc.ID).Order("chunk_index ASC").Find(&allChunks)
	if len(allChunks) == 0 {
		return
	}
	schemaDocs := make([]schema.Document, len(allChunks))
	for i, c := range allChunks {
		meta := map[string]any{
			"document_id":  doc.ID,
			"chunk_index":  c.ChunkIndex,
			"doc_name":     doc.Name,
			"rag_chunk_id": c.ID,
		}
		enrichMetadataRankBoostForChunk(meta, i, len(allChunks))
		applyDocumentPriorityFloorToRankBoost(meta, doc.Priority)
		schemaDocs[i] = schema.Document{
			PageContent: c.Content,
			Metadata:    meta,
		}
	}
	opts := []interfaces.VectorStoreOption{
		func(o *interfaces.VectorStoreOptions) { o.Namespace = ns },
	}
	newIDs, err := store.AddDocuments(ctx, schemaDocs, opts...)
	if err != nil {
		return
	}
	for i, c := range allChunks {
		vid := ""
		if i < len(newIDs) {
			vid = newIDs[i]
		}
		global.LRAG_DB.Model(&c).Update("vector_store_id", vid)
	}
}
