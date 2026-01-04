package config

import (
	"testing"
)

func TestGetContractConfig(t *testing.T) {
	tests := []struct {
		name      string
		chainID   int64
		shouldErr bool
	}{
		{"Polygon Amoy", 80002, false},
		{"Polygon Mainnet", 137, false},
		{"Invalid Chain", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := GetContractConfig(tt.chainID)
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if config == nil {
					t.Error("Config should not be nil")
				}
				if config.ChainID != tt.chainID {
					t.Errorf("ChainID = %d, want %d", config.ChainID, tt.chainID)
				}
			}
		})
	}
}

func TestContractConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *ContractConfig
		shouldErr bool
	}{
		{
			name: "valid config",
			config: &ContractConfig{
				ChainID:             80002,
				SafeFactory:         "0x123",
				SafeSingleton:       "0x456",
				SafeFallbackHandler: "0x789",
				SafeMultisend:       "0xabc",
			},
			shouldErr: false,
		},
		{
			name: "missing SafeFactory",
			config: &ContractConfig{
				ChainID:             80002,
				SafeSingleton:       "0x456",
				SafeFallbackHandler: "0x789",
				SafeMultisend:       "0xabc",
			},
			shouldErr: true,
		},
		{
			name: "invalid chain ID",
			config: &ContractConfig{
				ChainID:             0,
				SafeFactory:         "0x123",
				SafeSingleton:       "0x456",
				SafeFallbackHandler: "0x789",
				SafeMultisend:       "0xabc",
			},
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

func TestGetSupportedChainIDs(t *testing.T) {
	chainIDs := GetSupportedChainIDs()
	if len(chainIDs) < 2 {
		t.Errorf("Expected at least 2 supported chains, got %d", len(chainIDs))
	}
}
