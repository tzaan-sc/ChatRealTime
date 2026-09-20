package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"

	"github.com/redis/go-redis/v9"
)

type Hub struct {
	clients     map[string]*Client // Map: UserID -> *Client
	register    chan *Client
	Unregister  chan *Client
	mutex       sync.RWMutex
	redisClient *redis.Client
	msgRepo     *repository.MessageRepository
	convRepo    *repository.ConversationRepository
}

func NewHub(redisClient *redis.Client, msgRepo *repository.MessageRepository, convRepo *repository.ConversationRepository) *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		register:    make(chan *Client),
		Unregister:  make(chan *Client),
		redisClient: redisClient,
		msgRepo:     msgRepo,
		convRepo:    convRepo,
	}
}

func GetConversationID(userA, userB string) string {
	if strings.Compare(userA, userB) < 0 {
		return fmt.Sprintf("%s_%s", userA, userB)
	}
	return fmt.Sprintf("%s_%s", userB, userA)
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// BroadcastUserStatus phát trạng thái online/offline cho toàn hệ thống
func (h *Hub) BroadcastUserStatus(userID string, isOnline bool) {
	ctx := context.Background()
	if isOnline {
		h.redisClient.Set(ctx, fmt.Sprintf("user:online:%s", userID), "1", 30*time.Second)
	} else {
		h.redisClient.Del(ctx, fmt.Sprintf("user:online:%s", userID))
	}

	event := models.WSEvent{
		Event: "user:status",
		Payload: map[string]interface{}{
			"user_id":   userID,
			"is_online": isOnline,
		},
	}
	bytes, _ := json.Marshal(event)
	h.redisClient.Publish(ctx, "global:events", string(bytes))
}

// Run khởi động vòng lặp Hub
func (h *Hub) Run() {
	// Lắng nghe channel toàn cục trên 1 goroutine
	go h.subscribeGlobalEvents()

	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.UserID] = client
			h.mutex.Unlock()

			// Lắng nghe kênh riêng của user
			go h.subscribeUserChannel(client.UserID)

			// Đánh dấu online và thông báo cho mọi người
			h.BroadcastUserStatus(client.UserID, true)

			// Gửi danh sách toàn bộ user đang online hiện tại về cho riêng client này
			go h.sendInitialOnlineList(client)

			log.Printf("🟢 User connected: %s (ID: %s) | Tổng online: %d\n", client.Username, client.UserID, len(h.clients))

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mutex.Unlock()

			// Đánh dấu offline và thông báo
			h.BroadcastUserStatus(client.UserID, false)

			log.Printf("🔴 User disconnected: %s | Tổng online: %d\n", client.Username, len(h.clients))
		}
	}
}

// sendInitialOnlineList gửi danh sách những ai đang online khi client vừa kết nối
func (h *Hub) sendInitialOnlineList(client *Client) {
	ctx := context.Background()
	keys, err := h.redisClient.Keys(ctx, "user:online:*").Result()
	if err != nil {
		return
	}

	onlineIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) == 3 {
			onlineIDs = append(onlineIDs, parts[2])
		}
	}

	event := models.WSEvent{
		Event:   "user:online_list",
		Payload: onlineIDs,
	}
	bytes, _ := json.Marshal(event)
	client.Send <- bytes
}

// subscribeGlobalEvents lắng nghe các event phát cho toàn hệ thống
func (h *Hub) subscribeGlobalEvents() {
	pubsub := h.redisClient.Subscribe(context.Background(), "global:events")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		h.mutex.RLock()
		for _, client := range h.clients {
			select {
			case client.Send <- []byte(msg.Payload):
			default:
			}
		}
		h.mutex.RUnlock()
	}
}

// subscribeUserChannel lắng nghe Redis Channel của riêng user đích
func (h *Hub) subscribeUserChannel(userID string) {
	channelName := fmt.Sprintf("user:chat:%s", userID)
	pubsub := h.redisClient.Subscribe(context.Background(), channelName)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		h.mutex.RLock()
		client, isOnline := h.clients[userID]
		h.mutex.RUnlock()

		if isOnline {
			select {
			case client.Send <- []byte(msg.Payload):
			default:
				h.Unregister <- client
			}
		} else {
			break
		}
	}
}

// HandleClientEvent điều hướng sự kiện gửi lên từ Client
func (h *Hub) HandleClientEvent(client *Client, event models.WSEvent) {
	switch event.Event {
	case "chat:send":
		h.handleSendMessage(client, event.Payload)

	case "heartbeat":
		// Gia hạn TTL 30 giây trong Redis
		ctx := context.Background()
		h.redisClient.Set(ctx, fmt.Sprintf("user:online:%s", client.UserID), "1", 30*time.Second)

	case "typing:start", "typing:stop":
		h.handleTyping(client, event.Event, event.Payload)

	case "chat:read":
		h.handleMarkAsRead(client, event.Payload)

	case "ping":
		resp, _ := json.Marshal(models.WSEvent{Event: "pong", Payload: "pong"})
		client.Send <- resp

	default:
		log.Printf("Sự kiện không xác định: %s", event.Event)
	}
}

// handleSendMessage xử lý gửi tin nhắn
func (h *Hub) handleSendMessage(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}

	var req models.SendMessageRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}

	if strings.TrimSpace(req.Content) == "" || req.ReceiverID == "" {
		return
	}

	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	ctx := context.Background()
	convID := GetConversationID(client.UserID, req.ReceiverID)

	// 1. Lưu tin nhắn vào MongoDB
	newMsg := &models.Message{
		ConversationID: convID,
		SenderID:       client.UserID,
		ReceiverID:     req.ReceiverID,
		Content:        req.Content,
		Type:           msgType,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := h.msgRepo.Create(ctx, newMsg); err != nil {
		log.Printf("Lỗi lưu tin nhắn vào MongoDB: %v", err)
		return
	}

	// 2. Cập nhật hội thoại
	_, _ = h.convRepo.GetOrCreate(ctx, convID, client.UserID, req.ReceiverID)
	_ = h.convRepo.UpdateLastMessage(ctx, convID, req.Content, client.UserID)

	// 3. Đóng gói Event bắn qua Redis Pub/Sub cho người nhận
	eventReceive := models.WSEvent{
		Event:   "chat:receive",
		Payload: newMsg,
	}
	eventReceiveBytes, _ := json.Marshal(eventReceive)
	receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
	h.redisClient.Publish(ctx, receiverChannel, string(eventReceiveBytes))

	// 4. Trả gói tin ACK về cho người gửi
	eventACK := models.WSEvent{
		Event: "chat:ack",
		Payload: map[string]interface{}{
			"temp_id":    newMsg.ID.Hex(),
			"message":    newMsg,
			"created_at": newMsg.CreatedAt,
		},
	}
	eventACKBytes, _ := json.Marshal(eventACK)
	client.Send <- eventACKBytes
}

// handleTyping xử lý Typing Indicator (chuyển tiếp tức thời, không lưu DB)
func (h *Hub) handleTyping(client *Client, eventType string, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}

	var req struct {
		ReceiverID string `json:"receiver_id"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil || req.ReceiverID == "" {
		return
	}

	forwardEvent := models.WSEvent{
		Event: eventType,
		Payload: map[string]string{
			"sender_id": client.UserID,
		},
	}
	forwardBytes, _ := json.Marshal(forwardEvent)
	receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
	h.redisClient.Publish(context.Background(), receiverChannel, string(forwardBytes))
}

// handleMarkAsRead xử lý đánh dấu đã đọc tin nhắn
func (h *Hub) handleMarkAsRead(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}

	var req struct {
		ConversationID string `json:"conversation_id"`
		PartnerID      string `json:"partner_id"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil || req.ConversationID == "" {
		return
	}

	ctx := context.Background()
	// Đánh dấu đã đọc trong MongoDB
	_ = h.msgRepo.MarkAsRead(ctx, req.ConversationID, client.UserID)

	// Bắn sự kiện xác nhận về cho bạn chat (người gửi trước đó)
	if req.PartnerID != "" {
		readAckEvent := models.WSEvent{
			Event: "chat:read_ack",
			Payload: map[string]string{
				"conversation_id": req.ConversationID,
				"reader_id":       client.UserID,
			},
		}
		bytes, _ := json.Marshal(readAckEvent)
		partnerChannel := fmt.Sprintf("user:chat:%s", req.PartnerID)
		h.redisClient.Publish(ctx, partnerChannel, string(bytes))
	}
}
