package http

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type RelayerApiException struct {
    StatusCode int
    ErrorMsg   interface{}
}

func Request(endpoint, method string, headers map[string]string, data interface{}) (interface{}, error) {
    var body *bytes.Buffer
    if data != nil {
        jsonData, err := json.Marshal(data)
        if err != nil {
            return nil, err
        }
        body = bytes.NewBuffer(jsonData)
    } else {
        body = bytes.NewBuffer([]byte{})
    }

    req, err := http.NewRequest(method, endpoint, body)
    if err != nil {
        return nil, err
    }

    for key, value := range headers {
        req.Header.Set(key, value)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        var errorResponse RelayerApiException
        json.NewDecoder(resp.Body).Decode(&errorResponse)
        errorResponse.StatusCode = resp.StatusCode
        return nil, &errorResponse
    }

    var result interface{}
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, err
    }

    return result, nil
}

func Get(endpoint string, headers map[string]string) (interface{}, error) {
    return Request(endpoint, http.MethodGet, headers, nil)
}

func Post(endpoint string, headers map[string]string, data interface{}) (interface{}, error) {
    return Request(endpoint, http.MethodPost, headers, data)
}