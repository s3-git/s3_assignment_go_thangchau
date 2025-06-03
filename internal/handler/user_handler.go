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
	UserID1 int `json:"user_id_1" binding:"required"`
	UserID2 int `json:"user_id_2" binding:"required"`
}

func (h *UserHandler) CreateFriendships(c *gin.Context) {
	var req CreateFriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userController.CreateFriendships(req.UserID1, req.UserID2); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Friendship created successfully"})
}
