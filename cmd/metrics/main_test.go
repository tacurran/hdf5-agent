package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnforceGates(t *testing.T) {
	if err := enforceGates(map[string]any{"python_files": 1, "go_test_functions": 10, "github_actions_workflows": 1, "openapi_specs": 1}); err == nil {
		t.Fatal("expected python gate")
	}
	if err := enforceGates(map[string]any{"python_files": 0, "go_test_functions": 0, "github_actions_workflows": 1, "openapi_specs": 1}); err == nil {
		t.Fatal("expected test gate")
	}
	if err := enforceGates(map[string]any{"python_files": 0, "go_test_functions": 3, "github_actions_workflows": 1, "openapi_specs": 1}); err != nil {
		t.Fatal(err)
	}
}

func TestCoveragePercent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.out")
	content := "mode: set\nfile.go:1.1,2.2 2 1\nfile.go:3.1,4.2 2 0\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := coveragePercent(path)
	if got != 50 {
		t.Fatalf("got %v", got)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	s := snapshot{SchemaVersion: 1, Metrics: map[string]any{"go_test_files": 1}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var out snapshot
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.SchemaVersion != 1 {
		t.Fatal(out)
	}
}
