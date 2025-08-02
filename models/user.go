package models

import "time"

type User struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}
