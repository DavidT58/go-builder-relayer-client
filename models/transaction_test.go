package models

import (
	"encoding/json"
	"testing"
)

func TestOperationType_String(t *testing.T) {
	tests := []struct {
		op       OperationType
		expected string
	}{
		{Call, "Call"},
		{DelegateCall, "DelegateCall"},
		{OperationType(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.expected {
			t.Errorf("OperationType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestOperationType_MarshalJSON(t *testing.T) {
	op := Call
	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := "0"
	if string(data) != expected {
		t.Errorf("MarshalJSON() = %s, want %s", string(data), expected)
	}
}

func TestOperationType_UnmarshalJSON(t *testing.T) {
	var op OperationType
	data := []byte("1")
	err := json.Unmarshal(data, &op)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if op != DelegateCall {
		t.Errorf("UnmarshalJSON() = %v, want %v", op, DelegateCall)
	}
}

func TestRelayerTransactionState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    RelayerTransactionState
		terminal bool
	}{
		{STATE_NEW, false},
		{STATE_EXECUTED, false},
		{STATE_MINED, false},
		{STATE_CONFIRMED, true},
		{STATE_FAILED, true},
		{STATE_INVALID, true},
	}

	for _, tt := range tests {
		if got := tt.state.IsTerminal(); got != tt.terminal {
			t.Errorf("State %s IsTerminal() = %v, want %v", tt.state, got, tt.terminal)
		}
	}
}

func TestSafeTransaction_JSON(t *testing.T) {
	tx := SafeTransaction{
		To:        "0x1234567890123456789012345678901234567890",
		Value:     "1000000000000000000",
		Data:      "0xabcdef",
		Operation: Call,
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded SafeTransaction
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.To != tx.To {
		t.Errorf("To mismatch: got %s, want %s", decoded.To, tx.To)
	}
	if decoded.Value != tx.Value {
		t.Errorf("Value mismatch: got %s, want %s", decoded.Value, tx.Value)
	}
	if decoded.Operation != tx.Operation {
		t.Errorf("Operation mismatch: got %v, want %v", decoded.Operation, tx.Operation)
	}
}

func TestNewSafeTransaction(t *testing.T) {
	to := "0x1234567890123456789012345678901234567890"
	value := "1000"
	data := "0xabcd"

	tx := NewSafeTransaction(to, value, data)

	if tx.To != to {
		t.Errorf("To = %s, want %s", tx.To, to)
	}
	if tx.Value != value {
		t.Errorf("Value = %s, want %s", tx.Value, value)
	}
	if tx.Data != data {
		t.Errorf("Data = %s, want %s", tx.Data, data)
	}
	if tx.Operation != Call {
		t.Errorf("Operation = %v, want %v", tx.Operation, Call)
	}
}

func TestRelayerTransaction_IsMined(t *testing.T) {
	hash := "0xabc123"
	tests := []struct {
		name   string
		tx     RelayerTransaction
		expect bool
	}{
		{
			name:   "not mined",
			tx:     RelayerTransaction{},
			expect: false,
		},
		{
			name:   "mined",
			tx:     RelayerTransaction{Hash: &hash},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tx.IsMined(); got != tt.expect {
				t.Errorf("IsMined() = %v, want %v", got, tt.expect)
			}
		})
	}
}
