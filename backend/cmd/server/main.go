package main

import (
	"log"
	"os"

	"chatrealtime-backend/internal/database"
	"chatrealtime-backend/internal/handlers"
	"chatrealtime-backend/internal/middleware"
	"chatrealtime-backend/internal/repository"
	"chatrealtime-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Nạp file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Chú ý: Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	// 2. Khởi tạo Database (MongoDB & Redis)
	database.InitDatabase()

	// 3. Khởi tạo các tầng phụ thuộc (Dependency Injection)
	userRepo := repository.NewUserRepository(database.MongoDB)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// 4. Khởi tạo Gin Router
	r := gin.Default()

	// Cấu hình CORS để Frontend Svelte gọi API mà không bị chặn
	r.Use(cors.Default())

	// 5. Khai báo các Routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			// Route cần có Token
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.GetMe)
		}
	}

	// 6. Chạy Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server đang lắng nghe tại cổng http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Không thể khởi động server: %v", err)
	}
}
