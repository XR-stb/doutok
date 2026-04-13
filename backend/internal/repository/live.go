package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiaoran/doutok/internal/model"
	"github.com/xiaoran/doutok/internal/pkg/cache"
)

type LiveRepo struct{}

func NewLiveRepo() *LiveRepo {
	return &LiveRepo{}
}

func (r *LiveRepo) Create(ctx context.Context, room *model.LiveRoom) error {
	query := `INSERT INTO live_rooms (id, anchor_id, title, cover_url, stream_key, status, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	_, err := DB.ExecContext(ctx, query,
		room.ID, room.AnchorID, room.Title, room.CoverURL, room.StreamKey, room.Status, now, now)
	return err
}

func (r *LiveRepo) GetByID(ctx context.Context, id int64) (*model.LiveRoom, error) {
	query := `SELECT id, anchor_id, title, cover_url, stream_key, status, viewer_count, peak_viewer,
              like_count, gift_value, started_at, ended_at, created_at, updated_at
              FROM live_rooms WHERE id = ?`
	room := &model.LiveRoom{}
	err := DB.QueryRowContext(ctx, query, id).Scan(
		&room.ID, &room.AnchorID, &room.Title, &room.CoverURL, &room.StreamKey,
		&room.Status, &room.ViewerCount, &room.PeakViewer, &room.LikeCount,
		&room.GiftValue, &room.StartedAt, &room.EndedAt, &room.CreatedAt, &room.UpdatedAt)
	return room, err
}

// ListLive 获取正在直播的房间列表
func (r *LiveRepo) ListLive(ctx context.Context, offset, limit int) ([]*model.LiveRoom, error) {
	query := `SELECT id, anchor_id, title, cover_url, status, viewer_count, like_count
              FROM live_rooms WHERE status = 1
              ORDER BY viewer_count DESC LIMIT ? OFFSET ?`
	rows, err := DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*model.LiveRoom
	for rows.Next() {
		room := &model.LiveRoom{}
		if err := rows.Scan(&room.ID, &room.AnchorID, &room.Title, &room.CoverURL,
			&room.Status, &room.ViewerCount, &room.LikeCount); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *LiveRepo) StartLive(ctx context.Context, roomID int64) error {
	_, err := DB.ExecContext(ctx,
		"UPDATE live_rooms SET status = 1, started_at = ?, updated_at = ? WHERE id = ?",
		time.Now(), time.Now(), roomID)
	return err
}

func (r *LiveRepo) EndLive(ctx context.Context, roomID int64) error {
	_, err := DB.ExecContext(ctx,
		"UPDATE live_rooms SET status = 2, ended_at = ?, updated_at = ? WHERE id = ?",
		time.Now(), time.Now(), roomID)
	return err
}

// ==================== 礼物 ====================

func (r *LiveRepo) SendGift(ctx context.Context, gift *model.LiveGift) error {
	query := `INSERT INTO live_gifts (room_id, sender_id, gift_id, gift_name, gift_value, count, combo, created_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.ExecContext(ctx, query,
		gift.RoomID, gift.SenderID, gift.GiftID, gift.GiftName, gift.GiftValue,
		gift.Count, gift.Combo, time.Now())
	if err != nil {
		return err
	}

	// 更新直播间礼物总价值
	totalValue := int64(gift.GiftValue) * int64(gift.Count)
	DB.ExecContext(ctx,
		"UPDATE live_rooms SET gift_value = gift_value + ? WHERE id = ?",
		totalValue, gift.RoomID)

	return nil
}

// ==================== Redis 排行榜 ====================

func rankKey(roomID int64) string {
	return fmt.Sprintf("live:rank:%d", roomID)
}

func likeKey(roomID int64) string {
	return fmt.Sprintf("live:likes:%d", roomID)
}

func viewerKey(roomID int64) string {
	return fmt.Sprintf("live:viewers:%d", roomID)
}

// UpdateRank 更新送礼排行榜 (Redis ZSET)
func (r *LiveRepo) UpdateRank(ctx context.Context, roomID, userID int64, score float64) error {
	return cache.RDB.ZIncrBy(ctx, rankKey(roomID), score, strconv.FormatInt(userID, 10)).Err()
}

// GetTopRank 获取排行榜 Top N
func (r *LiveRepo) GetTopRank(ctx context.Context, roomID int64, n int64) ([]redis.Z, error) {
	return cache.RDB.ZRevRangeWithScores(ctx, rankKey(roomID), 0, n-1).Result()
}

// IncrLike 直播点赞 (Redis INCR, 高并发场景不走 MySQL)
func (r *LiveRepo) IncrLike(ctx context.Context, roomID int64) (int64, error) {
	return cache.RDB.Incr(ctx, likeKey(roomID)).Result()
}

// GetLikeCount 获取点赞总数
func (r *LiveRepo) GetLikeCount(ctx context.Context, roomID int64) (int64, error) {
	val, err := cache.RDB.Get(ctx, likeKey(roomID)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// AddViewer 观众进入直播间
func (r *LiveRepo) AddViewer(ctx context.Context, roomID, userID int64) error {
	cache.RDB.SAdd(ctx, viewerKey(roomID), userID)
	count, _ := cache.RDB.SCard(ctx, viewerKey(roomID)).Result()
	// 更新峰值
	DB.ExecContext(ctx,
		"UPDATE live_rooms SET viewer_count = ?, peak_viewer = GREATEST(peak_viewer, ?) WHERE id = ?",
		count, count, roomID)
	return nil
}

// RemoveViewer 观众离开直播间
func (r *LiveRepo) RemoveViewer(ctx context.Context, roomID, userID int64) error {
	cache.RDB.SRem(ctx, viewerKey(roomID), userID)
	count, _ := cache.RDB.SCard(ctx, viewerKey(roomID)).Result()
	DB.ExecContext(ctx, "UPDATE live_rooms SET viewer_count = ? WHERE id = ?", count, roomID)
	return nil
}

// GetViewerCount 获取当前观众数
func (r *LiveRepo) GetViewerCount(ctx context.Context, roomID int64) (int64, error) {
	return cache.RDB.SCard(ctx, viewerKey(roomID)).Result()
}

// CleanupRoom 直播结束时清理 Redis 数据，持久化到 MySQL
func (r *LiveRepo) CleanupRoom(ctx context.Context, roomID int64) error {
	// 持久化点赞数
	likes, _ := r.GetLikeCount(ctx, roomID)
	DB.ExecContext(ctx, "UPDATE live_rooms SET like_count = ? WHERE id = ?", likes, roomID)

	// 清理 Redis keys
	cache.RDB.Del(ctx, rankKey(roomID), likeKey(roomID), viewerKey(roomID))
	return nil
}
