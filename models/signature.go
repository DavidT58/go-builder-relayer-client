package models

type SignatureParams struct {
    GasPrice        *string `json:"gasPrice,omitempty"`
    Operation       *string `json:"operation,omitempty"`
    SafeTxnGas      *string `json:"safeTxnGas,omitempty"`
    BaseGas         *string `json:"baseGas,omitempty"`
    GasToken        *string `json:"gasToken,omitempty"`
    RefundReceiver   *string `json:"refundReceiver,omitempty"`
    PaymentToken    *string `json:"paymentToken,omitempty"`
    Payment         *string `json:"payment,omitempty"`
    PaymentReceiver  *string `json:"paymentReceiver,omitempty"`
}

type TransactionRequest struct {
    Type            string           `json:"type"`
    From            string           `json:"from"`
    To              string           `json:"to"`
    ProxyWallet     string           `json:"proxyWallet"`
    Data            string           `json:"data"`
    Signature       string           `json:"signature"`
    SignatureParams *SignatureParams  `json:"signatureParams,omitempty"`
    Value           *string          `json:"value,omitempty"`
    Nonce           *string          `json:"nonce,omitempty"`
    Metadata        *string          `json:"metadata,omitempty"`
}

type SplitSig struct {
    R string `json:"r"`
    S string `json:"s"`
    V string `json:"v"`
}