package interfaces

import "assignment/internal/domain/entities"

type UserRepositoryInterface interface {
    CreateFriendship(user1, user2 *entities.User) error
    GetFriendList(user *entities.User) ([]*entities.User, error)
    GetCommonFriends(user1, user2 *entities.User) ([]*entities.User, error)
    CreateSubscription(requestor, target *entities.User) error
    CreateBlock(requestor, target *entities.User) error
    GetRecipients(sender *entities.User, mentionedUsers []*entities.User) ([]*entities.User, error)
    UserExists(email string) (*entities.User, error)
    GetUserByEmail(email string) (*entities.User, error)
}

type Repositories interface {
    UserRepository() UserRepositoryInterface
}