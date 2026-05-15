package logger_test

import (
	json2 "encoding/json"
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
}

func TestBuildJSONReport_WithFailures(t *testing.T) {
	t.Parallel()
	responses := []runner.HTTPResponse{{Name: "example test", Error: "", StatusCode: 404, WantStatus: 200}}
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
		t.Fatalf("status: want %v, got %v", logger.JSONStdout{Ok: false, Failures: []runner.HTTPResponse{{Name: "example test", Error: "", StatusCode: 404, WantStatus: 200}}}, result)
	}

}
