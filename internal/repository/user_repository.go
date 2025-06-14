package repository

import (
	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
	"assignment/internal/infrastructure/database/models"
	"assignment/pkg/errors"
	"assignment/pkg/utils"
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepositoryInterface {
	return &userRepository{db: db}
}

func (r *userRepository) CreateFriendship(user1, user2 *entities.User) error {
	// Begin transaction
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to begin transaction")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Use the IDs from domain entities directly
	user1ID := user1.ID
	user2ID := user2.ID

	// Ensure consistent ordering (smaller ID first)
	firstUserID := user1ID
	secondUserID := user2ID
	if user1ID > user2ID {
		firstUserID = user2ID
		secondUserID = user1ID
	}

	// Try to insert friendship directly - let database constraint handle duplicates
	friend := &models.Friend{
		User1ID: firstUserID,
		User2ID: secondUserID,
	}

	err = friend.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		return errors.FromError(err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to commit transaction")
	}

	return nil
}

func (r *userRepository) GetFriendList(user *entities.User) ([]*entities.User, error) {
	// First verify that the user exists
	_, err := models.Users(
		models.UserWhere.ID.EQ(user.ID),
	).One(context.Background(), r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Newf(errors.ErrorTypeNotFound, "User with ID %d not found", user.ID)
		}
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch user")
	}

	// Get friendships where this user is user1
	user1Friends, err := models.Friends(
		models.FriendWhere.User1ID.EQ(user.ID),
		qm.Load(models.FriendRels.User2),
	).All(context.Background(), r.db)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch user1 friends")
	}

	// Get friendships where this user is user2
	user2Friends, err := models.Friends(
		models.FriendWhere.User2ID.EQ(user.ID),
		qm.Load(models.FriendRels.User1),
	).All(context.Background(), r.db)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch user2 friends")
	}

	// Collect all friend users
	friendMap := make(map[int]*entities.User)

	// Add friends from user1 relationships (where current user is user1)
	for _, friendship := range user1Friends {
		if friendship.R != nil && friendship.R.User2 != nil {
			friendUser := friendship.R.User2
			friendMap[friendUser.ID] = &entities.User{
				ID:    friendUser.ID,
				Email: friendUser.Email,
			}
		}
	}

	// Add friends from user2 relationships (where current user is user2)
	for _, friendship := range user2Friends {
		if friendship.R != nil && friendship.R.User1 != nil {
			friendUser := friendship.R.User1
			friendMap[friendUser.ID] = &entities.User{
				ID:    friendUser.ID,
				Email: friendUser.Email,
			}
		}
	}

	// Convert map to slice and sort by email
	var friends []*entities.User
	for _, friend := range friendMap {
		friends = append(friends, friend)
	}

	// Sort friends by email using utility function
	utils.SortUsersByEmail(friends)

	return friends, nil
}

func (r *userRepository) GetCommonFriends(user1, user2 *entities.User) ([]*entities.User, error) {
	// Get friends of user1
	user1Friends, err := r.GetFriendList(user1)
	if err != nil {
		return nil, err
	}

	// Get friends of user2
	user2Friends, err := r.GetFriendList(user2)
	if err != nil {
		return nil, err
	}

	// Create a map of user1's friends for efficient lookup
	user1FriendMap := make(map[string]*entities.User)
	for _, friend := range user1Friends {
		user1FriendMap[friend.Email] = friend
	}

	// Find common friends
	var commonFriends []*entities.User
	for _, friend := range user2Friends {
		if _, exists := user1FriendMap[friend.Email]; exists {
			commonFriends = append(commonFriends, friend)
		}
	}

	// Sort common friends by email using utility function
	utils.SortUsersByEmail(commonFriends)

	return commonFriends, nil
}

func (r *userRepository) CreateSubscription(requestor, target *entities.User) error {
	subscription := &models.Subscription{
		SubscriberID: requestor.ID,
		TargetID:     target.ID,
	}

	err := subscription.Insert(context.Background(), r.db, boil.Infer())
	if err != nil {
		return errors.FromError(err)
	}

	return nil
}

func (r *userRepository) CreateBlockTx(requestor, target *entities.User) error {
	// Begin transaction
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to begin transaction")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Remove friendship if it exists (bidirectional)
	firstUserID := requestor.ID
	secondUserID := target.ID
	if requestor.ID > target.ID {
		firstUserID = target.ID
		secondUserID = requestor.ID
	}

	_, err = models.Friends(
		models.FriendWhere.User1ID.EQ(firstUserID),
		models.FriendWhere.User2ID.EQ(secondUserID),
	).DeleteAll(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to delete friendship")
	}

	// 2. Remove subscriptions from both sides
	// Remove requestor's subscription to target
	_, err = models.Subscriptions(
		models.SubscriptionWhere.SubscriberID.EQ(requestor.ID),
		models.SubscriptionWhere.TargetID.EQ(target.ID),
	).DeleteAll(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to delete requestor subscription")
	}

	// Remove target's subscription to requestor
	_, err = models.Subscriptions(
		models.SubscriptionWhere.SubscriberID.EQ(target.ID),
		models.SubscriptionWhere.TargetID.EQ(requestor.ID),
	).DeleteAll(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to delete target subscription")
	}

	// 3. Create the block
	block := &models.Block{
		BlockerID: requestor.ID,
		BlockedID: target.ID,
	}

	err = block.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		return errors.FromError(err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to commit transaction")
	}

	return nil
}

func (r *userRepository) CheckBlockExists(requestorID, targetID int) (bool, error) {
	_, err := models.Blocks(
		models.BlockWhere.BlockerID.EQ(requestorID),
		models.BlockWhere.BlockedID.EQ(targetID),
	).One(context.Background(), r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to check block existence")
	}
	return true, nil
}

func (r *userRepository) CheckBidirectionalBlock(user1ID, user2ID int) (bool, error) {
	// Check if user1 blocks user2
	blocked1, err := r.CheckBlockExists(user1ID, user2ID)
	if err != nil {
		return false, err
	}
	if blocked1 {
		return true, nil
	}

	// Check if user2 blocks user1
	blocked2, err := r.CheckBlockExists(user2ID, user1ID)
	if err != nil {
		return false, err
	}
	return blocked2, nil
}

func (r *userRepository) CheckBidirectionalBlocksBatch(senderID int, userIDs []int) (map[int]bool, error) {
	if len(userIDs) == 0 {
		return make(map[int]bool), nil
	}

	// Build pairs for both directions: (sender -> user) and (user -> sender)
	var blockerIDs []int
	var blockedIDs []int

	for _, userID := range userIDs {
		// Check sender blocks user
		blockerIDs = append(blockerIDs, senderID)
		blockedIDs = append(blockedIDs, userID)
		// Check user blocks sender
		blockerIDs = append(blockerIDs, userID)
		blockedIDs = append(blockedIDs, senderID)
	}

	blocks, err := models.Blocks(
		qm.Where("(blocker_id, blocked_id) IN (SELECT unnest($1::int[]), unnest($2::int[]))",
			pq.Array(blockerIDs), pq.Array(blockedIDs)),
	).All(context.Background(), r.db)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to check bidirectional blocks")
	}

	// Build result map
	result := make(map[int]bool)
	for _, userID := range userIDs {
		result[userID] = false
	}

	// Mark users as blocked if any block exists
	for _, block := range blocks {
		if block.BlockerID == senderID {
			// Sender blocks this user
			result[block.BlockedID] = true
		} else {
			// This user blocks sender
			result[block.BlockerID] = true
		}
	}

	return result, nil
}

func (r *userRepository) GetUserByEmail(email string) (*entities.User, error) {
	user, err := models.Users(
		models.UserWhere.Email.EQ(email),
	).One(context.Background(), r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", email)
		}
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch user")
	}

	return &entities.User{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (r *userRepository) GetUsersByEmails(emails []string) ([]*entities.User, error) {
	if len(emails) == 0 {
		return []*entities.User{}, nil
	}

	users, err := models.Users(
		models.UserWhere.Email.IN(emails),
	).All(context.Background(), r.db)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch users by emails")
	}

	result := make([]*entities.User, len(users))
	for i, user := range users {
		result[i] = &entities.User{
			ID:    user.ID,
			Email: user.Email,
		}
	}

	return result, nil
}

func (r *userRepository) GetSubscribersByUserID(userID int) ([]*entities.User, error) {
	// Get all subscriptions where this user is the target
	subscriptions, err := models.Subscriptions(
		models.SubscriptionWhere.TargetID.EQ(userID),
		qm.Load(models.SubscriptionRels.Subscriber),
	).All(context.Background(), r.db)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch subscribers")
	}

	// Convert to entities
	var subscribers []*entities.User
	for _, subscription := range subscriptions {
		if subscription.R != nil && subscription.R.Subscriber != nil {
			subscriber := subscription.R.Subscriber
			subscribers = append(subscribers, &entities.User{
				ID:    subscriber.ID,
				Email: subscriber.Email,
			})
		}
	}

	// Sort subscribers by email using utility function
	utils.SortUsersByEmail(subscribers)

	return subscribers, nil
}
