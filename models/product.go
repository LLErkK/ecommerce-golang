package models

import "time"

type Product struct {
	ID           uint      `json:"id" gorm:"primary_key"`
	SellerID     uint      `json:"seller_id" gorm:"not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	Description  string    `json:"description" gorm:"type:text"`
	Price        float64   `json:"price" gorm:"not null"`
	Stock        uint      `json:"stock" gorm:"default:0"`
	Category     string    `json:"category" gorm:"size:255"`
	Images       string    `json:"images" gorm:"type:text"`     //json array of image url
	Weight       float64   `json:"weight" gorm:"not null"`      //dalam gram
	Dimensions   string    `json:"dimensions" gorm:"type:text"` //panjang lebar tinggi
	Brand        string    `json:"brand" gorm:"size:255"`
	IsActive     bool      `json:"is_active" gorm:"default:false"`
	Rating       float64   `json:"rating" gorm:"default:0"`
	TotalReviews uint      `json:"total_reviews" gorm:"default:0"`
	TotalSold    uint      `json:"total_sold" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Seller SellerProfile `gorm:"foreignKey:SellerID;references:SellerID" json:"seller,omitempty"`
}

// ProductListView - For search/browse (lightweight)
type ProductListView struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Image     string  `json:"image"` // First image only
	Rating    float64 `json:"rating"`
	TotalSold int     `json:"total_sold"`
	ShopName  string  `json:"shop_name"`
	City      string  `json:"city"`
}

// ProductDetailView - For product detail page
type ProductDetailView struct {
	Product
	ShopName   string  `json:"shop_name"`
	ShopLogo   string  `json:"shop_logo"`
	ShopCity   string  `json:"shop_city"`
	ShopRating float64 `json:"shop_rating"`
}
