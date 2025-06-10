package handler

import (
	"assignment/internal/domain/entities"
	"assignment/mocks"
	"assignment/pkg/errors"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateFriendships(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		setupMock      func(mockController *mocks.MockUserControllerInterface)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateFriendship("andy@example.com", "john@example.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true}`,
		},
		{
			name: "missing email validation",
			body: `{"friends":["andy@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails count: exactly 2 emails required"}}`,
		},
		{
			name: "user not found error",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateFriendship("andy@example.com", "john@example.com").Return(errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "andy@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User with email 'andy@example.com' not found"}}`,
		},
		{
			name: "cannot friend self",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateFriendship("andy@example.com", "john@example.com").Return(errors.New(errors.ErrorTypeBusiness, "Cannot add yourself as a friend"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"BUSINESS_ERROR","message":"Cannot add yourself as a friend"}}`,
		},
		{
			name: "invalid json",
			body: `{"friends": [}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := mocks.NewMockUserControllerInterface(ctrl)
			tt.setupMock(mockController)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		setupMock      func(mockController *mocks.MockUserControllerInterface)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with friends",
			body: `{"email":"andy@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().GetFriendList("andy@example.com").Return([]*entities.User{
					{ID: 1, Email: "john@example.com"},
					{ID: 2, Email: "jane@example.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":["john@example.com","jane@example.com"],"count":2}`,
		},
		{
			name: "success with no friends",
			body: `{"email":"andy@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().GetFriendList("andy@example.com").Return([]*entities.User{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":[],"count":0}`,
		},
		{
			name: "user not found error",
			body: `{"email":"nonexistent@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().GetFriendList("nonexistent@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User with email 'nonexistent@example.com' not found"}}`,
		},
		{
			name: "invalid email format",
			body: `{"email":"invalid-email"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name: "empty email",
			body: `{"email":""}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be provided"}}`,
		},
		{
			name: "invalid json",
			body: `{"email": }`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := mocks.NewMockUserControllerInterface(ctrl)
			tt.setupMock(mockController)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		setupMock      func(mockController *mocks.MockUserControllerInterface)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with common friends",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().GetCommonFriends("andy@example.com", "john@example.com").Return([]*entities.User{
					{ID: 3, Email: "common@example.com"},
					{ID: 4, Email: "mutual@example.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":["common@example.com","mutual@example.com"],"count":2}`,
		},
		{
			name: "success with no common friends",
			body: `{"friends":["andy@example.com", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().GetCommonFriends("andy@example.com", "john@example.com").Return([]*entities.User{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"friends":[],"count":0}`,
		},
		{
			name: "user not found error",
			body: `{"friends":["nonexistent@example.com", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().GetCommonFriends("nonexistent@example.com", "john@example.com").Return(nil, errors.Newf(errors.ErrorTypeNotFound, "User with email '%s' not found", "nonexistent@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User with email 'nonexistent@example.com' not found"}}`,
		},
		{
			name: "missing email validation",
			body: `{"friends":["andy@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails count: exactly 2 emails required"}}`,
		},
		{
			name: "invalid email format",
			body: `{"friends":["invalid-email", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name: "empty email",
			body: `{"friends":["", "john@example.com"]}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: email cannot be empty"}}`,
		},
		{
			name: "invalid json",
			body: `{"friends": [}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := mocks.NewMockUserControllerInterface(ctrl)
			tt.setupMock(mockController)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		setupMock      func(mockController *mocks.MockUserControllerInterface)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: `{"requestor":"andy@example.com","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateSubscription("andy@example.com", "john@example.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true}`,
		},
		{
			name: "user not found error - requestor",
			body: `{"requestor":"nonexistent@example.com","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateSubscription("nonexistent@example.com", "john@example.com").Return(errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", "nonexistent@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User not found: nonexistent@example.com"}}`,
		},
		{
			name: "user not found error - target", 
			body: `{"requestor":"andy@example.com","target":"nonexistent@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateSubscription("andy@example.com", "nonexistent@example.com").Return(errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", "nonexistent@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User not found: nonexistent@example.com"}}`,
		},
		{
			name: "database error",
			body: `{"requestor":"andy@example.com","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateSubscription("andy@example.com", "john@example.com").Return(errors.New(errors.ErrorTypeDatabase, "Database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":{"type":"DATABASE_ERROR","message":"Database connection failed"}}`,
		},
		{
			name:           "empty requestor email",
			body:           `{"requestor":"","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"requestor: requestor email cannot be empty; email: must be provided"}}`,
		},
		{
			name:           "empty target email",
			body:           `{"requestor":"andy@example.com","target":""}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"target: target email cannot be empty; email: must be provided"}}`,
		},
		{
			name:           "invalid requestor email format",
			body:           `{"requestor":"invalid-email","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name:           "invalid target email format",
			body:           `{"requestor":"andy@example.com","target":"invalid-email"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"email: must be valid email address"}}`,
		},
		{
			name:           "same requestor and target",
			body:           `{"requestor":"andy@example.com","target":"andy@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails: requestor and target cannot be the same"}}`,
		},
		{
			name:           "invalid json",
			body:           `{"requestor": }`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := mocks.NewMockUserControllerInterface(ctrl)
			tt.setupMock(mockController)

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
func TestCreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		setupMock      func(mockController *mocks.MockUserControllerInterface)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: `{"requestor":"andy@example.com","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateBlock("andy@example.com", "john@example.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true}`,
		},
		{
			name: "user not found error - requestor",
			body: `{"requestor":"nonexistent@example.com","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateBlock("nonexistent@example.com", "john@example.com").Return(errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", "nonexistent@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User not found: nonexistent@example.com"}}`,
		},
		{
			name: "user not found error - target",
			body: `{"requestor":"andy@example.com","target":"nonexistent@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateBlock("andy@example.com", "nonexistent@example.com").Return(errors.Newf(errors.ErrorTypeNotFound, "User not found: %s", "nonexistent@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"success":false,"error":{"type":"NOT_FOUND","message":"User not found: nonexistent@example.com"}}`,
		},
		{
			name: "database error",
			body: `{"requestor":"andy@example.com","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				mockController.EXPECT().CreateBlock("andy@example.com", "john@example.com").Return(errors.New(errors.ErrorTypeDatabase, "Database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":{"type":"DATABASE_ERROR","message":"Database connection failed"}}`,
		},
		{
			name:           "empty requestor email",
			body:           `{"requestor":"","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"Key: 'CreateBlockRequest.Requestor' Error:Field validation for 'Requestor' failed on the 'required' tag"}}`,
		},
		{
			name:           "empty target email",
			body:           `{"requestor":"andy@example.com","target":""}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"Key: 'CreateBlockRequest.Target' Error:Field validation for 'Target' failed on the 'required' tag"}}`,
		},
		{
			name:           "invalid requestor email format",
			body:           `{"requestor":"invalid-email","target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"Key: 'CreateBlockRequest.Requestor' Error:Field validation for 'Requestor' failed on the 'email' tag"}}`,
		},
		{
			name:           "invalid target email format",
			body:           `{"requestor":"andy@example.com","target":"invalid-email"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"Key: 'CreateBlockRequest.Target' Error:Field validation for 'Target' failed on the 'email' tag"}}`,
		},
		{
			name:           "same requestor and target",
			body:           `{"requestor":"andy@example.com","target":"andy@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Validation failed","details":"emails: cannot block yourself"}}`,
		},
		{
			name:           "invalid json",
			body:           `{"requestor": }`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"invalid character '}' looking for beginning of value"}}`,
		},
		{
			name:           "missing requestor field",
			body:           `{"target":"john@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"Key: 'CreateBlockRequest.Requestor' Error:Field validation for 'Requestor' failed on the 'required' tag"}}`,
		},
		{
			name:           "missing target field",
			body:           `{"requestor":"andy@example.com"}`,
			setupMock: func(mockController *mocks.MockUserControllerInterface) {
				// No mock expectations needed as validation happens before controller calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":{"type":"VALIDATION_ERROR","message":"Invalid request format","details":"Key: 'CreateBlockRequest.Target' Error:Field validation for 'Target' failed on the 'required' tag"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := mocks.NewMockUserControllerInterface(ctrl)
			tt.setupMock(mockController)

			handler := NewUserHandler(mockController)

			router := gin.New()
			router.POST("/blocks", handler.CreateBlock)

			req, err := http.NewRequest(http.MethodPost, "/blocks", bytes.NewBuffer([]byte(tt.body)))
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