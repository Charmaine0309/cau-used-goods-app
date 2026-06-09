package order

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware, readableMiddleware, verifiedMiddleware, adminMiddleware gin.HandlerFunc) {
	group := r.Group("/orders")
	group.Use(authMiddleware, readableMiddleware)
	{
		group.POST("", verifiedMiddleware, handler.Create)
		group.GET("", handler.ListMyOrders)
		group.GET("/:id", handler.GetByID)
		group.POST("/:id/confirm", verifiedMiddleware, handler.Confirm)
		group.POST("/:id/cancel", verifiedMiddleware, handler.Cancel)
		group.POST("/:id/complete", verifiedMiddleware, handler.Complete)
	}

	adminGroup := r.Group("/admin/orders")
	adminGroup.Use(authMiddleware, adminMiddleware)
	{
		adminGroup.POST("/:id/exception-close", handler.AdminExceptionClose)
		adminGroup.PUT("/:id/status", handler.AdminUpdateStatus)
	}
}
