# Đánh Giá Kiến Trúc: Mô Hình "Nhắn Tin Cá Nhân" (Direct Messaging)
## Stack Đề Xuất Thực Tế: Svelte + Golang + Redis + MongoDB

Tài liệu này đánh giá tính khả thi, phân tích thiết kế dữ liệu và hướng dẫn triển khai hệ thống chat 1-1 theo thời gian thực (tương tự Messenger, Zalo, Telegram) với mục tiêu: **Dễ học, linh hoạt, hiệu năng cao, nhẹ máy và 100% miễn phí khi làm đồ án / pet project**.

---

## 1. Tổng quan Stack công nghệ

* **Frontend:** **Svelte** (hoặc SvelteKit) – Xử lý WebSocket, UI reactivity tức thì không cần Virtual DOM, dung lượng siêu nhẹ.
* **Backend:** **Golang** – Server WebSocket hiệu năng cao, Goroutine xử lý đồng thời hàng nghìn kết nối mà tốn rất ít RAM.
* **Tầng Nhẹ / In-Memory:** **Redis** – Đảm nhận 3 nhiệm vụ:
  1. **Pub/Sub**: Cầu nối truyền tin giữa các worker/server Go.
  2. **Presence**: Lưu trạng thái `Online / Offline` với cơ chế TTL (tự hết hạn sau khi mất heartbeat).
  3. **Typing Indicator**: Bắn tín hiệu "đang soạn tin..." tạm thời, không bao giờ ghi xuống đĩa.
* **Tầng Lưu Trữ Lâu Dài (Database):** **MongoDB** – Lưu trữ thông tin tài khoản, danh sách cuộc trò chuyện và lịch sử tin nhắn.

---

## 2. Vì sao Redis + MongoDB là combo "hoàn hảo" để học và làm dự án cá nhân?

| Tiêu chí | Apache Cassandra (Ban đầu) | Combo Redis + MongoDB (Lựa chọn mới) |
| :--- | :--- | :--- |
| **Tài nguyên máy (RAM/CPU)** | Ngốn 8GB – 16GB RAM JVM tối thiểu, chạy Docker máy cá nhân rất dễ đơ/lag. | Docker Redis (~30MB RAM) + MongoDB (~150-250MB RAM). Chạy mượt mà trên bất kỳ laptop nào. |
| **Chi phí Cloud (Deploy)** | Không có Free Tier thực tế, chi phí thuê VPS đắt đỏ. | • **Redis**: Upstash Redis (Free 10.000 req/ngày) hoặc Docker trên VPS $3-5.<br>• **MongoDB Atlas**: Free Tier vĩnh viễn (512MB M0 cluster, không cần thẻ visa). |
| **Độ phức tạp học tập** | Khó học, schema cứng, truy vấn CQL hạn chế (không hỗ trợ JOIN/Aggregation linh hoạt). | Cực kỳ trực quan, document dạng JSON/BSON đồng nhất với dữ liệu gửi nhận từ Go và Svelte. |
| **Hệ sinh thái Golang** | Driver CQL khá cồng kềnh. | Thư viện chính chủ rất xịn: `go.mongodb.org/mongo-driver` và `go-redis/v9`. |

---

## 3. Thiết kế Data Model trong MongoDB cho Chat 1-1

Đối với MongoDB, có 2 cách thiết kế để lưu trữ tin nhắn:

### Cách 1: Mô hình 1 Tin nhắn = 1 Document (Khuyên dùng cho mới bắt đầu)
Mỗi tin nhắn gửi đi là 1 document riêng biệt trong collection `messages`.

```json
// Collection: messages
{
  "_id": ObjectId("65f1a2b3c4d5e6f7a8b9c0d1"),
  "conversation_id": "userA_userB", // Sắp xếp theo alphabet ID để luôn duy nhất
  "sender_id": "userA",
  "receiver_id": "userB",
  "content": "Chào bạn, hôm nay thế nào?",
  "type": "text",                  // text, image, file, audio
  "is_read": false,
  "created_at": ISODate("2026-09-15T22:15:00Z")
}
```

* **Index bắt buộc trong MongoDB:**
  ```javascript
  db.messages.createIndex({ conversation_id: 1, created_at: -1 })
  ```
  > **Tác dụng:** Giúp câu lệnh query "Lấy 30 hoặc 50 tin nhắn gần nhất của cuộc trò chuyện" quét thẳng vào index theo chiều thời gian giảm dần, tốc độ đạt **vài mili-giây**.

### Cách 2: Kỹ thuật Gom cụm (Bucket Pattern - Dành cho tối ưu nâng cao)
Gom 50 tin nhắn vào một document chung gọi là 1 `bucket`. Cách này tiết kiệm dung lượng index và tăng tốc đọc lịch sử theo trang, nhưng logic chèn tin nhắn trong Go sẽ phức tạp hơn một chút. 

👉 **Lời khuyên:** Hãy bắt đầu với **Cách 1** trước vì nó đơn giản, dễ query, dễ xóa và sửa.

---

## 4. Phân chia nhiệm vụ chi tiết giữa Redis và MongoDB

| Chức năng | Xử lý ở đâu? | Cơ chế thực hiện |
| :--- | :--- | :--- |
| **Nội dung tin nhắn** | **MongoDB** | Ghi bất đồng bộ xuống MongoDB để lưu trữ vĩnh viễn. |
| **Danh sách cuộc hội thoại** | **MongoDB** | Lưu `conversation_id`, `members`, `last_message_at`. |
| **Định tuyến tin nhắn** | **Redis Pub/Sub** | Khi User A gửi, Go publish vào channel `chat:userB`. Server Go nào đang giữ socket của User B sẽ sub và bắn xuống client. |
| **Trạng thái Online / Offline** | **Redis (Key + TTL)** | Mỗi khi client ping heartbeat (ví dụ 10s/lần), Go set key `user:online:{userId}` với TTL = 25s. Nếu hết 25s không thấy ping $\to$ Redis tự hủy key $\to$ User chuyển sang Offline. |
| **Typing indicator ("Đang gõ...")** | **Redis Pub/Sub** | Client gửi event `typing:start`. Go bắn qua Redis channel tới người nhận, **không lưu vào MongoDB** để tránh phình dữ liệu rác. |
| **Số tin chưa đọc (Unread badge)** | **Redis Hash / MongoDB** | Tăng đếm nhanh trên Redis `HINCRBY unread:{userId} {conversationId} 1` hoặc update trực tiếp counter trong document hội thoại. |

---

## 5. Sơ đồ kiến trúc hoàn chỉnh

```mermaid
flowchart TD
    subgraph Clients ["Giao diện người dùng"]
        ClientA["Svelte App (User A)"]
        ClientB["Svelte App (User B)"]
    end

    subgraph GoBackend ["Golang Realtime Service"]
        WSHandler["WebSocket & REST Handler"]
        MsgWorker["Message Worker / Dispatcher"]
    end

    subgraph MemoryLayer ["Redis (Port 6379)"]
        PubSub["Pub/Sub Channel\n(chat:userId, typing:userId)"]
        Presence["In-Memory Key/Value\n(user:online:id với TTL)"]
    end

    subgraph DBLayer ["MongoDB (Port 27017)"]
        ColUsers[("users")]
        ColConversations[("conversations")]
        ColMessages[("messages (Indexed)")]
    end

    ClientA <-->|WebSocket| WSHandler
    ClientB <-->|WebSocket| WSHandler

    WSHandler <-->|Check / Set Heartbeat| Presence
    WSHandler <-->|Publish & Subscribe| PubSub

    WSHandler --> MsgWorker
    MsgWorker -->|Lưu tin nhắn| ColMessages
    MsgWorker -->|Cập nhật last_message| ColConversations
    WSHandler -->|Đọc hồ sơ & lịch sử| DBLayer
```

---

## 6. Hướng dẫn chạy môi trường Local qua Docker

Theo cấu hình trong file [MyConfig.md](file:///d:/GIT/ChatRealTime/MyConfig.md):

```bash
# Khởi động Redis và MongoDB chạy ngầm
docker-compose up -d

# Kiểm tra trạng thái container
docker ps
```

* **Redis:** Kết nối tại `localhost:6379`
* **MongoDB:** Kết nối tại `mongodb://admin:password123@localhost:27017`

---

## 7. Đánh giá tổng kết

* **Độ khả thi:** **10 / 10** cho dự án học tập, đồ án hoặc MVP thực tế.
* **Chi phí:** **0 VNĐ** (Hoàn toàn miễn phí trên máy local, và có thể deploy bản demo lên Render / Fly.io / MongoDB Atlas Free Tier).
* **Trải nghiệm code:** Go làm việc với MongoDB và Redis qua JSON rất mượt và ít boilerplate hơn nhiều so với Cassandra.
