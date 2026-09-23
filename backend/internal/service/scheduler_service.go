package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"

	"github.com/redis/go-redis/v9"
)

type SchedulerService struct {
	redisClient   *redis.Client
	msgRepo       *repository.MessageRepository
	convRepo      *repository.ConversationRepository
	scheduledRepo *repository.ScheduledRepository
	reminderRepo  *repository.ReminderRepository
}

func NewSchedulerService(
	redisClient *redis.Client,
	msgRepo *repository.MessageRepository,
	convRepo *repository.ConversationRepository,
	scheduledRepo *repository.ScheduledRepository,
	reminderRepo *repository.ReminderRepository,
) *SchedulerService {
	return &SchedulerService{
		redisClient:   redisClient,
		msgRepo:       msgRepo,
		convRepo:      convRepo,
		scheduledRepo: scheduledRepo,
		reminderRepo:  reminderRepo,
	}
}

// Start bắt đầu goroutine quét tin hẹn giờ và nhắc việc mỗi 5 giây
func (s *SchedulerService) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.processScheduledMessages(ctx, now)
				s.processReminders(ctx, now)
			}
		}
	}()
	log.Println("⏰ Scheduler Service đã khởi chạy (quét mỗi 5s)...")
}

func (s *SchedulerService) processScheduledMessages(ctx context.Context, now time.Time) {
	messages, err := s.scheduledRepo.GetPendingDueMessages(ctx, now)
	if err != nil || len(messages) == 0 {
		return
	}

	for _, sm := range messages {
		msgType := sm.Type
		if msgType == "" {
			msgType = "text"
		}

		newMsg := &models.Message{
			SenderID:     sm.SenderID,
			SenderName:   sm.SenderName,
			SenderAvatar: sm.SenderAvatar,
			Content:      sm.Content,
			Type:         msgType,
			FileName:     sm.FileName,
			FileSize:     sm.FileSize,
			ReplyTo:      sm.ReplyTo,
			IsSilent:     sm.IsSilent,
			CreatedAt:    time.Now(),
		}

		if sm.GroupID != "" {
			newMsg.GroupID = sm.GroupID
			newMsg.IsRead = true
			if err := s.msgRepo.Create(ctx, newMsg); err != nil {
				log.Printf("Lỗi lưu scheduled msg: %v", err)
				continue
			}

			eventReceive := models.WSEvent{
				Event:   "group:receive",
				Payload: newMsg,
			}
			eventBytes, _ := json.Marshal(eventReceive)
			groupChannel := fmt.Sprintf("group:chat:%s", sm.GroupID)
			s.redisClient.Publish(ctx, groupChannel, string(eventBytes))

		} else if sm.ReceiverID != "" {
			convID := sm.ConversationID
			if convID == "" {
				if sm.ReceiverID == sm.SenderID {
					convID = fmt.Sprintf("saved_%s", sm.SenderID)
				} else if sm.SenderID < sm.ReceiverID {
					convID = fmt.Sprintf("%s_%s", sm.SenderID, sm.ReceiverID)
				} else {
					convID = fmt.Sprintf("%s_%s", sm.ReceiverID, sm.SenderID)
				}
			}

			newMsg.ConversationID = convID
			newMsg.ReceiverID = sm.ReceiverID
			newMsg.IsRead = (sm.ReceiverID == sm.SenderID)

			if err := s.msgRepo.Create(ctx, newMsg); err != nil {
				log.Printf("Lỗi lưu scheduled 1-1 msg: %v", err)
				continue
			}

			_, _ = s.convRepo.GetOrCreate(ctx, convID, sm.SenderID, sm.ReceiverID)
			_ = s.convRepo.UpdateLastMessage(ctx, convID, sm.Content, sm.SenderID)

			eventReceive := models.WSEvent{
				Event:   "chat:receive",
				Payload: newMsg,
			}
			eventBytes, _ := json.Marshal(eventReceive)

			// Gửi cho người nhận
			if sm.ReceiverID != sm.SenderID {
				s.redisClient.Publish(ctx, fmt.Sprintf("user:chat:%s", sm.ReceiverID), string(eventBytes))
			}
			// Gửi cho người gửi để cập nhật giao diện
			s.redisClient.Publish(ctx, fmt.Sprintf("user:chat:%s", sm.SenderID), string(eventBytes))
		}

		_ = s.scheduledRepo.MarkAsSent(ctx, sm.ID)
		log.Printf("📨 Đã tự động gửi tin nhắn hẹn giờ: ID=%s", sm.ID.Hex())
	}
}

func (s *SchedulerService) processReminders(ctx context.Context, now time.Time) {
	reminders, err := s.reminderRepo.GetPendingDueReminders(ctx, now)
	if err != nil || len(reminders) == 0 {
		return
	}

	for _, r := range reminders {
		alertEvent := models.WSEvent{
			Event: "reminder:alert",
			Payload: map[string]interface{}{
				"id":              r.ID.Hex(),
				"snippet":         r.ContentSnippet,
				"message_id":      r.MessageID,
				"conversation_id": r.ConversationID,
				"group_id":        r.GroupID,
				"remind_at":       r.RemindAt,
			},
		}
		bytes, _ := json.Marshal(alertEvent)
		s.redisClient.Publish(ctx, fmt.Sprintf("user:chat:%s", r.UserID), string(bytes))
		_ = s.reminderRepo.MarkAsTriggered(ctx, r.ID)
		log.Printf("🔔 Đã phát nhắc hẹn cho user %s: %s", r.UserID, r.ContentSnippet)
	}
}
