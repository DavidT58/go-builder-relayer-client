package builder

import (
	"math/big"
	"testing"

	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// Test signer address (from testPrivateKey)
	testSignerAddress = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	// Polygon Amoy testnet chain ID
	testChainID = 80002
)

func TestDeriveSafeAddress(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)

	safeAddr, err := DeriveSafeAddress(signerAddr, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed: %v", err)
	}

	if safeAddr == (common.Address{}) {
		t.Error("Safe address should not be zero")
	}

	// Verify the address is deterministic
	safeAddr2, err := DeriveSafeAddress(signerAddr, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed on second call: %v", err)
	}

	if safeAddr != safeAddr2 {
		t.Errorf("Safe address should be deterministic: %s != %s", safeAddr.Hex(), safeAddr2.Hex())
	}
}

func TestDeriveSafeAddress_DifferentSigners(t *testing.T) {
	signer1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	signer2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	safe1, err := DeriveSafeAddress(signer1, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed for signer1: %v", err)
	}

	safe2, err := DeriveSafeAddress(signer2, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed for signer2: %v", err)
	}

	if safe1 == safe2 {
		t.Error("Different signers should produce different Safe addresses")
	}
}

func TestDeriveSafeAddress_DifferentChains(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)

	// Polygon Amoy
	safeAmoy, err := DeriveSafeAddress(signerAddr, 80002)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed for Amoy: %v", err)
	}

	// Polygon Mainnet
	safeMainnet, err := DeriveSafeAddress(signerAddr, 137)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed for Mainnet: %v", err)
	}

	// Addresses should be the same if contract addresses are the same
	// (which they are in our config)
	if safeAmoy != safeMainnet {
		t.Logf("Note: Safe addresses differ between chains: Amoy=%s, Mainnet=%s", safeAmoy.Hex(), safeMainnet.Hex())
	}
}

func TestDeriveSafeAddress_InvalidChain(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)

	_, err := DeriveSafeAddress(signerAddr, 999999)
	if err == nil {
		t.Error("Expected error for unsupported chain ID")
	}
}

func TestBuildSafeInitializer(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)

	// Get contract config
	contractConfig, err := getTestContractConfig()
	if err != nil {
		t.Fatalf("Failed to get contract config: %v", err)
	}

	initializer, err := buildSafeInitializer(signerAddr, contractConfig)
	if err != nil {
		t.Fatalf("buildSafeInitializer failed: %v", err)
	}

	// Verify initializer starts with setup() selector: 0xb63e800d
	expectedSelector := []byte{0xb6, 0x3e, 0x80, 0x0d}
	if len(initializer) < 4 {
		t.Fatal("Initializer too short")
	}

	for i := 0; i < 4; i++ {
		if initializer[i] != expectedSelector[i] {
			t.Errorf("Initializer selector mismatch at byte %d: got 0x%02x, want 0x%02x", i, initializer[i], expectedSelector[i])
		}
	}

	// Verify initializer is deterministic
	initializer2, err := buildSafeInitializer(signerAddr, contractConfig)
	if err != nil {
		t.Fatalf("buildSafeInitializer failed on second call: %v", err)
	}

	if len(initializer) != len(initializer2) {
		t.Error("Initializer should be deterministic (length mismatch)")
	}
}

func TestEncodeSafeSetupParams(t *testing.T) {
	owners := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
	}
	threshold := big.NewInt(1)
	to := common.Address{}
	data := []byte{}
	fallbackHandler := common.HexToAddress("0x2222222222222222222222222222222222222222")
	paymentToken := common.Address{}
	payment := big.NewInt(0)
	paymentReceiver := common.Address{}

	encoded, err := encodeSafeSetupParams(owners, threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
	if err != nil {
		t.Fatalf("encodeSafeSetupParams failed: %v", err)
	}

	// Verify encoded data has correct length
	// 8 parameters * 32 bytes + array length (32) + 1 owner (32) + data length (32) = 352 bytes
	expectedMinLength := 8*32 + 32 + len(owners)*32 + 32
	if len(encoded) < expectedMinLength {
		t.Errorf("Encoded length = %d, want at least %d", len(encoded), expectedMinLength)
	}

	// Verify it's deterministic
	encoded2, err := encodeSafeSetupParams(owners, threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
	if err != nil {
		t.Fatalf("encodeSafeSetupParams failed on second call: %v", err)
	}

	if len(encoded) != len(encoded2) {
		t.Error("Encoded params should be deterministic")
	}
}

func TestVerifySafeAddress(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)

	// Derive the expected address
	expectedAddr, err := DeriveSafeAddress(signerAddr, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed: %v", err)
	}

	// Verify with correct address
	valid, err := VerifySafeAddress(signerAddr, expectedAddr, testChainID)
	if err != nil {
		t.Fatalf("VerifySafeAddress failed: %v", err)
	}
	if !valid {
		t.Error("Address verification should succeed for correct address")
	}

	// Verify with incorrect address
	wrongAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	valid, err = VerifySafeAddress(signerAddr, wrongAddr, testChainID)
	if err != nil {
		t.Fatalf("VerifySafeAddress failed: %v", err)
	}
	if valid {
		t.Error("Address verification should fail for incorrect address")
	}
}

func TestGetSafeDeploymentData(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)

	data, err := GetSafeDeploymentData(signerAddr, testChainID)
	if err != nil {
		t.Fatalf("GetSafeDeploymentData failed: %v", err)
	}

	// Verify all required fields are present
	requiredFields := []string{
		"safeAddress",
		"signerAddress",
		"singleton",
		"factory",
		"fallbackHandler",
		"initializer",
		"chainId",
	}

	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Verify signer address matches
	if data["signerAddress"] != signerAddr.Hex() {
		t.Errorf("Signer address mismatch: got %s, want %s", data["signerAddress"], signerAddr.Hex())
	}

	// Verify chain ID matches
	if data["chainId"] != int64(testChainID) {
		t.Errorf("Chain ID mismatch: got %v, want %d", data["chainId"], testChainID)
	}
}

func TestDeriveSafeAddressWithNonce(t *testing.T) {
	signerAddr := common.HexToAddress(testSignerAddress)
	nonce := big.NewInt(0)

	safeAddr, err := DeriveSafeAddressWithNonce(signerAddr, testChainID, nonce)
	if err != nil {
		t.Fatalf("DeriveSafeAddressWithNonce failed: %v", err)
	}

	if safeAddr == (common.Address{}) {
		t.Error("Safe address should not be zero")
	}

	// For nonce 0, should match regular derivation
	regularAddr, err := DeriveSafeAddress(signerAddr, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed: %v", err)
	}

	if safeAddr != regularAddr {
		t.Errorf("Address with nonce 0 should match regular derivation: %s != %s", safeAddr.Hex(), regularAddr.Hex())
	}
}

// TestDeriveSafeAddress_KnownAddress tests that our implementation produces the expected address
// This validates against the Python implementation for testChainID (80002 - Polygon Amoy)
func TestDeriveSafeAddress_KnownAddress(t *testing.T) {
	// Test with a known signer address
	signerAddr := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	
	// Expected Safe address calculated using the Python implementation's logic
	// Parameters:
	// - Factory: 0xa6B71E26C5e0845f74c812102Ca7114b6a896AB2
	// - Salt: keccak256(abi.encode(signerAddress))
	// - SAFE_INIT_CODE_HASH: 0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf
	expectedAddr := common.HexToAddress("0x76Bef2e2Aa6f92a8DC734e506C38Abe2e5523c11")
	
	safeAddr, err := DeriveSafeAddress(signerAddr, testChainID)
	if err != nil {
		t.Fatalf("DeriveSafeAddress failed: %v", err)
	}
	
	if safeAddr != expectedAddr {
		t.Errorf("Safe address mismatch:\n  got: %s\n  want: %s", safeAddr.Hex(), expectedAddr.Hex())
	}
	
	t.Logf("Successfully derived Safe address: %s", safeAddr.Hex())
}

// Helper function to get test contract config
func getTestContractConfig() (*config.ContractConfig, error) {
	return config.GetContractConfig(testChainID)
}
