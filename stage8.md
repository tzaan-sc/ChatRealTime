# Hướng Dẫn Chi Tiết Giai Đoạn 8: Nhắn Tin Nhóm (Group Chat)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: Golang + Redis Pub/Sub + MongoDB + Svelte 5

Giai đoạn này sẽ mở rộng hệ thống từ nhắn tin 1-1 sang **Trò chuyện nhóm (Group Chat)** tương tự các nhóm Zalo, Telegram hay Discord:
1. 👥 **Tạo nhóm trò chuyện:** Đặt tên nhóm, chọn danh sách bạn bè ban đầu để thêm vào nhóm.
2. 👑 **Phân quyền Admin / Member:** Người tạo nhóm là Quản trị viên (Admin) có quyền đổi tên nhóm, thêm thành viên mới, xóa thành viên khỏi nhóm hoặc giải tán nhóm.
3. ⚡ **Kiến trúc Realtime Redis Pub/Sub cho nhóm:** 
   - Thay vì gửi từng message tới từng user riêng lẻ, server chỉ cần Publish tin nhắn vào 1 channel Redis duy nhất: `group:chat:{groupId}`.
   - Bất kỳ instance server nào có thành viên trong nhóm đang kết nối đều nhận được tin nhắn và phân phát xuống cho client tương ứng.
4. 🎨 **Giao diện Chat nhóm:**
   - Bong bóng chat hiển thị Tên và Avatar của người gửi phía trên tin nhắn.
   - Danh sách thành viên bên thanh thông tin nhóm (Group Info Sidebar).

---

## 📋 Checklist Tiến Độ Giai Đoạn 8

- [x] **1. Thiết kế Model & Cơ sở dữ liệu MongoDB:**
  - [x] Collection `groups`: `id`, `name`, `avatar`, `creator_id`, `admin_ids`, `member_ids`, `created_at`.
  - [x] Bổ sung trường `group_id`, `sender_name`, `sender_avatar` vào Collection `messages` (phân biệt tin 1-1 và tin nhóm).
- [x] **2. Xây dựng REST API Quản lý Nhóm (Golang):**
  - [x] `POST /api/groups`: Tạo nhóm mới kèm danh sách thành viên ban đầu.
  - [x] `GET /api/groups`: Lấy danh sách các nhóm mà người dùng hiện tại đang tham gia.
  - [x] `GET /api/groups/:id`: Lấy chi tiết nhóm và danh sách thành viên.
  - [x] `POST /api/groups/:id/members`: Thêm thành viên mới (chỉ Admin).
  - [x] `DELETE /api/groups/:id/members/:userId`: Xóa thành viên hoặc rời nhóm.
  - [x] `GET /api/groups/:id/messages`: Tải lịch sử tin nhắn của nhóm.
- [x] **3. Nâng cấp WebSocket Hub cho Group Chat:**
  - [x] Khi client kết nối: Server tự động subscribe vào các channel Redis `group:chat:{groupId}` của tất cả các nhóm mà user này tham gia.
  - [x] Xử lý event gửi tin nhắn nhóm `group:send` $\to$ Lưu DB với `group_id` $\to$ Publish vào `group:chat:{groupId}`.
  - [x] Hỗ trợ reaction, edit, delete trên tin nhắn nhóm theo `group_id`.
- [x] **4. Nâng cấp Frontend Svelte 5:**
  - [x] Modal tạo nhóm mới (`CreateGroupModal.svelte` Form nhập tên nhóm + Checkbox chọn bạn bè).
  - [x] Danh sách hội thoại Sidebar: Tab "Cá nhân" & Tab "Nhóm", nút "+ Tạo nhóm".
  - [x] Khung chat `ChatArea.svelte`: Hiển thị tên & avatar người gửi trên từng bong bóng chat và thanh tiêu đề nhóm.
  - [x] Modal xem danh sách thành viên & Thêm người vào nhóm (`GroupMembersModal.svelte`).
- [x] **5. Kiểm thử toàn diện các luồng chat nhóm**.

---

## Mục lục
1. [Bước 1: Thiết kế Model MongoDB cho Group Chat](#buoc-1-models)
2. [Bước 2: Xây dựng Repository & Service Quản lý Nhóm (Golang)](#buoc-2-backend-services)
3. [Bước 3: Tích hợp Redis Pub/Sub cho Nhóm vào WebSocket Hub](#buoc-3-websocket-group)
4. [Bước 4: Nâng cấp Frontend Svelte 5 (Modal Tạo Nhóm & Giao Diện Nhóm)](#buoc-4-frontend-group)
5. [Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết](#buoc-5-kiem-thu-thuc-te)

---

<a name="buoc-1-models"></a>
## Bước 1: Thiết kế Model MongoDB cho Group Chat

### 1.1. Tạo file `backend/internal/models/group.go`
```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name      string               `bson:"name" json:"name"`
	Avatar    string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	CreatorID primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	AdminIDs  []primitive.ObjectID `bson:"admin_ids" json:"admin_ids"`
	MemberIDs []primitive.ObjectID `bson:"member_ids" json:"member_ids"`
	CreatedAt time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updated_at"`
}

type CreateGroupRequest struct {
	Name      string   `json:"name" binding:"required"`
	MemberIDs []string `json:"member_ids"` // Danh sách ID bạn bè mời vào nhóm
}

type AddMemberRequest struct {
	MemberIDs []string `json:"member_ids" binding:"required"`
}
```

### 1.2. Cập nhật Model Message để hỗ trợ Group ID
Trong `backend/internal/models/message.go`, thêm trường `GroupID`:
```go
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id,omitempty" json:"conversation_id,omitempty"` // Dùng cho chat 1-1
	GroupID        string             `bson:"group_id,omitempty" json:"group_id,omitempty"`               // Dùng cho chat nhóm
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	ReceiverID     string             `bson:"receiver_id,omitempty" json:"receiver_id,omitempty"`
	Content        string             `bson:"content" json:"content"`
	Type           string             `bson:"type" json:"type"` // text, image, file, voice
	FileName       string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileSize       int64              `bson:"file_size,omitempty" json:"file_size,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}
```

---

<a name="buoc-2-backend-services"></a>
## Bước 2: Xây dựng Repository & Service Quản lý Nhóm (Golang)

Tạo file `backend/internal/repository/group_repository.go`:
```go
package repository

import (
	"context"
	"chatrealtime-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GroupRepository struct {
	collection *mongo.Collection
}

func NewGroupRepository(db *mongo.Database) *GroupRepository {
	return &GroupRepository{collection: db.Collection("groups")}
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	res, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		return err
	}
	group.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *GroupRepository) GetUserGroups(ctx context.Context, userID primitive.ObjectID) ([]models.Group, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"member_ids": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []models.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}
```

---

<a name="buoc-3-websocket-group"></a>
## Bước 3: Tích hợp Redis Pub/Sub cho Nhóm vào WebSocket Hub

Khi người dùng tham gia vào phòng chat nhóm:
1. Tin nhắn gửi lên WebSocket dạng:
```json
{
  "type": "group:send",
  "payload": {
    "group_id": "65fc123456789...",
    "content": "Chào mừng cả nhóm!",
    "type": "text"
  }
}
```
2. Hub lưu tin nhắn vào MongoDB và bắn qua Redis:
```go
redisClient.Publish(ctx, fmt.Sprintf("group:chat:%s", groupID), messageJSON)
```
3. Mọi client đang mở kết nối thuộc nhóm đó sẽ nhận được tin nhắn tức thì!

---

<a name="buoc-4-frontend-group"></a>
## Bước 4: Nâng cấp Frontend Svelte 5

### 4.1. Modal Tạo Nhóm Mới (`CreateGroupModal.svelte`)
- Người dùng bấm nút **"+"** cạnh danh sách chat $\to$ Mở modal.
- Điền tên nhóm: *"Team Dự Án Svelte & Go"*.
- Danh sách bạn bè hiện ra kèm checkbox $\to$ Tick chọn người muốn mời $\to$ Bấm **"Tạo nhóm"**.
- Nhóm mới xuất hiện ngay trên đầu danh sách Sidebar với avatar nhóm đặc trưng.

### 4.2. Hiển thị Bong bóng Chat Nhóm
- Đối với tin nhắn nhóm, hiển thị thêm Tên của người gửi phía trên nội dung tin nhắn với các màu sắc ngẫu nhiên bắt mắt (giống Telegram).

---

<a name="buoc-5-kiem-thu-thuc-te"></a>
## Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết

| Kịch bản | Thao tác thực hiện | Kết quả mong đợi |
| :--- | :--- | :--- |
| 👥 **Tạo nhóm mới** | User A bấm nút "+ Nhóm", đặt tên "Dev Team", chọn User B và C $\to$ Bấm Tạo | Nhóm lập tức xuất hiện trên Sidebar của cả A, B và C! |
| ⚡ **Broadcast tin nhắn** | User A gửi "Xin chào cả nhóm!" vào group | Cả User B và User C đều nhận được tin nhắn tức thì cùng lúc qua Redis Pub/Sub. |
| 🏷️ **Hiển thị Tên người gửi** | User B gửi "Chào Alex!" | Trên màn hình của User C, tin nhắn hiện rõ tên `[Bob]` phía trên bong bóng chat. |
| ➕ **Thêm thành viên mới** | Admin A vào cài đặt nhóm, chọn thêm User D | User D lập tức thấy nhóm xuất hiện trong danh sách chat của mình và xem được tin nhắn mới. |
| 🚪 **Rời nhóm** | User C bấm "Rời nhóm" | C không còn thấy nhóm trong danh sách chat; nhóm vẫn hoạt động bình thường với các thành viên còn lại. |

---

### 🎉 Hoàn thành Giai đoạn 8!
Hệ thống chat của bạn đã sở hữu khả năng **Trò chuyện nhóm thời gian thực (Scalable Group Chat)** cực kỳ mạnh mẽ!
