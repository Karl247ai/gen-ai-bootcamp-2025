package errors

import (
	"fmt"
	"testing"
)

func TestAppError(t *testing.T) {
	tests := []struct {
		name           string
		err            *AppError
		expectedString string
		expectedCode   ErrorCode
	}{
		{
			name: "error with underlying error",
			err: Wrap(
				fmt.Errorf("underlying error"),
				ErrDBConnection,
				"connection failed",
			),
			expectedString: "DB_CONNECTION_ERROR: connection failed: underlying error",
			expectedCode:   ErrDBConnection,
		},
		{
			name: "error without underlying error",
			err: New(
				ErrValidation,
				"invalid input",
			),
			expectedString: "VALIDATION_ERROR: invalid input",
			expectedCode:   ErrValidation,
		},
		{
			name: "error with fields",
			err: WithFields(
				New(ErrInvalidInput, "invalid field"),
				map[string]interface{}{"field": "username"},
			),
			expectedString: "INVALID_INPUT: invalid field",
			expectedCode:   ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedString {
				t.Errorf("expected error string %q, got %q", tt.expectedString, tt.err.Error())
			}

			if !IsErrorCode(tt.err, tt.expectedCode) {
				t.Errorf("expected error code %v, got %v", tt.expectedCode, tt.err.Code)
			}
		})
	}
} 