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

func (c *userController) CreateFriendships(userID1, userID2 string) error {
    if userID1 == userID2 {
        return errors.New("cannot befriend yourself")
    }
    
    // Check if users exist
    // user1, err := c.userRepo.GetByID(userID1)
    // if err != nil || user1 == nil {
    //     return errors.New("user1 not found")
    // }
    
    // user2, err := c.userRepo.GetByID(userID2)
    // if err != nil || user2 == nil {
    //     return errors.New("user2 not found")
    // }

    // TODO: check if already friends
    
    return c.userRepo.CreateFriendship(userID1, userID2) //todo: pass email not id
}