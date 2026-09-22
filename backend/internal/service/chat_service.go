package service

import (
	"context"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"
)

type ChatService struct {
	msgRepo  *repository.MessageRepository
	convRepo *repository.ConversationRepository
	userRepo *repository.UserRepository
}

func NewChatService(msgRepo *repository.MessageRepository, convRepo *repository.ConversationRepository, userRepo *repository.UserRepository) *ChatService {
	return &ChatService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
		userRepo: userRepo,
	}
}

// GetUserConversations lấy danh sách hội thoại kèm thông tin đối phương và số tin chưa đọc
func (s *ChatService) GetUserConversations(ctx context.Context, currentUserID string) ([]models.ConversationResponse, error) {
	convs, err := s.convRepo.GetUserConversations(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	var responses []models.ConversationResponse
	for _, conv := range convs {
		// Tìm ID của người đối diện
		otherUserID := conv.Members[0]
		if otherUserID == currentUserID && len(conv.Members) > 1 {
			otherUserID = conv.Members[1]
		}

		otherUser, err := s.userRepo.FindByID(ctx, otherUserID)
		if err != nil || otherUser == nil {
			continue
		}

		unread, _ := s.msgRepo.CountUnread(ctx, conv.CustomID, currentUserID)

		responses = append(responses, models.ConversationResponse{
			Conversation: conv,
			OtherUser:    *otherUser,
			UnreadCount:  unread,
		})
	}

	return responses, nil
}

// GetMessages lấy lịch sử tin nhắn
func (s *ChatService) GetMessages(ctx context.Context, conversationID string, limit, offset int64) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.msgRepo.GetByConversation(ctx, conversationID, limit, offset)
}

// MarkMessagesAsRead đánh dấu tin nhắn là đã đọc
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, conversationID, currentUserID string) error {
	return s.msgRepo.MarkAsRead(ctx, conversationID, currentUserID)
}

// MarkMessagesAsUnread đánh dấu tin nhắn mới nhất là chưa đọc
func (s *ChatService) MarkMessagesAsUnread(ctx context.Context, conversationID, currentUserID string) error {
	return s.msgRepo.MarkAsUnread(ctx, conversationID, currentUserID)
}

// GetAllUsers lấy danh bạ để bắt đầu chat mới (trừ bản thân)
func (s *ChatService) GetAllUsers(ctx context.Context, currentUserID string) ([]models.User, error) {
	return s.userRepo.FindAllExcept(ctx, currentUserID)
}
