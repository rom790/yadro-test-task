package config_test

import (
	configPkg "github.com/rom790/yadro-test-task/internal/config"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestProcessConfig_ValidConfig(t *testing.T) {
	// Создаём временный файл
	content := `{
		"laps": 2,
		"lapLen": 3500,
		"penaltyLen": 150,
		"firingLines": 2,
		"start": "10:00:00.000",
		"startDelta": "00:01:30"
	}`
	tmpFile, err := os.CreateTemp("", "config_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(content))
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := configPkg.ProcessConfig(tmpFile.Name())
	require.NoError(t, err)

	require.Equal(t, 2, cfg.Laps)
	require.Equal(t, 3500, cfg.LapLen)
	require.Equal(t, 150, cfg.PenaltyLen)
	require.Equal(t, 2, cfg.FiringLines)
	require.Equal(t, "10:00:00.000", cfg.Start.Format("15:04:05.000"))
	require.Equal(t, 90*time.Second, cfg.StartDelta)
}

func TestProcessConfig_InvalidPath(t *testing.T) {
	_, err := configPkg.ProcessConfig("nonexistent.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "open config file error")
}

func TestProcessConfig_InvalidJSON(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "invalid_json_*.json")
	defer os.Remove(tmpFile.Name())

	tmpFile.Write([]byte(`{invalid json}`))
	tmpFile.Close()

	_, err := configPkg.ProcessConfig(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "read config error")
}

func TestProcessConfig_InvalidStartFormat(t *testing.T) {
	content := `{
		"laps": 2,
		"lapLen": 3500,
		"penaltyLen": 150,
		"firingLines": 2,
		"start": "bad format",
		"startDelta": "00:01:30"
	}`
	tmpFile, _ := os.CreateTemp("", "invalid_start_*.json")
	defer os.Remove(tmpFile.Name())

	tmpFile.Write([]byte(content))
	tmpFile.Close()

	_, err := configPkg.ProcessConfig(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid start time format")
}

func TestProcessConfig_InvalidStartDelta(t *testing.T) {
	content := `{
		"laps": 2,
		"lapLen": 3500,
		"penaltyLen": 150,
		"firingLines": 2,
		"start": "10:00:00.000",
		"startDelta": "bad delta"
	}`
	tmpFile, _ := os.CreateTemp("", "invalid_delta_*.json")
	defer os.Remove(tmpFile.Name())

	tmpFile.Write([]byte(content))
	tmpFile.Close()

	_, err := configPkg.ProcessConfig(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid start delta format")
}
