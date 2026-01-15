package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/davidt58/go-builder-relayer-client/client"
	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

// MaxUint256 is the maximum value for uint256
var MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

// Contract addresses for Polygon mainnet
var (
	USDC            = common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
	CTFExchange     = common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	NegRiskCTF      = common.HexToAddress("0xC5d563A36AE78145C45a50134d48A1215220f80a")
	NegRiskAdapter  = common.HexToAddress("0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296")
)

func parseInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse int64: %v", err)
	}
	return val
}

// encodeApprove encodes an ERC20 approve(address,uint256) function call
func encodeApprove(spender common.Address, amount *big.Int) (string, error) {
	// Define the approve function ABI
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)

	approveMethod := abi.NewMethod(
		"approve",
		"approve",
		abi.Function,
		"nonpayable",
		false,
		false,
		abi.Arguments{
			{Name: "spender", Type: addressType},
			{Name: "amount", Type: uint256Type},
		},
		nil,
	)

	// Pack the function call
	data, err := approveMethod.Inputs.Pack(spender, amount)
	if err != nil {
		return "", fmt.Errorf("failed to pack approve arguments: %w", err)
	}

	// Prepend function selector (first 4 bytes of keccak256("approve(address,uint256)"))
	return fmt.Sprintf("0x%x%x", approveMethod.ID, data), nil
}

// createUSDCApproveTxn creates a SafeTransaction for approving USDC spending
func createUSDCApproveTxn(token, spender common.Address) (models.SafeTransaction, error) {
	data, err := encodeApprove(spender, MaxUint256)
	if err != nil {
		return models.SafeTransaction{}, err
	}

	return models.SafeTransaction{
		To:        token.Hex(),
		Operation: models.Call,
		Data:      data,
		Value:     "0",
	}, nil
}

func main() {
	fmt.Println("Starting USDC approval transactions...")

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	relayerURL := os.Getenv("RELAYER_URL")
	if relayerURL == "" {
		relayerURL = "https://relayer-v2.polymarket.com"
	}

	chainIDStr := os.Getenv("CHAIN_ID")
	if chainIDStr == "" {
		chainIDStr = "137" // Polygon mainnet
	}
	chainID := parseInt64(chainIDStr)

	pk := os.Getenv("PK")
	if pk == "" {
		log.Fatal("PK environment variable is required")
	}

	builderConfig := config.NewBuilderConfig(
		os.Getenv("BUILDER_API_KEY"),
		os.Getenv("BUILDER_SECRET"),
		os.Getenv("BUILDER_PASS_PHRASE"),
	)

	c, err := client.NewRelayClient(relayerURL, chainID, pk, builderConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create approval transactions for all 3 contracts
	txn1, err := createUSDCApproveTxn(USDC, CTFExchange)
	if err != nil {
		log.Fatalf("Failed to create CTF Exchange approval: %v", err)
	}
	fmt.Printf("CTF Exchange approval data: %s\n", txn1.Data)

	txn2, err := createUSDCApproveTxn(USDC, NegRiskCTF)
	if err != nil {
		log.Fatalf("Failed to create NegRisk CTF approval: %v", err)
	}
	fmt.Printf("NegRisk CTF approval data: %s\n", txn2.Data)

	txn3, err := createUSDCApproveTxn(USDC, NegRiskAdapter)
	if err != nil {
		log.Fatalf("Failed to create NegRisk Adapter approval: %v", err)
	}
	fmt.Printf("NegRisk Adapter approval data: %s\n", txn3.Data)

	// Execute all 3 approval transactions in a single batch
	fmt.Println("\nSubmitting batch approval transaction...")
	resp, err := c.Execute([]models.SafeTransaction{txn1, txn2, txn3}, "approve USDC on CTF Exchange, NegRisk CTF, and NegRisk Adapter")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction ID:", resp.TransactionID)

	fmt.Println("Waiting for transaction to be confirmed...")
	result, err := resp.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transaction confirmed!\n")
	fmt.Printf("  State: %s\n", result.State)
	if result.Hash != nil {
		fmt.Printf("  Hash: %s\n", *result.Hash)
	}
	if result.BlockNumber != nil {
		fmt.Printf("  Block: %d\n", *result.BlockNumber)
	}
}
