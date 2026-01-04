package models

import (
	"encoding/json"
)

// SignatureParams contains the parameters for signing a Safe transaction
type SignatureParams struct {
	// PaymentReceiver is the address that receives the payment (optional)
	PaymentReceiver *string `json:"paymentReceiver,omitempty"`
	// Payment is the amount to be paid (optional)
	Payment *string `json:"payment,omitempty"`
	// PaymentToken is the token address for payment (optional)
	PaymentToken *string `json:"paymentToken,omitempty"`
	// GasPrice is the gas price for the transaction
	GasPrice string `json:"gasPrice"`
	// Operation is the operation type
	Operation OperationType `json:"operation"`
	// SafeTxGas is the gas for the Safe transaction
	SafeTxGas string `json:"safeTxGas"`
	// BaseGas is the base gas for the transaction
	BaseGas string `json:"baseGas"`
	// GasToken is the token address for gas payment
	GasToken string `json:"gasToken"`
	// RefundReceiver is the address that receives gas refunds
	RefundReceiver string `json:"refundReceiver"`
	// Nonce is the Safe transaction nonce
	Nonce string `json:"nonce"`
}

// NewSignatureParams creates a new SignatureParams with default values
func NewSignatureParams(nonce string) *SignatureParams {
	return &SignatureParams{
		GasPrice:       "0",
		Operation:      Call,
		SafeTxGas:      "0",
		BaseGas:        "0",
		GasToken:       "0x0000000000000000000000000000000000000000",
		RefundReceiver: "0x0000000000000000000000000000000000000000",
		Nonce:          nonce,
	}
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
	// Metadata is optional metadata for the transaction
	Metadata *string `json:"metadata,omitempty"`
	// SafeAddress is the address of the Safe wallet
	SafeAddress string `json:"safeAddress"`
	// To is the destination address(es) - can be string or array
	To json.RawMessage `json:"to"`
	// Value is the value(s) to send - can be string or array
	Value json.RawMessage `json:"value"`
	// Data is the transaction data - can be string or array
	Data json.RawMessage `json:"data"`
	// Operation is the operation type(s) - can be int or array
	Operation json.RawMessage `json:"operation,omitempty"`
	// Signatures is the array of signatures
	Signatures []Signature `json:"signatures"`
	// GasPrice is the gas price
	GasPrice string `json:"gasPrice,omitempty"`
	// SafeTxGas is the Safe transaction gas
	SafeTxGas string `json:"safeTxGas,omitempty"`
	// BaseGas is the base gas
	BaseGas string `json:"baseGas,omitempty"`
	// GasToken is the gas token address
	GasToken string `json:"gasToken,omitempty"`
	// RefundReceiver is the refund receiver address
	RefundReceiver string `json:"refundReceiver,omitempty"`
	// Nonce is the transaction nonce
	Nonce string `json:"nonce"`
	// ChainID is the blockchain chain ID
	ChainID int64 `json:"chainId,omitempty"`
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
