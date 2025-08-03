package controllers

import (
	"ecommerce-golang/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func SearchProduct(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	search := c.Query("search")
	category := c.Query("category")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	sortBy := c.Query("sort")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	//convert page dan limit
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}
	offset := (pageInt - 1) * limitInt

	query := db.Table("products").
		Select(`products.id, products.name, products.price, products.rating, 
                products.total_sold, products.images,
                seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true)

	if search != "" {
		query = query.Where("products.name LIKE = ?", "%"+search+"%")
	}
	if category != "" {
		query = query.Where("products.category = ?", category)
	}
	if minPrice != "" {
		query = query.Where("products.price >= ?", minPrice)
	}
	if maxPrice != "" {
		query = query.Where("products.price <= ?", maxPrice)
	}

	switch sortBy {
	case "price_asc":
		query = query.Order("products.price ASC")
	case "price_desc":
		query = query.Order("products.price DESC")
	case "rating":
		query = query.Order("products.rating DESC")
	case "newest":
		query = query.Order("products.created_at DESC")
	case "bestseller":
		query = query.Order("products.total_sold DESC")
	default:
		query = query.Order("products.total_sold DESC")
	}

	var total int64
	countQuery := db.Table("products").Where("is_active = ? AND products.stock > 0", true)
	if search != "" {
		countQuery = countQuery.Where("name LIKE ?", "%"+search+"%")
	}
	if category != "" {
		countQuery = countQuery.Where("category = ?", category)
	}
	if minPrice != "" {
		countQuery = countQuery.Where("price >= ?", minPrice)
	}
	if maxPrice != "" {
		countQuery = countQuery.Where("price <= ?", maxPrice)
	}
	countQuery.Count(&total)

	var products []models.ProductListView
	if err := query.Limit(limitInt).Offset(offset).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil products"})
		return
	}

	//mengambil foto pertama untuk thumbnail
	for i := range products {
		var imageList []string
		err := json.Unmarshal([]byte(products[i].Image), &imageList)
		if err == nil && len(imageList) > 0 {
			products[i].Image = imageList[0] // ambil thumbnail dari elemen pertama
		} else {
			products[i].Image = "" // fallback kosong jika parsing gagal
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products berhasil diambil",
		"data":    products,
		"pagination": gin.H{
			"page":        pageInt,
			"limit":       limitInt,
			"total":       total,
			"total_pages": (total + int64(limitInt) - 1) / int64(limitInt),
		},
	})
}

func GetProductDetail(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	productID := c.Param("id")

	var productDetail models.ProductDetailView
	err := db.Table("products").
		Select(`products.*, 
                seller_profiles.shop_name, seller_profiles.shop_logo, 
                seller_profiles.city as shop_city, 0 as shop_rating`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.id = ? AND products.is_active = ?", productID, true).
		First(&productDetail).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product tidaK ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Product detail berhasil diambil",
		"data":    productDetail,
	})
}

// aku ingin jika anonymus maka akan diberikan best rating dan best seller
// jika sudah login maka akan berdasarkan kategori riwayat pembelian +best rating dan best seller
// jika user belum pernah beli apapun maka berdasarkan kategori keranjang +best rating dan best seller
// jika user baru saja membuat maka akan seperti anonymus
func GetRecommendations(c *gin.Context) {

}
func GetCategories(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var categories []struct {
		Category string `json:"category"`
		Count    int    `json:"count"`
	}

	err := db.Table("products").
		Select("category, COUNT(*) as count").
		Where("is_active = ? AND stock > 0", true).
		Group("category").
		Order("count DESC").
		Find(&categories).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Categories berhasil diambil",
		"data":    categories,
	})
}
