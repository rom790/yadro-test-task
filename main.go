package main

import (
	"log"
)

func main() {
	log.SetFlags(0)

	args, err := parseFlags()
	if err != nil {
		log.Printf("Flag processing error: %v\n", err)
		return
	}

	var config *ParsedConfig
	config, err = processConfig(args.configPath)

	if err != nil {
		log.Printf("Config processing error: %v\n", err)
		return
	}

	competitors := make(map[int]*Competitor)

	err = processEvents(args.eventsPath, config, competitors)
	if err != nil {
		log.Printf("processing events error: %v\n", err)
	}

	report := generateReport(competitors, config)

	err = writeReport(report, args.outputPath)
	if err != nil {
		log.Printf("write report error: %v\n", err)
	}
}
