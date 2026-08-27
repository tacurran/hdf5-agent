package main

import "testing"

func TestDirOf(t *testing.T) {
	if got := dirOf("data/sample.h5"); got != "data" {
		t.Fatalf("got %q", got)
	}
	if got := dirOf("sample.h5"); got != "." {
		t.Fatalf("got %q", got)
	}
}
