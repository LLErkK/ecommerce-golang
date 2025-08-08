// middleware/rateLimiterConfigs.go
package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// Predefined rate limiter configurations

// IPBasedRateLimiter - rate limiter berdasarkan IP
func IPBasedRateLimiter(requestsPerMinute, burstSize int) *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: requestsPerMinute,
		BurstSize:        burstSize,
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// UserBasedRateLimiter - rate limiter berdasarkan user ID (untuk authenticated users)
func UserBasedRateLimiter(requestsPerMinute, burstSize int) *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: requestsPerMinute,
		BurstSize:        burstSize,
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("id"); exists {
				switch v := userID.(type) {
				case uint:
					return fmt.Sprintf("search:user:%d", v)
				case int:
					return fmt.Sprintf("search:user:%d", v)
				case float64:
					return fmt.Sprintf("search:user:%d", int(v))
				default:
					return "ip:" + c.ClientIP()
				}
			}
			// Kalau tidak login, fallback ke IP
			return "ip:" + c.ClientIP()
		},
	})
}

// EndpointBasedRateLimiter - rate limiter berdasarkan endpoint + IP
func EndpointBasedRateLimiter(requestsPerMinute, burstSize int) *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: requestsPerMinute,
		BurstSize:        burstSize,
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.Request.URL.Path + ":" + c.ClientIP()
		},
	})
}

// StrictRateLimiter - rate limiter ketat untuk endpoint sensitif
func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 10, // Hanya 10 requests per menit
		BurstSize:        2,  // Burst kecil
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// ModerateRateLimiter - rate limiter sedang untuk API umum
func ModerateRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 60, // 60 requests per menit
		BurstSize:        10, // Burst sedang
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// RelaxedRateLimiter - rate limiter longgar untuk read operations
func RelaxedRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 120, // 120 requests per menit
		BurstSize:        20,  // Burst besar
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// AuthRateLimiter - rate limiter khusus untuk authentication endpoints
func AuthRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 5, // Hanya 5 attempts per menit
		BurstSize:        3, // Burst sangat kecil
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Kombinasi IP + endpoint untuk security
			return c.Request.URL.Path + ":" + c.ClientIP()
		},
	})
}

// SearchRateLimiter - rate limiter untuk search endpoints
func SearchRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 30,
		BurstSize:        5,
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("id"); exists {
				switch v := userID.(type) {
				case uint:
					return fmt.Sprintf("search:user:%d", v)
				case int:
					return fmt.Sprintf("search:user:%d", v)
				case float64:
					return fmt.Sprintf("search:user:%d", int(v))
				default:
					return "search:ip:" + c.ClientIP()
				}
			}
			return "search:ip:" + c.ClientIP()
		},
	})
}

// UploadRateLimiter - rate limiter untuk file uploads
func UploadRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 10, // 10 uploads per menit
		BurstSize:        2,  // Burst sangat kecil
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("id"); exists {
				return "upload:user:" + string(rune(userID.(uint)))
			}
			return "upload:ip:" + c.ClientIP()
		},
	})
}

// Tiered Rate Limiting berdasarkan role
func TieredRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestPerMinute: 60, // Default
		BurstSize:        10,
		WindowSize:       time.Minute,
		KeyFunc: func(c *gin.Context) string {
			role, exists := c.Get("role")
			if !exists {
				return "anonymous:" + c.ClientIP()
			}

			userID, userExists := c.Get("id")
			if !userExists {
				return "anonymous:" + c.ClientIP()
			}

			return role.(string) + ":" + string(rune(userID.(uint)))
		},
	})
}

// Dynamic Rate Limiter - rate limit berdasarkan role
func DynamicRoleBasedMiddleware() gin.HandlerFunc {
	anonymousLimiter := IPBasedRateLimiter(30, 5)  // Anonymous: 30/min
	userLimiter := UserBasedRateLimiter(60, 10)    // User: 60/min
	sellerLimiter := UserBasedRateLimiter(100, 15) // Seller: 100/min

	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			// Anonymous user
			anonymousLimiter.TokenBucketMiddleware()(c)
			return
		}

		switch role.(string) {
		case "seller":
			sellerLimiter.TokenBucketMiddleware()(c)
		case "user":
			userLimiter.TokenBucketMiddleware()(c)
		default:
			anonymousLimiter.TokenBucketMiddleware()(c)
		}
	}
}
