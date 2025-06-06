package interfaces

import "assignment/internal/domain/entities"

type UserControllerInterface interface {
    CreateFriendship(user1Email, user2Email string) error
    GetFriendList(email string) ([]*entities.User, error)
    GetCommonFriends(email1, email2 string) ([]*entities.User, error)
    CreateSubscription(requestorEmail, targetEmail string) error
    CreateBlock(requestorEmail, targetEmail string) error
    GetRecipients(senderEmail, text string) ([]*entities.User, error)
}

type Controllers interface {
    UserController() UserControllerInterface
}