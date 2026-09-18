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

// Helper: Sinh Custom ID duy nhất cho 2 User
func GetConversationID(userA, userB string) string {
	if strings.Compare(userA, userB) < 0 {
		return fmt.Sprintf("%s_%s", userA, userB)
	}
	return fmt.Sprintf("%s_%s", userB, userA)
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// Run khởi động vòng lặp quản lý Client và lắng nghe Redis Pub/Sub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.UserID] = client
			h.mutex.Unlock()

			// Lắng nghe Redis Pub/Sub cho riêng User này trên một goroutine
			go h.subscribeUserChannel(client.UserID)
			log.Printf("🟢 User connected: %s (ID: %s) | Tổng online: %d\n", client.Username, client.UserID, len(h.clients))

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("🔴 User disconnected: %s | Tổng online: %d\n", client.Username, len(h.clients))
		}
	}
}

// subscribeUserChannel lắng nghe Redis Channel của user đích
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
				// Nếu buffer đầy, đóng kết nối
				h.Unregister <- client
			}
		} else {
			// User đã offline, ngắt subscription goroutine
			break
		}
	}
}

// HandleClientEvent điều hướng sự kiện gửi lên từ Client
func (h *Hub) HandleClientEvent(client *Client, event models.WSEvent) {
	switch event.Event {
	case "chat:send":
		h.handleSendMessage(client, event.Payload)
	case "ping":
		// Trả về pong
		resp, _ := json.Marshal(models.WSEvent{Event: "pong", Payload: "pong"})
		client.Send <- resp
	default:
		log.Printf("Sự kiện không xác định: %s", event.Event)
	}
}

// handleSendMessage xử lý khi User gửi tin nhắn
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

	// 2. Cập nhật hoặc tạo cuộc hội thoại
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

	// 4. Trả gói tin ACK (Xác nhận đã gửi thành công) về cho người gửi
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
