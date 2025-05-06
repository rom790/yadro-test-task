package event

import (
	"fmt"
	"strings"
	"time"
)

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
