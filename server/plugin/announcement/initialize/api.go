package initialize

import (
	"context"
	model "github.com/LightningRAG/LightningRAG/server/model/system"
	"github.com/LightningRAG/LightningRAG/server/plugin/plugin-tool/utils"
)

func Api(ctx context.Context) {
	entities := []model.SysApi{
		{
			Path:        "/info/createInfo",
			Description: "Create announcement",
			ApiGroup:    "Announcement",
			Method:      "POST",
		},
		{
			Path:        "/info/deleteInfo",
			Description: "Delete announcement",
			ApiGroup:    "Announcement",
			Method:      "DELETE",
		},
		{
			Path:        "/info/deleteInfoByIds",
			Description: "Batch delete announcements",
			ApiGroup:    "Announcement",
			Method:      "DELETE",
		},
		{
			Path:        "/info/updateInfo",
			Description: "Update announcement",
			ApiGroup:    "Announcement",
			Method:      "PUT",
		},
		{
			Path:        "/info/findInfo",
			Description: "Get announcement by ID",
			ApiGroup:    "Announcement",
			Method:      "GET",
		},
		{
			Path:        "/info/getInfoList",
			Description: "Get announcement list",
			ApiGroup:    "Announcement",
			Method:      "GET",
		},
	}
	utils.RegisterApis(entities...)
}
