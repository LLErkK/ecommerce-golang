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
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := config.ConnectDB()
	r := gin.Default()
	r.Use(middleware.InjectDB(db))

	routes.AuthRoutes(r, db) // Daftarkan routes ke router yang sama
	r.Run(":8080")           // Jalankan router yang sudah ada routes-nya

}
