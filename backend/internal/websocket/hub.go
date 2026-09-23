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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Hub struct {
	clients     map[string]*Client // Map: UserID -> *Client
	register    chan *Client
	Unregister  chan *Client
	mutex       sync.RWMutex
	redisClient *redis.Client
	msgRepo     *repository.MessageRepository
	convRepo    *repository.ConversationRepository
	groupRepo   *repository.GroupRepository
	userRepo    *repository.UserRepository
	draftRepo   *repository.DraftRepository
}

func NewHub(redisClient *redis.Client, msgRepo *repository.MessageRepository, convRepo *repository.ConversationRepository, groupRepo *repository.GroupRepository, userRepo *repository.UserRepository, draftRepo *repository.DraftRepository) *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		register:    make(chan *Client),
		Unregister:  make(chan *Client),
		redisClient: redisClient,
		msgRepo:     msgRepo,
		convRepo:    convRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		draftRepo:   draftRepo,
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

			// Lắng nghe các kênh nhóm của user
			go h.subscribeUserGroups(client)

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

// subscribeUserGroups đăng ký lắng nghe các kênh nhóm mà client tham gia
func (h *Hub) subscribeUserGroups(client *Client) {
	if h.groupRepo == nil {
		return
	}
	ctx := context.Background()
	userOID, err := primitive.ObjectIDFromHex(client.UserID)
	if err != nil {
		return
	}
	groups, err := h.groupRepo.GetUserGroups(ctx, userOID)
	if err != nil {
		return
	}
	for _, g := range groups {
		go h.subscribeGroupChannel(g.ID.Hex(), client.UserID)
	}
}

// subscribeGroupChannel lắng nghe Redis Channel của một nhóm
func (h *Hub) subscribeGroupChannel(groupID, userID string) {
	channelName := fmt.Sprintf("group:chat:%s", groupID)
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

	case "group:send":
		h.handleSendGroupMessage(client, event.Payload)


	case "chat:react":
		h.handleReact(client, event.Payload)

	case "chat:delete":
		h.handleDeleteMessage(client, event.Payload)

	case "chat:edit":
		h.handleEditMessage(client, event.Payload)

	case "heartbeat":
		// Gia hạn TTL 30 giây trong Redis
		ctx := context.Background()
		h.redisClient.Set(ctx, fmt.Sprintf("user:online:%s", client.UserID), "1", 30*time.Second)

	case "typing:start", "typing:stop":
		h.handleTyping(client, event.Event, event.Payload)

	case "chat:read":
		h.handleMarkAsRead(client, event.Payload)

	case "chat:unread":
		h.handleMarkAsUnread(client, event.Payload)

	case "chat:pin":
		h.handlePinMessage(client, event.Payload)

	case "draft:sync":
		h.handleDraftSync(client, event.Payload)

	case "call:request", "call:accept", "call:reject", "call:offer", "call:answer", "call:ice_candidate", "call:hangup":
		h.handleCallSignaling(client, event.Event, event.Payload)

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
	var convID string
	isSelfSaved := (req.ReceiverID == client.UserID)
	if isSelfSaved {
		convID = fmt.Sprintf("saved_%s", client.UserID)
	} else {
		convID = GetConversationID(client.UserID, req.ReceiverID)
	}

	// 1. Lưu tin nhắn vào MongoDB
	newMsg := &models.Message{
		ConversationID: convID,
		SenderID:       client.UserID,
		ReceiverID:     req.ReceiverID,
		Content:        req.Content,
		Type:           msgType,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		ReplyTo:        req.ReplyTo,
		IsSilent:       req.IsSilent,
		ThreadRootID:   req.ThreadRootID,
		ForwardFrom:    req.ForwardFrom,
		IsRead:         isSelfSaved,
		CreatedAt:      time.Now(),
	}

	if err := h.msgRepo.Create(ctx, newMsg); err != nil {
		log.Printf("Lỗi lưu tin nhắn vào MongoDB: %v", err)
		return
	}

	// Xóa bản nháp nếu có
	if h.draftRepo != nil {
		_ = h.draftRepo.Delete(ctx, client.UserID, convID)
	}

	// Nếu là tin nhắn phản hồi trong luồng (Thread Reply), cập nhật thread count của tin nhắn gốc
	if req.ThreadRootID != "" {
		if rootOID, err := primitive.ObjectIDFromHex(req.ThreadRootID); err == nil {
			_ = h.msgRepo.IncrementThreadCount(ctx, rootOID)
			if rootMsg, err := h.msgRepo.GetByID(ctx, rootOID); err == nil && rootMsg != nil {
				threadUpEvent := models.WSEvent{
					Event: "chat:thread_updated",
					Payload: map[string]interface{}{
						"root_id":         req.ThreadRootID,
						"thread_count":    rootMsg.ThreadCount + 1,
						"conversation_id": convID,
					},
				}
				bytes, _ := json.Marshal(threadUpEvent)
				if !isSelfSaved {
					h.redisClient.Publish(ctx, fmt.Sprintf("user:chat:%s", req.ReceiverID), string(bytes))
				}
				client.Send <- bytes
			}
		}
	}

	// 2. Cập nhật hội thoại
	_, _ = h.convRepo.GetOrCreate(ctx, convID, client.UserID, req.ReceiverID)
	lastSnippet := req.Content
	switch msgType {
	case "image":
		lastSnippet = "[Hình ảnh]"
	case "voice":
		lastSnippet = "[Tin nhắn thoại]"
	case "file":
		if req.FileName != "" {
			lastSnippet = "[Tệp] " + req.FileName
		} else {
			lastSnippet = "[Tệp đính kèm]"
		}
	}
	_ = h.convRepo.UpdateLastMessage(ctx, convID, lastSnippet, client.UserID)

	// 3. Đóng gói Event bắn qua Redis Pub/Sub cho người nhận (nếu không phải tự gửi cho mình)
	eventReceive := models.WSEvent{
		Event:   "chat:receive",
		Payload: newMsg,
	}
	eventReceiveBytes, _ := json.Marshal(eventReceive)
	if !isSelfSaved {
		receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
		h.redisClient.Publish(ctx, receiverChannel, string(eventReceiveBytes))
	}

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

	// Nếu client không truyền partner_id, tự động bóc tách từ conversation_id (format: user1_user2)
	if req.PartnerID == "" {
		parts := strings.Split(req.ConversationID, "_")
		if len(parts) == 2 {
			if parts[0] == client.UserID {
				req.PartnerID = parts[1]
			} else if parts[1] == client.UserID {
				req.PartnerID = parts[0]
			}
		}
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

// handleMarkAsUnread xử lý đánh dấu chưa đọc tin nhắn
func (h *Hub) handleMarkAsUnread(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}

	var req struct {
		ConversationID string `json:"conversation_id"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil || req.ConversationID == "" {
		return
	}

	ctx := context.Background()
	_ = h.msgRepo.MarkAsUnread(ctx, req.ConversationID, client.UserID)
}

// handleReact xử lý thả hoặc hủy biểu tượng cảm xúc
func (h *Hub) handleReact(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}
	var req struct {
		MessageID      string `json:"message_id"`
		ConversationID string `json:"conversation_id"`
		GroupID        string `json:"group_id"`
		ReceiverID     string `json:"receiver_id"`
		Emoji          string `json:"emoji"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}
	msgOID, err := primitive.ObjectIDFromHex(req.MessageID)
	if err != nil {
		return
	}

	ctx := context.Background()
	reactions, err := h.msgRepo.ToggleReaction(ctx, msgOID, client.UserID, req.Emoji)
	if err != nil {
		log.Printf("Lỗi ToggleReaction: %v", err)
		return
	}

	outEvent := models.WSEvent{
		Event: "chat:reaction_updated",
		Payload: map[string]interface{}{
			"message_id":      req.MessageID,
			"conversation_id": req.ConversationID,
			"group_id":        req.GroupID,
			"reactions":       reactions,
		},
	}
	outBytes, _ := json.Marshal(outEvent)

	// Gửi cho nhóm hoặc người nhận qua Redis
	if req.GroupID != "" {
		groupChannel := fmt.Sprintf("group:chat:%s", req.GroupID)
		h.redisClient.Publish(ctx, groupChannel, string(outBytes))
	} else if req.ReceiverID != "" {
		receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
		h.redisClient.Publish(ctx, receiverChannel, string(outBytes))
	}

	// Gửi cho người gửi
	client.Send <- outBytes
}

// handleDeleteMessage xử lý thu hồi tin nhắn
func (h *Hub) handleDeleteMessage(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}
	var req struct {
		MessageID      string `json:"message_id"`
		ConversationID string `json:"conversation_id"`
		GroupID        string `json:"group_id"`
		ReceiverID     string `json:"receiver_id"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}
	msgOID, err := primitive.ObjectIDFromHex(req.MessageID)
	if err != nil {
		return
	}

	ctx := context.Background()
	if err := h.msgRepo.DeleteMessage(ctx, msgOID, client.UserID); err != nil {
		log.Printf("Lỗi DeleteMessage: %v", err)
		return
	}

	outEvent := models.WSEvent{
		Event: "chat:message_deleted",
		Payload: map[string]interface{}{
			"message_id":      req.MessageID,
			"conversation_id": req.ConversationID,
			"group_id":        req.GroupID,
		},
	}
	outBytes, _ := json.Marshal(outEvent)

	if req.GroupID != "" {
		groupChannel := fmt.Sprintf("group:chat:%s", req.GroupID)
		h.redisClient.Publish(ctx, groupChannel, string(outBytes))
	} else if req.ReceiverID != "" {
		receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
		h.redisClient.Publish(ctx, receiverChannel, string(outBytes))
	}

	client.Send <- outBytes
}

// handleEditMessage xử lý chỉnh sửa nội dung tin nhắn
func (h *Hub) handleEditMessage(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}
	var req struct {
		MessageID      string `json:"message_id"`
		ConversationID string `json:"conversation_id"`
		GroupID        string `json:"group_id"`
		ReceiverID     string `json:"receiver_id"`
		Content        string `json:"content"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}
	cleanContent := strings.TrimSpace(req.Content)
	if cleanContent == "" {
		return
	}
	msgOID, err := primitive.ObjectIDFromHex(req.MessageID)
	if err != nil {
		return
	}

	ctx := context.Background()
	if err := h.msgRepo.EditMessage(ctx, msgOID, client.UserID, cleanContent); err != nil {
		log.Printf("Lỗi EditMessage: %v", err)
		return
	}

	outEvent := models.WSEvent{
		Event: "chat:message_edited",
		Payload: map[string]interface{}{
			"message_id":      req.MessageID,
			"conversation_id": req.ConversationID,
			"group_id":        req.GroupID,
			"content":         cleanContent,
		},
	}
	outBytes, _ := json.Marshal(outEvent)

	if req.GroupID != "" {
		groupChannel := fmt.Sprintf("group:chat:%s", req.GroupID)
		h.redisClient.Publish(ctx, groupChannel, string(outBytes))
	} else if req.ReceiverID != "" {
		receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
		h.redisClient.Publish(ctx, receiverChannel, string(outBytes))
	}

	client.Send <- outBytes
}

// handleSendGroupMessage xử lý gửi tin nhắn vào nhóm
func (h *Hub) handleSendGroupMessage(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}
	var req models.SendMessageRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}
	if strings.TrimSpace(req.Content) == "" || req.GroupID == "" {
		return
	}

	ctx := context.Background()
	groupOID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		return
	}
	userOID, err := primitive.ObjectIDFromHex(client.UserID)
	if err != nil {
		return
	}

	// Xác minh thành viên nhóm & kiểm tra quyền trong kênh
	if h.groupRepo != nil {
		isMember, err := h.groupRepo.IsMember(ctx, groupOID, userOID)
		if err != nil || !isMember {
			log.Printf("User %s không thuộc nhóm %s", client.UserID, req.GroupID)
			return
		}

		group, err := h.groupRepo.GetByID(ctx, groupOID)
		if err == nil && group != nil {
			isAdmin, _ := h.groupRepo.IsAdmin(ctx, groupOID, userOID)

			// Kiểm tra nếu thành viên đang bị cấm chat (Muted)
			if mutedUntil, isMuted := group.MutedMembers[client.UserID]; isMuted && mutedUntil.After(time.Now()) {
				errEvent := models.WSEvent{
					Event: "chat:error",
					Payload: map[string]interface{}{
						"type":        "member_muted",
						"message":     fmt.Sprintf("Bạn đang bị cấm chat trong nhóm này đến %s", mutedUntil.Format("15:04 02/01/2006")),
						"muted_until": mutedUntil,
					},
				}
				errBytes, _ := json.Marshal(errEvent)
				client.Send <- errBytes
				return
			}

			// Kiểm tra quyền gửi trong Kênh thông báo (Announcement channel)
			if req.ChannelID != "" {
				for _, ch := range group.Channels {
					if ch.ID == req.ChannelID && ch.Type == models.ChannelTypeAnnouncement {
						if !isAdmin && group.CreatorID != userOID {
							errEvent := models.WSEvent{
								Event: "chat:error",
								Payload: map[string]interface{}{
									"type":    "announcement_channel",
									"message": "Kênh thông báo chỉ dành cho quản trị viên đăng bài.",
								},
							}
							errBytes, _ := json.Marshal(errEvent)
							client.Send <- errBytes
							return
						}
						break
					}
				}
			}

			// Kiểm tra Chế độ chậm (Slow Mode)
			if group.SlowModeSeconds > 0 && !isAdmin && group.CreatorID != userOID {
				slowKey := fmt.Sprintf("slowmode:%s:%s", req.GroupID, client.UserID)
				ttl, _ := h.redisClient.TTL(ctx, slowKey).Result()
				if ttl > 0 {
					errEvent := models.WSEvent{
						Event: "chat:error",
						Payload: map[string]interface{}{
							"type":             "slow_mode",
							"message":          fmt.Sprintf("Chế độ chậm đang bật. Vui lòng chờ %d giây trước khi gửi tiếp.", int(ttl.Seconds())),
							"cooldown_seconds": int(ttl.Seconds()),
						},
					}
					errBytes, _ := json.Marshal(errEvent)
					client.Send <- errBytes
					return
				}
				h.redisClient.Set(ctx, slowKey, "1", time.Duration(group.SlowModeSeconds)*time.Second)
			}
		}
	}

	// Lấy profile người gửi
	senderName := client.Username
	senderAvatar := ""
	if h.userRepo != nil {
		sender, _ := h.userRepo.FindByID(ctx, client.UserID)
		if sender != nil {
			if sender.DisplayName != "" {
				senderName = sender.DisplayName
			}
			senderAvatar = sender.AvatarURL
		}
	}

	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	newMsg := &models.Message{
		GroupID:      req.GroupID,
		ChannelID:    req.ChannelID,
		SenderID:     client.UserID,
		SenderName:   senderName,
		SenderAvatar: senderAvatar,
		Content:      req.Content,
		Type:         msgType,
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		ReplyTo:      req.ReplyTo,
		IsSilent:     req.IsSilent,
		ThreadRootID: req.ThreadRootID,
		ForwardFrom:  req.ForwardFrom,
		IsRead:       true,
		CreatedAt:    time.Now(),
	}

	if err := h.msgRepo.Create(ctx, newMsg); err != nil {
		log.Printf("Lỗi lưu tin nhắn nhóm vào MongoDB: %v", err)
		return
	}

	// Xóa bản nháp nhóm nếu có
	if h.draftRepo != nil {
		_ = h.draftRepo.Delete(ctx, client.UserID, req.GroupID)
	}

	// Cập nhật thread count nếu là phản hồi trong luồng
	if req.ThreadRootID != "" {
		if rootOID, err := primitive.ObjectIDFromHex(req.ThreadRootID); err == nil {
			_ = h.msgRepo.IncrementThreadCount(ctx, rootOID)
			if rootMsg, err := h.msgRepo.GetByID(ctx, rootOID); err == nil && rootMsg != nil {
				threadUpEvent := models.WSEvent{
					Event: "group:thread_updated",
					Payload: map[string]interface{}{
						"root_id":      req.ThreadRootID,
						"thread_count": rootMsg.ThreadCount + 1,
						"group_id":     req.GroupID,
					},
				}
				bytes, _ := json.Marshal(threadUpEvent)
				groupChannel := fmt.Sprintf("group:chat:%s", req.GroupID)
				h.redisClient.Publish(ctx, groupChannel, string(bytes))
			}
		}
	}

	// Broadcast tin nhắn vào Redis channel của nhóm
	eventReceive := models.WSEvent{
		Event:   "group:receive",
		Payload: newMsg,
	}
	eventBytes, _ := json.Marshal(eventReceive)
	groupChannel := fmt.Sprintf("group:chat:%s", req.GroupID)
	h.redisClient.Publish(ctx, groupChannel, string(eventBytes))

	// Trả ACK về cho người gửi
	eventACK := models.WSEvent{
		Event: "group:ack",
		Payload: map[string]interface{}{
			"temp_id": newMsg.ID.Hex(),
			"message": newMsg,
		},
	}
	ackBytes, _ := json.Marshal(eventACK)
	client.Send <- ackBytes
}

// handleCallSignaling xử lý chuyển tiếp các gói tin báo hiệu WebRTC P2P (Offer, Answer, ICE, Hangup...)
func (h *Hub) handleCallSignaling(client *Client, eventType string, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return
	}

	receiverID, _ := data["receiver_id"].(string)
	if receiverID == "" {
		return
	}

	// Lấy thông tin người gửi (caller / callee)
	senderName := client.Username
	senderAvatar := ""
	if h.userRepo != nil {
		ctx := context.Background()
		sender, _ := h.userRepo.FindByID(ctx, client.UserID)
		if sender != nil {
			if sender.DisplayName != "" {
				senderName = sender.DisplayName
			}
			senderAvatar = sender.AvatarURL
		}
	}

	data["sender_id"] = client.UserID
	data["sender_name"] = senderName
	data["sender_avatar"] = senderAvatar

	outEvent := models.WSEvent{
		Event:   eventType,
		Payload: data,
	}
	outBytes, err := json.Marshal(outEvent)
	if err != nil {
		return
	}

	// Chuyển tiếp gói tin qua Redis channel của người nhận
	ctx := context.Background()
	receiverChannel := fmt.Sprintf("user:chat:%s", receiverID)
	h.redisClient.Publish(ctx, receiverChannel, string(outBytes))
}

// handlePinMessage xử lý ghim hoặc gỡ ghim tin nhắn
func (h *Hub) handlePinMessage(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}
	var req struct {
		MessageID      string `json:"message_id"`
		IsPinned       bool   `json:"is_pinned"`
		ConversationID string `json:"conversation_id"`
		GroupID        string `json:"group_id"`
		ReceiverID     string `json:"receiver_id"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}
	msgOID, err := primitive.ObjectIDFromHex(req.MessageID)
	if err != nil {
		return
	}

	ctx := context.Background()
	if req.IsPinned {
		_ = h.msgRepo.PinMessage(ctx, msgOID, client.UserID)
	} else {
		_ = h.msgRepo.UnpinMessage(ctx, msgOID)
	}

	outEvent := models.WSEvent{
		Event: "chat:pin_updated",
		Payload: map[string]interface{}{
			"message_id":      req.MessageID,
			"is_pinned":       req.IsPinned,
			"conversation_id": req.ConversationID,
			"group_id":        req.GroupID,
			"pinned_by":       client.UserID,
		},
	}
	outBytes, _ := json.Marshal(outEvent)

	if req.GroupID != "" {
		groupChannel := fmt.Sprintf("group:chat:%s", req.GroupID)
		h.redisClient.Publish(ctx, groupChannel, string(outBytes))
	} else if req.ReceiverID != "" {
		receiverChannel := fmt.Sprintf("user:chat:%s", req.ReceiverID)
		h.redisClient.Publish(ctx, receiverChannel, string(outBytes))
	}

	client.Send <- outBytes
}

// handleDraftSync xử lý đồng bộ tin nhắn nháp đa thiết bị
func (h *Hub) handleDraftSync(client *Client, rawPayload interface{}) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return
	}
	var req struct {
		TargetID   string `json:"target_id"`
		TargetType string `json:"target_type"`
		Content    string `json:"content"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}

	if h.draftRepo != nil {
		ctx := context.Background()
		_ = h.draftRepo.Upsert(ctx, client.UserID, req.TargetID, req.TargetType, req.Content)
	}
}




