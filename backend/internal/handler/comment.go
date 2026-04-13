package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/model"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/response"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
	"github.com/xiaoran/doutok/internal/repository"
)

var (
	commentRepo     = repository.NewCommentRepo()
	commentLikeRepo = repository.NewCommentLikeRepo()
)

type CreateCommentReq struct {
	VideoID  int64  `json:"video_id" binding:"required"`
	ParentID int64  `json:"parent_id"` // 0=一级评论
	RootID   int64  `json:"root_id"`   // 0=一级评论
	Content  string `json:"content" binding:"required,min=1,max=512"`
}

func CreateComment(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}

	// 如果是回复且未指定 root_id, 自动设置
	rootID := req.RootID
	if req.ParentID > 0 && rootID == 0 {
		rootID = req.ParentID
	}

	comment := &model.Comment{
		ID:         snowflake.GenID(),
		VideoID:    req.VideoID,
		UserID:     userID,
		ParentID:   req.ParentID,
		RootID:     rootID,
		Content:    req.Content,
		Status:     1,
		IPLocation: getIPLocation(c),
	}

	if err := commentRepo.Create(c.Request.Context(), comment); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "评论失败")
		return
	}

	// 记录行为
	go behaviorRepo.Record(c.Request.Context(), &model.UserBehavior{
		UserID:     userID,
		TargetType: 1,
		TargetID:   req.VideoID,
		Action:     "comment",
	})

	response.Success(c, gin.H{
		"comment_id": comment.ID,
	})
}

// GetComments 获取视频评论
// 支持 sort_by=hot (热度排序) 和 sort_by=new (最新)
func GetComments(c *gin.Context) {
	videoIDStr := c.Query("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "video_id 无效")
		return
	}

	sortBy := c.DefaultQuery("sort_by", "hot")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	comments, err := commentRepo.ListByVideo(c.Request.Context(), videoID, sortBy, offset, limit)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	response.Success(c, gin.H{
		"comments": comments,
		"offset":   offset + len(comments),
		"has_more": len(comments) == limit,
	})
}

func DeleteComment(c *gin.Context) {
	userID := c.GetInt64("user_id")
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	// 简化版：直接删除（生产环境需要验证是否为评论作者或视频作者）
	_ = userID
	if err := commentRepo.Delete(c.Request.Context(), commentID, 0); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{"msg": "ok"})
}

func LikeComment(c *gin.Context) {
	userID := c.GetInt64("user_id")
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	if err := commentLikeRepo.Like(c.Request.Context(), commentID, userID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{"msg": "ok"})
}

// getIPLocation 获取 IP 属地 (简化版)
func getIPLocation(c *gin.Context) string {
	ip := c.ClientIP()
	// 生产环境: 调用 IP 地理位置 API (如 ip2region)
	// 学习版: 直接返回 IP
	if ip == "127.0.0.1" || ip == "::1" {
		return "本地"
	}
	return ip
}
