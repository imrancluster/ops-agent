package llm

import (
	"errors"
	"fmt"
	"testing"
)

// TestAPIErrorRetryable is a table-driven test: one table of cases, one loop.
// Adding a case is one line, and every case gets its own name in the output.
func TestAPIErrorRetryable(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		// Not an error at all: nothing to retry.
		{200, false}, // OK: the request succeeded

		// 4xx = OUR request is wrong. Sending it again sends the same mistake.
		{400, false}, // Bad Request: malformed JSON or a bad parameter
		{401, false}, // Unauthorized: wrong API key; retrying won't fix the key
		{404, false}, // Not Found: wrong URL or model name

		// 429 is the one 4xx exception: the request is fine, we are just too fast.
		{429, true}, // Too Many Requests: wait, then try again

		// 5xx = the SERVER has a problem. It may be gone in a moment.
		{500, true}, // Internal Server Error
		{502, true}, // Bad Gateway: a proxy in front of the API failed
		{503, true}, // Service Unavailable: overloaded or restarting
		{504, true}, // Gateway Timeout: not in the code's list; proves "every 5xx"
		{529, true}, // Anthropic "overloaded" (non-standard, still 5xx)

		// Edges of the range: catch off-by-one mistakes like > instead of >=.
		{499, false}, // last 4xx code
		{599, true},  // last 5xx code
		{600, false}, // outside every valid range
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.status), func(t *testing.T) {
			e := &APIError{Provider: "test", Status: tt.status}
			if got := e.Retryable(); got != tt.want {
				t.Errorf("Retryable() for %d = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestAPIErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		// Body present: provider, status and body all come from the fields.
		{"with body", &APIError{Provider: "ollama", Status: 429, Body: "rate limited"}, "ollama: HTTP 429: rate limited"},
		// Body empty: no trailing ": " (tests the if-branch in Error()).
		{"empty body", &APIError{Provider: "claude", Status: 500}, "claude: HTTP 500"},
		// A different provider proves nothing is hardcoded to "ollama".
		{"other provider", &APIError{Provider: "openai", Status: 401, Body: "invalid api key"}, "openai: HTTP 401: invalid api key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAPIErrorSurvivesWrapping shows why APIError is a type and not a string:
// each layer adds context with %w, and the caller can still find the original.
func TestAPIErrorSurvivesWrapping(t *testing.T) {
	var err error = &APIError{Provider: "claude", Status: 503}
	err = fmt.Errorf("send request: %w", err) // the adapter adds context
	err = fmt.Errorf("agent step 2: %w", err) // the agent loop adds more

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As did not find *APIError in %q", err)
	}
	if apiErr.Status != 503 {
		t.Errorf("Status = %d, want 503", apiErr.Status)
	}
}
