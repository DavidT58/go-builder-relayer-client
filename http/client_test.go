package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "https://api.example.com"
	client := NewClient(baseURL)

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.GetBaseURL() != baseURL {
		t.Errorf("BaseURL = %s, want %s", client.GetBaseURL(), baseURL)
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	baseURL := "https://api.example.com"
	timeout := 10 * time.Second
	client := NewClientWithTimeout(baseURL, timeout)

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.httpClient.Timeout != timeout {
		t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, timeout)
	}
}

func TestClient_Get(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}

		response := map[string]string{"message": "success"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	data, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "success" {
		t.Errorf("Response message = %s, want success", response["message"])
	}
}

func TestClient_Post(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		// Read and verify request body
		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody["test"] != "data" {
			t.Errorf("Request body test = %s, want data", requestBody["test"])
		}

		response := map[string]string{"message": "created"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	body := map[string]string{"test": "data"}
	data, err := client.Post("/test", nil, body)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "created" {
		t.Errorf("Response message = %s, want created", response["message"])
	}
}

func TestClient_GetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":   123,
			"name": "test",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	var result map[string]interface{}
	err := client.GetJSON("/test", nil, &result)
	if err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Name = %v, want test", result["name"])
	}
}

func TestClient_PostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":      456,
			"message": "created",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	body := map[string]string{"test": "data"}
	var result map[string]interface{}
	err := client.PostJSON("/test", nil, body, &result)
	if err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}

	if result["message"] != "created" {
		t.Errorf("Message = %v, want created", result["message"])
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]string{"error": "bad request"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Error("Expected error for 400 status code")
	}
}

func TestClient_WithHeaders(t *testing.T) {
	expectedKey := "test-key"
	expectedValue := "test-value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(expectedKey) != expectedValue {
			t.Errorf("Header %s = %s, want %s", expectedKey, r.Header.Get(expectedKey), expectedValue)
		}

		response := map[string]string{"message": "success"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	headers := map[string]string{
		expectedKey: expectedValue,
	}
	_, err := client.Get("/test", headers)
	if err != nil {
		t.Fatalf("Get with headers failed: %v", err)
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client := NewClient("https://api.example.com")
	newTimeout := 5 * time.Second
	client.SetTimeout(newTimeout)

	if client.httpClient.Timeout != newTimeout {
		t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, newTimeout)
	}
}

func TestClient_SetBaseURL(t *testing.T) {
	client := NewClient("https://api.example.com")
	newURL := "https://new-api.example.com"
	client.SetBaseURL(newURL)

	if client.GetBaseURL() != newURL {
		t.Errorf("BaseURL = %s, want %s", client.GetBaseURL(), newURL)
	}
}
