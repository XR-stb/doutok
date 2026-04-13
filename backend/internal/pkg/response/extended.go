package response

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/pkg/errno"
)

// ErrorWithMsg 带自定义消息的错误响应
// 兼容 handler 中 response.Error(c, http.StatusXxx, errCode, "msg") 的调用方式
func ErrorWithMsg(c *gin.Context, httpCode int, err *errno.Errno, msg string) {
	if msg == "" {
		msg = err.Msg
	}
	c.JSON(httpCode, Response{Code: err.Code, Msg: msg})
}
