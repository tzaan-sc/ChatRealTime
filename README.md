# 🚀 Real-time Chat & WebRTC Communication System

Hệ thống nhắn tin thời gian thực (Real-time Messaging) và gọi thoại/video P2P (WebRTC) hiệu năng cao, được xây dựng với kiến trúc hướng sự kiện phân tán (**Event-Driven Architecture**) sử dụng **Go (Golang)**, **Redis Pub/Sub**, **MongoDB** và **Svelte 5**.

---

## 🌟 Tính Năng Nổi Bật

- 💬 **Nhắn tin Thời Gian Thực (1-1 & Nhóm):**
  - Hỗ trợ văn bản, hình ảnh xem trước (Lightbox), tệp đính kèm và tin nhắn thoại (Voice Note phát trực tiếp).
  - Trả lời tin nhắn (Reply quote) & Chỉnh sửa tin nhắn (Edit).
  - Thu hồi / Xóa tin nhắn theo thời gian thực (Unsend/Delete).
  - Thả biểu cảm cảm xúc đa dạng (Message Reactions).
- 👁️ **Cơ chế Đã xem / Chưa xem (Read Receipts):**
  - Hiển thị 1 tích xám `✓` (Đã gửi) ➔ 2 tích xanh biển `✓✓` (Đã xem).
  - Tự động nhận diện tab đang mở và tab bị thu nhỏ/ẩn để đánh dấu đã đọc chính xác.
  - Hỗ trợ tính năng **"Đánh dấu là chưa đọc" (Mark as Unread)** để lưu việc trả lời sau.
  - Huy hiệu số tin chưa đọc (Unread Counter Badges) cập nhật tức thời (Optimistic UI) cho cả Chat cá nhân và Nhóm.
- 🔔 **Hệ thống Thông báo Đa tầng:**
  - **In-App Toast Banner:** Thông báo nổi góc màn hình khi có tin nhắn từ bạn bè/nhóm khác.
  - **Desktop Notification:** Thông báo trình duyệt khi tab đang thu nhỏ hoặc ẩn.
  - **Dynamic Title:** Nhấp nháy tiêu đề trang web `(1) 💬 [Tên người gửi]...` khi có tin mới.
  - **Web Audio API:** Âm thanh Pop vui tai (nền) và âm Click êm dịu (khi đang chat), chống rò rỉ bộ nhớ, có nút bật/tắt chuông nhanh.
- 🟢 **Trạng thái Hoạt động:**
  - Nhận diện trạng thái Trực tuyến / Ngoại tuyến (Online / Offline) qua Redis TTL.
  - Hiển thị hiệu ứng ba chấm đang soạn thảo (Typing Indicator).
- 📞 **Gọi Thoại & Video P2P (WebRTC):**
  - Cuộc gọi thoại (Audio Call) và gọi Video trực tiếp giữa 2 người dùng.
  - Nhạc chuông reo khi có cuộc gọi đến và âm tút khi đang gọi đi.
  - Điều khiển Mic, Camera và chế độ toàn màn hình.

---

## 🛠️ Công Nghệ Sử Dụng

| Tầng | Công nghệ / Thư viện | Vai trò |
|---|---|---|
| **Backend** | Go (Golang 1.22+) | Xử lý logic máy chủ, RESTful APIs & WebSocket Hub |
| **Framework** | Gin Gonic | Web framework hiệu năng cao cho REST API |
| **WebSocket** | Gorilla WebSocket | Quản lý kết nối hai chiều thời gian thực |
| **Message Broker** | Redis 7 (Pub/Sub) | Điều phối tin nhắn tức thời giữa các kênh và phân tán session |
| **Database** | MongoDB 7 | Lưu trữ dữ liệu người dùng, hội thoại, tin nhắn và nhóm |
| **Frontend** | Svelte 5 + Vite | Giao diện Cyber Dark hiện đại, phản ứng cực nhanh |
| **Icon Set** | Lucide Svelte | Bộ icon hiện đại, tối giản |
| **Signaling** | WebRTC Native API | Kết nối truyền thông âm thanh/hình ảnh Peer-to-Peer |
| **Container** | Docker & Docker Compose | Đóng gói và chạy môi trường Redis & MongoDB |

---

## 📋 Yêu Cầu Hệ Thống (Prerequisites)

Trước khi bắt đầu, hãy đảm bảo máy tính của bạn đã cài đặt:

1. **Docker Desktop:** Đang chạy để khởi động MongoDB và Redis ([Tải Docker](https://www.docker.com/products/docker-desktop/)).
2. **Go (Golang):** Phiên bản **1.22+** ([Tải Golang](https://go.dev/dl/)).
3. **Node.js:** Phiên bản **18.x** hoặc **20.x+** ([Tải Node.js](https://nodejs.org/)).

---

## 🚀 Hướng Dẫn Cài Đặt & Chạy Hệ Thống

### Bước 1: Khởi động Cơ sở Dữ liệu (Redis & MongoDB)

Mở một cửa sổ Terminal (PowerShell hoặc Command Prompt) tại thư mục gốc dự án:

```powershell
# Khởi chạy Redis và MongoDB trong nền qua Docker Compose
docker compose up -d
```

> 💡 **Kiểm tra trạng thái:** Chạy `docker ps` để đảm bảo 2 container `mongo` (cổng `27017`) và `redis` (cổng `6379`) đang ở trạng thái **Up**.

---

### Bước 2: Khởi động Backend (Go Server)

Mở một cửa sổ Terminal mới:

```powershell
# 1. Di chuyển vào thư mục backend
cd backend

# 2. Tải các thư viện phụ thuộc (nếu là lần đầu chạy)
go mod download

# 3. Khởi chạy máy chủ Backend
go run ./cmd/server
```

Khi khởi động thành công, màn hình sẽ hiển thị:
```text
Kết nối MongoDB thành công!
Kết nối Redis thành công!
🚀 Server Chat Realtime đang chạy tại http://localhost:8080
```

> ⚙️ **Cấu hình môi trường Backend:** Tệp cấu hình nằm tại [backend/.env](file:///d:/GIT/ChatRealTime/backend/.env):
> ```env
> PORT=8080
> MONGO_URI=mongodb://admin:password123@localhost:27017
> DB_NAME=chatapp
> REDIS_ADDR=localhost:6379
> JWT_SECRET=super_secret_key_chatapp_2026
> ```

---

### Bước 3: Khởi động Frontend (Svelte App)

Mở một cửa sổ Terminal tiếp theo:

```powershell
# 1. Di chuyển vào thư mục frontend
cd frontend

# 2. Cài đặt các gói npm (nếu chưa cài)
npm install

# 3. Khởi chạy máy chủ phát triển Vite
npm run dev
```

Màn hình sẽ thông báo:
```text
  VITE v8.x.x  ready in ... ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

👉 Hãy truy cập trình duyệt tại: **`http://localhost:5173`**

---

## 🧪 Hướng Dẫn Trải Nghiệm Hệ Thống (Demo Flow)

Để kiểm thử trọn vẹn các tính năng Real-time giữa 2 người dùng:

1. **Mở 2 cửa sổ trình duyệt riêng biệt:**
   - Cửa sổ 1: Mở trình duyệt thông thường tại `http://localhost:5173`
   - Cửa sổ 2: Mở cửa sổ **Ẩn danh (Incognito)** tại `http://localhost:5173`
2. **Đăng ký / Đăng nhập:**
   - Cửa sổ 1: Đăng ký tài khoản `alex` (Tên hiển thị: **Alex Nguyễn**)
   - Cửa sổ 2: Đăng ký tài khoản `bob` (Tên hiển thị: **Bob Trần**)
3. **Bắt đầu nhắn tin:**
   - Trên tài khoản Alex, bấm nút **`+` (Nhắn tin cá nhân mới)** ở thanh Sidebar ➔ Chọn **Bob Trần**.
   - Gửi tin nhắn: Tin nhắn sẽ đến Bob ngay tức khắc kèm âm thanh thông báo.
4. **Kiểm tra trạng thái Đã xem (Read Receipts):**
   - Khi Bob đang mở cửa sổ chat: Tin nhắn của Alex lập tức đổi thành **2 tích xanh biển `✓✓`**.
   - Khi Bob thu nhỏ hoặc chuyển sang tab khác: Alex chỉ thấy **1 tích xám `✓`**, đồng thời tab của Bob sẽ nhấp nháy tiêu đề `(1) 💬 Alex Nguyễn:...` và hiện thông báo Desktop.
   - Khi Bob bấm chuột lại vào tab: Tích xanh lập tức hiện lên bên Alex.
5. **Kiểm tra tính năng "Đánh dấu là chưa đọc":**
   - Nhấp chuột phải (hoặc bấm biểu tượng 3 chấm) vào cuộc trò chuyện trên Sidebar ➔ Chọn **"Đánh dấu là chưa đọc"**.
   - Cuộc trò chuyện sẽ sáng huy hiệu số tin chưa đọc.
6. **Thử nghiệm Cuộc gọi P2P WebRTC:**
   - Trên thanh tiêu đề chat với bạn bè, bấm biểu tượng **Cuộc gọi thoại (Điện thoại)** hoặc **Gọi Video (Máy quay)**.
   - Bên người nhận sẽ vang chuông reo cùng modal chấp nhận cuộc gọi. Bấm **Chấp nhận** để kết nối đàm thoại hai chiều trực tiếp.

---

## 📁 Cấu Trúc Thư Mục Dự Án

```text
ChatRealTime/
├── backend/
│   ├── cmd/server/main.go          # Điểm khởi chạy chính của Backend
│   ├── internal/
│   │   ├── handlers/               # Điều phối HTTP REST APIs (Auth, Chat, Group, Upload)
│   │   ├── middleware/             # Middleware xác thực JWT
│   │   ├── models/                 # Structs dữ liệu (Message, Conversation, User, Group, WS)
│   │   ├── repository/             # Tầng truy xuất CSDL MongoDB
│   │   ├── service/                # Tầng nghiệp vụ xử lý logic
│   │   └── websocket/              # WebSocket Hub, Client connection & Redis Pub/Sub
│   ├── pkg/utils/                  # Hàm tiện ích (JWT, Password Hashing)
│   ├── uploads/                    # Thư mục lưu trữ tệp đính kèm & voice notes
│   ├── .env                        # Cấu hình cổng và chuỗi kết nối
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/         # Các thành phần giao diện (ChatArea, Sidebar, CallModal...)
│   │   │   ├── services/           # Kết nối API REST, WebSocket client, Notification service
│   │   │   ├── stores/             # Svelte Stores (auth, chat, call, notification)
│   │   │   └── utils/              # Tiện ích âm thanh Web Audio API (sound.js)
│   │   ├── App.svelte              # Component gốc điều phối toàn ứng dụng
│   │   └── app.css                 # Bảng màu, tokens và phong cách Cyber Dark
│   └── package.json
├── docker-compose.yml              # Cấu hình container Redis 7 & MongoDB 7
└── README.md                       # Tài liệu hướng dẫn sử dụng
```

---

## 🔧 Xử Lý Sự Cố Thường Gặp (Troubleshooting)

| Vấn đề | Nguyên nhân | Cách khắc phục |
|---|---|---|
| **Lỗi kết nối MongoDB/Redis** | Container Docker chưa được bật | Chạy `docker compose up -d` và kiểm tra lại trạng thái bằng `docker ps`. |
| **Cổng 8080 hoặc 5173 đã bị chiếm dụng** | Có ứng dụng khác đang dùng cổng | Đổi cổng trong tệp [backend/.env](file:///d:/GIT/ChatRealTime/backend/.env) (biến `PORT`) hoặc chạy frontend với `npm run dev -- --port 5174`. |
| **Không nghe thấy âm thanh thông báo** | Trình duyệt chặn tự động phát âm thanh (Autoplay) | Nhấp chuột ít nhất 1 lần vào trang web để kích hoạt quyền phát âm thanh của Web Audio API, hoặc kiểm tra nút chuông thông báo trên Sidebar đã bật chưa. |
| **Không thấy thông báo Desktop** | Chưa cấp quyền thông báo trình duyệt | Bấm vào biểu tượng ổ khóa cạnh URL trình duyệt ➔ Cho phép quyền **Thông báo (Notifications)**. |
| **Lỗi kết nối WebRTC (Không thấy hình)** | Chưa cấp quyền Camera / Microphone | Cho phép trình duyệt truy cập thiết bị Micro và Camera khi có thông báo hỏi. |

---

## 📄 Bản Quyền & Giấy Phép

Dự án được xây dựng phục vụ mục đích học tập, nghiên cứu và phát triển các hệ thống truyền thông thời gian thực quy mô lớn.
Mọi thắc mắc hoặc đóng góp vui lòng mở Issue hoặc Pull Request! 🚀
