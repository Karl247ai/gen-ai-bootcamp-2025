package errors

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	// Database error codes
	ErrDBConnection    ErrorCode = "DB_CONNECTION_ERROR"
	ErrDBQuery         ErrorCode = "DB_QUERY_ERROR"
	ErrDBTransaction   ErrorCode = "DB_TRANSACTION_ERROR"
	ErrDBDuplicate     ErrorCode = "DB_DUPLICATE_ERROR"
	ErrDBNotFound      ErrorCode = "DB_NOT_FOUND"

	// Validation error codes
	ErrValidation      ErrorCode = "VALIDATION_ERROR"
	ErrInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrInvalidID       ErrorCode = "INVALID_ID"

	// Authentication/Authorization
	ErrUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrForbidden       ErrorCode = "FORBIDDEN"

	// System errors
	ErrInternal        ErrorCode = "INTERNAL_ERROR"
	ErrTimeout         ErrorCode = "TIMEOUT"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
	Fields  map[string]interface{}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func WithFields(err *AppError, fields map[string]interface{}) *AppError {
	err.Fields = fields
	return err
}

// IsErrorCode checks if an error contains a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
} 