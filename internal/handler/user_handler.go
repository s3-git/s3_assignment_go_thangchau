package handler

import (
	"net/http"

	"assignment/internal/domain/interfaces"

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

type CreateFriendshipRequest struct {
	Friends []string `json:"friends"`
}

func (h *UserHandler) CreateFriendships(c *gin.Context) {
	var req CreateFriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	//TODO: custom validator

	if len(req.Friends) != 2 {
		//todo: util method
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"}) //todo: use utils to handle
		return
	}

	if req.Friends[0] == req.Friends[1] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot befriend self"}) //todo: use utils to handle
		return
	}

	if err := h.userController.CreateFriendships(req.Friends[0], req.Friends[1]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}
