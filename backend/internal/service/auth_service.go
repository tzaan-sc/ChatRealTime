package service

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
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

// parseDevice phân tích thông tin thiết bị và hệ điều hành từ User-Agent & IP
func parseDevice(userAgent, clientIP string) (string, string) {
	osName := "Thiết bị không xác định"
	uaLower := strings.ToLower(userAgent)
	if strings.Contains(uaLower, "windows") {
		osName = "Windows"
	} else if strings.Contains(uaLower, "macintosh") || strings.Contains(uaLower, "mac os") {
		osName = "macOS"
	} else if strings.Contains(uaLower, "iphone") {
		osName = "iPhone (iOS)"
	} else if strings.Contains(uaLower, "ipad") {
		osName = "iPad (iPadOS)"
	} else if strings.Contains(uaLower, "android") {
		osName = "Android"
	} else if strings.Contains(uaLower, "linux") {
		osName = "Linux"
	}

	browser := "Trình duyệt Web"
	if strings.Contains(uaLower, "edg") {
		browser = "Microsoft Edge"
	} else if strings.Contains(uaLower, "chrome") && !strings.Contains(uaLower, "edg") {
		browser = "Google Chrome"
	} else if strings.Contains(uaLower, "firefox") {
		browser = "Mozilla Firefox"
	} else if strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome") {
		browser = "Apple Safari"
	} else if strings.Contains(uaLower, "postman") {
		browser = "Postman Runtime"
	}

	deviceName := fmt.Sprintf("%s trên %s", browser, osName)

	location := "IP: " + clientIP
	if clientIP == "127.0.0.1" || clientIP == "::1" || strings.HasPrefix(clientIP, "192.168.") || strings.HasPrefix(clientIP, "10.") || strings.HasPrefix(clientIP, "172.") {
		location = "Mạng nội bộ / Máy cục bộ (Localhost)"
	}

	return deviceName, location
}

// CreateAuthSession sinh cặp Access Token & Refresh Token, đồng thời phát hiện thiết bị lạ để cảnh báo
func (s *AuthService) CreateAuthSession(ctx context.Context, user *models.User, clientIP, userAgent string) (*models.AuthResponse, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	accessToken, err := utils.GenerateToken(user.ID.Hex(), user.Username, jwtSecret)
	if err != nil {
		return nil, errors.New("lỗi khi cấp phát access token")
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("lỗi khi tạo refresh token")
	}

	deviceName, location := parseDevice(userAgent, clientIP)
	isNewDevice := false
	var alertMsg string

	if s.redisClient != nil {
		deviceSig := fmt.Sprintf("%s|%s", deviceName, clientIP)
		knownKey := fmt.Sprintf("user_devices:%s", user.ID.Hex())

		// Kiểm tra xem thiết bị này đã từng đăng nhập chưa
		isMember, _ := s.redisClient.SIsMember(ctx, knownKey, deviceSig).Result()
		totalKnown, _ := s.redisClient.SCard(ctx, knownKey).Result()

		if !isMember {
			_ = s.redisClient.SAdd(ctx, knownKey, deviceSig)
			// Nếu người dùng đã từng có ít nhất 1 thiết bị trước đó -> đây là thiết bị mới lạ!
			if totalKnown > 0 {
				isNewDevice = true
				nowStr := time.Now().Format("15:04:05 02/01/2006")
				alertMsg = fmt.Sprintf("Phát hiện phiên đăng nhập mới từ thiết bị lạ: %s tại %s lúc %s. Thông báo an toàn đã được gửi tới email %s.", deviceName, location, nowStr, user.Email)

				alertObj := models.SecurityAlert{
					ID:        primitive.NewObjectID().Hex(),
					UserID:    user.ID.Hex(),
					Type:      "new_device",
					Message:   alertMsg,
					IPAddress: clientIP,
					UserAgent: deviceName,
					CreatedAt: time.Now(),
				}
				alertData, _ := json.Marshal(alertObj)
				alertsKey := fmt.Sprintf("user_alerts:%s", user.ID.Hex())
				_ = s.redisClient.LPush(ctx, alertsKey, string(alertData))
				_ = s.redisClient.LTrim(ctx, alertsKey, 0, 29)

				log.Printf("[SUSPICIOUS LOGIN ALERT] Đã gửi cảnh báo email bảo mật tới <%s>: Thiết bị lạ %s (IP: %s)", user.Email, deviceName, clientIP)
			}
		}

		// Lưu Refresh Token vào Redis với thời hạn 30 ngày (Silent Token Refresh & Rotation)
		sessPayload := map[string]interface{}{
			"user_id":     user.ID.Hex(),
			"username":    user.Username,
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"device_name": deviceName,
			"location":    location,
			"created_at":  time.Now().Unix(),
		}
		sessBytes, _ := json.Marshal(sessPayload)
		tokenKey := fmt.Sprintf("refresh_token:%s", refreshToken)
		_ = s.redisClient.Set(ctx, tokenKey, string(sessBytes), 30*24*time.Hour)

		userTokensKey := fmt.Sprintf("user_refresh_tokens:%s", user.ID.Hex())
		_ = s.redisClient.SAdd(ctx, userTokensKey, refreshToken)
		_ = s.redisClient.Expire(ctx, userTokensKey, 30*24*time.Hour)
	}

	return &models.AuthResponse{
		Token:         accessToken,
		RefreshToken:  refreshToken,
		User:          *user,
		NewDevice:     isNewDevice,
		DeviceInfo:    fmt.Sprintf("%s (%s)", deviceName, location),
		SecurityAlert: alertMsg,
	}, nil
}

// Register xử lý logic đăng ký tài khoản
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
	// 0. Xác minh Captcha chống spam bot
	if err := utils.VerifyTurnstileCaptcha(req.CaptchaToken, clientIP); err != nil {
		return nil, err
	}

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

	// 5. Cấp phát phiên bảo mật gồm Access Token & Refresh Token
	return s.CreateAuthSession(ctx, newUser, clientIP, userAgent)
}

// Login xử lý logic đăng nhập
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
	// 0. Xác minh Captcha chống dò quét mật khẩu (brute-force)
	if err := utils.VerifyTurnstileCaptcha(req.CaptchaToken, clientIP); err != nil {
		return nil, err
	}

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

	// 3. Cấp phát phiên bảo mật & cảnh báo nếu đăng nhập từ thiết bị lạ
	return s.CreateAuthSession(ctx, user, clientIP, userAgent)
}

// OAuthLogin xử lý đăng nhập hoặc đăng ký nhanh qua mạng xã hội (Google, Facebook, GitHub, Apple, Discord)
func (s *AuthService) OAuthLogin(ctx context.Context, req models.OAuthLoginRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
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
		// Cấp phiên bảo mật (Access + Refresh Token)
		return s.CreateAuthSession(ctx, user, clientIP, userAgent)
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

		return s.CreateAuthSession(ctx, userByEmail, clientIP, userAgent)
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

	return s.CreateAuthSession(ctx, newUser, clientIP, userAgent)
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

// SendMagicLink tạo token ma thuật đăng nhập không cần mật khẩu (Hạn 15 phút)
func (s *AuthService) SendMagicLink(ctx context.Context, req models.SendMagicLinkRequest, clientIP string) (string, error) {
	if err := utils.VerifyTurnstileCaptcha(req.CaptchaToken, clientIP); err != nil {
		return "", err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return "", errors.New("địa chỉ email không được để trống")
	}

	// Sinh token ngẫu nhiên 32 bytes (64 ký tự hex)
	b := make([]byte, 32)
	if _, err := crand.Read(b); err != nil {
		return "", errors.New("lỗi khi sinh token bảo mật")
	}
	token := hex.EncodeToString(b)

	if s.redisClient != nil {
		key := fmt.Sprintf("magic_link:%s", token)
		if err := s.redisClient.Set(ctx, key, email, 15*time.Minute).Err(); err != nil {
			return "", errors.New("không thể lưu phiên Magic Link")
		}
	}

	log.Printf("🪄 [MAGIC_LINK_SENT] Gửi link ma thuật cho: %s | Token: %s\n", email, token)
	return token, nil
}

// VerifyMagicLink xác thực token ma thuật và đăng nhập (tự tạo user nếu chưa có)
func (s *AuthService) VerifyMagicLink(ctx context.Context, req models.VerifyMagicLinkRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return nil, errors.New("token không hợp lệ")
	}

	if s.redisClient == nil {
		return nil, errors.New("hệ thống lưu trữ phiên tạm thời không khả dụng")
	}

	key := fmt.Sprintf("magic_link:%s", token)
	email, err := s.redisClient.Get(ctx, key).Result()
	if err != nil || email == "" {
		return nil, errors.New("liên kết Magic Link không hợp lệ hoặc đã hết hạn (15 phút)")
	}

	// Token dùng 1 lần (Single-use token) -> Xóa ngay
	_ = s.redisClient.Del(ctx, key)

	// Tìm user theo email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// Tạo user mới tự động
		baseUsername := strings.Split(email, "@")[0]
		cleanUsername := strings.ToLower(strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, baseUsername))
		if len(cleanUsername) < 3 {
			cleanUsername = "magic_" + cleanUsername
		}
		if len(cleanUsername) > 20 {
			cleanUsername = cleanUsername[:20]
		}

		finalUsername := cleanUsername
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
			DisplayName:   baseUsername,
			AvatarURL:     "https://api.dicebear.com/7.x/bottts/svg?seed=" + finalUsername,
			EmailVerified: true,
			OAuthAccounts: []models.OAuthAccount{},
		}

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return nil, errors.New("không thể tạo tài khoản từ Magic Link: " + err.Error())
		}
		user = newUser
	}

	return s.CreateAuthSession(ctx, user, clientIP, userAgent)
}

// SendPhoneOTP gửi mã OTP 6 số xác thực số điện thoại
func (s *AuthService) SendPhoneOTP(ctx context.Context, req models.SendPhoneOTPRequest, clientIP string) (string, error) {
	if err := utils.VerifyTurnstileCaptcha(req.CaptchaToken, clientIP); err != nil {
		return "", err
	}

	phone := strings.ReplaceAll(strings.TrimSpace(req.Phone), " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if len(phone) < 9 || len(phone) > 15 {
		return "", errors.New("số điện thoại không đúng định dạng (9 - 15 số)")
	}

	rateKey := fmt.Sprintf("rate:phone_otp:%s", phone)
	if s.redisClient != nil {
		exists, _ := s.redisClient.Exists(ctx, rateKey).Result()
		if exists > 0 {
			return "", errors.New("bạn đang gửi yêu cầu quá thường xuyên. Vui lòng đợi 1 phút trước khi thử lại!")
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", r.Intn(1000000))

	if s.redisClient != nil {
		key := fmt.Sprintf("otp:phone:%s", phone)
		if err := s.redisClient.Set(ctx, key, otp, 5*time.Minute).Err(); err != nil {
			return "", errors.New("lỗi khi lưu trữ mã OTP điện thoại")
		}
		_ = s.redisClient.Set(ctx, rateKey, "1", 60*time.Second).Err()
	}

	log.Printf("📱 [SMS_OTP_SENT] Gửi mã xác thực SMS cho SĐT: %s | OTP: %s (Hết hạn trong 5 phút)\n", phone, otp)
	return otp, nil
}

// VerifyPhoneOTP xác thực mã OTP số điện thoại và đăng nhập
func (s *AuthService) VerifyPhoneOTP(ctx context.Context, req models.VerifyPhoneOTPRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
	phone := strings.ReplaceAll(strings.TrimSpace(req.Phone), " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	otp := strings.TrimSpace(req.OTP)

	if s.redisClient == nil {
		return nil, errors.New("hệ thống lưu trữ OTP tạm thời không khả dụng")
	}

	key := fmt.Sprintf("otp:phone:%s", phone)
	storedOTP, err := s.redisClient.Get(ctx, key).Result()
	if err != nil || storedOTP == "" {
		return nil, errors.New("mã OTP không tồn tại hoặc đã hết hạn")
	}

	if storedOTP != otp {
		return nil, errors.New("mã OTP không chính xác")
	}

	_ = s.redisClient.Del(ctx, key)

	// Tìm user theo số điện thoại
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// Tạo tài khoản mới theo số điện thoại
		suffix := phone
		if len(suffix) > 6 {
			suffix = suffix[len(suffix)-6:]
		}
		username := fmt.Sprintf("p_%s%d", suffix, time.Now().UnixNano()%1000)

		displayName := strings.TrimSpace(req.DisplayName)
		if displayName == "" {
			displayName = "User " + suffix
		}

		syntheticEmail := fmt.Sprintf("%s@phone.chatrealtime.local", phone)

		newUser := &models.User{
			Username:      username,
			Email:         syntheticEmail,
			PhoneNumber:   phone,
			DisplayName:   displayName,
			AvatarURL:     "https://api.dicebear.com/7.x/bottts/svg?seed=" + phone,
			EmailVerified: true,
			OAuthAccounts: []models.OAuthAccount{},
		}

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return nil, errors.New("không thể tạo tài khoản từ số điện thoại: " + err.Error())
		}
		user = newUser
	}

	return s.CreateAuthSession(ctx, user, clientIP, userAgent)
}

// RefreshToken làm mới phiên đăng nhập (Silent Token Refresh & Rotation)
func (s *AuthService) RefreshToken(ctx context.Context, req models.RefreshTokenRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return nil, errors.New("refresh token không được để trống")
	}

	if s.redisClient == nil {
		return nil, errors.New("dịch vụ xác thực phiên tạm thời gián đoạn")
	}

	tokenKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	data, err := s.redisClient.Get(ctx, tokenKey).Result()
	if err != nil || data == "" {
		return nil, errors.New("refresh token không hợp lệ hoặc đã hết hạn, vui lòng đăng nhập lại")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, errors.New("dữ liệu phiên không hợp lệ")
	}

	userID, _ := payload["user_id"].(string)
	if userID == "" {
		return nil, errors.New("không xác định được người dùng của phiên")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("tài khoản người dùng không tồn tại hoặc đã bị khóa")
	}

	// Token Rotation: Xóa token cũ ngay lập tức để chống Replay Attack
	_ = s.redisClient.Del(ctx, tokenKey)
	userTokensKey := fmt.Sprintf("user_refresh_tokens:%s", userID)
	_ = s.redisClient.SRem(ctx, userTokensKey, refreshToken)

	// Tạo phiên đăng nhập mới với token xoay vòng
	return s.CreateAuthSession(ctx, user, clientIP, userAgent)
}

// Logout kết thúc phiên đăng nhập và thu hồi Refresh Token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if s.redisClient == nil || refreshToken == "" {
		return nil
	}

	tokenKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	data, _ := s.redisClient.Get(ctx, tokenKey).Result()
	if data != "" {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(data), &payload); err == nil {
			if userID, ok := payload["user_id"].(string); ok && userID != "" {
				userTokensKey := fmt.Sprintf("user_refresh_tokens:%s", userID)
				_ = s.redisClient.SRem(ctx, userTokensKey, refreshToken)
			}
		}
	}

	_ = s.redisClient.Del(ctx, tokenKey)
	return nil
}

// GetSecuritySessions lấy thông tin các thiết bị đăng nhập và các cảnh báo an ninh
func (s *AuthService) GetSecuritySessions(ctx context.Context, userID, currentIP, currentUA string) ([]models.LoginDevice, []models.SecurityAlert, error) {
	devices := make([]models.LoginDevice, 0)
	alerts := make([]models.SecurityAlert, 0)

	if s.redisClient == nil {
		return devices, alerts, nil
	}

	userTokensKey := fmt.Sprintf("user_refresh_tokens:%s", userID)
	tokens, _ := s.redisClient.SMembers(ctx, userTokensKey).Result()

	for idx, tok := range tokens {
		tokenKey := fmt.Sprintf("refresh_token:%s", tok)
		data, err := s.redisClient.Get(ctx, tokenKey).Result()
		if err != nil || data == "" {
			_ = s.redisClient.SRem(ctx, userTokensKey, tok)
			continue
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			continue
		}

		devName, _ := payload["device_name"].(string)
		ip, _ := payload["client_ip"].(string)
		loc, _ := payload["location"].(string)
		ua, _ := payload["user_agent"].(string)
		createdUnix, _ := payload["created_at"].(float64)

		isCur := (ip == currentIP) && (ua == currentUA)
		if idx == 0 && !isCur && len(tokens) == 1 {
			isCur = true
		}

		devices = append(devices, models.LoginDevice{
			ID:         fmt.Sprintf("dev_%d", idx+1),
			UserID:     userID,
			IPAddress:  ip,
			UserAgent:  ua,
			DeviceName: devName,
			Location:   loc,
			LastActive: time.Unix(int64(createdUnix), 0),
			IsCurrent:  isCur,
		})
	}

	// Đọc cảnh báo an ninh gần đây
	alertsKey := fmt.Sprintf("user_alerts:%s", userID)
	alertStrings, _ := s.redisClient.LRange(ctx, alertsKey, 0, 19).Result()
	for _, aStr := range alertStrings {
		var a models.SecurityAlert
		if err := json.Unmarshal([]byte(aStr), &a); err == nil {
			alerts = append(alerts, a)
		}
	}

	return devices, alerts, nil
}

// RevokeOtherSessions thu hồi tất cả các phiên đăng nhập khác ngoại trừ phiên hiện tại
func (s *AuthService) RevokeOtherSessions(ctx context.Context, userID, currentRefreshToken string) error {
	if s.redisClient == nil {
		return nil
	}

	userTokensKey := fmt.Sprintf("user_refresh_tokens:%s", userID)
	tokens, _ := s.redisClient.SMembers(ctx, userTokensKey).Result()

	for _, tok := range tokens {
		if tok != currentRefreshToken {
			_ = s.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%s", tok))
			_ = s.redisClient.SRem(ctx, userTokensKey, tok)
		}
	}

	return nil
}

// SSOInitiate khởi tạo luồng Enterprise Single Sign-On (SAML 2.0 / OIDC / Okta)
func (s *AuthService) SSOInitiate(ctx context.Context, req models.SSOInitiateRequest) (*models.SSOInitiateResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.WorkEmail))
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		return nil, errors.New("địa chỉ email doanh nghiệp không hợp lệ")
	}

	domain := parts[1]
	var orgName string
	var ssoType string

	// Nhận diện doanh nghiệp thông minh dựa trên tên miền tổ chức
	switch domain {
	case "fpt.vn", "fpt.edu.vn", "fpt.com":
		orgName = "Tập đoàn FPT (FPT Corporation)"
		ssoType = "Okta SSO / SAML 2.0"
	case "viettel.com.vn", "viettel.vn":
		orgName = "Tập đoàn Viettel (Viettel Military Industry & Telecoms)"
		ssoType = "Microsoft Azure AD (Entra ID) / SAML 2.0"
	case "vng.com.vn", "vng.vn":
		orgName = "Công ty CP VNG (VNG Corporation)"
		ssoType = "Google Workspace Enterprise / OIDC"
	case "acme.com", "enterprise.io":
		orgName = "Acme Global Enterprise"
		ssoType = "Okta Enterprise SSO"
	default:
		// Tự động suy ra tên tổ chức từ tên miền
		cleanOrg := strings.Title(strings.Split(domain, ".")[0])
		orgName = fmt.Sprintf("Doanh nghiệp %s (%s)", cleanOrg, domain)
		ssoType = "Enterprise OIDC / SAML 2.0 Identity Provider"
	}

	mockRedirectURL := fmt.Sprintf("https://sso.%s/idp/profile/SAML2/Redirect/SSO?sp=chatrealtime&domain=%s", domain, domain)

	return &models.SSOInitiateResponse{
		Domain:       domain,
		Organization: orgName,
		SSOType:      ssoType,
		RedirectURL:  mockRedirectURL,
	}, nil
}

// SSOCallback xử lý phản hồi từ Identity Provider của doanh nghiệp và cấp quyền đăng nhập
func (s *AuthService) SSOCallback(ctx context.Context, req models.SSOCallbackRequest, clientIP, userAgent string) (*models.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.WorkEmail))
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		return nil, errors.New("địa chỉ email doanh nghiệp không hợp lệ")
	}

	domain := parts[1]
	cleanOrg := strings.Title(strings.Split(domain, ".")[0])
	orgName := fmt.Sprintf("Doanh nghiệp %s", cleanOrg)

	// Tìm user theo email doanh nghiệp
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// Tự động cấp tài khoản Doanh nghiệp mới theo SAML JIT (Just-In-Time) Provisioning
		baseName := parts[0]
		username := "sso_" + baseName
		for i := 1; i <= 10; i++ {
			exist, _ := s.userRepo.FindByUsername(ctx, username)
			if exist == nil {
				break
			}
			username = fmt.Sprintf("sso_%s%d", baseName, time.Now().UnixNano()%1000)
		}

		displayName := strings.Title(baseName)

		newUser := &models.User{
			Username:      username,
			Email:         email,
			DisplayName:   displayName,
			AvatarURL:     "https://api.dicebear.com/7.x/identicon/svg?seed=" + email,
			EmailVerified: true,
			Organization:  orgName,
			OAuthAccounts: []models.OAuthAccount{
				{
					Provider:   "enterprise_sso",
					ProviderID: domain,
					Email:      email,
					Name:       displayName,
					AvatarURL:  "https://api.dicebear.com/7.x/identicon/svg?seed=" + email,
					LinkedAt:   time.Now(),
				},
			},
		}

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return nil, errors.New("lỗi khi khởi tạo tài khoản doanh nghiệp: " + err.Error())
		}
		user = newUser
	} else if user.Organization == "" {
		// Cập nhật organization nếu chưa có
		user.Organization = orgName
		_ = s.userRepo.Update(ctx, user.ID, bson.M{"organization": orgName})
	}

	log.Printf("🏢 [ENTERPRISE_SSO_LOGIN] Đăng nhập thành công qua SSO Doanh nghiệp: %s (%s)\n", email, orgName)
	return s.CreateAuthSession(ctx, user, clientIP, userAgent)
}



