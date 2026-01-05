package config

import (
	"fmt"

	"github.com/davidt58/go-builder-relayer-client/errors"
)

// ContractConfig holds the contract addresses for a specific chain
type ContractConfig struct {
	// SafeFactory is the Safe Proxy Factory contract address
	SafeFactory string
	// SafeSingleton is the Safe Singleton (master copy) contract address
	SafeSingleton string
	// SafeFallbackHandler is the Safe Fallback Handler contract address
	SafeFallbackHandler string
	// SafeMultisend is the Safe MultiSend contract address
	SafeMultisend string
	// ChainID is the blockchain chain ID
	ChainID int64
}

// Polygon Amoy testnet (chainId: 80002) contract addresses
var polygonAmoyConfig = &ContractConfig{
	ChainID:             80002,
	SafeFactory:         "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b",
	SafeSingleton:       "0x3E5c63644E683549055b9Be8653de26E0B4CD36E",
	SafeFallbackHandler: "0xf48f2B2d2a534e402487b3ee7C18c33Aec0Fe5e4",
	SafeMultisend:       "0xA238CBeb142c10Ef7Ad8442C6D1f9E89e07e7761",
}

// Polygon mainnet (chainId: 137) contract addresses
var polygonMainnetConfig = &ContractConfig{
	ChainID:             137,
	SafeFactory:         "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b",
	SafeSingleton:       "0x3E5c63644E683549055b9Be8653de26E0B4CD36E",
	SafeFallbackHandler: "0xf48f2B2d2a534e402487b3ee7C18c33Aec0Fe5e4",
	SafeMultisend:       "0xA238CBeb142c10Ef7Ad8442C6D1f9E89e07e7761",
}

// chainConfigs maps chain IDs to their contract configurations
var chainConfigs = map[int64]*ContractConfig{
	80002: polygonAmoyConfig,
	137:   polygonMainnetConfig,
}

// GetContractConfig returns the contract configuration for a given chain ID
func GetContractConfig(chainID int64) (*ContractConfig, error) {
	config, exists := chainConfigs[chainID]
	if !exists {
		return nil, errors.ErrInvalidChainID(chainID)
	}
	return config, nil
}

// AddChainConfig adds or updates a contract configuration for a chain ID
func AddChainConfig(config *ContractConfig) {
	chainConfigs[config.ChainID] = config
}

// GetSupportedChainIDs returns a list of all supported chain IDs
func GetSupportedChainIDs() []int64 {
	chainIDs := make([]int64, 0, len(chainConfigs))
	for chainID := range chainConfigs {
		chainIDs = append(chainIDs, chainID)
	}
	return chainIDs
}

// Validate checks if the contract configuration is valid
func (c *ContractConfig) Validate() error {
	if c.SafeFactory == "" {
		return errors.ErrMissingRequiredField("SafeFactory")
	}
	if c.SafeSingleton == "" {
		return errors.ErrMissingRequiredField("SafeSingleton")
	}
	if c.SafeFallbackHandler == "" {
		return errors.ErrMissingRequiredField("SafeFallbackHandler")
	}
	if c.SafeMultisend == "" {
		return errors.ErrMissingRequiredField("SafeMultisend")
	}
	if c.ChainID <= 0 {
		return errors.ErrInvalidConfiguration("chain ID must be positive")
	}
	return nil
}

// String returns a string representation of the contract configuration
func (c *ContractConfig) String() string {
	return fmt.Sprintf("ContractConfig{ChainID: %d, SafeFactory: %s, SafeSingleton: %s}",
		c.ChainID, c.SafeFactory, c.SafeSingleton)
}
