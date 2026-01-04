package builder

import (
	"testing"

	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/davidt58/go-builder-relayer-client/signer"
)

func TestBuildSafeTransactionRequest(t *testing.T) {
	// Setup test data
	signerKey := "0x1234..."
	chainID := int64(80002)
	safeTransactionArgs := models.SafeTransactionArgs{
		// Populate with necessary fields for testing
	}
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		t.Fatalf("Failed to get contract config: %v", err)
	}

	s, err := signer.NewSigner(signerKey, chainID)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Call the function to test
	txnRequest, err := BuildSafeTransactionRequest(s, safeTransactionArgs, *contractConfig, "test metadata")
	if err != nil {
		t.Fatalf("Failed to build safe transaction request: %v", err)
	}

	// Validate the transaction request
	if txnRequest == nil {
		t.Fatal("Expected a transaction request, got nil")
	}
	// Add more assertions as needed
}

func TestBuildSafeCreateTransactionRequest(t *testing.T) {
	// Setup test data
	signerKey := "0x1234..."
	chainID := int64(80002)
	safeCreateArgs := models.SafeCreateTransactionArgs{
		// Populate with necessary fields for testing
	}
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		t.Fatalf("Failed to get contract config: %v", err)
	}

	s, err := signer.NewSigner(signerKey, chainID)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Call the function to test
	txnRequest, err := BuildSafeCreateTransactionRequest(s, safeCreateArgs, *contractConfig)
	if err != nil {
		t.Fatalf("Failed to build safe create transaction request: %v", err)
	}

	// Validate the transaction request
	if txnRequest == nil {
		t.Fatal("Expected a transaction request, got nil")
	}
	// Add more assertions as needed
}
