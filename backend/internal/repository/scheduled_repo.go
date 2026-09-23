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

type ScheduledRepository struct {
	collection *mongo.Collection
}

func NewScheduledRepository(db *mongo.Database) *ScheduledRepository {
	return &ScheduledRepository{
		collection: db.Collection("scheduled_messages"),
	}
}

// Create tạo tin nhắn hẹn giờ
func (r *ScheduledRepository) Create(ctx context.Context, msg *models.ScheduledMessage) error {
	msg.ID = primitive.NewObjectID()
	msg.CreatedAt = time.Now()
	msg.Status = "pending"
	_, err := r.collection.InsertOne(ctx, msg)
	return err
}

// GetPendingDueMessages lấy các tin nhắn đang chờ gửi đã đến hoặc quá hạn hẹn
func (r *ScheduledRepository) GetPendingDueMessages(ctx context.Context, now time.Time) ([]models.ScheduledMessage, error) {
	filter := bson.M{
		"status":       "pending",
		"scheduled_at": bson.M{"$lte": now},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.ScheduledMessage
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// MarkAsSent cập nhật trạng thái đã gửi
func (r *ScheduledRepository) MarkAsSent(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "sent"}})
	return err
}

// Cancel hủy lịch gửi tin
func (r *ScheduledRepository) Cancel(ctx context.Context, id primitive.ObjectID, senderID string) error {
	filter := bson.M{
		"_id":       id,
		"sender_id": senderID,
		"status":    "pending",
	}
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"status": "cancelled"}})
	return err
}

// GetUserScheduledMessages lấy danh sách tin đã hẹn giờ của user cho một cuộc trò chuyện hoặc nhóm
func (r *ScheduledRepository) GetUserScheduledMessages(ctx context.Context, senderID, targetID string, isGroup bool) ([]models.ScheduledMessage, error) {
	filter := bson.M{
		"sender_id": senderID,
		"status":    "pending",
	}
	if isGroup {
		filter["group_id"] = targetID
	} else {
		filter["conversation_id"] = targetID
	}

	opts := options.Find().SetSort(bson.D{{Key: "scheduled_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.ScheduledMessage
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
