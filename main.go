package main

import (
	"ecommerce-golang/config"
	"ecommerce-golang/middleware"
	"ecommerce-golang/routes"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	// Railway otomatis kasih environment variables
	// Jadi tidak perlu pakai .env file
	db := config.ConnectDB()
	r := gin.Default()
	r.Use(middleware.InjectDB(db))

	// Daftarkan routes
	routes.AuthRoutes(r, db)

	// Railway kasih PORT otomatis lewat env "PORT"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default kalau running lokal
	}

	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}
