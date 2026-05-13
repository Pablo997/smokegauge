package runner

import (
	"context"
	"net/http"
	time2 "time"
)

// Run performs one HTTP request; timeout is a Go duration string (e.g. "5s"). On failure, err is non-nil;
// if the server responded anyway, StatusCode may still be set—callers must branch on err before comparing codes.
func Run(method string, url string, timeout string) (int, error) {
	duration, err := time2.ParseDuration(timeout)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if resp == nil {
		return 0, err
	}

	defer resp.Body.Close()
	return resp.StatusCode, err
}
