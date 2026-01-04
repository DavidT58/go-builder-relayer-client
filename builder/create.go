package builder

import (
	"github.com/davidt58/go-builder-relayer-client/models"
)

// BuildSafeCreateTransactionRequest creates a signed Safe creation request
func BuildSafeCreateTransactionRequest(
	signer *Signer,
	args models.SafeCreateTransactionArgs,
	config ContractConfig,
) (*models.TransactionRequest, error) {
	// Implementation for creating a signed Safe creation request goes here
	return nil, nil
}
