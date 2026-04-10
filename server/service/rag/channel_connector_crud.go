package rag

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"gorm.io/gorm"
)

func randomWebhookSecret() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

// validateChannelConnectorExtraJSON 非空时须为 JSON 对象（{}），避免入库后 Webhook 侧 json.Unmarshal 静默失败。
func validateChannelConnectorExtraJSON(extra string) error {
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(extra), &m); err != nil {
		return fmt.Errorf("渠道扩展 extra 须为合法 JSON 对象: %w", err)
	}
	if m == nil {
		return fmt.Errorf("渠道扩展 extra 须为非 null 的 JSON 对象，空配置请留空或填写 {}")
	}
	return nil
}

// CreateChannelConnector 创建连接器；若未传 WebhookSecret 则生成并仅在返回体中暴露一次
func (s *ChannelConnectorService) CreateChannelConnector(ctx context.Context, uid, authorityID uint, req request.ChannelConnectorCreate) (*rag.RagChannelConnector, string, error) {
	var agent rag.RagAgent
	if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", req.AgentID, uid).First(&agent).Error; err != nil {
		return nil, "", err
	}
	secret := strings.TrimSpace(req.WebhookSecret)
	plainOnce := ""
	if secret == "" {
		var err error
		secret, err = randomWebhookSecret()
		if err != nil {
			return nil, "", err
		}
		plainOnce = secret
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	extraNorm := strings.TrimSpace(req.Extra)
	if err := validateChannelConnectorExtraJSON(extraNorm); err != nil {
		return nil, "", err
	}
	conn := &rag.RagChannelConnector{
		OwnerID:       uid,
		AuthorityID:   authorityID,
		Name:          req.Name,
		Channel:       strings.TrimSpace(req.Channel),
		AgentID:       req.AgentID,
		WebhookSecret: secret,
		Enabled:       enabled,
		Extra:         extraNorm,
	}
	if _, err := channel.Lookup(conn.Channel); err != nil {
		return nil, "", fmt.Errorf("不支持的渠道: %w", err)
	}
	if err := global.LRAG_DB.WithContext(ctx).Create(conn).Error; err != nil {
		return nil, "", err
	}
	return conn, plainOnce, nil
}

func (s *ChannelConnectorService) UpdateChannelConnector(ctx context.Context, uid uint, req request.ChannelConnectorUpdate) error {
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", req.ID, uid).First(&conn).Error; err != nil {
		return err
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AgentID > 0 {
		var agent rag.RagAgent
		if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", req.AgentID, uid).First(&agent).Error; err != nil {
			return err
		}
		updates["agent_id"] = req.AgentID
	}
	if req.WebhookSecret != "" {
		updates["webhook_secret"] = strings.TrimSpace(req.WebhookSecret)
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if strings.TrimSpace(req.Extra) != "" {
		extraNorm := strings.TrimSpace(req.Extra)
		if err := validateChannelConnectorExtraJSON(extraNorm); err != nil {
			return err
		}
		updates["extra"] = extraNorm
	}
	if len(updates) == 0 {
		return nil
	}
	return global.LRAG_DB.WithContext(ctx).Model(&rag.RagChannelConnector{}).Where("id = ? AND owner_id = ?", req.ID, uid).Updates(updates).Error
}

func (s *ChannelConnectorService) ListChannelConnectors(ctx context.Context, uid uint, req request.ChannelConnectorList) ([]rag.RagChannelConnector, int64, error) {
	db := global.LRAG_DB.Model(&rag.RagChannelConnector{}).Where("owner_id = ?", uid)
	if req.Channel != "" {
		db = db.Where("channel = ?", req.Channel)
	}
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	var list []rag.RagChannelConnector
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *ChannelConnectorService) GetChannelConnector(ctx context.Context, uid, id uint) (*rag.RagChannelConnector, error) {
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", id, uid).First(&conn).Error; err != nil {
		return nil, err
	}
	return &conn, nil
}

func (s *ChannelConnectorService) DeleteChannelConnector(ctx context.Context, uid, id uint) error {
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.Where("id = ? AND owner_id = ?", id, uid).First(&conn).Error; err != nil {
		return err
	}
	return global.LRAG_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("connector_id = ?", id).Delete(&rag.RagChannelWebhookEvent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("connector_id = ?", id).Delete(&rag.RagChannelSession{}).Error; err != nil {
			return err
		}
		return tx.Delete(&conn).Error
	})
}
