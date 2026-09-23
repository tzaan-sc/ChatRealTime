package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Draft đại diện cho tin nhắn nháp được lưu và đồng bộ đa thiết bị
type Draft struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	TargetID   string             `bson:"target_id" json:"target_id"`     // Custom conversation_id hoặc group_id
	TargetType string             `bson:"target_type" json:"target_type"` // "direct" hoặc "group"
	Content    string             `bson:"content" json:"content"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// SaveDraftRequest DTO lưu nháp
type SaveDraftRequest struct {
	TargetID   string `json:"target_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required"`
	Content    string `json:"content"`
}
