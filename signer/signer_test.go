package signer

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// Test private key (DO NOT use in production)
	testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	// Expected address for the test private key
	testAddress = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

func TestNewSigner(t *testing.T) {
	signer, err := NewSigner(testPrivateKey, 80002)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	if signer == nil {
		t.Fatal("Signer should not be nil")
	}

	// Check address
	if !strings.EqualFold(signer.AddressHex(), testAddress) {
		t.Errorf("Address = %s, want %s", signer.AddressHex(), testAddress)
	}

	// Check chain ID
	if signer.GetChainID().Int64() != 80002 {
		t.Errorf("ChainID = %d, want 80002", signer.GetChainID().Int64())
	}
}

func TestNewSigner_WithPrefix(t *testing.T) {
	// Test with "0x" prefix
	signer, err := NewSigner("0x"+testPrivateKey, 80002)
	if err != nil {
		t.Fatalf("NewSigner with prefix failed: %v", err)
	}

	if !strings.EqualFold(signer.AddressHex(), testAddress) {
		t.Errorf("Address = %s, want %s", signer.AddressHex(), testAddress)
	}
}

func TestNewSigner_InvalidKey(t *testing.T) {
	_, err := NewSigner("invalid", 80002)
	if err == nil {
		t.Error("Expected error for invalid private key")
	}
}

func TestSigner_SignAndVerify(t *testing.T) {
	signer, err := NewSigner(testPrivateKey, 80002)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	// Create a test message hash
	message := []byte("test message")
	messageHash := crypto.Keccak256Hash(message)

	// Sign the message hash
	signature, err := signer.SignEIP712StructHash(messageHash.Bytes())
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Verify signature is 65 bytes (132 hex chars + 0x prefix)
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		t.Fatalf("Failed to decode signature: %v", err)
	}
	if len(sigBytes) != 65 {
		t.Errorf("Signature length = %d, want 65", len(sigBytes))
	}

	// Verify the signature
	valid, err := signer.VerifySignature(messageHash.Bytes(), signature)
	if err != nil {
		t.Fatalf("VerifySignature failed: %v", err)
	}
	if !valid {
		t.Error("Signature verification failed")
	}
}

func TestSigner_SignMessage(t *testing.T) {
	signer, err := NewSigner(testPrivateKey, 80002)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	message := []byte("Hello, Ethereum!")
	signature, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	if signature == "" {
		t.Error("Signature should not be empty")
	}

	// Verify signature format
	if !strings.HasPrefix(signature, "0x") {
		t.Error("Signature should have 0x prefix")
	}
}

func TestSplitSignature(t *testing.T) {
	// Create a dummy signature (65 bytes)
	signature := make([]byte, 65)
	for i := range signature {
		signature[i] = byte(i)
	}
	signatureHex := hexutil.Encode(signature)

	r, s, v, err := SplitSignature(signatureHex)
	if err != nil {
		t.Fatalf("SplitSignature failed: %v", err)
	}

	// Verify r is 32 bytes
	rBytes, _ := hexutil.Decode(r)
	if len(rBytes) != 32 {
		t.Errorf("r length = %d, want 32", len(rBytes))
	}

	// Verify s is 32 bytes
	sBytes, _ := hexutil.Decode(s)
	if len(sBytes) != 32 {
		t.Errorf("s length = %d, want 32", len(sBytes))
	}

	// Verify v is the last byte
	if v != 64 {
		t.Errorf("v = %d, want 64", v)
	}
}

func TestPackSignatures(t *testing.T) {
	// Create two dummy signatures
	sig1 := make([]byte, 65)
	sig2 := make([]byte, 65)
	for i := range sig1 {
		sig1[i] = byte(i)
		sig2[i] = byte(i + 10)
	}

	signatures := []string{
		hexutil.Encode(sig1),
		hexutil.Encode(sig2),
	}

	packed, err := PackSignatures(signatures)
	if err != nil {
		t.Fatalf("PackSignatures failed: %v", err)
	}

	// Verify packed length is 130 bytes (65 * 2)
	packedBytes, _ := hexutil.Decode(packed)
	if len(packedBytes) != 130 {
		t.Errorf("Packed length = %d, want 130", len(packedBytes))
	}
}

func TestPackSignatures_Empty(t *testing.T) {
	_, err := PackSignatures([]string{})
	if err == nil {
		t.Error("Expected error for empty signatures")
	}
}

func TestRecoverAddress(t *testing.T) {
	signer, err := NewSigner(testPrivateKey, 80002)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	// Create and sign a message
	messageHash := crypto.Keccak256Hash([]byte("test"))
	signature, err := signer.SignEIP712StructHash(messageHash.Bytes())
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Recover the address
	sigBytes, _ := hexutil.Decode(signature)
	recovered, err := RecoverAddress(messageHash.Bytes(), sigBytes)
	if err != nil {
		t.Fatalf("RecoverAddress failed: %v", err)
	}

	if recovered != signer.Address() {
		t.Errorf("Recovered address = %s, want %s", recovered.Hex(), signer.AddressHex())
	}
}

func TestKeccak256(t *testing.T) {
	data := []byte("test")
	hash := Keccak256(data)

	if len(hash) != 32 {
		t.Errorf("Hash length = %d, want 32", len(hash))
	}

	// Verify against known hash
	expected := crypto.Keccak256(data)
	if hex.EncodeToString(hash) != hex.EncodeToString(expected) {
		t.Error("Hash mismatch")
	}
}

func TestKeccak256Hash(t *testing.T) {
	data := []byte("test")
	hash := Keccak256Hash(data)

	if hash == (common.Hash{}) {
		t.Error("Hash should not be zero")
	}

	// Verify against known hash
	expected := crypto.Keccak256Hash(data)
	if hash != expected {
		t.Error("Hash mismatch")
	}
}
