package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/model"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/logger"
	"github.com/xiaoran/doutok/internal/pkg/response"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
	"github.com/xiaoran/doutok/internal/storage"
	"github.com/xiaoran/doutok/internal/repository"
)

var (
	videoRepo    = repository.NewVideoRepo()
	likeRepo     = repository.NewLikeRepo()
	behaviorRepo = repository.NewBehaviorRepo()
)

// Feed 获取推荐视频流
// 学习版: 基于时间线 + cursor 分页
// 生产版: 应调用推荐算法服务，这里留了 behavior 记录为推荐算法提供数据
func Feed(c *gin.Context) {
	cursorStr := c.DefaultQuery("cursor", "0")
	limitStr := c.DefaultQuery("limit", "10")

	cursor, _ := strconv.ParseInt(cursorStr, 10, 64)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 30 {
		limit = 10
	}

	videos, err := videoRepo.Feed(c.Request.Context(), cursor, limit)
	if err != nil {
		logger.Error("feed query failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	// 批量查询当前用户的点赞状态
	userID := c.GetInt64("user_id")
	var likedMap map[int64]bool
	if userID > 0 && len(videos) > 0 {
		videoIDs := make([]int64, len(videos))
		for i, v := range videos {
			videoIDs[i] = v.ID
		}
		likedMap, _ = likeRepo.BatchIsLiked(c.Request.Context(), videoIDs, userID)

		// 记录曝光行为(异步)
		go func() {
			for _, v := range videos {
				behaviorRepo.Record(c.Request.Context(), &model.UserBehavior{
					UserID:     userID,
					TargetType: 1, // video
					TargetID:   v.ID,
					Action:     "impression",
				})
			}
		}()
	}

	// 组装响应
	type VideoItem struct {
		*model.Video
		IsLiked bool `json:"is_liked"`
	}
	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = VideoItem{
			Video:   v,
			IsLiked: likedMap[v.ID],
		}
	}

	// 返回下一页 cursor
	var nextCursor int64
	if len(videos) > 0 {
		nextCursor = videos[len(videos)-1].ID
	}

	response.Success(c, gin.H{
		"videos":      items,
		"next_cursor": nextCursor,
		"has_more":    len(videos) == limit,
	})
}

func GetVideo(c *gin.Context) {
	idStr := c.Param("id")
	videoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	video, err := videoRepo.GetByID(c.Request.Context(), videoID)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	if video == nil {
		response.ErrorWithMsg(c, 200, errno.ErrNotFound, "视频不存在")
		return
	}

	// 增加浏览量
	videoRepo.IncrCounter(c.Request.Context(), videoID, "view_count", 1)

	// 记录观看行为
	userID := c.GetInt64("user_id")
	if userID > 0 {
		go behaviorRepo.Record(c.Request.Context(), &model.UserBehavior{
			UserID:     userID,
			TargetType: 1,
			TargetID:   videoID,
			Action:     "view",
		})
	}

	// 查询点赞状态
	isLiked := false
	if userID > 0 {
		isLiked, _ = likeRepo.IsLiked(c.Request.Context(), videoID, userID)
	}

	// 查询作者信息
	author, _ := userRepo.GetByID(c.Request.Context(), video.AuthorID)
	if author != nil {
		author.Password = ""
	}

	response.Success(c, gin.H{
		"video":    video,
		"author":   author,
		"is_liked": isLiked,
	})
}

type UploadVideoReq struct {
	Title       string `json:"title" binding:"required,max=128"`
	Description string `json:"description" binding:"max=512"`
	CoverURL    string `json:"cover_url" binding:"required"`
	PlayURL     string `json:"play_url" binding:"required"`
	Duration    int    `json:"duration"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FileSize    int64  `json:"file_size"`
	Tags        string `json:"tags"`
	Visibility  int    `json:"visibility"` // 1=public, 2=friends, 3=private
}

func UploadVideo(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Support both JSON and multipart form
	title := c.PostForm("title")
	description := c.PostForm("description")
	tags := c.PostForm("tags")

	if title == "" {
		// Try JSON body as fallback
		var req UploadVideoReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "title is required")
			return
		}
		title = req.Title
		description = req.Description
		tags = req.Tags
	}

	// Handle file upload if present
	var playURL, coverURL string
	var fileSize int64

	file, header, err := c.Request.FormFile("video")
	if err == nil {
		defer file.Close()
		fileSize = header.Size

		// Upload to MinIO
		url, uploadErr := storage.Upload(c.Request.Context(), file, header.Size, header.Filename, header.Header.Get("Content-Type"))
		if uploadErr != nil {
			logger.Error("minio upload failed", "err", uploadErr)
			response.ErrorWithMsg(c, 200, errno.ErrInternal, "File upload failed")
			return
		}
		playURL = url
		coverURL = "" // TODO: generate thumbnail

		logger.Info("video uploaded to MinIO",
			"filename", header.Filename,
			"size", header.Size,
			"url", url)
	} else {
		// No file - use URLs from form
		playURL = c.PostForm("play_url")
		coverURL = c.PostForm("cover_url")
		if playURL == "" {
			response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "Video file is required")
			return
		}
	}

	video := &model.Video{
		ID:          snowflake.GenID(),
		AuthorID:    userID,
		Title:       title,
		Description: description,
		CoverURL:    coverURL,
		PlayURL:     playURL,
		Duration:    0,
		FileSize:    fileSize,
		Status:      1,
		Visibility:  1,
		Tags:        tags,
	}

	if err := videoRepo.Create(c.Request.Context(), video); err != nil {
		logger.Error("create video failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "Upload failed")
		return
	}

	userRepo.IncrCounter(c.Request.Context(), userID, "video_count", 1)

	logger.Info("video published", "video_id", video.ID, "author_id", userID, "title", title)
	response.Success(c, gin.H{
		"video_id":  video.ID,
		"play_url":  video.PlayURL,
		"cover_url": video.CoverURL,
		"title":     title,
	})
}

func DeleteVideo(c *gin.Context) {
	userID := c.GetInt64("user_id")
	videoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	// 验证是否为作者
	video, _ := videoRepo.GetByID(c.Request.Context(), videoID)
	if video == nil {
		response.ErrorWithMsg(c, 200, errno.ErrNotFound, "视频不存在")
		return
	}
	if video.AuthorID != userID {
		response.ErrorWithMsg(c, 200, errno.ErrForbidden, "无权操作")
		return
	}

	if err := videoRepo.Delete(c.Request.Context(), videoID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	userRepo.IncrCounter(c.Request.Context(), userID, "video_count", -1)
	response.Success(c, gin.H{"msg": "ok"})
}

// LikeVideo 点赞视频
// 高并发设计: Redis SISMEMBER 防重 -> Redis INCR 实时计数 -> 异步写 MySQL 持久化
// 学习版直接走 MySQL, 注释说明了生产版的做法
func LikeVideo(c *gin.Context) {
	userID := c.GetInt64("user_id")
	videoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	/*
		生产版高并发点赞方案:
		1. SISMEMBER like:video:{id} {user_id} -> 防重复点赞
		2. SADD like:video:{id} {user_id}     -> 记录点赞
		3. INCR video:like_count:{id}          -> 实时计数
		4. 发 Kafka 消息 -> consumer 异步写 MySQL
		5. 定时任务对账: Redis count vs MySQL count
	*/

	if err := likeRepo.Like(c.Request.Context(), videoID, userID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "点赞失败")
		return
	}

	// 记录行为
	go behaviorRepo.Record(c.Request.Context(), &model.UserBehavior{
		UserID:     userID,
		TargetType: 1,
		TargetID:   videoID,
		Action:     "like",
	})

	response.Success(c, gin.H{"msg": "ok"})
}

func UnlikeVideo(c *gin.Context) {
	userID := c.GetInt64("user_id")
	videoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	if err := likeRepo.Unlike(c.Request.Context(), videoID, userID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{"msg": "ok"})
}

// ==================== 文件上传 (MinIO) ====================

// UploadFile 通用文件上传 -> MinIO
// 返回 CDN URL 给前端，前端再把 URL 传给 UploadVideo
func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "请选择文件")
		return
	}
	defer file.Close()

	// 限制文件大小 (200MB)
	if header.Size > 200*1024*1024 {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "文件大小不能超过200MB")
		return
	}

	// 生成唯一文件名
	ext := ""
	if idx := len(header.Filename) - 1; idx > 0 {
		for i := idx; i >= 0; i-- {
			if header.Filename[i] == '.' {
				ext = header.Filename[i:]
				break
			}
		}
	}
	objectName := fmt.Sprintf("uploads/%s/%d%s",
		time.Now().Format("2006/01/02"), snowflake.GenID(), ext)

	// TODO: 实际写入 MinIO
	// 学习版先返回一个模拟的 URL
	fileURL := fmt.Sprintf("http://localhost:9000/doutok/%s", objectName)

	logger.Info("file uploaded", "name", header.Filename, "size", header.Size, "path", objectName)
	response.Success(c, gin.H{
		"url":  fileURL,
		"path": objectName,
		"size": header.Size,
	})
}
