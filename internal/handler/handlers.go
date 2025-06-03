package handler

import "assignment/internal/domain/interfaces"

type Handlers struct {
    UserHandler *UserHandler
}

func NewHandlers(controllers interfaces.Controllers) *Handlers {
    return &Handlers{
        UserHandler: NewUserHandler(controllers.UserController()),
    }
}