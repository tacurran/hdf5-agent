#!/usr/bin/env bash
set -euo pipefail

# The gonum.org/v1/hdf5 package hardcodes CGO paths for Intel Homebrew
# (/usr/local), which is wrong on Apple Silicon (/opt/homebrew). Derive the
# real paths from pkg-config so builds work regardless of prefix. Existing
# CGO_CFLAGS/CGO_LDFLAGS in the environment take precedence.
if command -v pkg-config >/dev/null 2>&1; then
  if pkg-config --exists hdf5 2>/dev/null; then
    hdf5_pc=hdf5
  elif pkg-config --exists hdf5-serial 2>/dev/null; then
    hdf5_pc=hdf5-serial
  fi
  if [ -n "${hdf5_pc:-}" ]; then
    export CGO_CFLAGS="${CGO_CFLAGS:-$(pkg-config --cflags-only-I "$hdf5_pc")}"
    export CGO_LDFLAGS="${CGO_LDFLAGS:-$(pkg-config --libs-only-L "$hdf5_pc")}"
  fi
fi

cmd="${1:-dev}"
case "$cmd" in
  check)
    command -v go >/dev/null
    command -v node >/dev/null
    pkg-config --exists hdf5 || pkg-config --exists hdf5-serial
    ;;
  setup) make setup ;;
  build) make build ;;
  test) make test ;;
  docker) make docker ;;
  testdata) make testdata ;;
  dev)
    mkdir -p data
    HDF5_DATA_DIR=./data STATIC_DIR= LOG_FORMAT=text go run ./cmd/hdf5-agent &
    backend=$!
    trap 'kill "$backend" 2>/dev/null || true' EXIT
    (cd frontend && npm run dev)
    ;;
  *)
    echo "Usage: $0 [check|setup|build|test|docker|testdata|dev]"
    exit 1
    ;;
esac
