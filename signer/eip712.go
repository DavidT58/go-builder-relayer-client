package signer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/davidt58/go-builder-relayer-client/errors"
)

// EIP712Domain represents the EIP-712 domain separator
type EIP712Domain struct {
	Name              string         `json:"name,omitempty"`
	Version           string         `json:"version,omitempty"`
	ChainId           *big.Int       `json:"chainId,omitempty"`
	VerifyingContract common.Address `json:"verifyingContract,omitempty"`
	Salt              *common.Hash   `json:"salt,omitempty"`
}

// EIP712Type represents a type definition in EIP-712
type EIP712Type struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TypedData represents EIP-712 typed data
type TypedData struct {
	Types       map[string][]EIP712Type `json:"types"`
	PrimaryType string                  `json:"primaryType"`
	Domain      EIP712Domain            `json:"domain"`
	Message     map[string]interface{}  `json:"message"`
}

// HashTypedData computes the EIP-712 hash of typed data
func HashTypedData(typedData *TypedData) (common.Hash, error) {
	// Hash the domain separator
	domainSeparator, err := hashDomain(typedData.Domain, typedData.Types)
	if err != nil {
		return common.Hash{}, err
	}

	// If the primary type is EIP712Domain, just return the domain hash wrapped
	if typedData.PrimaryType == "EIP712Domain" {
		// Special case: just hashing the domain itself
		rawData := []byte{0x19, 0x01}
		rawData = append(rawData, domainSeparator[:]...)
		// For domain-only, we hash it with itself
		rawData = append(rawData, domainSeparator[:]...)
		return crypto.Keccak256Hash(rawData), nil
	}

	// Hash the message
	messageHash, err := hashStruct(typedData.PrimaryType, typedData.Message, typedData.Types)
	if err != nil {
		return common.Hash{}, err
	}

	// Compute final hash: keccak256("\x19\x01" ‖ domainSeparator ‖ messageHash)
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator[:]...)
	rawData = append(rawData, messageHash[:]...)

	return crypto.Keccak256Hash(rawData), nil
}

// hashDomain hashes the EIP712Domain according to EIP-712
func hashDomain(domain EIP712Domain, types map[string][]EIP712Type) (common.Hash, error) {
	// Get the EIP712Domain type definition
	domainTypes, exists := types["EIP712Domain"]
	if !exists {
		// Use default EIP712Domain type
		domainTypes = []EIP712Type{
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		}
	}

	// Compute type hash
	typeHash, err := hashType("EIP712Domain", domainTypes)
	if err != nil {
		return common.Hash{}, err
	}

	// Encode domain data
	var encoded []byte
	encoded = append(encoded, typeHash[:]...)

	// Convert domain to map
	domainMap := structToMap(domain)

	// Encode each field in the domain type
	for _, field := range domainTypes {
		value, exists := domainMap[field.Name]
		if !exists {
			// Skip optional fields
			continue
		}

		encodedValue, err := encodeValue(field.Type, value, types)
		if err != nil {
			return common.Hash{}, err
		}
		encoded = append(encoded, encodedValue...)
	}

	return crypto.Keccak256Hash(encoded), nil
}

// hashStruct hashes a struct according to EIP-712
func hashStruct(primaryType string, data interface{}, types map[string][]EIP712Type) (common.Hash, error) {
	typeFields, exists := types[primaryType]
	if !exists {
		return common.Hash{}, errors.NewRelayerClientError(fmt.Sprintf("type %s not found", primaryType), nil)
	}

	// Compute type hash
	typeHash, err := hashType(primaryType, typeFields)
	if err != nil {
		return common.Hash{}, err
	}

	// Encode the data
	encodedData, err := encodeData(primaryType, data, types)
	if err != nil {
		return common.Hash{}, err
	}

	// Concatenate type hash and encoded data, then hash
	encoded := append(typeHash[:], encodedData...)
	return crypto.Keccak256Hash(encoded), nil
}

// hashType computes the type hash for a given type name and fields
func hashType(typeName string, typeFields []EIP712Type) (common.Hash, error) {
	typeStr := encodeTypeString(typeName, typeFields)
	return crypto.Keccak256Hash([]byte(typeStr)), nil
}

// encodeTypeString encodes a type definition as a string
func encodeTypeString(typeName string, typeFields []EIP712Type) string {
	var result strings.Builder
	result.WriteString(typeName)
	result.WriteString("(")

	for i, field := range typeFields {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(field.Type)
		result.WriteString(" ")
		result.WriteString(field.Name)
	}
	result.WriteString(")")

	return result.String()
}

// encodeType encodes a type name according to EIP-712 (kept for compatibility)
func encodeType(primaryType string, types map[string][]EIP712Type) (string, error) {
	typeFields, exists := types[primaryType]
	if !exists {
		return "", errors.NewRelayerClientError(fmt.Sprintf("type %s not found", primaryType), nil)
	}

	return encodeTypeString(primaryType, typeFields), nil
}

// typeHash computes the type hash for a given type name (kept for compatibility)
func typeHash(primaryType string, types map[string][]EIP712Type) (common.Hash, error) {
	typeStr, err := encodeType(primaryType, types)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash([]byte(typeStr)), nil
}

// encodeData encodes struct data according to EIP-712
func encodeData(primaryType string, data interface{}, types map[string][]EIP712Type) ([]byte, error) {
	typeFields, exists := types[primaryType]
	if !exists {
		return nil, errors.NewRelayerClientError(fmt.Sprintf("type %s not found", primaryType), nil)
	}

	// Convert data to map
	dataMap, err := toMap(data)
	if err != nil {
		return nil, err
	}

	// Encode each field
	var encoded []byte
	for _, field := range typeFields {
		value, exists := dataMap[field.Name]
		if !exists {
			return nil, errors.NewRelayerClientError(fmt.Sprintf("field %s not found in data", field.Name), nil)
		}

		encodedValue, err := encodeValue(field.Type, value, types)
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, encodedValue...)
	}

	return encoded, nil
}

// encodeValue encodes a single value according to EIP-712
func encodeValue(fieldType string, value interface{}, types map[string][]EIP712Type) ([]byte, error) {
	// Handle different types
	switch {
	case fieldType == "string":
		str, ok := value.(string)
		if !ok {
			return nil, errors.NewRelayerClientError(fmt.Sprintf("expected string, got %T", value), nil)
		}
		hash := crypto.Keccak256Hash([]byte(str))
		return hash[:], nil

	case fieldType == "bytes":
		var bytes []byte
		switch v := value.(type) {
		case string:
			bytes, _ = hexutil.Decode(v)
		case []byte:
			bytes = v
		default:
			return nil, errors.NewRelayerClientError(fmt.Sprintf("expected bytes, got %T", value), nil)
		}
		hash := crypto.Keccak256Hash(bytes)
		return hash[:], nil

	case strings.HasPrefix(fieldType, "bytes"):
		// Fixed-size bytes
		var bytes []byte
		switch v := value.(type) {
		case string:
			bytes, _ = hexutil.Decode(v)
		case []byte:
			bytes = v
		default:
			return nil, errors.NewRelayerClientError(fmt.Sprintf("expected bytes, got %T", value), nil)
		}
		// Pad to 32 bytes
		padded := make([]byte, 32)
		copy(padded, bytes)
		return padded, nil

	case fieldType == "address":
		var addr common.Address
		switch v := value.(type) {
		case string:
			addr = common.HexToAddress(v)
		case common.Address:
			addr = v
		default:
			return nil, errors.NewRelayerClientError(fmt.Sprintf("expected address, got %T", value), nil)
		}
		padded := make([]byte, 32)
		copy(padded[12:], addr[:])
		return padded, nil

	case strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "int"):
		// Integer types
		var bigInt *big.Int
		switch v := value.(type) {
		case string:
			bigInt = new(big.Int)
			bigInt.SetString(v, 0)
		case *big.Int:
			bigInt = v
		case int64:
			bigInt = big.NewInt(v)
		case int:
			bigInt = big.NewInt(int64(v))
		case float64:
			bigInt = big.NewInt(int64(v))
		default:
			return nil, errors.NewRelayerClientError(fmt.Sprintf("expected integer, got %T", value), nil)
		}
		bytes := bigInt.Bytes()
		padded := make([]byte, 32)
		copy(padded[32-len(bytes):], bytes)
		return padded, nil

	case fieldType == "bool":
		boolVal, ok := value.(bool)
		if !ok {
			return nil, errors.NewRelayerClientError(fmt.Sprintf("expected bool, got %T", value), nil)
		}
		padded := make([]byte, 32)
		if boolVal {
			padded[31] = 1
		}
		return padded, nil

	default:
		// Check if it's a struct type
		if _, exists := types[fieldType]; exists {
			// Recursively hash the struct
			hash, err := hashStruct(fieldType, value, types)
			if err != nil {
				return nil, err
			}
			return hash[:], nil
		}

		return nil, errors.NewRelayerClientError(fmt.Sprintf("unsupported type: %s", fieldType), nil)
	}
}

// toMap converts various data types to a map[string]interface{}
func toMap(data interface{}) (map[string]interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return v, nil
	case EIP712Domain:
		return structToMap(v), nil
	default:
		// Try JSON marshaling/unmarshaling
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, errors.ErrJSONMarshalFailed(err)
		}
		var result map[string]interface{}
		err = json.Unmarshal(jsonData, &result)
		if err != nil {
			return nil, errors.ErrJSONUnmarshalFailed(err)
		}
		return result, nil
	}
}

// structToMap converts a struct to a map using reflection
func structToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// Get JSON tag name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// Remove options from tag
		tagName := strings.Split(jsonTag, ",")[0]

		// Skip omitempty fields that are zero values
		if strings.Contains(jsonTag, "omitempty") && value.IsZero() {
			continue
		}

		result[tagName] = value.Interface()
	}

	return result
}
