package repository

import (
	"context"
	"errors"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PollRepository struct {
	collection *mongo.Collection
}

func NewPollRepository(db *mongo.Database) *PollRepository {
	return &PollRepository{
		collection: db.Collection("polls"),
	}
}

// Create tạo một cuộc bình chọn mới
func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.CreatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, poll)
	if err != nil {
		return err
	}
	poll.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID lấy thông tin cuộc bình chọn theo ID
func (r *PollRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

// UpdateMessageID gán MessageID tương ứng
func (r *PollRepository) UpdateMessageID(ctx context.Context, pollID primitive.ObjectID, messageID string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": pollID}, bson.M{
		"$set": bson.M{"message_id": messageID},
	})
	return err
}

// Vote thực hiện bỏ phiếu hoặc đổi phiếu
func (r *PollRepository) Vote(ctx context.Context, pollID primitive.ObjectID, userID string, optionID string) (*models.Poll, error) {
	poll, err := r.GetByID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	if poll.IsClosed {
		return nil, errors.New("cuộc bình chọn này đã kết thúc")
	}

	// Kiểm tra optionID có tồn tại
	optionExists := false
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			optionExists = true
			break
		}
	}
	if !optionExists {
		return nil, errors.New("phương án lựa chọn không tồn tại")
	}

	if !poll.MultipleChoice {
		// Đơn tuyển (Chỉ được chọn 1 phương án)
		for i := range poll.Options {
			if poll.Options[i].ID == optionID {
				// Nếu đã vote phương án này -> Hủy vote (toggle)
				if containsString(poll.Options[i].VoterIDs, userID) {
					poll.Options[i].VoterIDs = removeString(poll.Options[i].VoterIDs, userID)
				} else {
					poll.Options[i].VoterIDs = append(poll.Options[i].VoterIDs, userID)
				}
			} else {
				// Bỏ vote khỏi tất cả phương án khác
				poll.Options[i].VoterIDs = removeString(poll.Options[i].VoterIDs, userID)
			}
			poll.Options[i].VoteCount = len(poll.Options[i].VoterIDs)
		}
	} else {
		// Đa tuyển (Được chọn nhiều phương án)
		for i := range poll.Options {
			if poll.Options[i].ID == optionID {
				if containsString(poll.Options[i].VoterIDs, userID) {
					poll.Options[i].VoterIDs = removeString(poll.Options[i].VoterIDs, userID)
				} else {
					poll.Options[i].VoterIDs = append(poll.Options[i].VoterIDs, userID)
				}
			}
			poll.Options[i].VoteCount = len(poll.Options[i].VoterIDs)
		}
	}

	// Đếm tổng số người đã tham gia vote (unique voters)
	uniqueVoters := make(map[string]bool)
	for _, opt := range poll.Options {
		for _, v := range opt.VoterIDs {
			uniqueVoters[v] = true
		}
	}
	poll.TotalVotes = len(uniqueVoters)

	// Cập nhật lại database
	update := bson.M{
		"$set": bson.M{
			"options":     poll.Options,
			"total_votes": poll.TotalVotes,
		},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": pollID}, update)
	if err != nil {
		return nil, err
	}

	return poll, nil
}

// Close đóng cuộc bình chọn
func (r *PollRepository) Close(ctx context.Context, pollID primitive.ObjectID) (*models.Poll, error) {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_closed": true,
			"closed_at": &now,
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": pollID}, update)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, pollID)
}

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

func removeString(arr []string, target string) []string {
	res := make([]string, 0, len(arr))
	for _, s := range arr {
		if s != target {
			res = append(res, s)
		}
	}
	return res
}
