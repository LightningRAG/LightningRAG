// Package interfaces 定义 TTS 文字转语音接口，参考 references 目录内 tts_model
package interfaces

import (
	"context"
	"io"
)

// TTS 文字转语音接口，参考 references 目录内 tts_model.Base
type TTS interface {
	// Synthesize 将文字转为语音，返回音频流
	Synthesize(ctx context.Context, text string, voice string) (io.ReadCloser, error)

	// ProviderName 返回提供商名称
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string
}
