# Hướng Dẫn Chi Tiết Giai Đoạn 6: Tin Nhắn Đa Phương Tiện (Hình Ảnh, Tệp Đính Kèm & Tin Nhắn Thoại Voice Note)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: Golang (Static Uploads & Multipart) + Svelte 5 (MediaRecorder & Lightbox) + MongoDB + Redis

Giai đoạn này sẽ đưa ứng dụng chat của bạn lên một tầm cao mới, biến phòng chat chữ đơn điệu thành một không gian đa phương tiện sống động:
1. 🖼️ **Gửi Hình Ảnh (Image Messaging):** Chọn file từ máy hoặc dán trực tiếp (`Ctrl + V`) $\to$ Xem trước $\to$ Click xem ảnh phóng to toàn màn hình (Image Lightbox).
2. 📎 **Gửi Tệp Đính Kèm (File Attachments):** Hỗ trợ gửi file Word, Excel, PDF, Zip kèm icon định dạng, dung lượng và nút tải về máy.
3. 🎙️ **Tin Nhắn Thoại (Voice Note / Audio Message):** Nhấn giữ micro trên giao diện web để thu âm giọng nói $\to$ Gửi đoạn voice kèm trình phát âm thanh (Play/Pause, thanh thời lượng) như Zalo/Telegram.

---

## 📋 Checklist Tiến Độ Giai Đoạn 6

- [ ] **1. Xây dựng Backend Upload API (Golang):**
  - [ ] Tạo thư mục `backend/uploads` và cấu hình static files `r.Static("/uploads", "./uploads")` trong Gin router.
  - [ ] Viết Handler `UploadHandler` tiếp nhận multipart file (`POST /api/upload`), kiểm tra định dạng và sinh tên file an toàn (UUID / Timestamp).
  - [ ] Cập nhật model `Message` trong MongoDB để hỗ trợ các thuộc tính: `type` ("image", "file", "voice"), `file_name`, `file_size`.
- [ ] **2. Tạo Component Phóng to Ảnh (Image Lightbox Modal):**
  - [ ] Viết `frontend/src/lib/components/ImageModal.svelte` cho phép phóng to ảnh sắc nét trên nền tối mờ.
- [ ] **3. Tạo Component Thu âm & Phát Voice Note (Voice Message Player):**
  - [ ] Sử dụng `navigator.mediaDevices.getUserMedia` và `MediaRecorder` API để ghi âm giọng nói.
  - [ ] Viết trình phát Audio Player với nút Play/Pause và thanh tiến trình thời gian.
- [ ] **4. Nâng cấp `ChatArea.svelte`:**
  - [ ] Nút kẹp ghim 📎 chọn ảnh hoặc tài liệu từ máy tính.
  - [ ] Hỗ trợ dán ảnh trực tiếp từ bộ nhớ tạm (`Ctrl + V` clipboard paste).
  - [ ] Nút micro 🎙️ ghi âm giọng nói.
  - [ ] Render trực quan các bong bóng tin nhắn:
    - Ảnh: Hiển thị thumbnail bo góc, click để phóng to.
    - File: Hiển thị icon tệp, tên file, dung lượng (KB/MB) và nút tải về.
    - Voice: Hiển thị thanh nghe audio với nút Play/Pause.
- [ ] **5. Kiểm thử toàn diện 6 kịch bản đa phương tiện (Ảnh, PDF, Voice, Lightbox, Tải về)**.

---

## Mục lục
1. [Bước 1: Cập nhật Backend Go (API Upload & Static Serving)](#buoc-1-backend-upload)
2. [Bước 2: Nâng cấp Model Tin nhắn trong Backend](#buoc-2-update-models)
3. [Bước 3: Tạo Component Xem Ảnh Phóng To (`ImageModal.svelte`)](#buoc-3-image-modal)
4. [Bước 4: Nâng cấp Giao diện Khung Chat (`ChatArea.svelte`)](#buoc-4-chat-area-media)
5. [Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết](#buoc-5-kiem-thu-thuc-te)

---

<a name="buoc-1-backend-upload"></a>
## Bước 1: Cập nhật Backend Go (API Upload & Static Serving)

### 1.1. Tạo thư mục chứa file tải lên
Tại thư mục `backend`, tạo thư mục `uploads`:
```powershell
New-Item -ItemType Directory -Force -Path "d:\GIT\ChatRealTime\backend\uploads"
```

### 1.2. Viết file `backend/internal/handlers/upload_handler.go`
Tạo file [backend/internal/handlers/upload_handler.go](file:///d:/GIT/ChatRealTime/backend/internal/handlers/upload_handler.go):

```go
package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadFile tiếp nhận multipart form data, lưu trữ file vào thư mục uploads
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file gửi lên"})
		return
	}

	// Giới hạn dung lượng tối đa 20MB
	if file.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File vượt quá giới hạn 20MB"})
		return
	}

	// Tạo tên file duy nhất tránh trùng lặp
	ext := filepath.Ext(file.Filename)
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.ReplaceAll(filepath.Base(file.Filename), " ", "_"), "")
	if ext == "" {
		uniqueName += ".webm" // Mặc định cho file ghi âm voice
	}

	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	savePath := filepath.Join(uploadDir, uniqueName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu file trên máy chủ: " + err.Error()})
		return
	}

	// Trả về URL để truy cập file tĩnh
	fileURL := fmt.Sprintf("http://localhost:8080/uploads/%s", uniqueName)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Tải file thành công",
		"file_url":  fileURL,
		"file_name": file.Filename,
		"file_size": file.Size,
	})
}
```

---

<a name="buoc-2-update-models"></a>
## Bước 2: Nâng cấp Model Tin nhắn trong Backend

### 2.1. Bổ sung `file_name` và `file_size` vào `models/message.go`
Mở file [backend/internal/models/message.go](file:///d:/GIT/ChatRealTime/backend/internal/models/message.go) và cập nhật:

```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id" json:"conversation_id"`
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	ReceiverID     string             `bson:"receiver_id" json:"receiver_id"`
	Content        string             `bson:"content" json:"content"` // Chứa text hoặc URL của ảnh/file/voice
	Type           string             `bson:"type" json:"type"`       // "text", "image", "file", "voice"
	FileName       string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileSize       int64              `bson:"file_size,omitempty" json:"file_size,omitempty"`
	IsRead         bool               `bson:"is_read" json:"is_read"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       string `json:"type"` // "text", "image", "file", "voice"
	FileName   string `json:"file_name,omitempty"`
	FileSize   int64  `json:"file_size,omitempty"`
}
```

### 2.2. Đăng ký Static route và Upload route trong `cmd/server/main.go`
Thêm `r.Static("/uploads", "./uploads")` và `api.POST("/upload", uploadHandler.UploadFile)`.

---

<a name="buoc-3-image-modal"></a>
## Bước 3: Tạo Component Xem Ảnh Phóng To (`ImageModal.svelte`)

Tạo file [frontend/src/lib/components/ImageModal.svelte](file:///d:/GIT/ChatRealTime/frontend/src/lib/components/ImageModal.svelte):

```svelte
<script>
  import { X, Download } from 'lucide-svelte';

  export let imageUrl = '';
  export let onClose = () => {};
</script>

{#if imageUrl}
  <div class="lightbox-overlay" on:click={onClose}>
    <div class="lightbox-content" on:click|stopPropagation>
      <div class="lightbox-actions">
        <a href={imageUrl} download="image" target="_blank" class="action-btn" title="Tải ảnh gốc">
          <Download size={20} />
        </a>
        <button class="action-btn" on:click={onClose} title="Đóng">
          <X size={20} />
        </button>
      </div>
      <img src={imageUrl} alt="phóng to" class="lightbox-image" />
    </div>
  </div>
{/if}

<style>
  .lightbox-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.88);
    backdrop-filter: blur(12px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    animation: fadeIn 0.2s ease;
  }
  .lightbox-content {
    position: relative;
    max-width: 90vw;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .lightbox-actions {
    position: absolute;
    top: -45px;
    right: 0;
    display: flex;
    gap: 12px;
  }
  .action-btn {
    background: rgba(255, 255, 255, 0.15);
    border: none;
    color: #fff;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    text-decoration: none;
    transition: 0.2s;
  }
  .action-btn:hover { background: rgba(255, 255, 255, 0.3); transform: scale(1.08); }
  .lightbox-image {
    max-width: 100%;
    max-height: 85vh;
    border-radius: var(--radius-md);
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6);
    object-fit: contain;
  }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
```

---

<a name="buoc-4-chat-area-media"></a>
## Bước 4: Nâng cấp Giao diện Khung Chat (`ChatArea.svelte`)

Tại [ChatArea.svelte](file:///d:/GIT/ChatRealTime/frontend/src/lib/components/ChatArea.svelte), tích hợp:
1. **Upload File & Ảnh:**
   - Thẻ `<input type="file" bind:this={fileInput} on:change={handleFileUpload} />` ẩn.
   - Khi chọn file $\to$ gửi `POST /api/upload` $\to$ nhận `file_url` $\to$ gửi WebSocket `chat:send` với type `"image"` hoặc `"file"`.
2. **Ghi âm Voice Message:**
   - Dùng `navigator.mediaDevices.getUserMedia({ audio: true })`.
   - Nút Micro: Click để bắt đầu ghi âm (hiện sóng âm và số giây `00:04...`), click nút gửi hoặc dừng để nạp audio Blob lên server.
3. **Hiển thị Tin nhắn:**
   - **Ảnh (`type === "image"`):** Render thẻ `<img>` với hiệu ứng zoom khi rê chuột, bấm vào mở `ImageModal`.
   - **File (`type === "file"`):** Render thẻ file kèm icon tài liệu, tên file, kích thước dung lượng (KB/MB) và nút tải xuống.
   - **Voice (`type === "voice"`):** Render player với thẻ `<audio controls>` tùy biến tông màu Cyber Dark.

---

<a name="buoc-5-kiem-thu-thuc-te"></a>
## Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết

| Kịch bản | Thao tác thực hiện | Kết quả mong đợi |
| :--- | :--- | :--- |
| 📸 **Gửi ảnh & Lightbox** | Click nút kẹp ghim 📎 chọn 1 bức ảnh (PNG/JPG) | Ảnh được upload và hiển thị ngay trên màn hình cả 2 bên. Click vào ảnh $\to$ bung modal phóng to sắc nét! |
| 📋 **Dán ảnh (Ctrl + V)** | Copy 1 hình ảnh từ web hoặc chụp màn hình (`Win + Shift + S`), bấm `Ctrl + V` vào ô chat | Ảnh tự động upload và gửi đi tức thì mà không cần lưu ra desktop! |
| 📎 **Gửi tài liệu (PDF/Zip)** | Click 📎 chọn file PDF hoặc Zip bất kỳ | Cả 2 bên hiện bong bóng file với icon tài liệu, dung lượng chính xác và nút bấm tải về máy. |
| 🎙️ **Thu âm giọng nói (Voice)** | Click nút Micro 🎙️, cho phép quyền Microphone $\to$ nói "Alo alo 1 2 3 4" $\to$ bấm nút Gửi | Đáy khung chat của đối phương hiện tin nhắn voice dạng thanh phát audio, bấm nút Play để nghe rõ giọng nói! |

---

### 🎉 Hoàn thành Giai đoạn 6!
Ứng dụng chat của bạn đã trở thành một hệ thống **Đa phương tiện hoàn chỉnh (Full Media Messaging)**!
