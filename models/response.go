package models

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
	// client reference will be set when creating responses (Phase 8)
	client interface{}
}

// NewClientRelayerTransactionResponse creates a new response wrapper
func NewClientRelayerTransactionResponse(transactionID string) *ClientRelayerTransactionResponse {
	return &ClientRelayerTransactionResponse{
		TransactionID: transactionID,
	}
}
