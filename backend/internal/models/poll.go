package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollOption phương án lựa chọn trong cuộc bình chọn
type PollOption struct {
	ID        string   `bson:"id" json:"id"`
	Text      string   `bson:"text" json:"text"`
	VoterIDs  []string `bson:"voter_ids" json:"voter_ids"`
	VoteCount int      `bson:"vote_count" json:"vote_count"`
}

// Poll đại diện cho một cuộc bình chọn trong collection "polls"
type Poll struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID        primitive.ObjectID `bson:"group_id" json:"group_id"`
	ChannelID      string             `bson:"channel_id,omitempty" json:"channel_id,omitempty"`
	MessageID      string             `bson:"message_id,omitempty" json:"message_id,omitempty"`
	Question       string             `bson:"question" json:"question"`
	Options        []PollOption       `bson:"options" json:"options"`
	MultipleChoice bool               `bson:"multiple_choice" json:"multiple_choice"`
	IsAnonymous    bool               `bson:"is_anonymous" json:"is_anonymous"`
	IsClosed       bool               `bson:"is_closed" json:"is_closed"`
	TotalVotes     int                `bson:"total_votes" json:"total_votes"`
	CreatedBy      primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatorName    string             `bson:"creator_name" json:"creator_name"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	ClosedAt       *time.Time         `bson:"closed_at,omitempty" json:"closed_at,omitempty"`
}

// CreatePollRequest DTO tạo cuộc bình chọn mới
type CreatePollRequest struct {
	Question       string   `json:"question" binding:"required"`
	Options        []string `json:"options" binding:"required,min=2"`
	MultipleChoice bool     `json:"multiple_choice"`
	IsAnonymous    bool     `json:"is_anonymous"`
	ChannelID      string   `json:"channel_id"`
}

// VotePollRequest DTO bỏ phiếu
type VotePollRequest struct {
	OptionID string `json:"option_id" binding:"required"`
}
