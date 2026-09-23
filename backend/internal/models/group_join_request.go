package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupJoinRequest đại diện cho đơn xin tham gia nhóm trong collection "group_join_requests"
type GroupJoinRequest struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	GroupID    primitive.ObjectID  `bson:"group_id" json:"group_id"`
	UserID     primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Status     string              `bson:"status" json:"status"` // "pending", "approved", "rejected"
	Note       string              `bson:"note,omitempty" json:"note,omitempty"`
	CreatedAt  time.Time           `bson:"created_at" json:"created_at"`
	ReviewedAt *time.Time          `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	ReviewedBy *primitive.ObjectID `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
}

// JoinRequestDetail DTO trả về danh sách đơn chờ duyệt kèm thông tin thành viên
type JoinRequestDetail struct {
	ID        string          `json:"id"`
	GroupID   string          `json:"group_id"`
	User      GroupMemberInfo `json:"user"`
	Status    string          `json:"status"`
	Note      string          `json:"note"`
	CreatedAt time.Time       `json:"created_at"`
}
