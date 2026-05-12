package main

import (
	"errors"
	"flag"
	"io/fs"
	"os"

	"github.com/Pablo997/smokegauge/internal/config"
	"github.com/Pablo997/smokegauge/internal/logger"
	"gopkg.in/yaml.v3"
)

// Exit codes: 2 = bad CLI usage, missing file, invalid YAML, or failed config validation; 1 reserved for failed HTTP checks (runner).

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
}
