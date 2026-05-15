// Package logger formats CLI output: human-readable failures on stderr, JSON reports on stdout.
package logger

import (
	json2 "encoding/json"
	"fmt"
	"os"

	"github.com/Pablo997/smokegauge/internal/runner"
)

// JSONStdout is the root object written to stdout when -format json is used.
type JSONStdout struct {
	Ok       bool
	Failures []runner.HTTPResponse
}

// PrintErr writes a prefixed line to standard error. format and args follow [fmt.Fprintf] conventions.
func PrintErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "smokegauge: "+format, args...)
}

func printJSON(message []byte) {
	fmt.Print(string(message) + "\n")
}

// BuildJSONReport encodes check failures as JSON without writing to stdout.
func BuildJSONReport(responses []runner.HTTPResponse) ([]byte, error) {
	message := JSONStdout{Ok: len(responses) == 0, Failures: responses}
	json, err := json2.Marshal(message)
	return json, err
}

// ShowResults prints failures: format "text" to stderr (transport error preferred over status mismatch),
// format "json" as JSONStdout on stdout. Returns an error if JSON encoding fails.
func ShowResults(responses []runner.HTTPResponse, format string) error {
	if format != "json" {
		for _, resp := range responses {
			if resp.Error != "" {
				PrintErr("%s: %s\n", resp.Name, resp.Error)
			} else if resp.StatusCode != resp.WantStatus {
				PrintErr("%s: bad response. Got <%d> and expect <%d>\n", resp.Name, resp.StatusCode, resp.WantStatus)
			}
		}
	} else {
		json, err := BuildJSONReport(responses)
		if err != nil {
			return err
		}
		printJSON(json)
	}
	return nil
}
