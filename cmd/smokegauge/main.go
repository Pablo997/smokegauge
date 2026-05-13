// Command smokegauge loads a YAML check file, validates it, runs HTTP probes, and exits:
//
//	0 - all checks passed
//	1 - at least one check failed (network or unexpected status)
//	2 - invalid usage, missing file, invalid YAML, or invalid configuration
package main

import (
	"errors"
	"flag"
	"io/fs"
	"os"

	"github.com/Pablo997/smokegauge/internal/config"
	"github.com/Pablo997/smokegauge/internal/logger"
	"github.com/Pablo997/smokegauge/internal/runner"
	"gopkg.in/yaml.v3"
)

func check(path string, err error) {
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.PrintErr("config file does not exist: %s\n", path)
		} else {
			logger.PrintErr("cannot read config file: %s: %v\n", path, err)
		}
		os.Exit(2)
	}
}

func main() {
	flag.Usage = func() {
		logger.PrintErr("usage: smokegauge -file <path>\n")
		flag.PrintDefaults()
	}

	fileName := flag.String("file", "", "path to checks config file (YAML)")
	flag.Parse()

	if *fileName == "" {
		logger.PrintErr("required flag -file not provided\n")
		logger.PrintErr("usage: smokegauge -file <path>\n")
		os.Exit(2)
	}

	var err error
	data, err := os.ReadFile(*fileName)
	check(*fileName, err)

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

	for _, resp := range responses {
		if resp.Error != nil {
			logger.PrintErr("%s: %v\n", resp.Name, resp.Error)
		} else if resp.StatusCode != resp.WantStatus {
			logger.PrintErr("%s: bad response. Got <%d> and expect <%d>\n", resp.Name, resp.StatusCode, resp.WantStatus)
		}
	}
	if len(responses) > 0 {
		os.Exit(1)
	}
}
