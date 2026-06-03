package chat

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware, verifiedMiddleware gin.HandlerFunc) {
	group := r.Group("/chat")
	group.Use(authMiddleware, verifiedMiddleware)
	{
		group.POST("/conversations", handler.CreateOrGetConversation)
		group.GET("/conversations", handler.ListConversations)
		group.GET("/conversations/:id/messages", handler.ListMessages)
		group.POST("/conversations/:id/messages", handler.SendMessage)
		group.PUT("/conversations/:id/read", handler.MarkRead)
	}
}
