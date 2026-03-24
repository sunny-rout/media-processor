package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	redisClient *redis.Client
	useRedis    bool
	fallbackMu  sync.RWMutex
	fallback    = make(map[string]*visitor)
)

func initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Rate limiter: failed to parse redis URL, using in-memory: %v", err)
		useRedis = false
		return
	}

	redisClient = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Rate limiter: redis connection failed, using in-memory: %v", err)
		useRedis = false
		return
	}

	useRedis = true
	log.Println("Rate limiter: connected to Redis")
}

func getVisitorInMemory(ip string) bool {
	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	v, exists := fallback[ip]
	if !exists {
		fallback[ip] = &visitor{
			limiter:  rate.NewLimiter(10, 20),
			lastSeen: time.Now(),
		}
		return true
	}

	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func cleanupInMemory() {
	for {
		time.Sleep(time.Minute)
		fallbackMu.Lock()
		for ip, v := range fallback {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(fallback, ip)
			}
		}
		fallbackMu.Unlock()
	}
}

// RateLimiter returns a middleware that limits requests per IP
// Uses Redis if available, falls back to in-memory limiting
func RateLimiter() gin.HandlerFunc {
	initRedis()

	if !useRedis {
		go cleanupInMemory()
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if useRedis {
			key := fmt.Sprintf("ratelimit:%s", ip)
			ctx := context.Background()

			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				// Redis error - allow request
				c.Next()
				return
			}

			if count == 1 {
				redisClient.Expire(ctx, key, time.Second)
			}

			if count > 20 {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded. Please try again later.",
				})
				c.Abort()
				return
			}
		} else {
			// In-memory fallback
			if !getVisitorInMemory(ip) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
