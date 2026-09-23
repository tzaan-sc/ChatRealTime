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

type GroupRepository struct {
	collection *mongo.Collection
}

func NewGroupRepository(db *mongo.Database) *GroupRepository {
	return &GroupRepository{
		collection: db.Collection("groups"),
	}
}

// Create tạo nhóm mới
func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	group.ID = primitive.NewObjectID()
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, group)
	return err
}

// GetUserGroups lấy danh sách các nhóm mà user đang tham gia
func (r *GroupRepository) GetUserGroups(ctx context.Context, userID primitive.ObjectID) ([]models.Group, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"member_ids": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []models.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GetByID lấy thông tin nhóm theo ID
func (r *GroupRepository) GetByID(ctx context.Context, groupID primitive.ObjectID) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// AddMembers thêm danh sách thành viên mới vào nhóm
func (r *GroupRepository) AddMembers(ctx context.Context, groupID primitive.ObjectID, memberIDs []primitive.ObjectID) error {
	update := bson.M{
		"$addToSet": bson.M{"member_ids": bson.M{"$each": memberIDs}},
		"$set":      bson.M{"updated_at": time.Now()},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

// RemoveMember xóa hoặc để một thành viên rời nhóm
func (r *GroupRepository) RemoveMember(ctx context.Context, groupID primitive.ObjectID, memberID primitive.ObjectID) error {
	update := bson.M{
		"$pull": bson.M{
			"member_ids": memberID,
			"admin_ids":  memberID,
		},
		"$set": bson.M{"updated_at": time.Now()},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

// IsMember kiểm tra user có thuộc nhóm hay không
func (r *GroupRepository) IsMember(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"_id":        groupID,
		"member_ids": userID,
	})
	return count > 0, err
}

// IsAdmin kiểm tra user có phải là Admin nhóm hay không
func (r *GroupRepository) IsAdmin(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"_id":       groupID,
		"admin_ids": userID,
	})
	return count > 0, err
}

// UpdateSlowMode cập nhật thời gian giới hạn gửi tin trong nhóm
func (r *GroupRepository) UpdateSlowMode(ctx context.Context, groupID primitive.ObjectID, slowModeSeconds int) error {
	update := bson.M{
		"$set": bson.M{
			"slow_mode_seconds": slowModeSeconds,
			"updated_at":        time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

