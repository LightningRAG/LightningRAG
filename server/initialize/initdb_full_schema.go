package initialize

import (
	"context"

	"github.com/LightningRAG/LightningRAG/server/service/system"
	"gorm.io/gorm"
)

// 晚于 ensure_tables（99999），在 InitDB 的 InitTables 阶段执行，与 RegisterTables 对齐补齐 RAG/渠道等表。
const initOrderInitDBFullSchema = system.InitOrderExternal + 1

type initDBFullSchema struct{}

func init() {
	system.RegisterInit(initOrderInitDBFullSchema, &initDBFullSchema{})
}

func (i *initDBFullSchema) InitializerName() string {
	return "init_db_full_schema"
}

func (i *initDBFullSchema) InitializeData(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (i *initDBFullSchema) DataInserted(ctx context.Context) bool {
	return true
}

func (i *initDBFullSchema) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	if err := AutoMigrateAllSchema(db); err != nil {
		return ctx, err
	}
	if err := bizModelWithDB(db); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// TableCreated 恒为 false：允许在已有库上再次执行 InitDB 时仍能 AutoMigrate 出新表（如后续版本新增 rag_channel_outbounds）。
func (i *initDBFullSchema) TableCreated(ctx context.Context) bool {
	return false
}
