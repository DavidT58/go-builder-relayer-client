package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/davidt58/go-builder-relayer-client/errors"
)

// BuilderConfig holds the Builder API credentials
type BuilderConfig struct {
	// APIKey is the Builder API key
	APIKey string
	// Secret is the Builder API secret for HMAC signing
	Secret string
	// Passphrase is the Builder API passphrase
	Passphrase string
}

// NewBuilderConfig creates a new BuilderConfig
func NewBuilderConfig(apiKey, secret, passphrase string) *BuilderConfig {
	return &BuilderConfig{
		APIKey:     apiKey,
		Secret:     secret,
		Passphrase: passphrase,
	}
}

// Validate checks if the builder configuration is valid
func (b *BuilderConfig) Validate() error {
	if b.APIKey == "" {
		return errors.ErrMissingRequiredField("APIKey")
	}
	if b.Secret == "" {
		return errors.ErrMissingRequiredField("Secret")
	}
	if b.Passphrase == "" {
		return errors.ErrMissingRequiredField("Passphrase")
	}
	return nil
}

// GenerateBuilderHeaders creates the authentication headers for Builder API requests
// This implements HMAC-SHA256 signature as per Builder API authentication requirements
func (b *BuilderConfig) GenerateBuilderHeaders(method, requestPath string, body interface{}) (map[string]string, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	// Generate timestamp
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	// Prepare body string
	var bodyStr string
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, errors.ErrJSONMarshalFailed(err)
		}
		bodyStr = string(bodyBytes)
	} else {
		bodyStr = ""
	}

	// Create signature message: timestamp + method + requestPath + body
	message := fmt.Sprintf("%s%s%s%s", timestampStr, method, requestPath, bodyStr)

	// Decode the secret from base64
	secretBytes, err := base64.StdEncoding.DecodeString(b.Secret)
	if err != nil {
		return nil, errors.NewRelayerClientError("failed to decode secret", err)
	}

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Return headers
	headers := map[string]string{
		"POLY-API-KEY":    b.APIKey,
		"POLY-SIGNATURE":  signature,
		"POLY-TIMESTAMP":  timestampStr,
		"POLY-PASSPHRASE": b.Passphrase,
		"Content-Type":    "application/json",
	}

	return headers, nil
}

// String returns a string representation (without exposing secrets)
func (b *BuilderConfig) String() string {
	return fmt.Sprintf("BuilderConfig{APIKey: %s..., Passphrase: %s...}",
		truncate(b.APIKey, 8), truncate(b.Passphrase, 8))
}

// truncate helper function to safely display partial values
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
