package version

import "testing"

func TestConstants(t *testing.T) {
	if Name != "hdf5-agent" {
		t.Fatalf("Name=%q", Name)
	}
	if API != "v1" {
		t.Fatalf("API=%q", API)
	}
	if Version == "" {
		t.Fatal("empty Version")
	}
}
