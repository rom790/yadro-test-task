package main

import (
	configPkg "github.com/rom790/yadro-test-task/internal/config"
	eventPkg "github.com/rom790/yadro-test-task/internal/event"
	flags "github.com/rom790/yadro-test-task/internal/flags"
	"log"
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

	err = eventPkg.ProcessEvents(args.EventsPath, config, competitors)
	if err != nil {
		log.Printf("processing events error: %v\n", err)
	}

	report := eventPkg.GenerateReport(competitors, config)

	err = eventPkg.WriteReport(report, args.OutputPath)
	if err != nil {
		log.Printf("write report error: %v\n", err)
	}
}
