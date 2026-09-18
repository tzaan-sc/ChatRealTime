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

type MessageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection("messages"),
	}
}

// Create chèn tin nhắn mới
func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	msg.ID = primitive.NewObjectID()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, msg)
	return err
}

// GetByConversation lấy lịch sử tin nhắn có phân trang (mới nhất trước)
func (r *MessageRepository) GetByConversation(ctx context.Context, conversationID string, limit int64, offset int64) ([]models.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}). // Mới nhất lên đầu
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, bson.M{"conversation_id": conversationID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Đảo lại thứ tự tăng dần theo thời gian để client hiển thị từ trên xuống dưới
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountUnread đếm số tin nhắn chưa đọc của một người trong 1 cuộc trò chuyện
func (r *MessageRepository) CountUnread(ctx context.Context, conversationID, receiverID string) (int64, error) {
	filter := bson.M{
		"conversation_id": conversationID,
		"receiver_id":     receiverID,
		"is_read":         false,
	}
	return r.collection.CountDocuments(ctx, filter)
}

// MarkAsRead đánh dấu các tin nhắn trong cuộc trò chuyện là đã đọc
func (r *MessageRepository) MarkAsRead(ctx context.Context, conversationID, receiverID string) error {
	filter := bson.M{
		"conversation_id": conversationID,
		"receiver_id":     receiverID,
		"is_read":         false,
	}
	update := bson.M{"$set": bson.M{"is_read": true}}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}
