package domain

import "errors"

// Common errors used across the domain layer
var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates the input data is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates the request is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrConflict indicates a conflict with existing data
	ErrConflict = errors.New("resource already exists")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal server error")

	// ErrConnectionFailed indicates a connection failure
	ErrConnectionFailed = errors.New("connection failed")

	// ErrAuthenticationFailed indicates authentication failure
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrTimeout indicates a timeout occurred
	ErrTimeout = errors.New("operation timed out")

	// ErrPermissionDenied indicates permission is denied
	ErrPermissionDenied = errors.New("permission denied")

	// ErrForbidden indicates a forbidden action
	ErrForbidden = errors.New("forbidden")
)
