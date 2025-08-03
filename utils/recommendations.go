// utils/recommendations.go
package utils

import (
	"ecommerce-golang/models"
	"encoding/json"
	"gorm.io/gorm"
)

// CheckUserHasPurchaseHistory - mengecek apakah user pernah membeli
func CheckUserHasPurchaseHistory(db *gorm.DB, userID uint) bool {
	var count int64
	db.Table("product_user_histories").Where("user_id = ?", userID).Count(&count)
	return count > 0
}

// CheckUserHasCartItems - mengecek apakah user punya item di keranjang
func CheckUserHasCartItems(db *gorm.DB, userID uint) bool {
	var count int64
	db.Table("product_user_carts").Where("user_id = ?", userID).Count(&count)
	return count > 0
}

// GetTopCategoriesByPurchase - mendapatkan kategori favorit berdasarkan pembelian
func GetTopCategoriesByPurchase(db *gorm.DB, userID uint, limit int) []string {
	var results []struct {
		Category string
		Count    int64
	}

	db.Table("product_user_histories").
		Select("category, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("category").
		Order("count DESC, created_at DESC").
		Limit(limit).
		Find(&results)

	var categories []string
	for _, result := range results {
		categories = append(categories, result.Category)
	}

	return categories
}

// GetTopCategoriesByCart - mendapatkan kategori dari keranjang
func GetTopCategoriesByCart(db *gorm.DB, userID uint) []string {
	var categories []string

	db.Table("product_user_carts").
		Select("DISTINCT products.category").
		Joins("JOIN products ON product_user_carts.product_id = products.id").
		Where("product_user_carts.user_id = ? AND products.is_active = ?", userID, true).
		Order("product_user_carts.created_at DESC").
		Find(&categories)

	return categories
}

// GetAnonymousRecommendations - rekomendasi untuk anonymous user
func GetAnonymousRecommendations(db *gorm.DB) []models.ProductListView {
	var products []models.ProductListView

	err := db.Table("products").
		Select(`products.id, products.name, products.price, products.rating, 
                products.total_sold, products.images,
                seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true).
		Where("(products.rating >= 4.0 OR products.total_sold >= 10)").
		Order("(products.rating * 0.7 + (products.total_sold / 100.0) * 0.3) DESC").
		Limit(20).
		Find(&products).Error

	if err != nil {
		return []models.ProductListView{}
	}

	ProcessProductImages(&products)
	return products
}

// GetCategoryBasedRecommendations - rekomendasi berdasarkan kategori
func GetCategoryBasedRecommendations(db *gorm.DB, categories []string, userID uint) []models.ProductListView {
	var products []models.ProductListView

	// Query untuk produk dari kategori yang diminati
	categoryQuery := db.Table("products").
		Select(`products.id, products.name, products.price, products.rating, 
                products.total_sold, products.images,
                seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true).
		Where("products.category IN ?", categories).
		Order("(products.rating * 0.6 + (products.total_sold / 100.0) * 0.4) DESC").
		Limit(15)

	err := categoryQuery.Find(&products).Error
	if err != nil {
		return GetAnonymousRecommendations(db)
	}

	// Tambahkan beberapa best seller/rating dari kategori lain
	var additionalProducts []models.ProductListView
	additionalQuery := db.Table("products").
		Select(`products.id, products.name, products.price, products.rating, 
                products.total_sold, products.images,
                seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true).
		Where("products.category NOT IN ?", categories).
		Where("(products.rating >= 4.5 OR products.total_sold >= 50)").
		Order("(products.rating * 0.8 + (products.total_sold / 100.0) * 0.2) DESC").
		Limit(5)

	additionalQuery.Find(&additionalProducts)

	// Gabungkan hasil
	products = append(products, additionalProducts...)

	// Remove duplicates
	products = RemoveDuplicateProducts(products)

	ProcessProductImages(&products)
	return products
}

// ProcessProductImages - memproses gambar produk untuk thumbnail
func ProcessProductImages(products *[]models.ProductListView) {
	for i := range *products {
		var imageList []string
		err := json.Unmarshal([]byte((*products)[i].Image), &imageList)
		if err == nil && len(imageList) > 0 {
			(*products)[i].Image = imageList[0]
		} else {
			(*products)[i].Image = ""
		}
	}
}

// RemoveDuplicateProducts - menghapus duplikat produk
func RemoveDuplicateProducts(products []models.ProductListView) []models.ProductListView {
	seen := make(map[uint]bool)
	var result []models.ProductListView

	for _, product := range products {
		if !seen[product.ID] {
			seen[product.ID] = true
			result = append(result, product)
		}
	}

	return result
}

// GetBasisDescription - mendapatkan deskripsi basis rekomendasi
func GetBasisDescription(isAuthenticated bool, hasCategories bool) string {
	if !isAuthenticated {
		return "best_rating_and_bestseller"
	}

	if hasCategories {
		return "purchase_history_and_bestseller"
	}

	return "cart_categories_and_bestseller"
}
