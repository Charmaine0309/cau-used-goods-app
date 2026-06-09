package appeal

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware, readableMiddleware, verifiedMiddleware, adminMiddleware gin.HandlerFunc) {
	group := r.Group("/appeals")
	group.Use(authMiddleware, readableMiddleware)
	{
		group.POST("", handler.Create)
		group.GET("/my", handler.ListMy)
		group.GET("/:id", handler.GetByID)
		group.POST("/:id/close", handler.Close)
	}

	adminGroup := r.Group("/admin/appeals")
	adminGroup.Use(authMiddleware, adminMiddleware)
	{
		adminGroup.GET("", handler.AdminList)
		adminGroup.GET("/:id", handler.AdminGetByID)
		adminGroup.POST("/:id/processing", handler.AdminMarkProcessing)
		adminGroup.POST("/:id/handle", handler.AdminHandle)
	}
}
