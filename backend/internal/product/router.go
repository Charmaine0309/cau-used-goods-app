package product

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware, adminMiddleware gin.HandlerFunc) {
	r.GET("/categories", handler.ListCategories)
	r.GET("/products", handler.ListProducts)
	r.GET("/products/:id", handler.GetProductByID)

	adminCategories := r.Group("/admin/categories")
	adminCategories.Use(authMiddleware, adminMiddleware)
	{
		adminCategories.GET("", handler.AdminListCategories)
		adminCategories.POST("", handler.AdminCreateCategory)
		adminCategories.PUT("/:id", handler.AdminUpdateCategory)
		adminCategories.PUT("/:id/status", handler.AdminUpdateCategoryStatus)
		adminCategories.DELETE("/:id", handler.AdminDeleteCategory)
	}

	adminProducts := r.Group("/admin/products")
	adminProducts.Use(authMiddleware, adminMiddleware)
	{
		adminProducts.PUT("/:id/status", handler.AdminUpdateProductStatus)
	}

	products := r.Group("/products")
	products.Use(authMiddleware)
	{
		products.POST("", handler.CreateProduct)
		products.GET("/my", handler.ListMyProducts)
		products.PUT("/:id", handler.UpdateProduct)
		products.PUT("/:id/status", handler.UpdateProductStatus)
		products.POST("/:id/images", handler.AddProductImages)
		products.DELETE("/:id", handler.DeleteProduct)
	}
}
