package interfaces

//import "assignment/internal/domain"

type UserRepositoryInterface interface {
    CreateFriendship(user1Email, user2Email string) error
}

type Repositories interface {
    UserRepository() UserRepositoryInterface
}