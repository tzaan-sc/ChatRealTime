package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Reaction đại diện cho một biểu tượng cảm xúc thả vào tin nhắn
type Reaction struct {
	UserID string `bson:"user_id" json:"user_id"`
	Emoji  string `bson:"emoji" json:"emoji"`
}

// ReplySnippet chứa tóm tắt tin nhắn gốc khi được trả lời
type ReplySnippet struct {
	MessageID  string `bson:"message_id" json:"message_id"`
	SenderName string `bson:"sender_name" json:"sender_name"`
	Content    string `bson:"content" json:"content"`
}

// ForwardSnippet chứa thông tin người gửi gốc khi tin nhắn được chuyển tiếp
type ForwardSnippet struct {
	OriginalSenderID   string `bson:"original_sender_id,omitempty" json:"original_sender_id,omitempty"`
	OriginalSenderName string `bson:"original_sender_name,omitempty" json:"original_sender_name,omitempty"`
	OriginalMessageID  string `bson:"original_message_id,omitempty" json:"original_message_id,omitempty"`
	IsAnonymous        bool   `bson:"is_anonymous" json:"is_anonymous"`
}

// Message đại diện cho 1 tin nhắn trong collection "messages"
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID string             `bson:"conversation_id,omitempty" json:"conversation_id,omitempty"` // Dùng cho chat 1-1
	GroupID        string             `bson:"group_id,omitempty" json:"group_id,omitempty"`               // Dùng cho chat nhóm
	SenderID       string             `bson:"sender_id" json:"sender_id"`
	SenderName     string             `bson:"sender_name,omitempty" json:"sender_name,omitempty"`
	SenderAvatar   string             `bson:"sender_avatar,omitempty" json:"sender_avatar,omitempty"`
	ReceiverID     string             `bson:"receiver_id,omitempty" json:"receiver_id,omitempty"`
	Content        string             `bson:"content" json:"content"` // Chứa text hoặc URL tĩnh của ảnh/file/voice
	Type           string             `bson:"type" json:"type"`       // "text", "image", "file", "voice"
	FileName       string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileSize       int64              `bson:"file_size,omitempty" json:"file_size,omitempty"`
	Reactions      []Reaction         `bson:"reactions,omitempty" json:"reactions,omitempty"`
	ReplyTo        *ReplySnippet      `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	IsSilent       bool               `bson:"is_silent" json:"is_silent"`
	IsPinned       bool               `bson:"is_pinned" json:"is_pinned"`
	PinnedAt       *time.Time         `bson:"pinned_at,omitempty" json:"pinned_at,omitempty"`
	PinnedBy       string             `bson:"pinned_by,omitempty" json:"pinned_by,omitempty"`
	ThreadRootID   string             `bson:"thread_root_id,omitempty" json:"thread_root_id,omitempty"`
	ThreadCount    int                `bson:"thread_count" json:"thread_count"`
	ForwardFrom    *ForwardSnippet    `bson:"forward_from,omitempty" json:"forward_from,omitempty"`
	IsDeleted      bool               `bson:"is_deleted" json:"is_deleted"`
	IsEdited       bool               `bson:"is_edited" json:"is_edited"`
	IsRead         bool               `bson:"is_read" json:"is_read"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// SendMessageRequest DTO gửi từ client
type SendMessageRequest struct {
	ReceiverID   string          `json:"receiver_id,omitempty"`
	GroupID      string          `json:"group_id,omitempty"`
	Content      string          `json:"content" binding:"required"`
	Type         string          `json:"type"` // "text", "image", "file", "voice"
	FileName     string          `json:"file_name,omitempty"`
	FileSize     int64           `json:"file_size,omitempty"`
	ReplyTo      *ReplySnippet   `json:"reply_to,omitempty"`
	IsSilent     bool            `json:"is_silent,omitempty"`
	ThreadRootID string          `json:"thread_root_id,omitempty"`
	ForwardFrom  *ForwardSnippet `json:"forward_from,omitempty"`
}


