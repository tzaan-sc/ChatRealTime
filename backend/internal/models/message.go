package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message đại diện cho 1 tin nhắn trong collection "messages"
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id" json:"conversation_id"`
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	ReceiverID     string             `bson:"receiver_id" json:"receiver_id"`
	Content        string             `bson:"content" json:"content"`
	Type           string             `bson:"type" json:"type"` // "text", "image", "file"
	IsRead         bool               `bson:"is_read" json:"is_read"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// SendMessageRequest DTO gửi từ client
type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       string `json:"type"` // Mặc định là "text"
}
