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
	Error      string // transport error message; empty when the failure is status-only
	StatusCode int
	WantStatus int
}

// recordFailure stores at most one failure per check index; transport errors take precedence over status mismatches.
func recordFailure(response []*HTTPResponse, statusCode int, chk config.Check, i int, err error) {
	if err != nil {
		response[i] = &HTTPResponse{Name: chk.Name, Error: err.Error(), StatusCode: statusCode, WantStatus: chk.WantStatus}
	}
	if statusCode != chk.WantStatus {
		if response[i] == nil {
			response[i] = &HTTPResponse{Name: chk.Name, Error: "", StatusCode: statusCode, WantStatus: chk.WantStatus}
		}
	}
}

// Run executes one HTTP request with client (safe for concurrent use). timeout must parse as a Go duration (for example "5s").
// When err is non-nil, StatusCode may still be set if the server returned a response before the error.
func Run(client *http.Client, method string, url string, timeout string) (int, error) {
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

	resp, err := client.Do(req)
	if resp == nil {
		return 0, err
	}

	defer resp.Body.Close()
	return resp.StatusCode, err
}

// RunChecks runs all checks concurrently with one shared http.Client, bounded by file.Defaults.Concurrency.
// Each request uses file.Defaults.Timeout. Returns only failed checks; an empty slice means success.
func RunChecks(file config.Config) []HTTPResponse {
	var failures []HTTPResponse
	response := make([]*HTTPResponse, len(file.Checks))
	g := new(errgroup.Group)
	sem := make(chan struct{}, file.Defaults.Concurrency)
	client := &http.Client{}

	for i, chk := range file.Checks {
		chk := chk
		i := i
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()
			statusCode, err := Run(client, chk.Method, chk.URL, file.Defaults.Timeout)
			recordFailure(response, statusCode, chk, i, err)
			return nil
		})
	}
	_ = g.Wait()

	for _, slot := range response {
		if slot != nil {
			failures = append(failures, *slot)
		}
	}
	return failures
}
