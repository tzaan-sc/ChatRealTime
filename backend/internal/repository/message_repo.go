package repository

import (
	"context"
	"math"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepository struct {
	collection *mongo.Collection
	pollRepo   *PollRepository
	eventRepo  *EventRepository
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection("messages"),
	}
}

func (r *MessageRepository) SetPollAndEventRepos(pollRepo *PollRepository, eventRepo *EventRepository) {
	r.pollRepo = pollRepo
	r.eventRepo = eventRepo
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

	// Tải chi tiết Poll & Event mới nhất nếu có
	for i := range messages {
		if messages[i].PollID != "" && r.pollRepo != nil {
			if pOID, err := primitive.ObjectIDFromHex(messages[i].PollID); err == nil {
				if p, err := r.pollRepo.GetByID(ctx, pOID); err == nil {
					messages[i].Poll = p
				}
			}
		}
		if messages[i].EventID != "" && r.eventRepo != nil {
			if eOID, err := primitive.ObjectIDFromHex(messages[i].EventID); err == nil {
				if ev, err := r.eventRepo.GetByID(ctx, eOID); err == nil {
					messages[i].Event = ev
				}
			}
		}
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

// GetGroupAnalytics tổng hợp toàn diện các chỉ số phân tích của nhóm
func (r *MessageRepository) GetGroupAnalytics(ctx context.Context, groupID string, group *models.Group) (*models.GroupAnalyticsResponse, error) {
	resp := &models.GroupAnalyticsResponse{
		GroupID:          groupID,
		GroupName:        group.Name,
		TotalMembers:     len(group.MemberIDs),
		DailyActivity:    []models.DailyStat{},
		TopMembers:       []models.MemberLeaderboardItem{},
		ChannelStats:     []models.ChannelActivityItem{},
		MessageTypeStats: []models.MessageTypeItem{},
		PeakHour:         20,
	}

	// 1. Tổng số tin nhắn trong nhóm
	totalMsgs, err := r.collection.CountDocuments(ctx, bson.M{"group_id": groupID, "is_deleted": bson.M{"$ne": true}})
	if err != nil {
		return nil, err
	}
	resp.TotalMessages = int(totalMsgs)

	// 2. Số thành viên tích cực trong 7 ngày qua
	sevenDaysAgo := time.Now().AddDate(0, 0, -6)
	startOfSevenDays := time.Date(sevenDaysAgo.Year(), sevenDaysAgo.Month(), sevenDaysAgo.Day(), 0, 0, 0, 0, sevenDaysAgo.Location())

	activeMembersPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"group_id":   groupID,
			"created_at": bson.M{"$gte": startOfSevenDays},
			"is_deleted": bson.M{"$ne": true},
		}}},
		{{Key: "$group", Value: bson.M{"_id": "$sender_id"}}},
	}
	activeCursor, err := r.collection.Aggregate(ctx, activeMembersPipeline)
	if err == nil {
		var activeResults []bson.M
		if err := activeCursor.All(ctx, &activeResults); err == nil {
			resp.ActiveMembers7d = len(activeResults)
		}
	}

	// 3. Hoạt động 7 ngày gần nhất
	dailyPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"group_id":   groupID,
			"created_at": bson.M{"$gte": startOfSevenDays},
			"is_deleted": bson.M{"$ne": true},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}
	dailyCursor, err := r.collection.Aggregate(ctx, dailyPipeline)
	dailyMap := make(map[string]int)
	if err == nil {
		var dailyResults []struct {
			Date  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := dailyCursor.All(ctx, &dailyResults); err == nil {
			for _, item := range dailyResults {
				dailyMap[item.Date] = item.Count
			}
		}
	}
	for i := 0; i < 7; i++ {
		d := startOfSevenDays.AddDate(0, 0, i)
		dStr := d.Format("2006-01-02")
		resp.DailyActivity = append(resp.DailyActivity, models.DailyStat{
			Date:  dStr,
			Count: dailyMap[dStr],
		})
	}

	// 4. Top 10 thành viên tích cực nhất
	topMembersPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"group_id":   groupID,
			"is_deleted": bson.M{"$ne": true},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":           "$sender_id",
			"message_count": bson.M{"$sum": 1},
			"sender_name":   bson.M{"$last": "$sender_name"},
			"sender_avatar": bson.M{"$last": "$sender_avatar"},
		}}},
		{{Key: "$sort", Value: bson.M{"message_count": -1}}},
		{{Key: "$limit", Value: 10}},
	}
	topCursor, err := r.collection.Aggregate(ctx, topMembersPipeline)
	if err == nil {
		var topResults []struct {
			SenderID     string `bson:"_id"`
			MessageCount int    `bson:"message_count"`
			SenderName   string `bson:"sender_name"`
			SenderAvatar string `bson:"sender_avatar"`
		}
		if err := topCursor.All(ctx, &topResults); err == nil {
			for _, tr := range topResults {
				role := "member"
				if senderOID, err := primitive.ObjectIDFromHex(tr.SenderID); err == nil {
					if senderOID == group.CreatorID {
						role = "owner"
					} else {
						for _, adminOID := range group.AdminIDs {
							if adminOID == senderOID {
								role = "admin"
								break
							}
						}
						if role == "member" {
							for _, modOID := range group.ModeratorIDs {
								if modOID == senderOID {
									role = "moderator"
									break
								}
							}
						}
					}
				}

				resp.TopMembers = append(resp.TopMembers, models.MemberLeaderboardItem{
					UserID:       tr.SenderID,
					DisplayName:  tr.SenderName,
					AvatarURL:    tr.SenderAvatar,
					Role:         role,
					MessageCount: tr.MessageCount,
				})
			}
		}
	}

	// 5. Thống kê theo kênh chat
	channelMap := make(map[string]models.Channel)
	for _, ch := range group.Channels {
		channelMap[ch.ID] = ch
	}

	chanPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"group_id":   groupID,
			"is_deleted": bson.M{"$ne": true},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$channel_id",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
	}
	chanCursor, err := r.collection.Aggregate(ctx, chanPipeline)
	if err == nil {
		var chanResults []struct {
			ChannelID string `bson:"_id"`
			Count     int    `bson:"count"`
		}
		if err := chanCursor.All(ctx, &chanResults); err == nil {
			for _, cr := range chanResults {
				chName := "Chung"
				chType := "text"
				if ch, ok := channelMap[cr.ChannelID]; ok {
					chName = ch.Name
					chType = string(ch.Type)
				}
				pct := 0.0
				if totalMsgs > 0 {
					pct = math.Round((float64(cr.Count)/float64(totalMsgs))*1000) / 10
				}
				resp.ChannelStats = append(resp.ChannelStats, models.ChannelActivityItem{
					ChannelID:    cr.ChannelID,
					ChannelName:  chName,
					ChannelType:  chType,
					MessageCount: cr.Count,
					Percentage:   pct,
				})
			}
		}
	}

	// 6. Thống kê theo loại tin nhắn
	typePipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"group_id":   groupID,
			"is_deleted": bson.M{"$ne": true},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$type",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
	}
	typeCursor, err := r.collection.Aggregate(ctx, typePipeline)
	if err == nil {
		var typeResults []struct {
			Type  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := typeCursor.All(ctx, &typeResults); err == nil {
			for _, tr := range typeResults {
				tName := tr.Type
				if tName == "" {
					tName = "text"
				}
				pct := 0.0
				if totalMsgs > 0 {
					pct = math.Round((float64(tr.Count)/float64(totalMsgs))*1000) / 10
				}
				resp.MessageTypeStats = append(resp.MessageTypeStats, models.MessageTypeItem{
					Type:       tName,
					Count:      tr.Count,
					Percentage: pct,
				})
			}
		}
	}

	// 7. Khung giờ hoạt động sôi nổi nhất (Peak Hour)
	peakPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"group_id":   groupID,
			"is_deleted": bson.M{"$ne": true},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   bson.M{"$hour": "$created_at"},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 1}},
	}
	peakCursor, err := r.collection.Aggregate(ctx, peakPipeline)
	if err == nil {
		var peakResults []struct {
			Hour  int `bson:"_id"`
			Count int `bson:"count"`
		}
		if err := peakCursor.All(ctx, &peakResults); err == nil && len(peakResults) > 0 {
			resp.PeakHour = peakResults[0].Hour
		}
	}

	return resp, nil
}


