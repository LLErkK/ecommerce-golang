package middleware

import (
	"ecommerce-golang/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware(expectedType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token"})
			c.Abort()
			return
		}

		claims, err := utils.ParseJWT(tokenStr)
		if err != nil || claims["type"] != expectedType {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(claims["id"].(float64)))
		c.Next()
	}
}
