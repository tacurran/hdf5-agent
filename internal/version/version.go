// Package version holds build identity for the HDF5 Agent service.
package version

// Name is the service name reported on health endpoints and in logs.
const Name = "hdf5-agent"

// API is the current HTTP API version prefix (without a leading slash).
const API = "v1"

// Version is the service release version. Override at build time with:
//
//	go build -ldflags "-X github.com/tacurran/hdf5-agent/internal/version.Version=1.2.3"
var Version = "1.1.0"
