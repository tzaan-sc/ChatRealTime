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
	group.IsCommunity = true

	// Khởi tạo mặc định kênh và danh mục
	if len(group.Categories) == 0 {
		catID := primitive.NewObjectID().Hex()
		group.Categories = []models.ChannelCategory{
			{
				ID:        catID,
				Name:      "CHUNG",
				CreatedAt: now,
			},
		}
		if len(group.Channels) == 0 {
			group.Channels = []models.Channel{
				{
					ID:          primitive.NewObjectID().Hex(),
					Name:        "thong-bao",
					Description: "Kênh thông báo chính thức",
					Type:        models.ChannelTypeAnnouncement,
					CategoryID:  catID,
					CreatedAt:   now,
				},
				{
					ID:          primitive.NewObjectID().Hex(),
					Name:        "chung",
					Description: "Kênh trò chuyện thảo luận chung",
					Type:        models.ChannelTypeText,
					CategoryID:  catID,
					CreatedAt:   now,
				},
			}
		}
	}

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

// GetByID lấy thông tin nhóm theo ID (tự động nâng cấp kênh mặc định nếu nhóm cũ chưa có)
func (r *GroupRepository) GetByID(ctx context.Context, groupID primitive.ObjectID) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return nil, err
	}

	// Tự động nâng cấp nhóm cũ thành Không gian cộng đồng có kênh mặc định
	if len(group.Channels) == 0 {
		catID := primitive.NewObjectID().Hex()
		now := time.Now()
		group.Categories = []models.ChannelCategory{
			{
				ID:        catID,
				Name:      "CHUNG",
				CreatedAt: now,
			},
		}
		group.Channels = []models.Channel{
			{
				ID:          primitive.NewObjectID().Hex(),
				Name:        "thong-bao",
				Description: "Kênh thông báo chính thức",
				Type:        models.ChannelTypeAnnouncement,
				CategoryID:  catID,
				CreatedAt:   now,
			},
			{
				ID:          primitive.NewObjectID().Hex(),
				Name:        "chung",
				Description: "Kênh trò chuyện thảo luận chung",
				Type:        models.ChannelTypeText,
				CategoryID:  catID,
				CreatedAt:   now,
			},
		}
		group.IsCommunity = true
		_, _ = r.collection.UpdateOne(ctx, bson.M{"_id": group.ID}, bson.M{
			"$set": bson.M{
				"is_community": true,
				"categories":   group.Categories,
				"channels":     group.Channels,
			},
		})
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

// IsAdmin kiểm tra user có phải là Admin hoặc Creator của nhóm hay không
func (r *GroupRepository) IsAdmin(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"_id": groupID,
		"$or": []bson.M{
			{"creator_id": userID},
			{"admin_ids": userID},
		},
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

// AddCategory thêm danh mục kênh mới
func (r *GroupRepository) AddCategory(ctx context.Context, groupID primitive.ObjectID, category models.ChannelCategory) error {
	update := bson.M{
		"$push": bson.M{"categories": category},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

// DeleteCategory xóa danh mục kênh và chuyển các kênh thuộc danh mục này về không danh mục
func (r *GroupRepository) DeleteCategory(ctx context.Context, groupID primitive.ObjectID, catID string) error {
	update := bson.M{
		"$pull": bson.M{"categories": bson.M{"id": catID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

// AddChannel thêm kênh mới vào nhóm
func (r *GroupRepository) AddChannel(ctx context.Context, groupID primitive.ObjectID, channel models.Channel) error {
	update := bson.M{
		"$push": bson.M{"channels": channel},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

// DeleteChannel xóa kênh khỏi nhóm
func (r *GroupRepository) DeleteChannel(ctx context.Context, groupID primitive.ObjectID, chanID string) error {
	update := bson.M{
		"$pull": bson.M{"channels": bson.M{"id": chanID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	return err
}

