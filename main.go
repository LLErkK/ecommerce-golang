package main

import (
	"ecommerce-golang/config"
	"ecommerce-golang/middleware"
	"ecommerce-golang/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// Coba load file .env, kalau gagal berarti pakai environment variable bawaan
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	db := config.ConnectDB()
	r := gin.Default()
	r.Use(middleware.InjectDB(db))

	routes.AuthRoutes(r, db)

	// Jalankan server di port 8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
