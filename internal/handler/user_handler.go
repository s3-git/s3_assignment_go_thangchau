package handler

import (
	"assignment/internal/domain/interfaces"
	"assignment/pkg/errors"
	"assignment/pkg/validator"
	"net/http"

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

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *UserHandler) GetFriendList(c *gin.Context) {
	var req GetFriendListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if ValidateGetFriendListRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	friends, err := h.userController.GetFriendList(req.Email)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	friendEmails := make([]string, len(friends))
	for i, friend := range friends {
		friendEmails[i] = friend.Email
	}

	response := FriendListResponse{
		Success: true,
		Friends: friendEmails,
		Count:   len(friendEmails),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetCommonFriends(c *gin.Context) {
	var req GetCommonFriendsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if ValidateGetCommonFriendsRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	friends, err := h.userController.GetCommonFriends(req.Friends[0], req.Friends[1])
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	friendEmails := make([]string, len(friends))
	for i, friend := range friends {
		friendEmails[i] = friend.Email
	}

	response := CommonFriendsResponse{
		Success: true,
		Friends: friendEmails,
		Count:   len(friendEmails),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) CreateSubscription(c *gin.Context) {
	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if ValidateSubscriptionRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	if err := h.userController.CreateSubscription(req.Requestor, req.Target); err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *UserHandler) CreateBlock(c *gin.Context) {
	var req CreateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if ValidateCreateBlockRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	if err := h.userController.CreateBlock(req.Requestor, req.Target); err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *UserHandler) GetRecipients(c *gin.Context) {
	var req GetRecipientsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.SendBadRequest(c, "Invalid request format", err.Error())
		return
	}

	v := validator.New()
	if ValidateGetRecipientsRequest(v, &req); !v.Valid() {
		errors.HandleValidationErrors(c, v.Errors)
		return
	}

	recipients, err := h.userController.GetRecipients(req.Sender, req.Text)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	recipientEmails := make([]string, len(recipients))
	for i, recipient := range recipients {
		recipientEmails[i] = recipient.Email
	}

	response := RecipientsResponse{
		Success:    true,
		Recipients: recipientEmails,
	}

	c.JSON(http.StatusOK, response)
}
