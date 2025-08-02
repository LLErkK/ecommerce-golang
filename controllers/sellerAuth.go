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
	var input models.Seller
	db := c.MustGet("db").(*gorm.DB)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	input.Password = string(hashed)

	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": input, "message": "Register success"})
}

func SellerLogin(c *gin.Context) {
	var input struct {
		Email    string `form:"email" binding:"required"`
		Password string `form:"password" binding:"required"`
	}
	var seller models.Seller
	db := c.MustGet("db").(*gorm.DB)

	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//email salah
	if err := db.Where("email = ?", input.Email).First(&seller).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	//password salah
	if err := bcrypt.CompareHashAndPassword([]byte(seller.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, _ := utils.GenerateJWT(seller.ID, "seller")
	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login success"})
}

func SellerMe(c *gin.Context) {
	id := c.MustGet("id").(uint)
	c.JSON(http.StatusOK, gin.H{"id": id, "role": "Seller"})
}
