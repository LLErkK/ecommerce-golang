// routes/authRoutes.go - Updated with Rate Limiting
package routes

import (
	"ecommerce-golang/controllers"
	"ecommerce-golang/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize rate limiters
	authLimiter := middleware.AuthRateLimiter()
	searchLimiter := middleware.SearchRateLimiter()
	moderateLimiter := middleware.ModerateRateLimiter()
	relaxedLimiter := middleware.RelaxedRateLimiter()

	// Start cleanup routines
	authLimiter.StartCleanupRoutine()
	searchLimiter.StartCleanupRoutine()
	moderateLimiter.StartCleanupRoutine()
	relaxedLimiter.StartCleanupRoutine()

	// User routes dengan rate limiting ketat untuk auth
	user := r.Group("/user")
	user.Use(authLimiter.TokenBucketMiddleware()) // Rate limit untuk auth endpoints
	{
		user.POST("/register", func(c *gin.Context) {
			controllers.UserRegister(c)
		})
		user.POST("/login", func(c *gin.Context) {
			controllers.UserLogin(c)
		})
	}

	// Seller routes dengan rate limiting ketat untuk auth
	seller := r.Group("/seller")
	seller.Use(authLimiter.TokenBucketMiddleware()) // Rate limit untuk auth endpoints
	{
		seller.POST("/register", func(c *gin.Context) {
			controllers.SellerRegister(c)
		})
		seller.POST("/login", func(c *gin.Context) {
			controllers.SellerLogin(c)
		})
	}

	// Public product routes dengan rate limiting
	product := r.Group("/product")
	{
		// Search dengan rate limit khusus
		product.GET("", searchLimiter.TokenBucketMiddleware(), controllers.SearchProduct)

		// Product detail dengan rate limit relaxed
		product.GET("/:id", relaxedLimiter.TokenBucketMiddleware(), controllers.GetProductDetail)

		// Recommendations dengan optional auth dan rate limit
		product.GET("/recommendations",
			middleware.OptionalAuthMiddleware(),
			searchLimiter.TokenBucketMiddleware(),
			controllers.GetRecommendations)
	}

	// Categories dengan rate limit relaxed
	r.GET("/categories", relaxedLimiter.TokenBucketMiddleware(), controllers.GetCategories)

	// Protected user routes dengan dynamic rate limiting berdasarkan role
	userProtected := r.Group("/user")
	userProtected.Use(middleware.AuthMiddleware("user"))
	userProtected.Use(middleware.DynamicRoleBasedMiddleware()) // Dynamic rate limit
	{
		userProtected.GET("/me", controllers.UserMe)
		userProtected.GET("/profile", controllers.GetUserProfile)
		userProtected.PUT("/profile", controllers.UpdateUserProfile)
	}

	// Protected seller routes dengan dynamic rate limiting
	sellerProtected := r.Group("/seller")
	sellerProtected.Use(middleware.AuthMiddleware("seller"))
	sellerProtected.Use(middleware.DynamicRoleBasedMiddleware()) // Dynamic rate limit
	{
		sellerProtected.GET("/me", controllers.SellerMe)
		sellerProtected.GET("/profile", controllers.GetSellerProfile)
		sellerProtected.PUT("/profile", controllers.UpdateSellerProfile)

		// Seller product management dengan rate limit moderate
		sellerProtected.POST("/products", moderateLimiter.TokenBucketMiddleware(), controllers.CreateProduct)
		sellerProtected.GET("/products", relaxedLimiter.TokenBucketMiddleware(), controllers.GetSellerProducts)
		sellerProtected.GET("/products/:id", relaxedLimiter.TokenBucketMiddleware(), controllers.GetSellerProduct)
		sellerProtected.PUT("/products/:id", moderateLimiter.TokenBucketMiddleware(), controllers.UpdateProduct)
		sellerProtected.DELETE("/products/:id", moderateLimiter.TokenBucketMiddleware(), controllers.DeleteProduct)
	}
}

// Alternative: Simpler rate limiting setup
func SimpleAuthRoutes(r *gin.Engine, db *gorm.DB) {
	// Single global rate limiter
	globalLimiter := middleware.ModerateRateLimiter()
	globalLimiter.StartCleanupRoutine()

	// Apply global rate limiting
	r.Use(globalLimiter.TokenBucketMiddleware())

	// Rest of routes without additional rate limiting...
	// (sama seperti route sebelumnya tapi tanpa middleware tambahan)
}

// Advanced: Custom rate limiting per endpoint
func AdvancedAuthRoutes(r *gin.Engine, db *gorm.DB) {
	// Different rate limiters for different needs
	authLimiter := middleware.AuthRateLimiter()     // 5/min untuk auth
	searchLimiter := middleware.SearchRateLimiter() // 30/min untuk search
	crudLimiter := middleware.ModerateRateLimiter() // 60/min untuk CRUD
	readLimiter := middleware.RelaxedRateLimiter()  // 120/min untuk read

	// Start cleanup
	authLimiter.StartCleanupRoutine()
	searchLimiter.StartCleanupRoutine()
	crudLimiter.StartCleanupRoutine()
	readLimiter.StartCleanupRoutine()

	// Auth endpoints - very strict
	auth := r.Group("/auth")
	auth.Use(authLimiter.SlidingWindowMiddleware()) // Menggunakan sliding window untuk auth
	{
		auth.POST("/user/register", controllers.UserRegister)
		auth.POST("/user/login", controllers.UserLogin)
		auth.POST("/seller/register", controllers.SellerRegister)
		auth.POST("/seller/login", controllers.SellerLogin)
	}

	// Public read endpoints - relaxed
	public := r.Group("/public")
	public.Use(readLimiter.TokenBucketMiddleware())
	{
		public.GET("/products", controllers.SearchProduct)
		public.GET("/products/:id", controllers.GetProductDetail)
		public.GET("/categories", controllers.GetCategories)
	}

	// Search endpoints - moderate
	search := r.Group("/search")
	search.Use(middleware.OptionalAuthMiddleware())
	search.Use(searchLimiter.TokenBucketMiddleware())
	{
		search.GET("/products", controllers.SearchProduct)
		search.GET("/recommendations", controllers.GetRecommendations)
	}

	// Protected endpoints with role-based rate limiting
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware("user,seller")) // Accept both roles
	protected.Use(middleware.DynamicRoleBasedMiddleware())
	{
		// User endpoints
		userGroup := protected.Group("/user")
		{
			userGroup.GET("/me", controllers.UserMe)
			userGroup.GET("/profile", controllers.GetUserProfile)
			userGroup.PUT("/profile", controllers.UpdateUserProfile)
		}

		// Seller endpoints
		sellerGroup := protected.Group("/seller")
		{
			sellerGroup.GET("/me", controllers.SellerMe)
			sellerGroup.GET("/profile", controllers.GetSellerProfile)
			sellerGroup.PUT("/profile", controllers.UpdateSellerProfile)
			sellerGroup.POST("/products", controllers.CreateProduct)
			sellerGroup.GET("/products", controllers.GetSellerProducts)
			sellerGroup.GET("/products/:id", controllers.GetSellerProduct)
			sellerGroup.PUT("/products/:id", controllers.UpdateProduct)
			sellerGroup.DELETE("/products/:id", controllers.DeleteProduct)
		}
	}
}
