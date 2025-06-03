package controller

import "assignment/internal/domain/interfaces"

type controllers struct {
    userController interfaces.UserControllerInterface
}

func NewControllers(repos interfaces.Repositories) interfaces.Controllers {
    return &controllers{
        userController: NewUserController(repos.UserRepository()),
    }
}

func (c *controllers) UserController() interfaces.UserControllerInterface {
    return c.userController
}