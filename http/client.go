package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/davidt58/go-builder-relayer-client/errors"
	"github.com/davidt58/go-builder-relayer-client/models"
)

// Client is a wrapper around http.Client with custom error handling
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new HTTP client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// NewClientWithTimeout creates a new HTTP client with a custom timeout
func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// Request performs an HTTP request with the given parameters
func (c *Client) Request(method, path string, headers map[string]string, body interface{}) ([]byte, error) {
	// Construct full URL
	url := c.baseURL + path

	// Marshal body if present
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, errors.ErrJSONMarshalFailed(err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, errors.ErrHTTPRequestFailed(err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.ErrHTTPRequestFailed(err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.ErrHTTPRequestFailed(err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// Get performs a GET request
func (c *Client) Get(path string, headers map[string]string) ([]byte, error) {
	return c.Request(http.MethodGet, path, headers, nil)
}

// Post performs a POST request
func (c *Client) Post(path string, headers map[string]string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPost, path, headers, body)
}

// Put performs a PUT request
func (c *Client) Put(path string, headers map[string]string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPut, path, headers, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string, headers map[string]string) ([]byte, error) {
	return c.Request(http.MethodDelete, path, headers, nil)
}

// GetJSON performs a GET request and unmarshals the response into the target
func (c *Client) GetJSON(path string, headers map[string]string, target interface{}) error {
	data, err := c.Get(path, headers)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, target); err != nil {
		return errors.ErrJSONUnmarshalFailed(err)
	}

	return nil
}

// PostJSON performs a POST request and unmarshals the response into the target
func (c *Client) PostJSON(path string, headers map[string]string, body interface{}, target interface{}) error {
	data, err := c.Post(path, headers, body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, target); err != nil {
		return errors.ErrJSONUnmarshalFailed(err)
	}

	return nil
}

// parseAPIError attempts to parse an error response from the API
func parseAPIError(statusCode int, body []byte) error {
	var errorResp models.ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		// If we can't parse the error response, return a generic error
		return errors.NewRelayerApiError(statusCode, string(body))
	}

	// Create a detailed error from the parsed response
	if errorResp.Code != nil {
		return errors.NewRelayerApiErrorWithDetails(statusCode, errorResp.Error, *errorResp.Code, errorResp.Details)
	}

	return errors.NewRelayerApiError(statusCode, errorResp.Error)
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// GetBaseURL returns the base URL
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// SetBaseURL sets the base URL
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}
