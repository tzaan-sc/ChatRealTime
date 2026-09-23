package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChannelType định nghĩa loại kênh
type ChannelType string

const (
	ChannelTypeText         ChannelType = "text"
	ChannelTypeAnnouncement ChannelType = "announcement"
)

// Channel đại diện cho một kênh chat con trong nhóm/không gian cộng đồng
type Channel struct {
	ID          string      `bson:"id" json:"id"`
	Name        string      `bson:"name" json:"name"`
	Description string      `bson:"description,omitempty" json:"description,omitempty"`
	Type        ChannelType `bson:"type" json:"type"` // "text" hoặc "announcement"
	CategoryID  string      `bson:"category_id,omitempty" json:"category_id,omitempty"`
	CreatedAt   time.Time   `bson:"created_at" json:"created_at"`
}

// ChannelCategory định nghĩa danh mục phân chia kênh (ví dụ: "CHUNG", "KỸ THUẬT")
type ChannelCategory struct {
	ID        string    `bson:"id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// Group đại diện cho một nhóm chat hoặc không gian cộng đồng trong collection "groups"
type Group struct {
	ID              primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	Name            string                      `bson:"name" json:"name"`
	Avatar          string                      `bson:"avatar,omitempty" json:"avatar,omitempty"`
	CreatorID       primitive.ObjectID          `bson:"creator_id" json:"creator_id"`
	AdminIDs        []primitive.ObjectID        `bson:"admin_ids" json:"admin_ids"`
	ModeratorIDs    []primitive.ObjectID        `bson:"moderator_ids,omitempty" json:"moderator_ids,omitempty"`
	MemberIDs       []primitive.ObjectID        `bson:"member_ids" json:"member_ids"`
	SlowModeSeconds int                         `bson:"slow_mode_seconds,omitempty" json:"slow_mode_seconds"`
	RequireApproval bool                        `bson:"require_approval" json:"require_approval"`
	MutedMembers    map[string]time.Time        `bson:"muted_members,omitempty" json:"muted_members,omitempty"` // userID hex -> muted_until
	IsCommunity     bool                        `bson:"is_community" json:"is_community"`
	Categories      []ChannelCategory           `bson:"categories,omitempty" json:"categories,omitempty"`
	Channels        []Channel                   `bson:"channels,omitempty" json:"channels,omitempty"`
	CreatedAt       time.Time                   `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time                   `bson:"updated_at" json:"updated_at"`
}

// GroupMemberInfo chứa thông tin chi tiết một thành viên trong nhóm
type GroupMemberInfo struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Role        string     `json:"role"` // "owner", "admin", "moderator", "member"
	IsAdmin     bool       `json:"is_admin"`
	IsMuted     bool       `json:"is_muted"`
	MutedUntil  *time.Time `json:"muted_until,omitempty"`
}

// GroupDetailResponse DTO trả về chi tiết nhóm kèm danh sách thành viên và kênh
type GroupDetailResponse struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Avatar          string            `json:"avatar"`
	CreatorID       string            `json:"creator_id"`
	Members         []GroupMemberInfo `json:"members"`
	SlowModeSeconds int               `json:"slow_mode_seconds"`
	RequireApproval bool              `json:"require_approval"`
	IsCommunity     bool              `json:"is_community"`
	Categories      []ChannelCategory `json:"categories"`
	Channels        []Channel         `json:"channels"`
	CreatedAt       time.Time         `json:"created_at"`
}

// CreateGroupRequest DTO tạo nhóm mới
type CreateGroupRequest struct {
	Name            string   `json:"name" binding:"required"`
	MemberIDs       []string `json:"member_ids"` // Danh sách ID bạn bè mời vào nhóm
	IsCommunity     bool     `json:"is_community"`
	RequireApproval bool     `json:"require_approval"`
}

// AddMemberRequest DTO thêm thành viên vào nhóm
type AddMemberRequest struct {
	MemberIDs []string `json:"member_ids" binding:"required"`
}

// CreateCategoryRequest DTO tạo danh mục kênh mới
type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateChannelRequest DTO tạo kênh mới
type CreateChannelRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	Type        ChannelType `json:"type"` // "text" hoặc "announcement"
	CategoryID  string      `json:"category_id"`
}

// UpdateMemberRoleRequest DTO cập nhật vai trò thành viên
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"` // "admin", "moderator", "member"
}

// MuteMemberRequest DTO cấm chat thành viên
type MuteMemberRequest struct {
	DurationMinutes int `json:"duration_minutes"` // 0 = bỏ cấm, > 0 = số phút cấm
}

// UpdateGroupSettingsRequest DTO cập nhật cài đặt nhóm
type UpdateGroupSettingsRequest struct {
	RequireApproval *bool `json:"require_approval,omitempty"`
	SlowModeSeconds *int  `json:"slow_mode_seconds,omitempty"`
}
