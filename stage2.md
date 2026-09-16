# Hướng Dẫn Chi Tiết Giai Đoạn 2: Khởi Tạo Backend Golang & Xác Thực (Auth)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Ngôn ngữ: Go (Golang) | Database: MongoDB & Redis

Tài liệu này hướng dẫn bạn từ cài đặt Go, xây dựng khung dự án Backend theo mô hình **Clean Architecture**, kết nối với MongoDB/Redis đã tạo ở Giai đoạn 1, và hoàn thiện bộ API Đăng ký, Đăng nhập và Xác thực người dùng bằng **JWT Token**.

---

## Mục lục
1. [Bước 1: Cài đặt môi trường Golang trên Windows](#buoc-1-cai-dat-moi-truong-golang-tren-windows)
2. [Bước 2: Khởi tạo Go Module & Cấu trúc thư mục Clean Architecture](#buoc-2-khoi-tao-go-module--cau-truc-thu-muc-clean-architecture)
3. [Bước 3: Cài đặt các thư viện cần thiết](#buoc-3-cai-dat-cac-thu-vien-can-thiet)
4. [Bước 4: Viết mã nguồn chi tiết từng file](#buoc-4-viet-ma-nguon-chi-tiet-tung-file)
   - [4.1. File cấu hình môi trường `.env`](#41-file-cau-hinh-moi-truong-env)
   - [4.2. Khởi tạo Models (`internal/models/user.go`)](#42-khoi-tao-models)
   - [4.3. Tiện ích Mật khẩu & JWT (`pkg/utils/`)](#43-tien-ich-mat-khau--jwt)
   - [4.4. Kết nối Database (`internal/database/db.go`)](#44-ket-noi-database)
   - [4.5. Tầng Repository (`internal/repository/user_repo.go`)](#45-tang-repository)
   - [4.6. Tầng Service (`internal/service/auth_service.go`)](#46-tang-service)
   - [4.7. Tầng Handlers & Middleware (`internal/handlers/`, `internal/middleware/`)](#47-tang-handlers--middleware)
   - [4.8. File chạy chính (`cmd/server/main.go`)](#48-file-chay-chinh)
5. [Bước 5: Chạy Server và Kiểm thử API (Postman / cURL)](#buoc-5-chay-server-va-kiem-thu-api)
6. [Các lỗi thường gặp và cách xử lý](#cac-loi-thuong-gap-va-cach-xu-ly)

---

<a name="buoc-1-cai-dat-moi-truong-golang-tren-windows"></a>
## Bước 1: Cài đặt môi trường Golang trên Windows

### 1.1. Cài đặt Go
Mở **PowerShell** và gõ lệnh sau để cài đặt Go chính thức qua `winget`:
```powershell
winget install -e --id GoLang.Go
```
*(Hoặc tải bộ cài đặt `.msi` từ trang chủ: [https://go.dev/dl/](https://go.dev/dl/))*.

### 1.2. Kiểm tra cài đặt
Sau khi cài đặt xong, hãy **tắt cửa sổ PowerShell cũ và mở một cửa sổ mới** (để cập nhật biến môi trường), sau đó gõ:
```powershell
go version
# Kết quả mong đợi: go version go1.22.x (hoặc mới hơn) windows/amd64
```

---

<a name="buoc-2-khoi-tao-go-module--cau-truc-thu-muc-clean-architecture"></a>
## Bước 2: Khởi tạo Go Module & Cấu trúc thư mục Clean Architecture

Chúng ta sẽ tạo một thư mục riêng tên là `backend` trong dự án `ChatRealTime` để quản lý mã nguồn Go độc lập với Frontend Svelte.

### 2.1. Tạo thư mục backend và khởi tạo Go module
Mở Terminal/PowerShell tại `d:\GIT\ChatRealTime`:
```powershell
mkdir backend
cd backend
go mod init chatrealtime-backend
```

### 2.2. Tạo cấu trúc thư mục chuẩn
Chạy các lệnh tạo thư mục:
```powershell
mkdir cmd\server
mkdir internal\config
mkdir internal\database
mkdir internal\models
mkdir internal\repository
mkdir internal\service
mkdir internal\handlers
mkdir internal\middleware
mkdir pkg\utils
```

Cấu trúc thư mục sau khi tạo sẽ như sau:
```text
backend/
├── cmd/
│   └── server/
│       └── main.go          # Điểm khởi chạy chương trình (Entry point)
├── internal/
│   ├── database/            # Khởi tạo kết nối MongoDB và Redis
│   ├── handlers/            # Tiếp nhận HTTP Request và trả về JSON Response
│   ├── middleware/          # Kiểm tra JWT Token bảo vệ API
│   ├── models/              # Khai báo cấu trúc dữ liệu (User, Message,...)
│   ├── repository/          # Thao tác trực tiếp với cơ sở dữ liệu MongoDB
│   └── service/             # Logic nghiệp vụ (Mã hóa mật khẩu, kiểm tra trùng lặp)
├── pkg/
│   └── utils/               # Hàm tiện ích dùng chung (JWT, Hash password)
├── .env                     # Lưu biến môi trường (Mật khẩu, Port, Secret key)
├── go.mod                   # Quản lý phiên bản Go và các thư viện
└── go.sum                   # Checksum các thư viện
```

---

<a name="buoc-3-cai-dat-cac-thu-vien-can-thiet"></a>
## Bước 3: Cài đặt các thư viện cần thiết

Trong thư mục `backend`, chạy các lệnh sau để tải các thư viện chuẩn về dự án:

```powershell
# 1. Gin Web Framework (Xây dựng REST API siêu nhanh)
go get -u github.com/gin-gonic/gin

# 2. Driver chính thức của MongoDB
go get -u go.mongodb.org/mongo-driver/mongo

# 3. Driver Redis v9
go get -u github.com/redis/go-redis/v9

# 4. Thư viện mã hóa mật khẩu Bcrypt
go get -u golang.org/x/crypto/bcrypt

# 5. Thư viện JSON Web Token (JWT v5)
go get -u github.com/golang-jwt/jwt/v5

# 6. Thư viện đọc file cấu hình .env
go get -u github.com/joho/godotenv
```

---

<a name="buoc-4-viet-ma-nguon-chi-tiet-tung-file"></a>

#### Bước 4: Viết mã nguồn chi tiết từng file (Đăng ký, đăng nhập, lấy thông tin qua jwt)

### 4.1. File cấu hình môi trường `.env`
Tạo file `backend/.env`:
```ini
PORT=8080
MONGO_URI=mongodb://admin:password123@localhost:27017
DB_NAME=chatapp
REDIS_ADDR=localhost:6379
JWT_SECRET=super_secret_key_chatapp_2026
```

---

### 4.2. Khởi tạo Models (Định nghĩa cấu trúc dữ liệu User lưu trong MongoDB)
Tạo file `backend/internal/models/user.go`:
```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User đại diện cho một tài khoản trong collection "users"
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username    string             `bson:"username" json:"username"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"-"` // Dấu "-" để không bao giờ lộ password ra JSON
	DisplayName string             `bson:"display_name" json:"display_name"`
	AvatarURL   string             `bson:"avatar_url" json:"avatar_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
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

// Response DTO sau khi Đăng nhập thành công
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
```

---

### 4.3. Tiện ích bảo mật (Mật khẩu & JWT) (Băm mật khẩu bằng Bcrypt, tạo và giải mã JWT Token)

#### File `backend/pkg/utils/password.go`:
```go
package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword băm mật khẩu bằng thuật toán Bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash so sánh mật khẩu người dùng nhập với chuỗi băm trong DB
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

#### File `backend/pkg/utils/jwt.go`:
```go
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken tạo ra chuỗi JWT có thời hạn 7 ngày
func GenerateToken(userID, username, secretKey string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateToken giải mã và xác thực token gửi lên từ client
func ValidateToken(tokenString, secretKey string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token không hợp lệ")
}
```

---

### 4.4. Kết nối Database (Quản lý kết nối tới MongoDB và Redis)
Tạo file `backend/internal/database/db.go`:
```go
package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
	RedisClient *redis.Client
)

// InitDatabase kết nối cả MongoDB và Redis
func InitDatabase() {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	redisAddr := os.Getenv("REDIS_ADDR")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Kết nối MongoDB
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Lỗi kết nối MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Không thể ping tới MongoDB: %v", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)
	log.Println(" Kết nối MongoDB thành công!")

	// 2. Kết nối Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Không thể ping tới Redis: %v", err)
	}
	log.Println(" Kết nối Redis thành công!")
}
```

---

### 4.5. Tầng Repository (Tầng truy vấn DB: Thêm user mới, tìm user theo username/email/ID)
Tạo file `backend/internal/repository/user_repo.go`:
```go
package repository

import (
	"context"
	"errors"
	"time"

	"chatrealtime-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create chèn user mới vào MongoDB
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByUsername tìm user theo tên đăng nhập
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail tìm user theo email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID tìm user theo ObjectID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

---

### 4.6. Tầng Service (Tầng nghiệp vụ: Đăng ký, Đăng nhập, kiểm tra logic)
Tạo file `backend/internal/service/auth_service.go`:
```go
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
```

---

### 4.7. Tầng Handlers (Nhận HTTP request, trả JSON) & Middleware (Chặn request chưa đăng nhập & Xác thực Token)

#### Middleware xác thực JWT: `backend/internal/middleware/auth_middleware.go`
```go
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

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
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
```

#### Handler tiếp nhận Request: `backend/internal/handlers/auth_handler.go`
```go
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
```

---

### 4.8. File chạy chính (File chạy chính, định tuyến router và khởi động HTTP Server trên cổng 8080)
Tạo file `backend/cmd/server/main.go`:
```go
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
```

> **Lưu ý cài thêm cors:**
> Chạy thêm lệnh: `go get github.com/gin-contrib/cors` để cài thư viện xử lý CORS.

---

<a name="buoc-5-chay-server-va-kiem-thu-api"></a>
## Bước 5: Chạy Server và Kiểm thử API

### 5.1. Khởi động Server
Tại thư mục `backend`, chạy:
```powershell
go run cmd/server/main.go
```
Khi thấy dòng:
```text
 Kết nối MongoDB thành công!
 Kết nối Redis thành công!
🚀 Server đang lắng nghe tại cổng http://localhost:8080
```
$\to$ Server Go của bạn đã sẵn sàng nhận kết nối!

---

### 5.2. Kiểm thử API bằng cURL / Postman / Thunder Client

#### 1. Đăng ký tài khoản mới (`POST http://localhost:8080/api/auth/register`):
* **Body (JSON):**
  ```json
  {
    "username": "alex",
    "email": "alex@gmail.com",
    "password": "password123",
    "display_name": "Alex Nguyen"
  }
  ```
* **Kết quả trả về (201 Created):**
  ```json
  {
    "message": "Đăng ký tài khoản thành công",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "user": {
        "id": "65f1c9...",
        "username": "alex",
        "email": "alex@gmail.com",
        "display_name": "Alex Nguyen",
        "avatar_url": "https://api.dicebear.com/7.x/bottts/svg?seed=alex"
      }
    }
  }
  ```

#### 2. Đăng nhập (`POST http://localhost:8080/api/auth/login`):
* **Body (JSON):**
  ```json
  {
    "username": "alex",
    "password": "password123"
  }
  ```
* Nhận về chuỗi `token`.

#### 3. Lấy thông tin cá nhân (`GET http://localhost:8080/api/auth/me`):
* **Headers:**
  * `Authorization`: `Bearer <chuỗi_token_vừa_nhận_ở_trên>`
* **Kết quả:** Trả về đầy đủ thông tin User hiện tại mà không lộ mật khẩu.

---

<a name="cac-loi-thuong-gap-va-cach-xu-ly"></a>
## Các lỗi thường gặp và cách xử lý

1. **Lỗi `Không thể ping tới MongoDB`:**
   * Hãy kiểm tra xem container Docker MongoDB đã bật chưa bằng lệnh `docker ps`. Nếu chưa, hãy gõ `docker compose up -d` ở thư mục gốc.
2. **Lỗi `missing go.sum entry`:**
   * Trong thư mục `backend`, chạy:
     ```powershell
     go mod tidy
     ```
     Lệnh này sẽ tự động phân tích mã nguồn và tải đúng các gói phụ thuộc còn thiếu.
3. **Mở MongoDB Compass kiểm tra:**
   * Bạn mở MongoDB Compass, vào database `chatapp` $\to$ collection `users` sẽ thấy tài khoản vừa tạo xuất hiện với mật khẩu đã được băm an toàn dạng `$2a$10$...`.

---

### 🎉 Hoàn thành Giai đoạn 2!
Bạn đã có một hệ thống Backend Go chuẩn Clean Architecture với cơ chế bảo mật xác thực hoàn chỉnh, sẵn sàng bước sang **Giai đoạn 3: Xây dựng WebSocket Hub & Luồng Nhắn Tin Cốt Lõi**!
