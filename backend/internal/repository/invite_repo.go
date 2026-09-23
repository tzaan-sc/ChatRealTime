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

type InviteRepository struct {
	collection *mongo.Collection
}

func NewInviteRepository(db *mongo.Database) *InviteRepository {
	return &InviteRepository{
		collection: db.Collection("group_invites"),
	}
}

// Create tạo mã mời mới
func (r *InviteRepository) Create(ctx context.Context, invite *models.GroupInvite) error {
	invite.ID = primitive.NewObjectID()
	invite.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, invite)
	return err
}

// GetByCode tìm mã mời theo code
func (r *InviteRepository) GetByCode(ctx context.Context, code string) (*models.GroupInvite, error) {
	var inv models.GroupInvite
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&inv)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetActiveByGroup lấy danh sách mã mời đang có hiệu lực của nhóm
func (r *InviteRepository) GetActiveByGroup(ctx context.Context, groupID primitive.ObjectID) ([]models.GroupInvite, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.GroupInvite
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// IncrementUses tăng số lượt đã sử dụng mã mời
func (r *InviteRepository) IncrementUses(ctx context.Context, code string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"code": code}, bson.M{
		"$inc": bson.M{"uses_count": 1},
	})
	return err
}

// DeleteByCode xóa/thu hồi mã mời
func (r *InviteRepository) DeleteByCode(ctx context.Context, code string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"code": code})
	return err
}
