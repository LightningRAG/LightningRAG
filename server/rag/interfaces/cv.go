// Package interfaces 定义 CV 计算机视觉接口，参考 references 目录内 cv_model
package interfaces

import "context"

// CV 计算机视觉接口，参考 references 目录内 cv_model.Base
// 用于图像理解、描述、识别等
type CV interface {
	// Describe 描述图像内容
	Describe(ctx context.Context, image []byte) (string, error)

	// DescribeWithPrompt 按指定 prompt 描述图像
	DescribeWithPrompt(ctx context.Context, image []byte, prompt string) (string, error)

	// ProviderName 返回提供商名称
	ProviderName() string

	// ModelName 返回模型名称
	ModelName() string
}
