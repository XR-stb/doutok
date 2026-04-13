package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/repository"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/response"
)

func GetMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	if user == nil {
		response.ErrorWithMsg(c, 200, errno.ErrNotFound, "用户不存在")
		return
	}
	user.Password = "" // 不返回密码
	response.Success(c, user)
}

type UpdateMeReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Gender   *int   `json:"gender"`
	Birthday string `json:"birthday"`
}

func UpdateMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req UpdateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}

	fields := make(map[string]interface{})
	if req.Nickname != "" {
		fields["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		fields["avatar"] = req.Avatar
	}
	if req.Bio != "" {
		fields["bio"] = req.Bio
	}
	if req.Gender != nil {
		fields["gender"] = *req.Gender
	}
	if req.Birthday != "" {
		fields["birthday"] = req.Birthday
	}

	if err := userRepo.Update(c.Request.Context(), userID, fields); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "更新失败")
		return
	}
	response.Success(c, gin.H{"msg": "ok"})
}

func GetUserProfile(c *gin.Context) {
	idStr := c.Param("id")
	targetID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "无效的用户ID")
		return
	}

	user, err := userRepo.GetByID(c.Request.Context(), targetID)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	if user == nil {
		response.ErrorWithMsg(c, 200, errno.ErrNotFound, "用户不存在")
		return
	}
	user.Password = ""

	// 如果当前用户已登录，附加关注关系
	currentUserID := c.GetInt64("user_id")
	isFollowing := false
	if currentUserID > 0 && currentUserID != targetID {
		socialRepo := repository.NewSocialRepo()
		isFollowing, _ = socialRepo.IsFollowing(c.Request.Context(), currentUserID, targetID)
	}

	response.Success(c, gin.H{
		"user":         user,
		"is_following": isFollowing,
	})
}
