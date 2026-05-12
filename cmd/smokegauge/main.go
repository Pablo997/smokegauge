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
			fmt.Fprintf(os.Stderr, "smokegauge: config file does not exist: %s\n", path)
		} else {
			fmt.Fprintf(os.Stderr, "smokegauge: cannot read config file: %s: %s\n", path, err)
		}
		os.Exit(2)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "smokegauge: usage: smokegauge -file <path>")
		flag.PrintDefaults()
	}

	fileName := flag.String("file", "", "path to checks config file (YAML)")
	flag.Parse()

	if *fileName == "" {
		fmt.Fprintln(os.Stderr, "smokegauge: required flag -file not provided")
		fmt.Fprintln(os.Stderr, "smokegauge: usage: smokegauge -file <path>")
		os.Exit(2)
	}

	_, err := os.ReadFile(*fileName)
	check(*fileName, err)
}
