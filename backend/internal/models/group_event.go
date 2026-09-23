package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventAttendee đại diện cho một người tham gia sự kiện và trạng thái phản hồi
type EventAttendee struct {
	UserID     string    `bson:"user_id" json:"user_id"`
	UserName   string    `bson:"user_name" json:"user_name"`
	UserAvatar string    `bson:"user_avatar,omitempty" json:"user_avatar,omitempty"`
	Status     string    `bson:"status" json:"status"` // "going", "maybe", "declined"
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

// GroupEvent đại diện cho một sự kiện / lịch hẹn nhóm trong collection "group_events"
type GroupEvent struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID       primitive.ObjectID `bson:"group_id" json:"group_id"`
	ChannelID     string             `bson:"channel_id,omitempty" json:"channel_id,omitempty"`
	MessageID     string             `bson:"message_id,omitempty" json:"message_id,omitempty"`
	Title         string             `bson:"title" json:"title"`
	Description   string             `bson:"description,omitempty" json:"description,omitempty"`
	Location      string             `bson:"location,omitempty" json:"location,omitempty"`
	StartTime     time.Time          `bson:"start_time" json:"start_time"`
	EndTime       *time.Time         `bson:"end_time,omitempty" json:"end_time,omitempty"`
	CreatedBy     primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatorName   string             `bson:"creator_name" json:"creator_name"`
	CreatorAvatar string             `bson:"creator_avatar,omitempty" json:"creator_avatar,omitempty"`
	Attendees     []EventAttendee    `bson:"attendees" json:"attendees"`
	Status        string             `bson:"status" json:"status"` // "upcoming", "ongoing", "completed", "cancelled"
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// CreateEventRequest DTO tạo sự kiện mới
type CreateEventRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	StartTime   time.Time  `json:"start_time" binding:"required"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	ChannelID   string     `json:"channel_id"`
}

// RSVPEventRequest DTO phản hồi tham gia
type RSVPEventRequest struct {
	Status string `json:"status" binding:"required,oneof=going maybe declined"`
}
