package repository

import (
	"assignment/internal/domain/interfaces"
	"assignment/internal/infrastructure/database/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepositoryInterface {
	return &userRepository{db: db}
}

func (r *userRepository) CreateFriendship(user1Email, user2Email string) error {//TODO: some error status cannot be 500
	// Get first user's ID
	user1, err := models.Users(
		qm.Select(models.UserColumns.ID),
		models.UserWhere.Email.EQ(user1Email),
	).One(context.Background(), r.db)
	if err != nil {
		return err
	}

	// Get second user's ID
	user2, err := models.Users(
		qm.Select(models.UserColumns.ID),
		models.UserWhere.Email.EQ(user2Email),
	).One(context.Background(), r.db)
	if err != nil {
		return err
	}

	firstUserID := user1.ID
	secondUserID := user2.ID

	// Check which ID is smaller
	if user1.ID > user2.ID {
		firstUserID = user2.ID
		secondUserID = user1.ID
	}

	// Check if already friends
	exists, err := models.Friends(
		models.FriendWhere.User1ID.EQ(firstUserID),
		models.FriendWhere.User2ID.EQ(secondUserID),
	).Exists(context.Background(), r.db)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("friendship already created between %s and %s", user1Email, user2Email)
	}

	friend := &models.Friend{
		User1ID: firstUserID,
		User2ID: secondUserID,
	}

	err = friend.Insert(context.Background(), r.db, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetFriendList(email string) error {
	return nil
}