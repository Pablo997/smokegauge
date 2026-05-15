package runner_test

import (
	"net/http"
	"net/http/httptest"
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

	code, err := runner.Run(http.MethodGet, srv.URL+"/health", "5s")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("status: want %d, got %d", http.StatusOK, code)
	}
}

func TestRun_WrongHTTPStatusStillNoTransportError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	code, err := runner.Run(http.MethodGet, srv.URL+"/missing", "5s")
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if code != http.StatusNotFound {
		t.Fatalf("status: want %d, got %d", http.StatusNotFound, code)
	}
}

func TestRun_InvalidTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	_, err := runner.Run(http.MethodGet, srv.URL+"/", "not-a-duration")
	if err == nil {
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
			Timeout:     "5s",
			Concurrency: 2,
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
			Timeout:     "5s",
			Concurrency: 1,
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
			Timeout:     "5s",
			Concurrency: 1,
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
