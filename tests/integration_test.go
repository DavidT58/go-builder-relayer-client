package tests

import (
	"os"
	"testing"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/joho/godotenv"
)

func TestIntegration(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Fatalf("Error loading .env file")
	}

	relayerURL := os.Getenv("RELAYER_URL")
	chainID := parseInt64(os.Getenv("CHAIN_ID"))
	privateKey := os.Getenv("PK")

	builderConfig := config.NewBuilderConfig(
		os.Getenv("BUILDER_API_KEY"),
		os.Getenv("BUILDER_SECRET"),
		os.Getenv("BUILDER_PASS_PHRASE"),
	)

	c, err := client.NewRelayClient(relayerURL, chainID, privateKey, builderConfig)
	if err != nil {
		t.Fatalf("Failed to create RelayClient: %v", err)
	}

	// Example integration test for GetNonce
	nonceResp, err := c.GetNonce("0x6e0c80c90ea6c15917308F820Eac91Ce2724B5b5", "SAFE")
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	if nonceResp == nil {
		t.Fatal("Expected nonce response, got nil")
	}

	// Additional integration tests can be added here
}
