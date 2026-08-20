# Architecture

HDF5 Agent is a single HTTP service. File access stays inside this process.
Every other service in a data platform uses `/api/v1` (JSON) or `pkg/hdf5client`.

```
browser / catalog / batch job
        │  HTTP JSON  (X-Request-ID)
        ▼
   hdf5-agent  (:8080)
        │  internal/hdf5store (cgo → libhdf5)
        ▼
   HDF5_DATA_DIR/*.h5
```

## Packages

| Package | Role |
|---------|------|
| `cmd/hdf5-agent` | Process: env config, slog, signals, graceful shutdown |
| `internal/config` | Environment-only configuration |
| `internal/hdf5store` | HDF5 operations (replaces the old `backend.py` / h5py path) |
| `internal/httpserver` | Versioned routes, structured errors, CORS, request IDs |
| `internal/apierror` | Error envelope `{error:{code,message,request_id}}` |
| `pkg/hdf5client` | Typed HTTP client for sibling services |
| `api` | Embedded OpenAPI document |

HDF5 C calls are serialized with a mutex. Distro libhdf5 builds are often not
thread-safe.

## Test strategy

1. **Pure unit tests** (no CGO): config, path sanitization, error JSON, HTTP
   handlers against a fake `Store`, HTTP client against `httptest`.
2. **HDF5 integration tests** (`internal/hdf5store`): create a temp file with
   `WriteSample`, then list, walk, read, update, create, delete.
3. **Frontend**: Vitest for API URL/error helpers. The UI is still thin; Go
   owns the behavior that used to live in Python.
4. **CI**: `gofmt`, `go vet`, `go test ./...` (must fail the job), frontend
   lint/test/build, Docker image + `/healthz`. Metrics are generated every run
   and compared to `metrics/baseline.json`.

There is no Python remaining in the repository.

## Production behavior

- Timeouts on the `http.Server` (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`)
- `Shutdown` on SIGINT/SIGTERM
- Structured JSON logs (`log/slog`)
- `/healthz` vs `/readyz`
- Non-root container user 65532
- Path checks reject `..` and names that are not `*.h5` / `*.hdf5` basenames
