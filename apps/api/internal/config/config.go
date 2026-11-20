// Package config loads and validates process configuration from environment
// variables. Every variable is validated and all problems are reported
// together, so a misconfigured deployment fails fast with one actionable
// message instead of one error per restart.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Environment names the deployment tier.
type Environment string

// Supported deployment tiers.
const (
	EnvDevelopment Environment = "development"
	EnvTest        Environment = "test"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// Config is the validated configuration of the API process.
type Config struct {
	// Env is the deployment tier. It has no default on purpose: a deployment
	// that forgets to set it must fail instead of running with development
	// settings.
	Env Environment
	// HTTPAddr is the listen address, for example ":8080".
	HTTPAddr string
	// LogLevel is the minimum level written to the structured log.
	LogLevel slog.Level
	// DrainDelay is how long the process reports "not ready" before it stops
	// accepting connections, so load balancers can stop routing to it.
	DrainDelay time.Duration
	// ShutdownTimeout bounds how long in-flight requests may take to finish.
	ShutdownTimeout time.Duration
	// MaxBodyBytes is the largest request body the server will read.
	MaxBodyBytes int64
}

const (
	defaultHTTPAddr        = ":8080"
	defaultDrainDelay      = 5 * time.Second
	defaultShutdownTimeout = 25 * time.Second
	defaultMaxBodyBytes    = 1 << 20
	maxAllowedBodyBytes    = 64 << 20
)

// Load reads and validates configuration. getenv is os.Getenv in production
// and a map lookup in tests.
func Load(getenv func(string) string) (Config, error) {
	var problems []error

	cfg := Config{
		HTTPAddr:        defaultHTTPAddr,
		LogLevel:        slog.LevelInfo,
		DrainDelay:      defaultDrainDelay,
		ShutdownTimeout: defaultShutdownTimeout,
		MaxBodyBytes:    defaultMaxBodyBytes,
	}

	switch env := Environment(strings.ToLower(strings.TrimSpace(getenv("BRYNN_ENV")))); env {
	case EnvDevelopment, EnvTest, EnvStaging, EnvProduction:
		cfg.Env = env
	case "":
		problems = append(problems, errors.New("BRYNN_ENV is required (development, test, staging or production)"))
	default:
		problems = append(problems, fmt.Errorf("BRYNN_ENV %q is not one of development, test, staging, production", string(env)))
	}

	if v := strings.TrimSpace(getenv("BRYNN_HTTP_ADDR")); v != "" {
		cfg.HTTPAddr = v
	}

	if v := strings.TrimSpace(getenv("BRYNN_LOG_LEVEL")); v != "" {
		var lvl slog.Level
		if err := lvl.UnmarshalText([]byte(v)); err != nil {
			problems = append(problems, fmt.Errorf("BRYNN_LOG_LEVEL %q is invalid (debug, info, warn or error)", v))
		} else {
			cfg.LogLevel = lvl
		}
	}

	cfg.DrainDelay = durationVar(getenv, "BRYNN_DRAIN_DELAY", cfg.DrainDelay, 0, time.Minute, &problems)
	cfg.ShutdownTimeout = durationVar(getenv, "BRYNN_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout, time.Second, 5*time.Minute, &problems)

	if v := strings.TrimSpace(getenv("BRYNN_MAX_BODY_BYTES")); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 || n > maxAllowedBodyBytes {
			problems = append(problems, fmt.Errorf("BRYNN_MAX_BODY_BYTES %q must be an integer between 1 and %d", v, maxAllowedBodyBytes))
		} else {
			cfg.MaxBodyBytes = n
		}
	}

	if err := errors.Join(problems...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// durationVar parses a duration variable, falling back to def when it is unset
// and recording a problem when it is malformed or outside [lo, hi].
func durationVar(getenv func(string) string, name string, def, lo, hi time.Duration, problems *[]error) time.Duration {
	raw := strings.TrimSpace(getenv(name))
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < lo || d > hi {
		*problems = append(*problems, fmt.Errorf("%s %q must be a duration between %s and %s (for example 5s)", name, raw, lo, hi))
		return def
	}
	return d
}
