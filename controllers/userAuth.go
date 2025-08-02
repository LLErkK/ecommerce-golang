package controllers

import (
	"ecommerce-golang/models"
	"ecommerce-golang/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
)

func UserRegister(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		FullName string `json:"full_name" binding:"required"`
		Phone    string `json:"phone"`
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

	// Create user
	user := models.User{
		Email:    input.Email,
		Username: input.Username,
		Password: string(hashed),
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto create profile
	profile := models.UserProfile{
		UserID:       user.ID,
		FullName:     input.FullName,
		Phone:        input.Phone,
		PhotoProfile: "default-user.png",
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
			"user":    user,
			"profile": profile,
		},
	})
}
func UserLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var user models.User
	db := c.MustGet("db").(*gorm.DB)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//email salah
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//password salah
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(user.ID, "user")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal generate token" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login success"})
}

func UserMe(c *gin.Context) {
	id := c.MustGet("id").(uint)
	c.JSON(http.StatusOK, gin.H{"id": id, "role": "user"})
}
