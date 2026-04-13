package main

import (
	"github.com/LightningRAG/LightningRAG/server/core"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/initialize"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

//go:generate go mod tidy
//go:generate go mod download

// @Tag.Name        Base
// @Tag.Name        SysUser
// @Tag.Description Users

// @title                       LightningRAG API
// @version                     v2.9.0
// @description                 LightningRAG full-stack platform API
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /
func main() {
	// 初始化系统
	initializeSystem()
	// 运行服务器
	core.RunServer()
}

// initializeSystem 初始化系统所有组件
// 提取为单独函数以便于系统重载时调用
func initializeSystem() {
	global.LRAG_VP = core.Viper() // 初始化Viper
	initialize.OtherInit()
	global.LRAG_LOG = core.Zap() // 初始化zap日志库
	zap.ReplaceGlobals(global.LRAG_LOG)
	global.LRAG_DB = initialize.Gorm() // gorm连接数据库
	initialize.Timer()
	initialize.DBList()
	initialize.SetupHandlers() // 注册全局函数
	if global.LRAG_DB != nil {
		initialize.RegisterTables() // 初始化表
		initialize.LoadOAuthGlobalFromDB()
	}
}
