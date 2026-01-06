package builder

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// TestSignatureRecovery verifies that signatures can be recovered to the correct address
func TestSignatureRecovery(t *testing.T) {
	// Use the known test case
	structHashHex := "0x06d5102c3e356b62a75f8203cd5ce7ab1fa8fdab33875ef621eee102220d90b8"
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	expectedAddr := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" // Address for this private key

	// Create signer
	sig, err := signer.NewSigner(strings.TrimPrefix(privateKeyHex, "0x"), 137)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	t.Logf("Signer address: %s", sig.AddressHex())
	t.Logf("Expected address: %s", expectedAddr)

	// Sign the struct hash
	structHashBytes, _ := hex.DecodeString(strings.TrimPrefix(structHashHex, "0x"))
	signature, err := sig.SignEIP712StructHash(structHashBytes)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	t.Logf("Generated signature: %s", signature)

	// The signature was created by signing: keccak256("\x19Ethereum Signed Message:\n32" + structHash)
	prefix := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(structHashBytes)))
	prefixedMessage := append(prefix, structHashBytes...)
	finalHash := crypto.Keccak256(prefixedMessage)

	t.Logf("Final hash (with EIP-191): %s", hexutil.Encode(finalHash))

	// Decode signature
	sigBytes := common.FromHex(signature)
	if len(sigBytes) != 65 {
		t.Fatalf("Invalid signature length: %d", len(sigBytes))
	}

	// The signature has v=27/28, adjust for recovery (needs 0/1)
	v := sigBytes[64]
	t.Logf("V value in signature: %d", v)

	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Recover public key from the EIP-191 prefixed hash
	pubKey, err := crypto.SigToPub(finalHash, sigBytes)
	if err != nil {
		t.Fatalf("Failed to recover pubkey: %v", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	t.Logf("Recovered address from EIP-191 hash: %s", recoveredAddr.Hex())

	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr) {
		t.Errorf("Recovered address does not match expected!\nGot:      %s\nExpected: %s", recoveredAddr.Hex(), expectedAddr)
	}

	// Now test what Safe contract would do with v=31 (after our transformation)
	// Safe will do: ecrecover(keccak256("\x19Ethereum Signed Message:\n32" + dataHash), v-4, r, s)
	// So it will use v=27 (31-4) and apply EIP-191 prefix to the ORIGINAL struct hash

	// This means Safe expects the signature to be on the ORIGINAL hash (without EIP-191)
	// But then Safe applies EIP-191 during recovery

	// Let's test: sign WITHOUT EIP-191, then verify with Safe's logic
	t.Log("\n=== Testing Safe's verification logic ===")

	// Sign the raw struct hash (no EIP-191 prefix)
	rawSig, err := sig.Sign(structHashBytes)
	if err != nil {
		t.Fatalf("Failed to sign raw hash: %v", err)
	}

	t.Logf("Raw signature (no EIP-191): %s", rawSig)

	rawSigBytes := common.FromHex(rawSig)
	vRaw := rawSigBytes[64]
	t.Logf("V value in raw signature: %d", vRaw)

	// Transform v like our SplitAndPackSig does: 27→31, 28→32
	vTransformed := vRaw + 4
	t.Logf("V after transformation (Safe format): %d", vTransformed)

	// Now simulate Safe's recovery for v>30:
	// ecrecover(keccak256("\x19Ethereum Signed Message:\n32" + dataHash), v-4, r, s)
	vForRecovery := vTransformed - 4 - 27 // Subtract 4 (Safe does this), then 27 for ecrecover
	rawSigBytes[64] = byte(vForRecovery)

	// Safe applies EIP-191 to the original hash
	safeHash := crypto.Keccak256(prefixedMessage)
	pubKey2, err := crypto.SigToPub(safeHash, rawSigBytes)
	if err != nil {
		t.Fatalf("Failed to recover with Safe logic: %v", err)
	}

	recoveredAddr2 := crypto.PubkeyToAddress(*pubKey2)
	t.Logf("Recovered address with Safe logic: %s", recoveredAddr2.Hex())

	if !strings.EqualFold(recoveredAddr2.Hex(), expectedAddr) {
		t.Errorf("Safe logic recovery does not match!\nGot:      %s\nExpected: %s", recoveredAddr2.Hex(), expectedAddr)
	}
}

// TestActualWithdrawalSignature tests the signature from the actual error log
func TestActualWithdrawalSignature(t *testing.T) {
	// From error log
	structHashHex := "0xd1ae1f033e4a482b669d36f20baa5501bdce81e8bc82f2ffd99056c85927d5f9"
	generatedSig := "0x944a2858f1615becfdf1de5076fbe229b79b7abeb1a23c41e230491c7b3dbf9f5916e03bf1e1c37e7d4bbe2eddd3fe2f656cdede38cb43ddc2558fa2a58705441b"
	sentSig := "0x944a2858f1615becfdf1de5076fbe229b79b7abeb1a23c41e230491c7b3dbf9f5916e03bf1e1c37e7d4bbe2eddd3fe2f656cdede38cb43ddc2558fa2a58705441f"
	expectedEOA := "0x09f3293e08A8FA65EB0b7749A8f99B23318ccc17"

	t.Logf("Struct hash: %s", structHashHex)
	t.Logf("Generated signature: %s", generatedSig)
	t.Logf("Sent signature (v transformed): %s", sentSig)
	t.Logf("Expected EOA: %s", expectedEOA)

	// The signature was created with EIP-191 prefix
	structHashBytes := common.FromHex(structHashHex)
	prefix := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(structHashBytes)))
	prefixedMessage := append(prefix, structHashBytes...)
	finalHash := crypto.Keccak256(prefixedMessage)

	t.Logf("Hash with EIP-191 prefix: %s", hexutil.Encode(finalHash))

	// Recover from generated signature (v=27)
	genSigBytes := common.FromHex(generatedSig)
	if genSigBytes[64] >= 27 {
		genSigBytes[64] -= 27
	}

	pubKey, err := crypto.SigToPub(finalHash, genSigBytes)
	if err != nil {
		t.Fatalf("Failed to recover from generated sig: %v", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	t.Logf("Recovered from generated sig: %s", recoveredAddr.Hex())

	if !strings.EqualFold(recoveredAddr.Hex(), expectedEOA) {
		t.Errorf("Does not match expected EOA!\nGot:      %s\nExpected: %s", recoveredAddr.Hex(), expectedEOA)
	} else {
		t.Log("✓ Generated signature recovers to correct EOA")
	}

	// Now test what Safe would see with v=31
	// Safe does: ecrecover(keccak256("\x19Ethereum Signed Message:\n32" + originalHash), v-4, r, s)
	sentSigBytes := common.FromHex(sentSig)
	vSent := sentSigBytes[64] // Should be 31 (0x1f)
	t.Logf("V in sent signature: %d", vSent)

	// Safe subtracts 4: 31-4=27
	vSafe := vSent - 4
	t.Logf("V after Safe subtracts 4: %d", vSafe)

	// Safe applies EIP-191 to the ORIGINAL hash (but we already did that!)
	// So Safe will hash: keccak256("\x19Ethereum Signed Message:\n32" + keccak256("\x19Ethereum Signed Message:\n32" + originalHash))
	// This is DOUBLE prefixing!

	t.Log("\n⚠️  ISSUE: We're applying EIP-191, then Safe applies it again = double prefixing!")
	t.Log("Solution: Sign WITHOUT EIP-191 prefix, let Safe add it")
}
