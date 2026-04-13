package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiaoran/doutok/internal/model"
	"github.com/xiaoran/doutok/internal/pkg/cache"
)

type ChatRepo struct{}

func NewChatRepo() *ChatRepo {
	return &ChatRepo{}
}

// CreateConversation 创建私聊会话
func (r *ChatRepo) CreateConversation(ctx context.Context, convID, userA, userB int64) error {
	now := time.Now()
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO conversations (id, type, member_count, created_at, updated_at) VALUES (?, 1, 2, ?, ?)",
		convID, now, now)
	if err != nil {
		return err
	}

	memberQuery := "INSERT INTO conversation_members (conversation_id, user_id, joined_at) VALUES (?, ?, ?)"
	tx.ExecContext(ctx, memberQuery, convID, userA, now)
	tx.ExecContext(ctx, memberQuery, convID, userB, now)

	return tx.Commit()
}

// FindPrivateConversation 查找两个用户之间的私聊会话
func (r *ChatRepo) FindPrivateConversation(ctx context.Context, userA, userB int64) (int64, error) {
	query := `SELECT cm1.conversation_id FROM conversation_members cm1
              JOIN conversation_members cm2 ON cm1.conversation_id = cm2.conversation_id
              JOIN conversations c ON c.id = cm1.conversation_id
              WHERE cm1.user_id = ? AND cm2.user_id = ? AND c.type = 1
              LIMIT 1`
	var convID int64
	err := DB.QueryRowContext(ctx, query, userA, userB).Scan(&convID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return convID, err
}

// ListConversations 获取用户的会话列表
func (r *ChatRepo) ListConversations(ctx context.Context, userID int64, offset, limit int) ([]*model.ConversationItem, error) {
	query := `SELECT c.id, c.type, c.last_msg_at, cm.unread_count, cm.muted,
              m.content as last_msg, m.msg_type as last_msg_type,
              u.id as peer_id, u.nickname as peer_name, u.avatar as peer_avatar
              FROM conversation_members cm
              JOIN conversations c ON c.id = cm.conversation_id
              LEFT JOIN messages m ON m.id = c.last_msg_id
              LEFT JOIN conversation_members cm2 ON cm2.conversation_id = c.id AND cm2.user_id != ?
              LEFT JOIN users u ON u.id = cm2.user_id
              WHERE cm.user_id = ?
              ORDER BY c.last_msg_at DESC NULLS LAST
              LIMIT ? OFFSET ?`
	rows, err := DB.QueryContext(ctx, query, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.ConversationItem
	for rows.Next() {
		item := &model.ConversationItem{}
		var lastMsg, peerAvatar sql.NullString
		var lastMsgType sql.NullInt32
		var lastMsgAt sql.NullTime
		if err := rows.Scan(
			&item.ConversationID, &item.Type, &lastMsgAt, &item.UnreadCount, &item.Muted,
			&lastMsg, &lastMsgType, &item.PeerID, &item.PeerName, &peerAvatar); err != nil {
			return nil, err
		}
		if lastMsg.Valid {
			item.LastMsg = lastMsg.String
		}
		if lastMsgType.Valid {
			item.LastMsgType = int(lastMsgType.Int32)
		}
		if lastMsgAt.Valid {
			item.LastMsgAt = lastMsgAt.Time
		}
		if peerAvatar.Valid {
			item.PeerAvatar = peerAvatar.String
		}
		items = append(items, item)
	}
	return items, nil
}

// SaveMessage 保存消息
func (r *ChatRepo) SaveMessage(ctx context.Context, msg *model.Message) error {
	query := `INSERT INTO messages (id, conversation_id, sender_id, msg_type, content, extra, status, created_at)
              VALUES (?, ?, ?, ?, ?, ?, 1, ?)`
	now := time.Now()
	_, err := DB.ExecContext(ctx, query,
		msg.ID, msg.ConversationID, msg.SenderID, msg.MsgType, msg.Content, msg.Extra, now)
	if err != nil {
		return err
	}

	// 更新会话最后消息
	DB.ExecContext(ctx,
		"UPDATE conversations SET last_msg_id = ?, last_msg_at = ?, updated_at = ? WHERE id = ?",
		msg.ID, now, now, msg.ConversationID)

	// 更新其他成员的未读计数
	DB.ExecContext(ctx,
		"UPDATE conversation_members SET unread_count = unread_count + 1 WHERE conversation_id = ? AND user_id != ?",
		msg.ConversationID, msg.SenderID)

	return nil
}

// GetMessages 获取消息历史 - 基于 cursor (msg_id) 分页
// Snowflake ID 天然有序，所以用 ID 做 cursor 比用时间戳更精确
func (r *ChatRepo) GetMessages(ctx context.Context, convID, cursor int64, limit int) ([]*model.Message, error) {
	query := `SELECT id, conversation_id, sender_id, msg_type, content, extra, status, created_at
              FROM messages WHERE conversation_id = ? AND id < ? AND status != 4
              ORDER BY id DESC LIMIT ?`
	if cursor == 0 {
		cursor = 1<<63 - 1
	}
	rows, err := DB.QueryContext(ctx, query, convID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		m := &model.Message{}
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.MsgType,
			&m.Content, &m.Extra, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// MarkRead 标记已读 - 水位线模型
func (r *ChatRepo) MarkRead(ctx context.Context, convID, userID, lastMsgID int64) error {
	_, err := DB.ExecContext(ctx,
		"UPDATE conversation_members SET last_read_msg = ?, unread_count = 0 WHERE conversation_id = ? AND user_id = ?",
		lastMsgID, convID, userID)
	return err
}

// ==================== 在线状态 (Redis) ====================

func onlineKey() string {
	return "chat:online"
}

func userWSKey(userID int64) string {
	return fmt.Sprintf("chat:ws:%d", userID)
}

// SetOnline 设置用户在线
func (r *ChatRepo) SetOnline(ctx context.Context, userID int64, serverAddr string) error {
	pipe := cache.RDB.Pipeline()
	pipe.SAdd(ctx, onlineKey(), userID)
	pipe.Set(ctx, userWSKey(userID), serverAddr, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// SetOffline 设置用户离线
func (r *ChatRepo) SetOffline(ctx context.Context, userID int64) error {
	pipe := cache.RDB.Pipeline()
	pipe.SRem(ctx, onlineKey(), userID)
	pipe.Del(ctx, userWSKey(userID))
	_, err := pipe.Exec(ctx)
	return err
}

// IsOnline 检查用户是否在线
func (r *ChatRepo) IsOnline(ctx context.Context, userID int64) (bool, error) {
	return cache.RDB.SIsMember(ctx, onlineKey(), userID).Result()
}

// GetUserServer 获取用户连接的服务器地址（多节点路由）
func (r *ChatRepo) GetUserServer(ctx context.Context, userID int64) (string, error) {
	val, err := cache.RDB.Get(ctx, userWSKey(userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
