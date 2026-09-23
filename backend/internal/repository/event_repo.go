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

type EventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		collection: db.Collection("group_events"),
	}
}

// Create tạo một sự kiện mới
func (r *EventRepository) Create(ctx context.Context, event *models.GroupEvent) error {
	event.CreatedAt = time.Now()
	if event.Attendees == nil {
		event.Attendees = []models.EventAttendee{}
	}
	if event.Status == "" {
		event.Status = "upcoming"
	}
	res, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		return err
	}
	event.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID lấy chi tiết sự kiện
func (r *EventRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.GroupEvent, error) {
	var event models.GroupEvent
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetUpcomingByGroup lấy danh sách sự kiện của nhóm
func (r *EventRepository) GetUpcomingByGroup(ctx context.Context, groupID primitive.ObjectID) ([]models.GroupEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"group_id": groupID,
		"status":   bson.M{"$ne": "cancelled"},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.GroupEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	if events == nil {
		events = []models.GroupEvent{}
	}
	return events, nil
}

// UpdateMessageID gán message ID của sự kiện trong kênh chat
func (r *EventRepository) UpdateMessageID(ctx context.Context, eventID primitive.ObjectID, messageID string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": eventID}, bson.M{
		"$set": bson.M{"message_id": messageID},
	})
	return err
}

// RSVP cập nhật trạng thái tham gia của thành viên
func (r *EventRepository) RSVP(ctx context.Context, eventID primitive.ObjectID, attendee models.EventAttendee) (*models.GroupEvent, error) {
	event, err := r.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range event.Attendees {
		if event.Attendees[i].UserID == attendee.UserID {
			event.Attendees[i].Status = attendee.Status
			event.Attendees[i].UserName = attendee.UserName
			event.Attendees[i].UserAvatar = attendee.UserAvatar
			event.Attendees[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		attendee.UpdatedAt = time.Now()
		event.Attendees = append(event.Attendees, attendee)
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": eventID}, bson.M{
		"$set": bson.M{"attendees": event.Attendees},
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}

// Delete xóa sự kiện
func (r *EventRepository) Delete(ctx context.Context, eventID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": eventID})
	return err
}
