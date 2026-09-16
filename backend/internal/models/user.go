package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User đại diện cho một tài khoản trong collection "users"
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username    string             `bson:"username" json:"username"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"-"` // Dấu "-" để không bao giờ lộ password ra JSON
	DisplayName string             `bson:"display_name" json:"display_name"`
	AvatarURL   string             `bson:"avatar_url" json:"avatar_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// Request DTO cho Đăng ký
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=30"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

// Request DTO cho Đăng nhập
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Response DTO sau khi Đăng nhập thành công
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
