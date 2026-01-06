package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/davidt58/go-builder-relayer-client/builder"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	// From the error log
	privateKeyHex := "YOUR_PRIVATE_KEY" // You'll need to provide this
	chainID := int64(137)
	safeAddress := "0xe93E704C5f8aC34D0A179841C7661D4B2eCC46C6"
	usdcAddress := "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	recipientAddress := "0x7113C2394FcA480f4a3E7Ef30E70391c115E376c"
	amount := big.NewInt(3000000) // 3 USDC (6 decimals)

	// Create signer
	sig, err := signer.NewSigner(privateKeyHex, chainID)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	fmt.Printf("EOA Address: %s\n\n", sig.AddressHex())

	// Encode transfer function call
	// transfer(address recipient, uint256 amount)
	// Function selector: 0xa9059cbb
	transferData := make([]byte, 68)
	copy(transferData[0:4], common.Hex2Bytes("a9059cbb"))
	// Pad recipient address to 32 bytes
	copy(transferData[4+12:36], common.HexToAddress(recipientAddress).Bytes())
	// Amount as uint256
	amountBytes := amount.Bytes()
	copy(transferData[68-len(amountBytes):68], amountBytes)

	fmt.Printf("Transfer data: 0x%x\n\n", transferData)

	// Build Safe transaction args
	args := &models.SafeTransactionArgs{
		SafeAddress: safeAddress,
		Transactions: []models.SafeTransaction{
			{
				To:        usdcAddress,
				Value:     "0",
				Data:      common.Bytes2Hex(transferData),
				Operation: models.Call,
			},
		},
		Nonce:    "0",
		Metadata: "withdrawal",
	}

	// Create struct hash
	structHash, err := builder.CreateSafeStructHash(args, sig)
	if err != nil {
		log.Fatalf("Failed to create struct hash: %v", err)
	}

	fmt.Printf("EIP-712 Struct Hash: %s\n\n", structHash.Hex())

	// Create signature
	signature, err := builder.CreateSafeSignature(args, sig)
	if err != nil {
		log.Fatalf("Failed to create signature: %v", err)
	}

	fmt.Printf("Generated Signature: %s\n\n", signature)

	// Pack signature
	packedSig, err := builder.SplitAndPackSig(signature)
	if err != nil {
		log.Fatalf("Failed to pack signature: %v", err)
	}

	fmt.Printf("Packed Signature (v transformed): %s\n\n", packedSig)

	// Build full request
	request, err := builder.BuildSafeTransactionRequest(args, sig, chainID)
	if err != nil {
		log.Fatalf("Failed to build request: %v", err)
	}

	requestJSON, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("Full Request:\n%s\n", string(requestJSON))
}
