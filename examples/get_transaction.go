package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	client "github.com/davidt58/go-builder-relayer-client/client"
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

	c, err := client.NewRelayClient(relayerURL, chainID, "", nil)
	if err != nil {
		log.Fatal(err)
	}

	transactionID := "your_transaction_id_here" // Replace with the actual transaction ID
	resp, err := c.GetTransaction(transactionID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction details:", resp)
}
