// Command hdf5-agent is the HDF5 Agent HTTP service.
//
// Configuration is exclusively via environment variables (see internal/config).
// The process serves a versioned JSON API for other services and, when STATIC_DIR
// is set, the React file browser.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tacurran/hdf5-agent/internal/config"
	"github.com/tacurran/hdf5-agent/internal/hdf5store"
	"github.com/tacurran/hdf5-agent/internal/httpserver"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := newLogger(cfg)
	store, err := hdf5store.Open(cfg.DataDir, cfg.MaxDatasetPoints)
	if err != nil {
		log.Error("open store", "err", err)
		os.Exit(1)
	}
	srv := httpserver.New(cfg, store, log)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}

func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if cfg.LogFormat == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h)
}
