package builder

/*
Package builder provides utilities for building and deriving Safe wallet addresses and transactions.

Safe Address Derivation

Safe wallets use the CREATE2 opcode for deterministic address generation. This allows
predicting the address of a Safe before it's deployed on-chain.

The derivation process:
1. Build the Safe initializer data (call to Safe.setup())
2. Encode the proxy creation bytecode with the singleton address
3. Calculate CREATE2 address using: keccak256(0xff ++ factory ++ salt ++ keccak256(initCode))

Example usage:

    import (
        "github.com/ethereum/go-ethereum/common"
        "github.com/Polymarket/go-builder-relayer-client/builder"
    )

    signerAddr := common.HexToAddress("0x...")
    safeAddr, err := builder.DeriveSafeAddress(signerAddr, 80002)
    if err != nil {
        // handle error
    }

    fmt.Printf("Predicted Safe address: %s\n", safeAddr.Hex())

Safe Deployment Data

The GetSafeDeploymentData function returns all information needed to deploy a Safe:

    data, err := builder.GetSafeDeploymentData(signerAddr, 80002)
    // Returns: safeAddress, singleton, factory, fallbackHandler, initializer, etc.

Address Verification

Verify a Safe address matches the expected derivation:

    valid, err := builder.VerifySafeAddress(signerAddr, expectedAddr, 80002)
    if !valid {
        // address doesn't match
    }
*/
