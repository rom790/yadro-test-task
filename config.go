package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func processConfig(filePath string) (*ParsedConfig, error) {
	confFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Open config file error: %w\n", err)
	}
	defer confFile.Close()

	var rawCfg Config
	if err := json.NewDecoder(confFile).Decode(&rawCfg); err != nil {
		return nil, fmt.Errorf("Read config error: %w\n", err)
	}

	start, err := time.Parse("15:04:05.000", rawCfg.Start)
	if err != nil {
		return nil, fmt.Errorf("Invalid start time format: %w\n", err)
	}

	t, err := time.Parse("15:04:05", rawCfg.StartDelta)
	if err != nil {
		return nil, fmt.Errorf("Invalid start delta format: %w", err)
	}
	startDelta := time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second

	parsed := &ParsedConfig{
		Laps:        rawCfg.Laps,
		LapLen:      rawCfg.LapLen,
		PenaltyLen:  rawCfg.PenaltyLen,
		FiringLines: rawCfg.FiringLines,
		Start:       start,
		StartDelta:  startDelta,
	}
	return parsed, nil
}
