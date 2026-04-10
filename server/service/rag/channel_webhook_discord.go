package rag

import (
	"context"
	"errors"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/channel"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/rag"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *ChannelConnectorService) discordDeferredFollowUp(ctx context.Context, conn *rag.RagChannelConnector, dd *channel.DiscordDeferredInteraction) error {
	threadKey := dd.ChannelID + ":" + dd.UserID
	text := strings.TrimSpace(dd.Query)
	patchErr := func(msg string) error {
		if derr := channel.DiscordEditOriginalInteraction(ctx, dd.ApplicationID, dd.InteractionToken, msg); derr != nil {
			_ = s.enqueueDiscordDeferredEdit(ctx, conn.ID, msg, dd.ApplicationID, dd.InteractionToken)
		}
		return errors.New(msg)
	}

	var convID uint
	var sess rag.RagChannelSession
	err := global.LRAG_DB.Where("connector_id = ? AND thread_key = ?", conn.ID, threadKey).First(&sess).Error
	if err == nil {
		convID = sess.ConversationID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return patchErr("会话查询失败: " + err.Error())
	}

	globals := map[string]any{
		"channel":      conn.Channel,
		"connector_id": conn.ID,
	}
	out, newConvID, err := (&AgentService{}).Run(ctx, conn.OwnerID, conn.AuthorityID, conn.AgentID, nil, text, convID, globals)
	if err != nil {
		failMsg := "处理失败：" + truncateForDiscord(err.Error())
		if derr := channel.DiscordEditOriginalInteraction(ctx, dd.ApplicationID, dd.InteractionToken, failMsg); derr != nil {
			_ = s.enqueueDiscordDeferredEdit(ctx, conn.ID, failMsg, dd.ApplicationID, dd.InteractionToken)
		}
		return err
	}

	if convID == 0 && newConvID > 0 {
		ns := rag.RagChannelSession{
			ConnectorID:    conn.ID,
			ThreadKey:      threadKey,
			ConversationID: newConvID,
		}
		if err := global.LRAG_DB.WithContext(ctx).Create(&ns).Error; err != nil {
			var existing rag.RagChannelSession
			if err2 := global.LRAG_DB.Where("connector_id = ? AND thread_key = ?", conn.ID, threadKey).First(&existing).Error; err2 != nil {
				global.LRAG_LOG.Warn("discord session mapping failed", zap.Error(err))
			}
		}
	}

	reply := extractChannelAgentContent(out)
	if err := channel.DiscordEditOriginalInteraction(ctx, dd.ApplicationID, dd.InteractionToken, reply); err != nil {
		global.LRAG_LOG.Warn("discord follow-up failed", zap.Error(err))
		_ = s.enqueueDiscordDeferredEdit(ctx, conn.ID, reply, dd.ApplicationID, dd.InteractionToken)
		return err
	}
	return nil
}
