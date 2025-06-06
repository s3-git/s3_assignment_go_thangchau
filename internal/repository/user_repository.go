package repository

import (
	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
	"assignment/internal/infrastructure/database/models"
	"assignment/pkg/errors"
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"
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
	// TODO: Implement friend list retrieval
	return nil, nil
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
