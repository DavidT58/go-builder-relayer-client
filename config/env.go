package config

import (
	"os"
	"strconv"

	"github.com/davidt58/go-builder-relayer-client/errors"
)

// EnvConfig holds configuration loaded from environment variables
type EnvConfig struct {
	// RelayerURL is the Relayer API base URL
	RelayerURL string
	// ChainID is the blockchain chain ID
	ChainID int64
	// PrivateKey is the private key for signing (without 0x prefix)
	PrivateKey string
	// BuilderConfig contains Builder API credentials
	BuilderConfig *BuilderConfig
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*EnvConfig, error) {
	relayerURL := os.Getenv("RELAYER_URL")
	if relayerURL == "" {
		return nil, errors.ErrMissingRequiredField("RELAYER_URL")
	}

	chainIDStr := os.Getenv("CHAIN_ID")
	if chainIDStr == "" {
		return nil, errors.ErrMissingRequiredField("CHAIN_ID")
	}

	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		return nil, errors.NewRelayerClientError("invalid CHAIN_ID", err)
	}

	privateKey := os.Getenv("PK")
	// Private key is optional for some operations

	// Load builder credentials (optional)
	var builderConfig *BuilderConfig
	apiKey := os.Getenv("BUILDER_API_KEY")
	secret := os.Getenv("BUILDER_SECRET")
	passphrase := os.Getenv("BUILDER_PASS_PHRASE")

	if apiKey != "" && secret != "" && passphrase != "" {
		builderConfig = NewBuilderConfig(apiKey, secret, passphrase)
	}

	return &EnvConfig{
		RelayerURL:    relayerURL,
		ChainID:       chainID,
		PrivateKey:    privateKey,
		BuilderConfig: builderConfig,
	}, nil
}

// Validate checks if the environment configuration is valid
func (e *EnvConfig) Validate() error {
	if e.RelayerURL == "" {
		return errors.ErrMissingRequiredField("RelayerURL")
	}
	if e.ChainID <= 0 {
		return errors.ErrInvalidConfiguration("chain ID must be positive")
	}
	return nil
}

// HasSigner returns true if a private key is configured
func (e *EnvConfig) HasSigner() bool {
	return e.PrivateKey != ""
}

// HasBuilderConfig returns true if builder credentials are configured
func (e *EnvConfig) HasBuilderConfig() bool {
	return e.BuilderConfig != nil
}
