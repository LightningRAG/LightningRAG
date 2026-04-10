package initialize

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/example"
	"gorm.io/gorm"
)

func bizModelWithDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	return db.AutoMigrate(&example.ExaCustomer{}, &example.ExaFileUploadAndDownload{}, &example.ExaAttachmentCategory{}, &example.ExaFile{}, &example.ExaFileChunk{})
}

func bizModel() error {
	return bizModelWithDB(global.LRAG_DB)
}
