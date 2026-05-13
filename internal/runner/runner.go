// Package runner performs HTTP smoke checks using configuration loaded by [config.Config].
package runner

import (
	"context"
	"net/http"
	time2 "time"

	"github.com/Pablo997/smokegauge/internal/config"
	"golang.org/x/sync/errgroup"
)

// HTTPResponse captures one failed check for CLI reporting.
type HTTPResponse struct {
	Name       string
	Error      error
	StatusCode int
	WantStatus int
}

// Run executes a single HTTP request. timeout must parse as a Go duration (for example "5s").
// When err is non-nil, StatusCode may still be set if the server returned a response before the error.
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

// RunChecks runs all checks in file concurrently, bounded by file.Defaults.Concurrency, using
// file.Defaults.Timeout for each request. It returns only failed checks; an empty slice means success.
func RunChecks(file config.Config) []HTTPResponse {
	var responses []HTTPResponse
	response := make([]*HTTPResponse, len(file.Checks))
	g := new(errgroup.Group)
	sem := make(chan struct{}, file.Defaults.Concurrency)

	for i, chk := range file.Checks {
		chk := chk
		i := i
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			statusCode, err := Run(chk.Method, chk.URL, file.Defaults.Timeout)
			if err != nil {
				response[i] = &HTTPResponse{Name: chk.Name, Error: err, StatusCode: statusCode, WantStatus: chk.WantStatus}
			}
			if statusCode != chk.WantStatus {
				if response[i] == nil {
					response[i] = &HTTPResponse{Name: chk.Name, Error: err, StatusCode: statusCode, WantStatus: chk.WantStatus}
				}
			}
			return nil
		})
	}
	_ = g.Wait()

	for _, slot := range response {
		if slot != nil {
			responses = append(responses, *slot)
		}
	}
	return responses
}
