package handlers

import (
	"net/http"
	"strconv"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatHandler struct {
	chatService   *service.ChatService
	msgRepo       *repository.MessageRepository
	scheduledRepo *repository.ScheduledRepository
	reminderRepo  *repository.ReminderRepository
	draftRepo     *repository.DraftRepository
	userRepo      *repository.UserRepository
}

func NewChatHandler(
	chatService *service.ChatService,
	msgRepo *repository.MessageRepository,
	scheduledRepo *repository.ScheduledRepository,
	reminderRepo *repository.ReminderRepository,
	draftRepo *repository.DraftRepository,
	userRepo *repository.UserRepository,
) *ChatHandler {
	return &ChatHandler{
		chatService:   chatService,
		msgRepo:       msgRepo,
		scheduledRepo: scheduledRepo,
		reminderRepo:  reminderRepo,
		draftRepo:     draftRepo,
		userRepo:      userRepo,
	}
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

// GetPinnedMessages API lấy danh sách tin nhắn đã ghim
func (h *ChatHandler) GetPinnedMessages(c *gin.Context) {
	targetID := c.Param("target_id")
	isGroup := c.Query("is_group") == "true"

	msgs, err := h.msgRepo.GetPinnedMessages(c.Request.Context(), targetID, isGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": msgs})
}

// GetThreadMessages API lấy các phản hồi trong luồng
func (h *ChatHandler) GetThreadMessages(c *gin.Context) {
	rootID := c.Param("id")
	msgs, err := h.msgRepo.GetThreadMessages(c.Request.Context(), rootID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": msgs})
}

// CreateScheduledMessage API hẹn giờ gửi tin nhắn
func (h *ChatHandler) CreateScheduledMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateScheduledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderName := ""
	senderAvatar := ""
	if h.userRepo != nil {
		u, _ := h.userRepo.FindByID(c.Request.Context(), userID)
		if u != nil {
			senderName = u.DisplayName
			if senderName == "" {
				senderName = u.Username
			}
			senderAvatar = u.AvatarURL
		}
	}

	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	scheduled := &models.ScheduledMessage{
		SenderID:     userID,
		SenderName:   senderName,
		SenderAvatar: senderAvatar,
		ReceiverID:   req.ReceiverID,
		GroupID:      req.GroupID,
		Content:      req.Content,
		Type:         msgType,
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		ReplyTo:      req.ReplyTo,
		IsSilent:     req.IsSilent,
		ScheduledAt:  req.ScheduledAt,
	}

	if err := h.scheduledRepo.Create(c.Request.Context(), scheduled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": scheduled})
}

// GetScheduledMessages API lấy danh sách tin đã hẹn giờ
func (h *ChatHandler) GetScheduledMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	targetID := c.Query("target_id")
	isGroup := c.Query("is_group") == "true"

	list, err := h.scheduledRepo.GetUserScheduledMessages(c.Request.Context(), userID, targetID, isGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// CancelScheduledMessage API hủy hẹn giờ gửi tin
func (h *ChatHandler) CancelScheduledMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}
	userID := c.GetString("user_id")

	if err := h.scheduledRepo.Cancel(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã hủy lịch gửi tin"})
}

// CreateReminder API tạo hẹn giờ nhắc việc từ tin nhắn
func (h *ChatHandler) CreateReminder(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rem := &models.Reminder{
		UserID:         userID,
		MessageID:      req.MessageID,
		ConversationID: req.ConversationID,
		GroupID:        req.GroupID,
		ContentSnippet: req.ContentSnippet,
		RemindAt:       req.RemindAt,
	}

	if err := h.reminderRepo.Create(c.Request.Context(), rem); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rem})
}

// GetReminders API lấy danh sách nhắc việc đang chờ
func (h *ChatHandler) GetReminders(c *gin.Context) {
	userID := c.GetString("user_id")
	list, err := h.reminderRepo.GetUserReminders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// DismissReminder API bỏ qua hoặc hủy nhắc việc
func (h *ChatHandler) DismissReminder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}
	userID := c.GetString("user_id")

	if err := h.reminderRepo.Dismiss(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã tắt nhắc việc"})
}

// GetDrafts API lấy tất cả tin nhắn nháp của user
func (h *ChatHandler) GetDrafts(c *gin.Context) {
	userID := c.GetString("user_id")
	list, err := h.draftRepo.GetUserDrafts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// SaveDraft API lưu nháp tin nhắn
func (h *ChatHandler) SaveDraft(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.draftRepo.Upsert(c.Request.Context(), userID, req.TargetID, req.TargetType, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã lưu nháp"})
}

// DeleteDraft API xóa nháp tin nhắn
func (h *ChatHandler) DeleteDraft(c *gin.Context) {
	userID := c.GetString("user_id")
	targetID := c.Param("target_id")

	if err := h.draftRepo.Delete(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa nháp"})
}
