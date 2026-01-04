package models

type ClientRelayerTransactionResponse struct {
    TransactionID   string
    TransactionHash string
    client          *RelayClient
}

func (r *ClientRelayerTransactionResponse) GetTransaction() (interface{}, error) {
    // Implementation to retrieve transaction details from the Relayer API
}

func (r *ClientRelayerTransactionResponse) Wait() (interface{}, error) {
    // Implementation to wait for the transaction to be confirmed
}