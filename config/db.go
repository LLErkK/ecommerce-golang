package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() *sql.DB {
	// Railway env
	user := os.Getenv("MYSQLUSER")
	password := os.Getenv("MYSQLPASSWORD")
	host := os.Getenv("MYSQLHOST")
	port := os.Getenv("MYSQLPORT")
	database := os.Getenv("MYSQLDATABASE")

	// Format DSN MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database not reachable: ", err)
	}

	log.Println("Connected to MySQL Railway successfully 🚀")
	return db
}
