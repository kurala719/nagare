package model

import "errors"

// Common errors used across the domain layer
var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates the input data is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidEmail indicates the email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrWeakPassword indicates the password does not meet security requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters and include 3 of: lowercase, uppercase, digits, special characters")

	// ErrInvalidUsername indicates the username format is invalid
	ErrInvalidUsername = errors.New("username must be 3-32 characters, alphanumeric with underscores/hyphens only")

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
