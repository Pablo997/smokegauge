package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: smokegauge -file <path>")
		flag.PrintDefaults()
	}

	fileName := flag.String("file", "", "path to checks config file (YAML)")
	flag.Parse()

	if *fileName == "" {
		fmt.Fprintln(os.Stderr, "smokegauge: required flag -file not provided")
		fmt.Fprintln(os.Stderr, "usage: smokegauge -file <path>")
		os.Exit(2)
	}

	fmt.Println(*fileName)
}
