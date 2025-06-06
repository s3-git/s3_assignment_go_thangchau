package controller

import (
	"assignment/internal/domain/entities"
	"assignment/pkg/errors"
	stderrors "errors"
	"testing"
)

// mockUserRepo implements interfaces.UserRepositoryInterface for testing
type mockUserRepo struct {
	createFriendshipFunc func(user1, user2 *entities.User) error
	getUserByEmailFunc   func(email string) (*entities.User, error)
}

func (m *mockUserRepo) CreateFriendship(user1, user2 *entities.User) error {
	if m.createFriendshipFunc != nil {
		return m.createFriendshipFunc(user1, user2)
	}
	return nil
}

func (m *mockUserRepo) GetFriendList(user *entities.User) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func (m *mockUserRepo) GetCommonFriends(user1, user2 *entities.User) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func (m *mockUserRepo) CreateSubscription(requestor, target *entities.User) error {
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
