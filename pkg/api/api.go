// Package api provides specialized functionality for API testing in ffuf.
//
// This package contains modules for API endpoint discovery, request generation,
// response parsing, and authentication handling specifically tailored for
// testing and fuzzing web APIs.
package api

import (
	"fmt"
)

// Version represents the current version of the API package
const Version = "0.1.0"

// APIError represents an error that occurred during API operations
type APIError struct {
	Message string
	Code    int
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("API error (code %d): %s", e.Code, e.Message)
	}
	return fmt.Sprintf("API error: %s", e.Message)
}

// NewAPIError creates a new APIError with the given message and code
func NewAPIError(message string, code int) *APIError {
	return &APIError{
		Message: message,
		Code:    code,
	}
}