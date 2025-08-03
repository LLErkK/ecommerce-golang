package controllers

import (
	"ecommerce-golang/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func CreateProduct(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		Stock       uint    `json:"stock" binding:"required,gt=0"`
		Category    string  `json:"category" binding:"required"`
		Image       string  `json:"image"`
		Weight      float64 `json:"weight"`
		Dimensions  string  `json:"dimensions"`
		Brand       string  `json:"brand"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		SellerID:    sellerID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    input.Category,
		Images:      input.Image,
		Weight:      input.Weight,
		Dimensions:  input.Dimensions,
		Brand:       input.Brand,
		IsActive:    true,
	}

	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Membuat Product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"massage": "Product created",
		"data":    product,
	})
}

func GetSellerProducts(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)

	var products []models.Product
	if err := db.Find(&products, "seller_id = ?", sellerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Mengambil Product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"massage": "Products found",
		"data":    products,
		"total":   len(products),
	})
}

func GetSellerProduct(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)
	productID := c.Param("id")

	var product models.Product
	if err := db.Where("id = ? AND seller_id = ?", productID, sellerID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"massage": "Product found",
		"data":    product,
	})
}

func UpdateProduct(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)
	productID := c.Param("id")

	var input struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		Stock       uint    `json:"stock" binding:"required,gt=0"`
		Category    string  `json:"category" `
		Images      string  `json:"images"`
		Weight      float64 `json:"weight"`
		Dimensions  string  `json:"dimensions"`
		Brand       string  `json:"brand"`
		IsActive    *bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var product models.Product
	if err := db.First(&product, "id = ? AND seller_id = ?", productID, sellerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if input.Name != "" {
		product.Name = input.Name
	}
	if input.Description != "" {
		product.Description = input.Description
	}
	if input.Price > 0 {
		product.Price = input.Price
	}
	if input.Stock >= 0 {
		product.Stock = input.Stock
	}
	if input.Category != "" {
		product.Category = input.Category
	}
	if input.Images != "" {
		product.Images = input.Images
	}
	if input.Weight > 0 {
		product.Weight = input.Weight
	}
	if input.Dimensions != "" {
		product.Dimensions = input.Dimensions
	}
	if input.Brand != "" {
		product.Brand = input.Brand
	}
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}

	if err := db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Update Product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"massage": "Product updated",
		"data":    product,
	})
}

func DeleteProduct(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)
	productID := c.Param("id")

	var product models.Product
	if err := db.Where("id = ? AND seller_id = ?", productID, sellerID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := db.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Mengahapus Product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"massage": "Product deleted",
	})
}
