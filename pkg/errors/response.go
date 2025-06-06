package errors

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Error   ErrorDetails `json:"error"`
	Success bool         `json:"success"`
}

// ErrorDetails contains the detailed error information
type ErrorDetails struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// HandleError handles AppError and sends appropriate HTTP response
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	appErr := FromError(err)

	errorResponse := ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Type:    appErr.Type,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	c.JSON(appErr.GetStatusCode(), errorResponse)
}

// SendBadRequest sends a bad request error
func SendBadRequest(c *gin.Context, message string, details ...string) {
	appErr := New(ErrorTypeValidation, message)
	if len(details) > 0 {
		appErr.WithDetails(details[0])
	}
	HandleError(c, appErr)
}

// SendNotFound sends a not found error
func SendNotFound(c *gin.Context, message string) {
	HandleError(c, New(ErrorTypeNotFound, message))
}

// SendInternalError sends an internal server error
func SendInternalError(c *gin.Context, err error) {
	HandleError(c, Wrap(err, ErrorTypeInternal, "Internal server error"))
}

// SendConflict sends a conflict error
func SendConflict(c *gin.Context, message string) {
	HandleError(c, New(ErrorTypeConflict, message))
}

// HandleValidationErrors handles validation errors from the validator package
func HandleValidationErrors(c *gin.Context, validationErrors map[string]string) {
	appErr := New(ErrorTypeValidation, "Validation failed")

	// Convert validation errors to a details string
	if len(validationErrors) > 0 {
		var details string
		for field, msg := range validationErrors {
			if details != "" {
				details += "; "
			}
			details += field + ": " + msg
		}
		appErr.WithDetails(details)
	}

	HandleError(c, appErr)
}
