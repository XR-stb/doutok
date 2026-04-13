package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit 基于 Redis 的滑动窗口限流
func RateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("rate:%s:%s", c.ClientIP(), c.FullPath())
		ctx := context.Background()

		now := time.Now().UnixMilli()
		windowStart := now - window.Milliseconds()

		pipe := rdb.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
		pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, window)

		cmds, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		count := cmds[2].(*redis.IntCmd).Val()
		if count > int64(limit) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(limit)-count))
		c.Next()
	}
}
