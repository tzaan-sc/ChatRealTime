# 📋 DANH SÁCH TÍNH NĂNG NÂNG CAO - HỆ THỐNG CHAT REALTIME
> **Dạng:** Bảng Checklist tính năng thuần túy (Chỉ ghi tên tính năng, dùng để theo dõi tiến độ phát triển).

---

## 🛡️ 1. BẢO MẬT & QUYỀN RIÊNG TƯ (SECURITY & PRIVACY)
- [ ] Mã hóa đầu cuối 1-1 (E2EE - Signal Protocol / Double Ratchet)
- [ ] Mã hóa đầu cuối nhóm chat (MLS - Messaging Layer Security)
- [ ] Tin nhắn tự hủy (Disappearing Messages - hẹn giờ tự xóa 5s, 1m, 24h)
- [ ] Cảnh báo / Chặn chụp màn hình (Screenshot Prevention & Detection)
- [ ] Đăng nhập nhanh bằng quét mã QR (QR Code Login)
- [ ] Xác thực 2 bước (2FA - TOTP Google Authenticator)
- [ ] Đăng nhập sinh trắc học / Không cần mật khẩu (Passkeys / WebAuthn)
- [ ] Quản lý phiên đăng nhập đa thiết bị (Active Sessions & Đăng xuất từ xa)
- [ ] Khóa ứng dụng bằng mã PIN / FaceID / Vân tay
- [ ] Chế độ trò chuyện bí mật không lưu máy chủ (Secret Chat)
- [ ] Ẩn trạng thái hoạt động / Ẩn thời gian online cuối cùng (Last Seen Privacy)
- [ ] Chặn người dùng và báo cáo tài khoản vi phạm (Block & Report)

---

## 💬 2. NHẮN TIN NÂNG CAO & SOẠN THẢO (ADVANCED MESSAGING)
- [ ] Lên lịch hẹn giờ gửi tin nhắn (Scheduled Messages)
- [ ] Gửi tin nhắn im lặng (Silent Messages - không gây chuông/rung phía người nhận)
- [ ] Ghim nhiều tin nhắn cùng lúc trong cuộc trò chuyện (Multi-Pinned Messages)
- [ ] Luồng thảo luận theo từng tin nhắn (Message Threads / In-Thread Replies)
- [x] Đánh dấu tin nhắn là chưa đọc (Mark as Unread)
- [ ] Tự động đồng bộ tin nhắn nháp giữa các thiết bị (Draft Message Sync)
- [ ] Định dạng văn bản phong phú (Markdown: Đậm, Nghiêng, Gạch chân, Khung Code, Ẩn Spoiler)
- [ ] Chuyển tiếp tin nhắn thông minh (Forward kèm hoặc bỏ tên người gửi gốc)
- [ ] Hộp thư lưu trữ cá nhân (Saved Messages / Cloud Notebook)
- [ ] Hẹn giờ nhắc việc từ tin nhắn (Set Reminder on Message)
- [ ] Giới hạn tốc độ gửi tin trong phòng chat (Slow Mode chống spam)

---

## 👥 3. NHÓM, KÊNH & CỘNG ĐỒNG (SPACES, CHANNELS & GROUPS)
- [ ] Không gian cộng đồng đa kênh (Spaces / Servers theo phong cách Discord)
- [ ] Phân chia danh mục kênh (Channel Categories & Text Channels)
- [ ] Kênh thông báo phát sóng 1 chiều (Broadcast / Announcement Channels)
- [ ] Hệ thống phân quyền vai trò chi tiết (RBAC: Quyền gửi bài, xóa tin, mời người, cấm chat)
- [ ] Tạo link mời tham gia nhóm kèm giới hạn số lượt / thời hạn (Invite Links with Expiry)
- [ ] Chế độ duyệt thành viên trước khi vào nhóm (Member Approval Requests)
- [ ] Khảo sát & Bình chọn trực tuyến thời gian thực (Interactive Polls & Quizzes)
- [ ] Sự kiện nhóm & Lịch hẹn chung (Group Events)
- [ ] Thống kê phân tích hoạt động nhóm (Group Analytics & Member Insights)

---

## 📹 4. HỘI NGHỊ TRUYỀN HÌNH & PHÒNG THOẠI (WEBRTC SFU)
- [ ] Cuộc gọi nhóm Video / Voice đa điểm quy mô lớn (WebRTC SFU - LiveKit / Pion)
- [ ] Kênh thoại mở tự do tham gia / rời phòng (Discord Voice Channels)
- [ ] Chia sẻ màn hình độ nét cao kèm âm thanh hệ thống (Screen Sharing 1080p 60fps)
- [ ] Tự động nhận diện người đang phát biểu để làm nổi bật khung hình (Active Speaker Detection)
- [ ] Tính năng Giơ tay phát biểu & Quyền Admin tắt mic toàn bộ (Raise Hand & Mute All)
- [ ] Lọc tiếng ồn môi trường thông minh (AI Noise Suppression - RNNoise)
- [ ] Làm mờ phông nền & Nền ảo khi bật Camera (Virtual Background)
- [ ] Ghi âm / Ghi hình cuộc họp (Call Recording)
- [ ] Chia phòng họp nhóm nhỏ (Breakout Rooms)

---

## 🤖 5. TRÍ TUỆ NHÂN TẠO (AI & SMART ASSISTANT)
- [ ] Tóm tắt nội dung cuộc trò chuyện nhóm dài tự động (AI Chat Summarizer)
- [ ] Gợi ý phản hồi nhanh dựa trên ngữ cảnh (Smart Reply)
- [ ] Dịch tin nhắn trực tiếp đa ngôn ngữ (Real-time Inline Translation)
- [ ] Chuyển đổi tin nhắn thoại Voice Note thành văn bản (Speech-to-Text / Whisper)
- [ ] Trợ lý AI đồng hành giải đáp trực tiếp trong nhóm (`@AI Assistant`)
- [ ] Tự động phát hiện và trích xuất đầu việc cần làm (Action Items & Task Extractor)
- [ ] Soạn thảo và sửa lỗi chính tả bằng AI (AI Writing Assistant)
- [ ] Tạo ảnh và Sticker bằng AI theo câu lệnh (`/imagine`)

---

## 🔍 6. TÌM KIẾM TOÀN VĂN & QUẢN LÝ DỮ LIỆU ĐA PHƯƠNG TIỆN (SEARCH & MEDIA)
- [ ] Tìm kiếm toàn văn siêu tốc có dấu / không dấu (Meilisearch Engine)
- [ ] Bộ lọc tìm kiếm nâng cao (theo người gửi, ngày tháng, loại tệp tin, liên kết)
- [ ] Thanh phím tắt tìm kiếm nhanh toàn ứng dụng `Ctrl + K` (Command Palette)
- [ ] Thư viện Media tập trung của hội thoại (Shared Media Gallery: Ảnh, Video, Tài liệu, Audio, Links)
- [ ] Xem trước liên kết giàu thông tin (Rich OpenGraph Preview - YouTube, Spotify, Báo chí)
- [ ] Trình phát Audio / Video tích hợp xem ngay trong ứng dụng (In-app Media Player)
- [ ] Xem trước tài liệu PDF trực tiếp không cần tải về máy (In-app PDF Viewer)
- [ ] Tự động nén ảnh sang WebP và tối ưu video trước khi gửi

---

## 🔌 7. NỀN TẢNG BOT, TỰ ĐỘNG HÓA & MINI-APPS (BOT ECOSYSTEM)
- [ ] Nền tảng tạo và quản lý Bot (Bot API & Token Generator)
- [ ] Menu lệnh gạch chéo tự động (Slash Commands: `/poll`, `/remind`, `/help`...)
- [ ] Thẻ tin nhắn tương tác (Interactive Action Buttons & Forms)
- [ ] Hệ thống Webhooks nhận thông báo từ bên ngoài (GitHub, Trello, Jira, Google Calendar)
- [ ] Nền tảng nhúng Webview Mini-Apps trong chat (Telegram WebApp style)
- [ ] Tính năng xem chung video cùng bạn bè (Watch Together)

---

## 🎨 8. TƯƠNG TÁC XÃ HỘI & CÁ NHÂN HÓA (SOCIAL & CUSTOMIZATION)
- [ ] Kho Sticker hoạt họa vector chuyển động mượt (Lottie Animated Stickers)
- [ ] Tự tạo và tải lên bộ Sticker / Custom Emoji riêng
- [ ] Thả biểu cảm cảm xúc tùy biến bằng toàn bộ Emoji (Full Custom Emoji Reactions)
- [ ] Trạng thái hoạt động phong phú (Rich Presence: "Đang nghe nhạc Spotify", "Đang chơi game")
- [ ] Trạng thái cảm xúc cá nhân kèm biểu tượng (Custom Status & Mood)
- [ ] Tùy biến hình nền khung chat riêng cho từng bạn chat (Custom Chat Wallpapers)
- [ ] Bộ theme giao diện phong phú (Dark Cyber, AMOLED Black, Light Pastel, High Contrast)
- [ ] Tùy chỉnh âm thanh thông báo cho từng người / từng nhóm

---

## ⚡ 9. TRẢI NGHIỆM ĐA NỀN TẢNG & OFFLINE-FIRST (CLIENT EXPERIENCE)
- [ ] Hoạt động Offline hoàn chỉnh (Local DB: IndexedDB / SQLite WASM)
- [ ] Hàng đợi gửi tin Outbox (tự động gửi lại khi có mạng mà không mất tin)
- [ ] Đồng bộ sai khác siêu tốc khi mở ứng dụng (Delta Sync - nạp trong < 0.2 giây)
- [ ] Hỗ trợ Progressive Web App (PWA) cài đặt như app gốc trên máy tính / điện thoại
- [ ] Thông báo đẩy nền khi tắt hoàn toàn trình duyệt (Web Push / Firebase Cloud Messaging)
- [ ] Giao diện gọi điện tương thích hệ thống ngoài màn hình khóa (CallKit iOS / ConnectionService Android)
- [ ] Đóng gói phiên bản ứng dụng Desktop siêu nhẹ (Tauri / Electron)

---

## 📊 10. QUẢN TRỊ, KIỂM DUYỆT & VẬN HÀNH HỆ THỐNG (ADMIN & DEVOPS)
- [ ] Bộ lọc tự động phát hiện ngôn từ độc hại và chống spam (AI Auto-Moderation)
- [ ] Tự động quét và làm mờ hình ảnh nhạy cảm (NSFW Content Filtering)
- [ ] Cổng bảng điều khiển dành cho Quản trị viên (Admin Portal & User Management)
- [ ] Nhật ký kiểm toán các thao tác quản trị (Audit Logs)
- [ ] Cụm WebSocket Gateway phân tán chạy đồng thời nhiều node (Horizontal Scaling)
- [ ] Hàng đợi tin nhắn chịu tải cao (NATS JetStream / Apache Kafka Message Broker)
- [ ] Hệ thống lưu trữ tệp đám mây phân tán chuẩn S3 (MinIO Object Storage Cluster)
- [ ] Giám sát sức khỏe hệ thống và cảnh báo sự cố 24/7 (Prometheus + Grafana APM)
