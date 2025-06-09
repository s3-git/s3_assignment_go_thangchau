package handler

import "assignment/pkg/validator"

type CreateFriendshipRequest struct {
	Friends []string `json:"friends"`
}

func ValidateCreateFriendshipRequest(v *validator.Validator, r *CreateFriendshipRequest) {
	v.Check(len(r.Friends) == 2, "emails count", "exactly 2 emails required")

	for _, email := range r.Friends {
		v.Check(len(email) > 0, "email", "email cannot be empty")
		validator.ValidateEmail(v, email)
	}
}

type GetFriendListRequest struct {
	Email string `json:"email"`
}

func ValidateGetFriendListRequest(v *validator.Validator, r *GetFriendListRequest) {
	validator.ValidateEmail(v, r.Email)
}

type GetCommonFriendsRequest struct {
	Friends []string `json:"friends"`
}

func ValidateGetCommonFriendsRequest(v *validator.Validator, r *GetCommonFriendsRequest) {
	v.Check(len(r.Friends) == 2, "emails count", "exactly 2 emails required")

	for _, email := range r.Friends {
		v.Check(len(email) > 0, "email", "email cannot be empty")
		validator.ValidateEmail(v, email)
	}
}

type SubscriptionRequest struct {
	Requestor string `json:"requestor"`
	Target    string `json:"target"`
}

func ValidateSubscriptionRequest(v *validator.Validator, r *SubscriptionRequest) {
	v.Check(len(r.Requestor) > 0, "requestor", "requestor email cannot be empty")
	v.Check(len(r.Target) > 0, "target", "target email cannot be empty")
	validator.ValidateEmail(v, r.Requestor)
	validator.ValidateEmail(v, r.Target)
	v.Check(r.Requestor != r.Target, "emails", "requestor and target cannot be the same")
}

type CreateBlockRequest struct {
	Requestor string `json:"requestor" binding:"required,email"`
	Target    string `json:"target" binding:"required,email"`
}

func ValidateCreateBlockRequest(v *validator.Validator, r *CreateBlockRequest) {
	v.Check(len(r.Requestor) > 0, "requestor", "requestor email cannot be empty")
	v.Check(len(r.Target) > 0, "target", "target email cannot be empty")
	validator.ValidateEmail(v, r.Requestor)
	validator.ValidateEmail(v, r.Target)
	v.Check(r.Requestor != r.Target, "emails", "cannot block yourself")
}

type GetRecipientsRequest struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

func ValidateGetRecipientsRequest(v *validator.Validator, r *GetRecipientsRequest) {
	v.Check(len(r.Sender) > 0, "sender", "sender email cannot be empty")
	v.Check(len(r.Text) > 0, "text", "text cannot be empty")
	validator.ValidateEmail(v, r.Sender)
}

type FriendListResponse struct {
	Success bool     `json:"success"`
	Friends []string `json:"friends"`
	Count   int      `json:"count"`
}

type CommonFriendsResponse struct {
	Success bool     `json:"success"`
	Friends []string `json:"friends"`
	Count   int      `json:"count"`
}

type RecipientsResponse struct {
	Success    bool     `json:"success"`
	Recipients []string `json:"recipients"`
}