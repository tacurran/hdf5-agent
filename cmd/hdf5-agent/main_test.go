package main

import (
	"testing"

	"github.com/tacurran/hdf5-agent/internal/config"
)

func TestNewLogger(t *testing.T) {
	log := newLogger(config.Config{LogLevel: "debug", LogFormat: "text"})
	if log == nil {
		t.Fatal("nil logger")
	}
	if newLogger(config.Config{LogLevel: "error", LogFormat: "json"}) == nil {
		t.Fatal("json logger")
	}
}
