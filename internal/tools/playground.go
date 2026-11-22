package tools

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ErrNotFound to manage custom error
var ErrNotFound = errors.New("not found")

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Fetch is a function to fetch any valid URL's status
func Fetch(ctx context.Context, rawURL string) (int, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return 0, fmt.Errorf("invalid URL schema: %s", rawURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return 0, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return resp.StatusCode, nil
}

// Check is a function to check similar error of ErrNotFound
func Check(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// Save is a function to write byte into the out.txt file.
func Save(data []byte) error {
	err := os.WriteFile("out.txt", data, 0600) // two different problems here
	if err != nil {
		return err
	}

	return nil
}
