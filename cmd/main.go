package main

import (
	"io"
	"log"
	"os"

	configPkg "github.com/rom790/yadro-test-task/internal/config"
	eventPkg "github.com/rom790/yadro-test-task/internal/event"
	flags "github.com/rom790/yadro-test-task/internal/flags"
)

func main() {
	log.SetFlags(0)

	args, err := flags.ParseFlags()
	if err != nil {
		log.Printf("Flag processing error: %v\n", err)
		return
	}

	var config *configPkg.ParsedConfig
	config, err = configPkg.ProcessConfig(args.ConfigPath)

	if err != nil {
		log.Printf("Config processing error: %v\n", err)
		return
	}

	competitors := make(map[int]*eventPkg.Competitor)

	var logFile io.Writer
	logFile, err = os.Create(args.LogFilePath)
	if err != nil {
		log.Printf("Creating log file error: %v\n", err)
		return
	}

	err = eventPkg.ProcessEvents(args.EventsPath, config, competitors, logFile)
	if err != nil {
		log.Printf("processing events error: %v\n", err)
		return
	}

	report := eventPkg.GenerateReport(competitors, config)

	var reportFile io.Writer
	reportFile, err = os.Create(args.ReportPath)
	if err != nil {
		log.Printf("Creating report file error: %v\n", err)
		return
	}

	eventPkg.WriteReport(report, reportFile)

}
