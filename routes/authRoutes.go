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

	//public karena siapa saja bisa search
	product := r.Group("/product")
	{
		product.GET("", controllers.SearchProduct)
		product.GET("/:id", controllers.GetProductDetail)
		product.GET("/recommendations", controllers.GetRecommendations)
	}
	r.GET("/categories", controllers.GetCategories)
	//authorisasi User
	userProtected := r.Group("/user")
	userProtected.Use(middleware.AuthMiddleware("user"))
	{
		userProtected.GET("/me", controllers.UserMe)
		userProtected.GET("/profile", controllers.GetUserProfile)
		userProtected.PUT("/profile", controllers.UpdateUserProfile)
	}

	//authorisasi seller
	sellerProtected := r.Group("/seller")
	sellerProtected.Use(middleware.AuthMiddleware("seller"))
	{
		sellerProtected.GET("/me", controllers.SellerMe)
		sellerProtected.GET("/profile", controllers.GetSellerProfile)
		sellerProtected.PUT("/profile", controllers.UpdateSellerProfile)

		//seller product management
		sellerProtected.POST("/products", controllers.CreateProduct)
		sellerProtected.GET("/products", controllers.GetSellerProducts)
		sellerProtected.GET("/products/:id", controllers.GetSellerProduct)
		sellerProtected.PUT("products/:id", controllers.UpdateProduct)
		sellerProtected.DELETE("products/:id", controllers.DeleteProduct)
	}
}
