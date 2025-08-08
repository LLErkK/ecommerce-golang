package controllers

import (
	"ecommerce-golang/models"
	"ecommerce-golang/utils"
	"encoding/json"
	"errors"
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
            products.total_sold, 
            JSON_UNQUOTE(JSON_EXTRACT(products.images, '$[0]')) as image,
            seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true)

	if search != "" {
		query = query.Where("products.name LIKE  ?", "%"+search+"%")
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
	db := c.MustGet("db").(*gorm.DB)

	// Check authentication
	userID, exists := c.Get("id")
	var isAuthenticated bool
	var userIDUint uint

	if exists {
		isAuthenticated = true
		switch v := userID.(type) {
		case uint:
			userIDUint = v
		case int:
			userIDUint = uint(v)
		case float64:
			userIDUint = uint(v)
		default:
			isAuthenticated = false // tipe tidak dikenal, anggap tidak login
		}
	}

	var recommendedProducts []models.ProductListView
	var basis string
	var categories []string

	if !isAuthenticated {
		// Case 1: Anonymous user
		recommendedProducts = utils.GetAnonymousRecommendations(db)
		basis = "anonymous_bestseller_and_rating"
	} else {
		// Case 2: Authenticated user - check purchase history
		if utils.CheckUserHasPurchaseHistory(db, userIDUint) {
			// User has purchase history
			categories = utils.GetTopCategoriesByPurchase(db, userIDUint, 3)
			recommendedProducts = utils.GetCategoryBasedRecommendations(db, categories, userIDUint)
			basis = "purchase_history_based"
		} else if utils.CheckUserHasCartItems(db, userIDUint) {
			// User has cart items but no purchase history
			categories = utils.GetTopCategoriesByCart(db, userIDUint)
			recommendedProducts = utils.GetCategoryBasedRecommendations(db, categories, userIDUint)
			basis = "cart_based"
		} else {
			// Case 3: New user (no purchase, no cart) - treat like anonymous
			recommendedProducts = utils.GetAnonymousRecommendations(db)
			basis = "new_user_bestseller_and_rating"
		}
	}

	// Fallback jika tidak ada hasil
	if len(recommendedProducts) == 0 {
		recommendedProducts = utils.GetAnonymousRecommendations(db)
		basis = "fallback_bestseller_and_rating"
	}

	// Limit hasil akhir
	if len(recommendedProducts) > 20 {
		recommendedProducts = recommendedProducts[:20]
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recommendations berhasil diambil",
		"data":    recommendedProducts,
		"metadata": gin.H{
			"total":                len(recommendedProducts),
			"is_authenticated":     isAuthenticated,
			"recommendation_basis": basis,
			"categories_used":      categories,
		},
	})
}

func getUserPurchaseCategories(db *gorm.DB, userID uint) []string {
	var categories []string

	err := db.Table("product_user_histories").
		Select("DISTINCT category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5). // Ambil maksimal 5 kategori terakhir
		Find(&categories).Error

	if err != nil {
		return []string{}
	}

	return categories

}

func getUserCartCategories(db *gorm.DB, userID uint) []string {
	var categories []string
	err := db.Table("product_user_carts").
		Select("DISTINCT products.category").
		Joins("JOIN products ON product_user_carts.product_id = products.id").
		Where("product_user_carts.user_id = ? AND products.is_active = ?", userID, true).
		Find(&categories).Error

	if err != nil {
		return []string{}
	}

	return categories
}

func getAnonymousRecommendations(db *gorm.DB) []models.ProductListView {
	var products []models.ProductListView

	err := db.Table("products").
		Select(`products.id, products.name, products.price, products.rating, 
                products.total_sold, products.images,
                seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true).
		Where("(products.rating >= 4.0 OR products.total_sold >= 10)").              // Filter best rating atau best seller
		Order("(products.rating * 0.7 + (products.total_sold / 100.0) * 0.3) DESC"). // Weighted scoring
		Limit(20).
		Find(&products).Error

	if err != nil {
		return []models.ProductListView{}
	}

	// Process images untuk thumbnail
	processProductImages(&products)
	return products
}

func getCategoryBasedRecommendations(db *gorm.DB, categories []string, userID uint) []models.ProductListView {
	var products []models.ProductListView

	// Query untuk produk dari kategori yang diminati
	categoryQuery := db.Table("products").
		Select(`products.id, products.name, products.price, products.rating, 
                products.total_sold, products.images,
                seller_profiles.shop_name, seller_profiles.city`).
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("products.is_active = ? AND products.stock > 0", true).
		Where("products.category IN ?", categories).
		Order("(products.rating * 0.6 + (products.total_sold / 100.0) * 0.4) DESC"). // Weighted untuk kategori
		Limit(15)

	err := categoryQuery.Find(&products).Error
	if err != nil {
		return getAnonymousRecommendations(db)
	}

	// Tambahkan beberapa best seller/rating dari kategori lain untuk variasi
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

	// Remove duplicates jika ada
	products = removeDuplicateProducts(products)

	// Process images untuk thumbnail
	processProductImages(&products)

	return products
}

func processProductImages(products *[]models.ProductListView) {
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

func removeDuplicateProducts(products []models.ProductListView) []models.ProductListView {
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

func getBasisDescription(isAuthenticated bool, hasCategories bool, hasUserID bool) string {
	if !isAuthenticated {
		return "best_rating_and_bestseller"
	}

	if hasCategories {
		return "purchase_history_and_bestseller"
	}

	return "cart_categories_and_bestseller"
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

func AddProductToCart(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("id").(uint)
	productID := c.Param("id")

	var input struct {
		Quantity uint `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		input.Quantity = 1
	}

	productIDUint, err := strconv.ParseUint(productID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := db.First(&product, "id = ? AND is_active = ?", uint(productIDUint), true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found atau tidak aktif"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Cek stock availability
	if product.Stock < input.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":              "Insufficient stock",
			"available_stock":    product.Stock,
			"requested_quantity": input.Quantity,
		})
		return
	}

	// Cek apakah produk sudah ada di cart user
	var cartItem models.ProductUserCart
	err = db.Where("user_id = ? AND product_id = ?", userID, uint(productIDUint)).First(&cartItem).Error

	if err == nil {

		newQuantity := cartItem.Quantity + input.Quantity

		if product.Stock < newQuantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "Total quantity would exceed stock",
				"current_in_cart": cartItem.Quantity,
				"requested_add":   input.Quantity,
				"available_stock": product.Stock,
			})
			return
		}

		cartItem.Quantity = newQuantity
		if err := db.Save(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Product quantity updated in cart",
			"data": gin.H{
				"cart_id":      cartItem.ID,
				"product_id":   cartItem.ProductID,
				"new_quantity": cartItem.Quantity,
				"total_price":  float64(cartItem.Quantity) * product.Price,
			},
		})
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Kalau belum ada → buat baru
		newItem := models.ProductUserCart{
			UserID:    userID,
			ProductID: uint(productIDUint),
			Quantity:  input.Quantity,
		}

		if err := db.Create(&newItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Product added to cart successfully",
			"data": gin.H{
				"cart_id":     newItem.ID,
				"product_id":  newItem.ProductID,
				"quantity":    newItem.Quantity,
				"total_price": float64(newItem.Quantity) * product.Price,
			},
		})
	} else {
		// Error lain
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
}
func GetUserCart(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("id").(uint)

	var cartItems []struct {
		models.ProductUserCart
		ProductName  string  `json:"product_name"`
		ProductPrice float64 `json:"product_price"`
		ProductImage string  `json:"product_image"`
		ShopName     string  `json:"shop_name"`
		IsActive     bool    `json:"is_active"`
		Stock        uint    `json:"stock"`
		TotalPrice   float64 `json:"total_price"`
	}

	err := db.Table("product_user_carts").
		Select(`product_user_carts.*, 
				products.name as product_name, 
				products.price as product_price,
				products.images as product_image,
				products.is_active,
				products.stock,
				seller_profiles.shop_name,
				(product_user_carts.quantity * products.price) as total_price`).
		Joins("LEFT JOIN products ON product_user_carts.product_id = products.id").
		Joins("LEFT JOIN seller_profiles ON products.seller_id = seller_profiles.seller_id").
		Where("product_user_carts.user_id = ?", userID).
		Order("product_user_carts.created_at DESC").
		Find(&cartItems).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart items"})
		return
	}

	// Process product images and calculate totals
	var totalItems uint = 0
	var totalPrice float64 = 0

	for i := range cartItems {
		// Process first image
		if cartItems[i].ProductImage != "" {
			var imageList []string
			if json.Unmarshal([]byte(cartItems[i].ProductImage), &imageList) == nil && len(imageList) > 0 {
				cartItems[i].ProductImage = imageList[0]
			}
		}

		// Only count active products
		if cartItems[i].IsActive {
			totalItems += cartItems[i].Quantity
			totalPrice += cartItems[i].TotalPrice
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart berhasil diambil",
		"data":    cartItems,
		"summary": gin.H{
			"total_items": totalItems,
			"total_price": totalPrice,
			"item_count":  len(cartItems),
		},
	})
}

func UpdateCartItem(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("id").(uint)
	cartItemID := c.Param("cart_id")

	var input struct {
		Quantity uint `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find cart item
	var cartItem models.ProductUserCart
	err := db.Where("id = ? AND user_id = ?", cartItemID, userID).First(&cartItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Check product stock
	var product models.Product
	if err := db.First(&product, cartItem.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if product.Stock < input.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":              "Insufficient stock",
			"available_stock":    product.Stock,
			"requested_quantity": input.Quantity,
		})
		return
	}

	// Update quantity
	cartItem.Quantity = input.Quantity
	if err := db.Save(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart item updated successfully",
		"data": gin.H{
			"cart_id":      cartItem.ID,
			"product_id":   cartItem.ProductID,
			"new_quantity": cartItem.Quantity,
			"total_price":  float64(cartItem.Quantity) * product.Price,
		},
	})
}

func RemoveFromCart(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("id").(uint)
	cartItemID := c.Param("cart_id")

	// Find and delete cart item
	var cartItem models.ProductUserCart
	err := db.Where("id = ? AND user_id = ?", cartItemID, userID).First(&cartItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if err := db.Delete(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed from cart successfully",
	})
}

func ClearCart(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("id").(uint)

	if err := db.Where("user_id = ?", userID).Delete(&models.ProductUserCart{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}
