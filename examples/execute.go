package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/joho/godotenv"
)

func parseInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse int64: %v", err)
	}
	return val
}

func main() {
	godotenv.Load()

	relayerURL := os.Getenv("RELAYER_URL")
	chainID := parseInt64(os.Getenv("CHAIN_ID"))
	pk := os.Getenv("PK")

	builderConfig := config.NewBuilderConfig(
		os.Getenv("BUILDER_API_KEY"),
		os.Getenv("BUILDER_SECRET"),
		os.Getenv("BUILDER_PASS_PHRASE"),
	)

	c, err := client.NewRelayClient(relayerURL, chainID, pk, builderConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create USDC approval transaction
	usdc := "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"

	txn := models.SafeTransaction{
		To:        usdc,
		Operation: models.Call,
		// Example: ERC20 approve(spender, amount)
		// This is a placeholder - replace with actual encoded approve parameters
		// Format: 0x095ea7b3 + abi.encode(address spender, uint256 amount)
		Data:  "0x095ea7b3",
		Value: "0",
	}

	resp, err := c.Execute([]models.SafeTransaction{txn}, "approve USDC on CTF")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction ID:", resp.TransactionID)

	result, err := resp.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction confirmed:", result)
}
