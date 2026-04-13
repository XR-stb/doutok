package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xiaoran/doutok/internal/config"
	"github.com/xiaoran/doutok/internal/pkg/logger"
)

var DB *sql.DB

func InitDB(cfg config.DatabaseConfig) error {
	var err error
	DB, err = sql.Open("mysql", cfg.DSN())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetConnMaxLifetime(cfg.MaxLifetime)

	// 重试连接 - 等待 Docker MySQL 就绪
	for i := 0; i < 30; i++ {
		if err = DB.Ping(); err == nil {
			logger.Info("MySQL connected", "host", cfg.Host, "db", cfg.DBName)
			return nil
		}
		logger.Info("Waiting for MySQL...", "attempt", i+1)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("mysql connect timeout: %w", err)
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
