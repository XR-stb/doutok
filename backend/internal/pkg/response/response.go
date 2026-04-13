package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/pkg/errno"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type PageData struct {
	List    interface{} `json:"list"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	HasMore bool        `json:"has_more"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

func SuccessPage(c *gin.Context, list interface{}, total int64, page, size int) {
	Success(c, PageData{
		List: list, Total: total, Page: page, Size: size,
		HasMore: int64(page*size) < total,
	})
}

func Error(c *gin.Context, err *errno.Errno) {
	c.JSON(http.StatusOK, Response{Code: err.Code, Msg: err.Msg})
}

func ServerError(c *gin.Context) {
	Error(c, errno.ErrServer)
}
