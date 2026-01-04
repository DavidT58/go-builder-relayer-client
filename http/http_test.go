package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(resp) == 0 {
		t.Error("Expected response body, got empty")
	}
}

func TestPost(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	body := map[string]string{"key": "value"}
	resp, err := client.Post("/test", nil, body)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	if len(resp) == 0 {
		t.Error("Expected response body, got empty")
	}
}

func TestErrorResponse(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
