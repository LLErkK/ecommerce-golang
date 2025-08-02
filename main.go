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

	router := gin.Default()
	routes.AuthRoutes(r, db)
	router.Run(":8080")
}
