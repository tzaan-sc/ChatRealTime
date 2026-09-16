package service

import (
	"context"
	"errors"
	"os"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/pkg/utils"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register xử lý logic đăng ký tài khoản
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
	// 1. Kiểm tra username đã tồn tại chưa
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("tên đăng nhập đã được sử dụng")
	}

	// 2. Kiểm tra email đã tồn tại chưa
	existingEmail, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, errors.New("email đã được đăng ký")
	}

	// 3. Băm mật khẩu
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("lỗi khi mã hóa mật khẩu")
	}

	// 4. Tạo user mới
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	newUser := &models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    hashedPassword,
		DisplayName: displayName,
		AvatarURL:   "https://api.dicebear.com/7.x/bottts/svg?seed=" + req.Username,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, errors.New("không thể tạo tài khoản")
	}

	// 5. Cấp phát JWT Token
	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := utils.GenerateToken(newUser.ID.Hex(), newUser.Username, jwtSecret)
	if err != nil {
		return nil, errors.New("lỗi khi cấp phát token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *newUser,
	}, nil
}

// Login xử lý logic đăng nhập
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	// 1. Tìm user theo username
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("sai tên đăng nhập hoặc mật khẩu")
	}

	// 2. So khớp mật khẩu
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("sai tên đăng nhập hoặc mật khẩu")
	}

	// 3. Tạo Token
	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := utils.GenerateToken(user.ID.Hex(), user.Username, jwtSecret)
	if err != nil {
		return nil, errors.New("lỗi khi cấp phát token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetUserProfile lấy thông tin người dùng từ ID
func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}
