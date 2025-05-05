package flags

import (
	"flag"
	"fmt"
	"os"
)

type CLIArgs struct {
	ConfigPath string
	EventsPath string
	OutputPath string
}

func ParseFlags() (CLIArgs, error) {
	var args CLIArgs
	flag.StringVar(&args.ConfigPath, "config", "", "Path to config file (required)")
	flag.StringVar(&args.EventsPath, "events", "", "Path to events file (required)")
	flag.StringVar(&args.OutputPath, "out", "", "Path to output report file (optional)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -config <path> -events <path>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if args.ConfigPath == "" || args.EventsPath == "" {
		return CLIArgs{}, fmt.Errorf("missing required arguments flags\n")
	}

	return args, nil
}
