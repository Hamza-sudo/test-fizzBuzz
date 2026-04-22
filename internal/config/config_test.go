package config

import (
	"reflect"
	"testing"
	"time"
)

func TestLoadFromEnvWithoutWarnings(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantCfg Config
	}{
		{
			name: "defaults",
			env: map[string]string{
				"PORT":                "",
				"MAX_LIMIT":           "",
				"STATS_DB_DSN":        "",
				"STATS_DB_TIMEOUT_MS": "",
			},
			wantCfg: Config{
				Addr:           defaultAddr,
				MaxLimit:       defaultMaxLimit,
				StatsDBDSN:     defaultStatsDBDSN,
				StatsDBTimeout: defaultDBTimeout,
			},
		},
		{
			name: "custom values",
			env: map[string]string{
				"PORT":                "9090",
				"MAX_LIMIT":           "123",
				"STATS_DB_DSN":        "file:test.db",
				"STATS_DB_TIMEOUT_MS": "500",
			},
			wantCfg: Config{
				Addr:           ":9090",
				MaxLimit:       123,
				StatsDBDSN:     "file:test.db",
				StatsDBTimeout: 500 * time.Millisecond,
			},
		},
		{
			name: "full listen address",
			env: map[string]string{
				"PORT":                "localhost:8080",
				"MAX_LIMIT":           "",
				"STATS_DB_DSN":        "",
				"STATS_DB_TIMEOUT_MS": "",
			},
			wantCfg: Config{
				Addr:           "localhost:8080",
				MaxLimit:       defaultMaxLimit,
				StatsDBDSN:     defaultStatsDBDSN,
				StatsDBTimeout: defaultDBTimeout,
			},
		},
		{
			name: "colon prefixed port",
			env: map[string]string{
				"PORT":                ":9091",
				"MAX_LIMIT":           "",
				"STATS_DB_DSN":        "",
				"STATS_DB_TIMEOUT_MS": "",
			},
			wantCfg: Config{
				Addr:           ":9091",
				MaxLimit:       defaultMaxLimit,
				StatsDBDSN:     defaultStatsDBDSN,
				StatsDBTimeout: defaultDBTimeout,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)

			cfg, warnings := LoadFromEnv()
			assertNoWarnings(t, warnings)
			assertConfigEqual(t, cfg, tt.wantCfg)
		})
	}
}

func TestLoadInvalidNumericValuesFallbackToDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		"PORT":                "",
		"MAX_LIMIT":           "abc",
		"STATS_DB_DSN":        "",
		"STATS_DB_TIMEOUT_MS": "-1",
	})

	cfg, warnings := LoadFromEnv()
	assertConfigEqual(t, cfg, Config{
		Addr:           defaultAddr,
		MaxLimit:       defaultMaxLimit,
		StatsDBDSN:     defaultStatsDBDSN,
		StatsDBTimeout: defaultDBTimeout,
	})
	assertWarningsEqual(t, warnings, []Warning{
		{Field: "MAX_LIMIT", Message: "invalid value, falling back to default maximum limit", Value: "abc"},
		{Field: "STATS_DB_TIMEOUT_MS", Message: "invalid value, falling back to default milliseconds timeout", Value: "-1"},
	})
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for key, value := range env {
		t.Setenv(key, value)
	}
}

func assertNoWarnings(t *testing.T, got []Warning) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("expected no warnings, got %d: %+v", len(got), got)
	}
}

func assertConfigEqual(t *testing.T, got Config, want Config) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected config:\n got: %+v\nwant: %+v", got, want)
	}
}

func assertWarningsEqual(t *testing.T, got []Warning, want []Warning) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected warnings:\n got: %+v\nwant: %+v", got, want)
	}
}
