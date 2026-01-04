package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
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

	resp, err := c.Deploy()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction ID:", resp.TransactionID)

	txn, err := resp.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deployed:", txn)
}
