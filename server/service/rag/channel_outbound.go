package rag

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"github.com/LightningRAG/LightningRAG/server/model/rag/request"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// outboundKindDiscordDeferredEdit Discord 斜杠延迟响应：PATCH @original
const outboundKindDiscordDeferredEdit = "discord_deferred_edit"

type channelOutboundPayload struct {
	Kind   string         `json:"kind,omitempty"`
	Text   string         `json:"text"`
	Opaque map[string]any `json:"opaque,omitempty"`
}

// ChannelOutboundListRow 管理端列表展示
type ChannelOutboundListRow struct {
	ID            uint       `json:"id"`
	ConnectorID   uint       `json:"connectorId"`
	ConnectorName string     `json:"connectorName"`
	Channel       string     `json:"channel"`
	Kind          string     `json:"kind"`
	TextPreview   string     `json:"textPreview"`
	Attempts      int        `json:"attempts"`
	NextRetryAt   time.Time  `json:"nextRetryAt"`
	LeaseUntil    *time.Time `json:"leaseUntil,omitempty"`
	LastErr       string     `json:"lastErr"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func outboundRetryDelayAfterFailure(attempts int) time.Duration {
	switch attempts {
	case 1:
		return 15 * time.Second
	case 2:
		return 30 * time.Second
	case 3:
		return time.Minute
	case 4:
		return 2 * time.Minute
	case 5:
		return 5 * time.Minute
	case 6:
		return 10 * time.Minute
	case 7:
		return 20 * time.Minute
	default:
		return 30 * time.Minute
	}
}

func truncateChannelErr(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 1000 {
		return s
	}
	return s[:1000] + "…"
}

func effectiveChannelOutboundMaxAttempts() int {
	n := global.LRAG_CONFIG.Rag.ChannelOutboundMaxAttempts
	if n <= 0 {
		return 8
	}
	if n > 24 {
		return 24
	}
	return n
}

func effectiveChannelOutboundBatchSize(limitArg int) int {
	n := global.LRAG_CONFIG.Rag.ChannelOutboundBatchSize
	if n <= 0 {
		n = 32
	}
	if n > 128 {
		n = 128
	}
	if limitArg > 0 && limitArg < n {
		return limitArg
	}
	return n
}

func effectiveChannelOutboundClaimLease() time.Duration {
	sec := global.LRAG_CONFIG.Rag.ChannelOutboundClaimLeaseSeconds
	if sec <= 0 {
		sec = 180
	}
	if sec < 30 {
		sec = 30
	}
	if sec > 3600 {
		sec = 3600
	}
	return time.Duration(sec) * time.Second
}

func outboundDialectorName() string {
	if global.LRAG_DB == nil || global.LRAG_DB.Dialector == nil {
		return ""
	}
	return global.LRAG_DB.Dialector.Name()
}

// claimChannelOutboundRows 认领一批到期任务并写入 lease_until（MySQL/PostgreSQL 用 SKIP LOCKED；其余方言用逐行 CAS）。
func (s *ChannelConnectorService) claimChannelOutboundRows(ctx context.Context, maxAtt, limit int) ([]rag.RagChannelOutbound, error) {
	now := time.Now()
	claimUntil := now.Add(effectiveChannelOutboundClaimLease())
	switch outboundDialectorName() {
	case "mysql", "postgres":
		return s.claimChannelOutboundRowsSkipLocked(ctx, maxAtt, limit, claimUntil, now)
	default:
		return s.claimChannelOutboundRowsCAS(ctx, maxAtt, limit, claimUntil, now)
	}
}

func (s *ChannelConnectorService) claimChannelOutboundRowsSkipLocked(ctx context.Context, maxAtt, limit int, claimUntil, now time.Time) ([]rag.RagChannelOutbound, error) {
	var rows []rag.RagChannelOutbound
	err := global.LRAG_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&rag.RagChannelOutbound{}).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("next_retry_at <= ? AND attempts < ? AND (lease_until IS NULL OR lease_until < ?)", now, maxAtt, now).
			Order("id ASC").
			Limit(limit)
		if err := q.Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		ids := make([]uint, len(rows))
		for i := range rows {
			ids[i] = rows[i].ID
		}
		return tx.Model(&rag.RagChannelOutbound{}).Where("id IN ?", ids).Update("lease_until", claimUntil).Error
	})
	return rows, err
}

func (s *ChannelConnectorService) claimChannelOutboundRowsCAS(ctx context.Context, maxAtt, limit int, claimUntil, now time.Time) ([]rag.RagChannelOutbound, error) {
	var candidates []rag.RagChannelOutbound
	if err := global.LRAG_DB.WithContext(ctx).
		Model(&rag.RagChannelOutbound{}).
		Where("next_retry_at <= ? AND attempts < ? AND (lease_until IS NULL OR lease_until < ?)", now, maxAtt, now).
		Order("id ASC").
		Limit(limit).
		Find(&candidates).Error; err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	db := global.LRAG_DB.WithContext(ctx)
	var claimed []rag.RagChannelOutbound
	for i := range candidates {
		r := candidates[i]
		t := time.Now()
		res := db.Model(&rag.RagChannelOutbound{}).
			Where("id = ? AND (lease_until IS NULL OR lease_until < ?)", r.ID, t).
			Update("lease_until", claimUntil)
		if res.Error != nil {
			return claimed, res.Error
		}
		if res.RowsAffected == 1 {
			claimed = append(claimed, r)
		}
	}
	return claimed, nil
}

func (s *ChannelConnectorService) enqueueChannelOutboundPayload(ctx context.Context, connectorID uint, channelName string, pl channelOutboundPayload) error {
	pl.Text = strings.TrimSpace(pl.Text)
	if pl.Text == "" {
		return nil
	}
	raw, err := json.Marshal(pl)
	if err != nil {
		return err
	}
	row := rag.RagChannelOutbound{
		ConnectorID: connectorID,
		Channel:     strings.TrimSpace(channelName),
		Payload:     string(raw),
		Attempts:    0,
		NextRetryAt: time.Now(),
	}
	return global.LRAG_DB.WithContext(ctx).Create(&row).Error
}

// enqueueChannelOutbound SendReply 同步失败后写入队列，由定时任务重试。
func (s *ChannelConnectorService) enqueueChannelOutbound(ctx context.Context, connectorID uint, channelName, text string, opaque map[string]any) error {
	return s.enqueueChannelOutboundPayload(ctx, connectorID, channelName, channelOutboundPayload{Text: text, Opaque: opaque})
}

// enqueueDiscordDeferredEdit Discord PATCH @original 失败后入队（与 Adapter.SendReply 不同路径）。
func (s *ChannelConnectorService) enqueueDiscordDeferredEdit(ctx context.Context, connectorID uint, text, applicationID, interactionToken string) error {
	text = strings.TrimSpace(text)
	applicationID = strings.TrimSpace(applicationID)
	interactionToken = strings.TrimSpace(interactionToken)
	if text == "" || applicationID == "" || interactionToken == "" {
		return nil
	}
	pl := channelOutboundPayload{
		Kind: outboundKindDiscordDeferredEdit,
		Text: text,
		Opaque: map[string]any{
			"discord_application_id":    applicationID,
			"discord_interaction_token": interactionToken,
		},
	}
	return s.enqueueChannelOutboundPayload(ctx, connectorID, "discord", pl)
}

// ProcessChannelOutboundQueue 处理到期的出站重试（供定时任务调用）。
func (s *ChannelConnectorService) ProcessChannelOutboundQueue(ctx context.Context, limit int) (int, error) {
	if global.LRAG_DB == nil {
		return 0, nil
	}
	maxAtt := effectiveChannelOutboundMaxAttempts()
	limit = effectiveChannelOutboundBatchSize(limit)
	if limit <= 0 {
		return 0, nil
	}
	rows, err := s.claimChannelOutboundRows(ctx, maxAtt, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for i := range rows {
		row := &rows[i]
		if err := s.processOneChannelOutbound(ctx, row); err != nil {
			global.LRAG_LOG.Warn("channel outbound row failed", zap.Uint("id", row.ID), zap.Error(err))
		}
		n++
	}
	return n, nil
}

func (s *ChannelConnectorService) processOneChannelOutbound(ctx context.Context, row *rag.RagChannelOutbound) error {
	var pl channelOutboundPayload
	if err := json.Unmarshal([]byte(row.Payload), &pl); err != nil {
		return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
	}
	var conn rag.RagChannelConnector
	if err := global.LRAG_DB.WithContext(ctx).Where("id = ?", row.ConnectorID).First(&conn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
		}
		return err
	}
	if !conn.Enabled {
		return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
	}

	var sendErr error
	switch pl.Kind {
	case outboundKindDiscordDeferredEdit:
		if conn.Channel != "discord" || pl.Opaque == nil {
			return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
		}
		appID, _ := pl.Opaque["discord_application_id"].(string)
		tok, _ := pl.Opaque["discord_interaction_token"].(string)
		appID, tok = strings.TrimSpace(appID), strings.TrimSpace(tok)
		if appID == "" || tok == "" {
			return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
		}
		sendErr = channel.DiscordEditOriginalInteraction(ctx, appID, tok, pl.Text)
	default:
		extra := map[string]any{}
		if strings.TrimSpace(conn.Extra) != "" {
			_ = json.Unmarshal([]byte(conn.Extra), &extra)
		}
		cfg := &channel.ConnectorConfig{Channel: conn.Channel, Extra: extra}
		ad, err := channel.Lookup(conn.Channel)
		if err != nil {
			return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
		}
		ref := channel.ThreadRef{Opaque: pl.Opaque}
		sendErr = ad.SendReply(ctx, conn.WebhookSecret, cfg, ref, pl.Text)
	}
	if sendErr == nil {
		return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
	}

	row.Attempts++
	row.LastErr = truncateChannelErr(sendErr.Error())
	if row.Attempts >= effectiveChannelOutboundMaxAttempts() {
		global.LRAG_LOG.Warn("channel outbound abandoned",
			zap.Uint("connector", conn.ID),
			zap.String("channel", conn.Channel),
			zap.Int("attempts", row.Attempts),
			zap.String("err", row.LastErr))
		return global.LRAG_DB.WithContext(ctx).Delete(row, row.ID).Error
	}
	row.NextRetryAt = time.Now().Add(outboundRetryDelayAfterFailure(row.Attempts))
	return global.LRAG_DB.WithContext(ctx).Model(row).Updates(map[string]interface{}{
		"attempts":      row.Attempts,
		"next_retry_at": row.NextRetryAt,
		"last_err":      row.LastErr,
		"lease_until":   nil,
	}).Error
}

// ListChannelOutboundQueue 当前用户名下连接器关联的出站重试任务。
func (s *ChannelConnectorService) ListChannelOutboundQueue(ctx context.Context, uid uint, req request.ChannelOutboundList) ([]ChannelOutboundListRow, int64, error) {
	var connIDs []uint
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChannelConnector{}).Where("owner_id = ?", uid).Pluck("id", &connIDs).Error; err != nil {
		return nil, 0, err
	}
	if len(connIDs) == 0 {
		return nil, 0, nil
	}
	db := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChannelOutbound{}).Where("connector_id IN ?", connIDs)
	if req.ConnectorID > 0 {
		var own int64
		if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChannelConnector{}).
			Where("id = ? AND owner_id = ?", req.ConnectorID, uid).Count(&own).Error; err != nil {
			return nil, 0, err
		}
		if own == 0 {
			return nil, 0, nil
		}
		db = db.Where("connector_id = ?", req.ConnectorID)
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
	var rows []rag.RagChannelOutbound
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	var conns []rag.RagChannelConnector
	_ = global.LRAG_DB.WithContext(ctx).Select("id", "name").Where("owner_id = ?", uid).Find(&conns).Error
	nameByID := make(map[uint]string, len(conns))
	for i := range conns {
		nameByID[conns[i].ID] = conns[i].Name
	}
	out := make([]ChannelOutboundListRow, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		var pl channelOutboundPayload
		_ = json.Unmarshal([]byte(r.Payload), &pl)
		preview := pl.Text
		if len(preview) > 120 {
			preview = preview[:120] + "…"
		}
		kind := pl.Kind
		if kind == "" {
			kind = "send_reply"
		}
		out = append(out, ChannelOutboundListRow{
			ID:            r.ID,
			ConnectorID:   r.ConnectorID,
			ConnectorName: nameByID[r.ConnectorID],
			Channel:       r.Channel,
			Kind:          kind,
			TextPreview:   preview,
			Attempts:      r.Attempts,
			NextRetryAt:   r.NextRetryAt,
			LeaseUntil:    r.LeaseUntil,
			LastErr:       r.LastErr,
			CreatedAt:     r.CreatedAt,
		})
	}
	return out, total, nil
}

// DeleteChannelOutboundRow 删除一条出站任务（仅当所属连接器归当前用户）。
func (s *ChannelConnectorService) DeleteChannelOutboundRow(ctx context.Context, uid, id uint) error {
	var row rag.RagChannelOutbound
	if err := global.LRAG_DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return err
	}
	var own int64
	if err := global.LRAG_DB.WithContext(ctx).Model(&rag.RagChannelConnector{}).
		Where("id = ? AND owner_id = ?", row.ConnectorID, uid).Count(&own).Error; err != nil {
		return err
	}
	if own == 0 {
		return gorm.ErrRecordNotFound
	}
	return global.LRAG_DB.WithContext(ctx).Delete(&row, id).Error
}
