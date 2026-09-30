# Project Summary

**HDF5 Agent** is an HTTP service for browsing and manipulating HDF5 files through a versioned JSON API. It consists of a Go backend (using CGO bindings to libhdf5 via `gonum.org/v1/hdf5`) and a React/Vite frontend. Other services in a data platform call this agent's REST API rather than linking against the HDF5 C library directly.

## Architecture

The service is a single HTTP process that centralizes all HDF5 file access:

```
browser / catalog / batch job
        │  HTTP JSON  (X-Request-ID)
        ▼
   hdf5-agent  (:8080)
        │  internal/hdf5store (cgo → libhdf5)
        ▼
   HDF5_DATA_DIR/*.h5
```

HDF5 C calls are serialized with a mutex since distro libhdf5 builds are often not thread-safe.

**Note**: This describes the current implementation. See [ARCHITECTURE.md](ARCHITECTURE.md)
for planned future architecture including a Go BFF layer and integration with the
[Twothink Alembic Identity Plane](https://github.com/twothinkinc/alembic-identity-plane-go)
for authentication and authorization.

## Repository Structure

```
cmd/
  hdf5-agent/          Service entrypoint (config, logging, graceful shutdown)
  create-testdata/     Sample HDF5 file writer for tests
  metrics/             Quality metrics collector

internal/
  config/              Environment-only configuration
  hdf5store/           HDF5 operations (replaces old Python backend)
  httpserver/          Versioned routes, structured errors, CORS, request IDs
  apierror/            Error envelope {error:{code,message,request_id}}
  version/             API version constants

pkg/
  hdf5client/          Typed HTTP client for Go services that call this API

api/
  openapi.yaml         OpenAPI 3 contract (served at /api/v1/openapi.yaml)

frontend/              React 18 + Vite 8 UI
  src/                 Components, API client, and tests (Vitest)

examples/
  data-catalog/        Sibling service example using pkg/hdf5client

metrics/               Baseline and CI-generated quality snapshots
docs/                  Architecture ADRs and HDF5 library analysis
```

## Key Features

- **Versioned JSON API** (`/api/v1`) with OpenAPI contract
- **Health endpoints** (`/healthz`, `/readyz`) for orchestrators
- **File operations**: list, create, delete HDF5 files
- **Dataset operations**: read and in-place update with flattened indices
- **Request correlation** via `X-Request-ID` header
- **Structured errors** with stable error codes
- **Graceful shutdown** on SIGINT/SIGTERM
- **Configurable via environment variables** (no secrets required)
- **Docker support** with non-root user (uid 65532)
- **Quality metrics** tracked in CI against baseline

## API Overview

Base path: `/api/v1`

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/healthz`, `/api/v1/health` | Liveness |
| GET | `/readyz`, `/api/v1/ready` | Data directory readable |
| GET | `/api/v1/files` | List `.h5` / `.hdf5` files |
| POST | `/api/v1/files` | Create empty file |
| GET | `/api/v1/files/{name}` | Group/dataset tree |
| DELETE | `/api/v1/files/{name}` | Delete file |
| GET | `/api/v1/files/{name}/datasets?path=` | Read dataset |
| PUT | `/api/v1/files/{name}/datasets` | Update flattened indices |

Full contract: `api/openapi.yaml` (Apache-2.0 licensed)

## Technology Stack

- **Backend**: Go 1.26+, `gonum.org/v1/hdf5` (CGO bindings to libhdf5)
- **Frontend**: React 18, Vite 8, Vitest
- **API**: OpenAPI 3.0.3, structured JSON errors
- **Logging**: `log/slog` with JSON or text output
- **HTTP**: `net/http` with timeouts and graceful shutdown
- **Testing**: Go test with coverage, frontend Vitest
- **CI**: GitHub Actions (lint, test, metrics, Docker build)
- **Containerization**: Docker multi-stage build, non-root runtime

## Development

**Prerequisites**: Go 1.26+, Node 20+, HDF5 C headers (`libhdf5-dev` or `brew install hdf5`)

**Quick start**:

```bash
make setup      # Install Go and frontend dependencies
make testdata   # Generate data/sample.h5
make test       # Run Go and frontend tests
./run.sh dev    # Start backend + frontend with CGO flags
```

The helper script `run.sh` runs the Go backend and Vite dev server together and automatically sets HDF5 CGO flags.

**macOS / Apple Silicon note**: The vendored `gonum.org/v1/hdf5` hardcodes CGO paths for Intel Homebrew. Use `make`, `./run.sh`, or [direnv](https://direnv.net) (`.envrc` included) to set correct paths automatically. See README for manual export instructions.

**Manual process option**:

```bash
HDF5_DATA_DIR=./data STATIC_DIR= LOG_FORMAT=text go run ./cmd/hdf5-agent
cd frontend && npm run dev  # separate terminal
```

Frontend dev server: http://localhost:3000 (proxies `/api` to :8080)

**Docker**:

```bash
mkdir -p data
make testdata
docker compose up --build
```

UI and API served from same origin at http://localhost:8080

## Configuration

All configuration via environment variables (no secrets):

| Variable | Default | Meaning |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `HDF5_DATA_DIR` | `./data` | HDF5 files directory |
| `STATIC_DIR` | `./public` | Built frontend (empty disables) |
| `CORS_ORIGINS` | `*` | Comma-separated origins |
| `MAX_DATASET_POINTS` | `100000` | Read/update size cap |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `60s` | HTTP write timeout |

See README for full configuration table.

## Testing and CI

```bash
make test      # Go test ./... + frontend vitest
make lint      # gofmt + go vet + eslint (failures fail the build)
make metrics   # Generate metrics/current.json and compare to baseline
```

**Test strategy**:
- Pure unit tests (no CGO): config, path sanitization, error JSON, HTTP handlers against fake Store
- HDF5 integration tests: temp file operations (create, list, walk, read, update, delete)
- Frontend: Vitest for API client helpers
- CI: `.github/workflows/ci.yml` runs on PRs to `main`

CI enforces: lint passes, tests pass, Docker build succeeds, metrics delta reported.

## Composition Example

`examples/data-catalog` is a sibling Go service that builds a dataset inventory by calling this API. It demonstrates the intended usage pattern: downstream services use `pkg/hdf5client` and never link libhdf5 directly.

```bash
docker compose -f examples/docker-compose.yml up --build
curl -s http://127.0.0.1:8090/inventory
```

## Migration Notes

This service replaces a previous Python backend (`backend.py` with h5py). All HDF5 operations now happen in `internal/hdf5store`. The OpenAPI contract and error format are stable.

## License

Apache License 2.0 (see LICENSE file)

## Module

`github.com/tacurran/hdf5-agent`
