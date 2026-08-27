package config

import (
	"testing"
	"time"
)

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("PORT", "")
	t.Setenv("HDF5_DATA_DIR", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")
	t.Setenv("IDLE_TIMEOUT", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("MAX_DATASET_POINTS", "")
	t.Setenv("MAX_REQUEST_BYTES", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("DataDir = %q", cfg.DataDir)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout = %s", cfg.ReadTimeout)
	}
	if got := cfg.CORSOrigins; len(got) != 1 || got[0] != "*" {
		t.Errorf("CORSOrigins = %#v", got)
	}
}

func TestFromEnvPortAndOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("HDF5_DATA_DIR", "/var/lib/hdf5")
	t.Setenv("CORS_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("MAX_DATASET_POINTS", "50")
	t.Setenv("LOG_FORMAT", "text")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.DataDir != "/var/lib/hdf5" {
		t.Errorf("DataDir = %q", cfg.DataDir)
	}
	if cfg.MaxDatasetPoints != 50 {
		t.Errorf("MaxDatasetPoints = %d", cfg.MaxDatasetPoints)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q", cfg.LogFormat)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("CORSOrigins = %#v", cfg.CORSOrigins)
	}
}

func TestFromEnvHTTPAddrWinsOverPort(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:7777")
	t.Setenv("PORT", "9090")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:7777" {
		t.Errorf("HTTPAddr = %q", cfg.HTTPAddr)
	}
}

func TestFromEnvRejectsBadValues(t *testing.T) {
	t.Setenv("LOG_LEVEL", "verbose")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected error for bad LOG_LEVEL")
	}
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("READ_TIMEOUT", "not-a-duration")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected error for bad READ_TIMEOUT")
	}
	t.Setenv("READ_TIMEOUT", "15s")
	t.Setenv("MAX_DATASET_POINTS", "0")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected error for MAX_DATASET_POINTS=0")
	}
}
