package main

import (
	"fmt"
	"log"
	"os"

	"ecommerce-golang/middleware"
	"ecommerce-golang/routes"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ambil variabel env dari Railway
	dbUser := os.Getenv("MYSQLUSER")
	dbPass := os.Getenv("MYSQLPASSWORD")
	dbHost := os.Getenv("MYSQLHOST")
	dbPort := os.Getenv("MYSQLPORT")
	dbName := os.Getenv("MYSQLDATABASE")

	// Format DSN MySQL untuk GORM
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	// Koneksi pakai GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}

	// Setup Gin
	r := gin.Default()

	// Middleware inject DB
	r.Use(middleware.InjectDB(db))

	// Routes
	routes.AuthRoutes(r, db)

	// Listen pakai port Railway (default 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
