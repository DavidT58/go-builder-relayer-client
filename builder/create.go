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
// Matches the Python implementation using payment fields
func CreateSafeCreateStructHash(args *models.SafeCreateTransactionArgs, sig *signer.Signer, chainID int64) (common.Hash, error) {
	// Get contract configuration
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return common.Hash{}, err
	}

	// For SAFE-CREATE, we use payment fields (all zeros/constants)
	// This matches the Python implementation
	paymentToken := common.HexToAddress(constants.ZERO_ADDRESS)
	payment := big.NewInt(0)
	paymentReceiver := common.HexToAddress(constants.ZERO_ADDRESS)

	// Build CreateProxy struct with payment fields
	createProxy := &CreateProxy{
		PaymentToken:    paymentToken,
		Payment:         payment,
		PaymentReceiver: paymentReceiver,
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

	// The "to" for Safe creation is the factory address
	to := contractConfig.SafeFactory

	// The "data" is always "0x" for SAFE-CREATE (matching Python implementation)
	data := "0x"

	// Marshal to and data as JSON strings
	toJSON, err := json.Marshal(to)
	if err != nil {
		return nil, errors.ErrJSONMarshalFailed(err)
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, errors.ErrJSONMarshalFailed(err)
	}

	// For SAFE-CREATE, signature params contain payment info (all zeros/constants)
	paymentToken := constants.ZERO_ADDRESS
	payment := "0"
	paymentReceiver := constants.ZERO_ADDRESS

	signatureParams := &models.SignatureParams{
		PaymentToken:    &paymentToken,
		Payment:         &payment,
		PaymentReceiver: &paymentReceiver,
	}

	// Create the request (matching Python structure)
	request := &models.TransactionRequest{
		Type:            string(models.SAFE_CREATE),
		From:            args.SignerAddress,
		To:              toJSON,
		ProxyWallet:     args.SafeAddress,
		Data:            dataJSON,
		Signature:       signature,
		SignatureParams: signatureParams,
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
