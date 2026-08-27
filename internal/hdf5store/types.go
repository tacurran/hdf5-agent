package hdf5store

import (
	"encoding/json"
	"fmt"
)

// FileInfo is metadata about an HDF5 file in the data directory.
type FileInfo struct {
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
}

// Node is an HDF5 file, group, or dataset in the structure tree.
type Node struct {
	Name     string  `json:"name"`
	Path     string  `json:"path,omitempty"`
	Type     string  `json:"type"`
	Shape    []uint  `json:"shape,omitempty"`
	Dtype    string  `json:"dtype,omitempty"`
	NPoints  int     `json:"npoints,omitempty"`
	Children []*Node `json:"children,omitempty"`
}

// Dataset is a dataset payload returned by Read.
type Dataset struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Shape     []uint `json:"shape"`
	Dtype     string `json:"dtype"`
	NPoints   int    `json:"npoints"`
	Truncated bool   `json:"truncated"`
	Data      any    `json:"data"`
}

// UpdateRequest is the JSON body for dataset value updates.
type UpdateRequest struct {
	Path    string `json:"path"`
	Indices []int  `json:"indices,omitempty"`
	Values  []any  `json:"values,omitempty"`
	// Data is an alias for Values accepted for compatibility with the UI.
	Data []any `json:"data,omitempty"`
}

// ValuesOrData returns Values, falling back to Data.
func (r UpdateRequest) ValuesOrData() []any {
	if len(r.Values) > 0 {
		return r.Values
	}
	return r.Data
}

// CreateFileRequest is the JSON body for creating an empty HDF5 file.
type CreateFileRequest struct {
	Name string `json:"name"`
}

// FilesResponse is the list-files JSON envelope.
type FilesResponse struct {
	Files []FileInfo `json:"files"`
}

// HealthResponse is returned by liveness endpoints.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	API     string `json:"api"`
}

// ReadyResponse is returned by readiness endpoints.
type ReadyResponse struct {
	Status  string `json:"status"`
	DataDir string `json:"data_dir"`
}

// MarshalJSON keeps Files non-null so clients can iterate safely.
func (r FilesResponse) MarshalJSON() ([]byte, error) {
	type alias FilesResponse
	if r.Files == nil {
		r.Files = []FileInfo{}
	}
	return json.Marshal(alias(r))
}

func datatypeName(class int, size uint) string {
	switch class {
	case 0: // H5T_INTEGER
		return fmt.Sprintf("int%d", size*8)
	case 1: // H5T_FLOAT
		return fmt.Sprintf("float%d", size*8)
	case 3: // H5T_STRING
		return "string"
	case 6:
		return "compound"
	case 8:
		return "enum"
	case 9:
		return "vlen"
	case 10:
		return "array"
	default:
		return fmt.Sprintf("type_%d", class)
	}
}
