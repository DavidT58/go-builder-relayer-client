package models

import (
	"encoding/json"
	"testing"
)

func TestNewSignatureParams(t *testing.T) {
	nonce := "5"
	params := NewSignatureParams(nonce)

	if params.Nonce != nonce {
		t.Errorf("Nonce = %s, want %s", params.Nonce, nonce)
	}
	if params.GasPrice != "0" {
		t.Errorf("GasPrice = %s, want 0", params.GasPrice)
	}
	if params.Operation != Call {
		t.Errorf("Operation = %v, want %v", params.Operation, Call)
	}
}

func TestSplitSig_JSON(t *testing.T) {
	sig := SplitSig{
		R: "0x1234",
		S: "0x5678",
		V: 27,
	}

	data, err := json.Marshal(sig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded SplitSig
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.R != sig.R || decoded.S != sig.S || decoded.V != sig.V {
		t.Errorf("Decoded signature mismatch")
	}
}

func TestNewSignature(t *testing.T) {
	signer := "0x1234567890123456789012345678901234567890"
	data := "0xabcdef"

	sig := NewSignature(signer, data)

	if sig.Signer != signer {
		t.Errorf("Signer = %s, want %s", sig.Signer, signer)
	}
	if sig.Data != data {
		t.Errorf("Data = %s, want %s", sig.Data, data)
	}
}
