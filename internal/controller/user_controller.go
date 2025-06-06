package controller

import (
	"assignment/internal/domain/interfaces"
	"assignment/pkg/errors"
)

type userController struct {
	userRepo interfaces.UserRepositoryInterface
}

func NewUserController(userRepo interfaces.UserRepositoryInterface) interfaces.UserControllerInterface {
	return &userController{
		userRepo: userRepo,
	}
}

func (c *userController) CreateFriendship(user1Email, user2Email string) error {
	// Check for self-friendship
	if user1Email == user2Email {
		return errors.ErrCannotFriendSelf
	}

	return c.userRepo.CreateFriendship(user1Email, user2Email)
}

func (c *userController) GetFriendList(email string) error {
	return nil
}
