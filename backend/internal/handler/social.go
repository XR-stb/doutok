package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/response"
	"github.com/xiaoran/doutok/internal/repository"
)

var socialRepo = repository.NewSocialRepo()

func Follow(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}
	if userID == targetID {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "不能关注自己")
		return
	}

	if err := socialRepo.Follow(c.Request.Context(), userID, targetID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "关注失败")
		return
	}
	response.Success(c, gin.H{"msg": "ok"})
}

func Unfollow(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	if err := socialRepo.Unfollow(c.Request.Context(), userID, targetID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{"msg": "ok"})
}

func GetFollowing(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, err := socialRepo.ListFollowing(c.Request.Context(), userID, offset, limit)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{
		"list":     list,
		"offset":   offset + len(list),
		"has_more": len(list) == limit,
	})
}

func GetFollowers(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, err := socialRepo.ListFollowers(c.Request.Context(), userID, offset, limit)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{
		"list":     list,
		"offset":   offset + len(list),
		"has_more": len(list) == limit,
	})
}
