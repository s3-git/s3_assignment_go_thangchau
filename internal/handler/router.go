package handler

import (
	"github.com/gin-gonic/gin"

	"assignment/internal/domain/interfaces"
)

func SetupRoutes(r *gin.Engine, controllers interfaces.Controllers) {
	handlers := NewHandlers(controllers)

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/user")
		{
			users.POST("/friends", handlers.UserHandler.CreateFriendships)
			users.POST("/friends/list", handlers.UserHandler.GetFriendList)
			users.POST("/friends/common", handlers.UserHandler.GetCommonFriends)
			users.POST("/subscriptions", handlers.UserHandler.CreateSubscription)
			users.POST("/blocks", handlers.UserHandler.CreateBlock)
			users.POST("/recipients", handlers.UserHandler.GetRecipients)
		}
	}
}
