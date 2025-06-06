package handler

import (
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
	createFriendshipsFunc func(string, string) error
}

func (m *mockUserController) CreateFriendship(u1, u2 string) error {
	if m.createFriendshipsFunc != nil {
		return m.createFriendshipsFunc(u1, u2)
	}
	return nil
}

func (m *mockUserController) GetFriendList(email string) error {
	return nil
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
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"success":true,"message":"Friendship created successfully"}`,
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
			name: "business logic error - already friends",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.ErrAlreadyFriends
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"success":false,"error":{"type":"CONFLICT","message":"Users are already friends"}}`,
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