package flags

import (
	"flag"
	"fmt"
	"os"
)

type CLIArgs struct {
	ConfigPath  string
	EventsPath  string
	ReportPath  string
	LogFilePath string
}

func ParseFlags() (CLIArgs, error) {
	var args CLIArgs
	flag.StringVar(&args.ConfigPath, "config", "", "Path to config file (required)")
	flag.StringVar(&args.EventsPath, "events", "", "Path to events file (required)")
	flag.StringVar(&args.LogFilePath, "output", "output.log", "Path to ouptup log file (optional)")
	flag.StringVar(&args.ReportPath, "report", "report.txt", "Path to report file (optional)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -config <path> -events <path>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if args.ConfigPath == "" || args.EventsPath == "" {
		return CLIArgs{}, fmt.Errorf("missing required arguments flags")
	}

	return args, nil
}
