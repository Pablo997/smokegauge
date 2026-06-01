package main

import (
	"errors"
	"flag"
	"io/fs"
	"os"

	"github.com/Pablo997/smokegauge/internal/logger"
)

type flagsOptions struct {
	fileName string
	format   string
}

// check prints read errors to stderr and exits with code 2 when err is non-nil.
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

// parseFlags parses -file and -format; invalid usage exits with code 2.
func parseFlags() flagsOptions {
	flag.Usage = func() {
		logger.PrintErr("usage: smokegauge -file <path> [-format <text/json> -version]\n")
		flag.PrintDefaults()
	}

	fileName := flag.String("file", "", "path to checks config file (YAML)")
	format := flag.String("format", "text", "style of the response (text/json)")
	showVersion := flag.Bool("version", false, "print the version")
	flag.Parse()

	if *showVersion {
		logger.PrintLn("smokegauge " + version)
		os.Exit(0)
	}

	if *fileName == "" {
		logger.PrintErr("required flag -file not provided\n")
		logger.PrintErr("usage: smokegauge -file <path>\n")
		os.Exit(2)
	}

	if *format != "text" && *format != "json" {
		logger.PrintErr("invalid -format option <%s>\n", *format)
		logger.PrintErr("usage: smokegauge -file <path> [-format <text/json>]\n")
		os.Exit(2)
	}
	return flagsOptions{fileName: *fileName, format: *format}
}
