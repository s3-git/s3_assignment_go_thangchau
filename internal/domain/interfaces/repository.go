package interfaces

//import "assignment/internal/domain"

type UserRepositoryInterface interface {
    CreateFriendship(userID1, userID2 int) error
}

type Repositories interface {
    UserRepository() UserRepositoryInterface
}