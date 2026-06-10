package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware, readableMiddleware, normalMiddleware gin.HandlerFunc) {
	group := r.Group("/users")
	group.Use(authMiddleware, readableMiddleware)

	group.GET("/me", handler.Me)
	group.GET("/me/restriction", handler.Restriction)
	group.GET("/:id/public", handler.PublicProfile)
	group.PUT("/profile", normalMiddleware, handler.UpdateProfile)
	group.POST("/avatar", normalMiddleware, handler.UploadAvatar)
	group.POST("/student-verify", normalMiddleware, handler.SubmitStudentVerification)
	group.GET("/student-verify", handler.StudentVerification)
	group.POST("/cancel", normalMiddleware, handler.CancelAccount)
}

func RegisterAdminRoutes(r *gin.Engine, handler *Handler, authMiddleware, adminMiddleware, superAdminMiddleware gin.HandlerFunc) {
	group := r.Group("/admin/users")
	group.Use(authMiddleware, adminMiddleware)

	group.GET("", handler.ListAdminUsers)
	group.GET("/student-verifications", handler.ListStudentVerifications)
	group.GET("/:id", handler.AdminUserDetail)
	group.PUT("/:id/student-verify", handler.ReviewStudentVerification)
	group.PUT("/:id/status", handler.UpdateAccountStatus)
	group.PUT("/:id/role", superAdminMiddleware, handler.UpdateRole)
	group.GET("/:id/products", handler.ListUserProducts)
	group.GET("/:id/orders", handler.ListUserOrders)
	group.GET("/:id/reports", handler.ListUserReports)
	group.GET("/:id/appeals", handler.ListUserAppeals)
	group.GET("/:id/reviews", handler.ListUserReviews)
	group.GET("/:id/logs", handler.ListUserLogs)
}
