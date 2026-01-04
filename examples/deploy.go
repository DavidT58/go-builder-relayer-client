package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
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
	fmt.Println("starting...")
	
	godotenv.Load()

	relayerURL := os.Getenv("RELAYER_URL")
	chainID := parseInt64(os.Getenv("CHAIN_ID"))
	pk := os.Getenv("PK")

	// Debug: Print configuration
	fmt.Printf("[DEBUG] Configuration:\n")
	fmt.Printf("  RELAYER_URL: %s\n", relayerURL)
	fmt.Printf("  CHAIN_ID: %d\n", chainID)
	fmt.Printf("  PK (first 10 chars): %s...\n", pk[:10])
	fmt.Printf("  BUILDER_API_KEY: %s\n", os.Getenv("BUILDER_API_KEY"))

	builderConfig := config.NewBuilderConfig(
		os.Getenv("BUILDER_API_KEY"),
		os.Getenv("BUILDER_SECRET"),
		os.Getenv("BUILDER_PASS_PHRASE"),
	)

	fmt.Println("[DEBUG] Creating RelayClient...")
	c, err := client.NewRelayClient(relayerURL, chainID, pk, builderConfig)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[DEBUG] RelayClient created successfully")

	fmt.Println("[DEBUG] Calling Deploy()...")
	resp, err := c.Deploy()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)

	fmt.Println("[DEBUG] Calling Wait()...")
	txn, err := resp.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(txn.ToFormattedString())
}
