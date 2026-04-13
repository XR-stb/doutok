package handler

import (

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/config"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/response"
	"github.com/xiaoran/doutok/internal/repository"
)

// Search 全局搜索 (学习版: MySQL LIKE, 生产版: Elasticsearch)
func Search(c *gin.Context) {
	keyword := c.Query("q")
	searchType := c.DefaultQuery("type", "video") // video, user

	if keyword == "" {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "搜索关键词不能为空")
		return
	}

	ctx := c.Request.Context()

	switch searchType {
	case "user":
		// 简化版: MySQL LIKE
		user, err := userRepo.GetByUsername(ctx, keyword)
		if err != nil {
			response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
			return
		}
		var results []interface{}
		if user != nil {
			user.Password = ""
			results = append(results, user)
		}
		response.Success(c, gin.H{"results": results, "type": "user"})

	default:
		// TODO: 生产版接入 Elasticsearch
		// 学习版暂时返回空
		response.Success(c, gin.H{
			"results": []interface{}{},
			"type":    searchType,
			"hint":    "搜索功能需要 Elasticsearch, 当前为学习版占位",
		})
	}
}

// DebugConfig Debug 面板接口
func DebugConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Debug {
			response.ErrorWithMsg(c, 200, errno.ErrForbidden, "Debug mode disabled")
			return
		}

		// 收集系统状态
		response.Success(c, gin.H{
			"server": gin.H{
				"name": cfg.Server.Name,
				"host": cfg.Server.Host,
				"port": cfg.Server.Port,
				"mode": cfg.Server.Mode,
			},
			"database": gin.H{
				"host":     cfg.Database.Host,
				"db_name":  cfg.Database.DBName,
				"max_open": cfg.Database.MaxOpenConns,
			},
			"redis": gin.H{
				"host": cfg.Redis.Host,
				"db":   cfg.Redis.DB,
			},
			"kafka": gin.H{
				"brokers": cfg.Kafka.Brokers,
			},
			"minio": gin.H{
				"endpoint": cfg.MinIO.Endpoint,
				"bucket":   cfg.MinIO.Bucket,
			},
			"features": gin.H{
				"debug":      cfg.Debug,
				"log_level":  cfg.Log.Level,
				"log_format": cfg.Log.Format,
			},
		})
	}
}

// DebugBehaviors Debug: 查看用户行为日志
func DebugBehaviors(c *gin.Context) {
	userID := c.GetInt64("user_id")
	behaviors, _ := repository.NewBehaviorRepo().GetUserHistory(c.Request.Context(), userID, "view", 50)
	response.Success(c, gin.H{
		"user_id":  userID,
		"history":  behaviors,
	})
}
