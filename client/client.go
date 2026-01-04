package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davidt58/go-builder-relayer-client/builder"
	"github.com/davidt58/go-builder-relayer-client/config"
	"github.com/davidt58/go-builder-relayer-client/errors"
	"github.com/davidt58/go-builder-relayer-client/http"
	"github.com/davidt58/go-builder-relayer-client/models"
	"github.com/davidt58/go-builder-relayer-client/signer"
)

// RelayClient is the main client for interacting with the Relayer API
type RelayClient struct {
	relayerURL     string
	chainID        int64
	contractConfig *config.ContractConfig
	signer         *signer.Signer
	builderConfig  *config.BuilderConfig
	httpClient     *http.Client
	logger         *log.Logger
}

// NewRelayClient creates a new RelayClient instance
// privateKey can be empty if only read operations are needed
// builderConfig can be nil if only read operations are needed
func NewRelayClient(relayerURL string, chainID int64, privateKey string, builderConfig *config.BuilderConfig) (*RelayClient, error) {
	// Validate relayer URL
	if relayerURL == "" {
		return nil, errors.ErrMissingRequiredField("relayerURL")
	}

	// Get contract configuration for the chain
	contractConfig, err := config.GetContractConfig(chainID)
	if err != nil {
		return nil, err
	}

	// Create HTTP client
	httpClient := http.NewClient(relayerURL)

	// Create logger
	logger := log.New(os.Stdout, "[RelayClient] ", log.LstdFlags)

	// Create signer if private key is provided
	var sig *signer.Signer
	if privateKey != "" {
		sig, err = signer.NewSigner(privateKey, chainID)
		if err != nil {
			return nil, err
		}
	}

	client := &RelayClient{
		relayerURL:     relayerURL,
		chainID:        chainID,
		contractConfig: contractConfig,
		signer:         sig,
		builderConfig:  builderConfig,
		httpClient:     httpClient,
		logger:         logger,
	}

	return client, nil
}

// GetNonce retrieves the nonce for the signer
func (c *RelayClient) GetNonce(signerAddress, signerType string) (*models.NonceResponse, error) {
	// Build query parameters
	path := fmt.Sprintf("%s?signerAddress=%s&signerType=%s", GET_NONCE, signerAddress, signerType)

	// Make GET request
	var response models.NonceResponse
	if err := c.httpClient.GetJSON(path, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTransaction retrieves a transaction by ID
func (c *RelayClient) GetTransaction(transactionID string) (*models.RelayerTransaction, error) {
	// Build path
	path := fmt.Sprintf("%s/%s", GET_TRANSACTION, transactionID)

	// Make GET request
	var response models.RelayerTransaction
	if err := c.httpClient.GetJSON(path, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTransactions retrieves all transactions for the builder
func (c *RelayClient) GetTransactions() (*models.GetTransactionsResponse, error) {
	// Ensure builder credentials are configured
	if err := c.assertBuilderCredsNeeded(); err != nil {
		return nil, err
	}

	// Generate authentication headers
	headers, err := c.generateBuilderHeaders("GET", GET_TRANSACTIONS, nil)
	if err != nil {
		return nil, err
	}

	// Make GET request
	var response models.GetTransactionsResponse
	if err := c.httpClient.GetJSON(GET_TRANSACTIONS, headers, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetDeployed checks if a Safe wallet is deployed
func (c *RelayClient) GetDeployed(safeAddress string) (bool, error) {
	// Build query parameters
	path := fmt.Sprintf("%s?safeAddress=%s", GET_DEPLOYED, safeAddress)

	// Make GET request
	var response models.DeployedResponse
	if err := c.httpClient.GetJSON(path, nil, &response); err != nil {
		return false, err
	}

	return response.Deployed, nil
}

// Deploy creates and submits a Safe wallet deployment transaction
func (c *RelayClient) Deploy() (*models.ClientRelayerTransactionResponse, error) {
	// Ensure signer is configured
	if err := c.assertSignerNeeded(); err != nil {
		return nil, err
	}

	// Ensure builder credentials are configured
	if err := c.assertBuilderCredsNeeded(); err != nil {
		return nil, err
	}

	// Get expected Safe address
	safeAddress, err := c.GetExpectedSafe()
	if err != nil {
		return nil, err
	}

	// Check if already deployed
	deployed, err := c.GetDeployed(safeAddress)
	if err == nil && deployed {
		return nil, errors.NewRelayerClientError(fmt.Sprintf("Safe already deployed at %s", safeAddress), nil)
	}

	// Get nonce
	nonceResp, err := c.GetNonce(c.signer.AddressHex(), string(models.EOA))
	if err != nil {
		return nil, err
	}

	// Build Safe creation transaction request
	createArgs := &models.SafeCreateTransactionArgs{
		SignerAddress: c.signer.AddressHex(),
		SafeAddress:   safeAddress,
		Nonce:         nonceResp.Nonce,
		Metadata:      "",
	}

	request, err := builder.BuildSafeCreateTransactionRequest(createArgs, c.signer, c.chainID)
	if err != nil {
		return nil, err
	}

	// Submit the transaction
	return c.submitTransaction(request)
}

// Execute submits one or more transactions to be executed through the Safe
func (c *RelayClient) Execute(transactions []models.SafeTransaction, metadata string) (*models.ClientRelayerTransactionResponse, error) {
	// Ensure signer is configured
	if err := c.assertSignerNeeded(); err != nil {
		return nil, err
	}

	// Ensure builder credentials are configured
	if err := c.assertBuilderCredsNeeded(); err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return nil, errors.NewRelayerClientError("no transactions provided", nil)
	}

	// Get expected Safe address
	safeAddress, err := c.GetExpectedSafe()
	if err != nil {
		return nil, err
	}

	// Get nonce
	nonceResp, err := c.GetNonce(safeAddress, string(models.SAFE_SIGNER))
	if err != nil {
		return nil, err
	}

	// Build Safe transaction request
	txArgs := &models.SafeTransactionArgs{
		SafeAddress:  safeAddress,
		Transactions: transactions,
		Nonce:        nonceResp.Nonce,
		Metadata:     metadata,
	}

	var request *models.TransactionRequest
	if len(transactions) > 1 {
		// Use multisend for multiple transactions
		request, err = builder.BuildSafeTransactionRequestWithMultisend(txArgs, c.signer, c.chainID, c.contractConfig.SafeMultisend)
	} else {
		// Single transaction
		request, err = builder.BuildSafeTransactionRequest(txArgs, c.signer, c.chainID)
	}

	if err != nil {
		return nil, err
	}

	// Submit the transaction
	return c.submitTransaction(request)
}

// PollUntilState polls a transaction until it reaches one of the target states
func (c *RelayClient) PollUntilState(transactionID string, states []models.RelayerTransactionState, failState models.RelayerTransactionState, maxPolls, pollFrequency int) (*models.RelayerTransaction, error) {
	if maxPolls <= 0 {
		maxPolls = 100 // Default max polls
	}
	if pollFrequency <= 0 {
		pollFrequency = 2 // Default 2 seconds
	}

	// Create a map of target states for quick lookup
	targetStates := make(map[models.RelayerTransactionState]bool)
	for _, state := range states {
		targetStates[state] = true
	}

	// Poll until target state is reached or max polls exceeded
	for i := 0; i < maxPolls; i++ {
		// Get transaction
		txn, err := c.GetTransaction(transactionID)
		if err != nil {
			return nil, err
		}

		// Check if in target state
		if targetStates[txn.State] {
			return txn, nil
		}

		// Check if in fail state
		if failState != "" && txn.State == failState {
			return txn, errors.ErrTransactionFailed(transactionID, string(txn.State))
		}

		// Check if in a terminal failure state
		if txn.IsFailed() {
			return txn, errors.ErrTransactionFailed(transactionID, string(txn.State))
		}

		// Wait before next poll
		time.Sleep(time.Duration(pollFrequency) * time.Second)
	}

	return nil, errors.ErrPollingTimeout(transactionID)
}

// GetExpectedSafe derives the expected Safe address for the signer
func (c *RelayClient) GetExpectedSafe() (string, error) {
	if err := c.assertSignerNeeded(); err != nil {
		return "", err
	}

	safeAddress, err := builder.DeriveSafeAddress(c.signer.Address(), c.chainID)
	if err != nil {
		return "", err
	}

	return safeAddress.Hex(), nil
}

// submitTransaction submits a transaction request to the relayer
func (c *RelayClient) submitTransaction(request *models.TransactionRequest) (*models.ClientRelayerTransactionResponse, error) {
	// Generate authentication headers
	headers, err := c.generateBuilderHeaders("POST", SUBMIT_TRANSACTION, request)
	if err != nil {
		return nil, err
	}

	// Submit the transaction
	var response models.SubmitTransactionResponse
	if err := c.httpClient.PostJSON(SUBMIT_TRANSACTION, headers, request, &response); err != nil {
		return nil, err
	}

	// Create response wrapper
	clientResponse := models.NewClientRelayerTransactionResponse(response.TransactionID)
	clientResponse.SetClient(c)

	return clientResponse, nil
}

// generateBuilderHeaders creates authentication headers for Builder API requests
func (c *RelayClient) generateBuilderHeaders(method, requestPath string, body interface{}) (map[string]string, error) {
	if c.builderConfig == nil {
		return nil, errors.ErrBuilderCredsNotConfigured
	}

	return c.builderConfig.GenerateBuilderHeaders(method, requestPath, body)
}

// assertSignerNeeded checks if signer is configured
func (c *RelayClient) assertSignerNeeded() error {
	if c.signer == nil {
		return errors.ErrSignerNotConfigured
	}
	return nil
}

// assertBuilderCredsNeeded checks if builder credentials are configured
func (c *RelayClient) assertBuilderCredsNeeded() error {
	if c.builderConfig == nil {
		return errors.ErrBuilderCredsNotConfigured
	}
	return c.builderConfig.Validate()
}

// GetSigner returns the signer (if configured)
func (c *RelayClient) GetSigner() *signer.Signer {
	return c.signer
}

// GetChainID returns the chain ID
func (c *RelayClient) GetChainID() int64 {
	return c.chainID
}

// GetRelayerURL returns the relayer URL
func (c *RelayClient) GetRelayerURL() string {
	return c.relayerURL
}

// GetContractConfig returns the contract configuration
func (c *RelayClient) GetContractConfig() *config.ContractConfig {
	return c.contractConfig
}