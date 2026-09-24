package handlers

import (
	"net/http"

	"chatrealtime-backend/internal/models"
	"chatrealtime-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handler
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu gửi lên không hợp lệ: " + err.Error()})
		return
	}

	res, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Đăng ký tài khoản thành công",
		"data":    res,
	})
}

// Login handler
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu gửi lên không hợp lệ: " + err.Error()})
		return
	}

	res, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"data":    res,
	})
}

// GetMe handler (API được bảo vệ bởi Token)
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy phiên đăng nhập"})
		return
	}

	user, err := h.authService.GetUserProfile(c.Request.Context(), userID.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// OAuthLogin handler đăng nhập qua mạng xã hội
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	var req models.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu OAuth không hợp lệ: " + err.Error()})
		return
	}

	res, err := h.authService.OAuthLogin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xác thực mạng xã hội thành công",
		"data":    res,
	})
}

// LinkOAuth handler liên kết mạng xã hội vào tài khoản hiện tại
func (h *AuthHandler) LinkOAuth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy phiên đăng nhập"})
		return
	}

	var req models.LinkOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu liên kết không hợp lệ: " + err.Error()})
		return
	}

	if err := h.authService.LinkOAuth(c.Request.Context(), userID.(string), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Liên kết tài khoản mạng xã hội thành công",
	})
}

// UnlinkOAuth handler hủy liên kết mạng xã hội
func (h *AuthHandler) UnlinkOAuth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy phiên đăng nhập"})
		return
	}

	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin provider"})
		return
	}

	if err := h.authService.UnlinkOAuth(c.Request.Context(), userID.(string), provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gỡ liên kết tài khoản thành công",
	})
}

// GetProviders trả về danh sách các Provider hỗ trợ
func (h *AuthHandler) GetProviders(c *gin.Context) {
	providers := []gin.H{
		{"id": "google", "name": "Google", "icon": "google", "enabled": true},
		{"id": "facebook", "name": "Facebook", "icon": "facebook", "enabled": true},
		{"id": "github", "name": "GitHub", "icon": "github", "enabled": true},
		{"id": "apple", "name": "Apple", "icon": "apple", "enabled": true},
		{"id": "discord", "name": "Discord", "icon": "discord", "enabled": true},
	}
	c.JSON(http.StatusOK, gin.H{"data": providers})
}

// ForgotPassword gửi mã OTP khôi phục mật khẩu qua email
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email không hợp lệ: " + err.Error()})
		return
	}

	otp, err := h.authService.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Mã xác thực khôi phục mật khẩu đã được gửi về email của bạn (hết hạn trong 15 phút)",
		"demo_otp": otp, // Hỗ trợ dev/test nhanh khi chưa gắn server SMTP thực
	})
}

// ResetPassword đặt lại mật khẩu mới với mã OTP
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đặt lại mật khẩu không hợp lệ: " + err.Error()})
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đặt lại mật khẩu thành công! Bạn có thể đăng nhập bằng mật khẩu mới.",
	})
}

// SendVerificationEmail gửi mã kích hoạt email
func (h *AuthHandler) SendVerificationEmail(c *gin.Context) {
	var req models.SendVerificationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email không hợp lệ: " + err.Error()})
		return
	}

	otp, err := h.authService.SendVerificationEmail(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Mã kích hoạt tài khoản đã được gửi tới email của bạn (hết hạn trong 24 giờ)",
		"demo_otp": otp,
	})
}

// VerifyEmail kích hoạt tài khoản với mã OTP
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu xác thực không hợp lệ: " + err.Error()})
		return
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xác thực địa chỉ email thành công! Tài khoản của bạn đã được kích hoạt hoàn toàn.",
	})
}


