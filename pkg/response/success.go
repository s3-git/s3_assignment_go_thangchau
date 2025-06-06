package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents the structure of success responses
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
}

// SendSuccess sends a success response
func SendSuccess(c *gin.Context, data interface{}, message ...string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusOK, response)
}

// SendCreated sends a created response
func SendCreated(c *gin.Context, data interface{}, message ...string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusCreated, response)
}