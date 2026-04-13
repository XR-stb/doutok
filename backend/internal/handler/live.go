package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xiaoran/doutok/internal/model"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/logger"
	"github.com/xiaoran/doutok/internal/pkg/response"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
	"github.com/xiaoran/doutok/internal/repository"
)

var liveRepo = repository.NewLiveRepo()

// ==================== WebSocket 直播间 Hub ====================

// LiveHub 管理所有直播间的 WebSocket 连接
// 每个直播间一个 Room, 每个 Room 有多个 Client
type LiveHub struct {
	mu    sync.RWMutex
	rooms map[int64]*LiveRoom
}

type LiveRoom struct {
	mu      sync.RWMutex
	clients map[int64]*websocket.Conn // userID -> conn
}

type LiveMessage struct {
	Type    string      `json:"type"`    // danmaku, gift, like, system, rank_update
	UserID  int64       `json:"user_id"`
	Name    string      `json:"name"`
	Content interface{} `json:"content"`
	Time    int64       `json:"time"`
}

var hub = &LiveHub{rooms: make(map[int64]*LiveRoom)}

func (h *LiveHub) getRoom(roomID int64) *LiveRoom {
	h.mu.RLock()
	room := h.rooms[roomID]
	h.mu.RUnlock()
	if room != nil {
		return room
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = &LiveRoom{clients: make(map[int64]*websocket.Conn)}
	}
	return h.rooms[roomID]
}

func (r *LiveRoom) broadcast(msg LiveMessage) {
	data, _ := json.Marshal(msg)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for uid, conn := range r.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			logger.Error("ws write failed", "user_id", uid, "err", err)
		}
	}
}

func (r *LiveRoom) addClient(userID int64, conn *websocket.Conn) {
	r.mu.Lock()
	r.clients[userID] = conn
	r.mu.Unlock()
}

func (r *LiveRoom) removeClient(userID int64) {
	r.mu.Lock()
	delete(r.clients, userID)
	r.mu.Unlock()
}

// ==================== HTTP Handlers ====================

func ListLiveRooms(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	rooms, err := liveRepo.ListLive(c.Request.Context(), offset, limit)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{
		"rooms":    rooms,
		"offset":   offset + len(rooms),
		"has_more": len(rooms) == limit,
	})
}

func GetLiveRoom(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	room, err := liveRepo.GetByID(c.Request.Context(), roomID)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	// 从 Redis 获取实时数据
	viewers, _ := liveRepo.GetViewerCount(c.Request.Context(), roomID)
	likes, _ := liveRepo.GetLikeCount(c.Request.Context(), roomID)

	response.Success(c, gin.H{
		"room":         room,
		"viewer_count": viewers,
		"like_count":   likes,
	})
}

type CreateLiveReq struct {
	Title    string `json:"title" binding:"required,max=128"`
	CoverURL string `json:"cover_url"`
}

func CreateLiveRoom(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req CreateLiveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}

	roomID := snowflake.GenID()
	streamKey := fmt.Sprintf("live_%d_%d", userID, time.Now().Unix())

	room := &model.LiveRoom{
		ID:        roomID,
		AnchorID:  userID,
		Title:     req.Title,
		CoverURL:  req.CoverURL,
		StreamKey: streamKey,
		Status:    0, // created, 等待推流
	}

	if err := liveRepo.Create(c.Request.Context(), room); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "创建失败")
		return
	}

	response.Success(c, gin.H{
		"room_id":    roomID,
		"stream_key": streamKey,
		"rtmp_url":   fmt.Sprintf("rtmp://localhost:1935/live/%s", streamKey),
		"play_url":   fmt.Sprintf("http://localhost:8080/live/%s.flv", streamKey),
	})
}

func UpdateLiveRoom(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	action := c.Query("action") // start, end

	ctx := c.Request.Context()
	switch action {
	case "start":
		liveRepo.StartLive(ctx, roomID)
		room := hub.getRoom(roomID)
		room.broadcast(LiveMessage{
			Type:    "system",
			Content: "直播已开始",
			Time:    time.Now().UnixMilli(),
		})
	case "end":
		liveRepo.EndLive(ctx, roomID)
		liveRepo.CleanupRoom(ctx, roomID)
		room := hub.getRoom(roomID)
		room.broadcast(LiveMessage{
			Type:    "system",
			Content: "直播已结束",
			Time:    time.Now().UnixMilli(),
		})
	}
	response.Success(c, gin.H{"msg": "ok"})
}

// GetLiveRank 获取直播间送礼排行榜 Top 10
// 使用 Redis ZSET, O(logN) 更新, O(K+logN) 获取 TopK
func GetLiveRank(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, "")
		return
	}

	topN := int64(10)
	if n, _ := strconv.ParseInt(c.DefaultQuery("top", "10"), 10, 64); n > 0 && n <= 100 {
		topN = n
	}

	ranks, err := liveRepo.GetTopRank(c.Request.Context(), roomID, topN)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	// 补充用户信息
	type RankItem struct {
		Rank     int     `json:"rank"`
		UserID   int64   `json:"user_id"`
		Nickname string  `json:"nickname"`
		Avatar   string  `json:"avatar"`
		Score    float64 `json:"score"`
	}

	items := make([]RankItem, 0, len(ranks))
	for i, r := range ranks {
		uid, _ := strconv.ParseInt(r.Member.(string), 10, 64)
		user, _ := userRepo.GetByID(c.Request.Context(), uid)
		item := RankItem{
			Rank:   i + 1,
			UserID: uid,
			Score:  r.Score,
		}
		if user != nil {
			item.Nickname = user.Nickname
			item.Avatar = user.Avatar
		}
		items = append(items, item)
	}

	response.Success(c, gin.H{"rank": items})
}

type SendGiftReq struct {
	GiftID    int    `json:"gift_id" binding:"required"`
	GiftName  string `json:"gift_name" binding:"required"`
	GiftValue int    `json:"gift_value" binding:"required"`
	Count     int    `json:"count"`
	Combo     int    `json:"combo"`
}

// SendGift 送礼
// 写 MySQL 持久化 + 更新 Redis 排行榜 + WebSocket 广播动画
func SendGift(c *gin.Context) {
	userID := c.GetInt64("user_id")
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req SendGiftReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Combo <= 0 {
		req.Combo = 1
	}

	ctx := c.Request.Context()
	gift := &model.LiveGift{
		RoomID:    roomID,
		SenderID:  userID,
		GiftID:    req.GiftID,
		GiftName:  req.GiftName,
		GiftValue: req.GiftValue,
		Count:     req.Count,
		Combo:     req.Combo,
	}

	// 1. MySQL 持久化
	if err := liveRepo.SendGift(ctx, gift); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "送礼失败")
		return
	}

	// 2. Redis 更新排行榜
	score := float64(req.GiftValue * req.Count)
	liveRepo.UpdateRank(ctx, roomID, userID, score)

	// 3. WebSocket 广播礼物动画
	user, _ := userRepo.GetByID(ctx, userID)
	nickname := "匿名"
	if user != nil {
		nickname = user.Nickname
	}

	room := hub.getRoom(roomID)
	room.broadcast(LiveMessage{
		Type:   "gift",
		UserID: userID,
		Name:   nickname,
		Content: gin.H{
			"gift_id":    req.GiftID,
			"gift_name":  req.GiftName,
			"gift_value": req.GiftValue,
			"count":      req.Count,
			"combo":      req.Combo,
		},
		Time: time.Now().UnixMilli(),
	})

	// 4. 记录行为
	go behaviorRepo.Record(ctx, &model.UserBehavior{
		UserID:     userID,
		TargetType: 2, // live
		TargetID:   roomID,
		Action:     "gift",
	})

	response.Success(c, gin.H{"msg": "ok"})
}

// LikeLive 直播点赞 - 高并发，纯 Redis
func LikeLive(c *gin.Context) {
	userID := c.GetInt64("user_id")
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	count, err := liveRepo.IncrLike(c.Request.Context(), roomID)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	// 每 100 个赞广播一次（避免消息风暴）
	if count%100 == 0 {
		room := hub.getRoom(roomID)
		room.broadcast(LiveMessage{
			Type:    "like",
			UserID:  userID,
			Content: gin.H{"total": count},
			Time:    time.Now().UnixMilli(),
		})
	}

	response.Success(c, gin.H{"total_likes": count})
}

// ==================== WebSocket 直播弹幕 ====================

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 开发环境允许跨域
}

// WSLive WebSocket 直播间连接
// 客户端连接后可以发送弹幕，接收弹幕/礼物/系统消息
func WSLive(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()
	room := hub.getRoom(roomID)
	room.addClient(userID, conn)
	liveRepo.AddViewer(ctx, roomID, userID)

	// 通知有人进入
	user, _ := userRepo.GetByID(ctx, userID)
	nickname := "游客"
	if user != nil {
		nickname = user.Nickname
	}
	room.broadcast(LiveMessage{
		Type:    "system",
		UserID:  userID,
		Name:    nickname,
		Content: fmt.Sprintf("%s 进入了直播间", nickname),
		Time:    time.Now().UnixMilli(),
	})

	logger.Info("viewer joined live", "room_id", roomID, "user_id", userID)

	// 读取客户端消息（弹幕）
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var danmaku struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(msg, &danmaku); err != nil || danmaku.Content == "" {
			continue
		}

		// 广播弹幕
		room.broadcast(LiveMessage{
			Type:    "danmaku",
			UserID:  userID,
			Name:    nickname,
			Content: danmaku.Content,
			Time:    time.Now().UnixMilli(),
		})
	}

	// 断开连接清理
	room.removeClient(userID)
	liveRepo.RemoveViewer(ctx, roomID, userID)
	logger.Info("viewer left live", "room_id", roomID, "user_id", userID)
}
