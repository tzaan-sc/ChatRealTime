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
