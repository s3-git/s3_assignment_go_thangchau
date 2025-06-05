package handler

import (
	"github.com/gin-gonic/gin"

	"assignment/internal/domain/interfaces"
)

func SetupRoutes(r *gin.Engine, controllers interfaces.Controllers) {
	handlers := NewHandlers(controllers)

	// API version grouping
	v1 := r.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/user")
		{
			users.POST("/friends", handlers.UserHandler.CreateFriendships)
			// users.POST("/friends/list", controllers.UserController.GetFriendList)
			// users.POST("/friends/common", controllers.UserController.GetCommonFriends)
			// users.POST("/subscriptions", controllers.UserController.Subscription)
			// users.POST("/blocks", controllers.UserController.Block)
			// users.POST("/recipients", controllers.UserController.Recipients)
		}
	}
}
