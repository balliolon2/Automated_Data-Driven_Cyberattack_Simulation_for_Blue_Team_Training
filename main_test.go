package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedScenarioSeedIsAvailable(t *testing.T) {
	if len(scenarioSeedJSON) == 0 {
		t.Fatalf("embedded scenarios_seed.json is empty")
	}

	var scenarios []struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(scenarioSeedJSON, &scenarios); err != nil {
		t.Fatalf("embedded scenarios_seed.json is not valid JSON: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatalf("embedded scenarios_seed.json contains no scenarios")
	}
	for i, scenario := range scenarios {
		if strings.TrimSpace(scenario.Title) == "" {
			t.Fatalf("embedded scenario %d has no title", i)
		}
	}
}

func TestResolveDBPoolConfigBounds(t *testing.T) {
	t.Run("defaults without env", func(t *testing.T) {
		t.Setenv("DB_MAX_OPEN_CONNS", "")
		t.Setenv("DB_MAX_IDLE_CONNS", "")
		cfg := resolveDBPoolConfig()
		if cfg.MaxOpenConns != 50 || cfg.MaxIdleConns != 20 {
			t.Fatalf("got open=%d idle=%d, want 50/20", cfg.MaxOpenConns, cfg.MaxIdleConns)
		}
		if cfg.ConnMaxLifetime != 5*time.Minute || cfg.ConnMaxIdleTime != 2*time.Minute {
			t.Fatalf("unexpected lifetimes: life=%v idle=%v", cfg.ConnMaxLifetime, cfg.ConnMaxIdleTime)
		}
	})

	t.Run("clamps max open to 90 and idle to max open", func(t *testing.T) {
		t.Setenv("DB_MAX_OPEN_CONNS", "150")
		t.Setenv("DB_MAX_IDLE_CONNS", "80")
		cfg := resolveDBPoolConfig()
		if cfg.MaxOpenConns != 90 || cfg.MaxIdleConns != 80 {
			t.Fatalf("got open=%d idle=%d, want 90/80", cfg.MaxOpenConns, cfg.MaxIdleConns)
		}

		t.Setenv("DB_MAX_OPEN_CONNS", "15")
		t.Setenv("DB_MAX_IDLE_CONNS", "30")
		cfg = resolveDBPoolConfig()
		if cfg.MaxOpenConns != 15 || cfg.MaxIdleConns != 15 {
			t.Fatalf("got open=%d idle=%d, want 15/15", cfg.MaxOpenConns, cfg.MaxIdleConns)
		}
	})

	t.Run("ignores non-positive or malformed env values safely", func(t *testing.T) {
		t.Setenv("DB_MAX_OPEN_CONNS", "-5")
		t.Setenv("DB_MAX_IDLE_CONNS", "not-a-number")
		cfg := resolveDBPoolConfig()
		if cfg.MaxOpenConns != 50 || cfg.MaxIdleConns != 20 {
			t.Fatalf("got open=%d idle=%d, want 50/20", cfg.MaxOpenConns, cfg.MaxIdleConns)
		}

		t.Setenv("DB_MAX_OPEN_CONNS", "12")
		t.Setenv("DB_MAX_IDLE_CONNS", "0")
		cfg = resolveDBPoolConfig()
		if cfg.MaxOpenConns != 12 || cfg.MaxIdleConns != 12 {
			t.Fatalf("got open=%d idle=%d, want 12/12", cfg.MaxOpenConns, cfg.MaxIdleConns)
		}
	})
}
