# HDF5 Agent

HTTP service for browsing and manipulating HDF5 files. The backend is Go
(`gonum.org/v1/hdf5`). The UI is a Vite/React app. Other services talk to this
one over a versioned JSON API — they do not link libhdf5 and they do not use
Python.

## Quick start

**Library dependencies:** Go 1.26+, Node 20+, HDF5 C development headers
(`libhdf5-dev` or `brew install hdf5`).

```bash
make setup
make testdata
make test
HDF5_DATA_DIR=./data STATIC_DIR= LOG_FORMAT=text go run ./cmd/hdf5-agent
```

In another terminal:

```bash
cd frontend && npm run dev
```

Open http://localhost:3000. The Vite dev server proxies `/api` to the Go process
on port 8080.

**Docker (API + built UI on one port):**

```bash
mkdir -p data
make testdata
docker compose up --build
curl -fsS http://127.0.0.1:8080/healthz
```

The UI is served from the same origin as the API at http://localhost:8080.

## HTTP API

Canonical prefix: **`/api/v1`**. Full contract: [`api/openapi.yaml`](api/openapi.yaml)
(also `GET /api/v1/openapi.yaml` on a running process).

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/healthz`, `/api/v1/health` | Liveness |
| GET | `/readyz`, `/api/v1/ready` | Data directory readable |
| GET | `/api/v1/files` | List `.h5` / `.hdf5` files |
| POST | `/api/v1/files` | Create empty file `{ "name": "x.h5" }` |
| GET | `/api/v1/files/{name}` | Group/dataset tree |
| DELETE | `/api/v1/files/{name}` | Delete file |
| GET | `/api/v1/files/{name}/datasets?path=` | Read dataset |
| PUT | `/api/v1/files/{name}/datasets` | Update flattened indices |

Errors:

```json
{ "error": { "code": "not_found", "message": "not found: missing.h5", "request_id": "…" } }
```

Send `X-Request-ID` to correlate caller and service logs. Go callers should use
[`pkg/hdf5client`](pkg/hdf5client).

## Configuration

All configuration is environment variables. No secrets are required.

| Variable | Default | Meaning |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | Listen address (`PORT` used if `HTTP_ADDR` is unset) |
| `HDF5_DATA_DIR` | `./data` | Directory of HDF5 files |
| `STATIC_DIR` | `./public` | Built frontend; empty disables static serving |
| `CORS_ORIGINS` | `*` | Comma-separated origins |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `60s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout |
| `SHUTDOWN_TIMEOUT` | `20s` | Graceful shutdown |
| `MAX_DATASET_POINTS` | `100000` | Read/update size cap |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |

The container runs as uid 65532 and expects a writable `/data`.

## Composition example

[`examples/data-catalog`](examples/data-catalog) is a second Go service that
builds a dataset inventory by calling this API. It never uses the HDF5 C library.

```bash
docker compose -f examples/docker-compose.yml up --build
curl -s http://127.0.0.1:8090/inventory
```

## Tests and quality metrics

```bash
make test      # go test ./... and frontend vitest; failures fail the build
make lint      # gofmt + go vet + eslint (no || true)
make metrics   # writes metrics/current.json and prints a delta vs baseline
```

CI is [`.github/workflows/ci.yml`](.github/workflows/ci.yml) and runs on pull
requests to `main`. Metrics history lives in [`metrics/`](metrics/).

## Layout

```
cmd/hdf5-agent          service entrypoint
cmd/create-testdata     sample HDF5 writer
cmd/metrics             quality metrics collector
internal/hdf5store      Go HDF5 library used by the service
internal/httpserver     HTTP handlers, middleware, shutdown
pkg/hdf5client          HTTP client for other Go services
api/openapi.yaml        OpenAPI 3 contract
frontend/               React UI
examples/data-catalog   sibling catalog service
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
