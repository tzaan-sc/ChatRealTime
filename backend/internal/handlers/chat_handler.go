package handlers

import (
	"net/http"
	"strconv"

	"chatrealtime-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// GetConversations API lấy danh sách chat
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	convs, err := h.chatService.GetUserConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": convs})
}

// GetMessages API lấy lịch sử tin nhắn
func (h *ChatHandler) GetMessages(c *gin.Context) {
	convID := c.Param("conversation_id")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	messages, err := h.chatService.GetMessages(c.Request.Context(), convID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// MarkAsRead API đánh dấu đã xem
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	convID := c.Param("conversation_id")
	userID := c.GetString("user_id")

	if err := h.chatService.MarkMessagesAsRead(c.Request.Context(), convID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã đánh dấu đã đọc"})
}

// MarkAsUnread API đánh dấu chưa đọc
func (h *ChatHandler) MarkAsUnread(c *gin.Context) {
	convID := c.Param("conversation_id")
	userID := c.GetString("user_id")

	if err := h.chatService.MarkMessagesAsUnread(c.Request.Context(), convID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã đánh dấu chưa đọc"})
}

// GetUsers API lấy danh sách bạn bè để chọn chat
func (h *ChatHandler) GetUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	users, err := h.chatService.GetAllUsers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}
