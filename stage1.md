# Hướng Dẫn Chi Tiết Giai Đoạn 1: Khởi Tạo Hạ Tầng & Môi Trường Dữ Liệu
> **Dành cho người mới bắt đầu** | Hệ thống: Windows | Stack: Docker, Redis 7, MongoDB 7

Tài liệu này sẽ hướng dẫn bạn từ lúc máy tính chưa có gì đến khi có một cụm cơ sở dữ liệu hoàn chỉnh chạy ngầm và sẵn sàng kết nối với Backend Go.

---

## Mục lục
1. [Bước 1: Cài đặt WSL 2 & Docker Desktop](#buoc-1-cai-dat-wsl-2--docker-desktop)
2. [Bước 2: Hiểu cấu trúc file `docker-compose.yml`](#buoc-2-hieu-cau-truc-file-docker-composeyml)
3. [Bước 3: Khởi chạy cụm dịch vụ Redis & MongoDB](#buoc-3-khoi-chay-cum-dich-vu-redis--mongodb)
4. [Bước 4: Kiểm tra hoạt động của Redis](#buoc-4-kiem-tra-hoat-dong-cua-redis)
5. [Bước 5: Cài đặt MongoDB Compass & Kết nối CSDL](#buoc-5-cai-dat-mongodb-compass--ket-noi-csdl)
6. [Bước 6: Tạo Database, Collections và thiết lập Index tối ưu](#buoc-6-tao-database-collections-va-thiet-lap-index-toi-uu)
7. [Các câu lệnh thường dùng & Xử lý sự cố thường gặp (FAQ)](#cac-cau-lenh-thuong-dung--xu-ly-su-co-thuong-gap-faq)

---

## 📋 Checklist Các Bước Cần Thực Hiện

- [x] **1. Cài đặt WSL 2 & Docker Desktop** *(Đã hoàn thành & máy đã restart)*
- [ ] **2. Khởi chạy Redis & MongoDB:**
  ```powershell
  docker compose up -d
  ```
- [ ] **3. Kiểm tra Container đang chạy:**
  ```powershell
  docker ps
  ```
- [ ] **4. (Tuỳ chọn) Kiểm tra thử Redis CLI:**
  ```powershell
  docker exec -it redis redis-cli ping
  # Nhận lại: PONG là hoàn tất
  ```
- [ ] **5. Cài đặt MongoDB Compass & Kết nối:**
  - Tải/cài MongoDB Compass (`winget install -e --id MongoDB.Compass.Full`)
  - Kết nối URI: `mongodb://admin:password123@localhost:27017`
- [ ] **6. Tạo Database & Thiết lập Index trong MongoDB:**
  - Tạo Database: `chatapp`
  - Tạo 3 Collections: `users`, `conversations`, `messages`
  - Đánh Compound Index cho `messages`: `{ conversation_id: 1, created_at: -1 }`
- [ ] **7. Chuyển sang Giai đoạn 2 (Backend Go):** Xem [stage2.md](file:///d:/GIT/ChatRealTime/stage2.md)

---

<a name="buoc-1-cai-dat-wsl-2--docker-desktop"></a>
## Bước 1: Cài đặt WSL 2 & Docker Desktop

Docker trên Windows cần hệ thống Linux ảo hóa siêu nhẹ (WSL 2 (Windows Subsystem for Linux 2)) để chạy các container.

### 1.1. Bật WSL 2
1. Mở menu Start, gõ `powershell`.
2. Click chuột phải vào **Windows PowerShell** $\to$ chọn **Run as administrator** (Chạy với quyền Quản trị viên).
3. Gõ lệnh sau và nhấn Enter:
   ```powershell
   wsl --install
   ```
4. Khi màn hình báo cài đặt xong, hãy **Khởi động lại máy tính (Restart)** nếu hệ thống yêu cầu.

### 1.2. Cài đặt Docker Desktop
Bạn có thể cài đặt tự động bằng lệnh sau ngay trong PowerShell:
```powershell
winget install -e --id Docker.DockerDesktop
```
*(Hoặc tải file cài đặt thủ công tại: [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/))*.

### 1.3. Khởi động và kiểm tra Docker
1. Mở Start menu $\to$ mở ứng dụng **Docker Desktop**.
2. Chọn **Accept** các điều khoản $\to$ chọn **Use recommended settings** $\to$ bấm **Finish**.
3. **Quan sát:** Ở góc dưới bên trái màn hình Docker Desktop, khi nào biểu tượng chú cá voi chuyển sang **màu xanh lá (Engine running)** là Docker đã hoạt động thành công.
4. Mở cửa sổ PowerShell mới và gõ 2 lệnh kiểm tra:
   ```powershell
   docker --version
   # Kết quả: Docker version 27.x.x...

   docker compose version
   # Kết quả: Docker Compose version v2.x.x...
   ```

---

<a name="buoc-2-hieu-cau-truc-file-docker-composeyml"></a>
## Bước 2: Hiểu cấu trúc file `docker-compose.yml`

Trong thư mục dự án `d:\GIT\ChatRealTime\docker-compose.yml`, chúng ta đã có nội dung sau:

```yaml
version: '3.8'

services:
  redis:
    image: redis:7                  # Tải bản Redis 7 chính thức
    container_name: redis           # Đặt tên container cho dễ nhớ
    restart: always                 # Tự bật lại nếu bị crash hoặc restart máy
    ports:
      - "6379:6379"                 # Ánh xạ cổng: [Máy thật]:[Bên trong Docker]

  mongo:
    image: mongo:7                  # Tải bản MongoDB 7 chính thức
    container_name: mongo
    restart: always
    ports:
      - "27017:27017"               # Cổng mặc định của MongoDB
    environment:                    # Tài khoản quản trị khởi tạo
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password123
    volumes:
      - mongo_data:/data/db         # Lưu dữ liệu ra ổ đĩa máy thật (Không bị mất dữ liệu khi tắt container)

volumes:
  mongo_data:                       # Khai báo phân vùng lưu trữ
```

> 💡 **Giải thích:** Nhờ có dòng `volumes: mongo_data:/data/db`, toàn bộ tin nhắn chat bạn tạo ra sẽ không bao giờ bị mất đi dù bạn có tắt máy hay xóa container.

---

<a name="buoc-3-khoi-chay-cum-dich-vu-redis--mongodb"></a>
## Bước 3: Khởi chạy cụm dịch vụ Redis & MongoDB

Mở PowerShell tại thư mục dự án `d:\GIT\ChatRealTime` (hoặc mở Terminal tích hợp trong IDE):

```powershell
# Chạy cả Redis và MongoDB dưới dạng tiến trình ngầm (-d nghĩa là detached mode)
docker compose up -d
```

* Lần đầu chạy, Docker sẽ tự động tải các image về máy (mất khoảng 1–2 phút tùy tốc độ mạng).
* Sau khi tải xong, màn hình sẽ hiển thị:
  ```text
  [+] Running 3/3
   ✔ Network chatrealtime_default  Created
   ✔ Container redis              Started
   ✔ Container mongo              Started
  ```

### Kiểm tra danh sách container đang chạy:
```powershell
docker ps
```
Nếu thấy cả `redis` và `mongo` đều có trạng thái **Up ... (healthy)** là bạn đã khởi động thành công!

---

<a name="buoc-4-kiem-tra-hoat-dong-cua-redis"></a>
## Bước 4: Kiểm tra hoạt động của Redis

Bạn có thể chui thẳng vào bên trong container Redis để thử gửi dữ liệu bằng lệnh `redis-cli`:

1. Gõ lệnh mở CLI của Redis:
   ```powershell
   docker exec -it redis redis-cli
   ```
2. Gõ lệnh ping thử:
   ```text
   127.0.0.1:6379> ping
   PONG
   ```
3. Thử lưu và đọc một key kiểm tra:
   ```text
   127.0.0.1:6379> set hello "chatapp"
   OK
   127.0.0.1:6379> get hello
   "chatapp"
   ```
4. Thoát khỏi Redis CLI:
   ```text
   127.0.0.1:6379> exit
   ```
*(Redis hoạt động bình thường, sẵn sàng làm Pub/Sub và lưu Presence)*.

---

<a name="buoc-5-cai-dat-mongodb-compass--ket-noi-csdl"></a>
## Bước 5: Cài đặt MongoDB Compass & Kết nối CSDL

**MongoDB Compass** là phần mềm giao diện đồ họa chính thức của MongoDB, giúp bạn bấm chuột xem dữ liệu tin nhắn cực kỳ trực quan mà không cần gõ lệnh phức tạp.

1. **Tải phần mềm:**
   * Tải miễn phí tại: [https://www.mongodb.com/try/download/compass](https://www.mongodb.com/try/download/compass) (Chọn bản Windows x64 .exe).
   * Hoặc cài bằng winget:
     ```powershell
     winget install -e --id MongoDB.Compass.Full
     ```
2. **Mở MongoDB Compass và kết nối:**
   * Khởi động phần mềm MongoDB Compass.
   * Tại ô **URI** (hoặc ô nhập chuỗi kết nối), bạn dán chính xác chuỗi sau:
     ```text
     mongodb://admin:password123@localhost:27017
     ```
   * Bấm nút **Connect**.
   * Khi kết nối thành công, bạn sẽ thấy danh sách các CSDL mặc định của hệ thống (`admin`, `config`, `local`).

---

<a name="buoc-6-tao-database-collections-va-thiet-lap-index-toi-uu"></a>
## Bước 6: Tạo Database, Collections và thiết lập Index tối ưu

### 6.1. Tạo Database `chatapp` và các Collections
Ngay trong giao diện MongoDB Compass:
1. Bấm vào dấu **`+` (Create database)** ở góc trên thanh danh sách bên trái.
2. Điền thông tin:
   * **Database Name:** `chatapp`
   * **Collection Name:** `messages`
3. Bấm **Create Database**.
4. Tạo thêm 2 collection nữa trong database `chatapp` bằng cách di chuột vào `chatapp`, bấm dấu `+`:
   * Collection thứ hai: `users` (dùng lưu tài khoản, avatar, mật khẩu băm).
   * Collection thứ ba: `conversations` (dùng lưu danh sách phòng chat 1-1).

### 6.2. Tạo Index tối ưu cho tin nhắn (CỰC KỲ QUAN TRỌNG)
Trong ứng dụng chat, thao tác lấy lịch sử trò chuyện diễn ra liên tục. Nếu không đánh Index, MongoDB sẽ phải quét toàn bộ hàng triệu tin nhắn để tìm.

**Cách tạo Index trực tiếp trên Compass:**
1. Trong database `chatapp`, bấm chọn collection **`messages`**.
2. Chọn tab **Indexes** (nằm cạnh tab *Documents*, *Aggregations*).
3. Bấm nút **Create Index**.
4. Cấu hình 2 trường như sau:
   * Trường 1: Tên trường điền `conversation_id`, kiểu chọn `1` (Ascending).
   * Bấm **Add field** để thêm trường 2:
   * Trường 2: Tên trường điền `created_at`, kiểu chọn `-1` (Descending - để lấy tin mới nhất trước).
5. Bấm nút **Create Index**.

> 💡 **Cách 2 (Dành cho bạn nào thích dùng lệnh):**
> Ở góc dưới cùng bên trái cửa sổ MongoDB Compass, click vào chữ **`_MONGOSH`** và gõ lệnh sau rồi ấn Enter:
> ```javascript
> use chatapp
> db.messages.createIndex({ conversation_id: 1, created_at: -1 })
> ```

---

<a name="cac-cau-lenh-thuong-dung--xu-ly-su-co-thuong-gap-faq"></a>
## Các câu lệnh thường dùng & Xử lý sự cố thường gặp (FAQ)

### 1. Bảng lệnh Docker hàng ngày
| Lệnh | Ý nghĩa |
| :--- | :--- |
| `docker compose up -d` | Bật toàn bộ Redis và MongoDB chạy ngầm |
| `docker compose stop` | Tạm dừng các container (tiết kiệm pin/RAM khi không code) |
| `docker compose start` | Tiếp tục chạy lại các container đã tạm dừng |
| `docker compose down` | Tắt và gỡ container (dữ liệu trong volume vẫn an toàn) |
| `docker compose logs -f` | Xem log trực tiếp của cả Redis và MongoDB để xem lỗi nếu có |
| `docker ps` | Xem danh sách các container đang chạy |

### 2. Sự cố thường gặp
* **Lỗi `port is already allocated` (Trùng cổng 6379 hoặc 27017):**
  * *Nguyên nhân:* Trước đó máy bạn đã từng cài Redis hoặc MongoDB trực tiếp vào Windows rồi.
  * *Cách xử lý:* Mở Start menu $\to$ gõ `Services` $\to$ tìm dịch vụ `MongoDB` hoặc `Redis` $\to$ click chuột phải chọn **Stop**. Sau đó chạy lại `docker compose up -d`.
* **Docker báo `Docker Desktop is stopping...` hoặc không chuyển sang màu xanh:**
  * *Cách xử lý:* Khởi động lại Docker Desktop hoặc vào Task Manager tắt hẳn tiến trình `Docker Desktop` rồi bật lại.

---

### 🎉 Chúc mừng bạn!
Bạn đã hoàn thành trọn vẹn **Giai đoạn 1**. Hạ tầng dữ liệu đã sẵn sàng 100% để chúng ta bước sang **Giai đoạn 2: Viết mã nguồn Backend Golang (Clean Architecture, REST API Auth & WebSocket Hub)**!