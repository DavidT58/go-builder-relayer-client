package models

import (
	"encoding/json"
	"fmt"
)

// OperationType represents the type of operation for a Safe transaction
type OperationType int

const (
	// Call represents a standard call operation
	Call OperationType = 0
	// DelegateCall represents a delegate call operation
	DelegateCall OperationType = 1
)

// String returns the string representation of OperationType
func (o OperationType) String() string {
	switch o {
	case Call:
		return "Call"
	case DelegateCall:
		return "DelegateCall"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements json.Marshaler for OperationType
func (o OperationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(o))
}

// UnmarshalJSON implements json.Unmarshaler for OperationType
func (o *OperationType) UnmarshalJSON(data []byte) error {
	var val int
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*o = OperationType(val)
	return nil
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	// SAFE represents a standard Safe transaction
	SAFE TransactionType = "SAFE"
	// SAFE_CREATE represents a Safe wallet creation transaction
	SAFE_CREATE TransactionType = "SAFE-CREATE"
)

// String returns the string representation of TransactionType
func (t TransactionType) String() string {
	return string(t)
}

// RelayerTransactionState represents the state of a transaction in the relayer
type RelayerTransactionState string

const (
	// STATE_NEW indicates a newly submitted transaction
	STATE_NEW RelayerTransactionState = "STATE_NEW"
	// STATE_EXECUTED indicates the transaction has been executed
	STATE_EXECUTED RelayerTransactionState = "STATE_EXECUTED"
	// STATE_MINED indicates the transaction has been mined
	STATE_MINED RelayerTransactionState = "STATE_MINED"
	// STATE_CONFIRMED indicates the transaction has been confirmed
	STATE_CONFIRMED RelayerTransactionState = "STATE_CONFIRMED"
	// STATE_FAILED indicates the transaction has failed
	STATE_FAILED RelayerTransactionState = "STATE_FAILED"
	// STATE_INVALID indicates the transaction is invalid
	STATE_INVALID RelayerTransactionState = "STATE_INVALID"
)

// String returns the string representation of RelayerTransactionState
func (s RelayerTransactionState) String() string {
	return string(s)
}

// IsTerminal returns true if the state is a terminal state
func (s RelayerTransactionState) IsTerminal() bool {
	switch s {
	case STATE_CONFIRMED, STATE_FAILED, STATE_INVALID:
		return true
	default:
		return false
	}
}

// SafeTransaction represents a single transaction to be executed through a Safe
type SafeTransaction struct {
	// To is the destination address
	To string `json:"to"`
	// Value is the amount of native token to send (in wei, as string)
	Value string `json:"value"`
	// Data is the encoded function call data (hex string)
	Data string `json:"data"`
	// Operation is the type of operation (Call or DelegateCall)
	Operation OperationType `json:"operation"`
	// GasLimit is the gas limit for this transaction
	GasLimit string `json:"gasLimit,omitempty"`
}

// NewSafeTransaction creates a new SafeTransaction with default values
func NewSafeTransaction(to, value, data string) *SafeTransaction {
	return &SafeTransaction{
		To:        to,
		Value:     value,
		Data:      data,
		Operation: Call,
	}
}

// SafeTransactionArgs represents arguments for building a Safe transaction request
type SafeTransactionArgs struct {
	// SafeAddress is the address of the Safe wallet
	SafeAddress string
	// Transactions is the list of transactions to execute
	Transactions []SafeTransaction
	// Nonce is the Safe transaction nonce
	Nonce string
	// Metadata is optional metadata for the transaction
	Metadata string
}

// SafeCreateTransactionArgs represents arguments for building a Safe creation request
type SafeCreateTransactionArgs struct {
	// SignerAddress is the address of the signer who will own the Safe
	SignerAddress string
	// SafeAddress is the expected address of the Safe to be created
	SafeAddress string
	// Nonce is the nonce for the creation transaction
	Nonce string
	// Metadata is optional metadata for the transaction
	Metadata string
}

// RelayerTransaction represents a transaction in the relayer system
type RelayerTransaction struct {
	// TransactionID is the unique identifier for the transaction
	TransactionID string `json:"transactionId"`
	// State is the current state of the transaction
	State RelayerTransactionState `json:"state"`
	// Type is the type of transaction
	Type TransactionType `json:"type"`
	// SafeAddress is the address of the Safe
	SafeAddress string `json:"safeAddress"`
	// ChainID is the blockchain chain ID
	ChainID int64 `json:"chainId"`
	// Hash is the transaction hash (if mined)
	Hash *string `json:"hash,omitempty"`
	// BlockNumber is the block number (if mined)
	BlockNumber *int64 `json:"blockNumber,omitempty"`
	// CreatedAt is the timestamp when the transaction was created
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the timestamp when the transaction was last updated
	UpdatedAt string `json:"updatedAt"`
	// Metadata is optional metadata attached to the transaction
	Metadata *string `json:"metadata,omitempty"`
}

// IsMined returns true if the transaction has been mined
func (t *RelayerTransaction) IsMined() bool {
	return t.Hash != nil && *t.Hash != ""
}

// IsConfirmed returns true if the transaction has been confirmed
func (t *RelayerTransaction) IsConfirmed() bool {
	return t.State == STATE_CONFIRMED
}

// IsFailed returns true if the transaction has failed
func (t *RelayerTransaction) IsFailed() bool {
	return t.State == STATE_FAILED || t.State == STATE_INVALID
}

// ToFormattedString returns a formatted string representation of the transaction
// similar to Python's dictionary output
func (t *RelayerTransaction) ToFormattedString() string {
	hashStr := "nil"
	if t.Hash != nil {
		hashStr = fmt.Sprintf("\"%s\"", *t.Hash)
	}
	
	blockNumStr := "nil"
	if t.BlockNumber != nil {
		blockNumStr = fmt.Sprintf("%d", *t.BlockNumber)
	}
	
	metadataStr := "nil"
	if t.Metadata != nil {
		metadataStr = fmt.Sprintf("\"%s\"", *t.Metadata)
	}
	
	return fmt.Sprintf("Transaction: {transactionID: \"%s\", state: \"%s\", type: \"%s\", safeAddress: \"%s\", chainId: %d, hash: %s, blockNumber: %s, createdAt: \"%s\", updatedAt: \"%s\", metadata: %s}",
		t.TransactionID, t.State, t.Type, t.SafeAddress, t.ChainID, hashStr, blockNumStr, t.CreatedAt, t.UpdatedAt, metadataStr)
}

// SignerType represents the type of signer
type SignerType string

const (
	// EOA represents an Externally Owned Account
	EOA SignerType = "EOA"
	// SAFE represents a Safe wallet
	SAFE_SIGNER SignerType = "SAFE"
)

// String returns the string representation of SignerType
func (s SignerType) String() string {
	return string(s)
}

// NonceResponse represents the response from get-nonce endpoint
type NonceResponse struct {
	// Nonce is the current nonce value as a string
	Nonce string `json:"nonce"`
}

// DeployedResponse represents the response from get-deployed endpoint
type DeployedResponse struct {
	// Deployed indicates whether the Safe is deployed
	Deployed bool `json:"deployed"`
	// SafeAddress is the address of the Safe
	SafeAddress string `json:"safeAddress,omitempty"`
}
