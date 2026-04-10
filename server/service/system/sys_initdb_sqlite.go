package system

import (
	"context"
	"errors"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"gorm.io/gorm"
	"path/filepath"

	"github.com/LightningRAG/LightningRAG/server/config"
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"github.com/LightningRAG/LightningRAG/server/utils"
)

type SqliteInitHandler struct{}

func NewSqliteInitHandler() *SqliteInitHandler {
	return &SqliteInitHandler{}
}

// WriteConfig mysql回写配置
func (h SqliteInitHandler) WriteConfig(ctx context.Context) error {
	c, ok := ctx.Value("config").(config.Sqlite)
	if !ok {
		return errors.New("sqlite config invalid")
	}
	global.LRAG_CONFIG.System.DbType = "sqlite"
	global.LRAG_CONFIG.Sqlite = c
	global.LRAG_CONFIG.JWT.SigningKey = uuid.New().String()
	cs := utils.StructToMap(global.LRAG_CONFIG)
	for k, v := range cs {
		global.LRAG_VP.Set(k, v)
	}
	global.LRAG_ACTIVE_DBNAME = &c.Dbname
	return global.LRAG_VP.WriteConfig()
}

// EnsureDB 创建数据库并初始化 sqlite
func (h SqliteInitHandler) EnsureDB(ctx context.Context, conf *request.InitDB) (next context.Context, err error) {
	if s, ok := ctx.Value("dbtype").(string); !ok || s != "sqlite" {
		return ctx, ErrDBTypeMismatch
	}

	c := conf.ToSqliteConfig()
	next = context.WithValue(ctx, "config", c)
	if c.Dbname == "" {
		return ctx, nil
	} // 如果没有数据库名, 则跳出初始化数据

	dsn := conf.SqliteEmptyDsn()

	var db *gorm.DB
	if db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}); err != nil {
		return ctx, err
	}
	global.LRAG_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
	next = context.WithValue(next, "db", db)
	return next, err
}

func (h SqliteInitHandler) InitTables(ctx context.Context, inits initSlice) error {
	return createTables(ctx, inits)
}

func (h SqliteInitHandler) InitData(ctx context.Context, inits initSlice) error {
	next, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, init := range inits {
		if init.DataInserted(next) {
			color.Info.Printf(InitDataExist, Sqlite, init.InitializerName())
			continue
		}
		if n, err := init.InitializeData(next); err != nil {
			color.Info.Printf(InitDataFailed, Sqlite, init.InitializerName(), err)
			return err
		} else {
			next = n
			color.Info.Printf(InitDataSuccess, Sqlite, init.InitializerName())
		}
	}
	color.Info.Printf(InitSuccess, Sqlite)
	return nil
}
