package handler

import (
	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
	"assignment/pkg/errors"
	"assignment/pkg/response"
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
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if entities.ValidateCreateFriendshipRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	if err := h.userController.CreateFriendship(req.Friends[0], req.Friends[1]); err != nil {
		errors.HandleError(c, err)
		return
	}

	response.SendCreated(c, nil, "Friendship created successfully")
}

func (h *UserHandler) GetFriendList(c *gin.Context) {
	var req entities.GetFriendListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if entities.ValidateGetFriendlistRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	// TODO: Implement actual friend list retrieval
	response.SendSuccess(c, nil, "Friend list retrieved successfully")
}
