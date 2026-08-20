package main

import (
	"testing"

	"github.com/tacurran/hdf5-agent/internal/hdf5store"
)

func TestCollectDatasets(t *testing.T) {
	tree := &hdf5store.Node{
		Name: "a.h5",
		Type: "file",
		Children: []*hdf5store.Node{
			{
				Name: "g",
				Type: "group",
				Path: "/g",
				Children: []*hdf5store.Node{
					{Name: "ds", Type: "dataset", Path: "/g/ds", Dtype: "float64", Shape: []uint{3}, NPoints: 3},
				},
			},
		},
	}
	var out []inventoryDataset
	collectDatasets(tree, &out)
	if len(out) != 1 || out[0].Path != "/g/ds" {
		t.Fatalf("%#v", out)
	}
}
