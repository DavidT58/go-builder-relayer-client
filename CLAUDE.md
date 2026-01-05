# Claude Code Instructions for go-builder-relayer-client

## Project Context
Go implementation of Polymarket's Builder Relayer Client for Safe wallet deployment and transaction execution. Must maintain exact compatibility with the [Python reference implementation](https://github.com/Polymarket/py-builder-relayer-client).

## Critical Implementation Requirements

### 1. Builder API Authentication (CRITICAL)
The authentication MUST match the Python implementation exactly or all API calls will fail with 401 errors.

**Location**: [config/builder.go](config/builder.go)

```go
// CORRECT Implementation
secretBytes, err := base64.URLEncoding.DecodeString(b.Secret)  // URL-safe base64
signature := base64.URLEncoding.EncodeToString(h.Sum(nil))    // URL-safe base64

headers := map[string]string{
    "POLY_BUILDER_API_KEY":    b.APIKey,        // Underscore, not hyphen
    "POLY_BUILDER_SIGNATURE":  signature,        // Underscore, not hyphen
    "POLY_BUILDER_TIMESTAMP":  timestampStr,     // Underscore, not hyphen
    "POLY_BUILDER_PASSPHRASE": b.Passphrase,     // Underscore, not hyphen
}
```

**Common Mistakes**:
- ❌ Using `base64.StdEncoding` instead of `base64.URLEncoding`
- ❌ Using hyphens in header names (`POLY-API-KEY` instead of `POLY_BUILDER_API_KEY`)
- ❌ Missing `_BUILDER_` in header names

### 2. TransactionRequest Structure
Must match Python's model structure with correct field names and JSON tags.

**Location**: [models/signature.go](models/signature.go)

```go
type TransactionRequest struct {
    Type            string           `json:"type"`
    From            string           `json:"from"`              // Signer address (EOA)
    To              json.RawMessage  `json:"to"`
    ProxyWallet     string           `json:"proxyWallet"`       // Safe address
    Data            json.RawMessage  `json:"data"`
    Signature       string           `json:"signature"`
    SignatureParams *SignatureParams `json:"signatureParams,omitempty"`
    Value           json.RawMessage  `json:"value,omitempty"`
    Nonce           *string          `json:"nonce,omitempty"`
    Metadata        *string          `json:"metadata,omitempty"`
}
```

**For SAFE-CREATE transactions**:
- `From` = signer's EOA address
- `ProxyWallet` = derived Safe address
- `To` = factory address
- `Data` = "0x"
- `Nonce` is NOT included (relayer handles it)

**For SAFE transactions**:
- `From` = signer's EOA address
- `ProxyWallet` = Safe wallet address
- `To`, `Value`, `Data` = transaction details
- `Nonce` = Safe transaction nonce

### 3. CreateProxy EIP-712 Structure
The EIP-712 structure for Safe creation uses payment fields, NOT deployment parameters.

**Location**: [builder/eip712.go](builder/eip712.go)

```go
// CORRECT: Matches Python implementation
type CreateProxy struct {
    PaymentToken    common.Address  // Always ZERO_ADDRESS
    Payment         *big.Int        // Always 0
    PaymentReceiver common.Address  // Always ZERO_ADDRESS
}

// EIP-712 type string
"CreateProxy(address paymentToken,uint256 payment,address paymentReceiver)"
```

**Wrong approach** (old implementation):
```go
// ❌ INCORRECT - This does NOT match Python
type CreateProxy struct {
    Singleton   common.Address
    Initializer []byte
    SaltNonce   *big.Int
}
```

### 4. SAFE-CREATE Nonce Handling
For Safe wallet deployment, the nonce is always "0" in the transaction args.

**Location**: [client/client.go](client/client.go)

```go
// Correct: Nonce is always "0" for SAFE-CREATE
createArgs := &models.SafeCreateTransactionArgs{
    SignerAddress: c.signer.AddressHex(),
    SafeAddress:   safeAddress,
    Nonce:         "0",  // Always "0", relayer handles actual nonce
    Metadata:      "",
}
```

Do NOT call `GetNonce()` for SAFE-CREATE transactions.

## Python Reference Alignment

When implementing new features or fixing bugs, always check the Python reference:

### Key Python Files
- **Authentication**: `py_builder_signing_sdk/signing/hmac.py` - HMAC signature with URL-safe base64
- **Models**: `py_builder_relayer_client/models.py` - TransactionRequest, SignatureParams structures
- **CreateProxy**: `py_builder_relayer_client/model/create_proxy.py` - Payment-based EIP-712 structure
- **Safe Builder**: `py_builder_relayer_client/builder/safe.py` - SAFE transaction building
- **Create Builder**: `py_builder_relayer_client/builder/create.py` - SAFE-CREATE transaction building

### Verification Steps
When implementing a feature:
1. Check the Python implementation for field names and structure
2. Verify JSON serialization matches (use `json.Marshal` on test data)
3. Compare EIP-712 type strings character-by-character
4. Test with the same credentials that work in Python client

## Common Debugging Patterns

### Authentication Issues (401 errors)
```bash
# Verify secret format
echo "$BUILDER_SECRET" | base64 -d  # Should decode without errors

# Check header generation
go test -v ./config -run TestBuilderConfig_GenerateBuilderHeaders
```

### Signature Issues (400 "invalid signature")
- Verify EIP-712 domain separator matches Python
- Check that `CreateProxy` uses payment fields, not deployment fields
- Ensure signing method matches (regular `Sign` for SAFE-CREATE, `SignEIP712StructHash` for SAFE)

### Structure Issues (400 "validation error")
- Compare JSON output with Python's `TransactionRequest.to_dict()`
- Verify field names use correct casing (camelCase in JSON)
- Check that `from` and `proxyWallet` fields are present

## Testing Strategy

### Unit Tests
Run tests to verify authentication and structure:
```bash
go test ./config -v          # Builder auth headers
go test ./models -v          # Transaction structures
go test ./builder -v         # EIP-712 signing
```

### Integration Tests
Test against actual Polymarket API:
```bash
# Requires .env with valid credentials
go test -tags=integration ./tests/
go run examples/deploy.go    # Test Safe deployment
```

### Manual Verification
```bash
# Check JSON structure
go run examples/deploy.go 2>&1 | grep -A 20 "DEBUG"

# Compare with Python
cd ../py-builder-relayer-client
python examples/deploy.py
```

## File Organization by Concern

### Authentication & API Communication
- [config/builder.go](config/builder.go) - Builder API HMAC authentication
- [http/client.go](http/client.go) - HTTP client wrapper
- [client/endpoints.go](client/endpoints.go) - API endpoint constants

### Transaction Building & Signing
- [builder/create.go](builder/create.go) - SAFE-CREATE transaction building
- [builder/safe.go](builder/safe.go) - SAFE transaction building
- [builder/eip712.go](builder/eip712.go) - EIP-712 type definitions and hashing
- [signer/signer.go](signer/signer.go) - ECDSA signing operations

### Data Models
- [models/signature.go](models/signature.go) - TransactionRequest, SignatureParams
- [models/transaction.go](models/transaction.go) - Transaction types and states

### Safe Wallet Operations
- [builder/derive.go](builder/derive.go) - CREATE2 Safe address derivation
- [config/config.go](config/config.go) - Chain-specific contract addresses

## Environment Setup

Required `.env` variables:
```bash
# Relayer Configuration
RELAYER_URL=https://relayer-v2.polymarket.com
CHAIN_ID=137                    # 137 for Polygon mainnet, 80002 for Amoy testnet

# Signer Configuration
PK=0x...                        # Private key (with or without 0x prefix)

# Builder API Credentials (from Polymarket)
BUILDER_API_KEY=...             # API key
BUILDER_SECRET=...              # URL-safe base64-encoded secret (use as-is)
BUILDER_PASS_PHRASE=...         # Passphrase
```

**IMPORTANT**: The `BUILDER_SECRET` should be used exactly as provided by Polymarket - it's already URL-safe base64-encoded.

## Quick Reference: Key Differences from Standard Patterns

| Aspect | Standard Go | This Project | Reason |
|--------|-------------|--------------|--------|
| Base64 | `base64.StdEncoding` | `base64.URLEncoding` | Python uses URL-safe variant |
| Headers | Hyphens (`X-API-Key`) | Underscores (`POLY_BUILDER_API_KEY`) | Polymarket API requirement |
| CreateProxy | Deployment params | Payment params | Matches Safe contract interface |
| SAFE-CREATE nonce | Fetched via API | Hardcoded "0" | Relayer manages deployment nonce |
| Request structure | Varied naming | Python-matching fields | API compatibility requirement |

## Maintenance Notes

### When Adding New Features
1. Always check Python reference implementation first
2. Match field names, types, and JSON structure exactly
3. Add tests comparing JSON output with Python
4. Update both copilot-instructions.md and CLAUDE.md

### When Debugging API Errors
1. Enable debug logging in `http/client.go` to see actual requests
2. Compare request JSON with working Python requests
3. Verify authentication headers match exactly
4. Check that EIP-712 signatures match Python output
