package controller

import (
	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
	"assignment/pkg/errors"
	"assignment/pkg/utils"
	"maps"
	"slices"
)

type userController struct {
	userRepo interfaces.UserRepositoryInterface
}

func NewUserController(userRepo interfaces.UserRepositoryInterface) interfaces.UserControllerInterface {
	return &userController{
		userRepo: userRepo,
	}
}

func (c *userController) CreateFriendship(user1Email, user2Email string) error {
	// Check for self-friendship
	if user1Email == user2Email {
		return errors.ErrCannotFriendSelf
	}

	// Get users from repository
	user1, err := c.userRepo.GetUserByEmail(user1Email)
	if err != nil {
		return err
	}

	user2, err := c.userRepo.GetUserByEmail(user2Email)
	if err != nil {
		return err
	}

	// Check if either user has blocked the other
	isBlocked, err := c.userRepo.CheckBidirectionalBlock(user1.ID, user2.ID)
	if err != nil {
		return err
	}
	if isBlocked {
		return errors.ErrUserBlocked
	}

	return c.userRepo.CreateFriendship(user1, user2)
}

func (c *userController) GetFriendList(email string) ([]*entities.User, error) {
	user, err := c.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	return c.userRepo.GetFriendList(user)
}

func (c *userController) GetCommonFriends(email1, email2 string) ([]*entities.User, error) {
	// Check for same user
	if email1 == email2 {
		return nil, errors.ErrCannotGetCommonFriendsWithSelf
	}

	// Get users from repository
	user1, err := c.userRepo.GetUserByEmail(email1)
	if err != nil {
		return nil, err
	}

	user2, err := c.userRepo.GetUserByEmail(email2)
	if err != nil {
		return nil, err
	}

	return c.userRepo.GetCommonFriends(user1, user2)
}

func (c *userController) CreateSubscription(requestorEmail, targetEmail string) error {
	requestor, err := c.userRepo.GetUserByEmail(requestorEmail)
	if err != nil {
		return err
	}

	target, err := c.userRepo.GetUserByEmail(targetEmail)
	if err != nil {
		return err
	}

	// Check if either user has blocked the other
	isBlocked, err := c.userRepo.CheckBidirectionalBlock(requestor.ID, target.ID)
	if err != nil {
		return err
	}
	if isBlocked {
		return errors.ErrUserBlocked
	}

	return c.userRepo.CreateSubscription(requestor, target)
}

func (c *userController) CreateBlock(requestorEmail, targetEmail string) error {
	requestor, err := c.userRepo.GetUserByEmail(requestorEmail)
	if err != nil {
		return err
	}

	target, err := c.userRepo.GetUserByEmail(targetEmail)
	if err != nil {
		return err
	}

	return c.userRepo.CreateBlockTx(requestor, target)
}

func (c *userController) GetRecipients(senderEmail, text string) ([]*entities.User, error) {
	sender, err := c.userRepo.GetUserByEmail(senderEmail)
	if err != nil {
		return nil, err
	}

	mentionedEmails := utils.ExtractEmailsFromText(text)
	var mentionedUsers []*entities.User

	if len(mentionedEmails) > 0 {
		mentionedUsers, err = c.userRepo.GetUsersByEmails(mentionedEmails)
		if err != nil {
			return nil, err
		}
	}

	senderFriends, err := c.userRepo.GetFriendList(sender)
	if err != nil {
		return nil, err
	}

	subscribers, err := c.userRepo.GetSubscribersByUserID(sender.ID)
	if err != nil {
		return nil, err
	}

	recipients := make(map[int]*entities.User)

	for _, friend := range senderFriends {
		recipients[friend.ID] = friend
	}

	for _, subscriber := range subscribers {
		recipients[subscriber.ID] = subscriber
	}

	// Batch check bidirectional blocks for all mentioned users
	if len(mentionedUsers) > 0 {
		mentionedUserIDs := make([]int, len(mentionedUsers))
		for i, user := range mentionedUsers {
			mentionedUserIDs[i] = user.ID
		}
		
		blockedUsers, err := c.userRepo.CheckBidirectionalBlocksBatch(sender.ID, mentionedUserIDs)
		if err != nil {
			return nil, err
		}
		
		for _, mentioned := range mentionedUsers {
			if !blockedUsers[mentioned.ID] {
				recipients[mentioned.ID] = mentioned
			}
		}
	}

	return slices.Collect(maps.Values(recipients)), nil
}
