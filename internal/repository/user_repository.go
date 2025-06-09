package repository

import (
	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
	"assignment/internal/infrastructure/database/models"
	"assignment/pkg/errors"
	"context"
	"database/sql"

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

	// Sort by email for consistent ordering
	for i := range len(friends) - 1 {
		for j := i + 1; j < len(friends); j++ {
			if friends[i].Email > friends[j].Email {
				friends[i], friends[j] = friends[j], friends[i]
			}
		}
	}

	return friends, nil
}

func (r *userRepository) GetCommonFriends(user1, user2 *entities.User) ([]*entities.User, error) {
	// TODO: Implement common friends retrieval
	return nil, nil
}

func (r *userRepository) CreateSubscription(requestor, target *entities.User) error {
	// TODO: Implement subscription creation
	return nil
}

func (r *userRepository) CreateBlock(requestor, target *entities.User) error {
	// TODO: Implement block creation
	return nil
}

func (r *userRepository) GetRecipients(sender *entities.User, mentionedUsers []*entities.User) ([]*entities.User, error) {
	// TODO: Implement recipients retrieval
	return nil, nil
}

func (r *userRepository) UserExists(email string) (*entities.User, error) {
	// TODO: Implement user existence check
	return nil, nil
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
