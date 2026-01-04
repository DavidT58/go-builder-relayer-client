package builder

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/davidt58/go-builder-relayer-client/constants"
	"github.com/davidt58/go-builder-relayer-client/errors"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// SplitSignature splits a signature into r, s, v components
// signatureHex should be a 65-byte hex string (with or without 0x prefix)
// Returns r, s as hex strings with 0x prefix, and v as an integer
func SplitSignature(signatureHex string) (r, s string, v int, err error) {
	// Remove 0x prefix if present
	signatureHex = strings.TrimPrefix(signatureHex, "0x")

	// Decode the signature
	signature, err := hexutil.Decode("0x" + signatureHex)
	if err != nil {
		return "", "", 0, errors.ErrInvalidSignature(err)
	}

	if len(signature) != 65 {
		return "", "", 0, errors.ErrInvalidSignature(errors.NewRelayerClientError("signature must be 65 bytes", nil))
	}

	// Extract r, s, v components
	r = hexutil.Encode(signature[0:32])
	s = hexutil.Encode(signature[32:64])
	v = int(signature[64])

	// Adjust v value: convert 27,28 or 0,1 to 31,32 for Safe
	// Safe uses v = 31 or 32 for standard ECDSA signatures
	if v == 27 || v == 0 {
		v = 31
	} else if v == 28 || v == 1 {
		v = 32
	} else if v != 31 && v != 32 {
		// If v is already in a different format, try to normalize it
		if v < 27 {
			v += 31
		} else if v >= 27 && v <= 28 {
			v = v - 27 + 31
		}
	}

	return r, s, v, nil
}

// SplitAndPackSig splits a signature and packs it using eth_abi packed encoding
// The packed format is: encode_packed(["uint256", "uint256", "uint8"], [r, s, v])
// This returns a hex string with 0x prefix containing the 65 packed bytes
func SplitAndPackSig(signatureHex string) (string, error) {
	r, s, v, err := SplitSignature(signatureHex)
	if err != nil {
		return "", err
	}

	// Decode r and s
	rBytes, err := hexutil.Decode(r)
	if err != nil {
		return "", errors.ErrInvalidSignature(err)
	}
	sBytes, err := hexutil.Decode(s)
	if err != nil {
		return "", errors.ErrInvalidSignature(err)
	}

	// Pack: r (32 bytes) + s (32 bytes) + v (1 byte)
	packed := make([]byte, 65)
	copy(packed[0:32], rBytes)
	copy(packed[32:64], sBytes)
	packed[64] = byte(v)

	return hexutil.Encode(packed), nil
}

// CreateSafeStructHash builds the EIP-712 struct hash for a Safe transaction
// Note: This function only handles single transactions. For multiple transactions,
// use BuildSafeTransactionRequestWithMultisend which aggregates them first.
func CreateSafeStructHash(args *models.SafeTransactionArgs, sig *signer.Signer) (common.Hash, error) {
	// Get the transaction data
	var to common.Address
	var value *big.Int
	var data []byte
	var operation uint8

	if len(args.Transactions) == 0 {
		return common.Hash{}, errors.NewRelayerClientError("no transactions provided", nil)
	}

	if len(args.Transactions) > 1 {
		return common.Hash{}, errors.NewRelayerClientError("CreateSafeStructHash only supports single transactions; use BuildSafeTransactionRequestWithMultisend for multiple transactions", nil)
	}

	// Single transaction
	txn := args.Transactions[0]
	to = common.HexToAddress(txn.To)
	value = new(big.Int)
	if txn.Value != "" {
		value.SetString(txn.Value, 0)
	}

	if txn.Data != "" && txn.Data != "0x" {
		var err error
		data, err = hexutil.Decode(txn.Data)
		if err != nil {
			return common.Hash{}, errors.NewRelayerClientError("failed to decode transaction data", err)
		}
	}

	operation = uint8(txn.Operation)

	// Parse nonce
	nonce := new(big.Int)
	if args.Nonce != "" {
		nonce.SetString(args.Nonce, 0)
	}

	// Build SafeTx struct
	safeTx := &SafeTx{
		To:             to,
		Value:          value,
		Data:           data,
		Operation:      operation,
		SafeTxGas:      big.NewInt(0),
		BaseGas:        big.NewInt(0),
		GasPrice:       big.NewInt(0),
		GasToken:       common.HexToAddress(constants.ZERO_ADDRESS),
		RefundReceiver: common.HexToAddress(constants.ZERO_ADDRESS),
		Nonce:          nonce,
	}

	// Get verifying contract (the Safe address)
	verifyingContract := common.HexToAddress(args.SafeAddress)

	// Get chain ID from signer
	chainID := sig.GetChainID().Int64()

	// Build and return the hash
	return BuildSafeTxHash(safeTx, verifyingContract, chainID)
}

// CreateSafeSignature signs a Safe transaction and returns the signature
func CreateSafeSignature(args *models.SafeTransactionArgs, sig *signer.Signer) (string, error) {
	// Create the struct hash
	structHash, err := CreateSafeStructHash(args, sig)
	if err != nil {
		return "", err
	}

	// Sign the struct hash using EIP-712 signing
	signature, err := sig.SignEIP712StructHash(structHash.Bytes())
	if err != nil {
		return "", err
	}

	return signature, nil
}

// BuildSafeTransactionRequest builds a complete Safe transaction request
// This is the main function to use when preparing a Safe transaction for submission
func BuildSafeTransactionRequest(args *models.SafeTransactionArgs, sig *signer.Signer, chainID int64) (*models.TransactionRequest, error) {
	if args == nil {
		return nil, errors.ErrMissingRequiredField("args")
	}
	if sig == nil {
		return nil, errors.ErrSignerNotConfigured
	}

	// Create signature
	signature, err := CreateSafeSignature(args, sig)
	if err != nil {
		return nil, err
	}

	// Split and pack the signature
	packedSig, err := SplitAndPackSig(signature)
	if err != nil {
		return nil, err
	}

	// Build the transaction request
	var to, value, data, operation interface{}

	if len(args.Transactions) == 1 {
		// Single transaction
		txn := args.Transactions[0]
		to = txn.To
		value = txn.Value
		data = txn.Data
		operation = int(txn.Operation)
	} else {
		// Multiple transactions - need arrays
		tos := make([]string, len(args.Transactions))
		values := make([]string, len(args.Transactions))
		datas := make([]string, len(args.Transactions))
		operations := make([]int, len(args.Transactions))

		for i, txn := range args.Transactions {
			tos[i] = txn.To
			values[i] = txn.Value
			datas[i] = txn.Data
			operations[i] = int(txn.Operation)
		}

		to = tos
		value = values
		data = datas
		operation = operations
	}

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
		Type:           string(models.SAFE),
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

// BuildSafeTransactionRequestWithMultisend builds a Safe transaction request with multisend
// This should be used when you have multiple transactions to batch
func BuildSafeTransactionRequestWithMultisend(args *models.SafeTransactionArgs, sig *signer.Signer, chainID int64, multisendAddress string) (*models.TransactionRequest, error) {
	if len(args.Transactions) <= 1 {
		// No need for multisend with single transaction
		return BuildSafeTransactionRequest(args, sig, chainID)
	}

	// Aggregate transactions into a multisend
	multiSendTxn, err := AggregateSafeTransaction(args.Transactions, multisendAddress)
	if err != nil {
		return nil, err
	}

	// Create new args with the multisend transaction
	multiSendArgs := &models.SafeTransactionArgs{
		SafeAddress:  args.SafeAddress,
		Transactions: []models.SafeTransaction{*multiSendTxn},
		Nonce:        args.Nonce,
		Metadata:     args.Metadata,
	}

	return BuildSafeTransactionRequest(multiSendArgs, sig, chainID)
}
