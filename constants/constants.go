package constants

// SAFE_INIT_CODE_HASH is the keccak256 hash of the Safe proxy init code
// This is used for CREATE2 address derivation
const SAFE_INIT_CODE_HASH = "0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf"

// ZERO_ADDRESS is the Ethereum zero address
const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000"

// SAFE_FACTORY_NAME is the name used in EIP-712 domain for Safe proxy factory
const SAFE_FACTORY_NAME = "Polymarket Contract Proxy Factory"

// SAFE_TX_TYPEHASH is the EIP-712 type hash for SafeTx
// keccak256("SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)")
// This is computed dynamically in builder/eip712.go using GetSafeTxTypeHash()
const SAFE_TX_TYPEHASH = "0xbb8310d486368db6bd6f849402fdd73ad53d316b5a4b2644ad6efe0f941286d8"

// CREATE_PROXY_TYPEHASH is the EIP-712 type hash for CreateProxy
// keccak256("CreateProxy(address singleton,bytes initializer,uint256 saltNonce)")
// This is computed dynamically in builder/eip712.go using GetCreateProxyTypeHash()
const CREATE_PROXY_TYPEHASH = "0x7f4e0e3a2e8c0f5d9e6e5c7c4a0f0a2f7e9d5f4e7b0f0e0e0f0c0a0d0f0e0f0"

// MULTISEND_FUNCTION_SELECTOR is the function selector for multiSend(bytes)
// keccak256("multiSend(bytes)")[0:4]
const MULTISEND_FUNCTION_SELECTOR = "0x8d80ff0a"
