// models/sellerProfile.go
package models

import "time"

type SellerProfile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SellerID  uint      `gorm:"uniqueIndex;not null" json:"seller_id"`
	ShopName  string    `json:"shop_name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	ShopLogo  string    `json:"shop_logo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
