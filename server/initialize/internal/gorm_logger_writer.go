package internal

import (
	"fmt"
	"github.com/LightningRAG/LightningRAG/server/config"
	"github.com/LightningRAG/LightningRAG/server/global"
	"gorm.io/gorm/logger"
)

type Writer struct {
	config config.GeneralDB
	writer logger.Writer
}

func NewWriter(config config.GeneralDB) *Writer {
	return &Writer{config: config}
}

// Printf 格式化打印日志
func (c *Writer) Printf(message string, data ...any) {

	// 当有日志时候均需要输出到控制台
	fmt.Printf(message, data...)

	// 当开启了zap的情况，会打印到日志记录
	if c.config.LogZap {
		switch c.config.LogLevel() {
		case logger.Silent:
			global.LRAG_LOG.Debug(fmt.Sprintf(message, data...))
		case logger.Error:
			global.LRAG_LOG.Error(fmt.Sprintf(message, data...))
		case logger.Warn:
			global.LRAG_LOG.Warn(fmt.Sprintf(message, data...))
		case logger.Info:
			global.LRAG_LOG.Info(fmt.Sprintf(message, data...))
		default:
			global.LRAG_LOG.Info(fmt.Sprintf(message, data...))
		}
		return
	}
}
