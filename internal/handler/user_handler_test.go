package handler

import (
	"assignment/internal/domain/interfaces"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUserController struct {
	createFriendshipsFunc func(string, string) error
}

// Implement all interface methods
func (m *mockUserController) CreateFriendships(u1, u2 string) error {
	if m.createFriendshipsFunc != nil {
		return m.createFriendshipsFunc(u1, u2)
	}
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
			expectedBody:   `{"success":true}`,
		},
		{
			name: "same email",
			body: `{"friends":["andy@example.com", "andy@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.New("cannot befriend self")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"cannot befriend self"}`,
		},
		{
			name: "missing email",
			body: `{"friends":["andy@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.New("invalid body")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid body"}`,
		},
		{
			name: "internal server error",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			mockFunc: func(u1, u2 string) error {
				return errors.New("internal server error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name:           "invalid json",
			body:           `{"friends": [}`,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mUserController interfaces.UserControllerInterface
			if tt.mockFunc != nil {
				mUserController = &mockUserController{createFriendshipsFunc: tt.mockFunc}
			} else {
				mUserController = &mockUserController{}
			}
			handler := NewUserHandler(mUserController)

			router := gin.Default()
			router.POST("/friendships", handler.CreateFriendships)

			req, _ := http.NewRequest(http.MethodPost, "/friendships", bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
