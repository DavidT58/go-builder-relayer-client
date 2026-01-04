package errors

import (
	"errors"
	"testing"
)

func TestRelayerClientError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *RelayerClientError
		expected string
	}{
		{
			name:     "simple error",
			err:      NewRelayerClientError("test error", nil),
			expected: "relayer client error: test error",
		},
		{
			name:     "error with underlying",
			err:      NewRelayerClientError("test error", errors.New("underlying")),
			expected: "relayer client error: test error: underlying",
		},
		{
			name:     "error with code",
			err:      NewRelayerClientErrorWithCode("test error", "TEST_CODE", nil),
			expected: "relayer client error: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRelayerClientError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewRelayerClientError("test", underlying)

	if unwrapped := err.Unwrap(); unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

func TestRelayerApiError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *RelayerApiError
		expected string
	}{
		{
			name:     "simple error",
			err:      NewRelayerApiError(404, "not found"),
			expected: "relayer api error (status 404): not found",
		},
		{
			name:     "error with code",
			err:      NewRelayerApiErrorWithCode(400, "bad request", "INVALID_REQUEST"),
			expected: "relayer api error (status 400, code INVALID_REQUEST): bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrSignerNotConfigured", ErrSignerNotConfigured},
		{"ErrBuilderCredsNotConfigured", ErrBuilderCredsNotConfigured},
		{"ErrInvalidPrivateKey", ErrInvalidPrivateKey(nil)},
		{"ErrInvalidAddress", ErrInvalidAddress("0x123")},
		{"ErrInvalidChainID", ErrInvalidChainID(999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("Error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}
