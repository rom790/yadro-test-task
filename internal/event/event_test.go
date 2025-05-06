package event_test

import (
	"testing"
	"time"

	"github.com/rom790/yadro-test-task/internal/config"
	"github.com/rom790/yadro-test-task/internal/event"
	"github.com/stretchr/testify/require"
)

func TestParseEvent_Valid(t *testing.T) {
	line := "[10:00:00.000] 1 42 extra param"
	ev, err := event.ParseEvent(line)
	require.NoError(t, err)

	require.Equal(t, 1, ev.EventID)
	require.Equal(t, 42, ev.CompetitorID)
	require.Equal(t, "extra", ev.ExtraParams[0])
	require.Equal(t, "param", ev.ExtraParams[1])
	require.Equal(t, "10:00:00.000", ev.Time.Format("15:04:05.000"))
}

func TestParseEvent_InvalidTime(t *testing.T) {
	line := "[bad] 1 42"
	_, err := event.ParseEvent(line)
	require.Error(t, err)
	require.Contains(t, err.Error(), "time parsing error")
}

func TestParseEvent_TooFewFields(t *testing.T) {
	line := "[10:00:00.000] 1"
	_, err := event.ParseEvent(line)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid event line")
}

func TestHandleEvent_RegisterAndStart(t *testing.T) {
	cfg := &config.ParsedConfig{
		Laps:       2,
		LapLen:     3000,
		StartDelta: time.Minute,
	}

	comp := &event.Competitor{ID: 1}

	// Регистрация
	regEv := event.Event{
		Time:         mustParseTime("10:00:00.000"),
		EventID:      1,
		CompetitorID: 1,
	}
	out := callHandle(regEv, comp, cfg)
	require.Empty(t, out)
	require.Equal(t, mustParseTime("10:00:00.000"), comp.RegisteredAt)

	// Плановый старт
	startPlannedEv := event.Event{
		Time:         mustParseTime("10:01:00.000"),
		EventID:      2,
		CompetitorID: 1,
		ExtraParams:  []string{"10:01:00.000"},
	}
	out = callHandle(startPlannedEv, comp, cfg)
	require.Empty(t, out)
	require.Equal(t, mustParseTime("10:01:00.000"), comp.StartPlanned)
	require.Len(t, comp.Laps, 1)
}

func TestHandleEvent_LapFinishAndFinal(t *testing.T) {
	cfg := &config.ParsedConfig{
		Laps:       1,
		LapLen:     3000,
		StartDelta: time.Minute,
	}

	comp := &event.Competitor{
		ID: 1,
		Laps: []event.Lap{
			{Start: mustParseTime("10:00:00.000")},
		},
	}

	ev := event.Event{
		Time:         mustParseTime("10:02:00.000"),
		EventID:      10,
		CompetitorID: 1,
	}
	out := callHandle(ev, comp, cfg)

	require.Len(t, out, 1)
	require.Equal(t, 33, out[0].EventID)
	require.WithinDuration(t, mustParseTime("10:02:00.000"), comp.FinishTime, time.Millisecond)
}

func TestHandleEvent_PenaltyLaps(t *testing.T) {
	cfg := &config.ParsedConfig{}

	comp := &event.Competitor{ID: 1}

	comp.CurrentHits = 3
	out := callHandle(event.Event{
		EventID:      7,
		CompetitorID: 1,
	}, comp, cfg)

	require.Equal(t, 2, comp.PenaltyLaps.Count)
	require.Equal(t, 3, comp.TotalHits)
	require.Equal(t, 0, comp.CurrentHits)
	require.Empty(t, out)
}

func mustParseTime(tStr string) time.Time {
	t, err := time.Parse("15:04:05.000", tStr)
	if err != nil {
		panic(err)
	}
	return t
}

func callHandle(ev event.Event, comp *event.Competitor, cfg *config.ParsedConfig) []event.Event {
	return callEventHandle(ev, comp, cfg)
}

// Шорткат на экспортируемую функцию
var callEventHandle = func(ev event.Event, comp *event.Competitor, cfg *config.ParsedConfig) []event.Event {
	return event.HandleEvent(ev, comp, cfg)
}
