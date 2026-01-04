package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/joho/godotenv"
)

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
	ctf := "0x4d97dcd97ec945f40cf65f87097ace5ea0476045"

	txn := models.SafeTransaction{
		To:        usdc,
		Operation: models.Call,
		Data:      "0x095ea7b3.. .", // approve calldata
		Value:     "0",
	}

	resp, err := c.Execute([]models.SafeTransaction{txn}, "approve USDC on CTF")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction submitted:", resp.TransactionHash)

	result, err := resp.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction confirmed:", result)
}
