package repository

import (
	"context"
	"time"

	"github.com/xiaoran/doutok/internal/model"
)

type BehaviorRepo struct{}

func NewBehaviorRepo() *BehaviorRepo {
	return &BehaviorRepo{}
}

// Record 记录用户行为 - 推荐算法的核心数据源
// 生产环境: 发送到 Kafka, 异步写入 ClickHouse
// 学习版: 直接写 MySQL, 方便调试
func (r *BehaviorRepo) Record(ctx context.Context, b *model.UserBehavior) error {
	query := `INSERT INTO user_behaviors (user_id, target_type, target_id, action, duration, extra, created_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.ExecContext(ctx, query,
		b.UserID, b.TargetType, b.TargetID, b.Action, b.Duration, b.Extra, time.Now())
	return err
}

// GetUserHistory 获取用户最近的行为历史 - 用于推荐召回
func (r *BehaviorRepo) GetUserHistory(ctx context.Context, userID int64, action string, limit int) ([]int64, error) {
	query := `SELECT target_id FROM user_behaviors 
              WHERE user_id = ? AND action = ? AND target_type = 1
              ORDER BY created_at DESC LIMIT ?`
	rows, err := DB.QueryContext(ctx, query, userID, action, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetVideoInteractions 获取视频的互动用户列表 - 用于协同过滤
func (r *BehaviorRepo) GetVideoInteractions(ctx context.Context, videoID int64, action string, limit int) ([]int64, error) {
	query := `SELECT DISTINCT user_id FROM user_behaviors 
              WHERE target_id = ? AND target_type = 1 AND action = ?
              ORDER BY created_at DESC LIMIT ?`
	rows, err := DB.QueryContext(ctx, query, videoID, action, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
