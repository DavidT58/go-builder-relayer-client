package http

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildURL constructs a URL with query parameters
func BuildURL(baseURL, path string, params map[string]string) string {
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return baseURL + path
	}

	if len(params) > 0 {
		q := u.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// NormalizeURL ensures the URL has proper format
func NormalizeURL(rawURL string) string {
	// Remove trailing slash
	rawURL = strings.TrimSuffix(rawURL, "/")

	// Ensure it has a scheme
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	return rawURL
}

// MergeHeaders merges multiple header maps, with later maps overriding earlier ones
func MergeHeaders(headerMaps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, headers := range headerMaps {
		for key, value := range headers {
			result[key] = value
		}
	}

	return result
}

// FormatPath ensures the path starts with a forward slash
func FormatPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// BuildAuthHeaders creates authentication headers for Builder API
// This is a helper that calls BuilderConfig.GenerateBuilderHeaders
func BuildAuthHeaders(apiKey, secret, passphrase, method, path string, body interface{}) (map[string]string, error) {
	// This will be used in conjunction with config.BuilderConfig
	// For now, return basic structure
	return map[string]string{
		"POLY-API-KEY":    apiKey,
		"POLY-PASSPHRASE": passphrase,
		"Content-Type":    "application/json",
	}, nil
}

// RetryableError checks if an error is retryable
func RetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types that should be retried
	errStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"too many requests",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryable) {
			return true
		}
	}

	return false
}

// ValidateURL validates a URL format
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}
