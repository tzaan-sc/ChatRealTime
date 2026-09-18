package handlers

import (
	"log"
	"net/http"
	"os"

	"chatrealtime-backend/internal/websocket"
	"chatrealtime-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Cho phép mọi domain kết nối WebSocket (tránh bị chặn CORS)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	hub *websocket.Hub
}

func NewWSHandler(hub *websocket.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleWS nâng cấp kết nối HTTP lên WebSocket
func (h *WSHandler) HandleWS(c *gin.Context) {
	// Lấy token từ Query Parameter: ws://localhost:8080/ws?token=<jwt_token>
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu có token"})
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	claims, err := utils.ValidateToken(tokenString, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Lỗi nâng cấp kết nối WebSocket: %v", err)
		return
	}

	client := &websocket.Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   claims.UserID,
		Username: claims.Username,
	}

	// Đăng ký client vào Hub
	h.hub.RegisterClient(client)

	// Bật 2 goroutines song song để đọc & ghi
	go client.WritePump()
	go client.ReadPump()
}
