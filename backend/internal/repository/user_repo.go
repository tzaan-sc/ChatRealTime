package repository

import (
	"context"
	"errors"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create chèn user mới vào MongoDB
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByUsername tìm user theo tên đăng nhập
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail tìm user theo email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID tìm user theo ObjectID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAllExcept lấy danh sách user khác bản thân
func (r *UserRepository) FindAllExcept(ctx context.Context, currentUserID string) ([]models.User, error) {
	objID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$ne": objID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// FindByIDs lấy danh sách user theo mảng ObjectID
func (r *UserRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// FindByOAuth tìm user theo provider và provider_id
func (r *UserRepository) FindByOAuth(ctx context.Context, provider, providerID string) (*models.User, error) {
	var user models.User
	filter := bson.M{
		"oauth_accounts": bson.M{
			"$elemMatch": bson.M{
				"provider":    provider,
				"provider_id": providerID,
			},
		},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// LinkOAuthAccount liên kết tài khoản mạng xã hội vào User
func (r *UserRepository) LinkOAuthAccount(ctx context.Context, userID primitive.ObjectID, account models.OAuthAccount) error {
	filter := bson.M{"_id": userID}
	pullUpdate := bson.M{
		"$pull": bson.M{
			"oauth_accounts": bson.M{"provider": account.Provider},
		},
	}
	_, _ = r.collection.UpdateOne(ctx, filter, pullUpdate)

	pushUpdate := bson.M{
		"$push": bson.M{
			"oauth_accounts": account,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, pushUpdate)
	return err
}

// UnlinkOAuthAccount gỡ liên kết tài khoản mạng xã hội khỏi User
func (r *UserRepository) UnlinkOAuthAccount(ctx context.Context, userID primitive.ObjectID, provider string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$pull": bson.M{
			"oauth_accounts": bson.M{"provider": provider},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Update updates generic user fields
func (r *UserRepository) Update(ctx context.Context, userID primitive.ObjectID, updateFields bson.M) error {
	updateFields["updated_at"] = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": updateFields})
	return err
}

// UpdatePassword cập nhật mật khẩu băm mới cho User
func (r *UserRepository) UpdatePassword(ctx context.Context, userID primitive.ObjectID, hashedPassword string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// SetEmailVerified cập nhật trạng thái email đã xác thực
func (r *UserRepository) SetEmailVerified(ctx context.Context, email string, verified bool) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"email_verified": verified,
			"updated_at":     time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}




