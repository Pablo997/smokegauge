package logger

import (
	"fmt"
	"os"
)

// PrintErr writes to stderr with a fixed program prefix; format and args follow fmt.Fprintf rules.
func PrintErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "smokegauge: "+format, args...)
}
