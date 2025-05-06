package event

import (
	"bufio"
	"fmt"
	configPkg "github.com/rom790/yadro-test-task/internal/config"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Time         time.Time
	EventID      int
	CompetitorID int
	ExtraParams  []string
}

type Competitor struct {
	ID            int
	RegisteredAt  time.Time
	StartPlanned  time.Time
	StartActual   time.Time
	StartLineTime time.Time
	Laps          []Lap
	PenaltyLaps   PenaltyStats
	CurrentHits   int
	TotalHits     int
	Shots         int
	Status        string
	Comment       string
	FinishTime    time.Time
}
type Lap struct {
	Start        time.Time
	Duration     time.Duration
	AverageSpeed float64
}

type PenaltyStats struct {
	Start         time.Time
	TotalDuration time.Duration
	Count         int
}

func ProcessEvents(filePath string, config *configPkg.ParsedConfig, competitors map[int]*Competitor, w io.Writer) error {
	confFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open events file error: %w", err)
	}
	defer confFile.Close()

	formatters := initFormatters()
	scanner := bufio.NewScanner(confFile)

	for scanner.Scan() {
		line := scanner.Text()
		event, _ := ParseEvent(line)

		WriteEvent(formatOutput(event, formatters), w)

		comp, exists := competitors[event.CompetitorID]
		if !exists {
			comp = &Competitor{
				ID: event.CompetitorID,
			}
			competitors[event.CompetitorID] = comp
		}

		outgoing := HandleEvent(event, comp, config)

		for _, outEv := range outgoing {
			WriteEvent(formatOutput(outEv, formatters), w)
		}

	}
	return nil

}

func ParseEvent(line string) (Event, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return Event{}, fmt.Errorf("invalid event line")
	}

	evTime, err := parseEventTime(strings.Trim(fields[0], "[]"))
	if err != nil {
		return Event{}, fmt.Errorf("time parsing error: %w", err)
	}

	var eventID, competitorID int
	eventID, err = strconv.Atoi(fields[1])
	if err != nil {
		return Event{}, fmt.Errorf("event ID parsing error: %w", err)
	}

	competitorID, err = strconv.Atoi(fields[2])
	if err != nil {
		return Event{}, fmt.Errorf("competitor ID parsing error: %w", err)
	}

	var extrParams []string
	if len(fields) > 3 {
		extrParams = fields[3:]
	}

	return Event{
		Time:         evTime,
		EventID:      eventID,
		CompetitorID: competitorID,
		ExtraParams:  extrParams,
	}, nil

}
func HandleEvent(ev Event, comp *Competitor, config *configPkg.ParsedConfig) []Event {
	var outgoing []Event

	switch ev.EventID {
	case 1:
		comp.RegisteredAt = ev.Time

	case 2:
		t, _ := parseEventTime(ev.ExtraParams[0])
		comp.StartPlanned = t

		comp.Laps = append(comp.Laps, Lap{
			Start: t,
		})

	case 3:
		comp.StartLineTime = ev.Time

	case 4:
		comp.StartActual = ev.Time

		if !comp.StartPlanned.IsZero() && ev.Time.After(comp.StartPlanned.Add(config.StartDelta)) {
			comp.Status = "NotStarted"
			outgoing = append(outgoing, Event{
				Time:         ev.Time,
				EventID:      32,
				CompetitorID: comp.ID,
			})
		}

	case 5:
		comp.Shots += 5
	case 6:
		comp.CurrentHits++

	case 7:
		comp.TotalHits += comp.CurrentHits
		comp.PenaltyLaps.Count += 5 - comp.CurrentHits
		comp.CurrentHits = 0

	case 8:
		comp.PenaltyLaps.Start = ev.Time

	case 9:
		if !comp.PenaltyLaps.Start.IsZero() {
			duration := ev.Time.Sub(comp.PenaltyLaps.Start)
			comp.PenaltyLaps.TotalDuration += duration
		}

	case 10:
		endedLap := &comp.Laps[len(comp.Laps)-1]

		endedLap.Duration = ev.Time.Sub(endedLap.Start)
		endedLap.AverageSpeed = float64(config.LapLen) / endedLap.Duration.Seconds()

		if len(comp.Laps) < config.Laps {
			comp.Laps = append(comp.Laps, Lap{
				Start: ev.Time,
			})
		} else {
			comp.FinishTime = ev.Time
			outgoing = append(outgoing, Event{
				Time:         ev.Time,
				EventID:      33,
				CompetitorID: comp.ID,
			})
		}
	case 11:
		comp.Status = "NotFinished"
		comp.Comment = strings.Join(ev.ExtraParams, " ")
		comp.FinishTime = ev.Time
		outgoing = append(outgoing, Event{
			Time:         ev.Time,
			EventID:      32,
			CompetitorID: comp.ID,
		})
	}

	return outgoing
}

func parseEventTime(timeStr string) (time.Time, error) {
	const timeLayout = "15:04:05.000"
	return time.Parse(timeLayout, timeStr)
}
