package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `josn:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

type ParsedConfig struct {
	Laps        int
	LapLen      int
	PenaltyLen  int
	FiringLines int
	Start       time.Time
	StartDelta  time.Duration
}

func ProcessConfig(filePath string) (*ParsedConfig, error) {
	confFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open config file error: %w", err)
	}
	defer confFile.Close()

	var rawCfg Config
	if err := json.NewDecoder(confFile).Decode(&rawCfg); err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	start, err := time.Parse("15:04:05.000", rawCfg.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	t, err := time.Parse("15:04:05", rawCfg.StartDelta)
	if err != nil {
		return nil, fmt.Errorf("invalid start delta format: %w", err)
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
