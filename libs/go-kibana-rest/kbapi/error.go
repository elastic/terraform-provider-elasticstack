package kbapi

import (
	"fmt"
)

// APIError is the error object
type APIError struct {
	Code    int
	Message string
}

// Error return error message
func (e APIError) Error() string {
	return e.Message
}

// NewAPIError create new API error with code and message
func NewAPIError(code int, message string, params ...interface{}) APIError {
	return APIError{
		Code:    code,
		Message: fmt.Sprintf(message, params...),
	}
}
