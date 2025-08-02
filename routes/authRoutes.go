package routes

import (
	"ecommerce-golang/controllers"
	"ecommerce-golang/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(r *gin.Engine, db *gorm.DB) {
	user := r.Group("/user")
	{
		user.POST("/register", func(c *gin.Context) {
			controllers.UserRegister(c)
		})
		user.POST("/login", func(c *gin.Context) {
			controllers.UserLogin(c)
		})
	}

	seller := r.Group("/seller")
	{
		seller.POST("/register", func(c *gin.Context) {
			controllers.SellerRegister(c)
		})
		seller.POST("/login", func(c *gin.Context) {
			controllers.SellerLogin(c)
		})
	}

	//authorisasi User
	userProtected := r.Group("/user")
	userProtected.Use(middleware.AuthMiddleware("user"))
	{
		//logika bisnis user
	}

	//authorisasi seller
	sellerProtected := r.Group("/seller")
	sellerProtected.Use(middleware.AuthMiddleware("seller"))
	{
		//logika bisnis seller
	}
}
