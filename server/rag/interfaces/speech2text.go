// Package interfaces 定义 Speech2Text 语音转文字接口，参考 references 目录内 sequence2txt_model
package interfaces

import "context"

// Speech2Text 语音转文字接口，参考 references 目录内 sequence2txt_model.Base
type Speech2Text interface {
	// Transcribe 将音频转为文字，audio 为音频文件路径或字节
	Transcribe(ctx context.Context, audio interface{}) (string, error)

	// ProviderName 返回提供商名称
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string
}
