package main

import (
	"fmt"
	"log"
	"os"

	client "github.com/davidt58/go-builder-relayer-client/client"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	relayerURL := os.Getenv("RELAYER_URL")
	chainID := parseInt64(os.Getenv("CHAIN_ID"))

	c, err := client.NewRelayClient(relayerURL, chainID, "", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetNonce("0x6e0c80c90ea6c15917308F820Eac91Ce2724B5b5", "SAFE")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)
}
