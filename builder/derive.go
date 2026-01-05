package builder

import (
	"math/big"

	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// SAFE_INIT_CODE_HASH is the pre-computed keccak256 hash of the Safe proxy init code
// This constant matches the Python implementation's SAFE_INIT_CODE_HASH
// This hash is derived from the Safe proxy factory's deployment bytecode for creating minimal proxies (EIP-1167)
// Reference: https://github.com/safe-global/safe-contracts
const SAFE_INIT_CODE_HASH = "0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf"

// DeriveSafeAddress calculates the Safe address using CREATE2
// This matches the Python implementation's derive_safe_address function
func DeriveSafeAddress(signerAddress common.Address, chainID int64) (common.Address, error) {
	// Get contract configuration for the chain
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return common.Address{}, err
	}

	// Get factory address
	factoryAddress := common.HexToAddress(contractConfig.SafeFactory)

	// Calculate salt as keccak256(abi.encode(signerAddress))
	// In Solidity ABI encoding, an address is left-padded to 32 bytes
	salt := crypto.Keccak256Hash(common.LeftPadBytes(signerAddress.Bytes(), 32))

	// Get the init code hash
	initCodeHash := common.HexToHash(SAFE_INIT_CODE_HASH)

	// Calculate CREATE2 address
	// Formula: keccak256(0xff ++ factoryAddress ++ salt ++ initCodeHash)[12:]
	data := make([]byte, 1+20+32+32)
	data[0] = 0xff
	copy(data[1:21], factoryAddress.Bytes())
	copy(data[21:53], salt.Bytes())
	copy(data[53:85], initCodeHash.Bytes())

	hash := crypto.Keccak256Hash(data)
	safeAddress := common.BytesToAddress(hash[12:])

	return safeAddress, nil
}

// buildSafeInitializer creates the initializer data for Safe.setup()
// This encodes the call to setup(owners, threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
// This function is still needed for Safe creation transactions (not for address derivation)
func buildSafeInitializer(signerAddress common.Address, contractConfig *config.ContractConfig) ([]byte, error) {
	// Safe.setup() function selector: 0xb63e800d
	setupSelector := crypto.Keccak256([]byte("setup(address[],uint256,address,bytes,address,address,uint256,address)"))[:4]

	// Encode the parameters for Safe.setup()
	// owners: [signerAddress]
	// threshold: 1
	// to: 0x0 (no delegate call during setup)
	// data: 0x (empty bytes)
	// fallbackHandler: from config
	// paymentToken: 0x0 (ETH)
	// payment: 0
	// paymentReceiver: 0x0

	// Build the ABI-encoded parameters
	encodedParams, err := encodeSafeSetupParams(
		[]common.Address{signerAddress}, // owners
		big.NewInt(1),                   // threshold
		common.Address{},                // to (zero address)
		[]byte{},                        // data (empty)
		common.HexToAddress(contractConfig.SafeFallbackHandler), // fallbackHandler
		common.Address{}, // paymentToken (zero address for ETH)
		big.NewInt(0),    // payment
		common.Address{}, // paymentReceiver (zero address)
	)
	if err != nil {
		return nil, err
	}

	// Concatenate selector + encoded params
	initializerData := append(setupSelector, encodedParams...)
	return initializerData, nil
}

// encodeSafeSetupParams encodes the parameters for the Safe.setup() function
func encodeSafeSetupParams(
	owners []common.Address,
	threshold *big.Int,
	to common.Address,
	data []byte,
	fallbackHandler common.Address,
	paymentToken common.Address,
	payment *big.Int,
	paymentReceiver common.Address,
) ([]byte, error) {
	// ABI encoding layout:
	// - offset to owners array (32 bytes)
	// - threshold (32 bytes)
	// - to address (32 bytes, left-padded)
	// - offset to data bytes (32 bytes)
	// - fallbackHandler (32 bytes, left-padded)
	// - paymentToken (32 bytes, left-padded)
	// - payment (32 bytes)
	// - paymentReceiver (32 bytes, left-padded)
	// - owners array length (32 bytes)
	// - owners array elements (32 bytes each, left-padded)
	// - data length (32 bytes)
	// - data bytes (padded to 32 byte boundary)

	var encoded []byte

	// Calculate offsets
	// Offset to owners array: 8 * 32 = 256 bytes (after the 8 fixed parameters)
	ownersOffset := big.NewInt(256)
	encoded = append(encoded, common.LeftPadBytes(ownersOffset.Bytes(), 32)...)

	// threshold
	encoded = append(encoded, common.LeftPadBytes(threshold.Bytes(), 32)...)

	// to address
	encoded = append(encoded, common.LeftPadBytes(to.Bytes(), 32)...)

	// Calculate offset to data
	// data offset = owners offset + 32 (length) + len(owners) * 32
	dataOffset := big.NewInt(int64(256 + 32 + len(owners)*32))
	encoded = append(encoded, common.LeftPadBytes(dataOffset.Bytes(), 32)...)

	// fallbackHandler
	encoded = append(encoded, common.LeftPadBytes(fallbackHandler.Bytes(), 32)...)

	// paymentToken
	encoded = append(encoded, common.LeftPadBytes(paymentToken.Bytes(), 32)...)

	// payment
	encoded = append(encoded, common.LeftPadBytes(payment.Bytes(), 32)...)

	// paymentReceiver
	encoded = append(encoded, common.LeftPadBytes(paymentReceiver.Bytes(), 32)...)

	// Encode owners array
	// Array length
	encoded = append(encoded, common.LeftPadBytes(big.NewInt(int64(len(owners))).Bytes(), 32)...)
	// Array elements
	for _, owner := range owners {
		encoded = append(encoded, common.LeftPadBytes(owner.Bytes(), 32)...)
	}

	// Encode data bytes
	// Bytes length
	encoded = append(encoded, common.LeftPadBytes(big.NewInt(int64(len(data))).Bytes(), 32)...)
	// Bytes data (padded to 32-byte boundary)
	if len(data) > 0 {
		encoded = append(encoded, data...)
		// Pad to 32-byte boundary
		padding := (32 - (len(data) % 32)) % 32
		if padding > 0 {
			encoded = append(encoded, make([]byte, padding)...)
		}
	}

	return encoded, nil
}

// DeriveSafeAddressWithNonce calculates the Safe address with a specific nonce
// This is useful for predicting Safe addresses before deployment
func DeriveSafeAddressWithNonce(signerAddress common.Address, chainID int64, nonce *big.Int) (common.Address, error) {
	// This is similar to DeriveSafeAddress but allows specifying a nonce
	// For the default case (first Safe for an address), nonce is typically 0

	// For now, we'll use the same implementation as DeriveSafeAddress
	// The nonce is implicitly 0 in the CREATE2 calculation via the salt
	return DeriveSafeAddress(signerAddress, chainID)
}

// VerifySafeAddress checks if a given address matches the derived Safe address
func VerifySafeAddress(signerAddress common.Address, expectedAddress common.Address, chainID int64) (bool, error) {
	derivedAddress, err := DeriveSafeAddress(signerAddress, chainID)
	if err != nil {
		return false, err
	}

	return derivedAddress == expectedAddress, nil
}

// GetSafeDeploymentData returns the deployment data needed for Safe creation
func GetSafeDeploymentData(signerAddress common.Address, chainID int64) (map[string]interface{}, error) {
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return nil, err
	}

	safeAddress, err := DeriveSafeAddress(signerAddress, chainID)
	if err != nil {
		return nil, err
	}

	initializer, err := buildSafeInitializer(signerAddress, contractConfig)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"safeAddress":     safeAddress.Hex(),
		"signerAddress":   signerAddress.Hex(),
		"singleton":       contractConfig.SafeSingleton,
		"factory":         contractConfig.SafeFactory,
		"fallbackHandler": contractConfig.SafeFallbackHandler,
		"initializer":     common.Bytes2Hex(initializer),
		"chainId":         chainID,
	}, nil
}
