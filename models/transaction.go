package models

type OperationType int

const (
    Call OperationType = 0
    DelegateCall OperationType = 1
)

type TransactionType string

const (
    SAFE        TransactionType = "SAFE"
    SAFE_CREATE TransactionType = "SAFE-CREATE"
)

type SafeTransaction struct {
    To        string        `json:"to"`
    Operation OperationType `json:"operation"`
    Data      string        `json:"data"`
    Value     string        `json:"value"`
}

type RelayerTransactionState string

const (
    STATE_NEW       RelayerTransactionState = "STATE_NEW"
    STATE_EXECUTED  RelayerTransactionState = "STATE_EXECUTED"
    STATE_MINED     RelayerTransactionState = "STATE_MINED"
    STATE_CONFIRMED RelayerTransactionState = "STATE_CONFIRMED"
    STATE_FAILED    RelayerTransactionState = "STATE_FAILED"
    STATE_INVALID   RelayerTransactionState = "STATE_INVALID"
)