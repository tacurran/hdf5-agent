package hdf5store

import (
	"fmt"
	"math"

	"gonum.org/v1/hdf5"
)

// WriteSample creates a small HDF5 file used by tests and cmd/create-testdata.
// The file contains a measurements group with a float waveform and an integer matrix.
func WriteSample(path string) error {
	f, err := hdf5.CreateFile(path, hdf5.F_ACC_TRUNC)
	if err != nil {
		return fmt.Errorf("create sample file: %w", err)
	}
	defer f.Close()

	grp, err := f.CreateGroup("measurements")
	if err != nil {
		return fmt.Errorf("create group: %w", err)
	}
	defer grp.Close()

	waveform := make([]float64, 100)
	for i := range waveform {
		waveform[i] = math.Sin(float64(i) / 10.0)
	}
	if err := writeFloat64(grp, "waveform", []uint{100}, waveform); err != nil {
		return err
	}

	matrix := make([]int64, 200)
	for i := range matrix {
		matrix[i] = int64(i)
	}
	if err := writeInt64(grp, "matrix", []uint{10, 20}, matrix); err != nil {
		return err
	}
	return nil
}

func writeFloat64(g *hdf5.Group, name string, dims []uint, data []float64) error {
	space, err := hdf5.CreateSimpleDataspace(dims, nil)
	if err != nil {
		return fmt.Errorf("dataspace %s: %w", name, err)
	}
	defer space.Close()
	dset, err := g.CreateDataset(name, hdf5.T_NATIVE_DOUBLE, space)
	if err != nil {
		return fmt.Errorf("dataset %s: %w", name, err)
	}
	defer dset.Close()
	if err := dset.Write(&data); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

func writeInt64(g *hdf5.Group, name string, dims []uint, data []int64) error {
	space, err := hdf5.CreateSimpleDataspace(dims, nil)
	if err != nil {
		return fmt.Errorf("dataspace %s: %w", name, err)
	}
	defer space.Close()
	dset, err := g.CreateDataset(name, hdf5.T_NATIVE_INT64, space)
	if err != nil {
		return fmt.Errorf("dataset %s: %w", name, err)
	}
	defer dset.Close()
	if err := dset.Write(&data); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}
