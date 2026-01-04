package config

import (
	"encoding/base64"
	"testing"
)

func TestNewBuilderConfig(t *testing.T) {
	apiKey := "test-key"
	secret := "test-secret"
	passphrase := "test-pass"

	config := NewBuilderConfig(apiKey, secret, passphrase)

	if config.APIKey != apiKey {
		t.Errorf("APIKey = %s, want %s", config.APIKey, apiKey)
	}
	if config.Secret != secret {
		t.Errorf("Secret = %s, want %s", config.Secret, secret)
	}
	if config.Passphrase != passphrase {
		t.Errorf("Passphrase = %s, want %s", config.Passphrase, passphrase)
	}
}

func TestBuilderConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *BuilderConfig
		shouldErr bool
	}{
		{
			name:      "valid config",
			config:    NewBuilderConfig("key", "secret", "pass"),
			shouldErr: false,
		},
		{
			name:      "missing API key",
			config:    NewBuilderConfig("", "secret", "pass"),
			shouldErr: true,
		},
		{
			name:      "missing secret",
			config:    NewBuilderConfig("key", "", "pass"),
			shouldErr: true,
		},
		{
			name:      "missing passphrase",
			config:    NewBuilderConfig("key", "secret", ""),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestBuilderConfig_GenerateBuilderHeaders(t *testing.T) {
	// Use a valid base64 encoded secret
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key"))
	config := NewBuilderConfig("test-key", secret, "test-pass")

	headers, err := config.GenerateBuilderHeaders("POST", "/api/v1/test", map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("GenerateBuilderHeaders failed: %v", err)
	}

	// Check required headers are present
	requiredHeaders := []string{"POLY-API-KEY", "POLY-SIGNATURE", "POLY-TIMESTAMP", "POLY-PASSPHRASE", "Content-Type"}
	for _, header := range requiredHeaders {
		if _, exists := headers[header]; !exists {
			t.Errorf("Missing required header: %s", header)
		}
	}

	// Verify header values
	if headers["POLY-API-KEY"] != "test-key" {
		t.Errorf("POLY-API-KEY = %s, want test-key", headers["POLY-API-KEY"])
	}
	if headers["POLY-PASSPHRASE"] != "test-pass" {
		t.Errorf("POLY-PASSPHRASE = %s, want test-pass", headers["POLY-PASSPHRASE"])
	}
	if headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", headers["Content-Type"])
	}
}
