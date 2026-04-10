package response

import "github.com/LightningRAG/LightningRAG/server/model/example"

type ExaFileResponse struct {
	File example.ExaFileUploadAndDownload `json:"file"`
}
