package main

import (
	"log"
	"os"

	"chatrealtime-backend/internal/database"
	"chatrealtime-backend/internal/handlers"
	"chatrealtime-backend/internal/middleware"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/internal/service"
	"chatrealtime-backend/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Nạp biến môi trường
	if err := godotenv.Load(); err != nil {
		log.Println("Chú ý: Sử dụng biến môi trường hệ thống")
	}

	// 2. Khởi tạo Database (MongoDB & Redis)
	database.InitDatabase()

	// 3. Khởi tạo các Repositories
	userRepo := repository.NewUserRepository(database.MongoDB)
	msgRepo := repository.NewMessageRepository(database.MongoDB)
	convRepo := repository.NewConversationRepository(database.MongoDB)
	groupRepo := repository.NewGroupRepository(database.MongoDB)

	// 4. Khởi tạo WebSocket Hub và chạy ngầm
	hub := websocket.NewHub(database.RedisClient, msgRepo, convRepo, groupRepo, userRepo)
	go hub.Run()

	// 5. Khởi tạo Services
	authService := service.NewAuthService(userRepo)
	chatService := service.NewChatService(msgRepo, convRepo, userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)

	// 6. Khởi tạo Handlers
	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService)
	groupHandler := handlers.NewGroupHandler(groupService, msgRepo)
	wsHandler := handlers.NewWSHandler(hub)
	uploadHandler := handlers.NewUploadHandler()

	// 7. Khởi tạo Router với cấu hình CORS đầy đủ
	r := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	// Phục vụ file tĩnh đã tải lên (ảnh, voice, tài liệu)
	r.Static("/uploads", "./uploads")

	// Endpoint WebSocket
	r.GET("/ws", wsHandler.HandleWS)

	// REST API Routes
	api := r.Group("/api")
	{
		// Upload File (ảnh, tệp đính kèm, voice note)
		api.POST("/upload", middleware.AuthMiddleware(), uploadHandler.UploadFile)

		// Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.GetMe)
		}

		// Chat Routes (Cần xác thực Token)
		chat := api.Group("/chat", middleware.AuthMiddleware())
		{
			chat.GET("/conversations", chatHandler.GetConversations)
			chat.GET("/messages/:conversation_id", chatHandler.GetMessages)
			chat.POST("/messages/:conversation_id/read", chatHandler.MarkAsRead)
			chat.GET("/users", chatHandler.GetUsers)
		}

		// Group Routes (Cần xác thực Token)
		groups := api.Group("/groups", middleware.AuthMiddleware())
		{
			groups.POST("", groupHandler.Create)
			groups.GET("", groupHandler.GetMyGroups)
			groups.GET("/:id", groupHandler.GetDetails)
			groups.POST("/:id/members", groupHandler.AddMembers)
			groups.DELETE("/:id/members/:userId", groupHandler.RemoveMember)
			groups.GET("/:id/messages", groupHandler.GetGroupMessages)
		}
	}



	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server Chat Realtime đang chạy tại http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Không thể khởi động server: %v", err)
	}
}
