# Hướng Dẫn Chi Tiết Giai Đoạn 6: Đóng Gói Docker, Test Tải & Triển Khai (Deployment)
> **Dành cho người mới bắt đầu** | Hệ thống: Docker Multi-Stage | Nền tảng: Local & Cloud (Render / VPS)

Tài liệu này sẽ hướng dẫn bạn đóng gói toàn bộ hệ thống (Frontend Svelte + Backend Go + Redis + MongoDB) thành các Docker container chuẩn hóa, chạy 1-click bằng `docker compose` và sẵn sàng deploy lên môi trường Production trên Internet.

---

## 📋 Checklist Tiến Độ Giai Đoạn 6

- [ ] **1. Viết Dockerfile tối ưu cho Backend Go:**
  - Multi-stage build với Alpine Linux (Binary siêu nhẹ ~15MB).
  - `backend/Dockerfile`
- [ ] **2. Viết Dockerfile tối ưu cho Frontend Svelte:**
  - Multi-stage build với Nginx Alpine (File tĩnh siêu nhanh).
  - `frontend/Dockerfile` & `frontend/nginx.conf`
- [ ] **3. Cập nhật `docker-compose.yml` chạy 1-click toàn bộ hệ sinh thái:**
  - Gồm 4 dịch vụ: `mongo`, `redis`, `backend`, `frontend`.
- [ ] **4. Kiểm thử tính ổn định & Kịch bản mất mạng (Resilience Test):**
  - Test F5 / Tắt mạng đột ngột $\to$ Dữ liệu không bao giờ mất.
- [ ] **5. Hướng dẫn Deploy lên Cloud miễn phí / VPS giá rẻ.**

---

## Mục lục
1. [Bước 1: Dockerfile cho Backend Golang](#buoc-1-dockerfile-backend-golang)
2. [Bước 2: Dockerfile & Cấu hình Nginx cho Frontend Svelte](#buoc-2-dockerfile-frontend-svelte)
3. [Bước 3: File `docker-compose.yml` trọn gói](#buoc-3-docker-compose-tron-goi)
4. [Bước 4: Chạy toàn bộ hệ thống 1-click](#buoc-4-chay-he-thong-1-click)
5. [Bước 5: Kịch bản Kiểm thử Tải & Độ Ổn Định](#buoc-5-kiem-thu-do-on-dinh)
6. [Bước 6: Hướng dẫn Deploy lên Cloud (VPS / Render / MongoDB Atlas)](#buoc-6-deploy-cloud)

---

<a name="buoc-1-dockerfile-backend-golang"></a>
## Bước 1: Dockerfile cho Backend Golang

Tạo file `backend/Dockerfile`:

```dockerfile
# Stage 1: Build binary bằng Golang Compiler
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy mã nguồn và build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Stage 2: Chạy binary trên Scratch/Alpine siêu nhẹ (chỉ ~15MB)
FROM alpine:3.19

WORKDIR /app

# Cài đặt chứng chỉ SSL cho kết nối HTTPS
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server .
COPY .env .env

EXPOSE 8080

CMD ["./server"]
```

---

<a name="buoc-2-dockerfile-frontend-svelte"></a>
## Bước 2: Dockerfile & Cấu hình Nginx cho Frontend Svelte

### 2.1. File `frontend/nginx.conf`
```nginx
server {
    listen 80;
    server_name localhost;

    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
        try_files $uri $uri/ /index.html;
    }

    # Proxy các API và WebSocket tới Backend Go
    location /api/ {
        proxy_pass http://backend:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws {
        proxy_pass http://backend:8080/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }
}
```

### 2.2. File `frontend/Dockerfile`
```dockerfile
# Stage 1: Build mã nguồn Svelte
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# Stage 2: Phục vụ file tĩnh qua Web Server Nginx Alpine
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

---

<a name="buoc-3-docker-compose-tron-goi"></a>
## Bước 3: File `docker-compose.yml` trọn gói

Cập nhật lại file `d:\GIT\ChatRealTime\docker-compose.yml` ở thư mục gốc:

```yaml
version: '3.8'

services:
  # 1. Cơ sở dữ liệu Redis
  redis:
    image: redis:7
    container_name: redis
    restart: always
    ports:
      - "6379:6379"

  # 2. Cơ sở dữ liệu MongoDB
  mongo:
    image: mongo:7
    container_name: mongo
    restart: always
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password123
    volumes:
      - mongo_data:/data/db

  # 3. Backend Go Realtime Service
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: backend
    restart: always
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      MONGO_URI: mongodb://admin:password123@mongo:27017
      DB_NAME: chatapp
      REDIS_ADDR: redis:6379
      JWT_SECRET: super_secret_key_chatapp_2026
    depends_on:
      - redis
      - mongo

  # 4. Frontend Svelte Web App
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: frontend
    restart: always
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  mongo_data:
```

---

<a name="buoc-4-chay-he-thong-1-click"></a>
## Bước 4: Chạy toàn bộ hệ thống 1-click

Chỉ với **duy nhất 1 câu lệnh** tại thư mục gốc `d:\GIT\ChatRealTime`:

```powershell
docker compose up -d --build
```

- Toàn bộ 4 dịch vụ (Redis, Mongo, Go, Svelte) sẽ tự động được build và khởi chạy cùng lúc!
- Truy cập ngay trình duyệt tại địa chỉ: **`http://localhost`** để sử dụng ứng dụng chat!

---

<a name="buoc-5-kiem-thu-do-on-dinh"></a>
## Bước 5: Kịch bản Kiểm thử Tải & Độ Ổn Định

| Kịch bản kiểm thử | Hành động thực hiện | Kết quả mong đợi |
| :--- | :--- | :--- |
| **F5 / Tải lại trang** | Nhấn `Ctrl + F5` khi đang mở khung chat | Hệ thống tự động reconnect WebSocket, nạp lại toàn bộ lịch sử tin nhắn đầy đủ. |
| **Offline Messages** | Bob tắt trình duyệt, Alex gửi 3 tin nhắn, sau đó Bob mở lại | 3 tin nhắn của Alex đã được lưu trong MongoDB và hiển thị đầy đủ cho Bob kèm số lượng badge tin chưa đọc. |
| **Độ trễ truyền tin** | Gửi tin nhắn giữa 2 máy khác nhau | Độ trễ dưới 20ms nhờ Redis Pub/Sub và kết nối liên tục qua Gorilla WebSocket. |

---

<a name="buoc-6-deploy-cloud"></a>
## Bước 6: Hướng dẫn Deploy lên Cloud

### 1. Database Cloud Miễn Phí:
- **MongoDB:** Tạo cluster M0 Free vĩnh viễn trên [MongoDB Atlas](https://www.mongodb.com/atlas).
- **Redis:** Tạo cơ sở dữ liệu miễn phí trên [Upstash Redis](https://upstash.com).

### 2. Triển khai Server (VPS / Cloud):
- Thuê VPS Ubuntu giá rẻ (~$3 - $5/tháng trên DigitalOcean, Linode, Vultr hoặc Hetzner).
- Cài đặt Docker trên VPS: `curl -fsSL https://get.docker.com | sh`.
- Clone mã nguồn về VPS và gõ lệnh: `docker compose up -d --build`.
- Trỏ tên miền (Domain) và cấu hình SSL HTTPS miễn phí qua **Let's Encrypt / Certbot**.

---

### 🏆 Chúc mừng bạn đã hoàn thành toàn bộ Lộ trình Dự án!
Bạn đã xây dựng thành công một hệ thống **Direct Messaging Chat Realtime chuẩn kiến trúc doanh nghiệp** từ con số 0!
