# Hướng Dẫn Chi Tiết Giai Đoạn 5: Tính Năng Realtime Nâng Cao (Presence, Typing, Read Receipts)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: Golang + Redis (TTL & Pub/Sub) + MongoDB + Svelte

Tài liệu này sẽ hướng dẫn bạn nâng tầm trải nghiệm của ứng dụng chat từ cơ bản lên chuẩn **Messenger / Telegram** với 4 tính năng nâng cao:
1. **Trạng thái Online / Offline (Presence):** Sử dụng cơ chế Redis Key TTL (Time To Live).
2. **Đang gõ phím (Typing Indicator):** Bắn tín hiệu qua Redis Pub/Sub, hiệu ứng 3 chấm nhảy sống động trên UI.
3. **Đã đọc tin nhắn (Read Receipts):** Đánh dấu 2 tích xanh khi đối phương mở xem khung chat.
4. **Âm thanh thông báo:** Phát âm thanh nhẹ nhàng bằng Web Audio API khi nhận tin nhắn mới.

---

## 📋 Checklist Tiến Độ Giai Đoạn 5

- [ ] **1. Triển khai Trạng thái Online / Offline (Presence):**
  - Backend: Ghi nhận Heartbeat định kỳ $\to$ Lưu Redis key `user:online:{userId}` với TTL 30s.
  - Broadcast event `user:status` (online/offline) qua Redis cho bạn chat.
  - Frontend: Hiển thị chấm tròn xanh khi online hoặc "Đang hoạt động".
- [ ] **2. Triển khai Tính năng Typing Indicator ("Đang soạn tin..."):**
  - Frontend: Gõ phím gửi event `typing:start`, dừng gõ 2s gửi event `typing:stop`.
  - Backend: Chuyển tiếp event qua Redis Pub/Sub tới đối phương (không ghi MongoDB).
  - Frontend: Hiển thị animation 3 dấu chấm nhảy mềm mại.
- [ ] **3. Triển khai Đã xem tin nhắn (Read Receipts):**
  - Client mở khung chat $\to$ Bắn event `chat:read`.
  - Backend cập nhật MongoDB `is_read = true` và bắn event `chat:read_ack` về cho người gửi.
  - Frontend: Cập nhật icon 2 dấu tích xanh hoặc avatar nhỏ cạnh tin nhắn.
- [ ] **4. Thêm âm thanh thông báo:**
  - Tích hợp hàm phát tiếng beep/pop nhẹ khi có tin nhắn đến (nếu cửa sổ chat không focus).

---

## Mục lục
1. [Bước 1: Cơ chế Presence (Online/Offline) bằng Redis Key TTL](#buoc-1-co-che-presence)
2. [Bước 2: Triển khai Typing Indicator ("Đang gõ...")](#buoc-2-trien-khai-typing-indicator)
3. [Bước 3: Triển khai Xác nhận Đã đọc (Read Receipts)](#buoc-3-trien-khai-read-receipts)
4. [Bước 4: Tích hợp Âm thanh Thông báo trong Svelte](#buoc-4-tich-hop-am-thanh)
5. [Bước 5: Kiểm thử thực tế các tính năng nâng cao](#buoc-5-kiem-thu-thuc-te)

---

<a name="buoc-1-co-che-presence"></a>
## Bước 1: Cơ chế Presence (Online/Offline) bằng Redis Key TTL

### 1.1. Nguyên lý hoạt động
- Mỗi 15 giây, Frontend Svelte gửi một gói tin `{ "event": "heartbeat" }` lên Backend Go.
- Backend Go nhận được sẽ gọi lệnh Redis:
  ```go
  redisClient.Set(ctx, "user:online:"+userID, "1", 30*time.Second)
  ```
- Nếu sau 30 giây người dùng bị rớt mạng, tắt trình duyệt $\to$ Key trong Redis tự động biến mất $\to$ Hệ thống nhận diện User đã **Offline**.

### 1.2. Bổ sung vào Backend `internal/websocket/hub.go`
```go
// Xử lý Heartbeat trong HandleClientEvent
case "heartbeat":
    ctx := context.Background()
    // Lưu key với thời gian sống 30 giây
    h.redisClient.Set(ctx, fmt.Sprintf("user:online:%s", client.UserID), "1", 30*time.Second)
    
    // Broadcast trạng thái online
    h.broadcastUserStatus(client.UserID, true)
```

```go
// Hàm broadcast trạng thái Online / Offline
func (h *Hub) broadcastUserStatus(userID string, isOnline bool) {
    event := models.WSEvent{
        Event: "user:status",
        Payload: map[string]interface{}{
            "user_id":   userID,
            "is_online": isOnline,
        },
    }
    eventBytes, _ := json.Marshal(event)
    // Publish vào channel global
    h.redisClient.Publish(context.Background(), "user:status:global", string(eventBytes))
}
```

---

<a name="buoc-2-trien-khai-typing-indicator"></a>
## Bước 2: Triển khai Typing Indicator ("Đang gõ...")

### 2.1. Phía Backend Go (`internal/websocket/hub.go`)
Khi nhận sự kiện `typing:start` hoặc `typing:stop`, Server chỉ chuyển tiếp (forward) qua Redis Pub/Sub tới người nhận mà **không lưu vào MongoDB** để tiết kiệm tài nguyên:

```go
case "typing:start", "typing:stop":
    payloadBytes, _ := json.Marshal(event.Payload)
    var req struct {
        ReceiverID string `json:"receiver_id"`
    }
    json.Unmarshal(payloadBytes, &req)

    forwardEvent := models.WSEvent{
        Event: event.Event,
        Payload: map[string]string{
            "sender_id": client.UserID,
        },
    }
    forwardBytes, _ := json.Marshal(forwardEvent)
    h.redisClient.Publish(context.Background(), fmt.Sprintf("user:chat:%s", req.ReceiverID), string(forwardBytes))
```

### 2.2. Phía Frontend Svelte (`ChatArea.svelte`)
1. **Gửi sự kiện khi người dùng gõ vào ô nhập:**
```javascript
let typingTimeout;

function handleInput() {
  if (!isTyping) {
    isTyping = true;
    wsService.send('typing:start', { receiver_id: $activeConversation.other_user.id });
  }

  clearTimeout(typingTimeout);
  typingTimeout = setTimeout(() => {
    isTyping = false;
    wsService.send('typing:stop', { receiver_id: $activeConversation.other_user.id });
  }, 2000); // 2 giây không gõ sẽ tự tắt
}
```

2. **Hiển thị Animation 3 dấu chấm nhảy (CSS):**
```html
{#if isPartnerTyping}
  <div class="typing-indicator glass-card">
    <div class="dot"></div>
    <div class="dot"></div>
    <div class="dot"></div>
  </div>
{/if}
```

```css
.typing-indicator {
  display: inline-flex;
  gap: 4px;
  padding: 8px 12px;
  border-radius: 12px;
  width: fit-content;
}
.dot {
  width: 6px;
  height: 6px;
  background: var(--text-muted);
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}
.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}
```

---

<a name="buoc-3-trien-khai-read-receipts"></a>
## Bước 3: Triển khai Xác nhận Đã đọc (Read Receipts)

### 3.1. Phía Backend Go
Khi người nhận mở cửa sổ chat:
1. Cập nhật `is_read = true` trong MongoDB cho toàn bộ tin nhắn trước đó.
2. Bắn sự kiện `chat:read` qua Redis tới người gửi.

### 3.2. Hiển thị trên Frontend Svelte:
Bong bóng tin nhắn của mình sẽ hiển thị:
- 1 tích xám: Tin nhắn đã lưu trên máy chủ.
- 2 tích xanh: Người nhận đã xem.

```html
{#if isMe}
  <span class="read-status" class:seen={msg.is_read}>
    {msg.is_read ? '✓✓' : '✓'}
  </span>
{/if}
```

---

<a name="buoc-4-tich-hop-am-thanh"></a>
## Bước 4: Tích hợp Âm thanh Thông báo trong Svelte

Tạo file `frontend/src/lib/utils/sound.js` sử dụng Web Audio API tích hợp sẵn (không cần tải file mp3 bên ngoài):

```javascript
// Phát âm thanh thông báo "Pop" nhẹ nhàng
export function playNotificationSound() {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = 'sine';
    osc.frequency.setValueAtTime(587.33, ctx.currentTime); // Nốt D5
    osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.1); // Nốt A5

    gain.gain.setValueAtTime(0.15, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.15);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.15);
  } catch (e) {
    console.warn('Không thể phát âm thanh:', e);
  }
}
```

---

<a name="buoc-5-kiem-thu-thuc-te"></a>
## Bước 5: Kiểm thử thực tế các tính năng nâng cao

1. **Test Online/Offline:**
   - Mở 2 tab trình duyệt cho 2 tài khoản.
   - Khi đóng 1 tab $\to$ Tab còn lại chuyển trạng thái sang Offline sau khoảng 30s.
2. **Test Typing Indicator:**
   - Bắt đầu gõ phím ở Tab 1 $\to$ Tab 2 hiện ngay animation 3 dấu chấm `...` nhảy động!
3. **Test Âm thanh:**
   - Để Tab 2 ở chế độ nền (background), gửi tin từ Tab 1 $\to$ Nghe thấy tiếng thông báo pop vui tai.

---

### 🎉 Hoàn thành Giai đoạn 5!
Ứng dụng chat của bạn đã sở hữu đầy đủ trải nghiệm mượt mà của một ứng dụng nhắn tin thương mại hiện đại! Sẵn sàng sang **Giai đoạn 6: Đóng gói Docker & Triển khai**!
