package middleware

import (
	"net/http"
	"os"
	"strings"

	"chatrealtime-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware chặn và kiểm tra token trong Header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu có token xác thực"})
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng token phải là 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		jwtSecret := os.Getenv("JWT_SECRET")
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		// Lưu thông tin User vào Context để các Handler sau có thể đọc được
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
