package initialize

import (
	"time"

	"github.com/LightningRAG/LightningRAG/server/config"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/initialize/internal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormPgSql 初始化 Postgresql 数据库
// Author [piexlmax](https://github.com/piexlmax)
// Author [SliverHorn](https://github.com/SliverHorn)
func GormPgSql() *gorm.DB {
	p := global.LRAG_CONFIG.Pgsql
	return initPgSqlDatabase(p)
}

// GormPgSqlByConfig 初始化 Postgresql 数据库 通过指定参数
func GormPgSqlByConfig(p config.Pgsql) *gorm.DB {
	return initPgSqlDatabase(p)
}

// initPgSqlDatabase 初始化 Postgresql 数据库的辅助函数
func initPgSqlDatabase(p config.Pgsql) *gorm.DB {
	if p.Dbname == "" {
		return nil
	}
	pgsqlConfig := postgres.Config{
		DSN:                  p.Dsn(), // DSN data source name
		PreferSimpleProtocol: false,
	}
	// 数据库配置
	general := p.GeneralDB
	if db, err := gorm.Open(postgres.New(pgsqlConfig), internal.Gorm.Config(general)); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(p.MaxIdleConns)
		sqlDB.SetMaxOpenConns(p.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)
		return db
	}
}
