// Command create-testdata writes a sample HDF5 file for local exploration.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tacurran/hdf5-agent/internal/hdf5store"
)

func main() {
	out := flag.String("out", "data/sample.h5", "output HDF5 path")
	flag.Parse()
	if err := os.MkdirAll(dirOf(*out), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := hdf5store.WriteSample(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", *out)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
