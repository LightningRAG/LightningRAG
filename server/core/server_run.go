package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// initServer 启动服务并实现优雅关闭
func initServer(address string, router *gin.Engine, readTimeout, writeTimeout time.Duration) {
	h := middleware.WrapStripLeadingAPIPrefixBeforeRouting(router)
	// 创建服务
	srv := &http.Server{
		Addr:              address,
		Handler:           h,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("server listen failed", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("http server shutdown error", zap.Error(err))
	}

	global.LRAG_Timer.Close()
	zap.L().Info("timer tasks stopped")

	if global.LRAG_DB != nil {
		if sqlDB, err := global.LRAG_DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	for name, db := range global.LRAG_DBList {
		if db != nil {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
				zap.L().Info("closed db connection", zap.String("name", name))
			}
		}
	}

	if global.LRAG_REDIS != nil {
		_ = global.LRAG_REDIS.Close()
	}
	for _, rc := range global.LRAG_REDISList {
		if rc != nil {
			_ = rc.Close()
		}
	}

	if global.LRAG_MONGO != nil {
		_ = global.LRAG_MONGO.Close(ctx)
	}

	zap.L().Info("all resources released, server stopped")
}
