package builder

import (
	"testing"

	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/davidt58/go-builder-relayer-client/signer"
)

func TestBuildSafeTransactionRequest(t *testing.T) {
	// Setup test data
	signerKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80" // test private key
	chainID := int64(80002)

	s, err := signer.NewSigner(signerKey, chainID)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Get expected Safe address
	safeAddress, err := DeriveSafeAddress(s.Address(), chainID)
	if err != nil {
		t.Fatalf("Failed to derive safe address: %v", err)
	}

	safeTransactionArgs := &models.SafeTransactionArgs{
		SafeAddress: safeAddress.Hex(),
		Transactions: []models.SafeTransaction{
			{
				To:        "0x0000000000000000000000000000000000000001",
				Value:     "0",
				Data:      "0x",
				Operation: models.Call,
			},
		},
		Nonce:    "0",
		Metadata: "test metadata",
	}

	// Call the function to test
	txnRequest, err := BuildSafeTransactionRequest(safeTransactionArgs, s, chainID)
	if err != nil {
		t.Fatalf("Failed to build safe transaction request: %v", err)
	}

	// Validate the transaction request
	if txnRequest == nil {
		t.Fatal("Expected a transaction request, got nil")
	}

	if txnRequest.Type != string(models.SAFE) {
		t.Errorf("Expected type SAFE, got %s", txnRequest.Type)
	}

	if len(txnRequest.Signatures) == 0 {
		t.Error("Expected at least one signature")
	}
}

func TestBuildSafeCreateTransactionRequest(t *testing.T) {
	// Setup test data
	signerKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80" // test private key
	chainID := int64(80002)

	s, err := signer.NewSigner(signerKey, chainID)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Get expected Safe address
	safeAddress, err := DeriveSafeAddress(s.Address(), chainID)
	if err != nil {
		t.Fatalf("Failed to derive safe address: %v", err)
	}

	safeCreateArgs := &models.SafeCreateTransactionArgs{
		SignerAddress: s.AddressHex(),
		SafeAddress:   safeAddress.Hex(),
		Nonce:         "0",
		Metadata:      "test metadata",
	}

	// Call the function to test
	txnRequest, err := BuildSafeCreateTransactionRequest(safeCreateArgs, s, chainID)
	if err != nil {
		t.Fatalf("Failed to build safe create transaction request: %v", err)
	}

	// Validate the transaction request
	if txnRequest == nil {
		t.Fatal("Expected a transaction request, got nil")
	}

	if txnRequest.Type != string(models.SAFE_CREATE) {
		t.Errorf("Expected type SAFE-CREATE, got %s", txnRequest.Type)
	}

	if len(txnRequest.Signatures) == 0 {
		t.Error("Expected at least one signature")
	}
}

func TestSplitSignature(t *testing.T) {
	// Create a valid test signature (65 bytes)
	// r (32 bytes) + s (32 bytes) + v (1 byte)
	signerKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	chainID := int64(80002)

	s, err := signer.NewSigner(signerKey, chainID)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Sign a test message
	testMessage := []byte("test message")
	testHash := signer.Keccak256Hash(testMessage)
	signature, err := s.Sign(testHash.Bytes())
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Now test signature splitting
	r, sSig, v, err := SplitSignature(signature)
	if err != nil {
		t.Fatalf("Failed to split signature: %v", err)
	}

	if r == "" || sSig == "" {
		t.Error("r or s is empty")
	}

	// v should be adjusted to 31 or 32 for Safe
	if v != 31 && v != 32 {
		t.Errorf("Expected v to be 31 or 32, got %d", v)
	}
}

