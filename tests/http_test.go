package http

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
)

func TestGet(t *testing.T) {
    // Mock server for testing
    server := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "success"})
    })

    ts := httptest.NewServer(server)
    defer ts.Close()

    resp, err := Get(ts.URL, nil)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if resp["message"] != "success" {
        t.Errorf("expected success, got %v", resp["message"])
    }
}

func TestPost(t *testing.T) {
    // Mock server for testing
    server := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var body map[string]string
        json.NewDecoder(r.Body).Decode(&body)
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(body)
    })

    ts := httptest.NewServer(server)
    defer ts.Close()

    data := map[string]string{"key": "value"}
    resp, err := Post(ts.URL, nil, data)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if resp["key"] != "value" {
        t.Errorf("expected value, got %v", resp["key"])
    }
}