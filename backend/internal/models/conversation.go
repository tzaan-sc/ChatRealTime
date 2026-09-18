package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation đại diện cho 1 cuộc trò chuyện 1-1
type Conversation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomID      string             `bson:"custom_id" json:"custom_id"` // userA_userB (sort alphabet)
	Members       []string           `bson:"members" json:"members"`     // [userAID, userBID]
	LastMessage   string             `bson:"last_message" json:"last_message"`
	LastSenderID  string             `bson:"last_sender_id" json:"last_sender_id"`
	LastMessageAt time.Time          `bson:"last_message_at" json:"last_message_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// ConversationResponse DTO trả về cho Frontend danh sách chat
type ConversationResponse struct {
	Conversation Conversation `json:"conversation"`
	OtherUser    User         `json:"other_user"`
	UnreadCount  int64        `json:"unread_count"`
}
