package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupInvite đại diện cho một liên kết mời vào nhóm trong collection "group_invites"
type GroupInvite struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code      string             `bson:"code" json:"code"` // Mã mời duy nhất (VD: "inv_abc123")
	GroupID   primitive.ObjectID `bson:"group_id" json:"group_id"`
	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`
	MaxUses   int                `bson:"max_uses" json:"max_uses"` // 0 = không giới hạn
	UsesCount int                `bson:"uses_count" json:"uses_count"`
	ExpiresAt *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"` // nil = vĩnh viễn
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// CreateInviteRequest DTO tạo mã mời mới
type CreateInviteRequest struct {
	MaxUses     int `json:"max_uses"`     // 0, 1, 5, 25, 100
	ExpireHours int `json:"expire_hours"` // 0 = vĩnh viễn, 1, 24, 168 (7 ngày)
}

// InvitePreviewResponse DTO xem trước thông tin nhóm khi mở link mời
type InvitePreviewResponse struct {
	Code            string     `json:"code"`
	GroupID         string     `json:"group_id"`
	GroupName       string     `json:"group_name"`
	GroupAvatar     string     `json:"group_avatar"`
	MemberCount     int        `json:"member_count"`
	RequireApproval bool       `json:"require_approval"`
	InviterName     string     `json:"inviter_name"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IsExpired       bool       `json:"is_expired"`
	IsMaxedOut      bool       `json:"is_maxed_out"`
	IsAlreadyMember bool       `json:"is_already_member"`
	HasPendingReq   bool       `json:"has_pending_req"`
}
