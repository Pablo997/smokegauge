package runner_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Pablo997/smokegauge/internal/config"
	"github.com/Pablo997/smokegauge/internal/runner"
)

func TestRun_OK(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("want GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	client := &http.Client{}

	got := runner.Run(client, config.Check{URL: srv.URL + "/health", Method: http.MethodGet}, "5s", 1<<20)
	if got.Error != "" {
		t.Fatalf("Run: %v", got.Error)
	}
	if got.StatusCode != http.StatusOK {
		t.Fatalf("status: want %d, got %d", http.StatusOK, got.StatusCode)
	}
}

func TestRun_WrongHTTPStatusStillNoTransportError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	client := &http.Client{}

	got := runner.Run(client, config.Check{URL: srv.URL + "/missing", Method: http.MethodGet}, "5s", 1<<20)
	if got.Error != "" {
		t.Fatalf("unexpected transport error: %v", got.Error)
	}
	if got.StatusCode != http.StatusNotFound {
		t.Fatalf("status: want %d, got %d", http.StatusNotFound, got.StatusCode)
	}
}

func TestRun_InvalidTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	client := &http.Client{}

	got := runner.Run(client, config.Check{URL: srv.URL + "/", Method: http.MethodGet}, "not-a-duration", 1<<20)
	if got.Error == "" {
		t.Fatal("want error from invalid duration, got nil")
	}
}

func TestRunChecks_AllPass(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  2,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "first", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK},
			{Name: "second", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 0 {
		t.Fatalf("want no failures, got %d: %+v", len(got), got)
	}
}

func TestRunChecks_FailsOnUnexpectedStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  1,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "teapot", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 1 {
		t.Fatalf("want 1 failure, got %d: %+v", len(got), got)
	}
	if got[0].Name != "teapot" {
		t.Errorf("Name: want teapot, got %q", got[0].Name)
	}
	if got[0].StatusCode != http.StatusTeapot {
		t.Errorf("StatusCode: want %d, got %d", http.StatusTeapot, got[0].StatusCode)
	}
	if got[0].WantStatus != http.StatusOK {
		t.Errorf("WantStatus: want %d, got %d", http.StatusOK, got[0].WantStatus)
	}
	if got[0].Error != "" {
		t.Errorf("want nil transport error for HTTP teapot, got %v", got[0].Error)
	}
}

func TestRunChecks_ConcurrencyRunsAllChecks(t *testing.T) {
	t.Parallel()

	var seen int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&seen, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  1,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "a", Method: http.MethodGet, URL: srv.URL + "/a", WantStatus: http.StatusOK},
			{Name: "b", Method: http.MethodGet, URL: srv.URL + "/b", WantStatus: http.StatusOK},
			{Name: "c", Method: http.MethodGet, URL: srv.URL + "/c", WantStatus: http.StatusOK},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 0 {
		t.Fatalf("want no failures, got %+v", got)
	}
	// Handler invoked once per check (paths differ; server is same handler).
	if atomic.LoadInt32(&seen) != 3 {
		t.Fatalf("want 3 server hits, got %d", atomic.LoadInt32(&seen))
	}
}

func TestRunChecks_BodyContainsPasses(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("service is ok"))
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  1,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "body", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK, BodyContains: "ok"},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 0 {
		t.Fatalf("want no failures, got %+v", got)
	}
}

func TestRunChecks_FailsWhenBodyDoesNotContainExpectedText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("service is down"))
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  1,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "body", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK, BodyContains: "ok"},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 1 {
		t.Fatalf("want 1 failure, got %d: %+v", len(got), got)
	}
	if got[0].Name != "body" {
		t.Errorf("Name: want body, got %q", got[0].Name)
	}
	if got[0].StatusCode != http.StatusOK {
		t.Errorf("StatusCode: want %d, got %d", http.StatusOK, got[0].StatusCode)
	}
	if !strings.Contains(got[0].Error, `body does not contain "ok"`) {
		t.Errorf("Error: want body substring failure, got %q", got[0].Error)
	}
}

func TestRunChecks_BodyContainsIsOptional(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("anything"))
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  1,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "body", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 0 {
		t.Fatalf("want no failures, got %+v", got)
	}
}

func TestRunChecks_StatusFailureTakesPrecedenceOverBodyFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not ok"))
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		Version: 1,
		Defaults: config.Defaults{
			Timeout:      "5s",
			Concurrency:  1,
			MaxBodyBytes: 1 << 20,
		},
		Checks: []config.Check{
			{Name: "status", Method: http.MethodGet, URL: srv.URL + "/", WantStatus: http.StatusOK, BodyContains: "healthy"},
		},
	}

	got := runner.RunChecks(cfg)
	if len(got) != 1 {
		t.Fatalf("want 1 failure, got %d: %+v", len(got), got)
	}
	if got[0].StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode: want %d, got %d", http.StatusNotFound, got[0].StatusCode)
	}
	if got[0].Error != "" {
		t.Errorf("want status failure to take precedence, got error %q", got[0].Error)
	}
}
