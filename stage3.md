# Hướng Dẫn Chi Tiết Giai Đoạn 3: Xây Dựng WebSocket Hub & Nhắn Tin Thời Gian Thực
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Backend: Golang (Gorilla WebSocket, Goroutines) | Tầng trung gian: Redis Pub/Sub | CSDL: MongoDB

Tài liệu này sẽ hướng dẫn bạn từng bước xây dựng một hệ thống **WebSocket Hub** hoàn chỉnh trong Backend Go, kết nối Redis Pub/Sub để định tuyến tin nhắn siêu tốc, lưu trữ dữ liệu vào MongoDB và cung cấp API lấy lịch sử trò chuyện.

---

## 📋 Checklist Tiến Độ Giai Đoạn 3

- [ ] **1. Cài đặt thư viện WebSocket (`gorilla/websocket`):**
  ```powershell
  go get github.com/gorilla/websocket
  ```
- [ ] **2. Tạo Models cho Hội thoại & Tin nhắn:**
  - `internal/models/message.go`
  - `internal/models/conversation.go`
  - `internal/models/ws_event.go`
- [ ] **3. Xây dựng tầng Repository (MongoDB):**
  - `internal/repository/message_repo.go`
  - `internal/repository/conversation_repo.go`
- [ ] **4. Xây dựng WebSocket Hub & Client:**
  - `internal/websocket/client.go` (ReadPump, WritePump, Ping/Pong)
  - `internal/websocket/hub.go` (Client Map, Register, Unregister, Redis Pub/Sub)
- [ ] **5. Viết tầng Service & Handlers:**
  - `internal/service/chat_service.go`
  - `internal/handlers/chat_handler.go`
  - `internal/handlers/ws_handler.go`
- [ ] **6. Cập nhật Router trong `cmd/server/main.go`**
- [ ] **7. Kiểm thử WebSocket & Chat 1-1 qua Postman / Browser Console**

---

## Mục lục
1. [Bước 1: Cài đặt thư viện WebSocket](#buoc-1-cai-dat-thu-vien-websocket)
2. [Bước 2: Hiểu cấu trúc luồng dữ liệu WebSocket & Redis Pub/Sub](#buoc-2-hieu-cau-truc-luong-du-lieu)
3. [Bước 3: Khởi tạo Models & DTO](#buoc-3-khoi-tao-models--dto)
4. [Bước 4: Tầng Repository (Thao tác MongoDB)](#buoc-4-tang-repository)
5. [Bước 5: Tầng WebSocket Hub & Client Concurrency](#buoc-5-tang-websocket-hub--client)
6. [Bước 6: Tầng Service & Handlers](#buoc-6-tang-service--handlers)
7. [Bước 7: Cập nhật `cmd/server/main.go`](#buoc-7-cap-nhat-main-go)
8. [Bước 8: Chạy Server và Kiểm thử WebSocket](#buoc-8-chay-server-va-kiem-thu-websocket)
9. [Các lỗi thường gặp và cách xử lý](#cac-loi-thuong-gap-va-cach-xu-ly)

---

<a name="buoc-1-cai-dat-thu-vien-websocket"></a>
## Bước 1: Cài đặt thư viện WebSocket

Mở Terminal tại thư mục `backend` và chạy lệnh:
```powershell
cd d:\GIT\ChatRealTime\backend
go get github.com/gorilla/websocket
go mod tidy
```

---

<a name="buoc-2-hieu-cau-truc-luong-du-lieu"></a>
## Bước 2: Hiểu cấu trúc luồng dữ liệu (WebSocket + Redis Pub/Sub)

```
[Client A (Alex)] --- (1) Gửi tin JSON qua WebSocket ---> [Go Server]
                                                              |
    +---------------------------------------------------------+
    |
    |---> (2) Lưu tin nhắn vào MongoDB (Collection: messages)
    |---> (3) Publish tin nhắn vào Redis Channel: "user:chat:<Receiver_ID>"
    |---> (4) Trả ACK "Đã gửi" về Client A
    |
    v
[Redis Pub/Sub Channel] ---> [Go Server lắng nghe (Subscribe)]
                                   |
                                   v
                   (5) Đẩy tin xuống socket [Client B (Bob)]
```

### 💡 Quy tắc mã định danh cuộc hội thoại (`conversation_id`):
Để đảm bảo khi User A chat với User B hay User B chat với User A thì **luôn luôn thuộc về cùng 1 cuộc hội thoại**, ta sắp xếp 2 User ID theo thứ tự alphabet:
- Ví dụ: `id1 = "abc"`, `id2 = "xyz"` $\to$ `conversation_id = "abc_xyz"`.

---

<a name="buoc-3-khoi-tao-models--dto"></a>
## Bước 3: Khởi tạo Models & DTO

### 3.1. File `backend/internal/models/message.go`
```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message đại diện cho 1 tin nhắn trong collection "messages"
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id" json:"conversation_id"`
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	ReceiverID     string             `bson:"receiver_id" json:"receiver_id"`
	Content        string             `bson:"content" json:"content"`
	Type           string             `bson:"type" json:"type"` // "text", "image", "file"
	IsRead         bool               `bson:"is_read" json:"is_read"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// SendMessageRequest DTO gửi từ client
type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       string `json:"type"` // Mặc định là "text"
}
```

### 3.2. File `backend/internal/models/conversation.go`
```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation đại diện cho 1 cuộc trò chuyện 1-1
type Conversation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomID      string             `bson:"custom_id" json:"custom_id"` // userA_userB (sort alphabet)
	Members       []string           `bson:"members" json:"members"`     // [userAID, userBID]
	LastMessage   string             `bson:"last_message" json:"last_message"`
	LastSenderID  string             `bson:"last_sender_id" json:"last_sender_id"`
	LastMessageAt time.Time          `bson:"last_message_at" json:"last_message_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// ConversationResponse DTO trả về cho Frontend danh sách chat
type ConversationResponse struct {
	Conversation Conversation `json:"conversation"`
	OtherUser    User         `json:"other_user"`
	UnreadCount  int64        `json:"unread_count"`
}
```

### 3.3. File `backend/internal/models/ws_event.go`
```go
package models

// WSEvent chuẩn hóa cấu trúc gói tin gửi nhận qua WebSocket
type WSEvent struct {
	Event   string      `json:"event"`   // "chat:send", "chat:receive", "chat:ack", "ping", "pong"
	Payload interface{} `json:"payload"` // Nội dung dữ liệu tương ứng
}
```

---

<a name="buoc-4-tang-repository"></a>
## Bước 4: Tầng Repository (Thao tác MongoDB)

### 4.1. File `backend/internal/repository/message_repo.go`
```go
package repository

import (
	"context"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection("messages"),
	}
}

// Create chèn tin nhắn mới
func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	msg.ID = primitive.NewObjectID()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, msg)
	return err
}

// GetByConversation lấy lịch sử tin nhắn có phân trang (mới nhất trước)
func (r *MessageRepository) GetByConversation(ctx context.Context, conversationID string, limit int64, offset int64) ([]models.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}). // Mới nhất lên đầu
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, bson.M{"conversation_id": conversationID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Đảo lại thứ tự tăng dần theo thời gian để client hiển thị từ trên xuống dưới
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountUnread đếm số tin nhắn chưa đọc của một người trong 1 cuộc trò chuyện
func (r *MessageRepository) CountUnread(ctx context.Context, conversationID, receiverID string) (int64, error) {
	filter := bson.M{
		"conversation_id": conversationID,
		"receiver_id":     receiverID,
		"is_read":         false,
	}
	return r.collection.CountDocuments(ctx, filter)
}

// MarkAsRead đánh dấu các tin nhắn trong cuộc trò chuyện là đã đọc
func (r *MessageRepository) MarkAsRead(ctx context.Context, conversationID, receiverID string) error {
	filter := bson.M{
		"conversation_id": conversationID,
		"receiver_id":     receiverID,
		"is_read":         false,
	}
	update := bson.M{"$set": bson.M{"is_read": true}}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}
```

### 4.2. File `backend/internal/repository/conversation_repo.go`
```go
package repository

import (
	"context"
	"errors"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationRepository struct {
	collection *mongo.Collection
}

func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{
		collection: db.Collection("conversations"),
	}
}

// GetOrCreate tìm hoặc tạo mới conversation giữa 2 người
func (r *ConversationRepository) GetOrCreate(ctx context.Context, customID string, member1, member2 string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.collection.FindOne(ctx, bson.M{"custom_id": customID}).Decode(&conv)
	if err == nil {
		return &conv, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		now := time.Now()
		newConv := models.Conversation{
			ID:            primitive.NewObjectID(),
			CustomID:      customID,
			Members:       []string{member1, member2},
			LastMessage:   "",
			LastSenderID:  "",
			LastMessageAt: now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		_, err := r.collection.InsertOne(ctx, newConv)
		if err != nil {
			return nil, err
		}
		return &newConv, nil
	}

	return nil, err
}

// UpdateLastMessage cập nhật tin nhắn cuối cùng và thời gian
func (r *ConversationRepository) UpdateLastMessage(ctx context.Context, customID, lastMsg, senderID string) error {
	now := time.Now()
	filter := bson.M{"custom_id": customID}
	update := bson.M{
		"$set": bson.M{
			"last_message":    lastMsg,
			"last_sender_id":  senderID,
			"last_message_at": now,
			"updated_at":      now,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetUserConversations lấy danh sách hội thoại của 1 user, sắp xếp mới nhất lên đầu
func (r *ConversationRepository) GetUserConversations(ctx context.Context, userID string) ([]models.Conversation, error) {
	opts := options.Find().SetSort(bson.D{{Key: "last_message_at", Value: -1}})
	filter := bson.M{"members": userID}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Conversation
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
```

---

<a name="buoc-5-tang-websocket-hub--client"></a>
## Bước 5: Tầng WebSocket Hub & Client Concurrency

### 5.1. File `backend/internal/websocket/client.go`
```go
package websocket

import (
	"encoding/json"
	"log"
	"time"

	"chatrealtime-backend/internal/models"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   string
	Username string
}

// ReadPump đọc tin nhắn từ socket gửi lên Go server
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Lỗi kết nối client %s: %v", c.Username, err)
			}
			break
		}

		var event models.WSEvent
		if err := json.Unmarshal(messageBytes, &event); err != nil {
			log.Printf("Gói tin không hợp lệ từ %s: %v", c.Username, err)
			continue
		}

		// Xử lý sự kiện từ Client
		c.Hub.HandleClientEvent(c, event)
	}
}

// WritePump ghi tin nhắn từ channel Go server xuống socket client
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Gom các message còn lại trong buffer để gửi 1 lần (tối ưu thông lượng)
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
```

### 5.2. File `backend/internal/websocket/hub.go`
```go
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
	clients          map[string]*Client // Map: UserID -> *Client
	register         chan *Client
	Unregister       chan *Client
	mutex            sync.RWMutex
	redisClient      *redis.Client
	msgRepo          *repository.MessageRepository
	convRepo         *repository.ConversationRepository
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
```

---

<a name="buoc-6-tang-service--handlers"></a>
## Bước 6: Tầng Service & Handlers

### 6.1. File `backend/internal/service/chat_service.go`
```go
package service

import (
	"context"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"
)

type ChatService struct {
	msgRepo  *repository.MessageRepository
	convRepo *repository.ConversationRepository
	userRepo *repository.UserRepository
}

func NewChatService(msgRepo *repository.MessageRepository, convRepo *repository.ConversationRepository, userRepo *repository.UserRepository) *ChatService {
	return &ChatService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
		userRepo: userRepo,
	}
}

// GetUserConversations lấy danh sách hội thoại kèm thông tin đối phương và số tin chưa đọc
func (s *ChatService) GetUserConversations(ctx context.Context, currentUserID string) ([]models.ConversationResponse, error) {
	convs, err := s.convRepo.GetUserConversations(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	var responses []models.ConversationResponse
	for _, conv := range convs {
		// Tìm ID của người đối diện
		otherUserID := conv.Members[0]
		if otherUserID == currentUserID && len(conv.Members) > 1 {
			otherUserID = conv.Members[1]
		}

		otherUser, err := s.userRepo.FindByID(ctx, otherUserID)
		if err != nil || otherUser == nil {
			continue
		}

		unread, _ := s.msgRepo.CountUnread(ctx, conv.CustomID, currentUserID)

		responses = append(responses, models.ConversationResponse{
			Conversation: conv,
			OtherUser:    *otherUser,
			UnreadCount:  unread,
		})
	}

	return responses, nil
}

// GetMessages lấy lịch sử tin nhắn
func (s *ChatService) GetMessages(ctx context.Context, conversationID string, limit, offset int64) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.msgRepo.GetByConversation(ctx, conversationID, limit, offset)
}

// MarkMessagesAsRead đánh dấu tin nhắn là đã đọc
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, conversationID, currentUserID string) error {
	return s.msgRepo.MarkAsRead(ctx, conversationID, currentUserID)
}

// GetAllUsers lấy danh bạ để bắt đầu chat mới (trừ bản thân)
func (s *ChatService) GetAllUsers(ctx context.Context, currentUserID string) ([]models.User, error) {
	return s.userRepo.FindAllExcept(ctx, currentUserID)
}
```

> 💡 **Bổ sung hàm `FindAllExcept` vào `backend/internal/repository/user_repo.go`:**
```go
// FindAllExcept lấy danh sách user khác bản thân
func (r *UserRepository) FindAllExcept(ctx context.Context, currentUserID string) ([]models.User, error) {
	objID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$ne": objID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
```

### 6.2. File `backend/internal/handlers/chat_handler.go`
```go
package handlers

import (
	"net/http"
	"strconv"

	"chatrealtime-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// GetConversations API lấy danh sách chat
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	convs, err := h.chatService.GetUserConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": convs})
}

// GetMessages API lấy lịch sử tin nhắn
func (h *ChatHandler) GetMessages(c *gin.Context) {
	convID := c.Param("conversation_id")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	messages, err := h.chatService.GetMessages(c.Request.Context(), convID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// MarkAsRead API đánh dấu đã xem
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	convID := c.Param("conversation_id")
	userID := c.GetString("user_id")

	if err := h.chatService.MarkMessagesAsRead(c.Request.Context(), convID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã đánh dấu đã đọc"})
}

// GetUsers API lấy danh sách bạn bè để chọn chat
func (h *ChatHandler) GetUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	users, err := h.chatService.GetAllUsers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}
```

### 6.3. File `backend/internal/handlers/ws_handler.go`
```go
package handlers

import (
	"log"
	"net/http"
	"os"

	"chatrealtime-backend/internal/websocket"
	"chatrealtime-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Cho phép mọi domain kết nối WebSocket (tránh bị chặn CORS)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	hub *websocket.Hub
}

func NewWSHandler(hub *websocket.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleWS nâng cấp kết nối HTTP lên WebSocket
func (h *WSHandler) HandleWS(c *gin.Context) {
	// Lấy token từ Query Parameter: ws://localhost:8080/ws?token=<jwt_token>
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu có token"})
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	claims, err := utils.ValidateToken(tokenString, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Lỗi nâng cấp kết nối WebSocket: %v", err)
		return
	}

	client := &websocket.Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   claims.UserID,
		Username: claims.Username,
	}

	// Đăng ký client vào Hub
	h.hub.RegisterClient(client)

	// Bật 2 goroutines song song để đọc & ghi
	go client.WritePump()
	go client.ReadPump()
}
```

---

<a name="buoc-7-cap-nhat-main-go"></a>
## Bước 7: Cập nhật `cmd/server/main.go`

Cập nhật lại toàn bộ `backend/cmd/server/main.go`:
```go
package main

import (
	"log"
	"os"

	"chatrealtime-backend/internal/database"
	"chatrealtime-backend/internal/handlers"
	"chatrealtime-backend/internal/middleware"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/internal/service"
	"chatrealtime-backend/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Nạp biến môi trường
	if err := godotenv.Load(); err != nil {
		log.Println("Chú ý: Sử dụng biến môi trường hệ thống")
	}

	// 2. Khởi tạo Database (MongoDB & Redis)
	database.InitDatabase()

	// 3. Khởi tạo các Repositories
	userRepo := repository.NewUserRepository(database.MongoDB)
	msgRepo := repository.NewMessageRepository(database.MongoDB)
	convRepo := repository.NewConversationRepository(database.MongoDB)

	// 4. Khởi tạo WebSocket Hub và chạy ngầm
	hub := websocket.NewHub(database.RedisClient, msgRepo, convRepo)
	go hub.Run()

	// 5. Khởi tạo Services
	authService := service.NewAuthService(userRepo)
	chatService := service.NewChatService(msgRepo, convRepo, userRepo)

	// 6. Khởi tạo Handlers
	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService)
	wsHandler := handlers.NewWSHandler(hub)

	// 7. Khởi tạo Router
	r := gin.Default()
	r.Use(cors.Default())

	// Endpoint WebSocket
	r.GET("/ws", wsHandler.HandleWS)

	// REST API Routes
	api := r.Group("/api")
	{
		// Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.GetMe)
		}

		// Chat Routes (Cần xác thực Token)
		chat := api.Group("/chat", middleware.AuthMiddleware())
		{
			chat.GET("/conversations", chatHandler.GetConversations)
			chat.GET("/messages/:conversation_id", chatHandler.GetMessages)
			chat.POST("/messages/:conversation_id/read", chatHandler.MarkAsRead)
			chat.GET("/users", chatHandler.GetUsers)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server Chat Realtime đang chạy tại http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Không thể khởi động server: %v", err)
	}
}
```

---

<a name="buoc-8-chay-server-va-kiem-thu-websocket"></a>
## Bước 8: Chạy Server và Kiểm thử WebSocket

### 8.1. Khởi động Server
```powershell
cd d:\GIT\ChatRealTime\backend
go run cmd/server/main.go
```

---

### 8.2. Kịch bản kiểm thử Chat 2 người (Alex & Bob)

#### Bước A: Đăng ký 2 tài khoản qua Postman / Thunder Client
1. **Tài khoản 1 (Alex):**
   - `POST http://localhost:8080/api/auth/register` $\to$ nhận `token_alex` và `id_alex`.
2. **Tài khoản 2 (Bob):**
   - `POST http://localhost:8080/api/auth/register` (Username: `bob`, Password: `password123`) $\to$ nhận `token_bob` và `id_bob`.

---

#### Bước B: Kết nối WebSocket & Nhắn tin thời gian thực

##### 👉 Kết nối bằng Postman (Hỗ trợ WebSocket trực tiếp):
1. Trong Postman, bấm **New** $\to$ chọn **WebSocket Request**.
2. **Tab 1 (Alex):**
   - URL: `ws://localhost:8080/ws?token=<token_alex>` $\to$ Bấm **Connect**.
3. **Tab 2 (Bob):**
   - URL: `ws://localhost:8080/ws?token=<token_bob>` $\to$ Bấm **Connect**.

##### 👉 Gửi tin nhắn từ Alex sang Bob:
Tại cửa sổ của **Alex**, gửi gói tin JSON:
```json
{
  "event": "chat:send",
  "payload": {
    "receiver_id": "<id_bob>",
    "content": "Chào Bob! Tin nhắn thời gian thực qua Go + Redis!",
    "type": "text"
  }
}
```

##### 👉 Kết quả mong đợi:
1. **Bên Alex:** Nhận ngay gói tin `chat:ack` (xác nhận tin đã được gửi và lưu MongoDB).
2. **Bên Bob:** Nhận ngay gói tin `chat:receive` với nội dung tin nhắn của Alex trong chớp mắt!

---

<a name="cac-loi-thuong-gap-va-cach-xu-ly"></a>
## Các lỗi thường gặp và cách xử lý

1. **Lỗi `websocket: request origin not allowed`:**
   - *Nguyên nhân:* Bị chặn CORS do origin của trình duyệt khác với backend.
   - *Cách xử lý:* Đã được cấu hình `CheckOrigin: func(r *http.Request) bool { return true }` trong `ws_handler.go`.
2. **Lỗi `invalid character in string` khi gửi WS JSON:**
   - *Cách xử lý:* Đảm bảo gói tin gửi lên có định dạng chuẩn: `{"event":"chat:send", "payload":{...}}`.

---

### 🎉 Chúc mừng bạn!
Sau khi hoàn thành Giai đoạn 3, hệ thống Backend của bạn đã có khả năng xử lý **hàng nghìn kết nối WebSocket đồng thời**, truyền tin nhắn mượt mà qua Redis Pub/Sub và lưu trữ vĩnh viễn trong MongoDB!
Sẵn sàng bước tiếp sang **Giai đoạn 4: Xây dựng Giao diện Frontend Svelte**!
