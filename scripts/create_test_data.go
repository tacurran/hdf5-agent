package main

import (
	"fmt"
	"log"
	"math"

	"gonum.org/v1/hdf5"
)

func createTestData() error {
	// Create a new HDF5 file
	f, err := hdf5.CreateFile("test_data.h5", hdf5.F_ACC_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Create a group
	grp, err := f.CreateGroup("measurements")
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	defer grp.Close()

	// Create a 1D float array
	dims := []uint{1000}
	float_data := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		float_data[i] = math.Sin(float64(i) / 100.0)
	}

	space, err := hdf5.NewSimpleDataspace(dims, nil)
	if err != nil {
		return fmt.Errorf("failed to create dataspace: %w", err)
	}
	defer space.Close()

	dtype, err := hdf5.NewDatatypeFromType(float_data)
	if err != nil {
		return fmt.Errorf("failed to create datatype: %w", err)
	}
	defer dtype.Close()

	dset, err := grp.CreateDataset("waveform", dtype, space)
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}
	defer dset.Close()

	dset.Write(&float_data)

	// Create a 2D integer array
	dims2d := []uint{10, 20}
	int_data := make([]int64, 200)
	for i := 0; i < 200; i++ {
		int_data[i] = int64(i)
	}

	space2d, err := hdf5.NewSimpleDataspace(dims2d, nil)
	if err != nil {
		return fmt.Errorf("failed to create 2D dataspace: %w", err)
	}
	defer space2d.Close()

	dtype2d, err := hdf5.NewDatatypeFromType(int_data)
	if err != nil {
		return fmt.Errorf("failed to create 2D datatype: %w", err)
	}
	defer dtype2d.Close()

	dset2d, err := grp.CreateDataset("matrix", dtype2d, space2d)
	if err != nil {
		return fmt.Errorf("failed to create 2D dataset: %w", err)
	}
	defer dset2d.Close()

	dset2d.Write(&int_data)

	// Create string array
	strs := []string{"alpha", "beta", "gamma", "delta"}
	dset_str, err := grp.CreateDataset("labels", hdf5.S_Go_String, hdf5.NewSimpleDataspace([]uint{uint(len(strs))}, nil))
	if err != nil {
		return fmt.Errorf("failed to create string dataset: %w", err)
	}
	defer dset_str.Close()

	dset_str.Write(&strs)

	fmt.Println("✓ Created test_data.h5 with sample datasets")
	return nil
}

func main() {
	if err := createTestData(); err != nil {
		log.Fatal(err)
	}
}
