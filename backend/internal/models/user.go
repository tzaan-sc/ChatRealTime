package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OAuthAccount thông tin định danh từ bên thứ 3 (Google, Facebook, GitHub, Apple, Discord...)
type OAuthAccount struct {
	Provider   string    `bson:"provider" json:"provider"` // "google", "facebook", "github", "apple", "discord"
	ProviderID string    `bson:"provider_id" json:"provider_id"`
	Email      string    `bson:"email" json:"email"`
	Name       string    `bson:"name" json:"name"`
	AvatarURL  string    `bson:"avatar_url" json:"avatar_url"`
	LinkedAt   time.Time `bson:"linked_at" json:"linked_at"`
}

// User đại diện cho một tài khoản trong collection "users"
type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username      string             `bson:"username" json:"username"`
	Email         string             `bson:"email" json:"email"`
	Password      string             `bson:"password" json:"-"` // Dấu "-" để không bao giờ lộ password ra JSON
	DisplayName   string             `bson:"display_name" json:"display_name"`
	AvatarURL     string             `bson:"avatar_url" json:"avatar_url"`
	EmailVerified bool               `bson:"email_verified" json:"email_verified"`
	OAuthAccounts []OAuthAccount     `bson:"oauth_accounts,omitempty" json:"oauth_accounts,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
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

// Request DTO cho Đăng nhập / Đăng ký qua OAuth 2.0 (Google, Facebook, GitHub, Apple, Discord)
type OAuthLoginRequest struct {
	Provider   string `json:"provider" binding:"required"` // "google", "facebook", "github", "apple", "discord"
	Token      string `json:"token"`                       // ID Token / Access Token
	Code       string `json:"code"`                        // Auth Code nếu có
	ProviderID string `json:"provider_id"`                 // Unique ID phía OAuth provider
	Email      string `json:"email"`                       // Email nhận được từ OAuth
	Name       string `json:"name"`                        // Tên từ OAuth
	AvatarURL  string `json:"avatar_url"`                  // Link ảnh đại diện
}

// Request DTO cho Liên kết tài khoản OAuth khi đã đăng nhập
type LinkOAuthRequest struct {
	Provider   string `json:"provider" binding:"required"`
	Token      string `json:"token"`
	ProviderID string `json:"provider_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
}

// Request DTO cho Quên mật khẩu
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Request DTO cho Đặt lại mật khẩu với mã OTP
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Request DTO cho Gửi email xác thực tài khoản
type SendVerificationEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Request DTO cho Xác minh email với mã OTP
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// Response DTO sau khi Đăng nhập thành công
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}


