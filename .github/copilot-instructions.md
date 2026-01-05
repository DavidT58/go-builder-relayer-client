# Copilot Instructions for go-builder-relayer-client

## Project Overview
Go client library for interacting with Polymarket's Relayer API infrastructure. Handles Safe wallet deployment, EIP-712 transaction signing, Builder API HMAC authentication, and transaction status monitoring. Port of the [Python reference implementation](https://github.com/Polymarket/py-builder-relayer-client).

## Architecture

### Core Components Flow
```
client.RelayClient → config.BuilderConfig (HMAC auth) → http.Client → Relayer API
                  ↓
              signer.Signer (EIP-712) → builder.DeriveSafeAddress (CREATE2)
```

**Key pattern:** The `RelayClient` (in `client/client.go`) orchestrates three main operations:
1. **Deploy**: Create Safe wallet via `builder.GetSafeDeploymentData()` → sign with EIP-712 → POST to relayer
2. **Execute**: Build transaction via `builder/` helpers → EIP-712 signature → POST with Builder auth headers
3. **Poll**: Monitor transaction state transitions (NEW → EXECUTED → MINED → CONFIRMED)

### Critical Dependencies
- **Safe Contracts**: Chain-specific addresses in `config/config.go` (Polygon Amoy testnet: 80002, mainnet: 137)
- **Builder API Auth**: HMAC-SHA256 with URL-safe base64 encoding (`config/builder.go:GenerateBuilderHeaders`)
- **EIP-712 Signing**: Domain separator + struct hash for Safe transactions (`signer/eip712.go`)
- **CREATE2 Derivation**: Deterministic Safe address from signer+factory (`builder/derive.go`)
- **TransactionRequest Structure**: Matches Python with `from`, `proxyWallet`, `signature`, `signatureParams` fields

## Development Patterns

### Error Handling Convention
Use typed errors from `errors/` package - never return generic errors:
```go
// ✅ Correct
return nil, errors.ErrInvalidChainID(chainID)
return nil, errors.ErrSigningFailed(err)

// ❌ Wrong
return nil, fmt.Errorf("invalid chain: %d", chainID)
```

All errors should be either `RelayerClientError` (client-side) or `RelayerApiError` (API response).

### Model Enums and JSON
Enums like `OperationType` and `RelayerTransactionState` in `models/transaction.go` have custom JSON marshaling:
```go
type OperationType int
const (Call OperationType = 0; DelegateCall OperationType = 1)
// Must implement MarshalJSON/UnmarshalJSON to serialize as int, not string
```

### EIP-712 Signing Pattern
Two-step process (see `signer/eip712.go`):
1. Compute domain separator hash: `keccak256(typeHash ‖ domainData)`
2. Final signature: `Sign(keccak256("\x19\x01" ‖ domainSeparator ‖ structHash))`

The `SignEIP712StructHash` method signs WITHOUT EIP-191 prefix (unlike `Sign` for personal messages).

### Safe Address Derivation
CREATE2 formula in `builder/derive.go`:
```
safeAddress = keccak256(0xff ‖ factoryAddress ‖ salt ‖ keccak256(initCode))[12:]
```
Where `initCode = proxyBytecode ‖ abi.encode(singletonAddress)` and `salt = keccak256(initializer, saltNonce)`

## Testing Strategy

### Table-Driven Tests
Standard pattern across codebase (see `config/builder_test.go`, `http/helpers_test.go`):
```go
tests := []struct {
    name    string
    input   Type
    want    Type
    wantErr bool
}{/*...*/}
for _, tt := range tests { t.Run(tt.name, func(t *testing.T) {/*...*/}) }
```

### Integration Tests
Located in `tests/integration_test.go`. Require `.env` file with:
- `RELAYER_URL`, `CHAIN_ID`, `PK` (private key)
- `BUILDER_API_KEY`, `BUILDER_SECRET`, `BUILDER_PASS_PHRASE`

Run with: `go test -tags=integration ./tests/`

### Unit Test Coverage
Target 80%+ coverage. Each package has corresponding `*_test.go` files testing:
- Happy paths with valid inputs
- Error cases (invalid addresses, malformed data, auth failures)
- Edge cases (zero values, empty arrays, boundary conditions)

## Key Commands

```bash
# Standard development
go test ./...                          # Run all unit tests
go test -cover ./...                   # With coverage report
go test -tags=integration ./tests/     # Integration tests (needs .env)
go build ./examples/deploy.go          # Build example

# Environment setup
cp .env.example .env                   # Create config from template
go mod tidy                            # Update dependencies
```

## Configuration Notes

### Chain-Specific Contracts
Only Polygon Amoy (80002) and Polygon mainnet (137) are pre-configured. To add chains:
```go
config.AddChainConfig(&config.ContractConfig{
    ChainID: 42161, // Arbitrum
    SafeFactory: "0x...",
    // ... other addresses
})
```

### Builder API Authentication
**CRITICAL**: Must match Python implementation exactly:
- **Base64 Encoding**: Use URL-safe base64 (`base64.URLEncoding`) for both secret decoding AND signature encoding
- **Header Names**: Use underscores, not hyphens: `POLY_BUILDER_API_KEY`, `POLY_BUILDER_SIGNATURE`, `POLY_BUILDER_TIMESTAMP`, `POLY_BUILDER_PASSPHRASE`
- **HMAC Format**: `urlsafe_base64(HMAC-SHA256(urlsafe_base64_decode(secret), timestamp + method + path + body))`
- **Secret Format**: The `BUILDER_SECRET` from Polymarket is already URL-safe base64-encoded - decode it before HMAC, don't re-encode it

See `config/builder.go:GenerateBuilderHeaders()` for reference implementation.

## File Organization
- `client/`: Main entry point (`RelayClient` struct with public API)
- `signer/`: Cryptography (EIP-712, ECDSA signing, address derivation)
- `builder/`: Safe transaction builders and CREATE2 address derivation
- `models/`: Data types (transactions, signatures, responses, enums)
- `config/`: Contract addresses (per chain) and Builder API credentials
- `http/`: HTTP client wrapper with error handling
- `errors/`: Typed error definitions
- `examples/`: Runnable usage examples (deploy, execute, get_nonce, get_transaction)

## Important Gotchas
1. **URL-Safe Base64**: MUST use `base64.URLEncoding`, not `base64.StdEncoding` for Builder API (Python uses `urlsafe_b64decode/encode`)
2. **Header Naming**: Builder headers use underscores (`POLY_BUILDER_*`), not hyphens (`POLY-*`)
3. **CreateProxy EIP-712**: For SAFE-CREATE, use payment fields (`paymentToken`, `payment`, `paymentReceiver`), NOT singleton/initializer/saltNonce
4. **SAFE-CREATE Nonce**: Always use "0" - the relayer handles nonces internally for Safe deployments
5. **Transaction States**: Only `STATE_CONFIRMED`, `STATE_FAILED`, `STATE_INVALID` are terminal (see `IsTerminal()`)
6. **Signature V Value**: Must add 27 to recovery ID for Ethereum compatibility (handled in `signer.Sign()`)
7. **TransactionRequest**: Must include `from` (signer address) and `proxyWallet` (Safe address) fields to match Python API
