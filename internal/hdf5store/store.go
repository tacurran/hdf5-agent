// Package hdf5store is the Go replacement for the former Python h5py backend.
//
// It owns all HDF5 file access for the service: listing files, walking groups,
// reading and updating datasets, and creating or deleting files. Callers in other
// services should not import this package; they should use the versioned HTTP API
// or pkg/hdf5client.
package hdf5store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gonum.org/v1/hdf5"
)

// Store is a directory-backed HDF5 archive. All HDF5 C calls are serialized
// because typical distro libhdf5 builds are not thread-safe.
type Store struct {
	dir              string
	maxDatasetPoints int
	mu               sync.Mutex
}

// Open prepares a Store rooted at dir. The directory is created if missing.
func Open(dir string, maxDatasetPoints int) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("data directory is required")
	}
	if maxDatasetPoints <= 0 {
		return nil, fmt.Errorf("maxDatasetPoints must be positive")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return &Store{dir: dir, maxDatasetPoints: maxDatasetPoints}, nil
}

// Dir returns the configured data directory.
func (s *Store) Dir() string {
	return s.dir
}

// Ready reports whether the data directory is readable.
func (s *Store) Ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := os.ReadDir(s.dir)
	return err
}

// ListFiles returns HDF5 files in the data directory, sorted by name.
func (s *Store) ListFiles(ctx context.Context) ([]FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	out := make([]FileInfo, 0)
	for _, e := range entries {
		if e.IsDir() || !isHDF5Name(e.Name()) {
			continue
		}
		if err := ValidateFileName(e.Name()); err != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, FileInfo{Name: e.Name(), SizeBytes: info.Size()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Structure returns the hierarchical contents of an HDF5 file.
func (s *Store) Structure(ctx context.Context, name string) (*Node, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.filePath(name)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return nil, err
	}

	var root *Node
	err = s.withFile(path, hdf5.F_ACC_RDONLY, func(f *hdf5.File) error {
		children, err := listChildren(&f.CommonFG, "/")
		if err != nil {
			return err
		}
		root = &Node{Name: name, Path: "/", Type: "file", Children: children}
		return nil
	})
	return root, err
}

// ReadDataset loads values from a dataset. If the dataset has more than
// maxDatasetPoints elements, Data is omitted and Truncated is true.
func (s *Store) ReadDataset(ctx context.Context, name, datasetPath string) (*Dataset, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.filePath(name)
	if err != nil {
		return nil, err
	}
	dsPath, err := NormalizeDatasetPath(datasetPath)
	if err != nil {
		return nil, err
	}

	var out *Dataset
	err = s.withFile(path, hdf5.F_ACC_RDONLY, func(f *hdf5.File) error {
		dset, err := f.OpenDataset(dsPath)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrNotFound, dsPath)
		}
		defer dset.Close()
		out, err = readOpenedDataset(dset, s.maxDatasetPoints)
		return err
	})
	if err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return out, err
}

// CreateFile creates an empty HDF5 file. It fails if the name already exists.
func (s *Store) CreateFile(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.filePath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrConflict, name)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := hdf5.CreateFile(path, hdf5.F_ACC_EXCL)
	if err != nil {
		if _, statErr := os.Stat(path); statErr == nil {
			return fmt.Errorf("%w: %s", ErrConflict, name)
		}
		return fmt.Errorf("create hdf5 file: %w", err)
	}
	return f.Close()
}

// DeleteFile removes an HDF5 file from the data directory.
func (s *Store) DeleteFile(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.filePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return err
	}
	return nil
}

// UpdateDataset writes values at the given flattened indices.
func (s *Store) UpdateDataset(ctx context.Context, name, datasetPath string, indices []int, values []any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(indices) != len(values) {
		return fmt.Errorf("%w: indices and values length mismatch", ErrInvalidPath)
	}
	if len(indices) == 0 {
		return fmt.Errorf("%w: no values to write", ErrInvalidPath)
	}
	path, err := s.filePath(name)
	if err != nil {
		return err
	}
	dsPath, err := NormalizeDatasetPath(datasetPath)
	if err != nil {
		return err
	}

	return s.withFile(path, hdf5.F_ACC_RDWR, func(f *hdf5.File) error {
		dset, err := f.OpenDataset(dsPath)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrNotFound, dsPath)
		}
		defer dset.Close()

		space := dset.Space()
		if space == nil {
			return fmt.Errorf("dataset space: %w", ErrUnsupportedType)
		}
		defer space.Close()
		npoints := space.SimpleExtentNPoints()
		if npoints > s.maxDatasetPoints {
			return fmt.Errorf("%w: %d points", ErrTooLarge, npoints)
		}
		dtype, err := dset.Datatype()
		if err != nil {
			return err
		}
		defer dtype.Close()
		return writeIndices(dset, dtype, npoints, indices, values)
	})
}

func (s *Store) filePath(name string) (string, error) {
	if err := ValidateFileName(name); err != nil {
		return "", err
	}
	absDir, err := filepath.Abs(s.dir)
	if err != nil {
		return "", err
	}
	full := filepath.Join(absDir, name)
	if filepath.Dir(full) != absDir {
		return "", fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return full, nil
}

func (s *Store) withFile(path string, flags int, fn func(*hdf5.File) error) (err error) {
	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return fmt.Errorf("%w: %s", ErrNotFound, filepath.Base(path))
		}
		return statErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("hdf5 panic: %v", r)
		}
	}()
	f, err := hdf5.OpenFile(path, flags)
	if err != nil {
		return fmt.Errorf("open hdf5 file: %w", err)
	}
	defer f.Close()
	return fn(f)
}

func listChildren(fg *hdf5.CommonFG, parentPath string) ([]*Node, error) {
	n, err := fg.NumObjects()
	if err != nil {
		return nil, err
	}
	children := make([]*Node, 0, int(n))
	for i := uint(0); i < n; i++ {
		name, err := fg.ObjectNameByIndex(i)
		if err != nil {
			continue
		}
		gtyp, err := fg.ObjectTypeByIndex(i)
		if err != nil {
			continue
		}
		full := joinHDF5(parentPath, name)
		switch gtyp {
		case hdf5.H5G_GROUP:
			g, err := fg.OpenGroup(name)
			if err != nil {
				continue
			}
			sub, err := listChildren(&g.CommonFG, full)
			_ = g.Close()
			if err != nil {
				return nil, err
			}
			children = append(children, &Node{
				Name:     name,
				Path:     full,
				Type:     "group",
				Children: sub,
			})
		case hdf5.H5G_DATASET:
			node, err := datasetNode(fg, name, full)
			if err != nil {
				continue
			}
			children = append(children, node)
		}
	}
	return children, nil
}

func datasetNode(fg *hdf5.CommonFG, name, full string) (*Node, error) {
	dset, err := fg.OpenDataset(name)
	if err != nil {
		return nil, err
	}
	defer dset.Close()
	space := dset.Space()
	if space == nil {
		return nil, fmt.Errorf("no dataspace")
	}
	defer space.Close()
	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}
	dtype, err := dset.Datatype()
	if err != nil {
		return nil, err
	}
	defer dtype.Close()
	return &Node{
		Name:    name,
		Path:    full,
		Type:    "dataset",
		Shape:   dims,
		Dtype:   datatypeName(int(dtype.Class()), dtype.Size()),
		NPoints: space.SimpleExtentNPoints(),
	}, nil
}

func readOpenedDataset(dset *hdf5.Dataset, maxPoints int) (*Dataset, error) {
	space := dset.Space()
	if space == nil {
		return nil, fmt.Errorf("%w: no dataspace", ErrUnsupportedType)
	}
	defer space.Close()
	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}
	npoints := space.SimpleExtentNPoints()
	dtype, err := dset.Datatype()
	if err != nil {
		return nil, err
	}
	defer dtype.Close()
	out := &Dataset{
		Name:    filepath.Base(dset.Name()),
		Path:    dset.Name(),
		Shape:   dims,
		Dtype:   datatypeName(int(dtype.Class()), dtype.Size()),
		NPoints: npoints,
	}
	if npoints > maxPoints {
		out.Truncated = true
		out.Data = nil
		return out, nil
	}
	data, err := readValues(dset, dtype, npoints)
	if err != nil {
		return nil, err
	}
	out.Data = data
	return out, nil
}

func readValues(dset *hdf5.Dataset, dtype *hdf5.Datatype, npoints int) (any, error) {
	class := dtype.Class()
	size := dtype.Size()
	switch class {
	case hdf5.T_INTEGER:
		switch size {
		case 8:
			data := make([]int64, npoints)
			if err := dset.Read(&data); err != nil {
				return nil, err
			}
			return data, nil
		case 4:
			data := make([]int32, npoints)
			if err := dset.Read(&data); err != nil {
				return nil, err
			}
			return data, nil
		case 2:
			data := make([]int16, npoints)
			if err := dset.Read(&data); err != nil {
				return nil, err
			}
			return data, nil
		case 1:
			data := make([]int8, npoints)
			if err := dset.Read(&data); err != nil {
				return nil, err
			}
			return data, nil
		}
	case hdf5.T_FLOAT:
		if size == 8 {
			data := make([]float64, npoints)
			if err := dset.Read(&data); err != nil {
				return nil, err
			}
			return data, nil
		}
		data := make([]float32, npoints)
		if err := dset.Read(&data); err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, fmt.Errorf("%w: class=%d size=%d", ErrUnsupportedType, class, size)
}

func writeIndices(dset *hdf5.Dataset, dtype *hdf5.Datatype, npoints int, indices []int, values []any) error {
	for _, idx := range indices {
		if idx < 0 || idx >= npoints {
			return fmt.Errorf("%w: index %d out of range", ErrInvalidPath, idx)
		}
	}
	class := dtype.Class()
	size := dtype.Size()
	switch class {
	case hdf5.T_FLOAT:
		if size == 8 {
			data := make([]float64, npoints)
			if err := dset.Read(&data); err != nil {
				return err
			}
			for i, idx := range indices {
				v, err := asFloat64(values[i])
				if err != nil {
					return err
				}
				data[idx] = v
			}
			return dset.Write(&data)
		}
		data := make([]float32, npoints)
		if err := dset.Read(&data); err != nil {
			return err
		}
		for i, idx := range indices {
			v, err := asFloat64(values[i])
			if err != nil {
				return err
			}
			data[idx] = float32(v)
		}
		return dset.Write(&data)
	case hdf5.T_INTEGER:
		data := make([]int64, npoints)
		switch size {
		case 8:
			if err := dset.Read(&data); err != nil {
				return err
			}
		case 4:
			tmp := make([]int32, npoints)
			if err := dset.Read(&tmp); err != nil {
				return err
			}
			for i, v := range tmp {
				data[i] = int64(v)
			}
		default:
			return fmt.Errorf("%w: integer size %d", ErrUnsupportedType, size)
		}
		for i, idx := range indices {
			v, err := asInt64(values[i])
			if err != nil {
				return err
			}
			data[idx] = v
		}
		if size == 8 {
			return dset.Write(&data)
		}
		tmp := make([]int32, npoints)
		for i, v := range data {
			tmp[i] = int32(v)
		}
		return dset.Write(&tmp)
	default:
		return fmt.Errorf("%w: class=%d", ErrUnsupportedType, class)
	}
}

func asFloat64(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	case string:
		var f float64
		_, err := fmt.Sscan(n, &f)
		return f, err
	default:
		return 0, fmt.Errorf("%w: cannot convert %T to float", ErrInvalidPath, v)
	}
}

func asInt64(v any) (int64, error) {
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case json.Number:
		return n.Int64()
	case string:
		var i int64
		_, err := fmt.Sscan(n, &i)
		return i, err
	default:
		return 0, fmt.Errorf("%w: cannot convert %T to int", ErrInvalidPath, v)
	}
}
