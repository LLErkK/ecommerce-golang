package controllers

import (
	"ecommerce-golang/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func GetSellerProfile(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)

	var sellerProfile models.SellerProfile
	if err := db.Where("id = ?", sellerID).First(&sellerProfile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile berhasil diambil",
		"data":    sellerProfile,
	})
}

func UpdateSellerProfile(c *gin.Context) {
	sellerID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)

	var input models.SellerProfile
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var profile models.SellerProfile
	if err := db.Where("seller_id = ?", sellerID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if input.ShopName != "" {
		profile.ShopName = input.ShopName
	}
	if input.Phone != "" {
		profile.Phone = input.Phone
	}
	if input.Address != "" {
		profile.Address = input.Address
	}
	if input.City != "" {
		profile.City = input.City
	}
	if input.ShopLogo != "" {
		profile.ShopLogo = input.ShopLogo
	}

	if err := db.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile berhasil diambil",
		"data":    profile,
	})
}
