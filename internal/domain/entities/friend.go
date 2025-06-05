package entities

import (
	"errors"
	"fmt"
	"regexp"
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
		if !isValidEmail(email) {
			return fmt.Errorf("invalid email format: %s", email)
		}
	}

	return nil
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched && len(email) <= 254
}
