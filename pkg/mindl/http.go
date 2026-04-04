package mindl

import (
	"context"
	"io"
	"net/http"
	"os"
)

// Get creates a context-aware request and submits it using the http.DefaultClient.
func Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

// Download downloads a URL to the targeted path.
func Download(ctx context.Context, url, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := Get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
