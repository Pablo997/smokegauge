// Package runner performs HTTP smoke checks using configuration loaded by [config.Config].
package runner

import (
	"context"
	"net/http"
	time2 "time"

	"github.com/Pablo997/smokegauge/internal/config"
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

// RunChecks runs every check in file sequentially using file.Defaults.Timeout.
// It returns only failed checks; an empty slice means all checks succeeded.
func RunChecks(file config.Config) []HTTPResponse {
	var responses []HTTPResponse
	for _, chk := range file.Checks {
		statusCode, err := Run(chk.Method, chk.URL, file.Defaults.Timeout)
		var response HTTPResponse
		if err != nil {
			response = HTTPResponse{Name: chk.Name, Error: err, StatusCode: statusCode, WantStatus: chk.WantStatus}
			responses = append(responses, response)
		}
		if statusCode != chk.WantStatus {
			if response.Error == nil {
				response = HTTPResponse{Name: chk.Name, Error: err, StatusCode: statusCode, WantStatus: chk.WantStatus}
				responses = append(responses, response)
			}
		}
	}
	return responses
}
