package middleware

import (
	"ecommerce-golang/utils"
	"github.com/gin-gonic/gin"
	"strings"
)

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Ambil user_id dan role dari map
		userIDFloat, ok1 := claims["user_id"].(float64) // biasanya float64
		role, ok2 := claims["role"].(string)

		if ok1 {
			c.Set("id", int(userIDFloat)) // konversi float64 ke int
		}
		if ok2 {
			c.Set("role", role)
		}

		c.Next()
	}
}
