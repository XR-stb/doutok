package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/xiaoran/doutok/internal/model"
)

type VideoRepo struct{}

func NewVideoRepo() *VideoRepo {
	return &VideoRepo{}
}

func (r *VideoRepo) Create(ctx context.Context, v *model.Video) error {
	query := `INSERT INTO videos (id, author_id, title, description, cover_url, play_url, duration, 
              width, height, file_size, status, visibility, tags, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	_, err := DB.ExecContext(ctx, query,
		v.ID, v.AuthorID, v.Title, v.Description, v.CoverURL, v.PlayURL,
		v.Duration, v.Width, v.Height, v.FileSize, v.Status, v.Visibility,
		v.Tags, now, now)
	return err
}

func (r *VideoRepo) GetByID(ctx context.Context, id int64) (*model.Video, error) {
	query := `SELECT id, author_id, title, description, cover_url, play_url, duration,
              width, height, file_size, status, visibility, like_count, comment_count,
              share_count, view_count, tags, created_at, updated_at
              FROM videos WHERE id = ? AND status != 3`
	v := &model.Video{}
	err := DB.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.AuthorID, &v.Title, &v.Description, &v.CoverURL, &v.PlayURL,
		&v.Duration, &v.Width, &v.Height, &v.FileSize, &v.Status, &v.Visibility,
		&v.LikeCount, &v.CommentCount, &v.ShareCount, &v.ViewCount,
		&v.Tags, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

// Feed 获取推荐视频流 - 基于时间线 + cursor 分页
// 生产环境这里会调用推荐算法服务，这里先用时间线降序作为 baseline
func (r *VideoRepo) Feed(ctx context.Context, cursor int64, limit int) ([]*model.Video, error) {
	query := `SELECT id, author_id, title, description, cover_url, play_url, duration,
              width, height, file_size, status, visibility, like_count, comment_count,
              share_count, view_count, tags, created_at, updated_at
              FROM videos WHERE status = 1 AND id < ? ORDER BY id DESC LIMIT ?`
	if cursor == 0 {
		// 第一页，取最新的
		cursor = 1<<63 - 1
	}
	rows, err := DB.QueryContext(ctx, query, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*model.Video
	for rows.Next() {
		v := &model.Video{}
		if err := rows.Scan(
			&v.ID, &v.AuthorID, &v.Title, &v.Description, &v.CoverURL, &v.PlayURL,
			&v.Duration, &v.Width, &v.Height, &v.FileSize, &v.Status, &v.Visibility,
			&v.LikeCount, &v.CommentCount, &v.ShareCount, &v.ViewCount,
			&v.Tags, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

// GetByAuthor 获取用户的视频列表
func (r *VideoRepo) GetByAuthor(ctx context.Context, authorID int64, offset, limit int) ([]*model.Video, error) {
	query := `SELECT id, author_id, title, description, cover_url, play_url, duration,
              width, height, file_size, status, visibility, like_count, comment_count,
              share_count, view_count, tags, created_at, updated_at
              FROM videos WHERE author_id = ? AND status = 1 ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := DB.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*model.Video
	for rows.Next() {
		v := &model.Video{}
		if err := rows.Scan(
			&v.ID, &v.AuthorID, &v.Title, &v.Description, &v.CoverURL, &v.PlayURL,
			&v.Duration, &v.Width, &v.Height, &v.FileSize, &v.Status, &v.Visibility,
			&v.LikeCount, &v.CommentCount, &v.ShareCount, &v.ViewCount,
			&v.Tags, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func (r *VideoRepo) IncrCounter(ctx context.Context, id int64, field string, delta int64) error {
	query := "UPDATE videos SET " + field + " = " + field + " + ? WHERE id = ?"
	_, err := DB.ExecContext(ctx, query, delta, id)
	return err
}

func (r *VideoRepo) Delete(ctx context.Context, id int64) error {
	query := `UPDATE videos SET status = 3, updated_at = ? WHERE id = ?`
	_, err := DB.ExecContext(ctx, query, time.Now(), id)
	return err
}

// ==================== 点赞 ====================

type LikeRepo struct{}

func NewLikeRepo() *LikeRepo {
	return &LikeRepo{}
}

// Like 点赞 - Redis 做实时计数, MySQL 做持久化
func (r *LikeRepo) Like(ctx context.Context, videoID, userID int64) error {
	query := `INSERT IGNORE INTO video_likes (video_id, user_id, created_at) VALUES (?, ?, ?)`
	result, err := DB.ExecContext(ctx, query, videoID, userID, time.Now())
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		// 真正新增了，更新视频计数
		return NewVideoRepo().IncrCounter(ctx, videoID, "like_count", 1)
	}
	return nil // 已经点过赞了
}

func (r *LikeRepo) Unlike(ctx context.Context, videoID, userID int64) error {
	query := `DELETE FROM video_likes WHERE video_id = ? AND user_id = ?`
	result, err := DB.ExecContext(ctx, query, videoID, userID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return NewVideoRepo().IncrCounter(ctx, videoID, "like_count", -1)
	}
	return nil
}

func (r *LikeRepo) IsLiked(ctx context.Context, videoID, userID int64) (bool, error) {
	query := `SELECT 1 FROM video_likes WHERE video_id = ? AND user_id = ? LIMIT 1`
	var exists int
	err := DB.QueryRowContext(ctx, query, videoID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// BatchIsLiked 批量查询点赞状态 - Feed 页需要知道哪些视频已点赞
func (r *LikeRepo) BatchIsLiked(ctx context.Context, videoIDs []int64, userID int64) (map[int64]bool, error) {
	if len(videoIDs) == 0 || userID == 0 {
		return map[int64]bool{}, nil
	}
	query := "SELECT video_id FROM video_likes WHERE user_id = ? AND video_id IN ("
	args := []interface{}{userID}
	for i, vid := range videoIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, vid)
	}
	query += ")"

	rows, err := DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]bool)
	for rows.Next() {
		var vid int64
		if err := rows.Scan(&vid); err != nil {
			return nil, err
		}
		result[vid] = true
	}
	return result, nil
}
