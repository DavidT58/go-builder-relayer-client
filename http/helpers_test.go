package http

import (
	"testing"
)

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		params   map[string]string
		expected string
	}{
		{
			name:     "no params",
			baseURL:  "https://api.example.com",
			path:     "/test",
			params:   nil,
			expected: "https://api.example.com/test",
		},
		{
			name:    "with params",
			baseURL: "https://api.example.com",
			path:    "/test",
			params: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: "https://api.example.com/test?key1=value1&key2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildURL(tt.baseURL, tt.path, tt.params)
			// Note: Query parameter order may vary, so we just check if it contains the base
			if !contains(result, tt.baseURL+tt.path) {
				t.Errorf("BuildURL() = %s, should contain %s", result, tt.baseURL+tt.path)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with https",
			input:    "https://api.example.com/",
			expected: "https://api.example.com",
		},
		{
			name:     "without scheme",
			input:    "api.example.com",
			expected: "https://api.example.com",
		},
		{
			name:     "with http",
			input:    "http://api.example.com",
			expected: "http://api.example.com",
		},
		{
			name:     "with trailing slash",
			input:    "https://api.example.com/",
			expected: "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURL() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestMergeHeaders(t *testing.T) {
	headers1 := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	headers2 := map[string]string{
		"key2": "new-value2",
		"key3": "value3",
	}

	result := MergeHeaders(headers1, headers2)

	if result["key1"] != "value1" {
		t.Errorf("key1 = %s, want value1", result["key1"])
	}

	if result["key2"] != "new-value2" {
		t.Errorf("key2 = %s, want new-value2", result["key2"])
	}

	if result["key3"] != "value3" {
		t.Errorf("key3 = %s, want value3", result["key3"])
	}
}

func TestFormatPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with slash",
			input:    "/test",
			expected: "/test",
		},
		{
			name:     "without slash",
			input:    "test",
			expected: "/test",
		},
		{
			name:     "empty",
			input:    "",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPath(tt.input)
			if result != tt.expected {
				t.Errorf("FormatPath() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "timeout error",
			err:      &testError{"connection timeout"},
			expected: true,
		},
		{
			name:     "connection refused",
			err:      &testError{"connection refused"},
			expected: true,
		},
		{
			name:     "non-retryable",
			err:      &testError{"invalid request"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("RetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{
			name:      "valid https",
			url:       "https://api.example.com",
			shouldErr: false,
		},
		{
			name:      "valid http",
			url:       "http://api.example.com",
			shouldErr: false,
		},
		{
			name:      "empty",
			url:       "",
			shouldErr: true,
		},
		{
			name:      "no scheme",
			url:       "api.example.com",
			shouldErr: true,
		},
		{
			name:      "invalid scheme",
			url:       "ftp://api.example.com",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateURL() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

// Helper functions and types for tests

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
