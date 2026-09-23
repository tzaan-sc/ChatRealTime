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

type ReminderRepository struct {
	collection *mongo.Collection
}

func NewReminderRepository(db *mongo.Database) *ReminderRepository {
	return &ReminderRepository{
		collection: db.Collection("reminders"),
	}
}

// Create tạo lịch nhắc hẹn
func (r *ReminderRepository) Create(ctx context.Context, rem *models.Reminder) error {
	rem.ID = primitive.NewObjectID()
	rem.CreatedAt = time.Now()
	rem.Status = "pending"
	_, err := r.collection.InsertOne(ctx, rem)
	return err
}

// GetPendingDueReminders lấy danh sách các nhắc hẹn đã đến hạn
func (r *ReminderRepository) GetPendingDueReminders(ctx context.Context, now time.Time) ([]models.Reminder, error) {
	filter := bson.M{
		"status":    "pending",
		"remind_at": bson.M{"$lte": now},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Reminder
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// MarkAsTriggered đánh dấu nhắc việc đã phát thông báo
func (r *ReminderRepository) MarkAsTriggered(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "triggered"}})
	return err
}

// Dismiss bỏ qua nhắc việc
func (r *ReminderRepository) Dismiss(ctx context.Context, id primitive.ObjectID, userID string) error {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"status": "dismissed"}})
	return err
}

// GetUserReminders lấy danh sách nhắc việc còn pending của user
func (r *ReminderRepository) GetUserReminders(ctx context.Context, userID string) ([]models.Reminder, error) {
	filter := bson.M{
		"user_id": userID,
		"status":  "pending",
	}
	opts := options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Reminder
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
