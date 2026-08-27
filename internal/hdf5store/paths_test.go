package hdf5store

import "testing"

func TestValidateFileName(t *testing.T) {
	ok := []string{"a.h5", "data.hdf5", "run-01.h5", "A_b.9.h5"}
	for _, name := range ok {
		if err := ValidateFileName(name); err != nil {
			t.Errorf("ValidateFileName(%q) unexpected error: %v", name, err)
		}
	}
	bad := []string{"", "../x.h5", "/tmp/x.h5", "x.txt", ".h5", "foo/bar.h5", "x.h5.bak", "..h5"}
	for _, name := range bad {
		if err := ValidateFileName(name); err == nil {
			t.Errorf("ValidateFileName(%q) expected error", name)
		}
	}
}

func TestNormalizeDatasetPath(t *testing.T) {
	got, err := NormalizeDatasetPath("measurements/waveform")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/measurements/waveform" {
		t.Errorf("got %q", got)
	}
	got, err = NormalizeDatasetPath("/measurements/../secret")
	if err == nil {
		t.Fatalf("expected error, got %q", got)
	}
	if _, err := NormalizeDatasetPath(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestJoinHDF5(t *testing.T) {
	if got := joinHDF5("/", "g"); got != "/g" {
		t.Errorf("root join = %q", got)
	}
	if got := joinHDF5("/g", "ds"); got != "/g/ds" {
		t.Errorf("nested join = %q", got)
	}
}

func TestDatatypeName(t *testing.T) {
	if got := datatypeName(0, 8); got != "int64" {
		t.Errorf("int64 name = %q", got)
	}
	if got := datatypeName(1, 8); got != "float64" {
		t.Errorf("float64 name = %q", got)
	}
	if got := datatypeName(3, 8); got != "string" {
		t.Errorf("string name = %q", got)
	}
}

func TestUpdateRequestValuesOrData(t *testing.T) {
	r := UpdateRequest{Values: []any{1}}
	if len(r.ValuesOrData()) != 1 {
		t.Fatal("expected values")
	}
	r = UpdateRequest{Data: []any{2, 3}}
	if len(r.ValuesOrData()) != 2 {
		t.Fatal("expected data alias")
	}
}

func TestAsNumberConversion(t *testing.T) {
	f, err := asFloat64("1.5")
	if err != nil || f != 1.5 {
		t.Fatalf("asFloat64 string: %v %v", f, err)
	}
	i, err := asInt64(float64(9))
	if err != nil || i != 9 {
		t.Fatalf("asInt64 float: %v %v", i, err)
	}
}
