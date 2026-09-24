package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit tạo middleware giới hạn tần suất request theo IP client dựa trên Redis
// maxRequests: Số request tối đa trong khoảng thời gian window
// windowSeconds: Khung thời gian trượt tính theo giây
// actionName: Tên hành động để phân loại (e.g. "auth_login", "auth_register")
func RateLimit(redisClient *redis.Client, maxRequests int64, windowSeconds int, actionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = "unknown"
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		key := fmt.Sprintf("ratelimit:%s:%s", actionName, clientIP)

		// Tăng counter
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Fallback: nếu Redis gặp sự cố, không block request người dùng
			c.Next()
			return
		}

		// Nếu là lần đầu tiên, đặt TTL cho key
		if count == 1 {
			redisClient.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
		}

		// Tính toán TTL còn lại
		ttl, _ := redisClient.TTL(ctx, key).Result()
		remaining := maxRequests - count
		if remaining < 0 {
			remaining = 0
		}

		// Thêm các header chuẩn HTTP Rate Limit
		c.Header("X-RateLimit-Limit", strconv.FormatInt(maxRequests, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(int64(ttl.Seconds()), 10))

		if count > maxRequests {
			retrySeconds := int(ttl.Seconds())
			if retrySeconds <= 0 {
				retrySeconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(retrySeconds))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":        fmt.Sprintf("Bạn đã gửi quá nhiều yêu cầu trong thời gian ngắn. Vui lòng thử lại sau %d giây.", retrySeconds),
				"retry_after":  retrySeconds,
				"rate_limited": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
