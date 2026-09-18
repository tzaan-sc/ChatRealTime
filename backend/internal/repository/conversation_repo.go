package repository

import (
	"context"
	"errors"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationRepository struct {
	collection *mongo.Collection
}

func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{
		collection: db.Collection("conversations"),
	}
}

// GetOrCreate tìm hoặc tạo mới conversation giữa 2 người
func (r *ConversationRepository) GetOrCreate(ctx context.Context, customID string, member1, member2 string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.collection.FindOne(ctx, bson.M{"custom_id": customID}).Decode(&conv)
	if err == nil {
		return &conv, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		now := time.Now()
		newConv := models.Conversation{
			ID:            primitive.NewObjectID(),
			CustomID:      customID,
			Members:       []string{member1, member2},
			LastMessage:   "",
			LastSenderID:  "",
			LastMessageAt: now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		_, err := r.collection.InsertOne(ctx, newConv)
		if err != nil {
			return nil, err
		}
		return &newConv, nil
	}

	return nil, err
}

// UpdateLastMessage cập nhật tin nhắn cuối cùng và thời gian
func (r *ConversationRepository) UpdateLastMessage(ctx context.Context, customID, lastMsg, senderID string) error {
	now := time.Now()
	filter := bson.M{"custom_id": customID}
	update := bson.M{
		"$set": bson.M{
			"last_message":    lastMsg,
			"last_sender_id":  senderID,
			"last_message_at": now,
			"updated_at":      now,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetUserConversations lấy danh sách hội thoại của 1 user, sắp xếp mới nhất lên đầu
func (r *ConversationRepository) GetUserConversations(ctx context.Context, userID string) ([]models.Conversation, error) {
	opts := options.Find().SetSort(bson.D{{Key: "last_message_at", Value: -1}})
	filter := bson.M{"members": userID}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Conversation
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
