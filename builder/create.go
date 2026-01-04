package builder

import (
	"encoding/json"
	"math/big"

	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/davidt58/go-builder-relayer-client/constants"
	"github.com/davidt58/go-builder-relayer-client/errors"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// CreateSafeCreateStructHash builds the EIP-712 struct hash for Safe proxy creation
func CreateSafeCreateStructHash(args *models.SafeCreateTransactionArgs, sig *signer.Signer, chainID int64) (common.Hash, error) {
	// Get contract configuration
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return common.Hash{}, err
	}

	// Build the initializer data
	signerAddress := common.HexToAddress(args.SignerAddress)
	initializer, err := buildSafeInitializer(signerAddress, contractConfig)
	if err != nil {
		return common.Hash{}, err
	}

	// Parse nonce (saltNonce)
	saltNonce := new(big.Int)
	if args.Nonce != "" {
		saltNonce.SetString(args.Nonce, 0)
	}

	// Build CreateProxy struct
	createProxy := &CreateProxy{
		Singleton:   common.HexToAddress(contractConfig.SafeSingleton),
		Initializer: initializer,
		SaltNonce:   saltNonce,
	}

	// Get verifying contract (the Safe Factory)
	verifyingContract := common.HexToAddress(contractConfig.SafeFactory)

	// Build and return the hash
	return BuildCreateProxyHash(createProxy, verifyingContract, chainID)
}

// CreateSafeCreateSignature signs a Safe creation transaction and returns the signature
func CreateSafeCreateSignature(args *models.SafeCreateTransactionArgs, sig *signer.Signer, chainID int64) (string, error) {
	// Create the struct hash
	structHash, err := CreateSafeCreateStructHash(args, sig, chainID)
	if err != nil {
		return "", err
	}

	// Sign the struct hash using standard signing (not EIP-712)
	// For SAFE-CREATE, we use regular Sign method
	signature, err := sig.Sign(structHash.Bytes())
	if err != nil {
		return "", err
	}

	return signature, nil
}

// BuildSafeCreateTransactionRequest builds a complete Safe creation transaction request
// This is the main function to use when deploying a new Safe wallet
func BuildSafeCreateTransactionRequest(args *models.SafeCreateTransactionArgs, sig *signer.Signer, chainID int64) (*models.TransactionRequest, error) {
	if args == nil {
		return nil, errors.ErrMissingRequiredField("args")
	}
	if sig == nil {
		return nil, errors.ErrSignerNotConfigured
	}

	// Get contract configuration
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return nil, err
	}

	// Create signature
	signature, err := CreateSafeCreateSignature(args, sig, chainID)
	if err != nil {
		return nil, err
	}

	// Split and pack the signature
	packedSig, err := SplitAndPackSig(signature)
	if err != nil {
		return nil, err
	}

	// Build the initializer data
	signerAddress := common.HexToAddress(args.SignerAddress)
	initializer, err := buildSafeInitializer(signerAddress, contractConfig)
	if err != nil {
		return nil, err
	}

	// The "to" for Safe creation is the factory address
	to := contractConfig.SafeFactory

	// The "value" is always 0 for Safe creation
	value := "0"

	// The "data" is the encoded initializer
	data := hexutil.Encode(initializer)

	// Operation is always Call (0) for Safe creation
	operation := int(models.Call)

	// Marshal the to, value, data, operation fields
	toJSON, err := json.Marshal(to)
	if err != nil {
		return nil, errors.ErrJSONMarshalFailed(err)
	}
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return nil, errors.ErrJSONMarshalFailed(err)
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, errors.ErrJSONMarshalFailed(err)
	}
	operationJSON, err := json.Marshal(operation)
	if err != nil {
		return nil, errors.ErrJSONMarshalFailed(err)
	}

	// Create signature object
	sigObj := models.Signature{
		Signer: sig.AddressHex(),
		Data:   packedSig,
	}

	// Create the request
	request := &models.TransactionRequest{
		Type:           string(models.SAFE_CREATE),
		SafeAddress:    args.SafeAddress,
		To:             toJSON,
		Value:          valueJSON,
		Data:           dataJSON,
		Operation:      operationJSON,
		Signatures:     []models.Signature{sigObj},
		GasPrice:       "0",
		SafeTxGas:      "0",
		BaseGas:        "0",
		GasToken:       constants.ZERO_ADDRESS,
		RefundReceiver: constants.ZERO_ADDRESS,
		Nonce:          args.Nonce,
		ChainID:        chainID,
	}

	// Add metadata if provided
	if args.Metadata != "" {
		request.Metadata = &args.Metadata
	}

	return request, nil
}

// GetSafeCreationData returns the data needed for Safe creation
// This is a helper function that can be used to inspect the creation parameters
func GetSafeCreationData(signerAddress common.Address, chainID int64) (map[string]interface{}, error) {
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return nil, err
	}

	initializer, err := buildSafeInitializer(signerAddress, contractConfig)
	if err != nil {
		return nil, err
	}

	safeAddress, err := DeriveSafeAddress(signerAddress, chainID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"signerAddress":   signerAddress.Hex(),
		"safeAddress":     safeAddress.Hex(),
		"factory":         contractConfig.SafeFactory,
		"singleton":       contractConfig.SafeSingleton,
		"fallbackHandler": contractConfig.SafeFallbackHandler,
		"initializer":     hexutil.Encode(initializer),
		"chainId":         chainID,
	}, nil
}

// VerifySafeCreationSignature verifies a Safe creation signature
// This is useful for testing and debugging
func VerifySafeCreationSignature(args *models.SafeCreateTransactionArgs, sig *signer.Signer, signature string, chainID int64) (bool, error) {
	// Create the struct hash
	structHash, err := CreateSafeCreateStructHash(args, sig, chainID)
	if err != nil {
		return false, err
	}

	// Verify the signature
	return sig.VerifySignature(structHash.Bytes(), signature)
}
