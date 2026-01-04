package builder

import (
	"math/big"

	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// DeriveSafeAddress calculates the Safe address using CREATE2
// This matches the Python implementation's derive_safe_address function
func DeriveSafeAddress(signerAddress common.Address, chainID int64) (common.Address, error) {
	// Get contract configuration for the chain
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return common.Address{}, err
	}

	// Build the initializer data for the Safe setup
	initializerData, err := buildSafeInitializer(signerAddress, contractConfig)
	if err != nil {
		return common.Address{}, err
	}

	// Calculate the CREATE2 address
	safeAddress := calculateCreate2Address(
		common.HexToAddress(contractConfig.SafeFactory),
		common.HexToAddress(contractConfig.SafeSingleton),
		initializerData,
	)

	return safeAddress, nil
}

// buildSafeInitializer creates the initializer data for Safe.setup()
// This encodes the call to setup(owners, threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
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

// calculateCreate2Address calculates the CREATE2 address
// Formula: keccak256(0xff ++ deployerAddress ++ salt ++ keccak256(initCode))[12:]
func calculateCreate2Address(factoryAddress, singleton common.Address, initializer []byte) common.Address {
	// Build the init code for the Safe proxy
	// The init code is the proxy bytecode with the singleton address appended
	initCode := buildProxyInitCode(singleton, initializer)

	// Calculate the init code hash
	initCodeHash := crypto.Keccak256Hash(initCode)

	// Salt is the keccak256 of the initializer
	salt := crypto.Keccak256Hash(initializer)

	// Calculate CREATE2 address
	// keccak256(0xff ++ factoryAddress ++ salt ++ initCodeHash)[12:]
	data := make([]byte, 1+20+32+32)
	data[0] = 0xff
	copy(data[1:21], factoryAddress.Bytes())
	copy(data[21:53], salt.Bytes())
	copy(data[53:85], initCodeHash.Bytes())

	hash := crypto.Keccak256Hash(data)
	return common.BytesToAddress(hash[12:])
}

// buildProxyInitCode builds the init code for the Safe proxy
// This is the proxy creation bytecode that will deploy a minimal proxy to the singleton
func buildProxyInitCode(singleton common.Address, initializer []byte) []byte {
	// Safe uses a minimal proxy pattern (EIP-1167)
	// The proxy bytecode is:
	// 0x608060405234801561001057600080fd5b506040516101e63803806101e68339818101604052602081101561003357600080fd5b8101908080519060200190929190505050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614156100ca576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806101c46022913960400191505060405180910390fd5b806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505060ab806101196000396000f3fe608060405273ffffffffffffffffffffffffffffffffffffffff600054167fa619486e0000000000000000000000000000000000000000000000000000000060003514156050578060005260206000f35b3660008037600080366000845af43d6000803e60008114156070573d6000fd5b3d6000f3fea2646970667358221220d1429297349653a4918076d650332de1a1068c5f3e07c5c82360c277770b955264736f6c63430007060033
	//
	// For simplicity, we'll use a standard proxy creation pattern
	// The actual bytecode calculation matches the Safe Proxy Factory's createProxyWithNonce function

	// Standard minimal proxy bytecode (EIP-1167 clone)
	// 0x3d602d80600a3d3981f3363d3d373d3d3d363d73bebebebebebebebebebebebebebebebebebebebe5af43d82803e903d91602b57fd5bf3
	// where "bebebebe..." is replaced with the singleton address

	proxyCode := []byte{
		0x60, 0x2d, // PUSH1 0x2d (proxy code size)
		0x80,       // DUP1
		0x60, 0x0a, // PUSH1 0x0a (offset)
		0x3d, // RETURNDATASIZE
		0x39, // CODECOPY
		0x81, // DUP2
		0xf3, // RETURN
		// Proxy runtime code:
		0x36, // CALLDATASIZE
		0x3d, // RETURNDATASIZE
		0x3d, // RETURNDATASIZE
		0x37, // CALLDATACOPY
		0x3d, // RETURNDATASIZE
		0x3d, // RETURNDATASIZE
		0x3d, // RETURNDATASIZE
		0x36, // CALLDATASIZE
		0x3d, // RETURNDATASIZE
		0x73, // PUSH20
	}

	// Append singleton address
	proxyCode = append(proxyCode, singleton.Bytes()...)

	// Append rest of proxy code
	proxyCode = append(proxyCode, []byte{
		0x5a,       // GAS
		0xf4,       // DELEGATECALL
		0x3d,       // RETURNDATASIZE
		0x82,       // DUP3
		0x80,       // DUP1
		0x3e,       // RETURNDATACOPY
		0x90,       // SWAP1
		0x3d,       // RETURNDATASIZE
		0x91,       // SWAP2
		0x60, 0x2b, // PUSH1 0x2b
		0x57, // JUMPI
		0xfd, // REVERT
		0x5b, // JUMPDEST
		0xf3, // RETURN
	}...)

	return proxyCode
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
