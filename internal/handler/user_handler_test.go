package handler

import (
	"assignment/internal/domain/entities"
	"assignment/internal/domain/interfaces"
	"assignment/pkg/errors"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUserController struct {
	createFriendshipsFunc  func(string, string) error
	getFriendListFunc      func(string) ([]*entities.User, error)
	getCommonFriendsFunc   func(string, string) ([]*entities.User, error)
	createSubscriptionFunc func(string, string) error
}

func (m *mockUserController) CreateFriendship(u1, u2 string) error {
	if m.createFriendshipsFunc != nil {
		return m.createFriendshipsFunc(u1, u2)
	}
	return nil
}

func (m *mockUserController) GetFriendList(email string) ([]*entities.User, error) {
	if m.getFriendListFunc != nil {
		return m.getFriendListFunc(email)
	}
	return []*entities.User{}, nil
}

func (m *mockUserController) GetCommonFriends(email1, email2 string) ([]*entities.User, error) {
	if m.getCommonFriendsFunc != nil {
		return m.getCommonFriendsFunc(email1, email2)
	}
	return []*entities.User{}, nil
}

func (m *mockUserController) CreateSubscription(requestorEmail, targetEmail string) error {
	if m.createSubscriptionFunc != nil {
		return m.createSubscriptionFunc(requestorEmail, targetEmail)
	}
	return nil
}

func (m *mockUserController) CreateBlock(requestorEmail, targetEmail string) error {
	return nil
}

func (m *mockUserController) GetRecipients(senderEmail, text string) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func TestCreateFriendships(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockFunc       func(string, string) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true}`,
		},
		{
			name: "missing email validation",
			body: `{"friends":["andy@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.New(errors.ErrorTypeValidation, "invalid body")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails count: exactly 2 emails required"}}`,
		},
		{
			name: "user not found error",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "andy@example.com")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User with email 'andy@example.com' not found"}}`,
		},
		{
			name: "cannot friend self",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.ErrCannotFriendSelf
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"BUSINESS_ERROR","message":"Cannot add yourself as a friend"}}`,
		},
		{
			name:           "invalid json",
			body:           `{"friends": [}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockController interfaces.UserControllerInterface
			if tt.mockFunc != nil {
				mockController = &mockUserController{createFriendshipsFunc: tt.mockFunc}
			} else {
				mockController = &mockUserController{}
			}

			handler := NewUserHandler(mockController)

			router := gin.New()
			router.POST("/friends", handler.CreateFriendships)

			req, err := http.NewRequest(http.MethodPost, "/friends", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetFriendList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockFunc       func(string) ([]*entities.User, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with friends",
			body: `{"email":"andy@example.com"}`,
			mockFunc: func(email string) ([]*entities.User, error) {
				return []*entities.User{
					{ID: 1, Email: "john@example.com"},
					{ID: 2, Email: "jane@example.com"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":["john@example.com","jane@example.com"],"count":2}`,
		},
		{
			name: "success with no friends",
			body: `{"email":"andy@example.com"}`,
			mockFunc: func(email string) ([]*entities.User, error) {
				return []*entities.User{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":[],"count":0}`,
		},
		{
			name: "user not found error",
			body: `{"email":"nonexistent@example.com"}`,
			mockFunc: func(email string) ([]*entities.User, error) {
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User with email 'nonexistent@example.com' not found"}}`,
		},
		{
			name:           "invalid email format",
			body:           `{"email":"invalid-email"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name:           "empty email",
			body:           `{"email":""}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be provided"}}`,
		},
		{
			name:           "invalid json",
			body:           `{"email": }`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockController interfaces.UserControllerInterface
			if tt.mockFunc != nil {
				mockController = &mockUserController{getFriendListFunc: tt.mockFunc}
			} else {
				mockController = &mockUserController{}
			}

			handler := NewUserHandler(mockController)

			router := gin.New()
			router.POST("/friends/list", handler.GetFriendList)

			req, err := http.NewRequest(http.MethodPost, "/friends/list", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetCommonFriends(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockFunc       func(string, string) ([]*entities.User, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with common friends",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(email1, email2 string) ([]*entities.User, error) {
				return []*entities.User{
					{ID: 3, Email: "common@example.com"},
					{ID: 4, Email: "mutual@example.com"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":["common@example.com","mutual@example.com"],"count":2}`,
		},
		{
			name: "success with no common friends",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(email1, email2 string) ([]*entities.User, error) {
				return []*entities.User{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":[],"count":0}`,
		},
		{
			name: "user not found error",
			body: `{"friends":["nonexistent@example.com", "john@example.com"]}`,
			mockFunc: func(email1, email2 string) ([]*entities.User, error) {
				return nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", email1)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User with email 'nonexistent@example.com' not found"}}`,
		},
		{
			name:           "missing email validation",
			body:           `{"friends":["andy@example.com"]}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails count: exactly 2 emails required"}}`,
		},
		{
			name:           "invalid email format",
			body:           `{"friends":["invalid-email", "john@example.com"]}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name:           "empty email",
			body:           `{"friends":["", "john@example.com"]}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: email cannot be empty"}}`,
		},
		{
			name:           "invalid json",
			body:           `{"friends": [}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockController interfaces.UserControllerInterface
			if tt.mockFunc != nil {
				mockController = &mockUserController{getCommonFriendsFunc: tt.mockFunc}
			} else {
				mockController = &mockUserController{}
			}

			handler := NewUserHandler(mockController)

			router := gin.New()
			router.POST("/friends/common", handler.GetCommonFriends)

			req, err := http.NewRequest(http.MethodPost, "/friends/common", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestCreateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockFunc       func(string, string) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: `{"requestor":"andy@example.com","target":"john@example.com"}`,
			mockFunc: func(requestor, target string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true}`,
		},
		{
			name: "user not found error - requestor",
			body: `{"requestor":"nonexistent@example.com","target":"john@example.com"}`,
			mockFunc: func(requestor, target string) error {
				return errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", requestor)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User not found: nonexistent@example.com"}}`,
		},
		{
			name: "user not found error - target", 
			body: `{"requestor":"andy@example.com","target":"nonexistent@example.com"}`,
			mockFunc: func(requestor, target string) error {
				return errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", target)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User not found: nonexistent@example.com"}}`,
		},
		{
			name: "database error",
			body: `{"requestor":"andy@example.com","target":"john@example.com"}`,
			mockFunc: func(requestor, target string) error {
				return errors.New(errors.ErrorTypeDatabase, "Database connection failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":{"type":"DATABASE_ERROR","message":"Database connection failed"}}`,
		},
		{
			name:           "empty requestor email",
			body:           `{"requestor":"","target":"john@example.com"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"requestor: requestor email cannot be empty; email: must be provided"}}`,
		},
		{
			name:           "empty target email",
			body:           `{"requestor":"andy@example.com","target":""}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"target: target email cannot be empty; email: must be provided"}}`,
		},
		{
			name:           "invalid requestor email format",
			body:           `{"requestor":"invalid-email","target":"john@example.com"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name:           "invalid target email format",
			body:           `{"requestor":"andy@example.com","target":"invalid-email"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name:           "same requestor and target",
			body:           `{"requestor":"andy@example.com","target":"andy@example.com"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails: requestor and target cannot be the same"}}`,
		},
		{
			name:           "invalid json",
			body:           `{"requestor": }`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
		{
			name:           "missing requestor field",
			body:           `{"target":"john@example.com"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"requestor: requestor email cannot be empty; email: must be provided"}}`,
		},
		{
			name:           "missing target field",
			body:           `{"requestor":"andy@example.com"}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"target: target email cannot be empty; email: must be provided"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockController interfaces.UserControllerInterface
			if tt.mockFunc != nil {
				mockController = &mockUserController{createSubscriptionFunc: tt.mockFunc}
			} else {
				mockController = &mockUserController{}
			}

			handler := NewUserHandler(mockController)

			router := gin.New()
			router.POST("/subscriptions", handler.CreateSubscription)

			req, err := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}