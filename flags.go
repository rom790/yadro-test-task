package main

import (
	"flag"
	"fmt"
	"os"
)

type cliArgs struct {
	configPath string
	eventsPath string
	outputPath string
}

func parseFlags() (cliArgs, error) {
	var args cliArgs
	flag.StringVar(&args.configPath, "config", "", "Path to config file (required)")
	flag.StringVar(&args.eventsPath, "events", "", "Path to events file (required)")
	flag.StringVar(&args.outputPath, "out", "", "Path to output report file (optional)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -config <path> -events <path>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if args.configPath == "" || args.eventsPath == "" {
		return cliArgs{}, fmt.Errorf("missing required arguments flags\n")
	}

	return args, nil
}
