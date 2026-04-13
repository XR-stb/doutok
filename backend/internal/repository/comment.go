package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/xiaoran/doutok/internal/model"
)

type CommentRepo struct{}

func NewCommentRepo() *CommentRepo {
	return &CommentRepo{}
}

func (r *CommentRepo) Create(ctx context.Context, c *model.Comment) error {
	query := `INSERT INTO comments (id, video_id, user_id, parent_id, root_id, content, status, ip_location, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	_, err := DB.ExecContext(ctx, query,
		c.ID, c.VideoID, c.UserID, c.ParentID, c.RootID, c.Content, c.Status, c.IPLocation, now, now)
	if err != nil {
		return err
	}

	// 更新视频评论数
	NewVideoRepo().IncrCounter(ctx, c.VideoID, "comment_count", 1)

	// 如果是回复，更新父评论的 reply_count
	if c.ParentID > 0 {
		DB.ExecContext(ctx, "UPDATE comments SET reply_count = reply_count + 1 WHERE id = ?", c.RootID)
	}
	return nil
}

// ListByVideo 获取视频的一级评论 - 按热度排序
// 热度算法: 0.6 * hot_score + 0.3 * time_score + 0.1 * identity_boost
// hot_score = LOG2(like_count + 1) + 0.5 * LOG2(reply_count + 1)
// time_score = EXP(-0.693 * TIMESTAMPDIFF(HOUR, created_at, NOW()) / 12)
func (r *CommentRepo) ListByVideo(ctx context.Context, videoID int64, sortBy string, offset, limit int) ([]*model.Comment, error) {
	var orderClause string
	switch sortBy {
	case "hot":
		// 热度排序 - MySQL 内计算分数
		orderClause = `(0.6 * (LOG2(like_count + 1) + 0.5 * LOG2(reply_count + 1)) + 
                       0.3 * EXP(-0.693 * TIMESTAMPDIFF(HOUR, created_at, NOW()) / 12)) DESC`
	case "new":
		orderClause = `created_at DESC`
	default:
		orderClause = `created_at DESC`
	}

	query := `SELECT id, video_id, user_id, parent_id, root_id, content, like_count, reply_count,
              status, ip_location, created_at, updated_at
              FROM comments WHERE video_id = ? AND parent_id = 0 AND status = 1
              ORDER BY ` + orderClause + ` LIMIT ? OFFSET ?`

	rows, err := DB.QueryContext(ctx, query, videoID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		if err := rows.Scan(
			&c.ID, &c.VideoID, &c.UserID, &c.ParentID, &c.RootID, &c.Content,
			&c.LikeCount, &c.ReplyCount, &c.Status, &c.IPLocation,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// ListReplies 获取评论的子回复
func (r *CommentRepo) ListReplies(ctx context.Context, rootID int64, offset, limit int) ([]*model.Comment, error) {
	query := `SELECT id, video_id, user_id, parent_id, root_id, content, like_count, reply_count,
              status, ip_location, created_at, updated_at
              FROM comments WHERE root_id = ? AND parent_id != 0 AND status = 1
              ORDER BY created_at ASC LIMIT ? OFFSET ?`
	rows, err := DB.QueryContext(ctx, query, rootID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		if err := rows.Scan(
			&c.ID, &c.VideoID, &c.UserID, &c.ParentID, &c.RootID, &c.Content,
			&c.LikeCount, &c.ReplyCount, &c.Status, &c.IPLocation,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *CommentRepo) Delete(ctx context.Context, id, videoID int64) error {
	_, err := DB.ExecContext(ctx, "UPDATE comments SET status = 3 WHERE id = ?", id)
	if err == nil {
		NewVideoRepo().IncrCounter(ctx, videoID, "comment_count", -1)
	}
	return err
}

// ==================== 评论点赞 ====================

type CommentLikeRepo struct{}

func NewCommentLikeRepo() *CommentLikeRepo {
	return &CommentLikeRepo{}
}

func (r *CommentLikeRepo) Like(ctx context.Context, commentID, userID int64) error {
	query := `INSERT IGNORE INTO comment_likes (comment_id, user_id, created_at) VALUES (?, ?, ?)`
	result, err := DB.ExecContext(ctx, query, commentID, userID, time.Now())
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		DB.ExecContext(ctx, "UPDATE comments SET like_count = like_count + 1 WHERE id = ?", commentID)
	}
	return nil
}

func (r *CommentLikeRepo) IsLiked(ctx context.Context, commentID, userID int64) (bool, error) {
	var exists int
	err := DB.QueryRowContext(ctx,
		"SELECT 1 FROM comment_likes WHERE comment_id = ? AND user_id = ? LIMIT 1",
		commentID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
