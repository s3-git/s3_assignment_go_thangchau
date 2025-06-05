package controller

import (
	//"assignment/internal/domain"
	"assignment/internal/domain/interfaces"
	"errors"
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
    if user1Email == user2Email {
        return errors.New("cannot befriend self")
    }

	return c.userRepo.CreateFriendship(user1Email, user2Email)
}