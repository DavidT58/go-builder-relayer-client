package builder

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/davidt58/go-builder-relayer-client/constants"
	"github.com/davidt58/go-builder-relayer-client/errors"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// CreateSafeMultisendTransaction encodes multiple transactions into a single multisend transaction
// This follows the Safe MultiSend contract encoding format:
// - Each transaction is encoded as: uint8(operation) ++ address(to) ++ uint256(value) ++ uint256(dataLength) ++ bytes(data)
// - All transactions are concatenated
// - The result is wrapped with the multisend function selector
func CreateSafeMultisendTransaction(transactions []models.SafeTransaction, multiSendAddress string) (*models.SafeTransaction, error) {
	if len(transactions) == 0 {
		return nil, errors.NewRelayerClientError("no transactions to encode", nil)
	}

	// Encode all transactions using packed encoding
	var encodedTxns bytes.Buffer

	for _, txn := range transactions {
		// Encode each transaction in the format:
		// operation (uint8, 1 byte)
		// to (address, 20 bytes)
		// value (uint256, 32 bytes)
		// dataLength (uint256, 32 bytes)
		// data (bytes, variable length)

		// Operation (1 byte)
		encodedTxns.WriteByte(byte(txn.Operation))

		// To address (20 bytes)
		toAddr := common.HexToAddress(txn.To)
		encodedTxns.Write(toAddr.Bytes())

		// Value (32 bytes)
		value := new(big.Int)
		if txn.Value != "" {
			value.SetString(txn.Value, 0)
		}
		valueBytes := make([]byte, 32)
		value.FillBytes(valueBytes)
		encodedTxns.Write(valueBytes)

		// Decode data
		var dataBytes []byte
		if txn.Data != "" && txn.Data != "0x" {
			var err error
			dataBytes, err = hexutil.Decode(txn.Data)
			if err != nil {
				return nil, errors.NewRelayerClientError("failed to decode transaction data", err)
			}
		}

		// Data length (32 bytes)
		dataLength := big.NewInt(int64(len(dataBytes)))
		dataLengthBytes := make([]byte, 32)
		dataLength.FillBytes(dataLengthBytes)
		encodedTxns.Write(dataLengthBytes)

		// Data (variable length)
		if len(dataBytes) > 0 {
			encodedTxns.Write(dataBytes)
		}
	}

	// Wrap with multisend function selector
	// multiSend(bytes) - selector is 0x8d80ff0a
	selector, err := hexutil.Decode(constants.MULTISEND_FUNCTION_SELECTOR)
	if err != nil {
		return nil, errors.NewRelayerClientError("invalid multisend selector", err)
	}

	// Encode the bytes parameter for multiSend(bytes)
	// ABI encoding: selector + offset (32 bytes) + length (32 bytes) + data
	var callData bytes.Buffer
	callData.Write(selector)

	// Offset to the bytes data (always 32 for a single dynamic parameter)
	offset := make([]byte, 32)
	offset[31] = 32
	callData.Write(offset)

	// Length of the encoded transactions
	length := big.NewInt(int64(encodedTxns.Len()))
	lengthBytes := make([]byte, 32)
	length.FillBytes(lengthBytes)
	callData.Write(lengthBytes)

	// Encoded transactions
	callData.Write(encodedTxns.Bytes())

	// Pad to 32-byte boundary if needed
	remainder := callData.Len() % 32
	if remainder != 0 {
		padding := make([]byte, 32-remainder)
		callData.Write(padding)
	}

	// Create the multisend transaction
	multiSendTxn := &models.SafeTransaction{
		To:        multiSendAddress,
		Value:     "0",
		Data:      hexutil.Encode(callData.Bytes()),
		Operation: models.DelegateCall, // MultiSend uses DELEGATECALL
	}

	return multiSendTxn, nil
}

// EncodeMultiSendData encodes the inner data for multisend (without function selector)
// This is a helper function that can be used independently
func EncodeMultiSendData(transactions []models.SafeTransaction) ([]byte, error) {
	var encoded bytes.Buffer

	for _, txn := range transactions {
		// Operation (1 byte)
		encoded.WriteByte(byte(txn.Operation))

		// To address (20 bytes)
		toAddr := common.HexToAddress(txn.To)
		encoded.Write(toAddr.Bytes())

		// Value (32 bytes)
		value := new(big.Int)
		if txn.Value != "" {
			value.SetString(txn.Value, 0)
		}
		valueBytes := make([]byte, 32)
		value.FillBytes(valueBytes)
		encoded.Write(valueBytes)

		// Decode data
		var dataBytes []byte
		if txn.Data != "" && txn.Data != "0x" {
			var err error
			dataBytes, err = hexutil.Decode(txn.Data)
			if err != nil {
				return nil, errors.NewRelayerClientError("failed to decode transaction data", err)
			}
		}

		// Data length (32 bytes)
		dataLength := big.NewInt(int64(len(dataBytes)))
		dataLengthBytes := make([]byte, 32)
		dataLength.FillBytes(dataLengthBytes)
		encoded.Write(dataLengthBytes)

		// Data (variable length)
		if len(dataBytes) > 0 {
			encoded.Write(dataBytes)
		}
	}

	return encoded.Bytes(), nil
}

// AggregateSafeTransaction combines multiple Safe transactions into a single multisend transaction
// This is the main function to use when you need to batch multiple transactions
func AggregateSafeTransaction(transactions []models.SafeTransaction, safeMultisend string) (*models.SafeTransaction, error) {
	if len(transactions) == 0 {
		return nil, errors.NewRelayerClientError("no transactions to aggregate", nil)
	}

	// If there's only one transaction, return it as-is
	if len(transactions) == 1 {
		return &transactions[0], nil
	}

	// Otherwise, create a multisend transaction
	return CreateSafeMultisendTransaction(transactions, safeMultisend)
}

// DecodeMultiSendData decodes multisend data back into individual transactions
// This is useful for debugging and testing
func DecodeMultiSendData(data []byte) ([]models.SafeTransaction, error) {
	if len(data) == 0 {
		return nil, errors.NewRelayerClientError("empty multisend data", nil)
	}

	var transactions []models.SafeTransaction
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		// Read operation (1 byte)
		var operation uint8
		if err := binary.Read(reader, binary.BigEndian, &operation); err != nil {
			return nil, errors.NewRelayerClientError("failed to read operation", err)
		}

		// Read to address (20 bytes)
		toBytes := make([]byte, 20)
		if _, err := reader.Read(toBytes); err != nil {
			return nil, errors.NewRelayerClientError("failed to read to address", err)
		}
		to := common.BytesToAddress(toBytes)

		// Read value (32 bytes)
		valueBytes := make([]byte, 32)
		if _, err := reader.Read(valueBytes); err != nil {
			return nil, errors.NewRelayerClientError("failed to read value", err)
		}
		value := new(big.Int).SetBytes(valueBytes)

		// Read data length (32 bytes)
		dataLengthBytes := make([]byte, 32)
		if _, err := reader.Read(dataLengthBytes); err != nil {
			return nil, errors.NewRelayerClientError("failed to read data length", err)
		}
		dataLength := new(big.Int).SetBytes(dataLengthBytes)

		// Read data
		var txnData []byte
		if dataLength.Cmp(big.NewInt(0)) > 0 {
			txnData = make([]byte, dataLength.Int64())
			if _, err := reader.Read(txnData); err != nil {
				return nil, errors.NewRelayerClientError("failed to read data", err)
			}
		}

		txn := models.SafeTransaction{
			To:        to.Hex(),
			Value:     value.String(),
			Data:      hexutil.Encode(txnData),
			Operation: models.OperationType(operation),
		}

		transactions = append(transactions, txn)
	}

	return transactions, nil
}

// ComputeMultiSendHash computes the hash of a multisend transaction
// This is useful for verification and debugging
func ComputeMultiSendHash(transactions []models.SafeTransaction) (common.Hash, error) {
	encoded, err := EncodeMultiSendData(transactions)
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(encoded), nil
}
