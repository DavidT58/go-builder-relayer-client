# Go Client Implementation Plan for Polymarket Relayer API

**Source Reference:** [Polymarket/py-builder-relayer-client](https://github.com/Polymarket/py-builder-relayer-client)

**Date:** 2026-01-03  
**Prepared for:** @DavidT58

---

## 📋 Project Overview

The Python client is a library for interacting with the Polymarket Relayer infrastructure. It handles:
- Safe wallet deployment and transaction execution
- Transaction signing with EIP-712
- Builder API authentication with HMAC signatures
- Transaction status polling
- HTTP communication with the Relayer API

---

## 🏗️ Architecture & Structure

### Project Layout

```
go-builder-relayer-client/
├── client/
│   └── client.go           # Main RelayClient
├── models/
│   ├── transaction.go      # Transaction types and enums
│   ├── signature. go        # Signature-related structs
│   └── response. go         # API response types
├── signer/
│   └── signer.go           # Cryptographic signing
├── builder/
│   ├── safe. go            # Safe transaction builders
│   ├── create.go          # Safe creation builders
│   └── derive.go          # Safe address derivation
├── http/
│   └── client.go          # HTTP helpers
├── config/
│   └── config.go          # Contract configurations
├── errors/
│   └── errors. go          # Custom error types
├── examples/
│   ├── get_nonce.go
│   ├── deploy.go
│   ├── execute. go
│   └── get_transaction.go
├── go. mod
├── go.sum
└── README.md
```

---

## 🔑 Core Components

### 1. Models & Data Structures (`models/`)

#### **transaction.go**

- `OperationType` enum (Call=0, DelegateCall=1)
- `TransactionType` enum (SAFE, SAFE_CREATE)
- `SafeTransaction` struct
- `SafeTransactionArgs` struct
- `SafeCreateTransactionArgs` struct
- `RelayerTransactionState` enum (NEW, EXECUTED, MINED, CONFIRMED, FAILED, INVALID)

```go
type OperationType int

const (
    Call OperationType = 0
    DelegateCall OperationType = 1
)

type SafeTransaction struct {
    To        string        `json:"to"`
    Operation OperationType `json:"operation"`
    Data      string        `json:"data"`
    Value     string        `json:"value"`
}

type TransactionType string

const (
    SAFE        TransactionType = "SAFE"
    SAFE_CREATE TransactionType = "SAFE-CREATE"
)

type RelayerTransactionState string

const (
    STATE_NEW       RelayerTransactionState = "STATE_NEW"
    STATE_EXECUTED  RelayerTransactionState = "STATE_EXECUTED"
    STATE_MINED     RelayerTransactionState = "STATE_MINED"
    STATE_CONFIRMED RelayerTransactionState = "STATE_CONFIRMED"
    STATE_FAILED    RelayerTransactionState = "STATE_FAILED"
    STATE_INVALID   RelayerTransactionState = "STATE_INVALID"
)
```

#### **signature.go**

- `SignatureParams` struct (gas_price, operation, safe_txn_gas, base_gas, etc.)
- `TransactionRequest` struct
- `SplitSig` struct (r, s, v)
- JSON marshaling methods

```go
type SignatureParams struct {
    // SAFE signature params
    GasPrice     *string `json:"gasPrice,omitempty"`
    Operation    *string `json:"operation,omitempty"`
    SafeTxnGas   *string `json:"safeTxnGas,omitempty"`
    BaseGas      *string `json:"baseGas,omitempty"`
    GasToken     *string `json:"gasToken,omitempty"`
    RefundReceiver *string `json:"refundReceiver,omitempty"`
    
    // SAFE-CREATE signature params
    PaymentToken   *string `json:"paymentToken,omitempty"`
    Payment        *string `json:"payment,omitempty"`
    PaymentReceiver *string `json:"paymentReceiver,omitempty"`
}

type TransactionRequest struct {
    Type            string           `json:"type"`
    From            string           `json:"from"`
    To              string           `json:"to"`
    ProxyWallet     string           `json:"proxyWallet"`
    Data            string           `json:"data"`
    Signature       string           `json:"signature"`
    SignatureParams *SignatureParams `json:"signatureParams,omitempty"`
    Value           *string          `json:"value,omitempty"`
    Nonce           *string          `json:"nonce,omitempty"`
    Metadata        *string          `json:"metadata,omitempty"`
}

type SplitSig struct {
    R string `json:"r"`
    S string `json:"s"`
    V string `json:"v"`
}
```

#### **response.go**

```go
type ClientRelayerTransactionResponse struct {
    TransactionID   string
    TransactionHash string
    client          *RelayClient
}

func (r *ClientRelayerTransactionResponse) GetTransaction() (interface{}, error)
func (r *ClientRelayerTransactionResponse) Wait() (interface{}, error)
```

---

### 2. Signer Component (`signer/`)

**Requirements:**
- Import: `github.com/ethereum/go-ethereum/crypto`, `github.com/ethereum/go-ethereum/accounts`
- Support ECDSA private key handling
- Implement EIP-712 struct hash signing
- Implement EIP-191 message signing

**Methods:**

```go
package signer

import (
    "crypto/ecdsa"
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
    privateKey *ecdsa. PrivateKey
    chainID    *big.Int
}

func NewSigner(privateKeyHex string, chainID int64) (*Signer, error)
func (s *Signer) Address() common.Address
func (s *Signer) GetChainID() *big.Int
func (s *Signer) Sign(messageHash []byte) (string, error)
func (s *Signer) SignEIP712StructHash(messageHash []byte) (string, error)
```

---

### 3. HTTP Client (`http/`)

**Features:**
- Generic HTTP request wrapper
- Support GET/POST methods
- Error handling with custom `RelayerApiException`
- JSON request/response handling
- Builder authentication header injection

**Methods:**

```go
package http

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func Request(endpoint, method string, headers map[string]string, data interface{}) (interface{}, error)
func Get(endpoint string, headers map[string]string) (interface{}, error)
func Post(endpoint string, headers map[string]string, data interface{}) (interface{}, error)
```

---

### 4. Main Client (`client/`)

**RelayClient struct:**

```go
package client

import (
    "log"
)

type RelayClient struct {
    relayerURL     string
    chainID        int64
    contractConfig *ContractConfig
    signer         *Signer
    builderConfig  *BuilderConfig
    logger         *log. Logger
}

func NewRelayClient(relayerURL string, chainID int64, privateKey string, builderConfig *BuilderConfig) (*RelayClient, error)
```

**Key Methods:**

1. **Public API Methods:**

```go
// GetNonce retrieves the nonce for the signer
func (c *RelayClient) GetNonce(signerAddress, signerType string) (map[string]interface{}, error)

// GetTransaction retrieves a transaction by ID
func (c *RelayClient) GetTransaction(transactionID string) (interface{}, error)

// GetTransactions retrieves all transactions for the builder
func (c *RelayClient) GetTransactions() ([]interface{}, error)

// GetDeployed checks if a Safe wallet is deployed
func (c *RelayClient) GetDeployed(safeAddress string) (bool, error)

// Deploy creates and submits a Safe wallet deployment transaction
func (c *RelayClient) Deploy() (*ClientRelayerTransactionResponse, error)

// Execute submits one or more transactions to be executed through the Safe
func (c *RelayClient) Execute(transactions []SafeTransaction, metadata string) (*ClientRelayerTransactionResponse, error)

// PollUntilState polls a transaction until it reaches one of the target states
func (c *RelayClient) PollUntilState(transactionID string, states []string, failState string, maxPolls, pollFrequency int) (interface{}, error)
```

2. **Helper Methods:**

```go
// GetExpectedSafe derives the expected Safe address for the signer
func (c *RelayClient) GetExpectedSafe() (string, error)

// generateBuilderHeaders creates authentication headers for Builder API requests
func (c *RelayClient) generateBuilderHeaders(method, requestPath string, body interface{}) (map[string]string, error)

// postRequest makes an authenticated POST request to the relayer
func (c *RelayClient) postRequest(method, requestPath string, body interface{}) (interface{}, error)

// assertSignerNeeded checks if signer is configured
func (c *RelayClient) assertSignerNeeded() error

// assertBuilderCredsNeeded checks if builder credentials are configured
func (c *RelayClient) assertBuilderCredsNeeded() error
```

---

### 5. Builder Components (`builder/`)

#### **safe.go**

```go
package builder

// BuildSafeTransactionRequest creates a signed Safe transaction request
func BuildSafeTransactionRequest(
    signer *Signer,
    args SafeTransactionArgs,
    config ContractConfig,
    metadata string,
) (*TransactionRequest, error)
```

**Responsibilities:**
- EIP-712 domain separator construction
- Multi-transaction encoding
- Signature generation and packing

#### **create.go**

```go
// BuildSafeCreateTransactionRequest creates a signed Safe creation request
func BuildSafeCreateTransactionRequest(
    signer *Signer,
    args SafeCreateTransactionArgs,
    config ContractConfig,
) (*TransactionRequest, error)
```

#### **derive.go**

```go
// Derive calculates the expected Safe address using CREATE2
func Derive(signerAddress, safeFactory string) (string, error)
```

---

### 6. Configuration (`config/`)

#### **config.go**

```go
package config

type ContractConfig struct {
    SafeFactory          string
    ConditionalTokensCtf string
    // ... other contract addresses
}

func GetContractConfig(chainID int64) (*ContractConfig, error)
```

**Builder Config:**

```go
type BuilderConfig struct {
    APIKey     string
    Secret     string
    Passphrase string
}

func NewBuilderConfig(apiKey, secret, passphrase string) *BuilderConfig

// GenerateBuilderHeaders creates HMAC-signed headers for authentication
func (b *BuilderConfig) GenerateBuilderHeaders(method, path string, body interface{}) (map[string]string, error)
```

---

### 7. Error Handling (`errors/`)

**Custom Errors:**

```go
package errors

import "fmt"

type RelayerClientError struct {
    Message string
}

func (e *RelayerClientError) Error() string {
    return e.Message
}

type RelayerApiError struct {
    StatusCode int
    ErrorMsg   interface{}
}

func (e *RelayerApiError) Error() string {
    return fmt. Sprintf("RelayerApiError[status_code=%d, error_message=%v]", e.StatusCode, e.ErrorMsg)
}
```

---

## 📦 Dependencies

### Required Go Modules

```go
require (
    github.com/ethereum/go-ethereum v1.13.x  // Crypto, accounts, common
    github.com/joho/godotenv v1.5.x          // . env file support
    // Builder signing SDK equivalent (may need to port or create)
)
```

**go.mod example:**

```go
module github.com/davidt58/go-builder-relayer-client

go 1.21

require (
    github. com/ethereum/go-ethereum v1.13.8
    github.com/joho/godotenv v1.5.1
)
```

---

## 🔐 Authentication Flow

### 1. Builder API Authentication

- Generate HMAC signatures using API key, secret, passphrase
- Include timestamp, method, path, and body in signature
- Add headers: 
  - `POLYMARKET-SIGNATURE`
  - `POLYMARKET-TIMESTAMP`
  - `POLYMARKET-API-KEY`
  - `POLYMARKET-PASSPHRASE`

**HMAC Signature Algorithm:**

```go
func GenerateSignature(secret, timestamp, method, path, body string) string {
    message := timestamp + method + path + body
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(message))
    return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
```

### 2. Transaction Signing

- Use EIP-712 typed data signing
- Include domain separator with chain ID
- Sign with user's private key

---

## 🎯 API Endpoints to Support

```go
package client

const (
    GET_NONCE          = "/get-nonce"
    GET_DEPLOYED       = "/get-deployed"
    GET_TRANSACTION    = "/get-transaction"
    GET_TRANSACTIONS   = "/get-transactions"
    SUBMIT_TRANSACTION = "/submit-transaction"
)
```

---

## ✅ Implementation Checklist

### Phase 1: Foundation
- [ ] Set up Go module and project structure
- [ ] Define all models and enums
- [ ] Implement error types
- [ ] Create configuration loader

### Phase 2: Core Crypto
- [ ] Implement Signer with EIP-712 support
- [ ] Implement EIP-191 message signing
- [ ] Add Safe address derivation (CREATE2)

### Phase 3: HTTP & Auth
- [ ] Build HTTP client with error handling
- [ ] Implement Builder API authentication (HMAC)
- [ ] Create header generation logic

### Phase 4: Client Methods
- [ ] Implement GetNonce, GetDeployed, GetTransaction(s)
- [ ] Build Deploy() method
- [ ] Build Execute() method
- [ ] Add polling functionality

### Phase 5: Transaction Builders
- [ ] Port Safe transaction builder
- [ ] Port Safe create transaction builder
- [ ] Implement EIP-712 encoding

### Phase 6: Testing & Examples
- [ ] Create example programs (matching Python examples)
- [ ] Add unit tests for core components
- [ ] Integration tests with testnet
- [ ] Documentation

---

## 🎨 Key Differences from Python

1. **Type Safety:** Go's static typing requires explicit struct definitions
2. **Error Handling:** Use Go's explicit error returns instead of exceptions
3. **JSON Handling:** Use struct tags for JSON marshaling/unmarshaling
4. **Concurrency:** Can add goroutines for polling if needed
5. **Crypto Libraries:** Use `go-ethereum` instead of Python's `eth_account`
6. **Package Management:** Go modules instead of pip/setup.py
7. **Interfaces:** Can define interfaces for testability (e.g., HTTPClient interface)

---

## 📚 Additional Considerations

### 1. Builder Signing SDK

The Python client depends on `py-builder-signing-sdk`. You'll need to either:
- Port this to Go
- Implement HMAC signature generation directly in the client

**Recommendation:** Implement directly in the client for simplicity.

### 2. Testing Strategy

```
tests/
├── signer_test.go
├── client_test.go
├── builder_test.go
├── http_test.go
└── integration_test.go
```

**Test Coverage:**
- Unit tests for crypto operations (signing, hashing)
- Mock HTTP responses for client tests
- Integration tests with staging environment
- Table-driven tests for various scenarios

**Example Test:**

```go
func TestSignerAddress(t *testing.T) {
    pk := "0x1234..."
    signer, err := NewSigner(pk, 80002)
    if err != nil {
        t. Fatal(err)
    }
    
    addr := signer.Address()
    expected := "0xabcd..."
    
    if addr. Hex() != expected {
        t.Errorf("expected %s, got %s", expected, addr. Hex())
    }
}
```

### 3. Documentation

#### GoDoc Comments

```go
// Package client provides a Go client for interacting with the Polymarket Relayer API. 
//
// The RelayClient handles authentication, transaction signing, and API communication
// with the Polymarket Relayer infrastructure for Safe wallet deployment and execution.
//
// Example usage:
//
//     config := NewBuilderConfig(apiKey, secret, passphrase)
//     client, err := NewRelayClient(relayerURL, chainID, privateKey, config)
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     resp, err := client.Deploy()
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     txn, err := resp.Wait()
//     fmt.Println("Transaction mined:", txn)
package client
```

#### README.md Structure

```markdown
# go-builder-relayer-client

Go client library for interacting with the Polymarket Relayer infrastructure.

## Installation

```bash
go get github.com/davidt58/go-builder-relayer-client
```

## Quick Start

## Configuration

## Examples

## API Reference

## Contributing

## License
```

### 4. Configuration Management

**Environment Variables:**

```go
package main

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    RelayerURL      string
    ChainID         int64
    PrivateKey      string
    BuilderAPIKey   string
    BuilderSecret   string
    BuilderPassphrase string
}

func LoadConfig() (*Config, error) {
    godotenv.Load()
    
    return &Config{
        RelayerURL:        os.Getenv("RELAYER_URL"),
        ChainID:           parseInt64(os.Getenv("CHAIN_ID")),
        PrivateKey:        os. Getenv("PK"),
        BuilderAPIKey:     os.Getenv("BUILDER_API_KEY"),
        BuilderSecret:     os. Getenv("BUILDER_SECRET"),
        BuilderPassphrase: os.Getenv("BUILDER_PASS_PHRASE"),
    }, nil
}
```

**.env. example:**

```env
RELAYER_URL=https://relayer-v2-staging.polymarket.dev/
CHAIN_ID=80002
PK=your_private_key_here
BUILDER_API_KEY=your_api_key
BUILDER_SECRET=your_api_secret
BUILDER_PASS_PHRASE=your_passphrase
```

---

## 🚀 Example Programs

### examples/get_nonce.go

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/joho/godotenv"
    client "github.com/davidt58/go-builder-relayer-client/client"
)

func main() {
    godotenv.Load()
    
    relayerURL := os.Getenv("RELAYER_URL")
    chainID := parseInt64(os.Getenv("CHAIN_ID"))
    
    c, err := client.NewRelayClient(relayerURL, chainID, "", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    resp, err := c.GetNonce("0x6e0c80c90ea6c15917308F820Eac91Ce2724B5b5", "SAFE")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(resp)
}
```

### examples/deploy. go

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github. com/joho/godotenv"
    "github.com/davidt58/go-builder-relayer-client/client"
    "github.com/davidt58/go-builder-relayer-client/config"
)

func main() {
    godotenv.Load()
    
    relayerURL := os.Getenv("RELAYER_URL")
    chainID := parseInt64(os.Getenv("CHAIN_ID"))
    pk := os.Getenv("PK")
    
    builderConfig := config.NewBuilderConfig(
        os.Getenv("BUILDER_API_KEY"),
        os.Getenv("BUILDER_SECRET"),
        os.Getenv("BUILDER_PASS_PHRASE"),
    )
    
    c, err := client.NewRelayClient(relayerURL, chainID, pk, builderConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    resp, err := c.Deploy()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt. Println("Transaction ID:", resp.TransactionID)
    
    txn, err := resp.Wait()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt. Println("Deployed:", txn)
}
```

### examples/execute.go

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/joho/godotenv"
    "github. com/davidt58/go-builder-relayer-client/client"
    "github.com/davidt58/go-builder-relayer-client/config"
    "github.com/davidt58/go-builder-relayer-client/models"
)

func main() {
    godotenv.Load()
    
    relayerURL := os.Getenv("RELAYER_URL")
    chainID := parseInt64(os. Getenv("CHAIN_ID"))
    pk := os.Getenv("PK")
    
    builderConfig := config.NewBuilderConfig(
        os. Getenv("BUILDER_API_KEY"),
        os.Getenv("BUILDER_SECRET"),
        os.Getenv("BUILDER_PASS_PHRASE"),
    )
    
    c, err := client.NewRelayClient(relayerURL, chainID, pk, builderConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create USDC approval transaction
    usdc := "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
    ctf := "0x4d97dcd97ec945f40cf65f87097ace5ea0476045"
    
    txn := models.SafeTransaction{
        To:        usdc,
        Operation: models.Call,
        Data:      "0x095ea7b3.. .", // approve calldata
        Value:     "0",
    }
    
    resp, err := c.Execute([]models. SafeTransaction{txn}, "approve USDC on CTF")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Transaction submitted:", resp.TransactionHash)
    
    result, err := resp.Wait()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt. Println("Transaction confirmed:", result)
}
```

---

## 🔧 Utility Functions

### utils/utils.go

```go
package utils

import "strings"

// PrependZx adds "0x" prefix if not present
func PrependZx(hex string) string {
    if strings.HasPrefix(hex, "0x") {
        return hex
    }
    return "0x" + hex
}

// RemoveZx removes "0x" prefix if present
func RemoveZx(hex string) string {
    return strings.TrimPrefix(hex, "0x")
}
```

---

## 🎯 Implementation Priority

1. **Critical Path (Week 1):**
   - Models and data structures
   - Signer implementation
   - HTTP client
   - Basic RelayClient with GetNonce, GetTransaction

2. **Core Functionality (Week 2):**
   - Builder authentication
   - Safe transaction builder
   - Deploy() method
   - Execute() method

3. **Polish & Testing (Week 3):**
   - Polling functionality
   - Error handling refinement
   - Unit tests
   - Integration tests
   - Examples

4. **Documentation (Week 4):**
   - GoDoc comments
   - README
   - Usage guides
   - API reference

---

## 📝 Notes

- **Python Reference:** All implementation details can be cross-referenced with [Polymarket/py-builder-relayer-client](https://github.com/Polymarket/py-builder-relayer-client)
- **EIP-712 Resources:** [EIP-712 Specification](https://eips.ethereum.org/EIPS/eip-712)
- **Safe Contracts:** Understanding Gnosis Safe contract interfaces will be helpful
- **Testing:** Use Polygon Amoy testnet (chainID: 80002) for development

---

## 🤝 Next Steps

1. Review this plan and adjust priorities
2. Set up Git repository structure
3. Initialize Go module
4. Start with Phase 1 implementation
5. Create issues/tasks for each checklist item

---

**End of Implementation Plan**