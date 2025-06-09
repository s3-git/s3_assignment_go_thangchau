package interfaces

import "assignment/internal/domain/entities"

type UserRepositoryInterface interface {
    CreateFriendship(user1, user2 *entities.User) error
    GetFriendList(user *entities.User) ([]*entities.User, error)
    GetCommonFriends(user1, user2 *entities.User) ([]*entities.User, error)
    CreateSubscription(requestor, target *entities.User) error
    CreateBlockTx(requestor, target *entities.User) error
    CheckBlockExists(requestorID, targetID int) (bool, error)
    CheckBidirectionalBlock(user1ID, user2ID int) (bool, error)
    GetRecipients(sender *entities.User, mentionedUsers []*entities.User) ([]*entities.User, error)
    UserExists(email string) (*entities.User, error)
    GetUserByEmail(email string) (*entities.User, error)
}

type Repositories interface {
    UserRepository() UserRepositoryInterface
}