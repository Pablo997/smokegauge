package config

import (
	"strings"
	"testing"
)

func TestValidate_OK(t *testing.T) {
	t.Parallel()

	c := Config{
		Version: 1,
		Defaults: Defaults{
			Timeout:     "5s",
			Concurrency: 2,
		},
		Checks: []Check{
			{Name: "a", Method: "GET", URL: "https://example.com/", WantStatus: 200},
		},
	}
	errs := c.Validate()
	if len(errs) != 0 {
		t.Fatalf("want no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_version(t *testing.T) {
	t.Parallel()

	base := Config{
		Defaults: Defaults{Timeout: "1s", Concurrency: 1},
		Checks:   []Check{{Name: "x", Method: "GET", URL: "https://example.com/", WantStatus: 200}},
	}

	tests := []struct {
		name    string
		version int
		wantSub string
	}{
		{name: "missing", version: 0, wantSub: "schema version is missing"},
		{name: "unsupported", version: 2, wantSub: "unsupported schema version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := base
			c.Version = tt.version
			errs := c.Validate()
			if len(errs) == 0 {
				t.Fatal("want at least one error")
			}
			if !strings.Contains(errs[0].Error(), tt.wantSub) {
				t.Fatalf("error %q does not contain %q", errs[0].Error(), tt.wantSub)
			}
		})
	}
}

func TestValidate_timeout(t *testing.T) {
	t.Parallel()

	c := Config{
		Version:  1,
		Defaults: Defaults{Timeout: "not-a-duration", Concurrency: 1},
		Checks:   []Check{{Name: "x", Method: "GET", URL: "https://example.com/", WantStatus: 200}},
	}
	errs := c.Validate()
	if len(errs) == 0 {
		t.Fatal("want timeout parse error")
	}
}

func TestValidate_concurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		concurrency int
	}{
		{name: "zero", concurrency: 0},
		{name: "negative", concurrency: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := Config{
				Version:  1,
				Defaults: Defaults{Timeout: "1s", Concurrency: tt.concurrency},
				Checks:   []Check{{Name: "x", Method: "GET", URL: "https://example.com/", WantStatus: 200}},
			}
			errs := c.Validate()
			if len(errs) == 0 {
				t.Fatal("want concurrency error")
			}
			if !strings.Contains(errs[0].Error(), "concurrency") {
				t.Fatalf("error %q should mention concurrency", errs[0].Error())
			}
		})
	}
}

func TestValidate_checksEmpty(t *testing.T) {
	t.Parallel()

	c := Config{
		Version:  1,
		Defaults: Defaults{Timeout: "1s", Concurrency: 1},
		Checks:   nil,
	}
	errs := c.Validate()
	if len(errs) == 0 {
		t.Fatal("want empty checks error")
	}
	if !strings.Contains(errs[0].Error(), "check section is empty") {
		t.Fatalf("error %q", errs[0].Error())
	}
}

func TestValidate_checkURLEmpty(t *testing.T) {
	t.Parallel()

	c := Config{
		Version:  1,
		Defaults: Defaults{Timeout: "1s", Concurrency: 1},
		Checks: []Check{
			{Name: "good", Method: "GET", URL: "https://example.com/", WantStatus: 200},
			{Name: "bad", Method: "GET", URL: "", WantStatus: 200},
		},
	}
	errs := c.Validate()
	if len(errs) != 1 {
		t.Fatalf("want 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "index 1") {
		t.Fatalf("error %q should mention index 1", errs[0].Error())
	}
}

func TestValidate_multipleErrors(t *testing.T) {
	t.Parallel()

	c := Config{
		Version:  0,
		Defaults: Defaults{Timeout: "bad", Concurrency: 0},
		Checks:   nil,
	}
	errs := c.Validate()
	if len(errs) < 2 {
		t.Fatalf("want at least 2 errors (version + timeout + concurrency + checks accumulate), got %d: %v", len(errs), errs)
	}
}
