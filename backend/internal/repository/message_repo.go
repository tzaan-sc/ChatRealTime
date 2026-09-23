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

	filter := bson.M{
		"conversation_id": conversationID,
		"$or": []bson.M{
			{"thread_root_id": ""},
			{"thread_root_id": bson.M{"$exists": false}},
		},
	}
	cursor, err := r.collection.Find(ctx, filter, opts)
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

// GetByGroup lấy lịch sử tin nhắn của một nhóm chat hoặc theo kênh cụ thể
func (r *MessageRepository) GetByGroup(ctx context.Context, groupID string, channelID string, limit int64, offset int64) ([]models.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	filter := bson.M{
		"group_id": groupID,
		"$or": []bson.M{
			{"thread_root_id": ""},
			{"thread_root_id": bson.M{"$exists": false}},
		},
	}
	if channelID != "" {
		filter["channel_id"] = channelID
	}
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

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

// MarkAsUnread đánh dấu tin nhắn mới nhất trong cuộc trò chuyện là chưa đọc
func (r *MessageRepository) MarkAsUnread(ctx context.Context, conversationID, receiverID string) error {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var latestMsg models.Message
	err := r.collection.FindOne(ctx, bson.M{
		"conversation_id": conversationID,
		"receiver_id":     receiverID,
	}, opts).Decode(&latestMsg)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": latestMsg.ID}, bson.M{"$set": bson.M{"is_read": false}})
	return err
}

// GetByID lấy thông tin 1 tin nhắn theo ObjectID
func (r *MessageRepository) GetByID(ctx context.Context, messageID primitive.ObjectID) (*models.Message, error) {
	var msg models.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&msg)
	return &msg, err
}

// ToggleReaction thêm, đổi hoặc hủy reaction của 1 user trên tin nhắn
func (r *MessageRepository) ToggleReaction(ctx context.Context, messageID primitive.ObjectID, userID, emoji string) ([]models.Reaction, error) {
	var msg models.Message
	if err := r.collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&msg); err != nil {
		return nil, err
	}

	found := false
	newReactions := make([]models.Reaction, 0)
	for _, react := range msg.Reactions {
		if react.UserID == userID {
			found = true
			if react.Emoji != emoji {
				// Đổi sang biểu tượng cảm xúc mới
				newReactions = append(newReactions, models.Reaction{UserID: userID, Emoji: emoji})
			}
			// Nếu bấm lại cùng emoji -> Xóa reaction (toggle off)
		} else {
			newReactions = append(newReactions, react)
		}
	}

	if !found {
		newReactions = append(newReactions, models.Reaction{UserID: userID, Emoji: emoji})
	}

	update := bson.M{"$set": bson.M{"reactions": newReactions}}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	return newReactions, err
}

// DeleteMessage thu hồi / xóa tin nhắn của người gửi
func (r *MessageRepository) DeleteMessage(ctx context.Context, messageID primitive.ObjectID, senderID string) error {
	filter := bson.M{
		"_id":       messageID,
		"sender_id": senderID,
	}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"content":    "",
			"file_name":  "",
			"file_size":  0,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// EditMessage chỉnh sửa nội dung tin nhắn của người gửi
func (r *MessageRepository) EditMessage(ctx context.Context, messageID primitive.ObjectID, senderID, newContent string) error {
	filter := bson.M{
		"_id":        messageID,
		"sender_id":  senderID,
		"is_deleted": false,
	}
	update := bson.M{
		"$set": bson.M{
			"content":    newContent,
			"is_edited":  true,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// PinMessage ghim tin nhắn trong cuộc trò chuyện hoặc nhóm
func (r *MessageRepository) PinMessage(ctx context.Context, messageID primitive.ObjectID, userID string) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": messageID}, bson.M{
		"$set": bson.M{
			"is_pinned": true,
			"pinned_at": &now,
			"pinned_by": userID,
		},
	})
	return err
}

// UnpinMessage gỡ ghim tin nhắn
func (r *MessageRepository) UnpinMessage(ctx context.Context, messageID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": messageID}, bson.M{
		"$set": bson.M{
			"is_pinned": false,
			"pinned_at": nil,
			"pinned_by": "",
		},
	})
	return err
}

// GetPinnedMessages lấy tất cả tin nhắn đã ghim trong một cuộc trò chuyện hoặc nhóm
func (r *MessageRepository) GetPinnedMessages(ctx context.Context, targetID string, isGroup bool) ([]models.Message, error) {
	filter := bson.M{
		"is_pinned":  true,
		"is_deleted": false,
	}
	if isGroup {
		filter["group_id"] = targetID
	} else {
		filter["conversation_id"] = targetID
	}
	opts := options.Find().SetSort(bson.D{{Key: "pinned_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// GetThreadMessages lấy danh sách tin nhắn phản hồi trong một luồng
func (r *MessageRepository) GetThreadMessages(ctx context.Context, rootMessageID string) ([]models.Message, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"thread_root_id": rootMessageID, "is_deleted": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// IncrementThreadCount tăng số đếm phản hồi trong luồng của tin nhắn gốc
func (r *MessageRepository) IncrementThreadCount(ctx context.Context, rootMessageID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": rootMessageID}, bson.M{
		"$inc": bson.M{"thread_count": 1},
	})
	return err
}


