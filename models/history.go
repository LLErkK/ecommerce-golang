package models

import "time"

type ProductUserHistory struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Category  string    `json:"category" gorm:"size:255"` // Copy kategori saat pembelian
	Quantity  uint      `json:"quantity" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"` // Harga saat beli
	CreatedAt time.Time `json:"created_at"`

	Product Product `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"product,omitempty"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
}

type ProductUserCart struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Quantity  uint      `json:"quantity" gorm:"not null;default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Product Product `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"product,omitempty"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
}

func (ProductUserHistory) TableName() string {
	return "product_user_histories"
}

func (ProductUserCart) TableName() string {
	return "product_user_carts"
}
