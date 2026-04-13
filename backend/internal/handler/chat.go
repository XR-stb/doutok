package handler

import (
	"encoding/json"
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

var chatRepo = repository.NewChatRepo()

// ==================== Chat WebSocket Hub ====================

// ChatHub 管理所有聊天 WebSocket 连接
// 生产环境: 多实例部署时通过 Redis Pub/Sub 做跨实例消息路由
type ChatHub struct {
	mu      sync.RWMutex
	clients map[int64]*websocket.Conn // userID -> conn
}

var chatHub = &ChatHub{clients: make(map[int64]*websocket.Conn)}

func (h *ChatHub) addClient(userID int64, conn *websocket.Conn) {
	h.mu.Lock()
	// 踢掉旧连接（一个用户同时只能有一个 WS 连接）
	if old, ok := h.clients[userID]; ok {
		old.Close()
	}
	h.clients[userID] = conn
	h.mu.Unlock()
}

func (h *ChatHub) removeClient(userID int64) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
}

func (h *ChatHub) sendToUser(userID int64, data []byte) bool {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	return conn.WriteMessage(websocket.TextMessage, data) == nil
}

// ==================== HTTP Handlers ====================

func ListConversations(c *gin.Context) {
	userID := c.GetInt64("user_id")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, err := chatRepo.ListConversations(c.Request.Context(), userID, offset, limit)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{
		"conversations": items,
		"offset":        offset + len(items),
	})
}

type CreateConvReq struct {
	PeerID int64 `json:"peer_id" binding:"required"`
}

func CreateConversation(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req CreateConvReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}

	ctx := c.Request.Context()

	// 检查是否已有私聊会话
	convID, _ := chatRepo.FindPrivateConversation(ctx, userID, req.PeerID)
	if convID > 0 {
		response.Success(c, gin.H{"conversation_id": convID})
		return
	}

	// 创建新会话
	convID = snowflake.GenID()
	if err := chatRepo.CreateConversation(ctx, convID, userID, req.PeerID); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "创建失败")
		return
	}

	response.Success(c, gin.H{"conversation_id": convID})
}

func GetMessages(c *gin.Context) {
	convID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cursor, _ := strconv.ParseInt(c.DefaultQuery("cursor", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	messages, err := chatRepo.GetMessages(c.Request.Context(), convID, cursor, limit)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	// 标记已读
	userID := c.GetInt64("user_id")
	if len(messages) > 0 {
		chatRepo.MarkRead(c.Request.Context(), convID, userID, messages[0].ID)
	}

	var nextCursor int64
	if len(messages) > 0 {
		nextCursor = messages[len(messages)-1].ID
	}

	response.Success(c, gin.H{
		"messages":    messages,
		"next_cursor": nextCursor,
		"has_more":    len(messages) == limit,
	})
}

type SendMsgReq struct {
	ConversationID int64  `json:"conversation_id" binding:"required"`
	MsgType        int    `json:"msg_type"` // 1=text, 2=image, 3=video
	Content        string `json:"content" binding:"required"`
	Extra          string `json:"extra"`
}

func SendMessage(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req SendMsgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}
	if req.MsgType == 0 {
		req.MsgType = 1
	}

	msg := &model.Message{
		ID:             snowflake.GenID(),
		ConversationID: req.ConversationID,
		SenderID:       userID,
		MsgType:        req.MsgType,
		Content:        req.Content,
		Extra:          req.Extra,
	}

	ctx := c.Request.Context()
	if err := chatRepo.SaveMessage(ctx, msg); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "发送失败")
		return
	}

	// 尝试通过 WebSocket 实时推送给在线的对方
	// 生产版: 查 Redis 获取对方所在服务器 -> 通过 Redis Pub/Sub 路由
	go pushMessageToRecipients(msg)

	response.Success(c, gin.H{
		"msg_id": msg.ID,
	})
}

func pushMessageToRecipients(msg *model.Message) {
	data, _ := json.Marshal(gin.H{
		"type":            "new_message",
		"msg_id":          msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"msg_type":        msg.MsgType,
		"content":         msg.Content,
		"time":            time.Now().UnixMilli(),
	})

	// 查找会话的所有成员
	// 简化版：直接查 ChatHub 的所有连接，向非发送者推送
	chatHub.mu.RLock()
	defer chatHub.mu.RUnlock()
	for uid, conn := range chatHub.clients {
		if uid != msg.SenderID {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// WSChat WebSocket 聊天连接
// 客户端建立长连接后，可以收到实时消息推送
func WSChat(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("chat ws upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()
	chatHub.addClient(userID, conn)
	chatRepo.SetOnline(ctx, userID, "local")

	logger.Info("chat ws connected", "user_id", userID)

	// 心跳 + 读消息
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动心跳
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// 读取客户端消息（客户端也可以通过 WS 发消息，但主流程走 HTTP）
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	close(done)
	chatHub.removeClient(userID)
	chatRepo.SetOffline(ctx, userID)
	logger.Info("chat ws disconnected", "user_id", userID)
}
