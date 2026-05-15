// Command smokegauge loads a YAML check file, validates it, runs HTTP probes, and exits:
//
//	0 - all checks passed
//	1 - at least one check failed (network or unexpected status)
//	2 - invalid usage, missing file, invalid YAML, invalid configuration, or JSON encode error
//
// Flags: -file (required), -format text|json (default text).
package main

import (
	"os"

	"github.com/Pablo997/smokegauge/internal/config"
	"github.com/Pablo997/smokegauge/internal/logger"
	"github.com/Pablo997/smokegauge/internal/runner"
	"gopkg.in/yaml.v3"
)

func main() {
	flags := parseFlags()
	fileName := flags.fileName
	format := flags.format

	var err error
	data, err := os.ReadFile(fileName)
	check(fileName, err)

	var file config.Config
	err = yaml.Unmarshal(data, &file)
	if err != nil {
		logger.PrintErr("invalid config YAML: %v\n", err)
		os.Exit(2)
	}

	validationErrs := file.Validate()
	if len(validationErrs) > 0 {
		for _, validationErr := range validationErrs {
			logger.PrintErr("%v\n", validationErr)
		}
		os.Exit(2)
	}

	responses := runner.RunChecks(file)
	err = logger.ShowResults(responses, format)
	if err != nil {
		logger.PrintErr("cannot create the JSON: ", err)
		os.Exit(2)
	}

	if len(responses) > 0 {
		os.Exit(1)
	}
}
