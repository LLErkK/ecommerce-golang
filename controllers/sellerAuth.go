package controllers

import (
	"ecommerce-golang/models"
	"ecommerce-golang/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
)

func SellerRegister(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	db := c.MustGet("db").(*gorm.DB)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 12)

	// Start transaction
	tx := db.Begin()

	// Create seller
	seller := models.Seller{
		Email:    input.Email,
		Username: input.Username,
		Password: string(hashed),
	}

	if err := tx.Create(&seller).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto create profile
	profile := models.SellerProfile{
		SellerID: seller.ID,
		ShopName: "Toko " + seller.Username,
	}

	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membuat profile"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Register success",
		"data": gin.H{
			"seller":  seller,
			"profile": profile,
		},
	})
}

func SellerLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var seller models.Seller
	db := c.MustGet("db").(*gorm.DB)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//email salah
	if err := db.Where("email = ?", input.Email).First(&seller).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//password salah
	if err := bcrypt.CompareHashAndPassword([]byte(seller.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(seller.ID, "seller")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal generate token" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login success"})
}

func SellerMe(c *gin.Context) {
	id := c.MustGet("id").(uint)
	c.JSON(http.StatusOK, gin.H{"id": id, "role": "Seller"})
}
