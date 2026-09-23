package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScheduledMessage đại diện cho tin nhắn được hẹn giờ gửi
type ScheduledMessage struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id,omitempty" json:"conversation_id,omitempty"`
	GroupID        string             `bson:"group_id,omitempty" json:"group_id,omitempty"`
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	SenderName     string             `bson:"sender_name,omitempty" json:"sender_name,omitempty"`
	SenderAvatar   string             `bson:"sender_avatar,omitempty" json:"sender_avatar,omitempty"`
	ReceiverID     string             `bson:"receiver_id,omitempty" json:"receiver_id,omitempty"`
	Content        string             `bson:"content" json:"content"`
	Type           string             `bson:"type" json:"type"` // "text", "image", "file", "voice"
	FileName       string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileSize       int64              `bson:"file_size,omitempty" json:"file_size,omitempty"`
	ReplyTo        *ReplySnippet      `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	IsSilent       bool               `bson:"is_silent" json:"is_silent"`
	ScheduledAt    time.Time          `bson:"scheduled_at" json:"scheduled_at"`
	Status         string             `bson:"status" json:"status"` // "pending", "sent", "cancelled"
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// CreateScheduledRequest DTO hẹn giờ gửi tin
type CreateScheduledRequest struct {
	ReceiverID  string        `json:"receiver_id,omitempty"`
	GroupID     string        `json:"group_id,omitempty"`
	Content     string        `json:"content" binding:"required"`
	Type        string        `json:"type"`
	FileName    string        `json:"file_name,omitempty"`
	FileSize    int64         `json:"file_size,omitempty"`
	ReplyTo     *ReplySnippet `json:"reply_to,omitempty"`
	IsSilent    bool          `json:"is_silent"`
	ScheduledAt time.Time     `json:"scheduled_at" binding:"required"`
}
