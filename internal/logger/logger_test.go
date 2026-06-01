package logger_test

import (
	json2 "encoding/json"
	"strings"
	"testing"

	"github.com/Pablo997/smokegauge/internal/logger"
	"github.com/Pablo997/smokegauge/internal/runner"
)

func TestBuildJSONReport_OK(t *testing.T) {
	t.Parallel()
	var responses []runner.HTTPResponse
	json, err := logger.BuildJSONReport(responses)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var result logger.JSONStdout
	err = json2.Unmarshal(json, &result)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Ok || len(result.Failures) != 0 {
		t.Fatalf("status: want %v, got %v", logger.JSONStdout{Ok: true, Failures: nil}, result)
	}

	raw := string(json)
	if !strings.Contains(raw, `"ok"`) || !strings.Contains(raw, `"failures"`) {
		t.Fatalf("JSON should use lowercase keys, got %s", raw)
	}
	if strings.Contains(raw, `"Ok"`) || strings.Contains(raw, `"Failures"`) {
		t.Fatalf("JSON should not use Go field names, got %s", raw)
	}
}

func TestBuildJSONReport_WithFailures(t *testing.T) {
	t.Parallel()
	responses := []runner.HTTPResponse{{Name: "example test", Error: "body does not contain \"ok\"", StatusCode: 200, WantStatus: 200}}
	json, err := logger.BuildJSONReport(responses)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var result logger.JSONStdout
	err = json2.Unmarshal(json, &result)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Ok || len(result.Failures) != 1 {
		t.Fatalf("status: want %v, got %v", logger.JSONStdout{Ok: false, Failures: []runner.HTTPResponse{{Name: "example test", Error: "body does not contain \"ok\"", StatusCode: 200, WantStatus: 200}}}, result)
	}

	raw := string(json)
	for _, key := range []string{`"ok"`, `"failures"`, `"name"`, `"error"`, `"status_code"`, `"want_status"`} {
		if !strings.Contains(raw, key) {
			t.Fatalf("JSON should contain key %s, got %s", key, raw)
		}
	}
	for _, key := range []string{`"Ok"`, `"Failures"`, `"Name"`, `"Error"`, `"StatusCode"`, `"WantStatus"`} {
		if strings.Contains(raw, key) {
			t.Fatalf("JSON should not contain Go field key %s, got %s", key, raw)
		}
	}
}
