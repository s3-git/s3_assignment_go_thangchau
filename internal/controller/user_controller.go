package controller

import (
	"assignment/internal/domain/entities"
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

	// Get users from repository
	user1, err := c.userRepo.GetUserByEmail(user1Email)
	if err != nil {
		return err
	}

	user2, err := c.userRepo.GetUserByEmail(user2Email)
	if err != nil {
		return err
	}

	return c.userRepo.CreateFriendship(user1, user2)
}

func (c *userController) GetFriendList(email string) ([]*entities.User, error) {
	user, err := c.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	return c.userRepo.GetFriendList(user)
}

func (c *userController) GetCommonFriends(email1, email2 string) ([]*entities.User, error) {
	// Check for same user
	if email1 == email2 {
		return nil, errors.ErrCannotGetCommonFriendsWithSelf
	}

	// Get users from repository
	user1, err := c.userRepo.GetUserByEmail(email1)
	if err != nil {
		return nil, err
	}

	user2, err := c.userRepo.GetUserByEmail(email2)
	if err != nil {
		return nil, err
	}

	return c.userRepo.GetCommonFriends(user1, user2)
}

func (c *userController) CreateSubscription(requestorEmail, targetEmail string) error {
	requestor, err := c.userRepo.GetUserByEmail(requestorEmail)
	if err != nil {
		return err
	}

	target, err := c.userRepo.GetUserByEmail(targetEmail)
	if err != nil {
		return err
	}

	return c.userRepo.CreateSubscription(requestor, target)
}

func (c *userController) CreateBlock(requestorEmail, targetEmail string) error {
	// TODO: Implement block business logic
	return nil
}

func (c *userController) GetRecipients(senderEmail, text string) ([]*entities.User, error) {
	// TODO: Implement recipients business logic
	return nil, nil
}
