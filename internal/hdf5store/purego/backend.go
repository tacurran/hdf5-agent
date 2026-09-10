// Package purego is a spike: a cgo-free HDF5 read backend built on the pure-Go
// github.com/scigolib/hdf5 library, evaluated as a replacement for the cgo
// gonum.org/v1/hdf5 backend used by internal/hdf5store.
//
// It intentionally lives in its own package so it compiles and tests with
// CGO_ENABLED=0 (importing internal/hdf5store would pull in the cgo gonum
// dependency and defeat the purpose). The Node/Dataset types here mirror the
// DTOs in internal/hdf5store; productionizing the swap would extract those DTOs
// into a shared leaf package that both backends implement — see
// docs/pure-go-hdf5-analysis.md.
//
// Findings this spike encodes (all verified by backend_test.go against a
// gonum-written fixture):
//   - The pure-Go reader opens files written by the cgo/C library and returns
//     correct values — read interop holds.
//   - scigolib v0.14.1 exposes values via Dataset.Read() ([]float64) and names,
//     but NOT structured shape/dtype (its core package is internal). Shape and
//     dtype are recovered here by parsing Dataset.Info()'s human string, which
//     is fragile and should be treated as a gap to raise upstream.
//   - Read() returns []float64 for all numeric datasets, so 64-bit integers lose
//     precision beyond 2^53. Our current service returns typed []int64; this is
//     a fidelity difference to resolve before any production swap.
package purego

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	scigo "github.com/scigolib/hdf5"
)

// Node mirrors hdf5store.Node (a file/group/dataset in the structure tree).
type Node struct {
	Name     string
	Path     string
	Type     string
	Shape    []uint
	Dtype    string
	NPoints  int
	Children []*Node
}

// Dataset mirrors hdf5store.Dataset (a dataset payload).
type Dataset struct {
	Name      string
	Path      string
	Shape     []uint
	Dtype     string
	NPoints   int
	Truncated bool
	Data      any
}

// ReadBackend is the read seam a pure-Go backend must satisfy. The production
// store would delegate its Structure/ReadDataset methods to an implementation
// of this interface instead of calling a library directly.
type ReadBackend interface {
	Structure(path, displayName string) (*Node, error)
	ReadDataset(path, datasetPath string, maxPoints int) (*Dataset, error)
}

// Backend is a scigolib-backed, cgo-free ReadBackend.
type Backend struct{}

var _ ReadBackend = Backend{}

// Structure walks the file and returns the hierarchical node tree.
func (Backend) Structure(path, displayName string) (*Node, error) {
	f, err := scigo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	root := &Node{Name: displayName, Path: "/", Type: "file"}
	for _, child := range f.Root().Children() {
		n, err := nodeFromObject(child, "")
		if err != nil {
			return nil, err
		}
		root.Children = append(root.Children, n)
	}
	return root, nil
}

// ReadDataset finds a dataset by path and returns its values (as []float64).
func (Backend) ReadDataset(path, datasetPath string, maxPoints int) (*Dataset, error) {
	f, err := scigo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	want := normalizePath(datasetPath)
	ds := findDataset(f.Root(), "", want)
	if ds == nil {
		return nil, fmt.Errorf("not found: %s", datasetPath)
	}

	shape, dtype, _ := parseInfo(ds)
	npoints := 1
	if len(shape) == 0 {
		npoints = 0
	}
	for _, d := range shape {
		npoints *= int(d)
	}

	out := &Dataset{
		Name:    leaf(want),
		Path:    want,
		Shape:   shape,
		Dtype:   dtype,
		NPoints: npoints,
	}
	if npoints > maxPoints {
		out.Truncated = true
		return out, nil
	}
	data, err := ds.Read()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", datasetPath, err)
	}
	if out.NPoints == 0 {
		out.NPoints = len(data)
	}
	out.Data = data
	return out, nil
}

func nodeFromObject(obj scigo.Object, parentPath string) (*Node, error) {
	full := parentPath + "/" + strings.Trim(obj.Name(), "/")
	switch v := obj.(type) {
	case *scigo.Group:
		n := &Node{Name: leaf(full), Path: full, Type: "group"}
		for _, c := range v.Children() {
			child, err := nodeFromObject(c, full)
			if err != nil {
				return nil, err
			}
			n.Children = append(n.Children, child)
		}
		return n, nil
	case *scigo.Dataset:
		shape, dtype, _ := parseInfo(v)
		npoints := 1
		if len(shape) == 0 {
			npoints = 0
		}
		for _, d := range shape {
			npoints *= int(d)
		}
		return &Node{
			Name:    leaf(full),
			Path:    full,
			Type:    "dataset",
			Shape:   shape,
			Dtype:   dtype,
			NPoints: npoints,
		}, nil
	default:
		return &Node{Name: leaf(full), Path: full, Type: "unknown"}, nil
	}
}

func findDataset(g *scigo.Group, parentPath, want string) *scigo.Dataset {
	for _, obj := range g.Children() {
		full := parentPath + "/" + strings.Trim(obj.Name(), "/")
		switch v := obj.(type) {
		case *scigo.Group:
			if ds := findDataset(v, full, want); ds != nil {
				return ds
			}
		case *scigo.Dataset:
			if full == want {
				return v
			}
		}
	}
	return nil
}

// infoDtypeRe matches the leading "<class> (size=<n> bytes)" of Dataset.Info().
var infoDtypeRe = regexp.MustCompile(`Dataset:\s+(\w+)\s+\(size=(\d+)\s+bytes\)`)

// infoShapeRe matches the "array [a x b x ...]" portion of Dataset.Info().
var infoShapeRe = regexp.MustCompile(`array\s+\[([0-9 x]+)\]`)

// parseInfo extracts shape and a normalized dtype string from Dataset.Info().
// This is a spike-grade parser of a human-readable string because scigolib
// v0.14.1 offers no structured accessor (its core package is internal).
func parseInfo(ds *scigo.Dataset) (shape []uint, dtype string, err error) {
	info, err := ds.Info()
	if err != nil {
		return nil, "", err
	}
	if m := infoDtypeRe.FindStringSubmatch(info); m != nil {
		class, size := m[1], m[2]
		bytes, _ := strconv.Atoi(size)
		switch class {
		case "integer":
			dtype = fmt.Sprintf("int%d", bytes*8)
		case "float":
			dtype = fmt.Sprintf("float%d", bytes*8)
		case "string":
			dtype = "string"
		default:
			dtype = class
		}
	}
	if m := infoShapeRe.FindStringSubmatch(info); m != nil {
		for _, part := range strings.Split(m[1], "x") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if d, e := strconv.Atoi(part); e == nil {
				shape = append(shape, uint(d))
			}
		}
	}
	return shape, dtype, nil
}

func normalizePath(p string) string {
	p = "/" + strings.Trim(p, "/")
	return p
}

func leaf(p string) string {
	p = strings.Trim(p, "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}
