package signer

import (
    "crypto/ecdsa"
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
    privateKey *ecdsa.PrivateKey
    chainID    *big.Int
}

func NewSigner(privateKeyHex string, chainID int64) (*Signer, error) {
    // Implementation for creating a new Signer instance
}

func (s *Signer) Address() common.Address {
    // Implementation for retrieving the address of the signer
}

func (s *Signer) GetChainID() *big.Int {
    return s.chainID
}

func (s *Signer) Sign(messageHash []byte) (string, error) {
    // Implementation for signing a message hash
}

func (s *Signer) SignEIP712StructHash(messageHash []byte) (string, error) {
    // Implementation for signing an EIP-712 struct hash
}