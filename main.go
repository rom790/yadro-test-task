package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type cliArgs struct {
	configPath string
	eventsPath string
	outputPath string
}

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

type Event struct {
	Time         time.Time
	EventID      int
	CompetitorID int
	ExtraParams  []string
}

type Competitor struct {
	ID            int          // Идентификатор участника
	RegisteredAt  time.Time    // Время регистрации (event 1)
	StartPlanned  time.Time    // Назначенное время старта (event 2)
	StartActual   time.Time    // Фактическое время старта (event 4)
	StartLineTime time.Time    // Время выхода на старт (event 3)
	Laps          []Lap        // Основные круги
	PenaltyLaps   PenaltyStats // Информация по штрафным кругам
	CurrentHits   int
	TotalHits     int // Количество попаданий (event 6)
	Shots         int // Количество выстрелов (считается по количеству event 6)
	Status        string
	Comment       string    // Причина выхода, если есть (event 11)
	FinishTime    time.Time // Время выхода из гонки
}
type Lap struct {
	Start        time.Time // начало круга
	Duration     time.Duration
	AverageSpeed float64
}

type PenaltyStats struct {
	Start         time.Time // время входа в штрафные круги (event 8)
	TotalDuration time.Duration
	Count         int // количество пройденных штрафных кругов (вычисляется из промахов)
}

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
	fmt.Println(*config)

	competitors := make(map[int]*Competitor)

	processEvents(args.eventsPath, config, competitors)

	report := generateReport(competitors, config)

	writeReport(report, args.outputPath)
}

func writeReport(report []string, filePath string) error {
	for _, comp := range report {
		fmt.Println(comp)
	}

	if filePath == "" {
		return nil
	}

	content := strings.Join(report, "\n")

	return os.WriteFile(filePath, []byte(content), 0644)
}

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

func processEvents(filePath string, config *ParsedConfig, competitors map[int]*Competitor) error {
	confFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Open events file error: %w\n", err)
	}
	defer confFile.Close()

	formatters := initFormatters()
	scanner := bufio.NewScanner(confFile)

	for scanner.Scan() {
		line := scanner.Text()
		event, _ := parseEvent(line)

		fmt.Println(formatOutput(event, formatters))

		comp, exists := competitors[event.CompetitorID]
		if !exists {
			comp = &Competitor{
				ID: event.CompetitorID,
			}
			competitors[event.CompetitorID] = comp
		}

		outgoing := handleEvent(event, comp, config)

		for _, outEv := range outgoing {
			fmt.Println(formatOutput(outEv, formatters))
		}

	}
	return nil

}

func parseEvent(line string) (Event, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return Event{}, fmt.Errorf("invalid event line\n")
	}

	evTime, err := parseEventTime(strings.Trim(fields[0], "[]"))
	if err != nil {
		return Event{}, fmt.Errorf("time parsing error: %w\n", err)
	}

	var eventID, competitorID int
	eventID, err = strconv.Atoi(fields[1])
	if err != nil {
		return Event{}, fmt.Errorf("event ID parsing error: %w\n", err)
	}

	competitorID, err = strconv.Atoi(fields[2])
	if err != nil {
		return Event{}, fmt.Errorf("competitor ID parsing error: %w\n", err)
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

func parseEventTime(timeStr string) (time.Time, error) {
	const timeLayout = "15:04:05.000"
	return time.Parse(timeLayout, timeStr)
}

func formatOutput(ev Event, formatters map[int]func(Event, string) string) string {
	timeStr := formatEventTimeToStr(ev.Time)

	if formatter, ok := formatters[ev.EventID]; ok {
		return formatter(ev, timeStr)
	}
	return fmt.Sprintf("%s Unknown event(%d) for competitor(%d)", timeStr, ev.EventID, ev.CompetitorID)
}

func formatEventTimeToStr(t time.Time) string {
	const timeLayout = "15:04:05.000"
	return "[" + t.Format(timeLayout) + "]"
}
func formatDurationToStr(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

func initFormatters() map[int]func(Event, string) string {
	return map[int]func(Event, string) string{
		1: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) registered", t, ev.CompetitorID)
		},
		2: func(ev Event, t string) string {
			if len(ev.ExtraParams) >= 1 {
				return fmt.Sprintf("%s The start time for the competitor(%d) was set by a draw to %s", t, ev.CompetitorID, ev.ExtraParams[0])
			}
			return t + " Invalid start event"
		},
		3: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) is on the start line", t, ev.CompetitorID)
		},
		4: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) has started", t, ev.CompetitorID)
		},
		5: func(ev Event, t string) string {
			if len(ev.ExtraParams) >= 1 {
				return fmt.Sprintf("%s The competitor(%d) is on the firing range(%s)", t, ev.CompetitorID, ev.ExtraParams[0])
			}
			return t + " Invalid firing range event"
		},
		6: func(ev Event, t string) string {
			if len(ev.ExtraParams) >= 1 {
				return fmt.Sprintf("%s The target(%s) has been hit by competitor(%d)", t, ev.ExtraParams[0], ev.CompetitorID)
			}
			return t + " Invalid target event"
		},
		7: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) left the firing range", t, ev.CompetitorID)
		},
		8: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) entered the penalty laps", t, ev.CompetitorID)
		},
		9: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) left the penalty laps", t, ev.CompetitorID)
		},
		10: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) ended the main lap", t, ev.CompetitorID)
		},
		11: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) can`t continue: %s", t, ev.CompetitorID, strings.Join(ev.ExtraParams, " "))
		},
		32: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) is disqualified", t, ev.CompetitorID)
		},
		33: func(ev Event, t string) string {
			return fmt.Sprintf("%s The competitor(%d) has finished", t, ev.CompetitorID)
		},
	}
}

func handleEvent(ev Event, comp *Competitor, config *ParsedConfig) []Event {
	var outgoing []Event

	switch ev.EventID {
	case 1: // Registered
		comp.RegisteredAt = ev.Time

	case 2: // Start time set
		t, _ := parseEventTime(ev.ExtraParams[0])
		comp.StartPlanned = t

		comp.Laps = append(comp.Laps, Lap{
			Start: t,
		})

	case 3: // On start line
		comp.StartLineTime = ev.Time

	case 4: // Has started
		comp.StartActual = ev.Time
		// Проверка на опоздание

		if !comp.StartPlanned.IsZero() && ev.Time.After(comp.StartPlanned.Add(config.StartDelta)) {
			comp.Status = "NotStarted"
			outgoing = append(outgoing, Event{
				Time:         ev.Time,
				EventID:      32,
				CompetitorID: comp.ID,
			})
		}

	case 5: // On firing range
		comp.Shots += 5
	case 6: // Hit
		comp.CurrentHits++

	case 7:
		comp.TotalHits += comp.CurrentHits
		comp.PenaltyLaps.Count += 5 - comp.CurrentHits
		comp.CurrentHits = 0

	case 8: // Entered penalty laps
		comp.PenaltyLaps.Start = ev.Time

	case 9: // Left penalty laps
		if !comp.PenaltyLaps.Start.IsZero() {
			duration := ev.Time.Sub(comp.PenaltyLaps.Start)
			comp.PenaltyLaps.TotalDuration += duration
		}

	case 10: // Ended main lap
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
	case 11: // Can't continue
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

func generateReport(competitors map[int]*Competitor, config *ParsedConfig) []string {
	var report []string

	// Собираем всех участников в слайс и сортируем по времени (если есть)
	var sorted []*Competitor
	for _, c := range competitors {
		sorted = append(sorted, c)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return cmpCompsByTime(sorted[i], sorted[j])
	})

	for _, comp := range sorted {
		var status string
		if comp.Status == "NotStarted" || comp.StartActual.IsZero() {
			status = "[NotStarted]"
			// fmt.Println(comp.StartActual)
		} else if comp.Status == "NotFinished" {
			status = "[NotFinished]"
		} else {
			// fmt.Println(comp.FinishTime.Sub(comp.StartPlanned), comp.FinishTime)
			status = "[" + formatDurationToStr(comp.FinishTime.Sub(comp.StartPlanned)) + "]"
		}

		// Формируем строки по кругам
		var laps []string
		for _, lap := range comp.Laps {
			if lap.Duration > 0 {
				laps = append(laps, fmt.Sprintf("{%s, %.3f}", formatDurationToStr(lap.Duration), lap.AverageSpeed))
			} else {
				laps = append(laps, "{,}")
			}
		}

		// Строка по штрафным кругам
		var penalty string
		if comp.PenaltyLaps.TotalDuration > 0 && comp.PenaltyLaps.Count > 0 {
			avg := float64(comp.PenaltyLaps.Count*config.PenaltyLen) / comp.PenaltyLaps.TotalDuration.Seconds()
			penalty = fmt.Sprintf("{%s, %.3f}", formatDurationToStr(comp.PenaltyLaps.TotalDuration), avg)
		} else {
			penalty = "{,}"
		}

		// Хиты и выстрелы
		hitInfo := fmt.Sprintf("%d/%d", comp.TotalHits, comp.Shots)

		line := fmt.Sprintf("%s %d [%s] %s %s",
			status,
			comp.ID,
			strings.Join(laps, ", "),
			penalty,
			hitInfo,
		)
		report = append(report, line)
	}

	return report
}

func cmpCompsByTime(a, b *Competitor) bool {
	var da, db time.Duration

	if a.Status == "" && !a.FinishTime.IsZero() && !a.StartPlanned.IsZero() {
		da = a.FinishTime.Sub(a.StartPlanned)
	} else {
		da = time.Hour * 9999
	}

	if b.Status == "" && !b.FinishTime.IsZero() && !b.StartPlanned.IsZero() {
		db = b.FinishTime.Sub(b.StartPlanned)
	} else {
		db = time.Hour * 9999
	}

	return da < db
}
