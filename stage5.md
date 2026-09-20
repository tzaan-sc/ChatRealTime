# Hướng Dẫn Chi Tiết Giai Đoạn 5: Tính Năng Realtime Nâng Cao (Presence, Typing, Read Receipts, Âm Thanh)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: Golang + Redis (TTL & Pub/Sub) + MongoDB + Svelte 5

Tài liệu này sẽ hướng dẫn bạn nâng tầm trải nghiệm ứng dụng chat từ mức cơ bản lên chuẩn trải nghiệm hiện đại như **Telegram / Messenger** với 4 tính năng:
1. 🟢 **Trạng thái Online / Offline (Presence):** Quản lý qua Redis Key TTL 30s + Heartbeat định kỳ + broadcast tức thời khi đóng/mở tab.
2. ✍️ **Đang gõ phím (Typing Indicator):** Bắn tín hiệu qua Redis Pub/Sub, hiệu ứng animation 3 dấu chấm nhảy mềm mại.
3. 👁️ **Đã đọc tin nhắn (Read Receipts):** Đánh dấu 2 tích xanh khi bạn chat mở xem tin nhắn.
4. 🔔 **Âm thanh thông báo (Notification Sound):** Phát âm thanh Pop nhẹ bằng Web Audio API khi có tin nhắn mới (không phụ thuộc file ngoài).

---

## 📋 Checklist Tiến Độ Giai Đoạn 5

- [x] **1. Cập nhật Backend Go (`internal/websocket/hub.go`):**
  - [x] Xử lý sự kiện `heartbeat` & lưu Redis key `user:online:{userId}` với TTL 30s.
  - [x] Tự động broadcast `user:status` (online = true/false) khi client kết nối & ngắt kết nối.
  - [x] Chuyển tiếp sự kiện `typing:start` và `typing:stop` qua Redis Pub/Sub tới đối phương.
  - [x] Xử lý sự kiện `chat:read`: Đánh dấu đã đọc trong MongoDB và bắn `chat:read_ack` về cho người gửi.
- [x] **2. Tạo tiện ích âm thanh thông báo (`frontend/src/lib/utils/sound.js`):**
  - [x] Viết hàm `playNotificationSound()` bằng Web Audio API thuần.
- [x] **3. Nâng cấp Store Chat (`frontend/src/lib/stores/chat.js`):**
  - [x] Thêm store `onlineUsers` lưu danh sách ID người đang online.
  - [x] Thêm store `typingUsers` lưu trạng thái ai đang gõ tin nhắn.
  - [x] Thêm hàm đánh dấu tin nhắn đã đọc `markMessagesAsReadLocally()`.
- [x] **4. Nâng cấp `ChatArea.svelte`:**
  - [x] Bắt sự kiện gõ phím trong `<textarea>`, gửi `typing:start` và debounce 2s gửi `typing:stop`.
  - [x] Hiển thị animation 3 dấu chấm nhảy sinh động khi đối phương đang gõ.
  - [x] Hiển thị 1 dấu tích xám (đã gửi) và 2 dấu tích xanh (đã xem).
  - [x] Tự động gửi `chat:read` khi mở xem cuộc trò chuyện.
- [ ] **5. Nâng cấp `Sidebar.svelte`:**
  - [ ] Hiển thị chấm tròn màu xanh lá khi bạn chat đang online.
- [ ] **6. Nâng cấp `App.svelte`:**
  - [ ] Tự động gửi `heartbeat` định kỳ mỗi 15 giây.
  - [ ] Lắng nghe toàn bộ sự kiện nâng cao (`user:status`, `typing:start`, `typing:stop`, `chat:read_ack`).
  - [ ] Phát âm thanh thông báo khi có tin nhắn đến từ người khác.
- [ ] **7. Kiểm thử thực tế kịch bản 2 người dùng (Alex & Bob)**.

---

## Mục lục
1. [Bước 1: Nâng cấp Backend Go (`internal/websocket/hub.go`)](#buoc-1-backend-go)
2. [Bước 2: Tạo Utility Âm thanh Web Audio API (`sound.js`)](#buoc-2-utility-am-thanh)
3. [Bước 3: Nâng cấp Chat Stores (`stores/chat.js`)](#buoc-3-chat-stores)
4. [Bước 4: Nâng cấp Khung Chat Chính (`ChatArea.svelte`)](#buoc-4-chat-area)
5. [Bước 5: Nâng cấp Sidebar Danh bạ (`Sidebar.svelte`)](#buoc-5-sidebar)
6. [Bước 6: Nâng cấp Lắp ráp Ứng dụng (`App.svelte`)](#buoc-6-app-svelte)
7. [Bước 7: Kiểm thử thực tế chi tiết từng tính năng](#buoc-7-kiem-thu-thuc-te)

---

<a name="buoc-1-backend-go"></a>
## Bước 1: Nâng cấp Backend Go (`internal/websocket/hub.go`)

Chúng ta bổ sung vào `Hub` các khả năng:
1. Khi User connect $\to$ Set Redis TTL `user:online:{id}` và broadcast `user:status` (online = true).
2. Khi User disconnect $\to$ Xoá Redis key và broadcast `user:status` (online = false).
3. Lắng nghe thêm kênh Redis toàn cục `global:events` để mọi người dùng đều nhận được thông tin Online/Offline của nhau.
4. Xử lý các event: `heartbeat`, `typing:start`, `typing:stop`, `chat:read`.

Mở file [backend/internal/websocket/hub.go](file:///d:/GIT/ChatRealTime/backend/internal/websocket/hub.go) và cập nhật toàn bộ nội dung như sau:

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

// Broadcast trạng thái online/offline cho toàn hệ thống
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

// Gửi danh sách những ai đang online khi client vừa kết nối
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
		Event: "user:online_list",
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

// Xử lý gửi tin nhắn
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

// Xử lý Typing Indicator (chuyển tiếp tức thời, không lưu DB)
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

// Xử lý đánh dấu đã đọc tin nhắn
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
```

---

<a name="buoc-2-utility-am-thanh"></a>
## Bước 2: Tạo Utility Âm thanh Web Audio API (`sound.js`)

Tạo file [frontend/src/lib/utils/sound.js](file:///d:/GIT/ChatRealTime/frontend/src/lib/utils/sound.js) để tạo âm thanh nhẹ nhàng chuẩn web:

```javascript
// Phát âm thanh thông báo "Pop" nhẹ nhàng khi có tin nhắn đến
export function playNotificationSound() {
  try {
    const AudioCtx = window.AudioContext || window.webkitAudioContext;
    if (!AudioCtx) return;

    const ctx = new AudioCtx();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = 'sine';
    // Lướt từ nốt D5 (587Hz) lên A5 (880Hz) tạo cảm giác tin nhắn đến êm dịu
    osc.frequency.setValueAtTime(587.33, ctx.currentTime);
    osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.08);

    // Âm lượng fade out mềm mại trong 120ms
    gain.gain.setValueAtTime(0.12, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.12);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.12);
  } catch (e) {
    // Trình duyệt chưa cho phép autoplay nếu chưa click
    console.warn('Âm thanh thông báo chưa sẵn sàng:', e);
  }
}
```

---

<a name="buoc-3-chat-stores"></a>
## Bước 3: Nâng cấp Chat Stores (`stores/chat.js`)

Bổ sung các store theo dõi trạng thái `onlineUsers`, `typingUsers`, và hàm cập nhật trạng thái `is_read = true` cục bộ.

Cập nhật file [frontend/src/lib/stores/chat.js](file:///d:/GIT/ChatRealTime/frontend/src/lib/stores/chat.js):

```javascript
import { writable, get } from 'svelte/store';
import { apiRequest } from '../services/api';
import { token } from './auth';

export const conversations = writable([]);
export const activeConversation = writable(null); // { conversation, other_user }
export const currentMessages = writable([]);
export const userDirectory = writable([]);

// Set các User ID đang online (Ví dụ: Set(['id1', 'id2']))
export const onlineUsers = writable(new Set());

// Set các User ID đang gõ phím vào khung chat
export const typingUsers = writable(new Set());

// Tải danh sách hội thoại
export async function loadConversations() {
  const t = get(token);
  if (!t) return;
  try {
    const res = await apiRequest('/chat/conversations', 'GET', null, t);
    conversations.set(res.data || []);
  } catch (err) {
    console.error('Lỗi tải danh sách hội thoại:', err);
  }
}

// Chọn cuộc trò chuyện và tải lịch sử
export async function selectConversation(convItem) {
  activeConversation.set(convItem);
  const t = get(token);
  const convID = convItem.conversation.custom_id;
  try {
    const res = await apiRequest(`/chat/messages/${convID}?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    // Đánh dấu đã xem trên server
    apiRequest(`/chat/messages/${convID}/read`, 'POST', null, t);
  } catch (err) {
    console.error('Lỗi tải tin nhắn:', err);
  }
}

// Thêm tin nhắn mới vào danh sách hiện tại
export function appendMessage(msg) {
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === msg.conversation_id) {
    currentMessages.update((msgs) => [...msgs, msg]);
  }
  loadConversations();
}

// Cập nhật trạng thái tin nhắn đã đọc khi nhận event chat:read_ack
export function markMessagesAsReadLocally(convID) {
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === convID) {
    currentMessages.update((msgs) =>
      msgs.map((m) => ({ ...m, is_read: true }))
    );
  }
}

// Cập nhật trạng thái Online / Offline
export function setUserOnlineStatus(userID, isOnline) {
  onlineUsers.update((set) => {
    const next = new Set(set);
    if (isOnline) {
      next.add(userID);
    } else {
      next.delete(userID);
    }
    return next;
  });
}

// Cập nhật trạng thái Đang gõ phím
export function setUserTyping(userID, isTyping) {
  typingUsers.update((set) => {
    const next = new Set(set);
    if (isTyping) {
      next.add(userID);
    } else {
      next.delete(userID);
    }
    return next;
  });
}
```

---

<a name="buoc-4-chat-area"></a>
## Bước 4: Nâng cấp Khung Chat Chính (`ChatArea.svelte`)

Tại [frontend/src/lib/components/ChatArea.svelte](file:///d:/GIT/ChatRealTime/frontend/src/lib/components/ChatArea.svelte):
1. Thêm xử lý `on:input={handleInput}`: phát hiện gõ phím $\to$ gửi `typing:start`, sau 2 giây không gõ $\to$ gửi `typing:stop`.
2. Hiển thị trạng thái Online của người đối diện ngay tại Header.
3. Hiển thị Animation 3 dấu chấm nhảy mềm mại khi đối phương đang gõ.
4. Hiển thị 1 tích xám `✓` (đã gửi) hoặc 2 tích xanh `✓✓` (đã xem) cạnh thời gian tin nhắn.
5. Khi mở chat hoặc nhận tin mới, tự động gửi WebSocket event `chat:read` để báo cho đối phương.

Cập nhật file [ChatArea.svelte](file:///d:/GIT/ChatRealTime/frontend/src/lib/components/ChatArea.svelte):

```svelte
<script>
  import { activeConversation, currentMessages, onlineUsers, typingUsers } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';
  import { afterUpdate } from 'svelte';
  import { Send } from 'lucide-svelte';

  let inputContent = '';
  let messagesContainer;
  let typingTimeout;
  let isTyping = false;

  // Kiểm tra đối phương có online không
  $: partnerID = $activeConversation?.other_user?.id;
  $: isPartnerOnline = partnerID ? $onlineUsers.has(partnerID) : false;
  $: isPartnerTyping = partnerID ? $typingUsers.has(partnerID) : false;

  // Tự động cuộn xuống đáy khi có tin nhắn mới
  afterUpdate(() => {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  });

  // Bắt sự kiện gõ phím để bắn Typing Indicator
  function handleInput() {
    if (!$activeConversation) return;

    if (!isTyping) {
      isTyping = true;
      wsService.send('typing:start', { receiver_id: partnerID });
    }

    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
      isTyping = false;
      wsService.send('typing:stop', { receiver_id: partnerID });
    }, 2000);
  }

  // Gửi tin nhắn
  function handleSend() {
    if (!inputContent.trim() || !$activeConversation) return;

    // Dừng typing khi gửi tin
    clearTimeout(typingTimeout);
    if (isTyping) {
      isTyping = false;
      wsService.send('typing:stop', { receiver_id: partnerID });
    }

    wsService.send('chat:send', {
      receiver_id: partnerID,
      content: inputContent.trim(),
      type: 'text'
    });

    inputContent = '';
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  // Gửi event xác nhận đã đọc khi đang mở xem khung chat
  $: if ($activeConversation && partnerID) {
    wsService.send('chat:read', {
      conversation_id: $activeConversation.conversation.custom_id,
      partner_id: partnerID
    });
  }
</script>

{#if !$activeConversation}
  <div class="empty-chat glass-card">
    <div class="empty-content">
      <div class="icon-circle">💬</div>
      <h2>Chào mừng đến với Realtime Chat!</h2>
      <p>Chọn một người bạn ở danh sách bên trái để bắt đầu cuộc trò chuyện.</p>
    </div>
  </div>
{:else}
  <main class="chat-area">
    <!-- Header -->
    <div class="chat-header glass-card">
      <div class="partner-info">
        <div class="avatar-wrap">
          <img src={$activeConversation.other_user.avatar_url} alt="avatar" class="header-avatar" />
          <span class="status-dot" class:online={isPartnerOnline}></span>
        </div>
        <div>
          <h3>{$activeConversation.other_user.display_name || $activeConversation.other_user.username}</h3>
          <span class="status-sub">
            {#if isPartnerTyping}
              <span class="typing-text">đang soạn tin...</span>
            {:else if isPartnerOnline}
              <span class="online-text">Đang hoạt động</span>
            {:else}
              <span class="offline-text">Ngoại tuyến</span>
            {/if}
          </span>
        </div>
      </div>
    </div>

    <!-- Messages Body -->
    <div class="messages-viewport" bind:this={messagesContainer}>
      {#each $currentMessages as msg}
        {@const isMe = msg.sender_id === $currentUser.id}
        <div class="message-row" class:me={isMe}>
          <div class="bubble" class:bubble-me={isMe} class:bubble-other={!isMe}>
            <p class="text">{msg.content}</p>
            <div class="msg-meta">
              <span class="timestamp">
                {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
              {#if isMe}
                <span class="receipt-icon" class:seen={msg.is_read} title={msg.is_read ? 'Đã xem' : 'Đã gửi'}>
                  {msg.is_read ? '✓✓' : '✓'}
                </span>
              {/if}
            </div>
          </div>
        </div>
      {/each}

      <!-- Animation Đang Gõ Phím (Typing Dots) -->
      {#if isPartnerTyping}
        <div class="message-row">
          <div class="typing-bubble glass-card">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Input Footer -->
    <div class="chat-footer glass-card">
      <div class="input-wrapper">
        <textarea
          rows="1"
          placeholder="Nhập tin nhắn... (Nhấn Enter để gửi)"
          bind:value={inputContent}
          on:input={handleInput}
          on:keydown={handleKeyDown}
        ></textarea>
        <button class="send-btn" on:click={handleSend} disabled={!inputContent.trim()}>
          <Send size={18} />
        </button>
      </div>
    </div>
  </main>
{/if}

<style>
  .chat-area { flex: 1; height: 100vh; display: flex; flex-direction: column; background: var(--bg-primary); }
  .empty-chat { flex: 1; display: flex; align-items: center; justify-content: center; text-align: center; }
  .icon-circle { font-size: 50px; margin-bottom: 16px; }
  .empty-content h2 { font-size: 22px; margin-bottom: 8px; }
  .empty-content p { color: var(--text-muted); font-size: 14px; }
  .chat-header {
    padding: 16px 24px;
    border-bottom: 1px solid var(--border-glass);
    display: flex;
    align-items: center;
  }
  .partner-info { display: flex; align-items: center; gap: 14px; }
  .avatar-wrap { position: relative; }
  .header-avatar { width: 44px; height: 44px; border-radius: 50%; }
  .status-dot {
    position: absolute;
    bottom: 2px;
    right: 2px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--offline-color);
    border: 2px solid var(--bg-primary);
  }
  .status-dot.online { background: var(--online-color); }
  .partner-info h3 { font-size: 16px; font-weight: 600; color: #fff; }
  .status-sub { font-size: 12px; }
  .online-text { color: var(--online-color); }
  .offline-text { color: var(--text-dim); }
  .typing-text { color: #a78bfa; font-style: italic; animation: pulse 1.5s infinite; }
  @keyframes pulse { 0%, 100% { opacity: 0.6; } 50% { opacity: 1; } }

  .messages-viewport {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .message-row { display: flex; width: 100%; }
  .message-row.me { justify-content: flex-end; }
  .bubble {
    max-width: 65%;
    padding: 12px 16px;
    border-radius: var(--radius-md);
    position: relative;
    word-break: break-word;
  }
  .bubble-me {
    background: var(--bubble-me);
    color: #fff;
    border-bottom-right-radius: 4px;
  }
  .bubble-other {
    background: var(--bubble-other);
    color: #f3f4f6;
    border-bottom-left-radius: 4px;
    border: 1px solid var(--border-glass);
  }
  .text { font-size: 14px; line-height: 1.5; }
  .msg-meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
  }
  .timestamp { font-size: 10px; opacity: 0.75; }
  .receipt-icon { font-size: 11px; font-weight: 700; opacity: 0.6; }
  .receipt-icon.seen { color: #38bdf8; opacity: 1; }

  /* Typing Indicator Animation */
  .typing-bubble {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 10px 16px;
    border-radius: var(--radius-md);
    border-bottom-left-radius: 4px;
  }
  .typing-dot {
    width: 6px;
    height: 6px;
    background: #a78bfa;
    border-radius: 50%;
    animation: typingBounce 1.4s infinite ease-in-out both;
  }
  .typing-dot:nth-child(1) { animation-delay: -0.32s; }
  .typing-dot:nth-child(2) { animation-delay: -0.16s; }
  @keyframes typingBounce {
    0%, 80%, 100% { transform: scale(0); }
    40% { transform: scale(1); }
  }

  .chat-footer { padding: 16px 24px; border-top: 1px solid var(--border-glass); }
  .input-wrapper {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 8px 16px;
  }
  .input-wrapper textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 14px;
    outline: none;
    resize: none;
    font-family: inherit;
    max-height: 100px;
  }
  .send-btn {
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    width: 38px;
    height: 38px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .send-btn:not(:disabled):hover { transform: scale(1.05); }
</style>
```

---

<a name="buoc-5-sidebar"></a>
## Bước 5: Nâng cấp Sidebar Danh bạ (`Sidebar.svelte`)

Thêm chấm xanh báo Online ngay cạnh Avatar của từng bạn chat trong danh sách hội thoại.

Tại [frontend/src/lib/components/Sidebar.svelte](file:///d:/GIT/ChatRealTime/frontend/src/lib/components/Sidebar.svelte):
1. Import `onlineUsers` từ `../stores/chat`.
2. Kiểm tra `isOnline = $onlineUsers.has(item.other_user.id)` và render thẻ `<span class="user-status-dot" class:online={isOnline}></span>`.

Cập nhật phần hiển thị avatar trong `Sidebar.svelte`:

```svelte
          <div class="avatar-container">
            <img src={item.other_user.avatar_url} alt="avatar" class="avatar" />
            <span class="user-status-dot" class:online={$onlineUsers.has(item.other_user.id)}></span>
          </div>
```

Và thêm CSS:
```css
  .avatar-container { position: relative; }
  .user-status-dot {
    position: absolute;
    bottom: 2px;
    right: 2px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--offline-color);
    border: 2px solid var(--bg-secondary);
  }
  .user-status-dot.online { background: var(--online-color); }
```

---

<a name="buoc-6-app-svelte"></a>
## Bước 6: Nâng cấp Lắp ráp Ứng dụng (`App.svelte`)

Tại [frontend/src/App.svelte](file:///d:/GIT/ChatRealTime/frontend/src/App.svelte):
1. Gửi `heartbeat` lên WebSocket mỗi 15 giây.
2. Lắng nghe toàn bộ sự kiện nâng cao:
   - `user:online_list`: Nhận danh sách người đang online ban đầu.
   - `user:status`: Cập nhật trạng thái người khác vừa Online hoặc vừa Offline.
   - `typing:start` / `typing:stop`: Cập nhật trạng thái đang gõ phím.
   - `chat:read_ack`: Đổi dấu tích xám thành 2 tích xanh đã xem.
   - `chat:receive`: Nhận tin nhắn mới và **phát âm thanh thông báo pop** qua `playNotificationSound()`.

Cập nhật file [App.svelte](file:///d:/GIT/ChatRealTime/frontend/src/App.svelte):

```svelte
<script>
  import { onDestroy } from 'svelte';
  import { token, currentUser } from './lib/stores/auth';
  import { wsService, incomingMessages } from './lib/services/websocket';
  import {
    loadConversations,
    appendMessage,
    onlineUsers,
    setUserOnlineStatus,
    setUserTyping,
    markMessagesAsReadLocally
  } from './lib/stores/chat';
  import { playNotificationSound } from './lib/utils/sound';
  import AuthModal from './lib/components/AuthModal.svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import ChatArea from './lib/components/ChatArea.svelte';

  let heartbeatInterval;

  // 1. Tự động kết nối WebSocket và thiết lập Heartbeat định kỳ
  $: if ($token) {
    wsService.connect($token);
    loadConversations();

    clearInterval(heartbeatInterval);
    heartbeatInterval = setInterval(() => {
      wsService.send('heartbeat', {});
    }, 15000); // Mỗi 15 giây gửi 1 lần
  }

  onDestroy(() => {
    clearInterval(heartbeatInterval);
  });

  // 2. Lắng nghe và điều phối các sự kiện Real-time
  $: if ($incomingMessages) {
    const { event, payload } = $incomingMessages;

    switch (event) {
      case 'chat:receive':
        appendMessage(payload);
        // Phát âm thanh thông báo nếu tin nhắn đến từ người khác
        if (payload.sender_id !== $currentUser?.id) {
          playNotificationSound();
        }
        break;

      case 'chat:ack':
        appendMessage(payload.message);
        break;

      case 'user:online_list':
        // Danh sách ID đang online khi vừa vào app
        if (Array.isArray(payload)) {
          onlineUsers.set(new Set(payload));
        }
        break;

      case 'user:status':
        // Cập nhật người vừa online / offline
        if (payload?.user_id) {
          setUserOnlineStatus(payload.user_id, payload.is_online);
        }
        break;

      case 'typing:start':
        if (payload?.sender_id) {
          setUserTyping(payload.sender_id, true);
        }
        break;

      case 'typing:stop':
        if (payload?.sender_id) {
          setUserTyping(payload.sender_id, false);
        }
        break;

      case 'chat:read_ack':
        // Đánh dấu đã đọc thành 2 tích xanh
        if (payload?.conversation_id) {
          markMessagesAsReadLocally(payload.conversation_id);
        }
        break;
    }
  }
</script>

<div class="app-layout">
  {#if !$currentUser}
    <AuthModal />
  {:else}
    <Sidebar />
    <ChatArea />
  {/if}
</div>

<style>
  .app-layout {
    display: flex;
    width: 100vw;
    height: 100vh;
    background: var(--bg-primary);
  }
</style>
```

---

<a name="buoc-7-kiem-thu-thuc-te"></a>
## Bước 7: Kiểm thử thực tế chi tiết từng tính năng

Sau khi hoàn thành cập nhật code, mở 2 cửa sổ trình duyệt (Alex và Bob) để kiểm thử toàn diện:

### 🧪 Test 1: Kiểm thử Trạng thái Online / Offline
1. Mở đồng thời cửa sổ của Alex và Bob.
2. 👉 **Kết quả:** Tại Sidebar và Header chat của cả hai đều có **chấm tròn màu xanh lá** và dòng chữ `"Đang hoạt động"`.
3. Đóng hẳn tab của Bob $\to$ Quan sát màn hình Alex: Sau khi Bob ngắt kết nối, chấm tròn của Bob đổi sang màu xám và hiện `"Ngoại tuyến"`.

### 🧪 Test 2: Kiểm thử Animation Đang gõ phím (Typing Indicator)
1. Tại cửa sổ của Alex, nhấp chuột vào ô nhập tin nhắn và gõ vài chữ (chưa bấm Enter).
2. 👉 **Kết quả:** Tại cửa sổ của Bob, Header lập tức hiện dòng chữ tím *"đang soạn tin..."* và dưới đáy khung chat xuất hiện **bong bóng 3 dấu chấm nhảy mềm mại**!
3. Alex dừng gõ trong 2 giây $\to$ Bong bóng 3 dấu chấm bên Bob tự động biến mất!

### 🧪 Test 3: Kiểm thử Đã đọc (Read Receipts)
1. Alex gửi 1 tin nhắn cho Bob:
   - Bên Alex: Tin nhắn vừa gửi hiện **1 dấu tích xám `✓`** (Đã gửi tới server).
2. Bob đang mở cuộc trò chuyện với Alex xem tin:
   - 👉 **Kết quả bên Alex:** Dấu tích xám lập tức biến thành **2 dấu tích màu xanh biển `✓✓`** (Đã xem)!

### 🧪 Test 4: Kiểm thử Âm thanh thông báo (Sound Effect)
1. Để tab của Alex ở chế độ nền (hoặc chuyển sang tab khác).
2. Tại tab của Bob, gửi 1 tin nhắn sang cho Alex.
3. 👉 **Kết quả:** Tai nghe/loa máy tính của bạn sẽ phát ra tiếng **"Pop"** nhẹ nhàng, êm ái báo hiệu tin nhắn mới!

---

### 🎉 Xin chúc mừng!
Bạn đã sở hữu một hệ thống Chat Real-time đỉnh cao với đầy đủ tính năng tiêu chuẩn của các ứng dụng chat hàng đầu thế giới!
Sẵn sàng bước tiếp sang **Giai đoạn 6: Docker hóa và Triển khai Production**!
