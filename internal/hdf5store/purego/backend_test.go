package purego

import (
	"math"
	"testing"
)

// testdata/sample.h5 is written by the cgo gonum/C path (cmd/create-testdata /
// hdf5store.WriteSample). Reading it here with the pure-Go backend is the core
// interop proof: a cgo-free reader consuming a C-library-written file. This
// test builds and runs with CGO_ENABLED=0.

const fixture = "testdata/sample.h5"

func TestStructure(t *testing.T) {
	root, err := Backend{}.Structure(fixture, "sample.h5")
	if err != nil {
		t.Fatalf("Structure: %v", err)
	}
	if root.Type != "file" || root.Name != "sample.h5" {
		t.Fatalf("unexpected root: %+v", root)
	}
	if len(root.Children) != 1 {
		t.Fatalf("want 1 top-level child, got %d", len(root.Children))
	}
	grp := root.Children[0]
	if grp.Type != "group" || grp.Name != "measurements" || grp.Path != "/measurements" {
		t.Fatalf("unexpected group node: %+v", grp)
	}
	if len(grp.Children) != 2 {
		t.Fatalf("want 2 datasets under measurements, got %d", len(grp.Children))
	}

	byName := map[string]*Node{}
	for _, c := range grp.Children {
		byName[c.Name] = c
	}

	mat := byName["matrix"]
	if mat == nil {
		t.Fatal("matrix dataset missing")
	}
	if mat.Type != "dataset" || mat.Path != "/measurements/matrix" {
		t.Fatalf("unexpected matrix node: %+v", mat)
	}
	if mat.Dtype != "int64" {
		t.Errorf("matrix dtype = %q, want int64", mat.Dtype)
	}
	if len(mat.Shape) != 2 || mat.Shape[0] != 10 || mat.Shape[1] != 20 {
		t.Errorf("matrix shape = %v, want [10 20]", mat.Shape)
	}
	if mat.NPoints != 200 {
		t.Errorf("matrix npoints = %d, want 200", mat.NPoints)
	}

	wav := byName["waveform"]
	if wav == nil {
		t.Fatal("waveform dataset missing")
	}
	if wav.Dtype != "float64" {
		t.Errorf("waveform dtype = %q, want float64", wav.Dtype)
	}
	if len(wav.Shape) != 1 || wav.Shape[0] != 100 {
		t.Errorf("waveform shape = %v, want [100]", wav.Shape)
	}
	if wav.NPoints != 100 {
		t.Errorf("waveform npoints = %d, want 100", wav.NPoints)
	}
}

func TestReadWaveform(t *testing.T) {
	ds, err := Backend{}.ReadDataset(fixture, "/measurements/waveform", 100000)
	if err != nil {
		t.Fatalf("ReadDataset: %v", err)
	}
	if ds.Truncated {
		t.Fatal("unexpected truncation")
	}
	data, ok := ds.Data.([]float64)
	if !ok {
		t.Fatalf("data type = %T, want []float64", ds.Data)
	}
	if len(data) != 100 {
		t.Fatalf("len(data) = %d, want 100", len(data))
	}
	// WriteSample stores waveform[i] = sin(i/10).
	for _, i := range []int{0, 1, 50, 99} {
		want := math.Sin(float64(i) / 10.0)
		if math.Abs(data[i]-want) > 1e-12 {
			t.Errorf("waveform[%d] = %v, want %v", i, data[i], want)
		}
	}
}

func TestReadMatrix(t *testing.T) {
	ds, err := Backend{}.ReadDataset(fixture, "/measurements/matrix", 100000)
	if err != nil {
		t.Fatalf("ReadDataset: %v", err)
	}
	if ds.Dtype != "int64" || len(ds.Shape) != 2 {
		t.Fatalf("unexpected metadata: %+v", ds)
	}
	data, ok := ds.Data.([]float64) // scigolib returns []float64 for all numerics
	if !ok {
		t.Fatalf("data type = %T, want []float64", ds.Data)
	}
	if len(data) != 200 {
		t.Fatalf("len(data) = %d, want 200", len(data))
	}
	// WriteSample stores matrix[i] = i (row-major).
	for _, i := range []int{0, 1, 199} {
		if data[i] != float64(i) {
			t.Errorf("matrix[%d] = %v, want %d", i, data[i], i)
		}
	}
}

func TestReadDatasetTruncates(t *testing.T) {
	ds, err := Backend{}.ReadDataset(fixture, "/measurements/matrix", 10)
	if err != nil {
		t.Fatalf("ReadDataset: %v", err)
	}
	if !ds.Truncated {
		t.Error("expected truncation when npoints (200) exceeds maxPoints (10)")
	}
	if ds.Data != nil {
		t.Errorf("expected nil data when truncated, got %v", ds.Data)
	}
}

func TestReadDatasetNotFound(t *testing.T) {
	b := Backend{}
	if _, err := b.ReadDataset(fixture, "/measurements/missing", 100000); err == nil {
		t.Fatal("expected error for missing dataset")
	}
}
