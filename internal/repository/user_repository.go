package repository

import (
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

func (r *userRepository) CreateFriendship(user1Email, user2Email string) error {
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

	// Get both users' IDs
	users, err := models.Users(
		qm.Select(models.UserColumns.ID, models.UserColumns.Email),
		models.UserWhere.Email.IN([]string{user1Email, user2Email}),
	).All(context.Background(), tx)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "Failed to fetch users")
	}

	// Check that both users exist
	if len(users) != 2 {
		var missingEmails []string
		foundEmails := make(map[string]bool)
		for _, user := range users {
			foundEmails[user.Email] = true
		}
		if !foundEmails[user1Email] {
			missingEmails = append(missingEmails, user1Email)
		}
		if !foundEmails[user2Email] {
			missingEmails = append(missingEmails, user2Email)
		}
		return errors.Newf(errors.ErrorTypeNotFound, "User(s) not found: %v", missingEmails)
	}

	// Map emails to user IDs
	userIDMap := make(map[string]int)
	for _, user := range users {
		userIDMap[user.Email] = user.ID
	}

	user1ID := userIDMap[user1Email]
	user2ID := userIDMap[user2Email]

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

func (r *userRepository) GetFriendList(email string) error {
	return nil
}
