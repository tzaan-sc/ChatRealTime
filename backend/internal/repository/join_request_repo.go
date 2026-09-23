package repository

import (
	"context"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JoinRequestRepository struct {
	collection *mongo.Collection
}

func NewJoinRequestRepository(db *mongo.Database) *JoinRequestRepository {
	return &JoinRequestRepository{
		collection: db.Collection("group_join_requests"),
	}
}

// Create tạo yêu cầu tham gia nhóm
func (r *JoinRequestRepository) Create(ctx context.Context, req *models.GroupJoinRequest) error {
	req.ID = primitive.NewObjectID()
	req.Status = "pending"
	req.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, req)
	return err
}

// GetPendingByGroup lấy danh sách yêu cầu đang chờ duyệt của nhóm
func (r *JoinRequestRepository) GetPendingByGroup(ctx context.Context, groupID primitive.ObjectID) ([]models.GroupJoinRequest, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"group_id": groupID,
		"status":   "pending",
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.GroupJoinRequest
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// GetByID lấy thông tin yêu cầu theo ID
func (r *JoinRequestRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.GroupJoinRequest, error) {
	var req models.GroupJoinRequest
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// HasPending kiểm tra user có đang có đơn chờ duyệt vào nhóm không
func (r *JoinRequestRepository) HasPending(ctx context.Context, groupID, userID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
		"status":   "pending",
	})
	return count > 0, err
}

// UpdateStatus cập nhật trạng thái đơn (approved / rejected)
func (r *JoinRequestRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string, reviewerID primitive.ObjectID) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"reviewed_by": reviewerID,
			"reviewed_at": now,
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}
