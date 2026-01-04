package signer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestHashTypedData_SimpleDomain(t *testing.T) {
	typedData := &TypedData{
		Types: map[string][]EIP712Type{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "EIP712Domain",
		Domain: EIP712Domain{
			Name:              "Test",
			Version:           "1",
			ChainId:           big.NewInt(1),
			VerifyingContract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		},
		Message: map[string]interface{}{},
	}

	hash, err := HashTypedData(typedData)
	if err != nil {
		t.Fatalf("HashTypedData failed: %v", err)
	}

	if hash == (common.Hash{}) {
		t.Error("Hash should not be zero")
	}
}

func TestHashTypedData_WithMessage(t *testing.T) {
	typedData := &TypedData{
		Types: map[string][]EIP712Type{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"Person": {
				{Name: "name", Type: "string"},
				{Name: "wallet", Type: "address"},
			},
		},
		PrimaryType: "Person",
		Domain: EIP712Domain{
			Name:    "Test App",
			Version: "1",
			ChainId: big.NewInt(1),
		},
		Message: map[string]interface{}{
			"name":   "Alice",
			"wallet": "0x0000000000000000000000000000000000000001",
		},
	}

	hash, err := HashTypedData(typedData)
	if err != nil {
		t.Fatalf("HashTypedData failed: %v", err)
	}

	if hash == (common.Hash{}) {
		t.Error("Hash should not be zero")
	}

	// The hash should be deterministic
	hash2, err := HashTypedData(typedData)
	if err != nil {
		t.Fatalf("HashTypedData failed on second call: %v", err)
	}

	if hash != hash2 {
		t.Error("Hash should be deterministic")
	}
}

func TestEncodeType(t *testing.T) {
	types := map[string][]EIP712Type{
		"Person": {
			{Name: "name", Type: "string"},
			{Name: "wallet", Type: "address"},
		},
	}

	encoded, err := encodeType("Person", types)
	if err != nil {
		t.Fatalf("encodeType failed: %v", err)
	}

	expected := "Person(string name,address wallet)"
	if encoded != expected {
		t.Errorf("Encoded type = %s, want %s", encoded, expected)
	}
}

func TestEncodeType_NotFound(t *testing.T) {
	types := map[string][]EIP712Type{}

	_, err := encodeType("Person", types)
	if err == nil {
		t.Error("Expected error for non-existent type")
	}
}

func TestEncodeValue_String(t *testing.T) {
	types := map[string][]EIP712Type{}

	encoded, err := encodeValue("string", "hello", types)
	if err != nil {
		t.Fatalf("encodeValue failed: %v", err)
	}

	if len(encoded) != 32 {
		t.Errorf("Encoded length = %d, want 32", len(encoded))
	}
}

func TestEncodeValue_Address(t *testing.T) {
	types := map[string][]EIP712Type{}
	addr := "0x1234567890123456789012345678901234567890"

	encoded, err := encodeValue("address", addr, types)
	if err != nil {
		t.Fatalf("encodeValue failed: %v", err)
	}

	if len(encoded) != 32 {
		t.Errorf("Encoded length = %d, want 32", len(encoded))
	}

	// Verify the address is right-padded (last 20 bytes)
	decodedAddr := common.BytesToAddress(encoded[12:])
	if decodedAddr.Hex() != common.HexToAddress(addr).Hex() {
		t.Errorf("Decoded address = %s, want %s", decodedAddr.Hex(), addr)
	}
}

func TestEncodeValue_Uint256(t *testing.T) {
	types := map[string][]EIP712Type{}

	tests := []interface{}{
		"123",
		big.NewInt(123),
		int64(123),
		123,
	}

	for _, value := range tests {
		encoded, err := encodeValue("uint256", value, types)
		if err != nil {
			t.Fatalf("encodeValue failed for %T: %v", value, err)
		}

		if len(encoded) != 32 {
			t.Errorf("Encoded length = %d, want 32", len(encoded))
		}
	}
}

func TestEncodeValue_Bool(t *testing.T) {
	types := map[string][]EIP712Type{}

	// Test true
	encodedTrue, err := encodeValue("bool", true, types)
	if err != nil {
		t.Fatalf("encodeValue failed: %v", err)
	}
	if len(encodedTrue) != 32 {
		t.Errorf("Encoded length = %d, want 32", len(encodedTrue))
	}
	if encodedTrue[31] != 1 {
		t.Error("Bool true should encode to 1")
	}

	// Test false
	encodedFalse, err := encodeValue("bool", false, types)
	if err != nil {
		t.Fatalf("encodeValue failed: %v", err)
	}
	if encodedFalse[31] != 0 {
		t.Error("Bool false should encode to 0")
	}
}

func TestToMap(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "map",
			input: map[string]interface{}{"key": "value"},
		},
		{
			name: "struct",
			input: EIP712Domain{
				Name:    "Test",
				Version: "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toMap(tt.input)
			if err != nil {
				t.Fatalf("toMap failed: %v", err)
			}
			if result == nil {
				t.Error("Result should not be nil")
			}
		})
	}
}

func TestHashDomain(t *testing.T) {
	types := map[string][]EIP712Type{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
	}

	domain := EIP712Domain{
		Name:              "Test",
		Version:           "1",
		ChainId:           big.NewInt(1),
		VerifyingContract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	hash, err := hashDomain(domain, types)
	if err != nil {
		t.Fatalf("hashDomain failed: %v", err)
	}

	if hash == (common.Hash{}) {
		t.Error("Hash should not be zero")
	}

	// Hash should be deterministic
	hash2, err := hashDomain(domain, types)
	if err != nil {
		t.Fatalf("hashDomain failed on second call: %v", err)
	}

	if hash != hash2 {
		t.Error("Domain hash should be deterministic")
	}
}
