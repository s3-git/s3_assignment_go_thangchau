package controller

import (
	"errors"
	"testing"
)

// mockUserRepo implements interfaces.UserRepositoryInterface for testing
type mockUserRepo struct {
	createFriendshipFunc func(user1Email, user2Email string) error
}

func (m *mockUserRepo) CreateFriendship(u1, u2 string) error {
	if m.createFriendshipFunc != nil {
		return m.createFriendshipFunc(u1, u2)
	}
	return nil
}

func TestCreateFriendships(t *testing.T) {
	tests := []struct {
		name       string
		user1Email string
		user2Email string
		mockFunc   func(user1Email, user2Email string) error
		wantErr    string
	}{
		{
			name:       "same email",
			user1Email: "a@mail.com",
			user2Email: "a@mail.com",
			mockFunc:   nil,
			wantErr:    "cannot befriend self",
		},
		{
			name:       "repo error",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			mockFunc: func(user1Email, user2Email string) error {
				return errors.New("repo error")
			},
			wantErr: "repo error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepo{createFriendshipFunc: tt.mockFunc}
			controller := NewUserController(mockRepo)

			err := controller.CreateFriendship(tt.user1Email, tt.user2Email)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("expected error '%s', got %v", tt.wantErr, err)
				}
			}
		})
	}
}
