// main.go - Updated untuk Railway deployment
package main

import (
	"ecommerce-golang/config"
	"ecommerce-golang/middleware"
	"ecommerce-golang/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	// Load .env file hanya untuk development local
	// Railway akan menggunakan environment variables langsung
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Set Gin mode berdasarkan environment
	if os.Getenv("GIN_MODE") == "" {
		if os.Getenv("RAILWAY_ENVIRONMENT") == "production" {
			gin.SetMode(gin.ReleaseMode)
		}
	}

	// Connect ke database
	db := config.ConnectDB()

	// Initialize Gin router
	r := gin.Default()

	// Add middleware
	r.Use(middleware.InjectDB(db))

	// Add basic health check endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "E-commerce API is running",
			"version": "1.0.0",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "healthy",
			"database":    "connected",
			"environment": getEnv("RAILWAY_ENVIRONMENT", "development"),
		})
	})

	// Setup routes
	routes.AuthRoutes(r, db)

	// Railway menyediakan PORT via environment variable
	port := getEnv("PORT", "8080")

	log.Printf("Server starting on port %s", port)
	log.Printf("Environment: %s", getEnv("RAILWAY_ENVIRONMENT", "development"))

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Helper function untuk get environment variables
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
