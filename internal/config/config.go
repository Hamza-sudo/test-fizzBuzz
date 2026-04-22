package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr       = ":8080"
	defaultMaxLimit   = 10000
	defaultStatsDBDSN = "file:fizzbuzz_stats.db"
	defaultDBTimeout  = 200 * time.Millisecond
)

// Config groups runtime configuration loaded from environment variables.
type Config struct {
	Addr           string
	MaxLimit       int
	StatsDBDSN     string
	StatsDBTimeout time.Duration
}

// Warning describes a non-fatal configuration issue.
type Warning struct {
	Field   string
	Message string
	Value   string
}

// LoadFromEnv reads configuration from environment variables and applies defaults.
// It returns warnings when invalid non-critical values are provided.
func LoadFromEnv() (Config, []Warning) {
	warnings := make([]Warning, 0, 2)

	maxLimit, maxLimitWarning := serverMaxLimit()
	if maxLimitWarning != nil {
		warnings = append(warnings, *maxLimitWarning)
	}

	dbTimeout, timeoutWarning := statsDBTimeout()
	if timeoutWarning != nil {
		warnings = append(warnings, *timeoutWarning)
	}

	cfg := Config{
		Addr:           serverAddr(),
		MaxLimit:       maxLimit,
		StatsDBDSN:     statsDBDSN(),
		StatsDBTimeout: dbTimeout,
	}

	return cfg, warnings
}

func statsDBDSN() string {
	if dsn := os.Getenv("STATS_DB_DSN"); dsn != "" {
		return dsn
	}

	return defaultStatsDBDSN
}

func statsDBTimeout() (time.Duration, *Warning) {
	value := os.Getenv("STATS_DB_TIMEOUT_MS")
	if value == "" {
		return defaultDBTimeout, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultDBTimeout, &Warning{
			Field:   "STATS_DB_TIMEOUT_MS",
			Message: "invalid value, falling back to default milliseconds timeout",
			Value:   value,
		}
	}

	timeout := time.Duration(parsed) * time.Millisecond
	return timeout, nil
}

func serverAddr() string {
	if value := strings.TrimSpace(os.Getenv("PORT")); value != "" {
		// Accept either a raw port ("8080") or a full listen address
		// (":8080", "localhost:8080", "0.0.0.0:8080").
		if strings.Contains(value, ":") {
			return value
		}
		return ":" + value
	}

	// Default to a conventional local development port.
	return defaultAddr
}

func serverMaxLimit() (int, *Warning) {
	value := os.Getenv("MAX_LIMIT")
	if value == "" {
		return defaultMaxLimit, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultMaxLimit, &Warning{
			Field:   "MAX_LIMIT",
			Message: "invalid value, falling back to default maximum limit",
			Value:   value,
		}
	}

	return parsed, nil
}
