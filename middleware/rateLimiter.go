package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiterConfig struct {
	RequestPerMinute int
	BurstSize        int
	WindowSize       time.Duration
	KeyFunc          func(*gin.Context) string
}

type Client struct {
	Tokens     int
	LastRefill time.Time
	Request    []time.Time
	mu         sync.Mutex
}

type RateLimiter struct {
	clients map[string]*Client
	config  RateLimiterConfig
	mu      sync.RWMutex
}

func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.RequestPerMinute == 0 {
		config.RequestPerMinute = 60
	}
	if config.BurstSize == 0 {
		config.BurstSize = 10
	}
	if config.WindowSize == 0 {
		config.WindowSize = time.Minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}
	return &RateLimiter{
		clients: make(map[string]*Client),
		config:  config,
	}
}

func (rl *RateLimiter) TokenBucketMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.config.KeyFunc(c)

		rl.mu.Lock()
		client, exists := rl.clients[key]
		if !exists {
			client = &Client{
				Tokens:     rl.config.BurstSize,
				LastRefill: time.Now(),
			}
			rl.clients[key] = client
		}
		rl.mu.Unlock()
		client.mu.Lock()
		defer client.mu.Unlock()

		now := time.Now()
		timePassed := now.Sub(client.LastRefill)
		tokensToAdd := int(timePassed.Minutes() * float64(rl.config.RequestPerMinute))

		if tokensToAdd > 0 {
			client.Tokens += tokensToAdd
			if client.Tokens > rl.config.BurstSize {
				client.Tokens = rl.config.BurstSize
			}
			client.LastRefill = now
		}

		if client.Tokens <= 0 {
			retryAfter := time.Minute / time.Duration(rl.config.RequestPerMinute)

			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(retryAfter).Unix(), 10))
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d requests per minute", rl.config.RequestPerMinute),
				"retry_after": int(retryAfter.Seconds()),
			})
			c.Abort()
			return
		}

		client.Tokens--
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(client.Tokens))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(time.Minute).Unix(), 10))

		c.Next()
	}
}

func (rl *RateLimiter) SlidingWindowMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.config.KeyFunc(c)
		now := time.Now()

		rl.mu.Lock()
		client, exists := rl.clients[key]
		if !exists {
			client = &Client{
				Request: make([]time.Time, 0),
			}
			rl.clients[key] = client
		}
		rl.mu.Unlock()
		client.mu.Lock()
		defer client.mu.Unlock()

		windoStart := now.Add(-rl.config.WindowSize)
		var validRequests []time.Time
		for _, reqTime := range client.Request {
			if reqTime.After(windoStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		client.Request = validRequests

		if len(client.Request) >= rl.config.RequestPerMinute {
			retryAfter := client.Request[0].Add(rl.config.WindowSize).Sub(now)
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(retryAfter).Unix(), 10))
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d requests per %v", rl.config.RequestPerMinute, rl.config.WindowSize),
				"retry_after": int(retryAfter.Seconds()),
			})
			c.Abort()
			return
		}
		client.Request = append(client.Request, now)

		remaining := rl.config.RequestPerMinute - len(client.Request)
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(rl.config.WindowSize).Unix(), 10))

		c.Next()
	}
}
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, client := range rl.clients {
		client.mu.Lock()
		// Hapus client yang tidak aktif lebih dari 1 jam
		if now.Sub(client.LastRefill) > time.Hour {
			delete(rl.clients, key)
		}
		client.mu.Unlock()
	}
}

// StartCleanupRoutine - memulai routine untuk cleanup otomatis
func (rl *RateLimiter) StartCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			rl.Cleanup()
		}
	}()
}
