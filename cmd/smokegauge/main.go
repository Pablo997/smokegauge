package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
)

// Exit status 2: invalid CLI usage or unreadable config file. Failed HTTP checks will use 1.

func check(path string, err error) {
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			printErr("config file does not exist: %s\n", path)
		} else {
			printErr("cannot read config file: %s: %v\n", path, err)
		}
		os.Exit(2)
	}
}

func printErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "smokegauge: "+format, args...)
}

func main() {
	flag.Usage = func() {
		printErr("usage: smokegauge -file <path>")
		flag.PrintDefaults()
	}

	fileName := flag.String("file", "", "path to checks config file (YAML)")
	flag.Parse()

	if *fileName == "" {
		printErr("required flag -file not provided")
		printErr("usage: smokegauge -file <path>")
		os.Exit(2)
	}

	_, err := os.ReadFile(*fileName)
	check(*fileName, err)
}
