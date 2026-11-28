package llm

import "fmt"

// APIError is returned by every Provider when the model API answers with a
// non-200 HTTP status. Callers branch on it with errors.As, never by matching
// the error string:
//
//	var apiErr *APIError
//	if errors.As(err, &apiErr) && apiErr.Retryable() {
//		// back off and try again
//	}
type APIError struct {
	Provider string // which adapter failed: "ollama", "claude", ...
	Status   int    // HTTP status code, e.g. 429
	Body     string // response body; the adapter limits its size before storing it
}

// Error formats the error as "<provider>: HTTP <status>: <body>", or
// "<provider>: HTTP <status>" when Body is empty. Examples:
//
//	ollama: HTTP 429: rate limited
//	claude: HTTP 500
func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("%s: HTTP %d", e.Provider, e.Status)
	}

	return fmt.Sprintf("%s: HTTP %d: %s", e.Provider, e.Status, e.Body)
}

// Retryable reports whether sending the same request again might succeed:
// true for 429 (rate limited) and every 5xx (server-side failure), false for
// everything else. A 4xx other than 429 means our request is wrong, so
// repeating it only burns time and money.
func (e *APIError) Retryable() bool {

	return e.Status == 429 || (e.Status >= 500 && e.Status <= 599)
}
