package repository

import (
    "database/sql"
    "assignment/internal/domain/interfaces"
)

type repositories struct {
    userRepo interfaces.UserRepositoryInterface
}

func NewRepositories(db *sql.DB) interfaces.Repositories {
    return &repositories{
        userRepo: NewUserRepository(db),
    }
}

func (r *repositories) UserRepository() interfaces.UserRepositoryInterface {
    return r.userRepo
}