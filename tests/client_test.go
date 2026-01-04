package tests

import (
	"testing"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
)

func TestGetNonce(t *testing.T) {
	builderConfig := config.NewBuilderConfig("test_api_key", "test_secret", "test_passphrase")
	relayClient, err := client.NewRelayClient("https://relayer-v2-staging.polymarket.dev/", 80002, "your_private_key", builderConfig)
	if err != nil {
		t.Fatalf("Failed to create RelayClient: %v", err)
	}

	nonce, err := relayClient.GetNonce("0x6e0c80c90ea6c15917308F820Eac91Ce2724B5b5", "SAFE")
	if err != nil {
		t.Fatalf("Failed to get nonce: %v", err)
	}

	if nonce == nil {
		t.Error("Expected nonce to be non-nil")
	}
}

func TestGetTransaction(t *testing.T) {
	builderConfig := config.NewBuilderConfig("test_api_key", "test_secret", "test_passphrase")
	relayClient, err := client.NewRelayClient("https://relayer-v2-staging.polymarket.dev/", 80002, "your_private_key", builderConfig)
	if err != nil {
		t.Fatalf("Failed to create RelayClient: %v", err)
	}

	transactionID := "test_transaction_id"
	transaction, err := relayClient.GetTransaction(transactionID)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}

	if transaction == nil {
		t.Error("Expected transaction to be non-nil")
	}
}

func TestDeploy(t *testing.T) {
	builderConfig := config.NewBuilderConfig("test_api_key", "test_secret", "test_passphrase")
	relayClient, err := client.NewRelayClient("https://relayer-v2-staging.polymarket.dev/", 80002, "your_private_key", builderConfig)
	if err != nil {
		t.Fatalf("Failed to create RelayClient: %v", err)
	}

	response, err := relayClient.Deploy()
	if err != nil {
		t.Fatalf("Failed to deploy: %v", err)
	}

	if response.TransactionID == "" {
		t.Error("Expected TransactionID to be non-empty")
	}
}
