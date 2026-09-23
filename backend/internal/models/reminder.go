package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Reminder đại diện cho một lịch nhắc việc gắn với tin nhắn hoặc hội thoại
type Reminder struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         string             `bson:"user_id" json:"user_id"`
	MessageID      string             `bson:"message_id,omitempty" json:"message_id,omitempty"`
	ConversationID string             `bson:"conversation_id,omitempty" json:"conversation_id,omitempty"`
	GroupID        string             `bson:"group_id,omitempty" json:"group_id,omitempty"`
	ContentSnippet string             `bson:"content_snippet" json:"content_snippet"`
	RemindAt       time.Time          `bson:"remind_at" json:"remind_at"`
	Status         string             `bson:"status" json:"status"` // "pending", "triggered", "dismissed"
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// CreateReminderRequest DTO tạo nhắc việc
type CreateReminderRequest struct {
	MessageID      string    `json:"message_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	GroupID        string    `json:"group_id,omitempty"`
	ContentSnippet string    `json:"content_snippet" binding:"required"`
	RemindAt       time.Time `json:"remind_at" binding:"required"`
}
