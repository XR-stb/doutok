package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/pkg/auth"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/response"
)

const (
	ContextUserIDKey   = "user_id"
	ContextUsernameKey = "username"
	ContextRoleKey     = "role"
)

func JWTAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, errno.ErrAuth)
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errno.ErrAuth)
			c.Abort()
			return
		}

		claims, err := jwtMgr.Parse(parts[1])
		if err != nil {
			response.Error(c, errno.ErrTokenExpired)
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	if v, ok := c.Get(ContextUserIDKey); ok {
		return v.(int64)
	}
	return 0
}
