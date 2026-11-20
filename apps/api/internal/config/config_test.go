package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

// env builds a getenv function from key/value pairs.
func env(pairs ...string) func(string) string {
	m := make(map[string]string, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return func(key string) string { return m[key] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env("BRYNN_ENV", "development"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != EnvDevelopment {
		t.Errorf("Env = %q, want %q", cfg.Env, EnvDevelopment)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
	if cfg.DrainDelay != 5*time.Second {
		t.Errorf("DrainDelay = %v, want 5s", cfg.DrainDelay)
	}
	if cfg.ShutdownTimeout != 25*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 25s", cfg.ShutdownTimeout)
	}
	if cfg.MaxBodyBytes != 1<<20 {
		t.Errorf("MaxBodyBytes = %d, want %d", cfg.MaxBodyBytes, 1<<20)
	}
}

func TestLoadOverrides(t *testing.T) {
	cfg, err := Load(env(
		"BRYNN_ENV", "Production",
		"BRYNN_HTTP_ADDR", "127.0.0.1:9000",
		"BRYNN_LOG_LEVEL", "debug",
		"BRYNN_DRAIN_DELAY", "0s",
		"BRYNN_SHUTDOWN_TIMEOUT", "10s",
		"BRYNN_MAX_BODY_BYTES", "2048",
	))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != EnvProduction {
		t.Errorf("Env = %q, want %q (matching must be case-insensitive)", cfg.Env, EnvProduction)
	}
	if cfg.HTTPAddr != "127.0.0.1:9000" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.DrainDelay != 0 {
		t.Errorf("DrainDelay = %v, want 0", cfg.DrainDelay)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 10s", cfg.ShutdownTimeout)
	}
	if cfg.MaxBodyBytes != 2048 {
		t.Errorf("MaxBodyBytes = %d, want 2048", cfg.MaxBodyBytes)
	}
}

func TestLoadRequiresEnvironment(t *testing.T) {
	_, err := Load(env())
	if err == nil || !strings.Contains(err.Error(), "BRYNN_ENV is required") {
		t.Fatalf("Load() error = %v, want a \"BRYNN_ENV is required\" error", err)
	}
}

func TestLoadReportsEveryProblem(t *testing.T) {
	_, err := Load(env(
		"BRYNN_ENV", "prod-ish",
		"BRYNN_LOG_LEVEL", "loud",
		"BRYNN_SHUTDOWN_TIMEOUT", "abc",
		"BRYNN_MAX_BODY_BYTES", "0",
	))
	if err == nil {
		t.Fatal("Load() succeeded, want an error")
	}
	for _, name := range []string{"BRYNN_ENV", "BRYNN_LOG_LEVEL", "BRYNN_SHUTDOWN_TIMEOUT", "BRYNN_MAX_BODY_BYTES"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not mention %s: %v", name, err)
		}
	}
}
