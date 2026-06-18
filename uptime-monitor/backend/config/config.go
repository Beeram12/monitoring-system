package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, sourced from a .env file (or the
// process environment, which takes precedence if a variable is already set).
type Config struct {
	Port          string
	DatabaseURL   string
	PingTimeout   time.Duration
	SchedulerPoll time.Duration
	AllowedOrigin string
}

// Load reads .env (if present) into the process environment and builds a Config
// from the required variables. It fails fast if a required variable is missing.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	port, err := requireEnv("PORT")
	if err != nil {
		return Config{}, err
	}
	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	pingTimeout, err := requireEnvDuration("PING_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	schedulerPoll, err := requireEnvDuration("SCHEDULER_POLL_INTERVAL")
	if err != nil {
		return Config{}, err
	}
	allowedOrigin, err := requireEnv("ALLOWED_ORIGIN")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:          port,
		DatabaseURL:   databaseURL,
		PingTimeout:   pingTimeout,
		SchedulerPoll: schedulerPoll,
		AllowedOrigin: allowedOrigin,
	}, nil
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", key)
	}
	return v, nil
}

func requireEnvDuration(key string) (time.Duration, error) {
	v, err := requireEnv(key)
	if err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return d, nil
}
