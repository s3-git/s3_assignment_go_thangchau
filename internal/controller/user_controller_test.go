package controller

import (
	"assignment/internal/domain/entities"
	"assignment/mocks"
	"assignment/pkg/errors"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestCreateFriendships(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		user1Email  string
		user2Email  string
		setupMock   func(mockRepo *mocks.MockUserRepositoryInterface)
		wantErr     bool
		wantErrType errors.ErrorType
		wantErrMsg  string
	}{
		{
			name:       "successful friendship creation",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "a@example.com"}
				user2 := &entities.User{ID: 2, Email: "b@example.com"}
				mockRepo.EXPECT().GetUserByEmail("a@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("b@example.com").Return(user2, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, nil)
				mockRepo.EXPECT().CreateFriendship(user1, user2).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "same email friendship should fail",
			user1Email: "a@example.com",
			user2Email: "a@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				// No mock expectations needed as validation happens before repository calls
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeBusiness,
			wantErrMsg:  "Cannot add yourself as a friend",
		},
		{
			name:       "repo error",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "a@example.com"}
				user2 := &entities.User{ID: 2, Email: "b@example.com"}

				mockRepo.EXPECT().GetUserByEmail("a@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("b@example.com").Return(user2, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, nil)
				mockRepo.EXPECT().CreateFriendship(user1, user2).Return(errors.New(errors.ErrorTypeDatabase, "database connection failed"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
		{
			name:       "friendship blocked - user1 blocks user2",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "a@example.com"}
				user2 := &entities.User{ID: 2, Email: "b@example.com"}

				mockRepo.EXPECT().GetUserByEmail("a@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("b@example.com").Return(user2, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(true, nil)
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeForbidden,
			wantErrMsg:  "Cannot perform action on blocked user",
		},
		{
			name:       "block check error",
			user1Email: "a@example.com",
			user2Email: "b@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "a@example.com"}
				user2 := &entities.User{ID: 2, Email: "b@example.com"}

				mockRepo.EXPECT().GetUserByEmail("a@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("b@example.com").Return(user2, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, errors.New(errors.ErrorTypeDatabase, "Failed to check block existence"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "Failed to check block existence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		email           string
		setupMock       func(mockRepo *mocks.MockUserRepositoryInterface)
		wantErr         bool
		wantErrType     errors.ErrorType
		wantErrMsg      string
		expectedFriends []*entities.User
	}{
		{
			name:  "successful friend list retrieval with friends",
			email: "andy@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user := &entities.User{ID: 1, Email: "andy@example.com"}
				friends := []*entities.User{
					{ID: 2, Email: "john@example.com"},
					{ID: 3, Email: "jane@example.com"},
				}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user, nil)
				mockRepo.EXPECT().GetFriendList(user).Return(friends, nil)
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
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user := &entities.User{ID: 1, Email: "andy@example.com"}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user, nil)
				mockRepo.EXPECT().GetFriendList(user).Return([]*entities.User{}, nil)
			},
			wantErr:         false,
			expectedFriends: []*entities.User{},
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:  "repository error when getting friends",
			email: "andy@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user := &entities.User{ID: 1, Email: "andy@example.com"}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user, nil)
				mockRepo.EXPECT().GetFriendList(user).Return(nil, errors.New(errors.ErrorTypeDatabase, "database connection failed"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name                  string
		email1                string
		email2                string
		setupMock             func(mockRepo *mocks.MockUserRepositoryInterface)
		wantErr               bool
		wantErrType           errors.ErrorType
		wantErrMsg            string
		expectedCommonFriends []*entities.User
	}{
		{
			name:   "successful common friends retrieval with common friends",
			email1: "andy@example.com",
			email2: "john@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "andy@example.com"}
				user2 := &entities.User{ID: 2, Email: "john@example.com"}
				commonFriends := []*entities.User{
					{ID: 3, Email: "jane@example.com"},
					{ID: 4, Email: "bob@example.com"},
				}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("john@example.com").Return(user2, nil)
				mockRepo.EXPECT().GetCommonFriends(user1, user2).Return(commonFriends, nil)
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
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "andy@example.com"}
				user2 := &entities.User{ID: 2, Email: "john@example.com"}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("john@example.com").Return(user2, nil)
				mockRepo.EXPECT().GetCommonFriends(user1, user2).Return([]*entities.User{}, nil)
			},
			wantErr:               false,
			expectedCommonFriends: []*entities.User{},
		},
		{
			name:   "cannot get common friends with self",
			email1: "andy@example.com",
			email2: "andy@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				// No mock expectations needed as validation happens before repository calls
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeBusiness,
			wantErrMsg:  "Cannot get common friends with yourself",
		},
		{
			name:   "first user not found",
			email1: "nonexistent@example.com",
			email2: "john@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:   "second user not found",
			email1: "andy@example.com",
			email2: "nonexistent@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "andy@example.com"}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:   "repository error when getting common friends",
			email1: "andy@example.com",
			email2: "john@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				user1 := &entities.User{ID: 1, Email: "andy@example.com"}
				user2 := &entities.User{ID: 2, Email: "john@example.com"}
				mockRepo.EXPECT().GetUserByEmail("andy@example.com").Return(user1, nil)
				mockRepo.EXPECT().GetUserByEmail("john@example.com").Return(user2, nil)
				mockRepo.EXPECT().GetCommonFriends(user1, user2).Return(nil, errors.New(errors.ErrorTypeDatabase, "database connection failed"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		requestorEmail string
		targetEmail    string
		setupMock      func(mockRepo *mocks.MockUserRepositoryInterface)
		wantErr        bool
		wantErrType    errors.ErrorType
		wantErrMsg     string
	}{
		{
			name:           "successful subscription creation",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, nil)
				mockRepo.EXPECT().CreateSubscription(requestor, target).Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "requestor user not found",
			requestorEmail: "nonexistent@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:           "target user not found",
			requestorEmail: "requestor@example.com",
			targetEmail:    "nonexistent@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:           "repository error when creating subscription",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, nil)
				mockRepo.EXPECT().CreateSubscription(requestor, target).Return(errors.New(errors.ErrorTypeDatabase, "database connection failed"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "database connection failed",
		},
		{
			name:           "duplicate subscription",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, nil)
				mockRepo.EXPECT().CreateSubscription(requestor, target).Return(errors.New(errors.ErrorTypeBusiness, "Subscription already exists"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeBusiness,
			wantErrMsg:  "Subscription already exists",
		},
		{
			name:           "subscription blocked - requestor blocks target",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(true, nil)
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeForbidden,
			wantErrMsg:  "Cannot perform action on blocked user",
		},
		{
			name:           "subscription blocked - target blocks requestor",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(true, nil)
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeForbidden,
			wantErrMsg:  "Cannot perform action on blocked user",
		},
		{
			name:           "block check error during subscription",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, errors.New(errors.ErrorTypeDatabase, "Failed to check block existence"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "Failed to check block existence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

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
func TestCreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		requestorEmail string
		targetEmail    string
		setupMock      func(mockRepo *mocks.MockUserRepositoryInterface)
		wantErr        bool
		wantErrType    errors.ErrorType
		wantErrMsg     string
	}{
		{
			name:           "successful block creation",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CreateBlockTx(requestor, target).Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "requestor user not found",
			requestorEmail: "nonexistent@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:           "target user not found",
			requestorEmail: "requestor@example.com",
			targetEmail:    "nonexistent@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeNotFound,
			wantErrMsg:  "User with email 'nonexistent@example.com' not found",
		},
		{
			name:           "transaction failure during block creation",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CreateBlockTx(requestor, target).Return(errors.New(errors.ErrorTypeDatabase, "Failed to delete friendship"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "Failed to delete friendship",
		},
		{
			name:           "transaction failure during subscription deletion",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CreateBlockTx(requestor, target).Return(errors.New(errors.ErrorTypeDatabase, "Failed to delete requestor subscription"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "Failed to delete requestor subscription",
		},
		{
			name:           "block creation database error",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CreateBlockTx(requestor, target).Return(errors.New(errors.ErrorTypeDatabase, "Block already exists"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "Block already exists",
		},
		{
			name:           "transaction commit failure",
			requestorEmail: "requestor@example.com",
			targetEmail:    "target@example.com",
			setupMock: func(mockRepo *mocks.MockUserRepositoryInterface) {
				requestor := &entities.User{ID: 1, Email: "requestor@example.com"}
				target := &entities.User{ID: 2, Email: "target@example.com"}
				mockRepo.EXPECT().GetUserByEmail("requestor@example.com").Return(requestor, nil)
				mockRepo.EXPECT().GetUserByEmail("target@example.com").Return(target, nil)
				mockRepo.EXPECT().CreateBlockTx(requestor, target).Return(errors.New(errors.ErrorTypeDatabase, "Failed to commit transaction"))
			},
			wantErr:     true,
			wantErrType: errors.ErrorTypeDatabase,
			wantErrMsg:  "Failed to commit transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

			controller := NewUserController(mockRepo)
			err := controller.CreateBlock(tt.requestorEmail, tt.targetEmail)

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
