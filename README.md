# go-builder-relayer-client

Go client library for interacting with the Polymarket Relayer infrastructure.

## Overview

This client provides a Go implementation for deploying Safe wallets and executing transactions through the Polymarket Relayer API. It handles authentication, transaction signing with EIP-712, and transaction status monitoring.

## Features

- 🔐 Safe wallet deployment and transaction execution
- ✍️ EIP-712 typed data signing
- 🔑 Builder API authentication with HMAC signatures
- 📊 Transaction status polling
- 🌐 HTTP communication with the Relayer API

## Installation

```bash
go get github.com/davidt58/go-builder-relayer-client
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/joho/godotenv"
    "github.com/davidt58/go-builder-relayer-client/client"
    "github.com/davidt58/go-builder-relayer-client/config"
)

func main() {
    godotenv.Load()
    
    // Initialize builder configuration
    builderConfig := config.NewBuilderConfig(
        os.Getenv("BUILDER_API_KEY"),
        os.Getenv("BUILDER_SECRET"),
        os.Getenv("BUILDER_PASS_PHRASE"),
    )
    
    // Create relay client
    relayClient, err := client.NewRelayClient(
        os.Getenv("RELAYER_URL"),
        80002, // Polygon Amoy testnet
        os.Getenv("PK"),
        builderConfig,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Deploy Safe wallet
    resp, err := relayClient.Deploy()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Transaction ID:", resp.TransactionID)
}
```

## Configuration

Create a `.env` file based on `.env.example`:

```env
RELAYER_URL=https://relayer-v2-staging.polymarket.dev/
CHAIN_ID=80002
PK=your_private_key_here
BUILDER_API_KEY=your_api_key_here
BUILDER_SECRET=your_api_secret_here
BUILDER_PASS_PHRASE=your_passphrase_here
```

## Project Structure

```
go-builder-relayer-client/
├── client/          # Main RelayClient implementation
├── models/          # Data structures and types
├── signer/          # Cryptographic signing
├── builder/         # Transaction builders
├── http/            # HTTP client utilities
├── config/          # Configuration management
├── errors/          # Custom error types
├── utils/           # Helper functions
└── examples/        # Usage examples
```

## Examples

See the [examples](./examples) directory for complete working examples:

- `get_nonce.go` - Retrieve nonce for a signer
- `get_transaction.go` - Query transaction status
- `deploy.go` - Deploy a Safe wallet
- `execute.go` - Execute transactions through Safe

## API Reference

Full API documentation available at [GoDoc](https://pkg.go.dev/github.com/Polymarket/go-builder-relayer-client)

## Testing

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests (requires testnet setup)
go test -tags=integration ./tests/
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Related Projects

- [py-builder-relayer-client](https://github.com/Polymarket/py-builder-relayer-client) - Python implementation

## Support

For issues and questions, please open an issue on GitHub.