// package api provides HTTP handlers for the web server API.
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"nagare/internal/model"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// respondSuccess sends a successful JSON response
func respondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// respondSuccessMessage sends a successful JSON response with a message
func respondSuccessMessage(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
	})
}

// respondError sends an error JSON response with appropriate status code
func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := err.Error()

	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
		message = "resource not found"
	case errors.Is(err, model.ErrInvalidInput):
		status = http.StatusBadRequest
		message = "invalid input"
	case errors.Is(err, model.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, model.ErrForbidden), errors.Is(err, model.ErrPermissionDenied):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, model.ErrConflict):
		status = http.StatusConflict
		message = "resource already exists"
	case errors.Is(err, model.ErrAuthenticationFailed):
		status = http.StatusUnauthorized
		message = "authentication failed"
	case errors.Is(err, model.ErrConnectionFailed):
		status = http.StatusServiceUnavailable
		message = "connection failed"
	case errors.Is(err, model.ErrTimeout):
		status = http.StatusGatewayTimeout
		message = "operation timed out"
	}

	c.JSON(status, APIResponse{
		Success: false,
		Error:   message,
	})
}

// respondBadRequest sends a 400 Bad Request response
func respondBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error:   message,
	})
}
