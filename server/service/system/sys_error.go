package system

import (
	"context"
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	"go.uber.org/zap"
)

type SysErrorService struct{}

// CreateSysError 创建错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) CreateSysError(ctx context.Context, sysError *system.SysError) (err error) {
	if global.LRAG_DB == nil {
		return nil
	}
	err = global.LRAG_DB.Create(sysError).Error
	return err
}

// DeleteSysError 删除错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) DeleteSysError(ctx context.Context, ID string) (err error) {
	err = global.LRAG_DB.Delete(&system.SysError{}, "id = ?", ID).Error
	return err
}

// DeleteSysErrorByIds 批量删除错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) DeleteSysErrorByIds(ctx context.Context, IDs []string) (err error) {
	err = global.LRAG_DB.Delete(&[]system.SysError{}, "id in ?", IDs).Error
	return err
}

// UpdateSysError 更新错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) UpdateSysError(ctx context.Context, sysError system.SysError) (err error) {
	err = global.LRAG_DB.Model(&system.SysError{}).Where("id = ?", sysError.ID).Updates(&sysError).Error
	return err
}

// GetSysError 根据ID获取错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) GetSysError(ctx context.Context, ID string) (sysError system.SysError, err error) {
	err = global.LRAG_DB.Where("id = ?", ID).First(&sysError).Error
	return
}

// GetSysErrorInfoList 分页获取错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) GetSysErrorInfoList(ctx context.Context, info systemReq.SysErrorSearch) (list []system.SysError, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.LRAG_DB.Model(&system.SysError{}).Order("created_at desc")
	var sysErrors []system.SysError
	// 如果有条件搜索 下方会自动创建搜索语句
	if len(info.CreatedAtRange) == 2 {
		db = db.Where("created_at BETWEEN ? AND ?", info.CreatedAtRange[0], info.CreatedAtRange[1])
	}

	if info.Form != nil && *info.Form != "" {
		db = db.Where("form = ?", *info.Form)
	}
	if info.Info != nil && *info.Info != "" {
		db = db.Where("info LIKE ?", "%"+*info.Info+"%")
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&sysErrors).Error
	return sysErrors, total, err
}

// GetSysErrorSolution 异步处理错误
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) GetSysErrorSolution(ctx context.Context, ID string) (err error) {
	// 立即更新为处理中
	err = global.LRAG_DB.WithContext(ctx).Model(&system.SysError{}).Where("id = ?", ID).Update("status", "处理中").Error
	if err != nil {
		return err
	}

	// 异步协程在一分钟后更新为处理完成
	go func(id string) {
		defer func() {
			if r := recover(); r != nil {
				global.LRAG_LOG.Error("panic in GetSysErrorSolution goroutine", zap.Any("recover", r), zap.String("id", id))
			}
		}()

		var se system.SysError
		if err := global.LRAG_DB.Model(&system.SysError{}).Where("id = ?", id).First(&se).Error; err != nil {
			global.LRAG_LOG.Warn("failed to load sys error for solution", zap.String("id", id), zap.Error(err))
		}

		var form, info string
		if se.Form != nil {
			form = *se.Form
		}
		if se.Info != nil {
			info = *se.Info
		}

		llmReq := common.JSONMap{
			"mode": "solution",
			"info": info,
			"form": form,
		}

		var solution string
		if data, err := (&AutoCodeService{}).LLMAuto(context.Background(), llmReq); err == nil {
			if m, ok := data.(map[string]interface{}); ok {
				solution = fmt.Sprintf("%v", m["text"])
			}
			if uerr := global.LRAG_DB.Model(&system.SysError{}).Where("id = ?", id).Updates(map[string]interface{}{"status": "处理完成", "solution": solution}).Error; uerr != nil {
				global.LRAG_LOG.Warn("failed to update error status", zap.String("id", id), zap.String("status", "处理完成"), zap.Error(uerr))
			}
		} else {
			if uerr := global.LRAG_DB.Model(&system.SysError{}).Where("id = ?", id).Update("status", "处理失败").Error; uerr != nil {
				global.LRAG_LOG.Warn("failed to update error status", zap.String("id", id), zap.String("status", "处理失败"), zap.Error(uerr))
			}
		}
	}(ID)

	return nil
}
