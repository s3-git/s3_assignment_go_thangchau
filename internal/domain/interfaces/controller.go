package interfaces

//import "assignment/internal/domain"

type UserControllerInterface interface {
    CreateFriendships(userID1, userID2 int) error
    //GetFriendList(userID int) ([]*domain.User, error)
    //GetCommonFriends(userID1, userID2 int) ([]*domain.User, error)
    //Subscription(userID, targetID int) error
    //Block(userID, targetID int) error
    //GetRecipients(userID int) ([]*domain.User, error)
}

type Controllers interface {
    UserController() UserControllerInterface
}