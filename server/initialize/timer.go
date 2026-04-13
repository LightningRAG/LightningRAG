package initialize

import (
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/task"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())
		_, err := global.LRAG_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			if err := task.ClearTable(global.LRAG_DB); err != nil {
				global.LRAG_LOG.Error("ClearDB task failed", zap.Error(err))
			}
		}, "定时清理数据库【日志，黑名单】内容", option...)
		if err != nil {
			global.LRAG_LOG.Error("failed to register ClearDB timer", zap.Error(err))
		}

		_, err = global.LRAG_Timer.AddTaskByFunc("RagChannelWebhookEventsPrune", "@daily", func() {
			task.PruneRagChannelWebhookEvents()
		}, "RAG 第三方渠道 Webhook 幂等表按天清理", option...)
		if err != nil {
			global.LRAG_LOG.Error("failed to register webhook events prune timer", zap.Error(err))
		}

		poll := global.LRAG_CONFIG.Rag.ChannelOutboundPollSeconds
		if poll != -1 {
			if poll <= 0 {
				poll = 30
			}
			if poll < 5 {
				poll = 5
			}
			if poll > 300 {
				poll = 300
			}
			spec := fmt.Sprintf("*/%d * * * * *", poll)
			_, err = global.LRAG_Timer.AddTaskByFuncWithSecond("RagChannel", spec, func() {
				task.ProcessRagChannelOutbound()
			}, "RAG 第三方渠道出站重试队列", option...)
			if err != nil {
				global.LRAG_LOG.Error("failed to register channel outbound timer", zap.Error(err))
			}
		}

		// 其他定时任务定在这里 参考上方使用方法

		//_, err := global.LRAG_Timer.AddTaskByFunc("定时任务标识", "corn表达式", func() {
		//	具体执行内容...
		//  ......
		//}, option...)
		//if err != nil {
		//	fmt.Println("add timer error:", err)
		//}
	}()
}
