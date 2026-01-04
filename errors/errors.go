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
    return fmt.Sprintf("RelayerApiError[status_code=%d, error_message=%v]", e.StatusCode, e.ErrorMsg)
}