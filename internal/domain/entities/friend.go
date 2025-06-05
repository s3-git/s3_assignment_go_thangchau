package entities

import (
	"assignment/pkg/utils"
	"errors"
	"fmt"
	"strings"
)

type Friend struct {
	ID      string
	User1ID int
	User2ID int
}

type CreateFriendshipRequest struct {
	Friends []string `json:"friends"`
}

func (r *CreateFriendshipRequest) Validate() error {
	if len(r.Friends) != 2 {
		return errors.New("exactly 2 friends required")
	}

	if r.Friends == nil {
		return errors.New("friends field is required")
	}

	// Check for empty values
	for i, email := range r.Friends {
		if strings.TrimSpace(email) == "" {
			return fmt.Errorf("email at position %d cannot be empty", i)
		}
	}

	// Email format validation
	for _, email := range r.Friends {
		if !utils.ValidateEmail(email) {
			return fmt.Errorf("invalid email format: %s", email)
		}
	}

	return nil
}
