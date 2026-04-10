package core

import (
	"fmt"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/initialize"
	ragservice "github.com/LightningRAG/LightningRAG/server/service/rag"
	"github.com/LightningRAG/LightningRAG/server/service/system"
	"go.uber.org/zap"
)

func RunServer() {
	if global.LRAG_CONFIG.System.UseRedis {
		// 初始化redis服务
		initialize.Redis()
		if global.LRAG_CONFIG.System.UseMultipoint {
			initialize.RedisList()
		}
	}

	if global.LRAG_CONFIG.System.UseMongo {
		err := initialize.Mongo.Initialization()
		if err != nil {
			zap.L().Error(fmt.Sprintf("%+v", err))
		}
	}
	// 从db加载jwt数据
	if global.LRAG_DB != nil {
		system.LoadAll()
		ragservice.ResumeIncompleteDocumentJobs()
	}

	Router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.LRAG_CONFIG.System.Addr)

	fmt.Printf(`
	欢迎使用 LightningRAG
	当前版本:%s
	项目地址：https://github.com/LightningRAG/LightningRAG
	官网：https://lightningrag.com (云亿连旗下)
	LightningRAG 讨论社区:https://support.qq.com/products/9999999999
	默认自动化文档地址:http://127.0.0.1%s/swagger/index.html
	默认MCP SSE地址:http://127.0.0.1%s%s
	默认MCP Message地址:http://127.0.0.1%s%s
	--------------------------------------版权声明--------------------------------------
	** 版权所有方：LightningRAG 开源团队 **
	** 剔除授权标识需购买商用授权：https://LightningRAG.com/license **
`, global.Version, address, address, global.LRAG_CONFIG.MCP.SSEPath, address, global.LRAG_CONFIG.MCP.MessagePath)
	initServer(address, Router, 10*time.Minute, 10*time.Minute)
}
