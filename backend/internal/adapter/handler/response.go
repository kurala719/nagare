// package api provides HTTP handlers for the web server API.
package handler

import (
	"errors"
	"net/http"

	"nagare/internal/core/domain"

	"github.com/gin-gonic/gin"
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
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = "resource not found"
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		message = "invalid input"
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrPermissionDenied):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		message = "resource already exists"
	case errors.Is(err, domain.ErrAuthenticationFailed):
		status = http.StatusUnauthorized
		message = "authentication failed"
	case errors.Is(err, domain.ErrConnectionFailed):
		status = http.StatusServiceUnavailable
		message = "connection failed"
	case errors.Is(err, domain.ErrTimeout):
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
