package signer

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/davidt58/go-builder-relayer-client/errors"
)

// Signer handles cryptographic signing operations for Ethereum transactions
type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    *big.Int
}

// NewSigner creates a new Signer from a private key hex string
// privateKeyHex should not include the "0x" prefix
func NewSigner(privateKeyHex string, chainID int64) (*Signer, error) {
	// Remove "0x" prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Parse the private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, errors.ErrInvalidPrivateKey(err)
	}

	// Derive the address from the private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.ErrInvalidPrivateKey(fmt.Errorf("failed to get public key"))
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Signer{
		privateKey: privateKey,
		address:    address,
		chainID:    big.NewInt(chainID),
	}, nil
}

// Address returns the Ethereum address associated with the signer's private key
func (s *Signer) Address() common.Address {
	return s.address
}

// AddressHex returns the Ethereum address as a hex string with "0x" prefix
func (s *Signer) AddressHex() string {
	return s.address.Hex()
}

// GetChainID returns the chain ID
func (s *Signer) GetChainID() *big.Int {
	return new(big.Int).Set(s.chainID)
}

// Sign signs a message hash using EIP-191 personal sign format
// messageHash should be the 32-byte hash of the message
// Returns the signature as a hex string with "0x" prefix
func (s *Signer) Sign(messageHash []byte) (string, error) {
	if len(messageHash) != 32 {
		return "", errors.NewRelayerClientError("message hash must be 32 bytes", nil)
	}

	// Sign using Ethereum's personal sign format
	// This adds the "\x19Ethereum Signed Message:\n32" prefix
	signature, err := crypto.Sign(messageHash, s.privateKey)
	if err != nil {
		return "", errors.ErrSigningFailed(err)
	}

	// Adjust V value for Ethereum (add 27)
	if signature[64] < 27 {
		signature[64] += 27
	}

	return hexutil.Encode(signature), nil
}

// SignEIP712StructHash signs an EIP-712 struct hash
// This is used for signing structured data like Safe transactions
// messageHash should be the 32-byte hash of the EIP-712 typed data
// Returns the signature as a hex string with "0x" prefix
// NOTE: This applies EIP-191 prefix to match Python implementation behavior
func (s *Signer) SignEIP712StructHash(messageHash []byte) (string, error) {
	if len(messageHash) != 32 {
		return "", errors.NewRelayerClientError("message hash must be 32 bytes", nil)
	}

	// Apply EIP-191 prefix: "\x19Ethereum Signed Message:\n32" + messageHash
	// This matches the Python implementation's encode_defunct behavior
	prefixedHash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(messageHash))),
		messageHash,
	)

	// Sign the prefixed hash
	signature, err := crypto.Sign(prefixedHash.Bytes(), s.privateKey)
	if err != nil {
		return "", errors.ErrSigningFailed(err)
	}

	// Adjust V value for Ethereum (add 27)
	if signature[64] < 27 {
		signature[64] += 27
	}

	return hexutil.Encode(signature), nil
}

// SignMessage signs an arbitrary message using EIP-191 personal sign
// The message will be prefixed with "\x19Ethereum Signed Message:\n{length}"
func (s *Signer) SignMessage(message []byte) (string, error) {
	// Create the hash with EIP-191 prefix
	hash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))),
		message,
	)

	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", errors.ErrSigningFailed(err)
	}

	// Adjust V value for Ethereum (add 27)
	if signature[64] < 27 {
		signature[64] += 27
	}

	return hexutil.Encode(signature), nil
}

// RecoverAddress recovers the Ethereum address from a signature
// This can be used to verify signatures
func RecoverAddress(messageHash []byte, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, errors.ErrInvalidSignature(fmt.Errorf("signature must be 65 bytes"))
	}

	// Adjust V value (subtract 27 if needed)
	v := signature[64]
	if v >= 27 {
		signature[64] -= 27
	}

	// Recover the public key
	pubKey, err := crypto.SigToPub(messageHash, signature)
	if err != nil {
		return common.Address{}, errors.ErrInvalidSignature(err)
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}

// VerifySignature verifies that a signature was created by this signer
func (s *Signer) VerifySignature(messageHash []byte, signatureHex string) (bool, error) {
	// Decode the signature
	signature, err := hexutil.Decode(signatureHex)
	if err != nil {
		return false, errors.ErrInvalidSignature(err)
	}

	// Recover the address
	recoveredAddr, err := RecoverAddress(messageHash, signature)
	if err != nil {
		return false, err
	}

	// Compare with signer's address
	return recoveredAddr == s.address, nil
}

// SplitSignature splits a signature into r, s, v components
func SplitSignature(signatureHex string) (r, s string, v int, err error) {
	// Remove "0x" prefix if present
	signatureHex = strings.TrimPrefix(signatureHex, "0x")

	// Decode the signature
	signature, err := hexutil.Decode("0x" + signatureHex)
	if err != nil {
		return "", "", 0, errors.ErrInvalidSignature(err)
	}

	if len(signature) != 65 {
		return "", "", 0, errors.ErrInvalidSignature(fmt.Errorf("signature must be 65 bytes"))
	}

	// Extract r, s, v components
	r = "0x" + hexutil.Encode(signature[0:32])[2:]
	s = "0x" + hexutil.Encode(signature[32:64])[2:]
	v = int(signature[64])

	return r, s, v, nil
}

// PackSignatures packs multiple signatures into a single byte array
// This is used for Safe multi-signature transactions
func PackSignatures(signatures []string) (string, error) {
	if len(signatures) == 0 {
		return "", errors.NewRelayerClientError("no signatures provided", nil)
	}

	var packed []byte

	for _, sig := range signatures {
		// Remove "0x" prefix if present
		sig = strings.TrimPrefix(sig, "0x")

		// Decode the signature
		sigBytes, err := hexutil.Decode("0x" + sig)
		if err != nil {
			return "", errors.ErrInvalidSignature(err)
		}

		if len(sigBytes) != 65 {
			return "", errors.ErrInvalidSignature(fmt.Errorf("signature must be 65 bytes"))
		}

		packed = append(packed, sigBytes...)
	}

	return hexutil.Encode(packed), nil
}

// Keccak256 computes the Keccak256 hash of the input data
func Keccak256(data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}

// Keccak256Hash computes the Keccak256 hash and returns a common.Hash
func Keccak256Hash(data ...[]byte) common.Hash {
	return crypto.Keccak256Hash(data...)
}
