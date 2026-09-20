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

// UploadFile tiếp nhận multipart file upload và lưu trữ tĩnh
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file gửi lên: " + err.Error()})
		return
	}

	// Giới hạn dung lượng 25MB
	const maxFileSize = 25 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File vượt quá giới hạn 25MB"})
		return
	}

	// Làm sạch và tạo tên file duy nhất
	ext := strings.ToLower(filepath.Ext(file.Filename))
	cleanBase := strings.ReplaceAll(filepath.Base(file.Filename), " ", "_")
	// Xóa ký tự đặc biệt có thể gây nguy hiểm
	cleanBase = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, cleanBase)

	if ext == "" {
		// Mặc định cho âm thanh ghi âm voice
		ext = ".webm"
		cleanBase += ext
	}

	uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), cleanBase)
	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, 0755)

	savePath := filepath.Join(uploadDir, uniqueName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu file trên máy chủ: " + err.Error()})
		return
	}

	// Xác định loại media
	mediaType := "file"
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		mediaType = "image"
	case ".webm", ".ogg", ".mp3", ".wav", ".m4a":
		mediaType = "voice"
	}

	// Sinh URL tĩnh truy cập file
	// Sử dụng scheme và host từ request hoặc localhost:8080 mặc định
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost:8080"
	}
	fileURL := fmt.Sprintf("%s://%s/uploads/%s", scheme, host, uniqueName)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Tải file thành công",
		"file_url":  fileURL,
		"file_name": file.Filename,
		"file_size": file.Size,
		"type":      mediaType,
	})
}
