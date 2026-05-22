// Package runner performs HTTP smoke checks using configuration loaded by [config.Config].
package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	//Optional fields
	BodyContains bool
}

// recordFailure stores at most one failure per check index; transport errors take precedence over status mismatches.
func recordFailure(responses []*HTTPResponse, response HTTPResponse, chk config.Check, i int) {
	if response.Error != "" {
		responses[i] = &HTTPResponse{Name: chk.Name, Error: response.Error, StatusCode: response.StatusCode, WantStatus: chk.WantStatus, BodyContains: response.BodyContains}
	}
	if response.StatusCode != chk.WantStatus {
		if responses[i] == nil {
			responses[i] = &HTTPResponse{Name: chk.Name, Error: "", StatusCode: response.StatusCode, WantStatus: chk.WantStatus, BodyContains: response.BodyContains}
		}
	}
	if chk.BodyContains != "" && response.StatusCode == chk.WantStatus && !response.BodyContains {
		if responses[i] == nil {
			responses[i] = &HTTPResponse{Name: chk.Name, Error: fmt.Sprintf(`body does not contain %q`, chk.BodyContains), StatusCode: response.StatusCode, WantStatus: chk.WantStatus, BodyContains: response.BodyContains}
		}
	}
}

// Run executes one HTTP request with client (safe for concurrent use). timeout must parse as a Go duration (for example "5s").
// When err is non-nil, StatusCode may still be set if the server returned a response before the error.
func Run(client *http.Client, chk config.Check, timeout string, readBodyLimit int64) HTTPResponse {
	duration, err := time2.ParseDuration(timeout)
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, chk.Method, chk.URL, nil)
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}

	resp, err := client.Do(req)
	if resp == nil {
		return HTTPResponse{Error: err.Error()}
	}
	defer resp.Body.Close()

	var bodyReader io.Reader
	body := new(strings.Builder)
	var containsBody bool
	containsBody = true
	if chk.BodyContains != "" {
		bodyReader = io.LimitReader(resp.Body, readBodyLimit)
		_, err = io.Copy(body, bodyReader)
		if err != nil {
			return HTTPResponse{Error: err.Error()}
		}
		containsBody = strings.Contains(body.String(), chk.BodyContains)
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	var errorStr string
	if err != nil {
		errorStr = err.Error()
	} else {
		errorStr = ""
	}
	return HTTPResponse{StatusCode: resp.StatusCode, Error: errorStr, BodyContains: containsBody}
}

// RunChecks runs all checks concurrently with one shared http.Client, bounded by file.Defaults.Concurrency.
// Each request uses file.Defaults.Timeout. Returns only failed checks; an empty slice means success.
func RunChecks(file config.Config) []HTTPResponse {
	var failures []HTTPResponse
	responses := make([]*HTTPResponse, len(file.Checks))
	g := new(errgroup.Group)
	sem := make(chan struct{}, file.Defaults.Concurrency)
	client := &http.Client{}

	for i, chk := range file.Checks {
		chk := chk
		i := i
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()
			response := Run(client, chk, file.Defaults.Timeout, file.Defaults.MaxBodyBytes)
			recordFailure(responses, response, chk, i)
			return nil
		})
	}
	_ = g.Wait()

	for _, slot := range responses {
		if slot != nil {
			failures = append(failures, *slot)
		}
	}
	return failures
}
