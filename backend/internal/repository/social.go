package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/xiaoran/doutok/internal/model"
)

type SocialRepo struct{}

func NewSocialRepo() *SocialRepo {
	return &SocialRepo{}
}

// Follow 关注用户
// 双写模型：同时更新双方计数，并检查是否互关
func (r *SocialRepo) Follow(ctx context.Context, userID, targetID int64) error {
	// 1. 插入关注记录
	query := `INSERT INTO user_follows (user_id, target_id, status, created_at) VALUES (?, ?, 1, ?)
              ON DUPLICATE KEY UPDATE status = 1`
	_, err := DB.ExecContext(ctx, query, userID, targetID, time.Now())
	if err != nil {
		return err
	}

	// 2. 检查对方是否也关注了我 -> 互关
	var reverseStatus int
	err = DB.QueryRowContext(ctx,
		"SELECT status FROM user_follows WHERE user_id = ? AND target_id = ? AND status > 0",
		targetID, userID).Scan(&reverseStatus)
	if err == nil {
		// 对方也关注了我，双方设为互关
		DB.ExecContext(ctx, "UPDATE user_follows SET status = 2 WHERE user_id = ? AND target_id = ?", userID, targetID)
		DB.ExecContext(ctx, "UPDATE user_follows SET status = 2 WHERE user_id = ? AND target_id = ?", targetID, userID)
	}

	// 3. 更新计数
	userRepo := NewUserRepo()
	userRepo.IncrCounter(ctx, userID, "follow_count", 1)
	userRepo.IncrCounter(ctx, targetID, "fan_count", 1)
	return nil
}

// Unfollow 取消关注
func (r *SocialRepo) Unfollow(ctx context.Context, userID, targetID int64) error {
	result, err := DB.ExecContext(ctx,
		"UPDATE user_follows SET status = 0 WHERE user_id = ? AND target_id = ? AND status > 0",
		userID, targetID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil // 本来就没关注
	}

	// 如果之前是互关，对方改回普通关注
	DB.ExecContext(ctx,
		"UPDATE user_follows SET status = 1 WHERE user_id = ? AND target_id = ? AND status = 2",
		targetID, userID)

	userRepo := NewUserRepo()
	userRepo.IncrCounter(ctx, userID, "follow_count", -1)
	userRepo.IncrCounter(ctx, targetID, "fan_count", -1)
	return nil
}

// IsFollowing 是否已关注
func (r *SocialRepo) IsFollowing(ctx context.Context, userID, targetID int64) (bool, error) {
	var status int
	err := DB.QueryRowContext(ctx,
		"SELECT status FROM user_follows WHERE user_id = ? AND target_id = ? AND status > 0",
		userID, targetID).Scan(&status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// GetFollowRelation 获取关注关系详情
func (r *SocialRepo) GetFollowRelation(ctx context.Context, userID, targetID int64) (int, error) {
	var status int
	err := DB.QueryRowContext(ctx,
		"SELECT status FROM user_follows WHERE user_id = ? AND target_id = ?",
		userID, targetID).Scan(&status)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return status, err
}

// ListFollowing 关注列表
func (r *SocialRepo) ListFollowing(ctx context.Context, userID int64, offset, limit int) ([]*model.UserFollow, error) {
	query := `SELECT uf.target_id, u.nickname, u.avatar, uf.status, uf.created_at
              FROM user_follows uf
              JOIN users u ON u.id = uf.target_id
              WHERE uf.user_id = ? AND uf.status > 0
              ORDER BY uf.created_at DESC LIMIT ? OFFSET ?`
	rows, err := DB.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.UserFollow
	for rows.Next() {
		f := &model.UserFollow{}
		if err := rows.Scan(&f.TargetID, &f.Nickname, &f.Avatar, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}

// ListFollowers 粉丝列表
func (r *SocialRepo) ListFollowers(ctx context.Context, userID int64, offset, limit int) ([]*model.UserFollow, error) {
	query := `SELECT uf.user_id, u.nickname, u.avatar, uf.status, uf.created_at
              FROM user_follows uf
              JOIN users u ON u.id = uf.user_id
              WHERE uf.target_id = ? AND uf.status > 0
              ORDER BY uf.created_at DESC LIMIT ? OFFSET ?`
	rows, err := DB.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.UserFollow
	for rows.Next() {
		f := &model.UserFollow{}
		if err := rows.Scan(&f.TargetID, &f.Nickname, &f.Avatar, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}
