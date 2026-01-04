package models

import (
	"encoding/json"
	"strings"
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

func TestRelayerTransaction_ToFormattedString(t *testing.T) {
	hash := "0xabc123"
	blockNum := int64(12345)
	metadata := "test metadata"
	
	tests := []struct {
		name     string
		tx       RelayerTransaction
		contains []string
	}{
		{
			name: "with hash and block number",
			tx: RelayerTransaction{
				TransactionID: "tx-123",
				State:         STATE_MINED,
				Type:          SAFE,
				SafeAddress:   "0x1234",
				ChainID:       137,
				Hash:          &hash,
				BlockNumber:   &blockNum,
				CreatedAt:     "2024-01-01",
				UpdatedAt:     "2024-01-02",
			},
			contains: []string{"tx-123", "STATE_MINED", "0xabc123", "12345", "0x1234"},
		},
		{
			name: "without optional fields",
			tx: RelayerTransaction{
				TransactionID: "tx-456",
				State:         STATE_NEW,
				Type:          SAFE_CREATE,
				SafeAddress:   "0x5678",
				ChainID:       80002,
				CreatedAt:     "2024-01-03",
				UpdatedAt:     "2024-01-04",
			},
			contains: []string{"tx-456", "STATE_NEW", "nil", "0x5678"},
		},
		{
			name: "with metadata",
			tx: RelayerTransaction{
				TransactionID: "tx-789",
				State:         STATE_CONFIRMED,
				Type:          SAFE,
				SafeAddress:   "0x9abc",
				ChainID:       137,
				Metadata:      &metadata,
				CreatedAt:     "2024-01-05",
				UpdatedAt:     "2024-01-06",
			},
			contains: []string{"tx-789", "test metadata", "0x9abc"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tx.ToFormattedString()
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("ToFormattedString() = %s, want to contain %s", got, substr)
				}
			}
		})
	}
}
