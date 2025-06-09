package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Business logic errors
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	ErrorTypeBusiness       ErrorType = "BUSINESS_ERROR"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"
	ErrorTypeConflict       ErrorType = "CONFLICT"
	ErrorTypeUnauthorized   ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden      ErrorType = "FORBIDDEN"
	
	// System errors
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeDatabase       ErrorType = "DATABASE_ERROR"
	ErrorTypeExternal       ErrorType = "EXTERNAL_ERROR"
)

// AppError represents a standardized application error
type AppError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Internal   error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the internal error for error wrapping
func (e *AppError) Unwrap() error {
	return e.Internal
}

// GetStatusCode returns the HTTP status code for the error
func (e *AppError) GetStatusCode() int {
	if e.StatusCode != 0 {
		return e.StatusCode
	}
	
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeBusiness:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeInternal, ErrorTypeDatabase, ErrorTypeExternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new AppError
func New(errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
	}
}

// Newf creates a new AppError with formatted message
func Newf(errorType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:     errorType,
		Message:  message,
		Internal: err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:     errorType,
		Message:  fmt.Sprintf(format, args...),
		Internal: err,
	}
}

// WithStatusCode sets a custom status code
func (e *AppError) WithStatusCode(code int) *AppError {
	e.StatusCode = code
	return e
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// FromError converts a standard error to AppError
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, return as is
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	
	// Handle SQL specific errors
	if sqlErr := handleSQLError(err); sqlErr != nil {
		return sqlErr
	}
	
	// Default to internal error
	return &AppError{
		Type:     ErrorTypeInternal,
		Message:  "Internal server error",
		Internal: err,
	}
}

// handleSQLError handles database-specific errors
func handleSQLError(err error) *AppError {
	if err == nil {
		return nil
	}
	
	// Handle sql.ErrNoRows
	if errors.Is(err, sql.ErrNoRows) {
		return &AppError{
			Type:     ErrorTypeNotFound,
			Message:  "Resource not found",
			Internal: err,
		}
	}
	
	// Handle PostgreSQL specific errors
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return handlePostgreSQLError(pqErr)
	}
	
	// Handle other database errors
	if strings.Contains(err.Error(), "duplicate") || 
	   strings.Contains(err.Error(), "constraint") ||
	   strings.Contains(err.Error(), "unique") {
		return &AppError{
			Type:     ErrorTypeConflict,
			Message:  "Resource already exists or constraint violation",
			Internal: err,
		}
	}
	
	return nil
}

// handlePostgreSQLError handles PostgreSQL specific error codes
func handlePostgreSQLError(pqErr *pq.Error) *AppError {
	switch pqErr.Code {
	case "23505": // unique_violation
		return &AppError{
			Type:     ErrorTypeConflict,
			Message:  "Resource already exists",
			Details:  getConstraintMessage(pqErr.Constraint),
			Internal: pqErr,
		}
	case "23503": // foreign_key_violation
		return &AppError{
			Type:     ErrorTypeValidation,
			Message:  "Referenced resource does not exist",
			Details:  pqErr.Detail,
			Internal: pqErr,
		}
	case "23502": // not_null_violation
		return &AppError{
			Type:     ErrorTypeValidation,
			Message:  "Required field is missing",
			Details:  fmt.Sprintf("Field '%s' cannot be null", pqErr.Column),
			Internal: pqErr,
		}
	case "23514": // check_violation
		return &AppError{
			Type:     ErrorTypeValidation,
			Message:  "Invalid data provided",
			Details:  pqErr.Detail,
			Internal: pqErr,
		}
	case "42P01": // undefined_table
		return &AppError{
			Type:     ErrorTypeInternal,
			Message:  "Database configuration error",
			Internal: pqErr,
		}
	default:
		return &AppError{
			Type:     ErrorTypeDatabase,
			Message:  "Database operation failed",
			Details:  pqErr.Message,
			Internal: pqErr,
		}
	}
}

// getConstraintMessage returns user-friendly constraint messages
func getConstraintMessage(constraint string) string {
	switch {
	case strings.Contains(constraint, "email"):
		return "Email address already exists"
	case strings.Contains(constraint, "friend"):
		return "Friendship already exists"
	case strings.Contains(constraint, "block"):
		return "User is already blocked"
	case strings.Contains(constraint, "subscription"):
		return "Subscription already exists"
	default:
		return "Duplicate entry"
	}
}

// Predefined common errors
var (
	ErrNotFound        = New(ErrorTypeNotFound, "Resource not found")
	ErrUserNotFound    = New(ErrorTypeNotFound, "User not found")
	ErrInvalidInput    = New(ErrorTypeValidation, "Invalid input provided")
	ErrInternalServer  = New(ErrorTypeInternal, "Internal server error")
	ErrUnauthorized    = New(ErrorTypeUnauthorized, "Unauthorized access")
	ErrForbidden       = New(ErrorTypeForbidden, "Access forbidden")
	ErrConflict        = New(ErrorTypeConflict, "Resource conflict")
)

// Business logic errors
var (
	ErrCannotFriendSelf              = New(ErrorTypeBusiness, "Cannot add yourself as a friend")
	ErrCannotBlockSelf               = New(ErrorTypeBusiness, "Cannot block yourself")
	ErrCannotSubscribeSelf           = New(ErrorTypeBusiness, "Cannot subscribe to yourself")
	ErrCannotGetCommonFriendsWithSelf = New(ErrorTypeBusiness, "Cannot get common friends with yourself")
	ErrAlreadyFriends                = New(ErrorTypeConflict, "Users are already friends")
	ErrAlreadyBlocked                = New(ErrorTypeConflict, "User is already blocked")
	ErrAlreadySubscribed             = New(ErrorTypeConflict, "Already subscribed to user")
	ErrUserBlocked                   = New(ErrorTypeForbidden, "Cannot perform action on blocked user")
)