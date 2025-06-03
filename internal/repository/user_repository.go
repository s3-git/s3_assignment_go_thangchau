package repository

import (
	"database/sql"
	"assignment/internal/domain"
	"assignment/internal/domain/interfaces"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepositoryInterface {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id int) (domain.User, error) {
	return domain.User{}, nil
}

func (r *userRepository) CreateFriendship(userID1, userID2 int) error {
	query := "INSERT INTO friendships (user_id_1, user_id_2) VALUES ($1, $2)"
	_, err := r.db.Exec(query, userID1, userID2)
	return err
}
