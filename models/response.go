package models

import "fmt"

// SubmitTransactionResponse represents the response from submitting a transaction
type SubmitTransactionResponse struct {
	// TransactionID is the unique identifier for the submitted transaction
	TransactionID string `json:"transactionId"`
	// State is the initial state of the transaction
	State RelayerTransactionState `json:"state,omitempty"`
}

// GetTransactionResponse is an alias for RelayerTransaction
type GetTransactionResponse = RelayerTransaction

// GetTransactionsResponse represents the response from getting multiple transactions
type GetTransactionsResponse struct {
	// Transactions is the list of transactions
	Transactions []RelayerTransaction `json:"transactions"`
	// Total is the total number of transactions
	Total int `json:"total,omitempty"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	// Error is the error message
	Error string `json:"error"`
	// Code is the error code (optional)
	Code *string `json:"code,omitempty"`
	// Details contains additional error details (optional)
	Details interface{} `json:"details,omitempty"`
}

// ClientRelayerTransactionResponse wraps a transaction response with helper methods
type ClientRelayerTransactionResponse struct {
	// TransactionID is the unique identifier for the transaction
	TransactionID string
	// client reference for making API calls
	client RelayClientInterface
}

// String returns a string representation of the response
func (r *ClientRelayerTransactionResponse) String() string {
	return fmt.Sprintf("Response{TransactionID: \"%s\"}", r.TransactionID)
}

// RelayClientInterface defines the interface needed by ClientRelayerTransactionResponse
type RelayClientInterface interface {
	GetTransaction(transactionID string) (*RelayerTransaction, error)
	PollUntilState(transactionID string, states []RelayerTransactionState, failState RelayerTransactionState, maxPolls, pollFrequency int) (*RelayerTransaction, error)
}

// NewClientRelayerTransactionResponse creates a new response wrapper
func NewClientRelayerTransactionResponse(transactionID string) *ClientRelayerTransactionResponse {
	return &ClientRelayerTransactionResponse{
		TransactionID: transactionID,
	}
}

// SetClient sets the client reference for making API calls
func (r *ClientRelayerTransactionResponse) SetClient(client RelayClientInterface) {
	r.client = client
}

// GetTransaction fetches the current transaction details
func (r *ClientRelayerTransactionResponse) GetTransaction() (*RelayerTransaction, error) {
	if r.client == nil {
		return nil, &ClientError{Message: "client not configured"}
	}
	return r.client.GetTransaction(r.TransactionID)
}

// Wait polls until the transaction reaches a terminal state (mined or confirmed)
// Polls for both STATE_MINED and STATE_CONFIRMED to match Python implementation behavior.
// Note: STATE_MINED is not a terminal state (it can progress to STATE_CONFIRMED),
// but it's considered a valid completion state for this method. This allows callers
// to act on transactions as soon as they're mined, without waiting for full confirmation.
// Default polling: max 100 polls, every 2 seconds
func (r *ClientRelayerTransactionResponse) Wait() (*RelayerTransaction, error) {
	if r.client == nil {
		return nil, &ClientError{Message: "client not configured"}
	}

	// Poll until mined or confirmed (matching Python's wait() method behavior)
	targetStates := []RelayerTransactionState{STATE_MINED, STATE_CONFIRMED}
	failState := STATE_FAILED

	return r.client.PollUntilState(r.TransactionID, targetStates, failState, 100, 2)
}

// WaitWithOptions polls until the transaction reaches a terminal state with custom options
func (r *ClientRelayerTransactionResponse) WaitWithOptions(maxPolls, pollFrequency int) (*RelayerTransaction, error) {
	if r.client == nil {
		return nil, &ClientError{Message: "client not configured"}
	}

	targetStates := []RelayerTransactionState{STATE_CONFIRMED}
	failState := STATE_FAILED

	return r.client.PollUntilState(r.TransactionID, targetStates, failState, maxPolls, pollFrequency)
}

// WaitUntilMined polls until the transaction is mined (may not be confirmed yet)
func (r *ClientRelayerTransactionResponse) WaitUntilMined() (*RelayerTransaction, error) {
	if r.client == nil {
		return nil, &ClientError{Message: "client not configured"}
	}

	targetStates := []RelayerTransactionState{STATE_MINED, STATE_CONFIRMED}
	failState := STATE_FAILED

	return r.client.PollUntilState(r.TransactionID, targetStates, failState, 100, 2)
}

// ClientError represents an error from the client helper methods
type ClientError struct {
	Message string
}

// Error implements the error interface
func (e *ClientError) Error() string {
	return e.Message
}
