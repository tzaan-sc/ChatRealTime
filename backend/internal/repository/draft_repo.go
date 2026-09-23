package repository

import (
	"context"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DraftRepository struct {
	collection *mongo.Collection
}

func NewDraftRepository(db *mongo.Database) *DraftRepository {
	return &DraftRepository{
		collection: db.Collection("drafts"),
	}
}

// Upsert lưu hoặc cập nhật nội dung tin nhắn nháp
func (r *DraftRepository) Upsert(ctx context.Context, userID, targetID, targetType, content string) error {
	filter := bson.M{
		"user_id":   userID,
		"target_id": targetID,
	}

	if content == "" {
		_, err := r.collection.DeleteOne(ctx, filter)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"user_id":     userID,
			"target_id":   targetID,
			"target_type": targetType,
			"content":     content,
			"updated_at":  time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Delete xóa nháp khi tin đã được gửi đi
func (r *DraftRepository) Delete(ctx context.Context, userID, targetID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"user_id":   userID,
		"target_id": targetID,
	})
	return err
}

// GetUserDrafts lấy toàn bộ tin nhắn nháp của user
func (r *DraftRepository) GetUserDrafts(ctx context.Context, userID string) ([]models.Draft, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drafts []models.Draft
	if err := cursor.All(ctx, &drafts); err != nil {
		return nil, err
	}
	return drafts, nil
}
