package controllers

import (
	"ecommerce-golang/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func GetUserProfile(c *gin.Context) {
	userID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)

	var profile models.UserProfile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"massage": "Profile berhasil didapatkan",
		"data":    profile,
	})
}

func UpdateUserProfile(c *gin.Context) {
	userID := c.MustGet("id").(uint)
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		FullName     string `json:"full_name"`
		Phone        string `json:"phone"`
		Address      string `json:"address"`
		City         string `json:"city"`
		PhotoProfile string `json:"photo_profile"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var profile models.UserProfile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	// Update fields yang tidak kosong
	if input.FullName != "" {
		profile.FullName = input.FullName
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
	if input.PhotoProfile != "" {
		profile.PhotoProfile = input.PhotoProfile
	}
	if err := db.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile berhasil diupdate",
		"data":    profile,
	})

}
