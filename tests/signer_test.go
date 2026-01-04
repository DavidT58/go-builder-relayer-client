package tests

import (
	"math/big"
	"testing"

	"github.com/davidt58/go-builder-relayer-client/signer"
	"github.com/ethereum/go-ethereum/common"
)

func TestSignerAddress(t *testing.T) {
	pk := "0x1234..." // Replace with a valid private key
	signer, err := signer.NewSigner(pk, 80002)
	if err != nil {
		t.Fatal(err)
	}

	addr := signer.Address()
	expected := common.HexToAddress("0xabcd...") // Replace with the expected address

	if addr != expected {
		t.Errorf("expected %s, got %s", expected.Hex(), addr.Hex())
	}
}

func TestSignMessage(t *testing.T) {
	pk := "0x1234..." // Replace with a valid private key
	signer, err := signer.NewSigner(pk, 80002)
	if err != nil {
		t.Fatal(err)
	}

	messageHash := []byte("test message")
	signature, err := signer.Sign(messageHash)
	if err != nil {
		t.Fatal(err)
	}

	if signature == "" {
		t.Error("expected a signature, got empty string")
	}
}

func TestSignEIP712StructHash(t *testing.T) {
	pk := "0x1234..." // Replace with a valid private key
	signer, err := signer.NewSigner(pk, 80002)
	if err != nil {
		t.Fatal(err)
	}

	messageHash := []byte("test EIP-712 message")
	signature, err := signer.SignEIP712StructHash(messageHash)
	if err != nil {
		t.Fatal(err)
	}

	if signature == "" {
		t.Error("expected a signature, got empty string")
	}
}

func TestGetChainID(t *testing.T) {
	pk := "0x1234..." // Replace with a valid private key
	signer, err := signer.NewSigner(pk, 80002)
	if err != nil {
		t.Fatal(err)
	}

	chainID := signer.GetChainID()
	expectedChainID := big.NewInt(80002)

	if chainID.Cmp(expectedChainID) != 0 {
		t.Errorf("expected chain ID %d, got %d", expectedChainID, chainID)
	}
}
