// middleware/auth.go
package middleware

import (
	"ecommerce-golang/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(expectedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Cek format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenStr := tokenParts[1]

		// Parse dan validasi token
		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Cek apakah role sesuai
		userType, ok := claims["user_type"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: missing user_type",
			})
			c.Abort()
			return
		}

		if userType != expectedRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied. Required role: " + expectedRole + ", your role: " + userType,
			})
			c.Abort()
			return
		}

		// Ambil user_id dari claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: missing user_id",
			})
			c.Abort()
			return
		}

		// Set data user ke context untuk digunakan di controller
		c.Set("id", uint(userID))
		c.Set("user_type", userType)
		c.Next()
	}
}
