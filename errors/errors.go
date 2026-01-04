package errors

import (
	"fmt"
)

// RelayerClientError represents a client-side error
type RelayerClientError struct {
	// Message is the error message
	Message string
	// Code is an optional error code
	Code string
	// Err is the underlying error
	Err error
}

// Error implements the error interface
func (e *RelayerClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("relayer client error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("relayer client error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *RelayerClientError) Unwrap() error {
	return e.Err
}

// NewRelayerClientError creates a new RelayerClientError
func NewRelayerClientError(message string, err error) *RelayerClientError {
	return &RelayerClientError{
		Message: message,
		Err:     err,
	}
}

// NewRelayerClientErrorWithCode creates a new RelayerClientError with a code
func NewRelayerClientErrorWithCode(message, code string, err error) *RelayerClientError {
	return &RelayerClientError{
		Message: message,
		Code:    code,
		Err:     err,
	}
}

// RelayerApiError represents an error response from the Relayer API
type RelayerApiError struct {
	// StatusCode is the HTTP status code
	StatusCode int
	// Message is the error message from the API
	Message string
	// Code is the error code from the API
	Code string
	// Details contains additional error details
	Details interface{}
}

// Error implements the error interface
func (e *RelayerApiError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("relayer api error (status %d, code %s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("relayer api error (status %d): %s", e.StatusCode, e.Message)
}

// NewRelayerApiError creates a new RelayerApiError
func NewRelayerApiError(statusCode int, message string) *RelayerApiError {
	return &RelayerApiError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewRelayerApiErrorWithCode creates a new RelayerApiError with a code
func NewRelayerApiErrorWithCode(statusCode int, message, code string) *RelayerApiError {
	return &RelayerApiError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
	}
}

// NewRelayerApiErrorWithDetails creates a new RelayerApiError with details
func NewRelayerApiErrorWithDetails(statusCode int, message, code string, details interface{}) *RelayerApiError {
	return &RelayerApiError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Details:    details,
	}
}

// Common error constructors

// ErrSignerNotConfigured is returned when a signer is required but not configured
var ErrSignerNotConfigured = NewRelayerClientError("signer not configured", nil)

// ErrBuilderCredsNotConfigured is returned when builder credentials are required but not configured
var ErrBuilderCredsNotConfigured = NewRelayerClientError("builder credentials not configured", nil)

// ErrInvalidPrivateKey is returned when the private key is invalid
func ErrInvalidPrivateKey(err error) *RelayerClientError {
	return NewRelayerClientError("invalid private key", err)
}

// ErrInvalidAddress is returned when an address is invalid
func ErrInvalidAddress(address string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("invalid address: %s", address), nil)
}

// ErrInvalidChainID is returned when a chain ID is not supported
func ErrInvalidChainID(chainID int64) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("unsupported chain ID: %d", chainID), nil)
}

// ErrSigningFailed is returned when signature generation fails
func ErrSigningFailed(err error) *RelayerClientError {
	return NewRelayerClientError("signature generation failed", err)
}

// ErrInvalidSignature is returned when a signature is invalid
func ErrInvalidSignature(err error) *RelayerClientError {
	return NewRelayerClientError("invalid signature", err)
}

// ErrHTTPRequestFailed is returned when an HTTP request fails
func ErrHTTPRequestFailed(err error) *RelayerClientError {
	return NewRelayerClientError("HTTP request failed", err)
}

// ErrJSONMarshalFailed is returned when JSON marshaling fails
func ErrJSONMarshalFailed(err error) *RelayerClientError {
	return NewRelayerClientError("JSON marshal failed", err)
}

// ErrJSONUnmarshalFailed is returned when JSON unmarshaling fails
func ErrJSONUnmarshalFailed(err error) *RelayerClientError {
	return NewRelayerClientError("JSON unmarshal failed", err)
}

// ErrTransactionNotFound is returned when a transaction is not found
func ErrTransactionNotFound(transactionID string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("transaction not found: %s", transactionID), nil)
}

// ErrTransactionFailed is returned when a transaction fails
func ErrTransactionFailed(transactionID string, reason string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("transaction %s failed: %s", transactionID, reason), nil)
}

// ErrPollingTimeout is returned when polling times out
func ErrPollingTimeout(transactionID string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("polling timeout for transaction: %s", transactionID), nil)
}

// ErrInvalidResponse is returned when the API response is invalid
func ErrInvalidResponse(reason string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("invalid API response: %s", reason), nil)
}

// ErrMissingRequiredField is returned when a required field is missing
func ErrMissingRequiredField(fieldName string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("missing required field: %s", fieldName), nil)
}

// ErrInvalidConfiguration is returned when configuration is invalid
func ErrInvalidConfiguration(reason string) *RelayerClientError {
	return NewRelayerClientError(fmt.Sprintf("invalid configuration: %s", reason), nil)
}
