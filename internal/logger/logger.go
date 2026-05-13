// Package logger writes formatted messages to the standard error stream for CLI use.
package logger

import (
	"fmt"
	"os"
)

// PrintErr writes a prefixed line to standard error. format and args follow [fmt.Fprintf] conventions.
func PrintErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "smokegauge: "+format, args...)
}
