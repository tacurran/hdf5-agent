// Package config loads HDF5 Agent runtime settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the process configuration for the HDF5 Agent HTTP service.
// All fields are populated from environment variables; there are no secrets.
type Config struct {
	// HTTPAddr is the listen address, for example ":8080".
	HTTPAddr string
	// DataDir is the directory that contains HDF5 files served by the API.
	DataDir string
	// StaticDir is the optional directory of built frontend assets. Empty disables static serving.
	StaticDir string
	// CORSOrigins is the list of allowed Origin values. A single "*" allows any origin.
	CORSOrigins []string
	// ReadTimeout is the HTTP server read timeout, including the request body.
	ReadTimeout time.Duration
	// WriteTimeout is the HTTP server write timeout.
	WriteTimeout time.Duration
	// IdleTimeout is the HTTP keep-alive idle timeout.
	IdleTimeout time.Duration
	// ShutdownTimeout is how long graceful shutdown waits for in-flight requests.
	ShutdownTimeout time.Duration
	// MaxDatasetPoints is the maximum number of values returned from a dataset read.
	MaxDatasetPoints int
	// MaxRequestBytes is the maximum accepted JSON request body size.
	MaxRequestBytes int64
	// LogLevel is slog level name: debug, info, warn, or error.
	LogLevel string
	// LogFormat is "json" or "text".
	LogFormat string
}

// FromEnv loads Config from process environment variables.
//
// Supported variables:
//
//	HTTP_ADDR            listen address (default :8080); PORT is used if HTTP_ADDR is unset
//	HDF5_DATA_DIR        directory of HDF5 files (default ./data)
//	STATIC_DIR           built frontend directory (default ./public)
//	CORS_ORIGINS         comma-separated origins (default *)
//	READ_TIMEOUT         duration (default 15s)
//	WRITE_TIMEOUT        duration (default 60s)
//	IDLE_TIMEOUT         duration (default 60s)
//	SHUTDOWN_TIMEOUT     duration (default 20s)
//	MAX_DATASET_POINTS   integer (default 100000)
//	MAX_REQUEST_BYTES    integer (default 1048576)
//	LOG_LEVEL            debug|info|warn|error (default info)
//	LOG_FORMAT           json|text (default json)
func FromEnv() (Config, error) {
	addr := getenv("HTTP_ADDR", "")
	if addr == "" {
		port := getenv("PORT", "8080")
		addr = ":" + port
	}

	cfg := Config{
		HTTPAddr:         addr,
		DataDir:          getenv("HDF5_DATA_DIR", "./data"),
		StaticDir:        getenv("STATIC_DIR", "./public"),
		CORSOrigins:      splitList(getenv("CORS_ORIGINS", "*")),
		LogLevel:         strings.ToLower(getenv("LOG_LEVEL", "info")),
		LogFormat:        strings.ToLower(getenv("LOG_FORMAT", "json")),
		MaxDatasetPoints: 100000,
		MaxRequestBytes:  1 << 20,
	}

	var err error
	if cfg.ReadTimeout, err = parseDuration("READ_TIMEOUT", "15s"); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = parseDuration("WRITE_TIMEOUT", "60s"); err != nil {
		return Config{}, err
	}
	if cfg.IdleTimeout, err = parseDuration("IDLE_TIMEOUT", "60s"); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = parseDuration("SHUTDOWN_TIMEOUT", "20s"); err != nil {
		return Config{}, err
	}
	if cfg.MaxDatasetPoints, err = parseInt("MAX_DATASET_POINTS", 100000); err != nil {
		return Config{}, err
	}
	if cfg.MaxRequestBytes, err = parseInt64("MAX_REQUEST_BYTES", 1<<20); err != nil {
		return Config{}, err
	}
	if cfg.MaxDatasetPoints <= 0 {
		return Config{}, fmt.Errorf("MAX_DATASET_POINTS must be positive")
	}
	if cfg.MaxRequestBytes <= 0 {
		return Config{}, fmt.Errorf("MAX_REQUEST_BYTES must be positive")
	}
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error")
	}
	switch cfg.LogFormat {
	case "json", "text":
	default:
		return Config{}, fmt.Errorf("LOG_FORMAT must be json or text")
	}
	if cfg.DataDir == "" {
		return Config{}, fmt.Errorf("HDF5_DATA_DIR must not be empty")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func parseDuration(key, fallback string) (time.Duration, error) {
	raw := getenv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return d, nil
}

func parseInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, raw, err)
	}
	return n, nil
}

func parseInt64(key string, fallback int64) (int64, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, raw, err)
	}
	return n, nil
}
