package mindl

import (
	"context"
	"io"
	"net/http"
	"os"
)

// get makes a context-aware request to the given URL and returns the response.
func get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

// download downloads a URL to the targeted path.
func download(ctx context.Context, url, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
