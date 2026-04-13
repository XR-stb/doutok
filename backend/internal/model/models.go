package model

import (
	"time"
)

// ==================== 用户域 ====================

type User struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Phone       string    `json:"phone,omitempty"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"-"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Gender      int       `json:"gender"`
	Birthday    string    `json:"birthday,omitempty"`
	Status      int       `json:"status"`
	Role        string    `json:"role"`
	FollowCount int64     `json:"follow_count"`
	FanCount    int64     `json:"fan_count"`
	LikeCount   int64     `json:"like_count"`
	VideoCount  int64     `json:"video_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserFollow 关注列表项（JOIN 查询结果）
type UserFollow struct {
	TargetID  int64     `json:"target_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status"` // 1=following, 2=mutual
	CreatedAt time.Time `json:"created_at"`
}

// ==================== 视频域 ====================

type Video struct {
	ID           int64     `json:"id"`
	AuthorID     int64     `json:"author_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	CoverURL     string    `json:"cover_url"`
	PlayURL      string    `json:"play_url"`
	Duration     int       `json:"duration"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	FileSize     int64     `json:"file_size"`
	Status       int       `json:"status"`
	Visibility   int       `json:"visibility"`
	LikeCount    int64     `json:"like_count"`
	CommentCount int64     `json:"comment_count"`
	ShareCount   int64     `json:"share_count"`
	ViewCount    int64     `json:"view_count"`
	Tags         string    `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ==================== 评论域 ====================

type Comment struct {
	ID         int64     `json:"id"`
	VideoID    int64     `json:"video_id"`
	UserID     int64     `json:"user_id"`
	ParentID   int64     `json:"parent_id"`
	RootID     int64     `json:"root_id"`
	Content    string    `json:"content"`
	LikeCount  int64     `json:"like_count"`
	ReplyCount int       `json:"reply_count"`
	Status     int       `json:"status"`
	IPLocation string    `json:"ip_location"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ==================== 聊天域 ====================

type Message struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	MsgType        int       `json:"msg_type"`
	Content        string    `json:"content"`
	Extra          string    `json:"extra"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationItem 会话列表项（JOIN 查询结果）
type ConversationItem struct {
	ConversationID int64     `json:"conversation_id"`
	Type           int       `json:"type"`
	LastMsg        string    `json:"last_msg"`
	LastMsgType    int       `json:"last_msg_type"`
	LastMsgAt      time.Time `json:"last_msg_at"`
	UnreadCount    int       `json:"unread_count"`
	Muted          bool      `json:"muted"`
	PeerID         int64     `json:"peer_id"`
	PeerName       string    `json:"peer_name"`
	PeerAvatar     string    `json:"peer_avatar"`
}

// ==================== 直播域 ====================

type LiveRoom struct {
	ID          int64      `json:"id"`
	AnchorID    int64      `json:"anchor_id"`
	Title       string     `json:"title"`
	CoverURL    string     `json:"cover_url"`
	StreamKey   string     `json:"stream_key,omitempty"`
	Status      int        `json:"status"`
	ViewerCount int        `json:"viewer_count"`
	PeakViewer  int        `json:"peak_viewer"`
	LikeCount   int64      `json:"like_count"`
	GiftValue   int64      `json:"gift_value"`
	StartedAt   *time.Time `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type LiveGift struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	SenderID  int64     `json:"sender_id"`
	GiftID    int       `json:"gift_id"`
	GiftName  string    `json:"gift_name"`
	GiftValue int       `json:"gift_value"`
	Count     int       `json:"count"`
	Combo     int       `json:"combo"`
	CreatedAt time.Time `json:"created_at"`
}

// ==================== 行为日志 ====================

type UserBehavior struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TargetType int       `json:"target_type"` // 1=video, 2=live, 3=user
	TargetID   int64     `json:"target_id"`
	Action     string    `json:"action"` // view, like, comment, share, follow, gift, impression
	Duration   int       `json:"duration"`
	Extra      string    `json:"extra"`
	CreatedAt  time.Time `json:"created_at"`
}
