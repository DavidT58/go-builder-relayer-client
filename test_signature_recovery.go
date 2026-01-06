package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// The struct hash from the error log
	structHashHex := "0xd1ae1f033e4a482b669d36f20baa5501bdce81e8bc82f2ffd99056c85927d5f9"

	// Create signer with the private key
	privateKeyHex := "YOUR_PRIVATE_KEY_HERE" // Replace with actual key
	sig, err := signer.NewSigner(privateKeyHex, 137)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	fmt.Printf("Signer address: %s\n", sig.AddressHex())

	// Decode struct hash
	structHashBytes := common.FromHex(structHashHex)

	// Sign it
	signature, err := sig.SignEIP712StructHash(structHashBytes)
	if err != nil {
		log.Fatalf("Failed to sign: %v", err)
	}

	fmt.Printf("Generated signature: %s\n", signature)

	// Now let's recover the address from the signature
	// We need to recreate what was actually signed
	prefix := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(structHashBytes)))
	prefixedMessage := append(prefix, structHashBytes...)
	finalHash := crypto.Keccak256(prefixedMessage)

	fmt.Printf("Final hash that was signed: %s\n", hexutil.Encode(finalHash))

	// Decode signature
	sigBytes := common.FromHex(signature)
	if len(sigBytes) != 65 {
		log.Fatalf("Invalid signature length: %d", len(sigBytes))
	}

	// Adjust V for recovery (subtract 27)
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Recover public key
	pubKey, err := crypto.SigToPub(finalHash, sigBytes)
	if err != nil {
		log.Fatalf("Failed to recover pubkey: %v", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	fmt.Printf("Recovered address: %s\n", recoveredAddr.Hex())
	fmt.Printf("Expected address: %s\n", sig.AddressHex())

	if strings.EqualFold(recoveredAddr.Hex(), sig.AddressHex()) {
		fmt.Println("✓ Signature recovers to correct address!")
	} else {
		fmt.Println("✗ Signature does NOT recover to correct address!")
	}
}
