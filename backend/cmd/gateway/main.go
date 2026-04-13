package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiaoran/doutok/internal/config"
	"github.com/xiaoran/doutok/internal/pkg/auth"
	"github.com/xiaoran/doutok/internal/pkg/cache"
	"github.com/xiaoran/doutok/internal/pkg/logger"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
	"github.com/xiaoran/doutok/internal/repository"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	logger.Info("DouTok Gateway starting",
		"mode", cfg.Server.Mode,
		"debug", cfg.Debug)

	// 初始化 Snowflake ID 生成器
	snowflake.Init(1, 1) // datacenterID=1, workerID=1

	// 初始化全局 JWT Manager
	auth.InitJWT(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.ExpireHour)

	// 初始化数据库
	if err := repository.InitDB(cfg.Database); err != nil {
		logger.Error("Failed to init database", "err", err)
		logger.Info("Running WITHOUT database - some features disabled")
	} else {
		defer repository.Close()
	}

	// 初始化 Redis
	if err := cache.InitRedis(cfg.Redis); err != nil {
		logger.Error("Failed to init Redis", "err", err)
		logger.Info("Running WITHOUT Redis - some features disabled")
	}

	// 设置路由
	r := setupRouter(cfg)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "err", err)
			os.Exit(1)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "err", err)
	}
	logger.Info("Server exited")
}
