package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/config"
	"github.com/xiaoran/doutok/internal/handler"
	"github.com/xiaoran/doutok/internal/middleware"
	"github.com/xiaoran/doutok/internal/pkg/auth"
	"github.com/xiaoran/doutok/internal/pkg/cache"
	"time"
)

func setupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())

	// 限流：如果 Redis 可用，启用滑动窗口限流
	if cache.RDB != nil {
		r.Use(middleware.RateLimit(cache.RDB, 100, time.Minute))
	}

	// JWT Manager
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.ExpireHour)

	// ==================== 公开接口 ====================
	api := r.Group("/api/v1")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "service": "doutok-gateway"})
		})

		// 认证
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", handler.Register)
			authGroup.POST("/login", handler.Login)
		}

		// 公开 Feed（不需要登录也能看）
		api.GET("/feed", handler.Feed)
		api.GET("/videos/:id", handler.GetVideo)

		// 公开搜索
		api.GET("/search", handler.Search)

		// 评论（公开读取）
		api.GET("/comments", handler.GetComments)

		// 用户公开资料
		api.GET("/users/:id", handler.GetUserProfile)
		api.GET("/users/:id/following", handler.GetFollowing)
		api.GET("/users/:id/followers", handler.GetFollowers)

		// 直播（公开）
		api.GET("/lives", handler.ListLiveRooms)
		api.GET("/lives/:id", handler.GetLiveRoom)
		api.GET("/lives/:id/rank", handler.GetLiveRank)

		// WebSocket（不走 JWT 中间件，通过 query param 传 user_id）
		api.GET("/ws/live/:id", handler.WSLive)
		api.GET("/ws/chat", handler.WSChat)
	}

	// ==================== 需要认证的接口 ====================
	authed := api.Group("")
	authed.Use(middleware.JWTAuth(jwtMgr))
	{
		// Token 刷新
		authed.POST("/auth/refresh", handler.RefreshToken)

		// 用户
		authed.GET("/me", handler.GetMe)
		authed.PUT("/me", handler.UpdateMe)

		// 视频
		authed.POST("/videos", handler.UploadVideo)
		authed.DELETE("/videos/:id", handler.DeleteVideo)
		authed.POST("/videos/:id/like", handler.LikeVideo)
		authed.DELETE("/videos/:id/like", handler.UnlikeVideo)

		// 文件上传
		authed.POST("/upload", handler.UploadFile)

		// 评论
		authed.POST("/comments", handler.CreateComment)
		authed.DELETE("/comments/:id", handler.DeleteComment)
		authed.POST("/comments/:id/like", handler.LikeComment)

		// 社交
		authed.POST("/users/:id/follow", handler.Follow)
		authed.DELETE("/users/:id/follow", handler.Unfollow)

		// 聊天
		authed.GET("/conversations", handler.ListConversations)
		authed.POST("/conversations", handler.CreateConversation)
		authed.GET("/conversations/:id/messages", handler.GetMessages)
		authed.POST("/messages", handler.SendMessage)

		// 直播（需登录）
		authed.POST("/lives", handler.CreateLiveRoom)
		authed.PUT("/lives/:id", handler.UpdateLiveRoom)
		authed.POST("/lives/:id/gift", handler.SendGift)
		authed.POST("/lives/:id/like", handler.LikeLive)
	}

	// ==================== Debug 接口 ====================
	if cfg.Debug {
		debug := r.Group("/debug")
		{
			debug.GET("/config", handler.DebugConfig(cfg))
			debug.GET("/behaviors", middleware.JWTAuth(jwtMgr), handler.DebugBehaviors)
		}
	}

	return r
}
