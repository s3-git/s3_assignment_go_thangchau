package handler

import (
	"net/http"

	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
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
	var req entities.CreateFriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	v := validator.New()
	if entities.ValidateCreateFriendshipRequest(v, &req); !v.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": v.Errors})
		return
	}

	if err := h.userController.CreateFriendship(req.Friends[0], req.Friends[1]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func (h *UserHandler) GetFriendList(c *gin.Context) {
	var req entities.GetFriendListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	v := validator.New()
	if entities.ValidateGetFriendlistRequest(v, &req); !v.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": v.Errors})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}
