// config/db.go - Updated untuk Railway MySQL variables
package config

import (
	"ecommerce-golang/models"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/url"
	"os"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	var dsn string

	// Prioritas: Railway MySQL variables → MYSQL_URL → Manual config
	if host := os.Getenv("MYSQLHOST"); host != "" {
		// Gunakan Railway individual MySQL variables
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			getEnv("MYSQLUSER", "root"),
			getEnv("MYSQLPASSWORD", ""),
			getEnv("MYSQLHOST", "127.0.0.1"),
			getEnv("MYSQLPORT", "3306"),
			getEnv("MYSQLDATABASE", "ecommerce"),
		)
		log.Println("Using Railway MySQL individual variables")
	} else if mysqlURL := os.Getenv("MYSQL_URL"); mysqlURL != "" {
		// Parse MySQL URL dari Railway (backup method)
		dsn = convertMySQLURL(mysqlURL)
		log.Println("Using Railway MySQL URL")
	} else {
		// Fallback ke manual config untuk development local
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			getEnv("DB_USER", "root"),
			getEnv("DB_PASS", ""),
			getEnv("DB_HOST", "127.0.0.1"),
			getEnv("DB_PORT", "3306"),
			getEnv("DB_NAME", "ecommerce"),
		)
		log.Println("Using manual database configuration")
	}

	log.Printf("Connecting to database with DSN: %s", sanitizeDSN(dsn))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	DB = db

	// Optional: Set connection pool untuk production
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Warning: Failed to get underlying sql.DB: %v", err)
	} else {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(3600) // 1 hour
	}

	// Auto migrate semua tabel
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.Seller{},
		&models.UserProfile{},
		&models.SellerProfile{},
		&models.Product{},
		&models.ProductUserHistory{},
		&models.ProductUserCart{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("Database migrations completed successfully")

	return db
}

// Convert Railway MySQL URL format ke GORM format (backup method)
func convertMySQLURL(mysqlURL string) string {
	// Parse URL: mysql://user:password@host:port/database
	u, err := url.Parse(mysqlURL)
	if err != nil {
		log.Fatal("Error parsing MYSQL_URL:", err)
	}

	password, _ := u.User.Password()

	return fmt.Sprintf("%s:%s@tcp(%s)%s?charset=utf8mb4&parseTime=True&loc=Local",
		u.User.Username(),
		password,
		u.Host,
		u.Path,
	)
}

// Helper function untuk get environment variables
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// Helper function untuk sanitize DSN untuk logging (hide password)
func sanitizeDSN(dsn string) string {
	// Find password part dan replace dengan ***
	// Format: user:password@tcp(host:port)/database
	if len(dsn) == 0 {
		return dsn
	}

	// Simple regex replacement untuk hide password
	// Cari pattern :password@ dan replace password dengan ***
	start := 0
	for i, char := range dsn {
		if char == ':' && start == 0 {
			start = i + 1
		} else if char == '@' && start > 0 {
			return dsn[:start] + "***" + dsn[i:]
		}
	}
	return dsn
}
