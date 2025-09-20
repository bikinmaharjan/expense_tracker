package utils

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// RespondWithError sends a JSON error response with the given status code
func RespondWithError(c *gin.Context, status int, err error, details string) {
	c.JSON(status, ErrorResponse{
		Error:   err.Error(),
		Details: details,
	})
}

// HandleValidationError handles validation errors and returns appropriate response
func HandleValidationError(c *gin.Context, err error) {
	RespondWithError(c, 400, err, "Validation failed")
}

// HandleDatabaseError handles database errors and returns appropriate response
func HandleDatabaseError(c *gin.Context, err error) {
	RespondWithError(c, 500, err, "Database operation failed")
}