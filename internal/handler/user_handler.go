package handler

import (
	"assignment/internal/domain/interfaces"
	"assignment/pkg/errors"
	"assignment/pkg/validator"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userController interfaces.UserControllerInterface
}

func NewUserHandler(userController interfaces.UserControllerInterface) *UserHandler {
	return &UserHandler{
		userController: userController,
	}
}

func (h *UserHandler) CreateFriendships(c *gin.Context) {
	var req CreateFriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if ValidateCreateFriendshipRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	if err := h.userController.CreateFriendship(req.Friends[0], req.Friends[1]); err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func (h *UserHandler) GetFriendList(c *gin.Context) {
	// TODO: Implement friend list handler
	c.JSON(501, gin.H{"error": "Not implemented"})
}

func (h *UserHandler) GetCommonFriends(c *gin.Context) {
	// TODO: Implement common friends handler
	c.JSON(501, gin.H{"error": "Not implemented"})
}

func (h *UserHandler) CreateSubscription(c *gin.Context) {
	// TODO: Implement subscription handler
	c.JSON(501, gin.H{"error": "Not implemented"})
}

func (h *UserHandler) CreateBlock(c *gin.Context) {
	// TODO: Implement block handler
	c.JSON(501, gin.H{"error": "Not implemented"})
}

func (h *UserHandler) GetRecipients(c *gin.Context) {
	// TODO: Implement recipients handler
	c.JSON(501, gin.H{"error": "Not implemented"})
}
