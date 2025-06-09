package controller

import (
	"assignment/internal/domain/entities"
	"assignment/pkg/errors"
	stderrors "errors"
	"testing"
)

// mockUserRepo implements interfaces.UserRepositoryInterface for testing
type mockUserRepo struct {
	createFriendshipFunc    func(user1, user2 *entities.User) error
	getUserByEmailFunc      func(email string) (*entities.User, error)
	getFriendListFunc       func(user *entities.User) ([]*entities.User, error)
	getCommonFriendsFunc    func(user1, user2 *entities.User) ([]*entities.User, error)
	createSubscriptionFunc  func(requestor, target *entities.User) error
}

func (m *mockUserRepo) CreateFriendship(user1, user2 *entities.User) error {
	if m.createFriendshipFunc != nil {
		return m.createFriendshipFunc(user1, user2)
	}
	return nil
}

func (m *mockUserRepo) GetFriendList(user *entities.User) ([]*entities.User, error) {
	if m.getFriendListFunc != nil {
		return m.getFriendListFunc(user)
	}
	return []*entities.User{}, nil
}

func (m *mockUserRepo) GetCommonFriends(user1, user2 *entities.User) ([]*entities.User, error) {
	if m.getCommonFriendsFunc != nil {
		return m.getCommonFriendsFunc(user1, user2)
	}
	return []*entities.User{}, nil
}

func (m *mockUserRepo) CreateSubscription(requestor, target *entities.User) error {
	if m.createSubscriptionFunc != nil {
		return m.createSubscriptionFunc(requestor, target)
	}
	return nil
}

func (m *mockUserRepo) CreateBlock(requestor, target *entities.User) error {
	return nil
}

func (m *mockUserRepo) GetRecipients(sender *entities.User, mentionedUsers []*entities.User) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func (m *mockUserRepo) UserExists(email string) (*entities.User, error) {
	return &entities.User{ID: 1, Email: email}, nil
}

func (m *mockUserRepo) GetUserByEmail(email string) (*entities.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(email)
	}
	return &entities.User{ID: 1, Email: email}, nil
}

func TestCreateFriendships(t *testing.T) {
	tests := []struct {
		name         string
		user1Email   string
		user2Email   string
		mockFunc     func(user1, user2 *entities.User) error
		wantErr      bool
		wantErrType  errors.ErrorType
		wantErrMsg   string
	}{
		{
			name:       "successful friendship creation",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			mockFunc:   nil,
			wantErr:    false,
		},
		{
			name:       "same email friendship should fail",
			user1Email: "a@example.com",
			user2Email: "a@example.com",
			mockFunc:   nil,
			wantErr:    true,
			wantErrType: errors.ErrorTypeBusiness,
			wantErrMsg: "Cannot add yourself as a friend",
		},
		{
			name:       "repo error",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			mockFunc: func(user1, user2 *entities.User) error {
				return errors.New(errors.ErrorTypeDatabase, "database connection failed")
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepo{createFriendshipFunc: tt.mockFunc}
			controller := NewUserController(mockRepo)

			err := controller.CreateFriendship(tt.user1Email, tt.user2Email)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}

			var appErr *errors.AppError
			if !stderrors.As(err, &appErr) {
				t.Errorf("expected AppError, got %T", err)
				return
			}

			if appErr.Type != tt.wantErrType {
				t.Errorf("expected error type %s, got %s", tt.wantErrType, appErr.Type)
			}

			if appErr.Message != tt.wantErrMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.wantErrMsg, appErr.Message)
			}
		})
	}
}

func TestGetFriendList(t *testing.T) {
	tests := []struct {
		name              string
		email             string
		getUserByEmailFunc func(email string) (*entities.User, error)
		getFriendListFunc  func(user *entities.User) ([]*entities.User, error)
		wantErr           bool
		wantErrType       errors.ErrorType
		wantErrMsg        string
		expectedFriends   []*entities.User
	}{
		{
			name:  "successful friend list retrieval with friends",
			email: "andy@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				return &entities.User{ID: 1, Email: "andy@example.com"}, nil
			},
			getFriendListFunc: func(user *entities.User) ([]*entities.User, error) {
				return []*entities.User{
					{ID: 2, Email: "john@example.com"},
					{ID: 3, Email: "jane@example.com"},
				}, nil
			},
			wantErr: false,
			expectedFriends: []*entities.User{
				{ID: 2, Email: "john@example.com"},
				{ID: 3, Email: "jane@example.com"},
			},
		},
		{
			name:  "successful friend list retrieval with no friends",
			email: "andy@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				return &entities.User{ID: 1, Email: "andy@example.com"}, nil
			},
			getFriendListFunc: func(user *entities.User) ([]*entities.User, error) {
				return []*entities.User{}, nil
			},
			wantErr:         false,
			expectedFriends: []*entities.User{},
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			getFriendListFunc: nil,
			wantErr:          true,
			wantErrType:      errors.ErrorTypeNotFound,
			wantErrMsg:       "User with email 'nonexistent@example.com' not found",
		},
		{
			name:  "repository error when getting friends",
			email: "andy@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				return &entities.User{ID: 1, Email: "andy@example.com"}, nil
			},
			getFriendListFunc: func(user *entities.User) ([]*entities.User, error) {
				return nil, errors.New(errors.ErrorTypeDatabase, "database connection failed")
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepo{
				getUserByEmailFunc: tt.getUserByEmailFunc,
				getFriendListFunc:  tt.getFriendListFunc,
			}
			controller := NewUserController(mockRepo)

			friends, err := controller.GetFriendList(tt.email)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				if len(friends) != len(tt.expectedFriends) {
					t.Errorf("expected %d friends, got %d", len(tt.expectedFriends), len(friends))
					return
				}

				for i, expectedFriend := range tt.expectedFriends {
					if friends[i].ID != expectedFriend.ID || friends[i].Email != expectedFriend.Email {
						t.Errorf("expected friend %v, got %v", expectedFriend, friends[i])
					}
				}
				return
			}

			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}

			var appErr *errors.AppError
			if !stderrors.As(err, &appErr) {
				t.Errorf("expected AppError, got %T", err)
				return
			}

			if appErr.Type != tt.wantErrType {
				t.Errorf("expected error type %s, got %s", tt.wantErrType, appErr.Type)
			}

			if appErr.Message != tt.wantErrMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.wantErrMsg, appErr.Message)
			}
		})
	}
}

func TestGetCommonFriends(t *testing.T) {
	tests := []struct {
		name                  string
		email1                string
		email2                string
		getUserByEmailFunc    func(email string) (*entities.User, error)
		getCommonFriendsFunc  func(user1, user2 *entities.User) ([]*entities.User, error)
		wantErr               bool
		wantErrType           errors.ErrorType
		wantErrMsg            string
		expectedCommonFriends []*entities.User
	}{
		{
			name:   "successful common friends retrieval with common friends",
			email1: "andy@example.com",
			email2: "john@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "andy@example.com" {
					return &entities.User{ID: 1, Email: "andy@example.com"}, nil
				}
				if email == "john@example.com" {
					return &entities.User{ID: 2, Email: "john@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			getCommonFriendsFunc: func(user1, user2 *entities.User) ([]*entities.User, error) {
				return []*entities.User{
					{ID: 3, Email: "jane@example.com"},
					{ID: 4, Email: "bob@example.com"},
				}, nil
			},
			wantErr: false,
			expectedCommonFriends: []*entities.User{
				{ID: 3, Email: "jane@example.com"},
				{ID: 4, Email: "bob@example.com"},
			},
		},
		{
			name:   "successful common friends retrieval with no common friends",
			email1: "andy@example.com",
			email2: "john@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "andy@example.com" {
					return &entities.User{ID: 1, Email: "andy@example.com"}, nil
				}
				if email == "john@example.com" {
					return &entities.User{ID: 2, Email: "john@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			getCommonFriendsFunc: func(user1, user2 *entities.User) ([]*entities.User, error) {
				return []*entities.User{}, nil
			},
			wantErr:               false,
			expectedCommonFriends: []*entities.User{},
		},
		{
			name:                 "cannot get common friends with self",
			email1:               "andy@example.com",
			email2:               "andy@example.com",
			getUserByEmailFunc:   nil,
			getCommonFriendsFunc: nil,
			wantErr:              true,
			wantErrType:          errors.ErrorTypeBusiness,
			wantErrMsg:           "Cannot get common friends with yourself",
		},
		{
			name:   "first user not found",
			email1: "nonexistent@example.com",
			email2: "john@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "nonexistent@example.com" {
					return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
				}
				return &entities.User{ID: 2, Email: "john@example.com"}, nil
			},
			getCommonFriendsFunc: nil,
			wantErr:              true,
			wantErrType:          errors.ErrorTypeNotFound,
			wantErrMsg:           "User with email 'nonexistent@example.com' not found",
		},
		{
			name:   "second user not found",
			email1: "andy@example.com",
			email2: "nonexistent@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "andy@example.com" {
					return &entities.User{ID: 1, Email: "andy@example.com"}, nil
				}
				if email == "nonexistent@example.com" {
					return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			getCommonFriendsFunc: nil,
			wantErr:              true,
			wantErrType:          errors.ErrorTypeNotFound,
			wantErrMsg:           "User with email 'nonexistent@example.com' not found",
		},
		{
			name:   "repository error when getting common friends",
			email1: "andy@example.com",
			email2: "john@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "andy@example.com" {
					return &entities.User{ID: 1, Email: "andy@example.com"}, nil
				}
				if email == "john@example.com" {
					return &entities.User{ID: 2, Email: "john@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			getCommonFriendsFunc: func(user1, user2 *entities.User) ([]*entities.User, error) {
				return nil, errors.New(errors.ErrorTypeDatabase, "database connection failed")
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepo{
				getUserByEmailFunc:   tt.getUserByEmailFunc,
				getCommonFriendsFunc: tt.getCommonFriendsFunc,
			}
			controller := NewUserController(mockRepo)

			commonFriends, err := controller.GetCommonFriends(tt.email1, tt.email2)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				if len(commonFriends) != len(tt.expectedCommonFriends) {
					t.Errorf("expected %d common friends, got %d", len(tt.expectedCommonFriends), len(commonFriends))
					return
				}

				for i, expectedFriend := range tt.expectedCommonFriends {
					if commonFriends[i].ID != expectedFriend.ID || commonFriends[i].Email != expectedFriend.Email {
						t.Errorf("expected common friend %v, got %v", expectedFriend, commonFriends[i])
					}
				}
				return
			}

			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}

			var appErr *errors.AppError
			if !stderrors.As(err, &appErr) {
				t.Errorf("expected AppError, got %T", err)
				return
			}

			if appErr.Type != tt.wantErrType {
				t.Errorf("expected error type %s, got %s", tt.wantErrType, appErr.Type)
			}

			if appErr.Message != tt.wantErrMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.wantErrMsg, appErr.Message)
			}
		})
	}
}


func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name                   string
		requestorEmail         string
		targetEmail            string
		getUserByEmailFunc     func(email string) (*entities.User, error)
		createSubscriptionFunc func(requestor, target *entities.User) error
		wantErr                bool
		wantErrType            errors.ErrorType
		wantErrMsg             string
	}{
		{
			name:           "successful subscription creation",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "requestor@example.com" {
					return &entities.User{ID: 1, Email: "requestor@example.com"}, nil
				}
				if email == "target@example.com" {
					return &entities.User{ID: 2, Email: "target@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			createSubscriptionFunc: nil,
			wantErr:                false,
		},
		{
			name:           "requestor user not found",
			requestorEmail: "nonexistent@example.com",
			targetEmail:    "target@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "nonexistent@example.com" {
					return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
				}
				if email == "target@example.com" {
					return &entities.User{ID: 2, Email: "target@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			createSubscriptionFunc: nil,
			wantErr:                true,
			wantErrType:            errors.ErrorTypeNotFound,
			wantErrMsg:             "User with email 'nonexistent@example.com' not found",
		},
		{
			name:           "target user not found",
			requestorEmail: "requestor@example.com",
			targetEmail:    "nonexistent@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "requestor@example.com" {
					return &entities.User{ID: 1, Email: "requestor@example.com"}, nil
				}
				if email == "nonexistent@example.com" {
					return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			createSubscriptionFunc: nil,
			wantErr:                true,
			wantErrType:            errors.ErrorTypeNotFound,
			wantErrMsg:             "User with email 'nonexistent@example.com' not found",
		},
		{
			name:           "repository error when creating subscription",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "requestor@example.com" {
					return &entities.User{ID: 1, Email: "requestor@example.com"}, nil
				}
				if email == "target@example.com" {
					return &entities.User{ID: 2, Email: "target@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			createSubscriptionFunc: func(requestor, target *entities.User) error {
				return errors.New(errors.ErrorTypeDatabase, "database connection failed")
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
		{
			name:           "duplicate subscription",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			getUserByEmailFunc: func(email string) (*entities.User, error) {
				if email == "requestor@example.com" {
					return &entities.User{ID: 1, Email: "requestor@example.com"}, nil
				}
				if email == "target@example.com" {
					return &entities.User{ID: 2, Email: "target@example.com"}, nil
				}
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			createSubscriptionFunc: func(requestor, target *entities.User) error {
				return errors.New(errors.ErrorTypeBusiness, "Subscription already exists")
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeBusiness,
			wantErrMsg:  "Subscription already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepo{
				getUserByEmailFunc:     tt.getUserByEmailFunc,
				createSubscriptionFunc: tt.createSubscriptionFunc,
			}
			controller := NewUserController(mockRepo)

			err := controller.CreateSubscription(tt.requestorEmail, tt.targetEmail)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}

			var appErr *errors.AppError
			if !stderrors.As(err, &appErr) {
				t.Errorf("expected AppError, got %T", err)
				return
			}

			if appErr.Type != tt.wantErrType {
				t.Errorf("expected error type %s, got %s", tt.wantErrType, appErr.Type)
			}

			if appErr.Message != tt.wantErrMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.wantErrMsg, appErr.Message)
			}
		})
	}
}