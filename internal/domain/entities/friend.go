package entities

import (
	"assignment/pkg/utils"
	"assignment/pkg/validator"
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

func ValidateCreateFriendshipRequest(v *validator.Validator, r *CreateFriendshipRequest) {
	v.Check(len(r.Friends) == 2, "friends", "exactly 2 friends required")

	// Validate email
	for _, email := range r.Friends {
		validator.ValidateEmail(v, email)
	}
}

type GetFriendListRequest struct {
	Email string `json:"email"`
}
