package event

import (
	"fmt"
	configPkg "github.com/rom790/yadro-test-task/internal/config"
	"os"
	"sort"
	"strings"
	"time"
)

func WriteEvent(ev string) {
	fmt.Println(ev)
}

func WriteReport(report []string, filePath string) error {
	for _, comp := range report {
		fmt.Println(comp)
	}

	if filePath == "" {
		return nil
	}

	content := strings.Join(report, "\n")

	return os.WriteFile(filePath, []byte(content), 0644)
}

func GenerateReport(competitors map[int]*Competitor, config *configPkg.ParsedConfig) []string {
	var report []string
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
		} else if comp.Status == "NotFinished" {
			status = "[NotFinished]"
		} else {
			status = "[" + formatDurationToStr(comp.FinishTime.Sub(comp.StartPlanned)) + "]"
		}

		var laps []string
		for _, lap := range comp.Laps {
			if lap.Duration > 0 {
				laps = append(laps, fmt.Sprintf("{%s, %.3f}", formatDurationToStr(lap.Duration), lap.AverageSpeed))
			} else {
				laps = append(laps, "{,}")
			}
		}

		var penalty string
		if comp.PenaltyLaps.TotalDuration > 0 && comp.PenaltyLaps.Count > 0 {
			avg := float64(comp.PenaltyLaps.Count*config.PenaltyLen) / comp.PenaltyLaps.TotalDuration.Seconds()
			penalty = fmt.Sprintf("{%s, %.3f}", formatDurationToStr(comp.PenaltyLaps.TotalDuration), avg)
		} else {
			penalty = "{,}"
		}

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
