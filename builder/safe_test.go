package builder

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
)

// TestSignEIP712StructHash verifies that our signature generation matches Python implementation
func TestSignEIP712StructHash(t *testing.T) {
	// Test data from Python test: tests/builder/test_safe.py
	structHashHex := "0x06d5102c3e356b62a75f8203cd5ce7ab1fa8fdab33875ef621eee102220d90b8"
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	expectedSigHex := "0xad62657208a0d885f91bba7490de238741bf7c51eb792f00856171aafc9e012373156fb672e55d840733c8bf723ec458545fcd5749aa5e547f808c222e7e11701c"

	// Remove 0x prefix
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	structHashHex = strings.TrimPrefix(structHashHex, "0x")

	// Create signer
	sig, err := signer.NewSigner(privateKeyHex, 137)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Decode struct hash
	structHashBytes, err := hex.DecodeString(structHashHex)
	if err != nil {
		t.Fatalf("Failed to decode struct hash: %v", err)
	}

	// Sign the struct hash
	signature, err := sig.SignEIP712StructHash(structHashBytes)
	if err != nil {
		t.Fatalf("Failed to sign struct hash: %v", err)
	}

	// Compare signatures (case-insensitive)
	if !strings.EqualFold(signature, expectedSigHex) {
		t.Errorf("Signature mismatch:\nGot:      %s\nExpected: %s", signature, expectedSigHex)
	} else {
		t.Logf("✓ Signature matches Python implementation: %s", signature)
	}
}

// TestSplitAndPackSig verifies signature packing matches Python
func TestSplitAndPackSig(t *testing.T) {
	// Test data from Python test: tests/builder/test_safe.py
	inputSig := "0xad62657208a0d885f91bba7490de238741bf7c51eb792f00856171aafc9e012373156fb672e55d840733c8bf723ec458545fcd5749aa5e547f808c222e7e11701c"
	expectedPacked := "0xad62657208a0d885f91bba7490de238741bf7c51eb792f00856171aafc9e012373156fb672e55d840733c8bf723ec458545fcd5749aa5e547f808c222e7e117020"

	// Pack the signature
	packed, err := SplitAndPackSig(inputSig)
	if err != nil {
		t.Fatalf("Failed to pack signature: %v", err)
	}

	// Compare (case-insensitive)
	if !strings.EqualFold(packed, expectedPacked) {
		t.Errorf("Packed signature mismatch:\nGot:      %s\nExpected: %s", packed, expectedPacked)
	} else {
		t.Logf("✓ Packed signature matches Python implementation: %s", packed)
	}
}

// TestCreateSafeStructHash verifies struct hash computation matches Python
func TestCreateSafeStructHash(t *testing.T) {
	// Test data from Python test: tests/model/test_safe_tx.py
	expectedStructHash := "0x06d5102c3e356b62a75f8203cd5ce7ab1fa8fdab33875ef621eee102220d90b8"

	// Build SafeTx struct
	safeTx := &SafeTx{
		To:             common.HexToAddress("0xA238CBeb142c10Ef7Ad8442C6D1f9E89e07e7761"),
		Value:          common.Big0,
		Data:           common.FromHex("0x8d80ff0a00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000132002791bca1f2de4661ed88a30c99a7a9449aa8417400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000044095ea7b30000000000000000000000004d97dcd97ec945f40cf65f87097ace5ea0476045ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff002791bca1f2de4661ed88a30c99a7a9449aa8417400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000044095ea7b30000000000000000000000004d97dcd97ec945f40cf65f87097ace5ea0476045ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000"),
		Operation:      1, // DelegateCall
		SafeTxGas:      common.Big0,
		BaseGas:        common.Big0,
		GasPrice:       common.Big0,
		GasToken:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		RefundReceiver: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Nonce:          big.NewInt(8),
	}

	verifyingContract := common.HexToAddress("0xd93B25cb943D14d0d34FBaF01Fc93a0f8b5F6E47")
	chainID := int64(137)

	// Build struct hash
	structHash, err := BuildSafeTxHash(safeTx, verifyingContract, chainID)
	if err != nil {
		t.Fatalf("Failed to build struct hash: %v", err)
	}

	// Compare (case-insensitive)
	if !strings.EqualFold(structHash.Hex(), expectedStructHash) {
		t.Errorf("Struct hash mismatch:\nGot:      %s\nExpected: %s", structHash.Hex(), expectedStructHash)
	} else {
		t.Logf("✓ Struct hash matches Python implementation: %s", structHash.Hex())
	}
}
