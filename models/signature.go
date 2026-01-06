package models

import (
	"encoding/json"
)

// SignatureParams contains the parameters for signing a Safe transaction
type SignatureParams struct {
	// SAFE transaction params
	GasPrice       *string `json:"gasPrice,omitempty"`
	Operation      *string `json:"operation,omitempty"`
	SafeTxGas      *string `json:"safeTxnGas,omitempty"`
	BaseGas        *string `json:"baseGas,omitempty"`
	GasToken       *string `json:"gasToken,omitempty"`
	RefundReceiver *string `json:"refundReceiver,omitempty"`

	// SAFE-CREATE transaction params
	PaymentToken    *string `json:"paymentToken,omitempty"`
	Payment         *string `json:"payment,omitempty"`
	PaymentReceiver *string `json:"paymentReceiver,omitempty"`
}

// SplitSig represents a split ECDSA signature (r, s, v components)
type SplitSig struct {
	// R is the r component of the signature
	R string `json:"r"`
	// S is the s component of the signature
	S string `json:"s"`
	// V is the v component of the signature (recovery id)
	V int `json:"v"`
}

// NewSplitSig creates a new SplitSig from signature bytes
func NewSplitSig(r, s string, v int) *SplitSig {
	return &SplitSig{
		R: r,
		S: s,
		V: v,
	}
}

// Signature represents a complete signature with all components
type Signature struct {
	// Signer is the address of the signer
	Signer string `json:"signer"`
	// Data is the signature data (full signature as hex string)
	Data string `json:"data,omitempty"`
	// Split contains the split signature components (r, s, v)
	Split *SplitSig `json:"split,omitempty"`
}

// NewSignature creates a new Signature
func NewSignature(signer, data string) *Signature {
	return &Signature{
		Signer: signer,
		Data:   data,
	}
}

// TransactionRequest represents a request to submit a transaction to the relayer
type TransactionRequest struct {
	// Type is the transaction type (SAFE or SAFE-CREATE)
	Type string `json:"type"`
	// From is the signer address (EOA for SAFE-CREATE, Safe address for SAFE)
	From string `json:"from"`
	// To is the destination address(es) - can be string or array
	To json.RawMessage `json:"to"`
	// ProxyWallet is the Safe wallet address
	ProxyWallet string `json:"proxyWallet"`
	// Data is the transaction data - can be string or array
	Data json.RawMessage `json:"data"`
	// Signature is the transaction signature
	Signature string `json:"signature"`
	// SignatureParams contains additional signature parameters
	SignatureParams *SignatureParams `json:"signatureParams,omitempty"`
	// Value is the value(s) to send - can be string or array (optional)
	Value json.RawMessage `json:"value,omitempty"`
	// Operation is the operation type(s) - can be int or array (optional)
	Operation json.RawMessage `json:"operation,omitempty"`
	// Nonce is the transaction nonce (optional)
	Nonce *string `json:"nonce,omitempty"`
	// Metadata is optional metadata for the transaction
	Metadata *string `json:"metadata,omitempty"`
}

// SafeTransactionData represents the structured data for a Safe transaction
type SafeTransactionData struct {
	// To is the destination address
	To string `json:"to"`
	// Value is the value to send
	Value string `json:"value"`
	// Data is the transaction data
	Data string `json:"data"`
	// Operation is the operation type
	Operation OperationType `json:"operation"`
	// SafeTxGas is the Safe transaction gas
	SafeTxGas string `json:"safeTxGas"`
	// BaseGas is the base gas
	BaseGas string `json:"baseGas"`
	// GasPrice is the gas price
	GasPrice string `json:"gasPrice"`
	// GasToken is the gas token address
	GasToken string `json:"gasToken"`
	// RefundReceiver is the refund receiver address
	RefundReceiver string `json:"refundReceiver"`
	// Nonce is the transaction nonce
	Nonce string `json:"nonce"`
}

// EIP712Domain represents the EIP-712 domain separator
type EIP712Domain struct {
	// ChainID is the chain ID
	ChainID int64 `json:"chainId"`
	// VerifyingContract is the contract address
	VerifyingContract string `json:"verifyingContract"`
}

// EIP712TypedData represents EIP-712 typed data for signing
type EIP712TypedData struct {
	// Types contains the type definitions
	Types map[string][]EIP712Type `json:"types"`
	// PrimaryType is the primary type name
	PrimaryType string `json:"primaryType"`
	// Domain is the domain separator
	Domain EIP712Domain `json:"domain"`
	// Message is the message data
	Message interface{} `json:"message"`
}

// EIP712Type represents a type definition in EIP-712
type EIP712Type struct {
	// Name is the field name
	Name string `json:"name"`
	// Type is the field type
	Type string `json:"type"`
}
