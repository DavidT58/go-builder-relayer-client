package client

// API endpoints for the Relayer API
const (
	// GET_NONCE returns the current nonce for a signer
	GET_NONCE = "/nonce"

	// GET_DEPLOYED checks if a Safe wallet is deployed
	GET_DEPLOYED = "/deployed"

	// GET_TRANSACTION retrieves a specific transaction by ID
	GET_TRANSACTION = "/transaction"

	// GET_TRANSACTIONS retrieves all transactions for the builder
	GET_TRANSACTIONS = "/transactions"

	// SUBMIT_TRANSACTION submits a new transaction to the relayer
	SUBMIT_TRANSACTION = "/submit"
)
