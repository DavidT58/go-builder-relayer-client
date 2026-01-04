package builder

import (
	"math/big"

	"github.com/davidt58/go-builder-relayer-client/constants"
	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// SafeTx represents the EIP-712 SafeTx typed data structure
// This matches the Safe contract's SafeTx structure
type SafeTx struct {
	To             common.Address
	Value          *big.Int
	Data           []byte
	Operation      uint8
	SafeTxGas      *big.Int
	BaseGas        *big.Int
	GasPrice       *big.Int
	GasToken       common.Address
	RefundReceiver common.Address
	Nonce          *big.Int
}

// CreateProxy represents the EIP-712 CreateProxy typed data structure
// This is used for Safe wallet creation via the proxy factory
type CreateProxy struct {
	Singleton      common.Address
	Initializer    []byte
	SaltNonce      *big.Int
}

// BuildSafeTxHash builds the EIP-712 hash for a Safe transaction
// This follows the EIP-712 standard for typed data hashing
func BuildSafeTxHash(safeTx *SafeTx, verifyingContract common.Address, chainID int64) (common.Hash, error) {
	// Build the EIP-712 typed data
	typedData := &signer.TypedData{
		Types: map[string][]signer.EIP712Type{
			"EIP712Domain": {
				{Name: "verifyingContract", Type: "address"},
			},
			"SafeTx": {
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "data", Type: "bytes"},
				{Name: "operation", Type: "uint8"},
				{Name: "safeTxGas", Type: "uint256"},
				{Name: "baseGas", Type: "uint256"},
				{Name: "gasPrice", Type: "uint256"},
				{Name: "gasToken", Type: "address"},
				{Name: "refundReceiver", Type: "address"},
				{Name: "nonce", Type: "uint256"},
			},
		},
		PrimaryType: "SafeTx",
		Domain: signer.EIP712Domain{
			VerifyingContract: verifyingContract,
		},
		Message: map[string]interface{}{
			"to":             safeTx.To.Hex(),
			"value":          safeTx.Value.String(),
			"data":           common.Bytes2Hex(safeTx.Data),
			"operation":      int(safeTx.Operation),
			"safeTxGas":      safeTx.SafeTxGas.String(),
			"baseGas":        safeTx.BaseGas.String(),
			"gasPrice":       safeTx.GasPrice.String(),
			"gasToken":       safeTx.GasToken.Hex(),
			"refundReceiver": safeTx.RefundReceiver.Hex(),
			"nonce":          safeTx.Nonce.String(),
		},
	}

	// Hash the typed data
	return signer.HashTypedData(typedData)
}

// BuildCreateProxyHash builds the EIP-712 hash for Safe proxy creation
// This is used when deploying a new Safe wallet
func BuildCreateProxyHash(createProxy *CreateProxy, verifyingContract common.Address, chainID int64) (common.Hash, error) {
	// Build the EIP-712 typed data
	typedData := &signer.TypedData{
		Types: map[string][]signer.EIP712Type{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"CreateProxy": {
				{Name: "singleton", Type: "address"},
				{Name: "initializer", Type: "bytes"},
				{Name: "saltNonce", Type: "uint256"},
			},
		},
		PrimaryType: "CreateProxy",
		Domain: signer.EIP712Domain{
			Name:              constants.SAFE_FACTORY_NAME,
			ChainId:           big.NewInt(chainID),
			VerifyingContract: verifyingContract,
		},
		Message: map[string]interface{}{
			"singleton":   createProxy.Singleton.Hex(),
			"initializer": common.Bytes2Hex(createProxy.Initializer),
			"saltNonce":   createProxy.SaltNonce.String(),
		},
	}

	// Hash the typed data
	return signer.HashTypedData(typedData)
}

// ComputeSafeTxHash is a helper function that creates a SafeTx struct and computes its hash
func ComputeSafeTxHash(
	to common.Address,
	value *big.Int,
	data []byte,
	operation uint8,
	safeTxGas *big.Int,
	baseGas *big.Int,
	gasPrice *big.Int,
	gasToken common.Address,
	refundReceiver common.Address,
	nonce *big.Int,
	verifyingContract common.Address,
	chainID int64,
) (common.Hash, error) {
	safeTx := &SafeTx{
		To:             to,
		Value:          value,
		Data:           data,
		Operation:      operation,
		SafeTxGas:      safeTxGas,
		BaseGas:        baseGas,
		GasPrice:       gasPrice,
		GasToken:       gasToken,
		RefundReceiver: refundReceiver,
		Nonce:          nonce,
	}

	return BuildSafeTxHash(safeTx, verifyingContract, chainID)
}

// ComputeCreateProxyHash is a helper function that creates a CreateProxy struct and computes its hash
func ComputeCreateProxyHash(
	singleton common.Address,
	initializer []byte,
	saltNonce *big.Int,
	verifyingContract common.Address,
	chainID int64,
) (common.Hash, error) {
	createProxy := &CreateProxy{
		Singleton:   singleton,
		Initializer: initializer,
		SaltNonce:   saltNonce,
	}

	return BuildCreateProxyHash(createProxy, verifyingContract, chainID)
}

// GetSafeTxTypeHash returns the type hash for SafeTx
// This is keccak256("SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)")
func GetSafeTxTypeHash() common.Hash {
	typeString := "SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)"
	return crypto.Keccak256Hash([]byte(typeString))
}

// GetCreateProxyTypeHash returns the type hash for CreateProxy
// This is keccak256("CreateProxy(address singleton,bytes initializer,uint256 saltNonce)")
func GetCreateProxyTypeHash() common.Hash {
	typeString := "CreateProxy(address singleton,bytes initializer,uint256 saltNonce)"
	return crypto.Keccak256Hash([]byte(typeString))
}

// GetDomainSeparator computes the EIP-712 domain separator
func GetDomainSeparator(name string, chainID int64, verifyingContract common.Address) common.Hash {
	typeString := "EIP712Domain(string name,uint256 chainId,address verifyingContract)"
	typeHash := crypto.Keccak256Hash([]byte(typeString))

	nameHash := crypto.Keccak256Hash([]byte(name))
	chainIDBig := big.NewInt(chainID)
	chainIDBytes := make([]byte, 32)
	chainIDBig.FillBytes(chainIDBytes)

	verifyingContractBytes := make([]byte, 32)
	copy(verifyingContractBytes[12:], verifyingContract.Bytes())

	// Concatenate: typeHash + nameHash + chainID + verifyingContract
	var data []byte
	data = append(data, typeHash[:]...)
	data = append(data, nameHash[:]...)
	data = append(data, chainIDBytes...)
	data = append(data, verifyingContractBytes...)

	return crypto.Keccak256Hash(data)
}
