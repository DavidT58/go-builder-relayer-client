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
	// Use a valid URL-safe base64 encoded secret (matching Python implementation)
	secret := base64.URLEncoding.EncodeToString([]byte("test-secret-key"))
	config := NewBuilderConfig("test-key", secret, "test-pass")

	headers, err := config.GenerateBuilderHeaders("POST", "/api/v1/test", map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("GenerateBuilderHeaders failed: %v", err)
	}

	// Check required headers are present (note: underscores, not hyphens)
	requiredHeaders := []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_SIGNATURE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_PASSPHRASE", "Content-Type"}
	for _, header := range requiredHeaders {
		if _, exists := headers[header]; !exists {
			t.Errorf("Missing required header: %s", header)
		}
	}

	// Verify header values
	if headers["POLY_BUILDER_API_KEY"] != "test-key" {
		t.Errorf("POLY_BUILDER_API_KEY = %s, want test-key", headers["POLY_BUILDER_API_KEY"])
	}
	if headers["POLY_BUILDER_PASSPHRASE"] != "test-pass" {
		t.Errorf("POLY_BUILDER_PASSPHRASE = %s, want test-pass", headers["POLY_BUILDER_PASSPHRASE"])
	}
	if headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", headers["Content-Type"])
	}
}
