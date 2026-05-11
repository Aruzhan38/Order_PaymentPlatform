package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		ctx := c.Request.Context()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "rate limiter error",
			})
			c.Abort()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			ttl, _ := redisClient.TTL(ctx, key).Result()

			c.Header("Retry-After", strconv.Itoa(int(ttl.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"limit":       limit,
				"retry_after": ttl.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
