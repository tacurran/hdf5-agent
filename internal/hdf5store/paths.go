package hdf5store

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ErrInvalidName is returned when a file name is not a safe HDF5 basename.
	ErrInvalidName = errors.New("invalid file name")
	// ErrInvalidPath is returned when a dataset path is not a safe HDF5 object path.
	ErrInvalidPath = errors.New("invalid dataset path")
	// ErrNotFound is returned when a file or dataset does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned when creating a file that already exists.
	ErrConflict = errors.New("already exists")
	// ErrUnsupportedType is returned when a dataset datatype cannot be serialized.
	ErrUnsupportedType = errors.New("unsupported datatype")
	// ErrTooLarge is returned when an in-place update exceeds the configured point limit.
	ErrTooLarge = errors.New("dataset too large")
)

var fileNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*\.(h5|hdf5)$`)

// ValidateFileName checks that name is a basename with an HDF5 extension and
// contains no path separators or parent-directory segments.
func ValidateFileName(name string) error {
	if name == "" || name != filepath.Base(name) || strings.Contains(name, "\\") {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	if strings.Contains(name, "..") || strings.ContainsRune(name, 0) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	if !fileNameRE.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return nil
}

// NormalizeDatasetPath validates and canonicalizes an HDF5 object path.
func NormalizeDatasetPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" || strings.ContainsRune(p, 0) {
		return "", fmt.Errorf("%w: empty", ErrInvalidPath)
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return "", fmt.Errorf("%w: %q", ErrInvalidPath, p)
		}
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	clean := path.Clean(p)
	if clean == "." || strings.Contains(clean, "..") {
		return "", fmt.Errorf("%w: %q", ErrInvalidPath, p)
	}
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	if clean != "/" && strings.HasSuffix(clean, "/") {
		clean = strings.TrimSuffix(clean, "/")
	}
	for _, part := range strings.Split(strings.TrimPrefix(clean, "/"), "/") {
		if part == "" && clean != "/" {
			return "", fmt.Errorf("%w: %q", ErrInvalidPath, p)
		}
	}
	return clean, nil
}

func joinHDF5(parent, name string) string {
	if parent == "" || parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

func isHDF5Name(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".h5") || strings.HasSuffix(lower, ".hdf5")
}
