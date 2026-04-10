package initialize

import (
	"fmt"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/task"
	"github.com/robfig/cron/v3"
)

func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())
		// 清理DB定时任务
		_, err := global.LRAG_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			err := task.ClearTable(global.LRAG_DB) // 定时任务方法定在task文件包中
			if err != nil {
				fmt.Println("timer error:", err)
			}
		}, "定时清理数据库【日志，黑名单】内容", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}

		_, err = global.LRAG_Timer.AddTaskByFunc("RagChannelWebhookEventsPrune", "@daily", func() {
			task.PruneRagChannelWebhookEvents()
		}, "RAG 第三方渠道 Webhook 幂等表按天清理", option...)
		if err != nil {
			fmt.Println("add rag channel webhook events prune timer error:", err)
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
				fmt.Println("add rag channel outbound timer error:", err)
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
