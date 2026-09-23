package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Group đại diện cho một nhóm chat trong collection "groups"
type Group struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name            string               `bson:"name" json:"name"`
	Avatar          string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	CreatorID       primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	AdminIDs        []primitive.ObjectID `bson:"admin_ids" json:"admin_ids"`
	MemberIDs       []primitive.ObjectID `bson:"member_ids" json:"member_ids"`
	SlowModeSeconds int                  `bson:"slow_mode_seconds,omitempty" json:"slow_mode_seconds"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
}

// GroupMemberInfo chứa thông tin chi tiết một thành viên trong nhóm
type GroupMemberInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	IsAdmin     bool   `json:"is_admin"`
}

// GroupDetailResponse DTO trả về chi tiết nhóm kèm danh sách thành viên
type GroupDetailResponse struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Avatar          string            `json:"avatar"`
	CreatorID       string            `json:"creator_id"`
	Members         []GroupMemberInfo `json:"members"`
	SlowModeSeconds int               `json:"slow_mode_seconds"`
	CreatedAt       time.Time         `json:"created_at"`
}

// CreateGroupRequest DTO tạo nhóm mới
type CreateGroupRequest struct {
	Name      string   `json:"name" binding:"required"`
	MemberIDs []string `json:"member_ids"` // Danh sách ID bạn bè mời vào nhóm
}

// AddMemberRequest DTO thêm thành viên vào nhóm
type AddMemberRequest struct {
	MemberIDs []string `json:"member_ids" binding:"required"`
}
