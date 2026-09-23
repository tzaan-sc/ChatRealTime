package main

import (
	"context"
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
	scheduledRepo := repository.NewScheduledRepository(database.MongoDB)
	reminderRepo := repository.NewReminderRepository(database.MongoDB)
	draftRepo := repository.NewDraftRepository(database.MongoDB)
	inviteRepo := repository.NewInviteRepository(database.MongoDB)
	joinReqRepo := repository.NewJoinRequestRepository(database.MongoDB)
	pollRepo := repository.NewPollRepository(database.MongoDB)
	eventRepo := repository.NewEventRepository(database.MongoDB)

	// Gán pollRepo và eventRepo vào msgRepo để tự động populate khi đọc tin nhắn
	msgRepo.SetPollAndEventRepos(pollRepo, eventRepo)

	// 4. Khởi tạo WebSocket Hub và chạy ngầm
	hub := websocket.NewHub(database.RedisClient, msgRepo, convRepo, groupRepo, userRepo, draftRepo)
	go hub.Run()

	// 5. Khởi tạo Services
	authService := service.NewAuthService(userRepo)
	chatService := service.NewChatService(msgRepo, convRepo, userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo, inviteRepo, joinReqRepo, pollRepo, eventRepo, msgRepo, database.RedisClient)

	// Khởi tạo Background Scheduler Service (quét tin hẹn giờ & nhắc việc)
	schedulerService := service.NewSchedulerService(database.RedisClient, msgRepo, convRepo, scheduledRepo, reminderRepo)
	schedulerService.Start(context.Background())

	// 6. Khởi tạo Handlers
	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService, msgRepo, scheduledRepo, reminderRepo, draftRepo, userRepo)
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

		// Invite Links preview & join
		api.GET("/invites/:code/preview", middleware.AuthMiddleware(), groupHandler.PreviewInvite)
		api.POST("/invites/:code/join", middleware.AuthMiddleware(), groupHandler.JoinViaInvite)

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
			chat.POST("/messages/:conversation_id/unread", chatHandler.MarkAsUnread)
			chat.GET("/users", chatHandler.GetUsers)

			// Advanced Messaging Routes
			chat.GET("/pinned/:target_id", chatHandler.GetPinnedMessages)
			chat.GET("/messages/:id/thread", chatHandler.GetThreadMessages)
			chat.POST("/scheduled", chatHandler.CreateScheduledMessage)
			chat.GET("/scheduled", chatHandler.GetScheduledMessages)
			chat.DELETE("/scheduled/:id", chatHandler.CancelScheduledMessage)
			chat.POST("/reminders", chatHandler.CreateReminder)
			chat.GET("/reminders", chatHandler.GetReminders)
			chat.DELETE("/reminders/:id", chatHandler.DismissReminder)
			chat.GET("/drafts", chatHandler.GetDrafts)
			chat.POST("/drafts", chatHandler.SaveDraft)
			chat.DELETE("/drafts/:target_id", chatHandler.DeleteDraft)
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
			groups.PATCH("/:id/slowmode", groupHandler.UpdateSlowMode)
			groups.POST("/:id/categories", groupHandler.CreateCategory)
			groups.DELETE("/:id/categories/:catId", groupHandler.DeleteCategory)
			groups.POST("/:id/channels", groupHandler.CreateChannel)
			groups.DELETE("/:id/channels/:chanId", groupHandler.DeleteChannel)

			// Phase 2: RBAC, Invites & Approvals
			groups.PATCH("/:id/members/:userId/role", groupHandler.UpdateMemberRole)
			groups.POST("/:id/members/:userId/mute", groupHandler.MuteMember)
			groups.PATCH("/:id/settings", groupHandler.UpdateSettings)
			groups.POST("/:id/invites", groupHandler.CreateInvite)
			groups.GET("/:id/invites", groupHandler.GetGroupInvites)
			groups.DELETE("/:id/invites/:code", groupHandler.RevokeInvite)
			groups.GET("/:id/join-requests", groupHandler.GetPendingJoinRequests)
			groups.POST("/:id/join-requests/:requestId/approve", groupHandler.ApproveJoinRequest)
			groups.POST("/:id/join-requests/:requestId/reject", groupHandler.RejectJoinRequest)

			// Phase 3: Interactive Polls & Group Events
			groups.POST("/:id/polls", groupHandler.CreatePoll)
			groups.GET("/:id/polls/:pollId", groupHandler.GetPoll)
			groups.POST("/:id/polls/:pollId/vote", groupHandler.VotePoll)
			groups.POST("/:id/polls/:pollId/close", groupHandler.ClosePoll)
			groups.POST("/:id/events", groupHandler.CreateEvent)
			groups.GET("/:id/events", groupHandler.GetGroupEvents)
			groups.POST("/:id/events/:eventId/rsvp", groupHandler.RSVPEvent)
			groups.DELETE("/:id/events/:eventId", groupHandler.DeleteEvent)

			// Phase 4: Group Analytics & Insights
			groups.GET("/:id/analytics", groupHandler.GetGroupAnalytics)
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
