package hdf5store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(dir, 100000)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if err := store.Ready(ctx); err != nil {
		t.Fatalf("Ready: %v", err)
	}

	sample := filepath.Join(dir, "sample.h5")
	if err := WriteSample(sample); err != nil {
		t.Fatalf("WriteSample: %v", err)
	}

	files, err := store.ListFiles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "sample.h5" {
		t.Fatalf("ListFiles = %#v", files)
	}
	if files[0].SizeBytes <= 0 {
		t.Fatal("expected non-zero size")
	}

	tree, err := store.Structure(ctx, "sample.h5")
	if err != nil {
		t.Fatal(err)
	}
	if tree.Type != "file" || len(tree.Children) != 1 {
		t.Fatalf("structure = %#v", tree)
	}
	group := tree.Children[0]
	if group.Type != "group" || group.Name != "measurements" {
		t.Fatalf("group = %#v", group)
	}
	if len(group.Children) != 2 {
		t.Fatalf("datasets = %#v", group.Children)
	}

	ds, err := store.ReadDataset(ctx, "sample.h5", "/measurements/waveform")
	if err != nil {
		t.Fatal(err)
	}
	if ds.Dtype != "float64" || ds.NPoints != 100 || ds.Truncated {
		t.Fatalf("waveform meta = %#v", ds)
	}
	data, ok := ds.Data.([]float64)
	if !ok || len(data) != 100 {
		t.Fatalf("waveform data type %T len %v", ds.Data, ds.Data)
	}

	matrix, err := store.ReadDataset(ctx, "sample.h5", "measurements/matrix")
	if err != nil {
		t.Fatal(err)
	}
	if matrix.Dtype != "int64" || len(matrix.Shape) != 2 {
		t.Fatalf("matrix = %#v", matrix)
	}

	if err := store.UpdateDataset(ctx, "sample.h5", "/measurements/waveform", []int{0}, []any{42.0}); err != nil {
		t.Fatal(err)
	}
	ds, err = store.ReadDataset(ctx, "sample.h5", "/measurements/waveform")
	if err != nil {
		t.Fatal(err)
	}
	data = ds.Data.([]float64)
	if data[0] != 42.0 {
		t.Fatalf("updated value = %v", data[0])
	}

	if err := store.CreateFile(ctx, "empty.h5"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateFile(ctx, "empty.h5"); err == nil {
		t.Fatal("expected conflict")
	}
	if err := store.DeleteFile(ctx, "empty.h5"); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteFile(ctx, "empty.h5"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestStoreRejectsTraversal(t *testing.T) {
	store, err := Open(t.TempDir(), 10)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := store.Structure(ctx, "../etc.h5"); err == nil {
		t.Fatal("expected invalid name")
	}
	if _, err := store.ReadDataset(ctx, "ok.h5", "/../../etc/passwd"); err == nil {
		t.Fatal("expected invalid path")
	}
}

func TestStoreTruncatesLargeDataset(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteSample(filepath.Join(dir, "sample.h5")); err != nil {
		t.Fatal(err)
	}
	ds, err := store.ReadDataset(context.Background(), "sample.h5", "/measurements/waveform")
	if err != nil {
		t.Fatal(err)
	}
	if !ds.Truncated || ds.Data != nil {
		t.Fatalf("expected truncation, got %#v", ds)
	}
	if err := store.UpdateDataset(context.Background(), "sample.h5", "/measurements/waveform", []int{0}, []any{1}); err == nil {
		t.Fatal("expected too large on update")
	}
}

func TestOpenRejectsBadArgs(t *testing.T) {
	if _, err := Open("", 10); err == nil {
		t.Fatal("expected error")
	}
	file := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(file, 10); err == nil {
		t.Fatal("expected not a directory")
	}
}

func TestMissingFile(t *testing.T) {
	store, err := Open(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Structure(context.Background(), "missing.h5"); err == nil {
		t.Fatal("expected not found")
	}
}
