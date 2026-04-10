package rag

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/registry"
	"go.uber.org/zap"
)

var (
	errNoOCRProvider         = errors.New("未配置 OCR 模型，请在模型管理中添加支持 ocr 场景的模型")
	errNoCVProvider          = errors.New("未配置 CV 模型，请在模型管理中添加支持 cv 场景的模型")
	errNoSpeech2TextProvider = errors.New("未配置 Speech2Text 模型，请在模型管理中添加支持 speech2text 场景的模型")
)

// resolveModelProvider 通用模型解析，完整回退链：KB配置 → 用户默认 → 系统全局默认 → 全局搜索
func resolveModelProvider(ctx context.Context, uid uint, modelID uint, modelSource string, modelType string) (provider, modelName, baseURL, apiKey string, ok bool) {
	// 使用 ResolveModelWithFallback，authorityId 传 0 因为文档解析场景不涉及角色
	return ResolveModelWithFallback(ctx, uid, 0, modelID, modelSource, modelType)
}

// resolveOCRProvider 解析 OCR 模型（兼容旧接口）
func resolveOCRProvider(ctx context.Context, uid uint) (provider, modelName, baseURL, apiKey string, ok bool) {
	return resolveModelProvider(ctx, uid, 0, "", interfaces.ModelTypeOCR)
}

// resolveOCRProviderFromKB 从知识库配置解析 OCR 模型
func resolveOCRProviderFromKB(ctx context.Context, uid uint, kb *rag.RagKnowledgeBase) (provider, modelName, baseURL, apiKey string, ok bool) {
	if kb != nil {
		return resolveModelProvider(ctx, uid, kb.OCRID, kb.OCRSource, interfaces.ModelTypeOCR)
	}
	return resolveModelProvider(ctx, uid, 0, "", interfaces.ModelTypeOCR)
}

// resolveCVProviderFromKB 从知识库配置解析 CV 模型
func resolveCVProviderFromKB(ctx context.Context, uid uint, kb *rag.RagKnowledgeBase) (provider, modelName, baseURL, apiKey string, ok bool) {
	if kb != nil {
		return resolveModelProvider(ctx, uid, kb.ImageDescriptionID, kb.ImageDescriptionSource, interfaces.ModelTypeCV)
	}
	return resolveModelProvider(ctx, uid, 0, "", interfaces.ModelTypeCV)
}

// resolveSpeech2TextProviderFromKB 从知识库配置解析 Speech2Text 模型
func resolveSpeech2TextProviderFromKB(ctx context.Context, uid uint, kb *rag.RagKnowledgeBase) (provider, modelName, baseURL, apiKey string, ok bool) {
	if kb != nil {
		return resolveModelProvider(ctx, uid, kb.Speech2TextID, kb.Speech2TextSource, interfaces.ModelTypeSpeech2Txt)
	}
	return resolveModelProvider(ctx, uid, 0, "", interfaces.ModelTypeSpeech2Txt)
}

// ParseImageWithOCR 使用 OCR 从图片字节中提取文本（兼容旧接口）
func ParseImageWithOCR(ctx context.Context, data []byte, filename string, uid uint) (string, error) {
	return ParseImageWithOCRFromKB(ctx, data, filename, uid, nil)
}

// ParseImageWithOCRFromKB 使用知识库配置的 OCR 模型从图片/PDF 中提取文本
func ParseImageWithOCRFromKB(ctx context.Context, data []byte, filename string, uid uint, kb *rag.RagKnowledgeBase) (string, error) {
	provider, modelName, baseURL, apiKey, ok := resolveOCRProviderFromKB(ctx, uid, kb)
	if !ok {
		return "", errNoOCRProvider
	}
	ocrInst, err := registry.CreateOCR(registry.OCRConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
	if err != nil || ocrInst == nil {
		return "", fmt.Errorf("创建 OCR 模型失败(%s/%s): %v", provider, modelName, err)
	}
	result, err := ocrInst.ExtractText(ctx, data, filename)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	global.LRAG_LOG.Info("OCR 提取文本完成",
		zap.String("provider", provider),
		zap.String("model", modelName),
		zap.Int("textLen", len(result.Text)))
	return result.Text, nil
}

// ParseImageContent 从图片文件解析文本内容（兼容旧接口）
func ParseImageContent(ctx context.Context, r io.Reader, filename string, uid uint) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return ParseImageWithOCR(ctx, data, filename, uid)
}

// ParseImageContentFromKB 从图片解析内容，支持 OCR 提取文本 + CV 图片描述
func ParseImageContentFromKB(ctx context.Context, data []byte, filename string, uid uint, kb *rag.RagKnowledgeBase) (string, error) {
	var parts []string

	// OCR 提取文本
	useOCR := kb == nil || kb.UseOCR
	if useOCR {
		ocrText, ocrErr := ParseImageWithOCRFromKB(ctx, data, filename, uid, kb)
		if ocrErr != nil {
			global.LRAG_LOG.Warn("图片 OCR 提取失败", zap.String("file", filename), zap.Error(ocrErr))
		} else if ocrText != "" {
			parts = append(parts, ocrText)
		}
	}

	// CV 图片描述
	useCV := kb != nil && kb.UseImageDescription
	if useCV {
		cvText, cvErr := DescribeImageFromKB(ctx, data, uid, kb)
		if cvErr != nil {
			global.LRAG_LOG.Warn("图片 CV 描述失败", zap.String("file", filename), zap.Error(cvErr))
		} else if cvText != "" {
			parts = append(parts, cvText)
		}
	}

	if len(parts) == 0 {
		if useOCR {
			return "", errNoOCRProvider
		}
		return "", fmt.Errorf("图片解析未提取到任何内容")
	}
	return strings.Join(parts, "\n\n"), nil
}

// DescribeImageFromKB 使用 CV 模型描述图片内容
func DescribeImageFromKB(ctx context.Context, data []byte, uid uint, kb *rag.RagKnowledgeBase) (string, error) {
	provider, modelName, baseURL, apiKey, ok := resolveCVProviderFromKB(ctx, uid, kb)
	if !ok {
		return "", errNoCVProvider
	}
	cvInst, err := registry.CreateCV(registry.CVConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
	if err != nil || cvInst == nil {
		return "", fmt.Errorf("创建 CV 模型失败(%s/%s): %v", provider, modelName, err)
	}
	desc, err := cvInst.DescribeWithPrompt(ctx, data, "请详细描述这张图片的内容，包括文字、表格、图表等信息。如果图片包含文字，请完整提取。")
	if err != nil {
		return "", err
	}
	global.LRAG_LOG.Info("CV 图片描述完成",
		zap.String("provider", provider),
		zap.String("model", modelName),
		zap.Int("descLen", len(desc)))
	return desc, nil
}

// ParseAudioContent 使用 Speech2Text 模型将音频转为文字
func ParseAudioContent(ctx context.Context, data []byte, filename string, uid uint, kb *rag.RagKnowledgeBase) (string, error) {
	useSpeech := kb == nil || kb.UseSpeech2Text
	if !useSpeech {
		return "", fmt.Errorf("知识库未启用语音转文字功能，请在知识库设置中启用")
	}
	provider, modelName, baseURL, apiKey, ok := resolveSpeech2TextProviderFromKB(ctx, uid, kb)
	if !ok {
		return "", errNoSpeech2TextProvider
	}
	s2tInst, err := registry.CreateSpeech2Text(registry.Speech2TextConfig{
		Provider:  provider,
		ModelName: modelName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
	})
	if err != nil || s2tInst == nil {
		return "", fmt.Errorf("创建 Speech2Text 模型失败(%s/%s): %v", provider, modelName, err)
	}
	text, err := s2tInst.Transcribe(ctx, data)
	if err != nil {
		return "", fmt.Errorf("音频转文字失败: %w", err)
	}
	global.LRAG_LOG.Info("Speech2Text 转写完成",
		zap.String("provider", provider),
		zap.String("model", modelName),
		zap.String("file", filename),
		zap.Int("textLen", len(text)))
	return text, nil
}
