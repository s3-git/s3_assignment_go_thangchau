package repository

import (
	"assignment/internal/domain/interfaces"
	"database/sql"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepositoryInterface {
	return &userRepository{db: db}
}

// TODO: replace with sqlboiler
func (r *userRepository) CreateFriendship(userID1, userID2 string) error {
	query := "INSERT INTO friendships (user_id_1, user_id_2) VALUES ($1, $2)"
	_, err := r.db.Exec(query, userID1, userID2)
	return err
}
