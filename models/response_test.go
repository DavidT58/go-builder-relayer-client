package models

import (
	"strings"
	"testing"
)

func TestClientRelayerTransactionResponse_String(t *testing.T) {
	resp := &ClientRelayerTransactionResponse{
		TransactionID: "019b88b1-2839-7ae5-abf2-89ea78c1ce19",
	}
	
	got := resp.String()
	
	// Check that the string contains the transaction ID
	if !strings.Contains(got, "019b88b1-2839-7ae5-abf2-89ea78c1ce19") {
		t.Errorf("String() = %s, want to contain transaction ID", got)
	}
	
	// Check that it starts with "Response{"
	if !strings.HasPrefix(got, "Response{") {
		t.Errorf("String() = %s, want to start with 'Response{'", got)
	}
	
	// Check that it ends with "}"
	if !strings.HasSuffix(got, "}") {
		t.Errorf("String() = %s, want to end with '}'", got)
	}
}

func TestClientRelayerTransactionResponse_Wait_NoClient(t *testing.T) {
	resp := &ClientRelayerTransactionResponse{
		TransactionID: "test-id",
		client:        nil,
	}
	
	_, err := resp.Wait()
	if err == nil {
		t.Error("Wait() with no client should return error")
	}
	
	if !strings.Contains(err.Error(), "client not configured") {
		t.Errorf("Wait() error = %v, want to contain 'client not configured'", err)
	}
}
