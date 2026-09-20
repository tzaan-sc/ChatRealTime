# Hướng Dẫn Chi Tiết Giai Đoạn 7: Tương Tác Tin Nhắn (Reaction Emoji, Trả Lời / Quote, Thu Hồi & Chỉnh Sửa)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: Golang + Redis Pub/Sub + MongoDB + Svelte 5

Giai đoạn này sẽ mang lại các tính năng tương tác mượt mà và trực quan quen thuộc của **Telegram & Messenger**:
1. ❤️ **Thả Reaction Emoji:** Bấm hoặc rê chuột vào tin nhắn để chọn cảm xúc: ❤️, 😂, 👍, 😢, 🔥, 🚀 $\to$ Icon cảm xúc nhảy ra dưới góc tin nhắn cả 2 bên.
2. 💬 **Trả lời / Trích dẫn (Reply / Quote):** Bấm "Trả lời" một tin nhắn cụ thể $\to$ Phía trên ô nhập hiện trích dẫn $\to$ Khi gửi, bong bóng chat sẽ trích dẫn kèm tin nhắn gốc (bấm vào trích dẫn sẽ tự cuộn tới tin gốc).
3. 🗑️ **Thu hồi tin nhắn (Unsend / Delete for Everyone):** Cho phép người gửi xoá tin nhắn $\to$ Tin nhắn đổi thành *"Tin nhắn đã được thu hồi"* ở cả hai màn hình.
4. ✏️ **Chỉnh sửa nội dung (Edit Message):** Cho phép sửa nhanh lỗi chính tả tin nhắn vừa gửi $\to$ Tin nhắn cập nhật nội dung mới tức thì kèm nhãn nhỏ `(đã chỉnh sửa)`.

---

## 📋 Checklist Tiến Đo Giai Đoạn 7

- [x] **1. Cập nhật Model & Database Schema trong Backend (Golang):**
  - [x] Thêm struct `Reaction` và trường `reactions []Reaction` vào `models.Message`.
  - [x] Thêm struct `ReplySnippet` và trường `reply_to *ReplySnippet` vào `models.Message`.
  - [x] Thêm các cờ `is_deleted bool` và `is_edited bool` vào `models.Message`.
- [x] **2. Cập nhật WebSocket Hub (`backend/internal/websocket/hub.go`):**
  - [x] Xử lý event `chat:react`: Cập nhật reaction trong MongoDB $\to$ Publish `chat:reaction_updated` qua Redis.
  - [x] Xử lý event `chat:delete`: Đặt `is_deleted = true` trong MongoDB $\to$ Publish `chat:message_deleted`.
  - [x] Xử lý event `chat:edit`: Cập nhật `content` mới trong MongoDB $\to$ Publish `chat:message_edited`.
  - [x] Hỗ trợ gửi tin nhắn có `reply_to` trong event `chat:send`.
- [x] **3. Nâng cấp Store Chat trên Frontend (`frontend/src/lib/stores/chat.js`):**
  - [x] Thêm hàm `updateReactionLocally(messageId, userId, emoji)`.
  - [x] Thêm hàm `markDeletedLocally(messageId)`.
  - [x] Thêm hàm `updateEditedLocally(messageId, newContent)`.
- [x] **4. Nâng cấp Giao diện Khung Chat (`frontend/src/lib/components/ChatArea.svelte`):**
  - [x] Thanh Quick Reaction Bar hiển thị khi hover chuột vào bong bóng chat.
  - [x] Nút Action Menu: Trả lời 💬, Sửa ✏️ (nếu là tin của mình), Thu hồi 🗑️ (nếu là tin của mình).
  - [x] Banner trích dẫn phía trên ô gõ text khi đang trong chế độ Reply.
  - [x] Banner chỉnh sửa tin nhắn phía trên ô gõ text khi đang sửa tin.
  - [x] Hiển thị danh sách Reaction gộp (ví dụ: ❤️ 2, 😂 1) ở góc dưới tin nhắn.
- [x] **5. Kiểm thử toàn diện 6 kịch bản tương tác tin nhắn**.

---

## Mục lục
1. [Bước 1: Cập nhật Model MongoDB & Backend Go](#buoc-1-models)
2. [Bước 2: Xử lý Event Tương Tác trong WebSocket Hub](#buoc-2-hub)
3. [Bước 3: Nâng cấp Stores Frontend (`stores/chat.js`)](#buoc-3-stores)
4. [Bước 4: Nâng cấp Giao diện Khung Chat (`ChatArea.svelte`)](#buoc-4-ui)
5. [Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết](#buoc-5-kiem-thu-thuc-te)

---

<a name="buoc-1-models"></a>
## Bước 1: Cập nhật Model MongoDB & Backend Go

Mở file [backend/internal/models/message.go](file:///d:/GIT/ChatRealTime/backend/internal/models/message.go):

```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reaction struct {
	UserID string `bson:"user_id" json:"user_id"`
	Emoji  string `bson:"emoji" json:"emoji"`
}

type ReplySnippet struct {
	MessageID  string `bson:"message_id" json:"message_id"`
	SenderName string `bson:"sender_name" json:"sender_name"`
	Content    string `bson:"content" json:"content"`
}

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id" json:"conversation_id"`
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	ReceiverID     string             `bson:"receiver_id" json:"receiver_id"`
	Content        string             `bson:"content" json:"content"`
	Type           string             `bson:"type" json:"type"`
	FileName       string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileSize       int64              `bson:"file_size,omitempty" json:"file_size,omitempty"`
	Reactions      []Reaction         `bson:"reactions,omitempty" json:"reactions,omitempty"`
	ReplyTo        *ReplySnippet      `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	IsDeleted      bool               `bson:"is_deleted" json:"is_deleted"`
	IsEdited       bool               `bson:"is_edited" json:"is_edited"`
	IsRead         bool               `bson:"is_read" json:"is_read"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
```

---

<a name="buoc-2-hub"></a>
## Bước 2: Xử lý Event Tương Tác trong WebSocket Hub

Trong [backend/internal/websocket/hub.go](file:///d:/GIT/ChatRealTime/backend/internal/websocket/hub.go), bổ sung các trường hợp:

### 2.1. Thả Reaction (`chat:react`)
```go
case "chat:react":
    // payload: { "message_id": "...", "receiver_id": "...", "emoji": "❤️" }
    // Cập nhật MongoDB $addToSet hoặc $pull reaction
    // Publish event "chat:reaction_updated" qua Redis channel của cả 2 bên
```

### 2.2. Thu hồi tin nhắn (`chat:delete`)
```go
case "chat:delete":
    // payload: { "message_id": "...", "receiver_id": "..." }
    // Cập nhật MongoDB: set is_deleted = true, content = ""
    // Publish event "chat:message_deleted" cho người nhận
```

### 2.3. Chỉnh sửa tin nhắn (`chat:edit`)
```go
case "chat:edit":
    // payload: { "message_id": "...", "receiver_id": "...", "content": "Nội dung mới" }
    // Cập nhật MongoDB: set content = newContent, is_edited = true, updated_at = time.Now()
    // Publish event "chat:message_edited" cho người nhận
```

---

<a name="buoc-3-stores"></a>
## Bước 3: Nâng cấp Stores Frontend (`stores/chat.js`)

Thêm các helper hàm xử lý tức thì trên giao diện Svelte:
```javascript
export function updateMessageReaction(messageId, userId, emoji) {
  activeMessages.update(msgs => msgs.map(m => {
    if (m.id === messageId) {
      let reactions = m.reactions ? [...m.reactions] : [];
      const existingIdx = reactions.findIndex(r => r.user_id === userId);
      if (existingIdx !== -1) {
        if (reactions[existingIdx].emoji === emoji) {
          // Bấm lại emoji cũ -> Xóa reaction
          reactions.splice(existingIdx, 1);
        } else {
          // Đổi emoji
          reactions[existingIdx].emoji = emoji;
        }
      } else {
        reactions.push({ user_id: userId, emoji });
      }
      return { ...m, reactions };
    }
    return m;
  }));
}

export function updateMessageDeleted(messageId) {
  activeMessages.update(msgs => msgs.map(m => {
    if (m.id === messageId) {
      return { ...m, is_deleted: true, content: '' };
    }
    return m;
  }));
}

export function updateMessageEdited(messageId, newContent) {
  activeMessages.update(msgs => msgs.map(m => {
    if (m.id === messageId) {
      return { ...m, content: newContent, is_edited: true };
    }
    return m;
  }));
}
```

---

<a name="buoc-4-ui"></a>
## Bước 4: Nâng cấp Giao diện Khung Chat (`ChatArea.svelte`)

### 4.1. Quick Reaction Bar & Action Menu
Khi rê chuột vào một bong bóng tin nhắn bất kỳ:
- Hiện thanh popup gồm 6 emoji: ❤️ 😂 👍 😢 🔥 🚀
- Các nút hành động:
  - 💬 **Trả lời (Reply):** Kích hoạt trích dẫn.
  - ✏️ **Chỉnh sửa (Edit):** Đưa nội dung tin vào ô nhập để sửa (chỉ hiện với tin của chính mình).
  - 🗑️ **Thu hồi (Delete):** Bấm để xoá vĩnh viễn nội dung tin nhắn.

### 4.2. Hiển thị Bong bóng Chat có Trích dẫn & Thu hồi
- **Nếu `is_deleted === true`:** Hiển thị khối chữ xám mờ in nghiêng: *"🚫 Tin nhắn đã được thu hồi"*.
- **Nếu có `reply_to`:** Hiển thị khung nhỏ viền màu tím nhạt trích dẫn tên người gửi và nội dung tin gốc.
- **Nếu có `is_edited === true`:** Hiển thị nhãn nhỏ `(đã sửa)` cạnh thời gian gửi.

---

<a name="buoc-5-kiem-thu-thuc-te"></a>
## Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết

| Kịch bản | Thao tác thực hiện | Kết quả mong đợi |
| :--- | :--- | :--- |
| ❤️ **Thả Reaction Emoji** | Rê chuột vào tin nhắn của đối phương $\to$ bấm emoji ❤️ | Góc dưới bong bóng chat hiện biểu tượng `❤️ 1`. Đối phương thấy ngay tức thì mà không cần tải lại trang. Bấm lại vào tim để hủy reaction. |
| 💬 **Trả lời tin nhắn (Reply)** | Bấm nút Reply trên một tin nhắn $\to$ Gõ "Tôi đồng ý với ý này" $\to$ Gửi | Tin nhắn gửi đi có thanh trích dẫn hiển thị nội dung tin cũ bên trên. Click vào thanh trích dẫn tự động cuộn lên xem tin gốc! |
| ✏️ **Chỉnh sửa tin nhắn** | Bấm nút Sửa ✏️ trên tin vừa gửi $\to$ Sửa thành nội dung đúng $\to$ Bấm Lưu | Cả 2 màn hình thấy nội dung tin nhắn được cập nhật ngay lập tức kèm nhãn `(đã sửa)`. |
| 🗑️ **Thu hồi tin nhắn** | Bấm nút Xóa 🗑️ trên tin vừa gửi $\to$ Xác nhận thu hồi | Tin nhắn biến đổi thành *"🚫 Tin nhắn đã được thu hồi"* trên cả màn hình của mình và đối phương! |
| 🔀 **Đa phản ứng (Multiple Reactions)** | Alex thả ❤️, Bob thả 🔥 vào cùng 1 tin nhắn | Góc dưới tin nhắn nhóm lại thành 2 icon: `❤️ 1` `🔥 1`. |

---

### 🎉 Hoàn thành Giai đoạn 7!
Hệ thống chat của bạn đã có đầy đủ trải nghiệm tương tác cao cấp không thua kém bất kỳ ứng dụng nhắn tin thương mại nào!
