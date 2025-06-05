package entities

import (
	"assignment/pkg/validator"
)

type Friend struct {
	ID      string
	User1ID int
	User2ID int
}

//TODO: movel somewhere else
type CreateFriendshipRequest struct {
	Friends []string `json:"friends"`
}

func ValidateCreateFriendshipRequest(v *validator.Validator, r *CreateFriendshipRequest) {
	v.Check(len(r.Friends) == 2, "friends", "exactly 2 emails required")

	for _, email := range r.Friends {
		validator.ValidateEmail(v, email)
	}
}

//TODO: movel somewhere else
type GetFriendListRequest struct {
	Email string `json:"email"`
}

func ValidateGetFriendlistRequest(v *validator.Validator, r *GetFriendListRequest) {
	validator.ValidateEmail(v, r.Email)
}
