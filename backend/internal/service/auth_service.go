package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/pkg/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	redisClient *redis.Client
}

func NewAuthService(userRepo *repository.UserRepository, redisClient *redis.Client) *AuthService {
	return &AuthService{userRepo: userRepo, redisClient: redisClient}
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
		Username:      req.Username,
		Email:         req.Email,
		Password:      hashedPassword,
		DisplayName:   displayName,
		AvatarURL:     "https://api.dicebear.com/7.x/bottts/svg?seed=" + req.Username,
		EmailVerified: false,
		OAuthAccounts: []models.OAuthAccount{},
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

// OAuthLogin xử lý đăng nhập hoặc đăng ký nhanh qua mạng xã hội (Google, Facebook, GitHub, Apple, Discord)
func (s *AuthService) OAuthLogin(ctx context.Context, req models.OAuthLoginRequest) (*models.AuthResponse, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		return nil, errors.New("nhà cung cấp xác thực (provider) không hợp lệ")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	avatarURL := strings.TrimSpace(req.AvatarURL)
	providerID := strings.TrimSpace(req.ProviderID)

	// Nếu có token thật gửi lên và chưa có info, thử xác thực từ API của Provider
	if req.Token != "" && (email == "" || providerID == "") {
		verifiedEmail, verifiedName, verifiedAvatar, verifiedID, err := s.verifyOAuthToken(provider, req.Token)
		if err == nil && verifiedID != "" {
			providerID = verifiedID
			if verifiedEmail != "" {
				email = verifiedEmail
			}
			if verifiedName != "" {
				name = verifiedName
			}
			if verifiedAvatar != "" {
				avatarURL = verifiedAvatar
			}
		}
	}

	if providerID == "" {
		// Nếu là chế độ test / demo nhanh, tự tạo providerID từ email
		if email != "" {
			providerID = fmt.Sprintf("%s_%s", provider, strings.ReplaceAll(email, "@", "_"))
		} else {
			return nil, errors.New("thiếu thông tin định danh Provider ID hoặc Email")
		}
	}

	if email == "" {
		email = fmt.Sprintf("%s_%s@chatrealtime.local", provider, providerID)
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}
	if avatarURL == "" {
		avatarURL = "https://api.dicebear.com/7.x/bottts/svg?seed=" + providerID
	}

	oauthAcc := models.OAuthAccount{
		Provider:   provider,
		ProviderID: providerID,
		Email:      email,
		Name:       name,
		AvatarURL:  avatarURL,
		LinkedAt:   time.Now(),
	}

	// 1. Kiểm tra tài khoản đã từng liên kết với Provider + ProviderID này chưa
	user, err := s.userRepo.FindByOAuth(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		// Đã tìm thấy tài khoản! Đồng bộ lại avatar/tên mới nhất nếu có
		if avatarURL != "" && user.AvatarURL == "" {
			_ = s.userRepo.Update(ctx, user.ID, bson.M{"avatar_url": avatarURL})
			user.AvatarURL = avatarURL
		}
		// Cấp Token
		jwtSecret := os.Getenv("JWT_SECRET")
		token, err := utils.GenerateToken(user.ID.Hex(), user.Username, jwtSecret)
		if err != nil {
			return nil, errors.New("lỗi khi cấp phát token")
		}
		return &models.AuthResponse{Token: token, User: *user}, nil
	}

	// 2. Nếu chưa liên kết theo ProviderID, kiểm tra xem Email đã tồn tại trong hệ thống chưa
	userByEmail, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if userByEmail != nil {
		// Email đã có tài khoản sẵn (đăng ký thường hoặc qua bên khác) -> Tự động liên kết
		if err := s.userRepo.LinkOAuthAccount(ctx, userByEmail.ID, oauthAcc); err != nil {
			return nil, errors.New("không thể liên kết tài khoản mạng xã hội")
		}
		userByEmail.OAuthAccounts = append(userByEmail.OAuthAccounts, oauthAcc)
		userByEmail.EmailVerified = true
		_ = s.userRepo.Update(ctx, userByEmail.ID, bson.M{"email_verified": true})

		jwtSecret := os.Getenv("JWT_SECRET")
		token, err := utils.GenerateToken(userByEmail.ID.Hex(), userByEmail.Username, jwtSecret)
		if err != nil {
			return nil, errors.New("lỗi khi cấp phát token")
		}
		return &models.AuthResponse{Token: token, User: *userByEmail}, nil
	}

	// 3. User mới hoàn toàn -> Tạo tài khoản tự động
	baseUsername := strings.Split(email, "@")[0]
	cleanUsername := strings.ToLower(strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, baseUsername))
	if len(cleanUsername) < 3 {
		cleanUsername = provider + "_" + cleanUsername
	}
	if len(cleanUsername) > 20 {
		cleanUsername = cleanUsername[:20]
	}

	finalUsername := cleanUsername
	// Kiểm tra trùng username
	for i := 1; i <= 10; i++ {
		exist, _ := s.userRepo.FindByUsername(ctx, finalUsername)
		if exist == nil {
			break
		}
		finalUsername = fmt.Sprintf("%s%d", cleanUsername, time.Now().UnixNano()%1000)
	}

	newUser := &models.User{
		Username:      finalUsername,
		Email:         email,
		DisplayName:   name,
		AvatarURL:     avatarURL,
		EmailVerified: true,
		OAuthAccounts: []models.OAuthAccount{oauthAcc},
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, errors.New("không thể tạo tài khoản từ mạng xã hội: " + err.Error())
	}

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

// LinkOAuth liên kết mạng xã hội vào tài khoản đang đăng nhập
func (s *AuthService) LinkOAuth(ctx context.Context, currentUserID string, req models.LinkOAuthRequest) error {
	objID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return errors.New("ID người dùng không hợp lệ")
	}

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	providerID := strings.TrimSpace(req.ProviderID)
	if providerID == "" {
		providerID = fmt.Sprintf("%s_%s", provider, strings.ReplaceAll(req.Email, "@", "_"))
	}

	// Kiểm tra tài khoản mạng xã hội này đã bị ai khác liên kết chưa
	existOther, err := s.userRepo.FindByOAuth(ctx, provider, providerID)
	if err != nil {
		return err
	}
	if existOther != nil && existOther.ID != objID {
		return errors.New("tài khoản mạng xã hội này đã được liên kết với một người dùng khác")
	}

	account := models.OAuthAccount{
		Provider:   provider,
		ProviderID: providerID,
		Email:      req.Email,
		Name:       req.Name,
		AvatarURL:  req.AvatarURL,
		LinkedAt:   time.Now(),
	}

	return s.userRepo.LinkOAuthAccount(ctx, objID, account)
}

// UnlinkOAuth gỡ liên kết mạng xã hội
func (s *AuthService) UnlinkOAuth(ctx context.Context, currentUserID string, provider string) error {
	objID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return errors.New("ID người dùng không hợp lệ")
	}

	user, err := s.userRepo.FindByID(ctx, currentUserID)
	if err != nil || user == nil {
		return errors.New("không tìm thấy người dùng")
	}

	// Kiểm tra điều kiện bảo mật: User phải có mật khẩu hoặc còn tài khoản OAuth khác
	if user.Password == "" && len(user.OAuthAccounts) <= 1 {
		return errors.New("không thể gỡ liên kết vì đây là phương thức đăng nhập duy nhất của bạn. Hãy tạo mật khẩu trước!")
	}

	return s.userRepo.UnlinkOAuthAccount(ctx, objID, strings.ToLower(provider))
}

// verifyOAuthToken xác thực token trực tiếp với Google / Facebook / GitHub
func (s *AuthService) verifyOAuthToken(provider, token string) (email, name, avatar, providerID string, err error) {
	client := &http.Client{Timeout: 5 * time.Second}

	switch provider {
	case "google":
		resp, e := client.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + token)
		if e != nil || resp.StatusCode != http.StatusOK {
			return "", "", "", "", errors.New("token Google không hợp lệ")
		}
		defer resp.Body.Close()
		var gData struct {
			Sub     string `json:"sub"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&gData); err == nil {
			return gData.Email, gData.Name, gData.Picture, gData.Sub, nil
		}

	case "github":
		req, e := http.NewRequest("GET", "https://api.github.com/user", nil)
		if e != nil {
			return "", "", "", "", e
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, e := client.Do(req)
		if e != nil || resp.StatusCode != http.StatusOK {
			return "", "", "", "", errors.New("token GitHub không hợp lệ")
		}
		defer resp.Body.Close()
		var ghData struct {
			ID        int64  `json:"id"`
			Login     string `json:"login"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&ghData); err == nil {
			displayName := ghData.Name
			if displayName == "" {
				displayName = ghData.Login
			}
			return ghData.Email, displayName, ghData.AvatarURL, fmt.Sprintf("%d", ghData.ID), nil
		}

	case "facebook":
		resp, e := client.Get(fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email,picture.type(large)&access_token=%s", token))
		if e != nil || resp.StatusCode != http.StatusOK {
			return "", "", "", "", errors.New("token Facebook không hợp lệ")
		}
		defer resp.Body.Close()
		var fbData struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Email   string `json:"email"`
			Picture struct {
				Data struct {
					URL string `json:"url"`
				} `json:"data"`
			} `json:"picture"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&fbData); err == nil {
			return fbData.Email, fbData.Name, fbData.Picture.Data.URL, fbData.ID, nil
		}
	}

	return "", "", "", "", errors.New("không hỗ trợ xác thực tự động cho provider này")
}

// GetUserProfile lấy thông tin người dùng từ ID
func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// ForgotPassword tạo mã OTP 6 chữ số gửi qua email và lưu Redis với hạn 15 phút
func (s *AuthService) ForgotPassword(ctx context.Context, req models.ForgotPasswordRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("không tìm thấy tài khoản liên kết với địa chỉ email này")
	}

	// Rate limiting chống gửi liên tục: mỗi 60s tối đa 1 lần
	rateKey := fmt.Sprintf("rate:forgot_pw:%s", email)
	if s.redisClient != nil {
		exists, _ := s.redisClient.Exists(ctx, rateKey).Result()
		if exists > 0 {
			return "", errors.New("yêu cầu gửi mã quá nhanh. Vui lòng đợi 1 phút trước khi thử lại!")
		}
	}

	// Sinh mã OTP ngẫu nhiên 6 chữ số
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", r.Intn(1000000))

	// Lưu OTP vào Redis (15 phút)
	if s.redisClient != nil {
		otpKey := fmt.Sprintf("otp:forgot_pw:%s", email)
		if err := s.redisClient.Set(ctx, otpKey, otp, 15*time.Minute).Err(); err != nil {
			return "", errors.New("lỗi máy chủ khi tạo mã OTP")
		}
		_ = s.redisClient.Set(ctx, rateKey, "1", 60*time.Second).Err()
	}

	log.Printf("📧 [FORGOT_PASSWORD_OTP] Gửi mã đặt lại mật khẩu cho: %s | OTP: %s (Hết hạn trong 15 phút)\n", email, otp)
	return otp, nil
}

// ResetPassword kiểm tra mã OTP và cập nhật mật khẩu mới
func (s *AuthService) ResetPassword(ctx context.Context, req models.ResetPasswordRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	otp := strings.TrimSpace(req.OTP)

	if s.redisClient == nil {
		return errors.New("hệ thống lưu trữ OTP tạm thời không khả dụng")
	}

	otpKey := fmt.Sprintf("otp:forgot_pw:%s", email)
	storedOTP, err := s.redisClient.Get(ctx, otpKey).Result()
	if err != nil || storedOTP == "" {
		return errors.New("mã OTP không tồn tại hoặc đã hết hạn. Vui lòng gửi yêu cầu mới!")
	}

	if storedOTP != otp {
		return errors.New("mã OTP không chính xác. Vui lòng kiểm tra lại!")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("không tìm thấy người dùng")
	}

	// Băm mật khẩu mới
	hashedPw, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("lỗi khi mã hóa mật khẩu mới")
	}

	// Cập nhật vào DB
	if err := s.userRepo.UpdatePassword(ctx, user.ID, hashedPw); err != nil {
		return errors.New("không thể cập nhật mật khẩu: " + err.Error())
	}

	// Xóa OTP khỏi Redis
	_ = s.redisClient.Del(ctx, otpKey)
	log.Printf("🔑 [RESET_PASSWORD_SUCCESS] Đặt lại mật khẩu thành công cho tài khoản: %s\n", email)
	return nil
}

// SendVerificationEmail tạo mã OTP kích hoạt tài khoản gửi qua email (Hạn 24 giờ)
func (s *AuthService) SendVerificationEmail(ctx context.Context, req models.SendVerificationEmailRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("không tìm thấy tài khoản với email này")
	}
	if user.EmailVerified {
		return "", errors.New("địa chỉ email này đã được xác thực trước đó")
	}

	// Rate limiting chống gửi liên tục: 60s
	rateKey := fmt.Sprintf("rate:verify_email:%s", email)
	if s.redisClient != nil {
		exists, _ := s.redisClient.Exists(ctx, rateKey).Result()
		if exists > 0 {
			return "", errors.New("yêu cầu gửi email quá thường xuyên. Vui lòng thử lại sau 1 phút!")
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", r.Intn(1000000))

	if s.redisClient != nil {
		otpKey := fmt.Sprintf("otp:verify_email:%s", email)
		if err := s.redisClient.Set(ctx, otpKey, otp, 24*time.Hour).Err(); err != nil {
			return "", errors.New("lỗi máy chủ khi tạo mã xác thực email")
		}
		_ = s.redisClient.Set(ctx, rateKey, "1", 60*time.Second).Err()
	}

	log.Printf("✉️ [EMAIL_VERIFICATION_OTP] Mã kích hoạt tài khoản cho: %s | OTP: %s (Hết hạn trong 24 giờ)\n", email, otp)
	return otp, nil
}

// VerifyEmail kiểm tra mã kích hoạt và đánh dấu email_verified = true
func (s *AuthService) VerifyEmail(ctx context.Context, req models.VerifyEmailRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	otp := strings.TrimSpace(req.OTP)

	if s.redisClient == nil {
		return errors.New("hệ thống lưu trữ OTP tạm thời không khả dụng")
	}

	otpKey := fmt.Sprintf("otp:verify_email:%s", email)
	storedOTP, err := s.redisClient.Get(ctx, otpKey).Result()
	if err != nil || storedOTP == "" {
		return errors.New("mã xác thực email không hợp lệ hoặc đã hết hạn")
	}

	if storedOTP != otp {
		return errors.New("mã xác thực không chính xác")
	}

	if err := s.userRepo.SetEmailVerified(ctx, email, true); err != nil {
		return errors.New("không thể cập nhật trạng thái xác thực email: " + err.Error())
	}

	_ = s.redisClient.Del(ctx, otpKey)
	log.Printf("✅ [EMAIL_VERIFICATION_SUCCESS] Kích hoạt email thành công cho: %s\n", email)
	return nil
}


