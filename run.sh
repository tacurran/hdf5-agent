#!/usr/bin/env bash
set -euo pipefail
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
